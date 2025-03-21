// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package usermanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/juju/errors"
	"github.com/juju/names/v6"

	"github.com/juju/juju/api/base"
	internallogger "github.com/juju/juju/internal/logger"
	"github.com/juju/juju/rpc/params"
)

// Option is a function that can be used to configure a Client.
type Option = base.Option

// WithTracer returns an Option that configures the Client to use the
// supplied tracer.
var WithTracer = base.WithTracer

var logger = internallogger.GetLogger("juju.api.usermanager")

// Client provides methods that the Juju client command uses to interact
// with users stored in the Juju Server.
type Client struct {
	base.ClientFacade
	facade base.FacadeCaller
}

// NewClient creates a new `Client` based on an existing authenticated API
// connection.
func NewClient(st base.APICallCloser, options ...Option) *Client {
	frontend, backend := base.NewClientFacade(st, "UserManager", options...)
	return &Client{ClientFacade: frontend, facade: backend}
}

// AddUser creates a new local user in the controller, sharing with that user any specified models.
func (c *Client) AddUser(
	ctx context.Context,
	username, displayName, password string,
) (_ names.UserTag, secretKey []byte, _ error) {
	if !names.IsValidUser(username) {
		return names.UserTag{}, nil, fmt.Errorf("invalid user name %q", username)
	}

	userArgs := params.AddUsers{
		Users: []params.AddUser{{
			Username:    username,
			DisplayName: displayName,
			Password:    password,
		}},
	}
	var results params.AddUserResults
	err := c.facade.FacadeCall(ctx, "AddUser", userArgs, &results)
	if err != nil {
		return names.UserTag{}, nil, errors.Trace(err)
	}
	if count := len(results.Results); count != 1 {
		logger.Errorf(context.TODO(), "expected 1 result, got %#v", results)
		return names.UserTag{}, nil, errors.Errorf("expected 1 result, got %d", count)
	}
	result := results.Results[0]
	if result.Error != nil {
		return names.UserTag{}, nil, errors.Trace(result.Error)
	}
	tag, err := names.ParseUserTag(result.Tag)
	if err != nil {
		return names.UserTag{}, nil, errors.Trace(err)
	}
	return tag, result.SecretKey, nil
}

func (c *Client) userCall(ctx context.Context, username string, methodCall string) error {
	if !names.IsValidUser(username) {
		return errors.Errorf("%q is not a valid username", username)
	}
	tag := names.NewUserTag(username)

	var results params.ErrorResults
	args := params.Entities{
		[]params.Entity{{tag.String()}},
	}
	err := c.facade.FacadeCall(ctx, methodCall, args, &results)
	if err != nil {
		return errors.Trace(err)
	}
	return results.OneError()
}

// DisableUser disables a user.  If the user is already disabled, the action
// is considered a success.
func (c *Client) DisableUser(ctx context.Context, username string) error {
	return c.userCall(ctx, username, "DisableUser")
}

// EnableUser enables a users.  If the user is already enabled, the action is
// considered a success.
func (c *Client) EnableUser(ctx context.Context, username string) error {
	return c.userCall(ctx, username, "EnableUser")
}

// RemoveUser deletes a user. That is it permanently removes the user, while
// retaining the record of the user to maintain provenance.
func (c *Client) RemoveUser(ctx context.Context, username string) error {
	return c.userCall(ctx, username, "RemoveUser")
}

// IncludeDisabled is a type alias to avoid bare true/false values
// in calls to the client method.
type IncludeDisabled bool

var (
	// ActiveUsers indicates to only return active users.
	ActiveUsers IncludeDisabled = false
	// AllUsers indicates that both enabled and disabled users should be
	// returned.
	AllUsers IncludeDisabled = true
)

// UserInfo returns information about the specified users.  If no users are
// specified, the call should return all users.  If includeDisabled is set to
// ActiveUsers, only enabled users are returned.
func (c *Client) UserInfo(ctx context.Context, usernames []string, all IncludeDisabled) ([]params.UserInfo, error) {
	var results params.UserInfoResults
	var entities []params.Entity
	for _, username := range usernames {
		if !names.IsValidUser(username) {
			return nil, errors.Errorf("%q is not a valid username", username)
		}
		tag := names.NewUserTag(username)
		entities = append(entities, params.Entity{Tag: tag.String()})
	}
	args := params.UserInfoRequest{
		Entities:        entities,
		IncludeDisabled: bool(all),
	}
	err := c.facade.FacadeCall(ctx, "UserInfo", args, &results)
	if err != nil {
		return nil, errors.Trace(err)
	}
	// Only need to look for errors if users were explicitly specified, because
	// if we didn't ask for any, we should get all, and we shouldn't get any
	// errors for listing all.  We care here because we index into the users
	// slice.
	if len(results.Results) == len(usernames) {
		var errorStrings []string
		for i, result := range results.Results {
			if result.Error != nil {
				annotated := errors.Annotate(result.Error, usernames[i])
				errorStrings = append(errorStrings, annotated.Error())
			}
		}
		// TODO(wallyworld) - we should return these errors to the caller so that any
		// users which are successfully found can be handled.
		if len(errorStrings) > 0 {
			return nil, errors.New(strings.Join(errorStrings, ", "))
		}
	}
	info := []params.UserInfo{}
	for i, result := range results.Results {
		if result.Result == nil {
			return nil, errors.Errorf("unexpected nil result at position %d", i)
		}
		info = append(info, *result.Result)
	}
	return info, nil
}

// ModelUserInfo returns information on all users in the model.
func (c *Client) ModelUserInfo(ctx context.Context, modelUUID string) ([]params.ModelUserInfo, error) {
	var results params.ModelUserInfoResults
	args := params.Entities{
		Entities: []params.Entity{{
			Tag: names.NewModelTag(modelUUID).String(),
		}},
	}
	err := c.facade.FacadeCall(ctx, "ModelUserInfo", args, &results)
	if err != nil {
		return nil, errors.Trace(err)
	}

	info := []params.ModelUserInfo{}
	for i, result := range results.Results {
		if result.Result == nil {
			return nil, errors.Errorf("unexpected nil result at position %d", i)
		}
		info = append(info, *result.Result)
	}
	return info, nil
}

// SetPassword changes the password for the specified user.
func (c *Client) SetPassword(ctx context.Context, username, password string) error {
	if !names.IsValidUser(username) {
		return errors.Errorf("%q is not a valid username", username)
	}
	tag := names.NewUserTag(username)
	args := params.EntityPasswords{
		Changes: []params.EntityPassword{{
			Tag:      tag.String(),
			Password: password}},
	}
	var results params.ErrorResults
	err := c.facade.FacadeCall(ctx, "SetPassword", args, &results)
	if err != nil {
		return err
	}
	return results.OneError()
}

// ResetPassword resets password for the specified user.
func (c *Client) ResetPassword(ctx context.Context, username string) ([]byte, error) {
	if !names.IsValidUser(username) {
		return nil, fmt.Errorf("invalid user name %q", username)
	}

	in := params.Entities{
		Entities: []params.Entity{{
			Tag: names.NewUserTag(username).String(),
		}},
	}
	var out params.AddUserResults
	err := c.facade.FacadeCall(ctx, "ResetPassword", in, &out)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if count := len(out.Results); count != 1 {
		logger.Errorf(context.TODO(), "expected 1 result, got %#v", out)
		return nil, errors.Errorf("expected 1 result, got %d", count)
	}
	result := out.Results[0]
	if result.Error != nil {
		return nil, errors.Trace(result.Error)
	}
	return result.SecretKey, nil
}
