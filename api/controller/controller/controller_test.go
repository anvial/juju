// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package controller_test

import (
	"context"
	"encoding/json"
	"time"

	"github.com/juju/errors"
	"github.com/juju/names/v6"
	jujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon.v2"

	"github.com/juju/juju/api/base"
	apitesting "github.com/juju/juju/api/base/testing"
	"github.com/juju/juju/api/controller/controller"
	apiservererrors "github.com/juju/juju/apiserver/errors"
	corecontroller "github.com/juju/juju/controller"
	"github.com/juju/juju/core/life"
	"github.com/juju/juju/core/permission"
	environscloudspec "github.com/juju/juju/environs/cloudspec"
	proxyfactory "github.com/juju/juju/internal/proxy/factory"
	coretesting "github.com/juju/juju/internal/testing"
	"github.com/juju/juju/internal/uuid"
	"github.com/juju/juju/rpc/params"
)

type Suite struct {
	jujutesting.IsolationSuite
}

var _ = gc.Suite(&Suite{})

func (s *Suite) TestDestroyController(c *gc.C) {
	var stub jujutesting.Stub
	apiCaller := apitesting.BestVersionCaller{
		BestVersion: 11,
		APICallerFunc: func(objType string, version int, id, request string, arg, result interface{}) error {
			stub.AddCall(objType+"."+request, arg)
			return stub.NextErr()
		},
	}
	client := controller.NewClient(apiCaller)

	destroyStorage := true
	force := true
	maxWait := time.Minute
	timeout := time.Hour
	err := client.DestroyController(context.Background(), controller.DestroyControllerParams{
		DestroyModels:  true,
		DestroyStorage: &destroyStorage,
		Force:          &force,
		MaxWait:        &maxWait,
		ModelTimeout:   &timeout,
	})
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCalls(c, []jujutesting.StubCall{
		{FuncName: "Controller.DestroyController", Args: []interface{}{params.DestroyControllerArgs{
			DestroyModels:  true,
			DestroyStorage: &destroyStorage,
			Force:          &force,
			MaxWait:        &maxWait,
			ModelTimeout:   &timeout,
		}}},
	})
}

func (s *Suite) TestDestroyControllerError(c *gc.C) {
	apiCaller := apitesting.BestVersionCaller{
		BestVersion: 4,
		APICallerFunc: func(objType string, version int, id, request string, arg, result interface{}) error {
			return errors.New("nope")
		},
	}
	client := controller.NewClient(apiCaller)
	err := client.DestroyController(context.Background(), controller.DestroyControllerParams{})
	c.Assert(err, gc.ErrorMatches, "nope")
}

func (s *Suite) TestInitiateMigration(c *gc.C) {
	s.checkInitiateMigration(c, makeSpec())
}

func (s *Suite) TestInitiateMigrationEmptyCACert(c *gc.C) {
	spec := makeSpec()
	spec.TargetCACert = ""
	s.checkInitiateMigration(c, spec)
}

func (s *Suite) checkInitiateMigration(c *gc.C, spec controller.MigrationSpec) {
	client, stub := makeInitiateMigrationClient(params.InitiateMigrationResults{
		Results: []params.InitiateMigrationResult{{
			MigrationId: "id",
		}},
	})
	id, err := client.InitiateMigration(context.Background(), spec)
	c.Assert(err, jc.ErrorIsNil)
	c.Check(id, gc.Equals, "id")
	stub.CheckCalls(c, []jujutesting.StubCall{
		{FuncName: "Controller.InitiateMigration", Args: []interface{}{specToArgs(spec)}},
	})
}

func specToArgs(spec controller.MigrationSpec) params.InitiateMigrationArgs {
	var macsJSON []byte
	if len(spec.TargetMacaroons) > 0 {
		var err error
		macsJSON, err = json.Marshal(spec.TargetMacaroons)
		if err != nil {
			panic(err)
		}
	}
	return params.InitiateMigrationArgs{
		Specs: []params.MigrationSpec{{
			ModelTag: names.NewModelTag(spec.ModelUUID).String(),
			TargetInfo: params.MigrationTargetInfo{
				ControllerTag:   names.NewControllerTag(spec.TargetControllerUUID).String(),
				ControllerAlias: spec.TargetControllerAlias,
				Addrs:           spec.TargetAddrs,
				CACert:          spec.TargetCACert,
				AuthTag:         names.NewUserTag(spec.TargetUser).String(),
				Password:        spec.TargetPassword,
				Macaroons:       string(macsJSON),
			},
		}},
	}
}

