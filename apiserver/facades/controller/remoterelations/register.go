// Copyright 2022 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package remoterelations

import (
	"context"
	"reflect"

	"github.com/juju/errors"

	"github.com/juju/juju/apiserver/common"
	commoncrossmodel "github.com/juju/juju/apiserver/common/crossmodel"
	"github.com/juju/juju/apiserver/facade"
	corelogger "github.com/juju/juju/core/logger"
	secretservice "github.com/juju/juju/domain/secret/service"
)

// Register is called to expose a package of facades onto a given registry.
func Register(registry facade.FacadeRegistry) {
	registry.MustRegister("RemoteRelations", 2, func(stdCtx context.Context, ctx facade.ModelContext) (facade.Facade, error) {
		return newAPI(ctx) // Adds UpdateControllersForModels and WatchLocalRelationChanges.
	}, reflect.TypeOf((*API)(nil)))
}

// newAPI creates a new server-side API facade backed by global state.
func newAPI(ctx facade.ModelContext) (*API, error) {
	systemState, err := ctx.StatePool().SystemState()
	if err != nil {
		return nil, errors.Trace(err)
	}
	serviceFactory := ctx.ServiceFactory()
	controllerConfigService := serviceFactory.ControllerConfig()
	externalControllerService := serviceFactory.ExternalController()
	return NewRemoteRelationsAPI(
		stateShim{st: ctx.State(), Backend: commoncrossmodel.GetBackend(ctx.State())},
		externalControllerService,
		serviceFactory.Secret(secretservice.NotImplementedBackendConfigGetter),
		common.NewControllerConfigAPI(systemState, controllerConfigService, externalControllerService),
		ctx.Resources(), ctx.Auth(),
		ctx.Logger().Child("remoterelations", corelogger.CMR),
	)
}
