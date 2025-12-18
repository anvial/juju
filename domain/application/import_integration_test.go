// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package application_test

import (
	context "context"
	"testing"

	"github.com/juju/clock"
	"github.com/juju/description/v11"
	"github.com/juju/tc"

	"github.com/juju/juju/core/database"
	"github.com/juju/juju/core/model"
	"github.com/juju/juju/core/modelmigration"
	"github.com/juju/juju/core/semversion"
	"github.com/juju/juju/domain"
	"github.com/juju/juju/domain/application/charm"
	applicationmodelmigration "github.com/juju/juju/domain/application/modelmigration"
	"github.com/juju/juju/domain/application/service"
	"github.com/juju/juju/domain/application/state"
	schematesting "github.com/juju/juju/domain/schema/testing"
	domaintesting "github.com/juju/juju/domain/testing"
	internalcharm "github.com/juju/juju/internal/charm"
	"github.com/juju/juju/internal/charm/assumes"
	"github.com/juju/juju/internal/charm/resource"
	loggertesting "github.com/juju/juju/internal/logger/testing"
)

type importSuite struct {
	schematesting.ModelSuite
}

func TestImportSuite(t *testing.T) {
	tc.Run(t, &importSuite{})
}

func (s *importSuite) TestImportMaximalCharmMetadata(c *tc.C) {
	desc := description.NewModel(description.ModelArgs{
		Type: string(model.IAAS),
	})

	// Create a charm description and write it into the model, then import
	// it using the model migration framework. Verify that the charm has been
	// imported correctly into the database.

	// This skips both Payloads and LXD profiles, as it's not longer used, so
	// can be skipped.

	app := desc.AddApplication(description.ApplicationArgs{
		Name:     "foo",
		CharmURL: "ch:foo-1",
	})
	app.SetCharmOrigin(description.CharmOriginArgs{
		Source:   "charm-hub",
		ID:       "deadbeef",
		Hash:     "deadbeef2",
		Revision: 1,
		Channel:  "latest/stable",
		Platform: "amd64/ubuntu/20.04",
	})
	app.SetCharmMetadata(description.CharmMetadataArgs{
		Name:           "foo",
		Summary:        "summary foo",
		Description:    "description foo",
		Subordinate:    false,
		MinJujuVersion: "4.0.0",
		RunAs:          "root",
		Assumes:        "[]",
		Categories:     []string{"bar", "baz"},
		Tags:           []string{"alpha", "beta"},
		Terms:          []string{"terms", "and", "conditions"},
		Provides: map[string]description.CharmMetadataRelation{
			"db": relation{
				name:          "db",
				role:          "provider",
				interfaceName: "db",
				optional:      true,
				limit:         1,
				scope:         "global",
			},
		},
		Peers: map[string]description.CharmMetadataRelation{
			"restart": relation{
				name:          "restart",
				role:          "peer",
				interfaceName: "restarter",
				optional:      true,
				limit:         2,
				scope:         "global",
			},
		},
		Requires: map[string]description.CharmMetadataRelation{
			"cache": relation{
				name:          "cache",
				role:          "requirer",
				interfaceName: "cache",
				optional:      true,
				limit:         3,
				scope:         "container",
			},
		},
		Storage: map[string]description.CharmMetadataStorage{
			"ebs": storage{
				name:        "ebs",
				description: "ebs storage",
				shared:      false,
				readonly:    true,
				countMin:    1,
				countMax:    1,
				minimumSize: 10,
				location:    "/ebs",
				properties:  []string{"fast", "encrypted"},
				stype:       "filesystem",
			},
		},
		ExtraBindings: map[string]string{
			"db-admin": "db-admin",
		},
		Devices: map[string]description.CharmMetadataDevice{
			"gpu": device{
				name:        "gpu",
				description: "A GPU device",
				dtype:       "gpu",
				countMin:    1,
				countMax:    2,
			},
		},
		Containers: map[string]description.CharmMetadataContainer{
			"deadbeef": container{
				resource: "deadbeef",
				mounts: []description.CharmMetadataContainerMount{
					containerMount{
						storage:  "tmpfs",
						location: "/tmp",
					},
				},
				uid: ptr(1000),
				gid: ptr(1000),
			},
		},
		Resources: map[string]description.CharmMetadataResource{
			"file1": resourceMeta{
				name:        "file1",
				rtype:       "file",
				description: "A resource file",
				path:        "resources/deadbeef1",
			},
			"oci2": resourceMeta{
				name:        "oci2",
				rtype:       "oci-image",
				description: "A resource oci image",
				path:        "resources/deadbeef2",
			},
		},
	})
	app.SetCharmManifest(description.CharmManifestArgs{
		Bases: []description.CharmManifestBase{
			manifestBase{
				name:          "ubuntu",
				channel:       "stable",
				architectures: []string{"amd64"},
			},
		},
	})

	coordinator := modelmigration.NewCoordinator(loggertesting.WrapCheckLog(c))
	applicationmodelmigration.RegisterImport(coordinator, clock.WallClock, loggertesting.WrapCheckLog(c))
	err := coordinator.Perform(c.Context(), modelmigration.NewScope(nil, s.TxnRunnerFactory(), nil, model.UUID(s.ModelUUID())), desc)
	c.Assert(err, tc.ErrorIsNil)

	svc := s.setupService(c)
	metadata, err := svc.GetCharmMetadata(c.Context(), charm.CharmLocator{
		Name:     "foo",
		Revision: 1,
		Source:   charm.CharmHubSource,
	})
	c.Assert(err, tc.ErrorIsNil)

	c.Check(metadata, tc.DeepEquals, internalcharm.Meta{
		Name:           "foo",
		Summary:        "summary foo",
		Description:    "description foo",
		Subordinate:    false,
		MinJujuVersion: semversion.MustParse("4.0.0"),
		CharmUser:      internalcharm.RunAsRoot,
		Assumes: &assumes.ExpressionTree{
			Expression: assumes.CompositeExpression{
				ExprType:       assumes.AllOfExpression,
				SubExpressions: []assumes.Expression{},
			},
		},
		Categories: []string{"bar", "baz"},
		Tags:       []string{"alpha", "beta"},
		Terms:      []string{"terms", "and", "conditions"},
		Provides: map[string]internalcharm.Relation{
			"db": {
				Name:      "db",
				Role:      internalcharm.RoleProvider,
				Interface: "db",
				Optional:  true,
				Limit:     1,
				Scope:     "global",
			},
			"juju-info": {
				Name:      "juju-info",
				Role:      internalcharm.RoleProvider,
				Interface: "juju-info",
				Scope:     "global",
			},
		},
		Peers: map[string]internalcharm.Relation{
			"restart": {
				Name:      "restart",
				Role:      internalcharm.RolePeer,
				Interface: "restarter",
				Optional:  true,
				Limit:     2,
				Scope:     "global",
			},
		},
		Requires: map[string]internalcharm.Relation{
			"cache": {
				Name:      "cache",
				Role:      internalcharm.RoleRequirer,
				Interface: "cache",
				Optional:  true,
				Limit:     3,
				Scope:     "container",
			},
		},
		Storage: map[string]internalcharm.Storage{
			"ebs": {
				Name:        "ebs",
				Type:        internalcharm.StorageFilesystem,
				Description: "ebs storage",
				Shared:      false,
				ReadOnly:    true,
				CountMin:    1,
				CountMax:    1,
				MinimumSize: 10,
				Location:    "/ebs",
				Properties:  []string{"fast", "encrypted"},
			},
		},
		ExtraBindings: map[string]internalcharm.ExtraBinding{
			"db-admin": {Name: "db-admin"},
		},
		Devices: map[string]internalcharm.Device{
			"gpu": {
				Name:        "gpu",
				Description: "A GPU device",
				Type:        "gpu",
				CountMin:    1,
				CountMax:    2,
			},
		},
		Containers: map[string]internalcharm.Container{
			"deadbeef": {
				Resource: "deadbeef",
				Mounts: []internalcharm.Mount{
					{
						Storage:  "tmpfs",
						Location: "/tmp",
					},
				},
				Uid: ptr(1000),
				Gid: ptr(1000),
			},
		},
		Resources: map[string]resource.Meta{
			"file1": {
				Name:        "file1",
				Type:        resource.TypeFile,
				Description: "A resource file",
				Path:        "resources/deadbeef1",
			},
			"oci2": {
				Name:        "oci2",
				Type:        resource.TypeContainerImage,
				Description: "A resource oci image",
				Path:        "resources/deadbeef2",
			},
		},
	})
}