func (s *Suite) TestInitiateMigrationError(c *gc.C) {
	client, _ := makeInitiateMigrationClient(params.InitiateMigrationResults{
		Results: []params.InitiateMigrationResult{{
			Error: apiservererrors.ServerError(errors.New("boom")),
		}},
	})
	id, err := client.InitiateMigration(context.Background(), makeSpec())
	c.Check(id, gc.Equals, "")
	c.Check(err, gc.ErrorMatches, "boom")
}

func (s *Suite) TestInitiateMigrationResultMismatch(c *gc.C) {
	client, _ := makeInitiateMigrationClient(params.InitiateMigrationResults{
		Results: []params.InitiateMigrationResult{
			{MigrationId: "id"},
			{MigrationId: "wtf"},
		},
	})
	id, err := client.InitiateMigration(context.Background(), makeSpec())
	c.Check(id, gc.Equals, "")
	c.Check(err, gc.ErrorMatches, "unexpected number of results returned")
}

func (s *Suite) TestInitiateMigrationCallError(c *gc.C) {
	apiCaller := apitesting.APICallerFunc(func(string, int, string, string, interface{}, interface{}) error {
		return errors.New("boom")
	})
	client := controller.NewClient(apiCaller)
	id, err := client.InitiateMigration(context.Background(), makeSpec())
	c.Check(id, gc.Equals, "")
	c.Check(err, gc.ErrorMatches, "boom")
}

func (s *Suite) TestInitiateMigrationValidationError(c *gc.C) {
	client, stub := makeInitiateMigrationClient(params.InitiateMigrationResults{})
	spec := makeSpec()
	spec.ModelUUID = "not-a-uuid"
	id, err := client.InitiateMigration(context.Background(), spec)
	c.Check(id, gc.Equals, "")
	c.Check(err, gc.ErrorMatches, "client-side validation failed: model UUID not valid")
	c.Check(stub.Calls(), gc.HasLen, 0) // API call shouldn't have happened
}

func (s *Suite) TestHostedModelConfigs_CallError(c *gc.C) {
	apiCaller := apitesting.APICallerFunc(func(string, int, string, string, interface{}, interface{}) error {
		return errors.New("boom")
	})
	client := controller.NewClient(apiCaller)
	config, err := client.HostedModelConfigs(context.Background())
	c.Check(config, gc.HasLen, 0)
	c.Check(err, gc.ErrorMatches, "boom")
}

func (s *Suite) TestHostedModelConfigs_FormatResults(c *gc.C) {
	apiCaller := apitesting.APICallerFunc(func(objType string, version int, id, request string, arg, result interface{}) error {
		c.Assert(objType, gc.Equals, "Controller")
		c.Assert(request, gc.Equals, "HostedModelConfigs")
		c.Assert(arg, gc.IsNil)
		out := result.(*params.HostedModelConfigsResults)
		c.Assert(out, gc.NotNil)
		*out = params.HostedModelConfigsResults{
			Models: []params.HostedModelConfig{
				{
					Name:     "first",
					OwnerTag: "user-foo@bar",
					Config: map[string]interface{}{
						"name": "first",
					},
					CloudSpec: &params.CloudSpec{
						Type: "magic",
						Name: "first",
					},
				}, {
					Name:     "second",
					OwnerTag: "bad-tag",
				}, {
					Name:     "third",
					OwnerTag: "user-foo@bar",
					Config: map[string]interface{}{
						"name": "third",
					},
					CloudSpec: &params.CloudSpec{
						Name: "third",
					},
				},
			},
		}
		return nil
	})
	client := controller.NewClient(apiCaller)
	config, err := client.HostedModelConfigs(context.Background())
	c.Assert(config, gc.HasLen, 3)
	c.Assert(err, jc.ErrorIsNil)
	first := config[0]
	c.Assert(first.Name, gc.Equals, "first")
	c.Assert(first.Owner, gc.Equals, names.NewUserTag("foo@bar"))
	c.Assert(first.Config, gc.DeepEquals, map[string]interface{}{
		"name": "first",
	})
	c.Assert(first.CloudSpec, gc.DeepEquals, environscloudspec.CloudSpec{
		Type: "magic",
		Name: "first",
	})
	second := config[1]
	c.Assert(second.Name, gc.Equals, "second")
	c.Assert(second.Error.Error(), gc.Equals, `"bad-tag" is not a valid tag`)
	third := config[2]
	c.Assert(third.Name, gc.Equals, "third")
	c.Assert(third.Error.Error(), gc.Equals, "validating CloudSpec: empty Type not valid")
}

