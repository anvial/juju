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

func (s *importSuite) TestImportCharm(c *tc.C) {
	desc := description.NewModel(description.ModelArgs{
		Type: string(model.IAAS),
	})

	// Create a charm description and write it into the model, then import
	// it using the model migration framework. Verify that the charm has been
	// imported correctly into the database.

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