func (s *importSuite) TestImportMinimalCharmMetadata(c *tc.C) {
	desc := description.NewModel(description.ModelArgs{
		Type: string(model.IAAS),
	})

	app := desc.AddApplication(description.ApplicationArgs{
		Name:     "foo",
		CharmURL: "ch:foo-1",
	})
	app.SetCharmOrigin(description.CharmOriginArgs{
		Source:   "charm-hub",
		ID:       "deadbeef",
		Hash:     "deadbeef2",
		Revision: 1,
		Channel:  "latest/stable",
		Platform: "amd64/ubuntu/20.04",
	})
	app.SetCharmMetadata(description.CharmMetadataArgs{
		Name: "foo",
	})
	app.SetCharmManifest(description.CharmManifestArgs{
		Bases: []description.CharmManifestBase{
			manifestBase{
				name:          "ubuntu",
				channel:       "stable",
				architectures: []string{"amd64"},
			},
		},
	})

	coordinator := modelmigration.NewCoordinator(loggertesting.WrapCheckLog(c))
	applicationmodelmigration.RegisterImport(coordinator, clock.WallClock, loggertesting.WrapCheckLog(c))
	err := coordinator.Perform(c.Context(), modelmigration.NewScope(nil, s.TxnRunnerFactory(), nil, model.UUID(s.ModelUUID())), desc)
	c.Assert(err, tc.ErrorIsNil)

	svc := s.setupService(c)
	metadata, err := svc.GetCharmMetadata(c.Context(), charm.CharmLocator{
		Name:     "foo",
		Revision: 1,
		Source:   charm.CharmHubSource,
	})
	c.Assert(err, tc.ErrorIsNil)

	c.Check(metadata, tc.DeepEquals, internalcharm.Meta{
		Name: "foo",
		Provides: map[string]internalcharm.Relation{
			"juju-info": {
				Name:      "juju-info",
				Role:      internalcharm.RoleProvider,
				Interface: "juju-info",
				Scope:     "global",
			},
		},
	})
}