func makeInitiateMigrationClient(results params.InitiateMigrationResults) (
	*controller.Client, *jujutesting.Stub,
) {
	var stub jujutesting.Stub
	apiCaller := apitesting.APICallerFunc(
		func(objType string, version int, id, request string, arg, result interface{}) error {
			stub.AddCall(objType+"."+request, arg)
			out := result.(*params.InitiateMigrationResults)
			*out = results
			return nil
		},
	)
	client := controller.NewClient(apiCaller)
	return client, &stub
}

func makeSpec() controller.MigrationSpec {
	mac, err := macaroon.New([]byte("secret"), []byte("id"), "location", macaroon.LatestVersion)
	if err != nil {
		panic(err)
	}
	return controller.MigrationSpec{
		ModelUUID:             randomUUID(),
		TargetControllerUUID:  randomUUID(),
		TargetControllerAlias: "target-controller",
		TargetAddrs:           []string{"1.2.3.4:5"},
		TargetCACert:          "cert",
		TargetUser:            "someone",
		TargetPassword:        "secret",
		TargetMacaroons:       []macaroon.Slice{{mac}},
	}
}

func randomUUID() string {
	return uuid.MustNewUUID().String()
}

func (s *Suite) TestModelStatusEmpty(c *gc.C) {
	apiCaller := apitesting.APICallerFunc(func(objType string, version int, id, request string, arg, result interface{}) error {
		c.Check(objType, gc.Equals, "Controller")
		c.Check(id, gc.Equals, "")
		c.Check(request, gc.Equals, "ModelStatus")
		c.Check(result, gc.FitsTypeOf, &params.ModelStatusResults{})

		return nil
	})

	client := controller.NewClient(apiCaller)
	results, err := client.ModelStatus(context.Background())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, jc.DeepEquals, []base.ModelStatus{})
}

func (s *Suite) TestModelStatus(c *gc.C) {
	apiCaller := apitesting.BestVersionCaller{
		BestVersion: 4,
		APICallerFunc: func(objType string, version int, id, request string, arg, result interface{}) error {
			c.Check(objType, gc.Equals, "Controller")
			c.Check(id, gc.Equals, "")
			c.Check(request, gc.Equals, "ModelStatus")
			c.Check(arg, jc.DeepEquals, params.Entities{
				Entities: []params.Entity{
					{Tag: coretesting.ModelTag.String()},
					{Tag: coretesting.ModelTag.String()},
				},
			})
			c.Check(result, gc.FitsTypeOf, &params.ModelStatusResults{})

			out := result.(*params.ModelStatusResults)
			out.Results = []params.ModelStatus{
				{
					ModelTag:           coretesting.ModelTag.String(),
					OwnerTag:           "user-glenda",
					ApplicationCount:   3,
					HostedMachineCount: 2,
					Life:               "alive",
					Machines: []params.ModelMachineInfo{{
						Id:         "0",
						InstanceId: "inst-ance",
						Status:     "pending",
					}},
				},
				{Error: apiservererrors.ServerError(errors.New("model error"))},
			}
			return nil
		},
	}

	client := controller.NewClient(apiCaller)
	results, err := client.ModelStatus(context.Background(), coretesting.ModelTag, coretesting.ModelTag)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results[0], jc.DeepEquals, base.ModelStatus{
		UUID:               coretesting.ModelTag.Id(),
		TotalMachineCount:  1,
		HostedMachineCount: 2,
		ApplicationCount:   3,
		Owner:              "glenda",
		Life:               life.Alive,
		Machines:           []base.Machine{{Id: "0", InstanceId: "inst-ance", Status: "pending"}},
	})
	c.Assert(results[1].Error, gc.ErrorMatches, "model error")
}

func (s *Suite) TestModelStatusError(c *gc.C) {
	apiCaller := apitesting.APICallerFunc(
		func(objType string, version int, id, request string, args, result interface{}) error {
			return errors.New("model error")
		})
	client := controller.NewClient(apiCaller)
	out, err := client.ModelStatus(context.Background(), coretesting.ModelTag, coretesting.ModelTag)
	c.Assert(err, gc.ErrorMatches, "model error")
	c.Assert(out, gc.IsNil)
}

