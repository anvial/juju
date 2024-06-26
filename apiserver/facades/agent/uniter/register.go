// Copyright 2022 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package uniter

import (
	"context"
	"reflect"

	"github.com/juju/errors"

	"github.com/juju/juju/apiserver/common"
	"github.com/juju/juju/apiserver/common/cloudspec"
	"github.com/juju/juju/apiserver/common/unitcommon"
	apiservererrors "github.com/juju/juju/apiserver/errors"
	"github.com/juju/juju/apiserver/facade"
	"github.com/juju/juju/core/model"
	"github.com/juju/juju/internal/secrets/provider"
)

// Register is called to expose a package of facades onto a given registry.
func Register(registry facade.FacadeRegistry) {
	registry.MustRegister("Uniter", 19, func(stdCtx context.Context, ctx facade.ModelContext) (facade.Facade, error) {
		return newUniterAPI(stdCtx, ctx)
	}, reflect.TypeOf((*UniterAPI)(nil)))
}

// newUniterAPI creates a new instance of the core Uniter API.
func newUniterAPI(_ context.Context, ctx facade.ModelContext) (*UniterAPI, error) {
	m, err := ctx.State().Model()
	if err != nil {
		return nil, errors.Trace(err)
	}
	serviceFactory := ctx.ServiceFactory()
	secretBackendService := serviceFactory.SecretBackend()
	secretBackendAdminConfigGetter := func(stdCtx context.Context) (*provider.ModelBackendConfigInfo, error) {
		return secretBackendService.GetSecretBackendConfigForAdmin(stdCtx, model.UUID(m.UUID()))
	}
	return newUniterAPIWithServices(
		ctx,
		serviceFactory.ControllerConfig(),
		serviceFactory.Secret(secretBackendAdminConfigGetter),
		serviceFactory.Cloud(),
		serviceFactory.Credential(),
		serviceFactory.Unit(),
	)
}

// newUniterAPIWithServices creates a new instance using the services.
func newUniterAPIWithServices(
	context facade.ModelContext,
	controllerConfigService ControllerConfigService,
	secretService SecretService,
	cloudService CloudService,
	credentialService CredentialService,
	unitRemover UnitRemover,
) (*UniterAPI, error) {
	authorizer := context.Auth()
	if !authorizer.AuthUnitAgent() && !authorizer.AuthApplicationAgent() {
		return nil, apiservererrors.ErrPerm
	}
	st := context.State()
	aClock := context.StatePool().Clock()
	resources := context.Resources()
	leadershipChecker, err := context.LeadershipChecker()
	if err != nil {
		return nil, errors.Trace(err)
	}
	leadershipRevoker, err := context.LeadershipRevoker()
	if err != nil {
		return nil, errors.Trace(err)
	}

	accessUnit := unitcommon.UnitAccessor(authorizer, unitcommon.Backend(st))
	accessApplication := applicationAccessor(authorizer, st)
	accessMachine := machineAccessor(authorizer, st)
	accessCloudSpec := cloudSpecAccessor(authorizer, st)

	m, err := st.Model()
	if err != nil {
		return nil, errors.Trace(err)
	}

	storageAccessor, err := getStorageState(st)
	if err != nil {
		return nil, errors.Trace(err)
	}
	storageAPI, err := newStorageAPI(
		stateShim{st}, storageAccessor, context.ServiceFactory().BlockDevice(), resources, accessUnit)
	if err != nil {
		return nil, errors.Trace(err)
	}

	accessUnitOrApplication := common.AuthAny(accessUnit, accessApplication)

	cloudSpec := cloudspec.NewCloudSpecV2(resources,
		cloudspec.MakeCloudSpecGetterForModel(st, cloudService, credentialService),
		cloudspec.MakeCloudSpecWatcherForModel(st, cloudService),
		cloudspec.MakeCloudSpecCredentialWatcherForModel(st),
		cloudspec.MakeCloudSpecCredentialContentWatcherForModel(st, credentialService),
		common.AuthFuncForTag(m.ModelTag()),
	)

	systemState, err := context.StatePool().SystemState()
	if err != nil {
		return nil, errors.Trace(err)
	}
	logger := context.Logger().Child("uniter")
	return &UniterAPI{
		LifeGetter:                 common.NewLifeGetter(st, accessUnitOrApplication),
		DeadEnsurer:                common.NewDeadEnsurer(st, common.RevokeLeadershipFunc(leadershipRevoker), accessUnit),
		AgentEntityWatcher:         common.NewAgentEntityWatcher(st, resources, accessUnitOrApplication),
		APIAddresser:               common.NewAPIAddresser(systemState, resources),
		MongoModelWatcher:          common.NewMongoModelWatcher(m, resources),
		RebootRequester:            common.NewRebootRequester(st, accessMachine),
		UpgradeSeriesAPI:           common.NewExternalUpgradeSeriesAPI(st, resources, authorizer, accessMachine, accessUnit, logger),
		UnitStateAPI:               common.NewExternalUnitStateAPI(controllerConfigService, st, resources, authorizer, accessUnit, logger),
		LeadershipSettingsAccessor: leadershipSettingsAccessorFactory(st, leadershipChecker, resources, authorizer),
		lxdProfileAPI:              NewExternalLXDProfileAPIv2(st, resources, authorizer, accessUnit, logger),
		// TODO(fwereade): so *every* unit should be allowed to get/set its
		// own status *and* its application's? This is not a pleasing arrangement.
		StatusAPI: NewStatusAPI(m, accessUnitOrApplication, leadershipChecker),

		m:                       m,
		st:                      st,
		controllerConfigService: controllerConfigService,
		secretService:           secretService,
		networkService:          context.ServiceFactory().Network(),
		cloudService:            cloudService,
		credentialService:       credentialService,
		unitRemover:             unitRemover,
		clock:                   aClock,
		auth:                    authorizer,
		resources:               resources,
		leadershipChecker:       leadershipChecker,
		accessUnit:              accessUnit,
		accessApplication:       accessApplication,
		accessMachine:           accessMachine,
		accessCloudSpec:         accessCloudSpec,
		cloudSpecer:             cloudSpec,
		StorageAPI:              storageAPI,
		logger:                  logger,
		store:                   context.ObjectStore(),
	}, nil
}