func (s *importSuite) TestImportMaximalCharmManifest(c *tc.C) {
	desc := description.NewModel(description.ModelArgs{
		Type: string(model.IAAS),
	})

	app := desc.AddApplication(description.ApplicationArgs{
		Name:     "foo",
		CharmURL: "ch:foo-1",
	})
	app.SetCharmOrigin(description.CharmOriginArgs{
		Source:   "charm-hub",
		ID:       "deadbeef",
		Hash:     "deadbeef2",
		Revision: 1,
		Channel:  "latest/stable",
		Platform: "amd64/ubuntu/20.04",
	})
	app.SetCharmMetadata(description.CharmMetadataArgs{
		Name: "foo",
	})
	app.SetCharmManifest(description.CharmManifestArgs{
		Bases: []description.CharmManifestBase{
			manifestBase{
				name:    "ubuntu",
				channel: "stable",
			},
			manifestBase{
				name:          "ubuntu",
				channel:       "4.0/stable/foo",
				architectures: []string{"arm64"},
			},
			manifestBase{
				name:          "ubuntu",
				channel:       "latest/stable",
				architectures: []string{"amd64", "s390x", "ppc64el"},
			},
		},
	})

	coordinator := modelmigration.NewCoordinator(loggertesting.WrapCheckLog(c))
	applicationmodelmigration.RegisterImport(coordinator, clock.WallClock, loggertesting.WrapCheckLog(c))
	err := coordinator.Perform(c.Context(), modelmigration.NewScope(nil, s.TxnRunnerFactory(), nil, model.UUID(s.ModelUUID())), desc)
	c.Assert(err, tc.ErrorIsNil)

	svc := s.setupService(c)
	manifest, err := svc.GetCharmManifest(c.Context(), charm.CharmLocator{
		Name:     "foo",
		Revision: 1,
		Source:   charm.CharmHubSource,
	})
	c.Assert(err, tc.ErrorIsNil)

	c.Check(manifest, tc.DeepEquals, internalcharm.Manifest{
		Bases: []internalcharm.Base{
			{
				Name:          "ubuntu",
				Channel:       internalcharm.Channel{Risk: "stable"},
				Architectures: []string{"amd64"},
			},
			{
				Name:          "ubuntu",
				Channel:       internalcharm.Channel{Track: "4.0", Risk: "stable", Branch: "foo"},
				Architectures: []string{"arm64"},
			},
			{
				Name:          "ubuntu",
				Channel:       internalcharm.Channel{Track: "latest", Risk: "stable"},
				Architectures: []string{"amd64", "s390x", "ppc64el"},
			},
		},
	})
}

