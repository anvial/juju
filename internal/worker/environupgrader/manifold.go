// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package environupgrader

import (
	"context"

	"github.com/juju/errors"
	"github.com/juju/names/v5"
	"github.com/juju/worker/v4"
	"github.com/juju/worker/v4/dependency"

	"github.com/juju/juju/api/base"
	"github.com/juju/juju/core/logger"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/internal/worker/common"
	"github.com/juju/juju/internal/worker/gate"
)

// ManifoldConfig describes how to configure and construct a Worker,
// and what registered resources it may depend upon.
type ManifoldConfig struct {
	APICallerName string
	EnvironName   string
	GateName      string
	ControllerTag names.ControllerTag
	ModelTag      names.ModelTag
	Logger        logger.Logger

	NewFacade                    func(base.APICaller) (Facade, error)
	NewWorker                    func(context.Context, Config) (worker.Worker, error)
	NewCredentialValidatorFacade func(base.APICaller) (common.CredentialAPI, error)
}

func (config ManifoldConfig) start(ctx context.Context, getter dependency.Getter) (worker.Worker, error) {
	var environ environs.Environ
	if err := getter.Get(config.EnvironName, &environ); err != nil {
		if errors.Cause(err) != dependency.ErrMissing {
			return nil, errors.Trace(err)
		}
		// Only the controller's leader is given an Environ; the
		// other controller units will watch the model and wait
		// for its environ version to be updated.
		environ = nil
	}

	var apiCaller base.APICaller
	if err := getter.Get(config.APICallerName, &apiCaller); err != nil {
		return nil, errors.Trace(err)
	}

	var gate gate.Unlocker
	if err := getter.Get(config.GateName, &gate); err != nil {
		return nil, errors.Trace(err)
	}

	facade, err := config.NewFacade(apiCaller)
	if err != nil {
		return nil, errors.Trace(err)
	}

	credentialAPI, err := config.NewCredentialValidatorFacade(apiCaller)
	if err != nil {
		return nil, errors.Trace(err)
	}

	worker, err := config.NewWorker(ctx, Config{
		Facade:        facade,
		Environ:       environ,
		GateUnlocker:  gate,
		ControllerTag: config.ControllerTag,
		ModelTag:      config.ModelTag,
		CredentialAPI: credentialAPI,
		Logger:        config.Logger,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	return worker, nil
}

// Manifold returns a dependency.Manifold that will run a Worker as
// configured.
func Manifold(config ManifoldConfig) dependency.Manifold {
	return dependency.Manifold{
		Inputs: []string{
			config.APICallerName,
			config.EnvironName,
			config.GateName,
		},
		Start:  config.start,
		Filter: bounceErrChanged,
	}
}

// bounceErrChanged converts ErrModelRemoved to dependency.ErrUninstall.
func bounceErrChanged(err error) error {
	if errors.Cause(err) == ErrModelRemoved {
		return dependency.ErrUninstall
	}
	return err
}
