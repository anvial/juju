// Copyright 2020 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package uniter

import (
	"context"

	"github.com/juju/errors"
	"github.com/juju/names/v6"

	"github.com/juju/juju/apiserver/common"
	apiservererrors "github.com/juju/juju/apiserver/errors"
	"github.com/juju/juju/apiserver/facade"
	"github.com/juju/juju/apiserver/internal"
	"github.com/juju/juju/core/instance"
	corelogger "github.com/juju/juju/core/logger"
	"github.com/juju/juju/core/lxdprofile"
	coremachine "github.com/juju/juju/core/machine"
	"github.com/juju/juju/core/unit"
	machineerrors "github.com/juju/juju/domain/machine/errors"
	"github.com/juju/juju/rpc/params"
)

// LXDProfileMachine describes machine-receiver state methods
// for executing a lxd profile upgrade.
type LXDProfileMachine interface {
	ContainerType() instance.ContainerType
}

type LXDProfileAPI struct {
	machineService  MachineService
	watcherRegistry facade.WatcherRegistry

	logger     corelogger.Logger
	accessUnit common.GetAuthFunc

	modelInfoService   ModelInfoService
	applicationService ApplicationService
}

// NewLXDProfileAPI returns a new LXDProfileAPI. Currently both
// GetAuthFuncs can used to determine current permissions.
func NewLXDProfileAPI(
	machineService MachineService,
	watcherRegistry facade.WatcherRegistry,
	authorizer facade.Authorizer,
	accessUnit common.GetAuthFunc,
	logger corelogger.Logger,
	modelInfoService ModelInfoService,
	applicationService ApplicationService,
) *LXDProfileAPI {
	return &LXDProfileAPI{
		machineService:     machineService,
		watcherRegistry:    watcherRegistry,
		accessUnit:         accessUnit,
		logger:             logger,
		modelInfoService:   modelInfoService,
		applicationService: applicationService,
	}
}

// NewExternalLXDProfileAPI can be used for API registration.
func NewExternalLXDProfileAPI(
	machineService MachineService,
	watcherRegistry facade.WatcherRegistry,
	authorizer facade.Authorizer,
	accessUnit common.GetAuthFunc,
	logger corelogger.Logger,
	modelInfoService ModelInfoService,
	applicationService ApplicationService,
) *LXDProfileAPI {
	return NewLXDProfileAPI(
		machineService,
		watcherRegistry,
		authorizer,
		accessUnit,
		logger,
		modelInfoService,
		applicationService,
	)
}

// WatchInstanceData returns a NotifyWatcher for observing
// changes to the lxd profile for one unit.
func (u *LXDProfileAPI) WatchInstanceData(ctx context.Context, args params.Entities) (params.NotifyWatchResults, error) {
	u.logger.Tracef(ctx, "Starting WatchInstanceData with %+v", args)
	result := params.NotifyWatchResults{
		Results: make([]params.NotifyWatchResult, len(args.Entities)),
	}
	canAccess, err := u.accessUnit(ctx)
	if err != nil {
		u.logger.Tracef(ctx, "WatchInstanceData error %+v", err)
		return params.NotifyWatchResults{}, err
	}
	for i, entity := range args.Entities {
		tag, err := names.ParseTag(entity.Tag)
		if err != nil {
			result.Results[i].Error = apiservererrors.ServerError(apiservererrors.ErrPerm)
			continue
		}
		if !canAccess(tag) {
			result.Results[i].Error = apiservererrors.ServerError(apiservererrors.ErrPerm)
			continue
		}
		unitName, err := unit.NewName(tag.Id())
		if err != nil {
			result.Results[i].Error = apiservererrors.ServerError(err)
			continue
		}
		// TODO(nvinuesa): we could save this call if we move the lxd profile
		// watcher to the unit domain. Then, the watcher would be already
		// notifying for changes on the unit directly.
		machineUUID, err := u.applicationService.GetUnitMachineUUID(ctx, unitName)
		if err != nil {
			result.Results[i].Error = apiservererrors.ServerError(err)
			continue
		}
		watcherId, err := u.watchOneInstanceData(ctx, machineUUID)
		if err != nil {
			result.Results[i].Error = apiservererrors.ServerError(err)
			continue
		}

		result.Results[i].NotifyWatcherId = watcherId

	}
	u.logger.Tracef(ctx, "WatchInstanceData returning %+v", result)
	return result, nil
}

func (u *LXDProfileAPI) watchOneInstanceData(ctx context.Context, machineUUID coremachine.UUID) (string, error) {
	watcher, err := u.machineService.WatchLXDProfiles(ctx, machineUUID)
	if err != nil {
		return "", errors.Trace(err)
	}
	watcherID, _, err := internal.EnsureRegisterWatcher[struct{}](ctx, u.watcherRegistry, watcher)
	return watcherID, err
}

// LXDProfileName returns the name of the lxd profile applied to the unit's
// machine for the current charm version.
func (u *LXDProfileAPI) LXDProfileName(ctx context.Context, args params.Entities) (params.StringResults, error) {
	u.logger.Tracef(ctx, "Starting LXDProfileName with %+v", args)
	result := params.StringResults{
		Results: make([]params.StringResult, len(args.Entities)),
	}
	canAccess, err := u.accessUnit(ctx)
	if err != nil {
		return params.StringResults{}, err
	}
	for i, entity := range args.Entities {
		tag, err := names.ParseTag(entity.Tag)
		if err != nil {
			result.Results[i].Error = apiservererrors.ServerError(apiservererrors.ErrPerm)
			continue
		}

		if !canAccess(tag) {
			result.Results[i].Error = apiservererrors.ServerError(apiservererrors.ErrPerm)
			continue
		}
		unitName, err := unit.NewName(tag.Id())
		if err != nil {
			result.Results[i].Error = apiservererrors.ServerError(err)
			continue
		}
		machineUUID, err := u.applicationService.GetUnitMachineUUID(ctx, unitName)
		if err != nil {
			result.Results[i].Error = apiservererrors.ServerError(err)
			continue
		}
		name, err := u.getOneLXDProfileName(ctx, unitName.Application(), machineUUID)
		if errors.Is(err, machineerrors.NotProvisioned) {
			result.Results[i].Error = apiservererrors.ServerError(errors.NotProvisionedf("machine %q", machineUUID))
		} else if err != nil {
			result.Results[i].Error = apiservererrors.ServerError(err)
			continue
		}

		result.Results[i].Result = name

	}
	return result, nil
}

func (u *LXDProfileAPI) getOneLXDProfileName(ctx context.Context, appName string, machineUUID coremachine.UUID) (string, error) {
	profileNames, err := u.machineService.AppliedLXDProfileNames(ctx, machineUUID)
	if err != nil {
		u.logger.Errorf(ctx, "unable to retrieve LXD profiles for machine %q: %v", machineUUID, err)
		return "", err
	}
	return lxdprofile.MatchProfileNameByAppName(profileNames, appName)
}

// CanApplyLXDProfile returns false results. LXD Profiles are not supported.
func (u *LXDProfileAPI) CanApplyLXDProfile(ctx context.Context, args params.Entities) (params.BoolResults, error) {
	result := params.BoolResults{
		Results: make([]params.BoolResult, len(args.Entities)),
	}
	return result, nil
}

// LXDProfileRequired returns false results. LXD profiles are not supported.
func (u *LXDProfileAPI) LXDProfileRequired(ctx context.Context, args params.CharmURLs) (params.BoolResults, error) {
	result := params.BoolResults{
		Results: make([]params.BoolResult, len(args.URLs)),
	}
	return result, nil
}