func (s *importSuite) TestImportMinimalCharmManifest(c *tc.C) {
	desc := description.NewModel(description.ModelArgs{
		Type: string(model.IAAS),
	})

	app := desc.AddApplication(description.ApplicationArgs{
		Name:     "foo",
		CharmURL: "ch:foo-1",
	})
	app.SetCharmOrigin(description.CharmOriginArgs{
		Source:   "charm-hub",
		ID:       "deadbeef",
		Hash:     "deadbeef2",
		Revision: 1,
		Channel:  "latest/stable",
		Platform: "amd64/ubuntu/20.04",
	})
	app.SetCharmMetadata(description.CharmMetadataArgs{
		Name: "foo",
	})
	app.SetCharmManifest(description.CharmManifestArgs{
		Bases: []description.CharmManifestBase{
			manifestBase{
				name:    "ubuntu",
				channel: "stable",
			},
		},
	})

	coordinator := modelmigration.NewCoordinator(loggertesting.WrapCheckLog(c))
	applicationmodelmigration.RegisterImport(coordinator, clock.WallClock, loggertesting.WrapCheckLog(c))
	err := coordinator.Perform(c.Context(), modelmigration.NewScope(nil, s.TxnRunnerFactory(), nil, model.UUID(s.ModelUUID())), desc)
	c.Assert(err, tc.ErrorIsNil)

	svc := s.setupService(c)
	manifest, err := svc.GetCharmManifest(c.Context(), charm.CharmLocator{
		Name:     "foo",
		Revision: 1,
		Source:   charm.CharmHubSource,
	})
	c.Assert(err, tc.ErrorIsNil)

	c.Check(manifest, tc.DeepEquals, internalcharm.Manifest{
		Bases: []internalcharm.Base{
			{
				Name:          "ubuntu",
				Channel:       internalcharm.Channel{Risk: "stable"},
				Architectures: []string{"amd64"},
			},
		},
	})
}

func (s *importSuite) TestImportMinimalCharmConfig(c *tc.C) {
	desc := description.NewModel(description.ModelArgs{
		Type: string(model.IAAS),
	})

	app := desc.AddApplication(description.ApplicationArgs{
		Name:     "foo",
		CharmURL: "ch:foo-1",
	})
	app.SetCharmOrigin(description.CharmOriginArgs{
		Source:   "charm-hub",
		ID:       "deadbeef",
		Hash:     "deadbeef2",
		Revision: 1,
		Channel:  "latest/stable",
		Platform: "amd64/ubuntu/20.04",
	})
	app.SetCharmMetadata(description.CharmMetadataArgs{
		Name: "foo",
	})
	app.SetCharmManifest(description.CharmManifestArgs{
		Bases: []description.CharmManifestBase{
			manifestBase{
				name:          "ubuntu",
				channel:       "stable",
				architectures: []string{"amd64"},
			},
		},
	})
	app.SetCharmConfigs(description.CharmConfigsArgs{
		Configs: map[string]description.CharmConfig{
			"foo": config{
				configType:   "string",
				defaultValue: "bar",
				description:  "foo description",
			},
		},
	})

	coordinator := modelmigration.NewCoordinator(loggertesting.WrapCheckLog(c))
	applicationmodelmigration.RegisterImport(coordinator, clock.WallClock, loggertesting.WrapCheckLog(c))
	err := coordinator.Perform(c.Context(), modelmigration.NewScope(nil, s.TxnRunnerFactory(), nil, model.UUID(s.ModelUUID())), desc)
	c.Assert(err, tc.ErrorIsNil)

	svc := s.setupService(c)
	config, err := svc.GetCharmConfig(c.Context(), charm.CharmLocator{
		Name:     "foo",
		Revision: 1,
		Source:   charm.CharmHubSource,
	})
	c.Assert(err, tc.ErrorIsNil)

	c.Check(config, tc.DeepEquals, internalcharm.ConfigSpec{
		Options: map[string]internalcharm.Option{
			"foo": {
				Type:        "string",
				Default:     "bar",
				Description: "foo description",
			},
		},
	})
}

