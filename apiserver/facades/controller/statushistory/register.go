// Copyright 2022 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package statushistory

import (
	"context"
	"reflect"

	"github.com/juju/errors"

	"github.com/juju/juju/apiserver/common"
	"github.com/juju/juju/apiserver/facade"
)

// Register is called to expose a package of facades onto a given registry.
func Register(registry facade.FacadeRegistry) {
	registry.MustRegister("StatusHistory", 2, func(stdCtx context.Context, ctx facade.ModelContext) (facade.Facade, error) {
		return newAPI(ctx)
	}, reflect.TypeOf((*API)(nil)))
}

// newAPI returns an API Instance.
func newAPI(ctx facade.ModelContext) (*API, error) {
	m, err := Model(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &API{
		MongoModelWatcher: common.NewMongoModelWatcher(m, ctx.Resources()),
		st:                ctx.State(),
		authorizer:        ctx.Auth(),
	}, nil
}