func (s *Suite) TestConfigSet(c *gc.C) {
	apiCaller := apitesting.BestVersionCaller{
		BestVersion: 5,
		APICallerFunc: func(objType string, version int, id, request string, args, result interface{}) error {
			c.Assert(objType, gc.Equals, "Controller")
			c.Assert(version, gc.Equals, 5)
			c.Assert(request, gc.Equals, "ConfigSet")
			c.Assert(result, gc.IsNil)
			c.Assert(args, gc.DeepEquals, params.ControllerConfigSet{Config: map[string]interface{}{
				"some-setting": 345,
			}})
			return errors.New("ruth mundy")
		},
	}
	client := controller.NewClient(apiCaller)
	err := client.ConfigSet(context.Background(), map[string]interface{}{
		"some-setting": 345,
	})
	c.Assert(err, gc.ErrorMatches, "ruth mundy")
}

func (s *Suite) TestWatchModelSummaries(c *gc.C) {
	apiCaller := apitesting.BestVersionCaller{
		BestVersion: 9,
		APICallerFunc: func(objType string, version int, id, request string, args, result interface{}) error {
			c.Check(objType, gc.Equals, "Controller")
			c.Check(version, gc.Equals, 9)
			c.Check(request, gc.Equals, "WatchModelSummaries")
			c.Check(result, gc.FitsTypeOf, &params.SummaryWatcherID{})
			c.Check(args, gc.IsNil)
			return errors.New("some error")
		},
	}
	client := controller.NewClient(apiCaller)
	watcher, err := client.WatchModelSummaries(context.Background())
	c.Assert(err, gc.ErrorMatches, "some error")
	c.Assert(watcher, gc.IsNil)
}

func (s *Suite) TestWatchAllModelSummaries(c *gc.C) {
	apiCaller := apitesting.BestVersionCaller{
		BestVersion: 9,
		APICallerFunc: func(objType string, version int, id, request string, args, result interface{}) error {
			c.Check(objType, gc.Equals, "Controller")
			c.Check(version, gc.Equals, 9)
			c.Check(request, gc.Equals, "WatchAllModelSummaries")
			c.Check(result, gc.FitsTypeOf, &params.SummaryWatcherID{})
			c.Check(args, gc.IsNil)
			return errors.New("some error")
		},
	}
	client := controller.NewClient(apiCaller)
	watcher, err := client.WatchAllModelSummaries(context.Background())
	c.Assert(err, gc.ErrorMatches, "some error")
	c.Assert(watcher, gc.IsNil)
}

func (s *Suite) TestDashboardConnectionInfo(c *gc.C) {
	apiCaller := apitesting.APICallerFunc(
		func(objType string, version int, id, request string, args, result interface{}) error {
			c.Assert(objType, gc.Equals, "Controller")
			c.Assert(request, gc.Equals, "DashboardConnectionInfo")
			c.Assert(args, gc.IsNil)
			c.Assert(result, gc.FitsTypeOf, &params.DashboardConnectionInfo{})
			*(result.(*params.DashboardConnectionInfo)) = params.DashboardConnectionInfo{
				SSHConnection: &params.DashboardConnectionSSHTunnel{
					Model:  "c:controller",
					Entity: "dashboard/leader",
					Host:   "10.1.1.1",
					Port:   "1234",
				},
			}
			return nil
		})
	client := controller.NewClient(apiCaller)
	connectionInfo, err := client.DashboardConnectionInfo(context.Background(), proxyfactory.NewFactory())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(connectionInfo.SSHTunnel, gc.NotNil)
}