func (s *importSuite) TestImportMaximalCharmConfig(c *tc.C) {
	desc := description.NewModel(description.ModelArgs{
		Type: string(model.IAAS),
	})

	app := desc.AddApplication(description.ApplicationArgs{
		Name:     "foo",
		CharmURL: "ch:foo-1",
	})
	app.SetCharmOrigin(description.CharmOriginArgs{
		Source:   "charm-hub",
		ID:       "deadbeef",
		Hash:     "deadbeef2",
		Revision: 1,
		Channel:  "latest/stable",
		Platform: "amd64/ubuntu/20.04",
	})
	app.SetCharmMetadata(description.CharmMetadataArgs{
		Name: "foo",
	})
	app.SetCharmManifest(description.CharmManifestArgs{
		Bases: []description.CharmManifestBase{
			manifestBase{
				name:          "ubuntu",
				channel:       "stable",
				architectures: []string{"amd64"},
			},
		},
	})
	app.SetCharmConfigs(description.CharmConfigsArgs{
		Configs: map[string]description.CharmConfig{
			"foo": config{
				configType:   "string",
				defaultValue: "bar",
				description:  "foo description",
			},
			"baz": config{
				configType:   "int",
				defaultValue: 42,
				description:  "baz description",
			},
			"qux": config{
				configType:   "boolean",
				defaultValue: true,
				description:  "qux description",
			},
			"norf": config{
				configType:   "secret",
				defaultValue: "foo-bar-baz",
				description:  "norf description",
			},
		},
	})

	coordinator := modelmigration.NewCoordinator(loggertesting.WrapCheckLog(c))
	applicationmodelmigration.RegisterImport(coordinator, clock.WallClock, loggertesting.WrapCheckLog(c))
	err := coordinator.Perform(c.Context(), modelmigration.NewScope(nil, s.TxnRunnerFactory(), nil, model.UUID(s.ModelUUID())), desc)
	c.Assert(err, tc.ErrorIsNil)

	svc := s.setupService(c)
	config, err := svc.GetCharmConfig(c.Context(), charm.CharmLocator{
		Name:     "foo",
		Revision: 1,
		Source:   charm.CharmHubSource,
	})
	c.Assert(err, tc.ErrorIsNil)

	c.Check(config, tc.DeepEquals, internalcharm.ConfigSpec{
		Options: map[string]internalcharm.Option{
			"foo": {
				Type:        "string",
				Default:     "bar",
				Description: "foo description",
			},
			"baz": {
				Type:        "int",
				Default:     42,
				Description: "baz description",
			},
			"qux": {
				Type:        "boolean",
				Default:     true,
				Description: "qux description",
			},
			"norf": {
				Type:        "secret",
				Default:     "foo-bar-baz",
				Description: "norf description",
			},
		},
	})
}

