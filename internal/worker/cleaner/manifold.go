// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package cleaner

import (
	"context"

	"github.com/juju/clock"
	"github.com/juju/errors"
	"github.com/juju/worker/v4"
	"github.com/juju/worker/v4/dependency"

	"github.com/juju/juju/api/base"
	"github.com/juju/juju/api/controller/cleaner"
	"github.com/juju/juju/core/logger"
)

// ManifoldConfig describes the resources used by the cleanup worker.
type ManifoldConfig struct {
	APICallerName string
	Clock         clock.Clock
	Logger        logger.Logger
}

// Validate is called by start to check for bad configuration.
func (config ManifoldConfig) Validate() error {
	if config.APICallerName == "" {
		return errors.NotValidf("empty APICallerName")
	}
	if config.Clock == nil {
		return errors.NotValidf("nil Clock")
	}
	if config.Logger == nil {
		return errors.NotValidf("nil Logger")
	}
	return nil
}

// Manifold returns a Manifold that encapsulates the cleanup worker.
func Manifold(config ManifoldConfig) dependency.Manifold {
	return dependency.Manifold{
		Inputs: []string{config.APICallerName},
		Start:  config.start,
	}
}

// start is a StartFunc for a Worker manifold.
func (config ManifoldConfig) start(ctx context.Context, getter dependency.Getter) (worker.Worker, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Trace(err)
	}
	var apiCaller base.APICaller
	if err := getter.Get(config.APICallerName, &apiCaller); err != nil {
		return nil, errors.Trace(err)
	}
	api := cleaner.NewAPI(apiCaller)
	w, err := NewCleaner(ctx, api, config.Clock, config.Logger)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return w, nil
}
