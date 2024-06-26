// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package secretsdrain

import (
	"context"
	"reflect"

	"github.com/juju/errors"

	commonsecrets "github.com/juju/juju/apiserver/common/secrets"
	apiservererrors "github.com/juju/juju/apiserver/errors"
	"github.com/juju/juju/apiserver/facade"
	coremodel "github.com/juju/juju/core/model"
	"github.com/juju/juju/internal/secrets/provider"
)

// Register is called to expose a package of facades onto a given registry.
func Register(registry facade.FacadeRegistry) {
	registry.MustRegister("SecretsDrain", 1, func(stdCtx context.Context, ctx facade.ModelContext) (facade.Facade, error) {
		return newSecretsDrainAPI(ctx)
	}, reflect.TypeOf((*commonsecrets.SecretsDrainAPI)(nil)))
}

// newSecretsDrainAPI creates a SecretsDrainAPI.
func newSecretsDrainAPI(ctx facade.ModelContext) (*commonsecrets.SecretsDrainAPI, error) {
	if !ctx.Auth().AuthUnitAgent() {
		return nil, apiservererrors.ErrPerm
	}
	leadershipChecker, err := ctx.LeadershipChecker()
	if err != nil {
		return nil, errors.Trace(err)
	}
	model, err := ctx.State().Model()
	if err != nil {
		return nil, errors.Trace(err)
	}
	modelUUID := coremodel.UUID(model.UUID())
	authTag := ctx.Auth().GetAuthTag()

	secretBackendService := ctx.ServiceFactory().SecretBackend()
	secretBackendAdminConfigGetter := func(stdCtx context.Context) (*provider.ModelBackendConfigInfo, error) {
		return secretBackendService.GetSecretBackendConfigForAdmin(stdCtx, modelUUID)
	}
	return commonsecrets.NewSecretsDrainAPI(
		authTag,
		ctx.Auth(),
		ctx.Logger().Child("secretsdrain"),
		leadershipChecker,
		modelUUID,
		ctx.ServiceFactory().Secret(secretBackendAdminConfigGetter),
		secretBackendService,
		ctx.WatcherRegistry(),
	)
}