func (s *importSuite) TestImportMinimalCharmActions(c *tc.C) {
	desc := description.NewModel(description.ModelArgs{
		Type: string(model.IAAS),
	})

	app := desc.AddApplication(description.ApplicationArgs{
		Name:     "foo",
		CharmURL: "ch:foo-1",
	})
	app.SetCharmOrigin(description.CharmOriginArgs{
		Source:   "charm-hub",
		ID:       "deadbeef",
		Hash:     "deadbeef2",
		Revision: 1,
		Channel:  "latest/stable",
		Platform: "amd64/ubuntu/20.04",
	})
	app.SetCharmMetadata(description.CharmMetadataArgs{
		Name: "foo",
	})
	app.SetCharmManifest(description.CharmManifestArgs{
		Bases: []description.CharmManifestBase{
			manifestBase{
				name:          "ubuntu",
				channel:       "stable",
				architectures: []string{"amd64"},
			},
		},
	})
	app.SetCharmActions(description.CharmActionsArgs{
		Actions: map[string]description.CharmAction{
			"foo": action{
				description:    "foo description",
				parallel:       true,
				executionGroup: "bar",
				params: map[string]any{
					"a": int(1),
				},
			},
		},
	})

	coordinator := modelmigration.NewCoordinator(loggertesting.WrapCheckLog(c))
	applicationmodelmigration.RegisterImport(coordinator, clock.WallClock, loggertesting.WrapCheckLog(c))
	err := coordinator.Perform(c.Context(), modelmigration.NewScope(nil, s.TxnRunnerFactory(), nil, model.UUID(s.ModelUUID())), desc)
	c.Assert(err, tc.ErrorIsNil)

	svc := s.setupService(c)
	actions, err := svc.GetCharmActions(c.Context(), charm.CharmLocator{
		Name:     "foo",
		Revision: 1,
		Source:   charm.CharmHubSource,
	})
	c.Assert(err, tc.ErrorIsNil)

	c.Check(actions, tc.DeepEquals, internalcharm.Actions{
		ActionSpecs: map[string]internalcharm.ActionSpec{
			"foo": {
				Description:    "foo description",
				Parallel:       true,
				ExecutionGroup: "bar",
				Params: map[string]any{
					// All params are marshalled to JSON to try and keep all
					// types consistent when there are complex types. But, the
					// downside is that numbers become float64.
					"a": float64(1),
				},
			},
		},
	})
}

func (s *importSuite) TestImportMaximalCharmActions(c *tc.C) {
	desc := description.NewModel(description.ModelArgs{
		Type: string(model.IAAS),
	})

	app := desc.AddApplication(description.ApplicationArgs{
		Name:     "foo",
		CharmURL: "ch:foo-1",
	})
	app.SetCharmOrigin(description.CharmOriginArgs{
		Source:   "charm-hub",
		ID:       "deadbeef",
		Hash:     "deadbeef2",
		Revision: 1,
		Channel:  "latest/stable",
		Platform: "amd64/ubuntu/20.04",
	})
	app.SetCharmMetadata(description.CharmMetadataArgs{
		Name: "foo",
	})
	app.SetCharmManifest(description.CharmManifestArgs{
		Bases: []description.CharmManifestBase{
			manifestBase{
				name:          "ubuntu",
				channel:       "stable",
				architectures: []string{"amd64"},
			},
		},
	})
	app.SetCharmActions(description.CharmActionsArgs{
		Actions: map[string]description.CharmAction{
			"foo": action{
				description:    "foo description",
				parallel:       true,
				executionGroup: "bar",
				params: map[string]any{
					"a": int(1),
					"b": "string param",
					"c": true,
					"d": 3.14,
					"e": []any{1, 2.0, "x"},
					"f": map[string]any{
						"nested": "value",
					},
				},
			},
			"baz": action{
				description: "baz description",
			},
		},
	})

	coordinator := modelmigration.NewCoordinator(loggertesting.WrapCheckLog(c))
	applicationmodelmigration.RegisterImport(coordinator, clock.WallClock, loggertesting.WrapCheckLog(c))
	err := coordinator.Perform(c.Context(), modelmigration.NewScope(nil, s.TxnRunnerFactory(), nil, model.UUID(s.ModelUUID())), desc)
	c.Assert(err, tc.ErrorIsNil)

	svc := s.setupService(c)
	actions, err := svc.GetCharmActions(c.Context(), charm.CharmLocator{
		Name:     "foo",
		Revision: 1,
		Source:   charm.CharmHubSource,
	})
	c.Assert(err, tc.ErrorIsNil)

	c.Check(actions, tc.DeepEquals, internalcharm.Actions{
		ActionSpecs: map[string]internalcharm.ActionSpec{
			"foo": {
				Description:    "foo description",
				Parallel:       true,
				ExecutionGroup: "bar",
				Params: map[string]any{
					// All params are marshalled to JSON to try and keep all
					// types consistent when there are complex types. But, the
					// downside is that numbers become float64.
					"a": float64(1),
					"b": "string param",
					"c": true,
					"d": 3.14,
					"e": []any{float64(1), 2.0, "x"},
					"f": map[string]any{
						"nested": "value",
					},
				},
			},
			"baz": {
				Description: "baz description",
			},
		},
	})
}