func (s *Suite) TestAllModels(c *gc.C) {
	now := time.Now()
	apiCaller := apitesting.APICallerFunc(func(objType string, version int, id, request string, args, result interface{}) error {
		c.Check(objType, gc.Equals, "Controller")
		c.Check(id, gc.Equals, "")
		c.Check(request, gc.Equals, "AllModels")
		c.Check(args, gc.IsNil)
		c.Check(result, gc.FitsTypeOf, &params.UserModelList{})
		*(result.(*params.UserModelList)) = params.UserModelList{
			UserModels: []params.UserModel{{
				Model: params.Model{
					Name:     "test",
					UUID:     coretesting.ModelTag.Id(),
					Type:     "iaas",
					OwnerTag: "user-fred",
				},
				LastConnection: &now,
			}},
		}
		return nil
	})

	client := controller.NewClient(apiCaller)
	m, err := client.AllModels(context.Background())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(m, jc.DeepEquals, []base.UserModel{{
		Name:           "test",
		UUID:           coretesting.ModelTag.Id(),
		Type:           "iaas",
		Owner:          "fred",
		LastConnection: &now,
	}})
}

func (s *Suite) TestControllerConfig(c *gc.C) {
	apiCaller := apitesting.APICallerFunc(func(objType string, version int, id, request string, args, result interface{}) error {
		c.Check(objType, gc.Equals, "Controller")
		c.Check(id, gc.Equals, "")
		c.Check(request, gc.Equals, "ControllerConfig")
		c.Check(args, gc.IsNil)
		c.Check(result, gc.FitsTypeOf, &params.ControllerConfigResult{})
		*(result.(*params.ControllerConfigResult)) = params.ControllerConfigResult{
			Config: map[string]interface{}{
				"api-port": 666,
			},
		}
		return nil
	})

	client := controller.NewClient(apiCaller)
	m, err := client.ControllerConfig(context.Background())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(m, jc.DeepEquals, corecontroller.Config{"api-port": 666})
}

func (s *Suite) TestListBlockedModels(c *gc.C) {
	apiCaller := apitesting.APICallerFunc(func(objType string, version int, id, request string, args, result interface{}) error {
		c.Check(objType, gc.Equals, "Controller")
		c.Check(id, gc.Equals, "")
		c.Check(request, gc.Equals, "ListBlockedModels")
		c.Check(args, gc.IsNil)
		c.Check(result, gc.FitsTypeOf, &params.ModelBlockInfoList{})
		*(result.(*params.ModelBlockInfoList)) = params.ModelBlockInfoList{
			Models: []params.ModelBlockInfo{{
				Name:     "controller",
				UUID:     coretesting.ModelTag.Id(),
				OwnerTag: "user-fred",
				Blocks: []string{
					"BlockChange",
					"BlockDestroy",
				},
			}},
		}
		return nil
	})

	client := controller.NewClient(apiCaller)
	results, err := client.ListBlockedModels(context.Background())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, jc.DeepEquals, []params.ModelBlockInfo{
		{
			Name:     "controller",
			UUID:     coretesting.ModelTag.Id(),
			OwnerTag: "user-fred",
			Blocks: []string{
				"BlockChange",
				"BlockDestroy",
			},
		},
	})
}

func (s *Suite) TestRemoveBlocks(c *gc.C) {
	apiCaller := apitesting.APICallerFunc(func(objType string, version int, id, request string, args, result interface{}) error {
		c.Check(objType, gc.Equals, "Controller")
		c.Check(id, gc.Equals, "")
		c.Check(request, gc.Equals, "RemoveBlocks")
		c.Check(args, jc.DeepEquals, params.RemoveBlocksArgs{All: true})
		c.Check(result, gc.IsNil)
		return errors.New("some error")
	})

	client := controller.NewClient(apiCaller)
	err := client.RemoveBlocks(context.Background())
	c.Assert(err, gc.ErrorMatches, "some error")
}

func (s *Suite) TestGetControllerAccess(c *gc.C) {
	apiCaller := apitesting.APICallerFunc(func(objType string, version int, id, request string, args, result interface{}) error {
		c.Check(objType, gc.Equals, "Controller")
		c.Check(id, gc.Equals, "")
		c.Check(request, gc.Equals, "GetControllerAccess")
		c.Check(args, jc.DeepEquals, params.Entities{
			Entities: []params.Entity{{Tag: "user-fred"}},
		})
		c.Check(result, gc.FitsTypeOf, &params.UserAccessResults{})
		*(result.(*params.UserAccessResults)) = params.UserAccessResults{
			Results: []params.UserAccessResult{{
				Result: &params.UserAccess{
					UserTag: "user-fred",
					Access:  "superuser",
				},
			}},
		}
		return nil
	})

	client := controller.NewClient(apiCaller)
	access, err := client.GetControllerAccess(context.Background(), "fred")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(access, gc.Equals, permission.SuperuserAccess)
}