func (s *importSuite) setupService(c *tc.C) *service.Service {
	modelDB := func(context.Context) (database.TxnRunner, error) {
		return s.ModelTxnRunner(), nil
	}
	modelUUID := model.UUID(s.ModelUUID())

	return service.NewService(
		state.NewState(modelDB, modelUUID, clock.WallClock, loggertesting.WrapCheckLog(c)),
		domaintesting.NoopLeaderEnsurer(),
		nil,
		domain.NewStatusHistory(loggertesting.WrapCheckLog(c), clock.WallClock),
		modelUUID,
		clock.WallClock,
		loggertesting.WrapCheckLog(c),
	)
}

type manifestBase struct {
	name          string
	channel       string
	architectures []string
}

func (m manifestBase) Name() string {
	return m.name
}
func (m manifestBase) Channel() string {
	return m.channel
}
func (m manifestBase) Architectures() []string {
	return m.architectures
}

type relation struct {
	name          string
	role          string
	interfaceName string
	optional      bool
	limit         int
	scope         string
}

func (r relation) Name() string {
	return r.name
}

func (r relation) Role() string {
	return r.role
}

func (r relation) Interface() string {
	return r.interfaceName
}

func (r relation) Optional() bool {
	return r.optional
}

func (r relation) Limit() int {
	return r.limit
}

func (r relation) Scope() string {
	return r.scope
}

type storage struct {
	name        string
	stype       string
	description string
	shared      bool
	readonly    bool
	countMin    int
	countMax    int
	minimumSize int
	location    string
	properties  []string
}

func (s storage) Name() string {
	return s.name
}

func (s storage) Description() string {
	return s.description
}

func (s storage) Type() string {
	return s.stype
}

func (s storage) Shared() bool {
	return s.shared
}

func (s storage) Readonly() bool {
	return s.readonly
}

func (s storage) CountMin() int {
	return s.countMin
}

func (s storage) CountMax() int {
	return s.countMax
}

func (s storage) MinimumSize() int {
	return s.minimumSize
}

func (s storage) Location() string {
	return s.location
}

func (s storage) Properties() []string {
	return s.properties
}

type device struct {
	name        string
	description string
	dtype       string
	countMin    int
	countMax    int
}

func (d device) Name() string {
	return d.name
}

func (d device) Description() string {
	return d.description
}

func (d device) Type() string {
	return d.dtype
}

func (d device) CountMin() int {
	return d.countMin
}

func (d device) CountMax() int {
	return d.countMax
}

type container struct {
	resource string
	mounts   []description.CharmMetadataContainerMount
	uid      *int
	gid      *int
}

func (c container) Resource() string {
	return c.resource
}

func (c container) Mounts() []description.CharmMetadataContainerMount {
	return c.mounts
}

func (c container) Uid() *int {
	return c.uid
}

func (c container) Gid() *int {
	return c.gid
}

type containerMount struct {
	storage  string
	location string
}

func (cm containerMount) Storage() string {
	return cm.storage
}

func (cm containerMount) Location() string {
	return cm.location
}

type resourceMeta struct {
	name        string
	rtype       string
	description string
	path        string
}

func (r resourceMeta) Name() string {
	return r.name
}

func (r resourceMeta) Type() string {
	return r.rtype
}

func (r resourceMeta) Description() string {
	return r.description
}

func (r resourceMeta) Path() string {
	return r.path
}

func ptr[T any](i T) *T {
	return &i
}

type config struct {
	configType   string
	defaultValue any
	description  string
}

func (c config) Type() string {
	return c.configType
}

func (c config) Default() any {
	return c.defaultValue
}

func (c config) Description() string {
	return c.description
}

type action struct {
	description    string
	parallel       bool
	executionGroup string
	params         map[string]any
}

func (a action) Description() string {
	return a.description
}

func (a action) Parallel() bool {
	return a.parallel
}

func (a action) ExecutionGroup() string {
	return a.executionGroup
}

func (a action) Parameters() map[string]any {
	return a.params
}
