// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/juju/description/v8"
	"github.com/juju/errors"
	"github.com/juju/names/v5"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/version/v2"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/environschema.v1"
	"gopkg.in/macaroon.v2"

	"github.com/juju/juju/core/arch"
	"github.com/juju/juju/core/constraints"
	"github.com/juju/juju/core/instance"
	"github.com/juju/juju/core/network"
	"github.com/juju/juju/core/payloads"
	"github.com/juju/juju/core/permission"
	"github.com/juju/juju/core/resources"
	resourcetesting "github.com/juju/juju/core/resources/testing"
	"github.com/juju/juju/core/status"
	jujuversion "github.com/juju/juju/core/version"
	domainstorage "github.com/juju/juju/domain/storage"
	"github.com/juju/juju/internal/charm"
	charmresource "github.com/juju/juju/internal/charm/resource"
	"github.com/juju/juju/internal/featureflag"
	internalpassword "github.com/juju/juju/internal/password"
	"github.com/juju/juju/internal/storage"
	"github.com/juju/juju/internal/testing"
	"github.com/juju/juju/internal/testing/factory"
	"github.com/juju/juju/state"
	"github.com/juju/juju/state/cloudimagemetadata"
)

// Constraints stores megabytes by default for memory and root disk.
const (
	gig uint64 = 1024

	addedHistoryCount = 5
	// 6 for the one initial + 5 added.
	expectedHistoryCount = addedHistoryCount + 1
)

type MigrationBaseSuite struct {
	ConnWithWallClockSuite
}

func (s *MigrationBaseSuite) setLatestTools(c *gc.C, latestTools version.Number) {
	err := s.Model.UpdateLatestToolsVersion(latestTools)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *MigrationBaseSuite) setRandSequenceValue(c *gc.C, name string) int {
	var value int
	var err error
	count := rand.Intn(5) + 1
	for i := 0; i < count; i++ {
		value, err = state.Sequence(s.State, name)
		c.Assert(err, jc.ErrorIsNil)
	}
	// The value stored in the doc is one higher than what it returns.
	return value + 1
}

func (s *MigrationBaseSuite) primeStatusHistory(c *gc.C, entity statusSetter, statusVal status.Status, count int) {
	primeStatusHistory(c, s.StatePool.Clock(), entity, statusVal, count, func(i int) map[string]interface{} {
		return map[string]interface{}{"index": count - i}
	}, 0, "")
}

func (s *MigrationBaseSuite) makeApplicationWithUnits(c *gc.C, applicationName string, count int) {
	units := make([]*state.Unit, count)
	application := s.Factory.MakeApplication(c, &factory.ApplicationParams{
		Name: applicationName,
		Charm: s.Factory.MakeCharm(c, &factory.CharmParams{
			Name: applicationName,
		}),
	})
	for i := 0; i < count; i++ {
		units[i] = s.Factory.MakeUnit(c, &factory.UnitParams{
			Application: application,
		})
	}
}

func (s *MigrationBaseSuite) makeUnitWithStorage(c *gc.C) (*state.Application, *state.Unit, names.StorageTag) {
	pool := "modelscoped"
	kind := "block"
	// Create a default pool for block devices.
	s.policy.Providers = map[string]domainstorage.StoragePoolDetails{
		pool: {Name: pool, Provider: "loop"},
	}

	// There are test charms called "storage-block" and
	// "storage-filesystem" which are what you'd expect.
	ch := s.AddTestingCharm(c, "storage-"+kind)
	storage := map[string]state.StorageConstraints{
		"data": makeStorageCons(pool, 1024, 1),
	}
	application := s.AddTestingApplicationWithStorage(c, "storage-"+kind, ch, storage)
	unit, err := application.AddUnit(state.AddUnitParams{})
	c.Assert(err, jc.ErrorIsNil)

	machine := s.Factory.MakeMachine(c, nil)
	err = unit.AssignToMachine(machine)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(err, jc.ErrorIsNil)
	storageTag := names.NewStorageTag("data/0")
	agentVersion := version.MustParseBinary("2.0.1-ubuntu-and64")
	err = unit.SetAgentVersion(agentVersion)
	c.Assert(err, jc.ErrorIsNil)
	return application, unit, storageTag
}

type MigrationExportSuite struct {
	MigrationBaseSuite
}

var _ = gc.Suite(&MigrationExportSuite{})

func (s *MigrationExportSuite) SetUpTest(c *gc.C) {
	s.MigrationBaseSuite.SetUpTest(c)
	s.SetFeatureFlags(featureflag.StrictMigration)
}

func (s *MigrationExportSuite) checkStatusHistory(c *gc.C, history []description.Status, statusVal status.Status) {
	for i, st := range history {
		c.Logf("status history #%d: %s", i, st.Updated())
		c.Check(st.Value(), gc.Equals, string(statusVal))
		c.Check(st.Message(), gc.Equals, "")
		c.Check(st.Data(), jc.DeepEquals, map[string]interface{}{"index": i + 1})
	}
}

func (s *MigrationExportSuite) TestModelInfo(c *gc.C) {
	latestTools := version.MustParse("2.0.1")
	s.setLatestTools(c, latestTools)
	err := s.State.SetModelConstraints(constraints.MustParse("arch=amd64 mem=8G"))
	c.Assert(err, jc.ErrorIsNil)
	machineSeq := s.setRandSequenceValue(c, "machine")
	fooSeq := s.setRandSequenceValue(c, "application-foo")

	err = s.State.SwitchBlockOn(state.ChangeBlock, "locked down")
	c.Assert(err, jc.ErrorIsNil)

	err = s.Model.SetPassword("supppperrrrsecret1235556667777")
	c.Assert(err, jc.ErrorIsNil)

	environVersion := 123
	err = s.Model.SetEnvironVersion(environVersion)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(model.PasswordHash(), gc.Equals, internalpassword.AgentPasswordHash("supppperrrrsecret1235556667777"))
	c.Assert(model.Type(), gc.Equals, string(s.Model.Type()))
	c.Assert(model.Tag(), gc.Equals, s.Model.ModelTag())
	c.Assert(model.Owner(), gc.Equals, s.Model.Owner())
	dbModelCfg, err := s.Model.Config()
	c.Assert(err, jc.ErrorIsNil)
	modelAttrs := dbModelCfg.AllAttrs()
	modelCfg := model.Config()
	// Config as read from state has resources tags coerced to a map.
	modelCfg["resource-tags"] = map[string]string{}
	c.Assert(modelCfg, jc.DeepEquals, modelAttrs)
	c.Assert(model.LatestToolsVersion(), gc.Equals, latestTools)
	c.Assert(model.EnvironVersion(), gc.Equals, environVersion)
	constraints := model.Constraints()
	c.Assert(constraints, gc.NotNil)
	c.Assert(constraints.Architecture(), gc.Equals, "amd64")
	c.Assert(constraints.Memory(), gc.Equals, 8*gig)
	c.Assert(model.Sequences(), jc.DeepEquals, map[string]int{
		"machine":         machineSeq,
		"application-foo": fooSeq,
		// blocks is added by the switch block on call above.
		"block": 1,
	})
	c.Assert(model.Blocks(), jc.DeepEquals, map[string]string{
		"all-changes": "locked down",
	})
}

func (s *MigrationExportSuite) TestModelUsers(c *gc.C) {
	// Make sure we have some last connection times for the admin user,
	// and create a few other users.
	lastConnection := state.NowToTheSecond(s.State)
	owner, err := s.State.UserAccess(s.Owner, s.Model.ModelTag())
	c.Assert(err, jc.ErrorIsNil)
	err = state.UpdateModelUserLastConnection(s.State, owner, lastConnection)
	c.Assert(err, jc.ErrorIsNil)

	bobTag := names.NewUserTag("bob@external")
	bob, err := s.Model.AddUser(state.UserAccessSpec{
		User:      bobTag,
		CreatedBy: s.Owner,
		Access:    permission.ReadAccess,
	})
	c.Assert(err, jc.ErrorIsNil)
	err = state.UpdateModelUserLastConnection(s.State, bob, lastConnection)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	users := model.Users()
	c.Assert(users, gc.HasLen, 2)

	exportedBob := users[0]
	// admin is "test-admin", and results are sorted
	exportedAdmin := users[1]

	c.Assert(exportedAdmin.Name(), gc.Equals, s.Owner)
	c.Assert(exportedAdmin.DisplayName(), gc.Equals, owner.DisplayName)
	c.Assert(exportedAdmin.CreatedBy(), gc.Equals, s.Owner)
	c.Assert(exportedAdmin.DateCreated(), gc.Equals, owner.DateCreated)
	c.Assert(exportedAdmin.LastConnection(), gc.Equals, lastConnection)
	c.Assert(exportedAdmin.Access(), gc.Equals, "admin")

	c.Assert(exportedBob.Name(), gc.Equals, bobTag)
	c.Assert(exportedBob.DisplayName(), gc.Equals, "")
	c.Assert(exportedBob.CreatedBy(), gc.Equals, s.Owner)
	c.Assert(exportedBob.DateCreated(), gc.Equals, bob.DateCreated)
	c.Assert(exportedBob.LastConnection(), gc.Equals, lastConnection)
	c.Assert(exportedBob.Access(), gc.Equals, "read")
}

func (s *MigrationExportSuite) TestMachines(c *gc.C) {
	s.assertMachinesMigrated(c, constraints.MustParse("arch=amd64 mem=8G tags=foo,bar spaces=dmz"))
}

func (s *MigrationExportSuite) TestMachinesWithVirtConstraint(c *gc.C) {
	s.assertMachinesMigrated(c, constraints.MustParse("arch=amd64 mem=8G virt-type=lxd"))
}

func (s *MigrationExportSuite) TestMachinesWithRootDiskSourceConstraint(c *gc.C) {
	s.assertMachinesMigrated(c, constraints.MustParse("arch=amd64 mem=8G root-disk-source=aldous"))
}

func (s *MigrationExportSuite) assertMachinesMigrated(c *gc.C, cons constraints.Value) {
	// Add a machine with an LXC container.
	source := "vashti"
	displayName := "test-display-name"

	addr := network.NewSpaceAddress("1.1.1.1")
	addr.SpaceID = "0"

	machine1 := s.Factory.MakeMachine(c, &factory.MachineParams{
		Constraints: cons,
		Characteristics: &instance.HardwareCharacteristics{
			RootDiskSource: &source,
		},
		DisplayName: displayName,
		Addresses:   network.SpaceAddresses{addr},
	})
	nested := s.Factory.MakeMachineNested(c, machine1.Id(), nil)

	s.primeStatusHistory(c, machine1, status.Started, addedHistoryCount)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	machines := model.Machines()
	c.Assert(machines, gc.HasLen, 1)

	exported := machines[0]
	c.Assert(exported.Tag(), gc.Equals, machine1.MachineTag())
	c.Assert(exported.Base(), gc.Equals, machine1.Base().String())

	expCons := exported.Constraints()
	c.Assert(expCons, gc.NotNil)
	c.Assert(expCons.Architecture(), gc.Equals, *cons.Arch)
	c.Assert(expCons.Memory(), gc.Equals, *cons.Mem)
	if cons.HasVirtType() {
		c.Assert(expCons.VirtType(), gc.Equals, *cons.VirtType)
	}
	if cons.HasRootDiskSource() {
		c.Assert(expCons.RootDiskSource(), gc.Equals, *cons.RootDiskSource)
	}
	if cons.HasRootDisk() {
		c.Assert(expCons.RootDisk(), gc.Equals, *cons.RootDisk)
	}

	tools, err := machine1.AgentTools()
	c.Assert(err, jc.ErrorIsNil)
	exTools := exported.Tools()
	c.Assert(exTools, gc.NotNil)
	c.Assert(exTools.Version(), jc.DeepEquals, tools.Version)

	history := exported.StatusHistory()
	c.Assert(history, gc.HasLen, expectedHistoryCount)
	s.checkStatusHistory(c, history[:addedHistoryCount], status.Started)

	containers := exported.Containers()
	c.Assert(containers, gc.HasLen, 1)
	container := containers[0]
	c.Assert(container.Tag(), gc.Equals, nested.MachineTag())

	// Ensure that a new machine has a modification set to its initial state.
	inst := exported.Instance()
	c.Assert(inst.ModificationStatus().Value(), gc.Equals, "idle")
	c.Assert(inst.RootDiskSource(), gc.Equals, "vashti")
	c.Assert(inst.DisplayName(), gc.Equals, displayName)

	c.Assert(exported.ProviderAddresses(), gc.HasLen, 1)
	exAddr := exported.ProviderAddresses()[0]
	c.Assert(exAddr.Value(), gc.Equals, "1.1.1.1")
	c.Assert(exAddr.SpaceID(), gc.Equals, "0")
}

func (s *MigrationExportSuite) TestApplications(c *gc.C) {
	s.assertMigrateApplications(c, s.State, constraints.MustParse("arch=amd64 mem=8G"))
}

func (s *MigrationExportSuite) TestCAASSidecarApplications(c *gc.C) {
	caasSt := s.Factory.MakeCAASModel(c, nil)
	s.AddCleanup(func(_ *gc.C) { caasSt.Close() })

	s.assertMigrateApplications(c, caasSt, constraints.MustParse("arch=amd64 mem=8G"))
}

func (s *MigrationExportSuite) TestApplicationsWithVirtConstraint(c *gc.C) {
	s.assertMigrateApplications(c, s.State, constraints.MustParse("arch=amd64 mem=8G virt-type=lxd"))
}

func (s *MigrationExportSuite) TestApplicationsWithRootDiskSourceConstraint(c *gc.C) {
	s.assertMigrateApplications(c, s.State, constraints.MustParse("arch=amd64 mem=8G root-disk-source=vonnegut"))
}

func (s *MigrationExportSuite) assertMigrateApplications(c *gc.C, st *state.State, cons constraints.Value) {
	f := factory.NewFactory(st, s.StatePool, testing.FakeControllerConfig())

	dbModel, err := st.Model()
	c.Assert(err, jc.ErrorIsNil)
	series := "quantal"
	if dbModel.Type() == state.ModelTypeCAAS {
		series = "focal"
	}
	ch := f.MakeCharmV2(c, &factory.CharmParams{
		Name:   "snappass-test",
		Series: series,
	})
	application := f.MakeApplication(c, &factory.ApplicationParams{
		Charm: ch,
		CharmConfig: map[string]interface{}{
			"foo": "bar",
		},
		CharmOrigin: &state.CharmOrigin{
			Channel: &state.Channel{
				Risk: "beta",
			},
			Platform: &state.Platform{
				Architecture: "amd64",
				OS:           "ubuntu",
				Channel:      "20.04/stable",
			},
		},
		ApplicationConfig: map[string]interface{}{
			"app foo": "app bar",
		},
		ApplicationConfigFields: environschema.Fields{
			"app foo": environschema.Attr{Type: environschema.Tstring}},
		Constraints:  cons,
		DesiredScale: 3,
	})

	err = application.UpdateLeaderSettings(&goodToken{}, map[string]string{
		"leader": "true",
	})
	c.Assert(err, jc.ErrorIsNil)

	if dbModel.Type() == state.ModelTypeCAAS {
		_, err = application.AddUnit(state.AddUnitParams{ProviderId: strPtr("provider-id1")})
		c.Assert(err, jc.ErrorIsNil)
		err = application.SetOperatorStatus(status.StatusInfo{Status: status.Running})
		c.Assert(err, jc.ErrorIsNil)

		addr := network.NewSpaceAddress("192.168.1.1", network.WithScope(network.ScopeCloudLocal))
		err = application.UpdateCloudService("provider-id", []network.SpaceAddress{addr})
		c.Assert(err, jc.ErrorIsNil)

		err = application.SetProvisioningState(state.ApplicationProvisioningState{
			Scaling:     true,
			ScaleTarget: 3,
		})
		c.Assert(err, jc.ErrorIsNil)
	}

	agentVer, err := version.ParseBinary("2.9.1-ubuntu-amd64")
	c.Assert(err, jc.ErrorIsNil)
	units, err := application.AllUnits()
	c.Assert(err, jc.ErrorIsNil)
	for _, unit := range units {
		err = unit.SetAgentVersion(agentVer)
		c.Assert(err, jc.ErrorIsNil)
	}

	s.primeStatusHistory(c, application, status.Active, addedHistoryCount)

	model, err := st.Export(map[string]string{}, state.NewObjectStore(c, st.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	applications := model.Applications()
	c.Assert(applications, gc.HasLen, 1)

	exported := applications[0]
	c.Assert(exported.Name(), gc.Equals, application.Name())
	c.Assert(exported.Tag(), gc.Equals, application.ApplicationTag())
	c.Assert(exported.Type(), gc.Equals, string(dbModel.Type()))

	origin := exported.CharmOrigin()
	c.Assert(origin.Channel(), gc.Equals, "beta")
	c.Assert(origin.Platform(), gc.Equals, "amd64/ubuntu/20.04/stable")

	c.Assert(exported.CharmConfig(), jc.DeepEquals, map[string]interface{}{
		"foo": "bar",
	})
	c.Assert(exported.ApplicationConfig(), jc.DeepEquals, map[string]interface{}{
		"app foo": "app bar",
	})
	c.Assert(exported.LeadershipSettings(), jc.DeepEquals, map[string]interface{}{
		"leader": "true",
	})

	constraints := exported.Constraints()
	c.Assert(constraints, gc.NotNil)
	c.Assert(constraints.Architecture(), gc.Equals, *cons.Arch)
	c.Assert(constraints.Memory(), gc.Equals, *cons.Mem)
	if cons.HasVirtType() {
		c.Assert(constraints.VirtType(), gc.Equals, *cons.VirtType)
	}
	if cons.HasRootDiskSource() {
		c.Assert(constraints.RootDiskSource(), gc.Equals, *cons.RootDiskSource)
	}
	if cons.HasRootDisk() {
		c.Assert(constraints.RootDisk(), gc.Equals, *cons.RootDisk)
	}

	history := exported.StatusHistory()
	c.Assert(history, gc.HasLen, expectedHistoryCount)
	s.checkStatusHistory(c, history[:addedHistoryCount], status.Active)

	if dbModel.Type() == state.ModelTypeCAAS {
		units, err := application.AllUnits()
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(len(units), gc.Equals, len(exported.Units()))

		for _, exportedUnit := range exported.Units() {
			tools := exportedUnit.Tools()
			c.Assert(tools.Version(), gc.Equals, agentVer)
		}
		c.Assert(exported.CloudService().ProviderId(), gc.Equals, "provider-id")
		c.Assert(exported.DesiredScale(), gc.Equals, 3)
		c.Assert(exported.Placement(), gc.Equals, "")
		c.Assert(exported.HasResources(), jc.IsTrue)
		addresses := exported.CloudService().Addresses()
		addr := addresses[0]
		c.Assert(addr.Value(), gc.Equals, "192.168.1.1")
		c.Assert(addr.Scope(), gc.Equals, "local-cloud")
		c.Assert(addr.Type(), gc.Equals, "ipv4")
		c.Assert(addr.Origin(), gc.Equals, "provider")
	} else {
		c.Assert(exported.CloudService(), gc.IsNil)
		_, err := application.AgentTools()
		c.Assert(err, jc.ErrorIs, errors.NotFound)
	}

	if dbModel.Type() == state.ModelTypeCAAS {
		ps := exported.ProvisioningState()
		c.Assert(ps, gc.NotNil)
		c.Assert(ps.Scaling(), jc.IsTrue)
		c.Assert(ps.ScaleTarget(), gc.Equals, 3)
	}

	// Check that we're exporting the metadata.

	exportedCharmMetadata := exported.CharmMetadata()
	c.Assert(exportedCharmMetadata, gc.NotNil)
	c.Check(exportedCharmMetadata.Name(), gc.Equals, ch.Meta().Name)
	c.Check(exportedCharmMetadata.Summary(), gc.Equals, ch.Meta().Summary)
	c.Check(exportedCharmMetadata.Description(), gc.Equals, ch.Meta().Description)
	c.Check(exportedCharmMetadata.Categories(), jc.DeepEquals, ch.Meta().Categories)
	c.Check(exportedCharmMetadata.Tags(), jc.DeepEquals, ch.Meta().Tags)
	c.Check(exportedCharmMetadata.Subordinate(), gc.Equals, ch.Meta().Subordinate)
	c.Check(exportedCharmMetadata.Terms(), jc.DeepEquals, ch.Meta().Terms)

	extraBindings := make(map[string]string)
	for name, binding := range ch.Meta().ExtraBindings {
		extraBindings[name] = binding.Name
	}
	c.Check(exportedCharmMetadata.ExtraBindings(), jc.DeepEquals, extraBindings)

	expectedProvides := make(map[string]string)
	for name, provider := range ch.Meta().Provides {
		expectedProvides[name] = provider.Name
	}
	exportedProvides := make(map[string]string)
	for name, provider := range exportedCharmMetadata.Provides() {
		exportedProvides[name] = provider.Name()
	}
	c.Check(exportedProvides, jc.DeepEquals, expectedProvides)

	expectedRequires := make(map[string]string)
	for name, provider := range ch.Meta().Requires {
		expectedRequires[name] = provider.Name
	}
	exportedRequires := make(map[string]string)
	for name, provider := range exportedCharmMetadata.Requires() {
		exportedRequires[name] = provider.Name()
	}
	c.Check(exportedRequires, jc.DeepEquals, expectedRequires)

	expectedPeers := make(map[string]string)
	for name, provider := range ch.Meta().Peers {
		expectedPeers[name] = provider.Name
	}
	exportedPeers := make(map[string]string)
	for name, provider := range exportedCharmMetadata.Peers() {
		exportedPeers[name] = provider.Name()
	}
	c.Check(exportedPeers, jc.DeepEquals, expectedPeers)

	// Check that we're exporting the manifest.
	exportedCharmManifest := exported.CharmManifest()
	c.Assert(exportedCharmManifest, gc.NotNil)

	expectedManifestBases := make([]string, 0)
	for _, base := range ch.Manifest().Bases {
		expectedManifestBases = append(expectedManifestBases, fmt.Sprintf("%s %s %v",
			base.Name,
			base.Channel.String(),
			base.Architectures,
		))
	}
	exportedManifestBases := make([]string, 0)
	for _, base := range exportedCharmManifest.Bases() {
		exportedManifestBases = append(exportedManifestBases, fmt.Sprintf("%s %s %v",
			base.Name(),
			base.Channel(),
			base.Architectures(),
		))
	}
	c.Check(exportedManifestBases, jc.DeepEquals, expectedManifestBases)
}

func (s *MigrationExportSuite) TestMalformedApplications(c *gc.C) {
	f := factory.NewFactory(s.State, s.StatePool, testing.FakeControllerConfig())

	dbModel, err := s.State.Model()
	c.Assert(err, jc.ErrorIsNil)
	series := "quantal"
	ch := f.MakeCharm(c, &factory.CharmParams{Series: series})

	application := f.MakeApplication(c, &factory.ApplicationParams{
		Charm: ch,
		CharmConfig: map[string]interface{}{
			"foo": "bar",
		},
		CharmOrigin: &state.CharmOrigin{
			Channel: &state.Channel{},
			Platform: &state.Platform{
				Architecture: "amd64",
				OS:           "ubuntu",
				Channel:      "20.04/stable",
			},
		},
		ApplicationConfig: map[string]interface{}{
			"app foo": "app bar",
		},
		ApplicationConfigFields: environschema.Fields{
			"app foo": environschema.Attr{Type: environschema.Tstring}},
		DesiredScale: 3,
	})

	err = application.UpdateLeaderSettings(&goodToken{}, map[string]string{
		"leader": "true",
	})
	c.Assert(err, jc.ErrorIsNil)

	agentVer, err := version.ParseBinary("2.9.1-ubuntu-amd64")
	c.Assert(err, jc.ErrorIsNil)

	units, err := application.AllUnits()
	c.Assert(err, jc.ErrorIsNil)
	for _, unit := range units {
		err = unit.SetAgentVersion(agentVer)
		c.Assert(err, jc.ErrorIsNil)
	}

	s.primeStatusHistory(c, application, status.Active, addedHistoryCount)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	applications := model.Applications()
	c.Assert(applications, gc.HasLen, 1)

	exported := applications[0]
	c.Assert(exported.Name(), gc.Equals, application.Name())
	c.Assert(exported.Tag(), gc.Equals, application.ApplicationTag())
	c.Assert(exported.Type(), gc.Equals, string(dbModel.Type()))

	origin := exported.CharmOrigin()
	c.Assert(origin.Channel(), gc.Equals, "stable")
	c.Assert(origin.Platform(), gc.Equals, "amd64/ubuntu/20.04/stable")
}

func (s *MigrationExportSuite) TestMultipleApplications(c *gc.C) {
	s.Factory.MakeApplication(c, &factory.ApplicationParams{Name: "first"})
	s.Factory.MakeApplication(c, &factory.ApplicationParams{Name: "second"})
	s.Factory.MakeApplication(c, &factory.ApplicationParams{Name: "third"})

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	applications := model.Applications()
	c.Assert(applications, gc.HasLen, 3)
}

func (s *MigrationExportSuite) TestOfferConnections(c *gc.C) {
	stOffer, err := s.State.AddOfferConnection(state.AddOfferConnectionParams{
		OfferUUID:       "offer-uuid",
		RelationId:      1,
		RelationKey:     "relation-key",
		SourceModelUUID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
		Username:        "fred",
	})
	c.Assert(err, jc.ErrorIsNil)

	// We only care for the offer connections
	model, err := s.State.ExportPartial(state.ExportConfig{
		SkipActions:              true,
		SkipAnnotations:          true,
		SkipCloudImageMetadata:   true,
		SkipCredentials:          true,
		SkipIPAddresses:          true,
		SkipSettings:             true,
		SkipSSHHostKeys:          true,
		SkipStatusHistory:        true,
		SkipLinkLayerDevices:     true,
		SkipUnitAgentBinaries:    true,
		SkipMachineAgentBinaries: true,
		SkipRelationData:         true,
		SkipInstanceData:         true,
		SkipApplicationOffers:    true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	offers := model.OfferConnections()
	c.Assert(offers, gc.HasLen, 1)
	offer := offers[0]
	c.Assert(offer.OfferUUID(), gc.Equals, stOffer.OfferUUID())
	c.Assert(offer.RelationID(), gc.Equals, stOffer.RelationId())
	c.Assert(offer.RelationKey(), gc.Equals, stOffer.RelationKey())
	c.Assert(offer.SourceModelUUID(), gc.Equals, stOffer.SourceModelUUID())
	c.Assert(offer.UserName(), gc.Equals, stOffer.UserName())
}

func (s *MigrationExportSuite) TestUnits(c *gc.C) {
	f := factory.NewFactory(s.State, s.StatePool, testing.FakeControllerConfig())
	unit := f.MakeUnit(c, &factory.UnitParams{
		Constraints: constraints.MustParse("arch=amd64 mem=8G"),
	})
	s.assertMigrateUnits(c, s.State, unit)
}

func (s *MigrationExportSuite) TestCAASUnits(c *gc.C) {
	caasSt := s.Factory.MakeCAASModel(c, nil)
	s.AddCleanup(func(_ *gc.C) { caasSt.Close() })

	f := factory.NewFactory(caasSt, s.StatePool, testing.FakeControllerConfig())
	app := f.MakeApplication(c, &factory.ApplicationParams{
		Constraints: constraints.MustParse("arch=amd64 mem=8G"),
	})
	unit := f.MakeUnit(c, &factory.UnitParams{
		Application: app,
	})
	s.assertMigrateUnits(c, caasSt, unit)
}

func (s *MigrationExportSuite) assertMigrateUnits(c *gc.C, st *state.State, unit *state.Unit) {
	for _, version := range []string{"garnet", "amethyst", "pearl", "steven"} {
		err := unit.SetWorkloadVersion(version)
		c.Assert(err, jc.ErrorIsNil)
	}
	us := state.NewUnitState()
	us.SetCharmState(map[string]string{"payload": "b4dc0ffee"})
	us.SetRelationState(map[int]string{42: "magic"})
	us.SetUniterState("uniter state")
	us.SetStorageState("storage state")
	err := unit.SetState(us, state.UnitStateSizeLimits{})
	c.Assert(err, jc.ErrorIsNil)

	dbModel, err := st.Model()
	c.Assert(err, jc.ErrorIsNil)

	if dbModel.Type() == state.ModelTypeCAAS {
		// need to set a cloud container status so that SetStatus for
		// the unit doesn't throw away the history writes.
		var updateUnits state.UpdateUnitsOperation
		updateUnits.Updates = []*state.UpdateUnitOperation{
			unit.UpdateOperation(state.UnitUpdateProperties{
				ProviderId: strPtr("provider-id"),
				Address:    strPtr("192.168.1.1"),
				Ports:      &[]string{"80"},
				CloudContainerStatus: &status.StatusInfo{
					Status:  status.Running,
					Message: "cloud container running",
				},
			})}
		app, err := unit.Application()
		c.Assert(err, jc.ErrorIsNil)
		err = app.UpdateUnits(&updateUnits)
		c.Assert(err, jc.ErrorIsNil)
	}

	s.primeStatusHistory(c, unit, status.Active, addedHistoryCount)
	s.primeStatusHistory(c, unit.Agent(), status.Idle, addedHistoryCount)

	model, err := st.Export(map[string]string{}, state.NewObjectStore(c, st.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	applications := model.Applications()
	c.Assert(applications, gc.HasLen, 1)

	application := applications[0]
	units := application.Units()
	c.Assert(units, gc.HasLen, 1)

	exported := units[0]

	c.Assert(exported.Name(), gc.Equals, unit.Name())
	c.Assert(exported.Tag(), gc.Equals, unit.UnitTag())
	c.Assert(exported.Validate(), jc.ErrorIsNil)
	c.Assert(exported.WorkloadVersion(), gc.Equals, "steven")
	c.Assert(exported.CharmState(), jc.DeepEquals, map[string]string{"payload": "b4dc0ffee"})
	c.Assert(exported.RelationState(), jc.DeepEquals, map[int]string{42: "magic"})
	c.Assert(exported.UniterState(), gc.Equals, "uniter state")
	c.Assert(exported.StorageState(), gc.Equals, "storage state")
	obtainedConstraints := exported.Constraints()
	c.Assert(obtainedConstraints, gc.NotNil)
	c.Assert(obtainedConstraints.Architecture(), gc.Equals, "amd64")
	c.Assert(obtainedConstraints.Memory(), gc.Equals, 8*gig)

	workloadHistory := exported.WorkloadStatusHistory()
	if dbModel.Type() == state.ModelTypeCAAS {
		// Account for the extra cloud container status history addition.
		c.Assert(workloadHistory, gc.HasLen, expectedHistoryCount+1)
		c.Assert(workloadHistory[expectedHistoryCount].Message(), gc.Equals, "installing agent")
		c.Assert(workloadHistory[expectedHistoryCount].Value(), gc.Equals, "waiting")
		c.Assert(workloadHistory[expectedHistoryCount-1].Message(), gc.Equals, "cloud container running")
		c.Assert(workloadHistory[expectedHistoryCount-1].Value(), gc.Equals, "running")
	} else {
		c.Assert(workloadHistory, gc.HasLen, expectedHistoryCount)
	}
	s.checkStatusHistory(c, workloadHistory[:addedHistoryCount], status.Active)

	agentHistory := exported.AgentStatusHistory()
	c.Assert(agentHistory, gc.HasLen, expectedHistoryCount)
	s.checkStatusHistory(c, agentHistory[:addedHistoryCount], status.Idle)

	versionHistory := exported.WorkloadVersionHistory()
	// There are extra entries at the start that we don't care about.
	c.Assert(len(versionHistory) >= 4, jc.IsTrue)
	versions := make([]string, 4)
	for i, s := range versionHistory[:4] {
		versions[i] = s.Message()
	}
	// The exporter reads history in reverse time order.
	c.Assert(versions, gc.DeepEquals, []string{"steven", "pearl", "amethyst", "garnet"})

	if dbModel.Type() == state.ModelTypeCAAS {
		containerInfo := exported.CloudContainer()
		c.Assert(containerInfo.ProviderId(), gc.Equals, "provider-id")
		c.Assert(containerInfo.Ports(), jc.DeepEquals, []string{"80"})
		addr := containerInfo.Address()
		c.Assert(addr, gc.NotNil)
		c.Assert(addr.Value(), gc.Equals, "192.168.1.1")
		c.Assert(addr.Scope(), gc.Equals, "local-machine")
		c.Assert(addr.Type(), gc.Equals, "ipv4")
		c.Assert(addr.Origin(), gc.Equals, "provider")
	}

	tools, err := unit.AgentTools()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(exported.Tools().Version(), gc.Equals, tools.Version)
}

func (s *MigrationExportSuite) TestApplicationLeadership(c *gc.C) {
	s.makeApplicationWithUnits(c, "mysql", 2)
	s.makeApplicationWithUnits(c, "wordpress", 4)

	model, err := s.State.Export(map[string]string{
		"mysql":     "mysql/1",
		"wordpress": "wordpress/2",
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	leaders := make(map[string]string)
	for _, application := range model.Applications() {
		leaders[application.Name()] = application.Leader()
	}
	c.Assert(leaders, jc.DeepEquals, map[string]string{
		"mysql":     "mysql/1",
		"wordpress": "wordpress/2",
	})
}

func (s *MigrationExportSuite) TestUnitOpenPortRanges(c *gc.C) {
	machine := s.Factory.MakeMachine(c, nil)
	unit := s.Factory.MakeUnit(c, &factory.UnitParams{
		Machine: machine,
	})
	c.Assert(unit.AssignToMachine(machine), jc.ErrorIsNil)

	state.MustOpenUnitPortRange(c, s.State, machine, unit.Name(), allEndpoints, network.MustParsePortRange("1234-2345/tcp"))

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	machines := model.Machines()
	c.Assert(machines, gc.HasLen, 1)

	unitPortRanges := machines[0].OpenedPortRanges().ByUnit()[unit.Name()]
	c.Assert(unitPortRanges, gc.Not(gc.IsNil), gc.Commentf("opened port ranges for unit not included in exported model"))

	unitPortRangesByEndpoint := unitPortRanges.ByEndpoint()
	c.Assert(unitPortRangesByEndpoint, gc.HasLen, 1)
	c.Assert(unitPortRangesByEndpoint[allEndpoints], gc.HasLen, 1)

	portRange := unitPortRangesByEndpoint[allEndpoints][0]
	c.Assert(portRange.FromPort(), gc.Equals, 1234)
	c.Assert(portRange.ToPort(), gc.Equals, 2345)
	c.Assert(portRange.Protocol(), gc.Equals, "tcp")
}

func (s *MigrationExportSuite) TestRemoteEntities(c *gc.C) {
	remotes := s.State.RemoteEntities()
	remoteCtrl := names.NewControllerTag("uuid-223412")

	err := remotes.ImportRemoteEntity(remoteCtrl, "aaa-bbb-ccc")
	c.Assert(err, jc.ErrorIsNil)

	mac, err := macaroon.New(nil, []byte(remoteCtrl.Id()), "", macaroon.LatestVersion)
	c.Assert(err, jc.ErrorIsNil)
	err = remotes.SaveMacaroon(remoteCtrl, mac)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	remoteEntities := model.RemoteEntities()
	c.Assert(remoteEntities, gc.HasLen, 1)

	entity := remoteEntities[0]
	c.Assert(entity.ID(), gc.Equals, names.NewControllerTag("uuid-223412").String())
	c.Assert(entity.Token(), gc.Equals, "aaa-bbb-ccc")
	c.Assert(entity.Macaroon(), gc.Equals, "")
}

func (s *MigrationExportSuite) TestRelationNetworks(c *gc.C) {
	wordpress := s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	wordpressEP, err := wordpress.Endpoint("db")
	c.Assert(err, jc.ErrorIsNil)
	mysql := s.AddTestingApplication(c, "mysql", s.AddTestingCharm(c, "mysql"))
	mysqlEP, err := mysql.Endpoint("server")
	c.Assert(err, jc.ErrorIsNil)
	_, err = s.State.AddRelation(wordpressEP, mysqlEP)
	c.Assert(err, jc.ErrorIsNil)

	_, err = state.NewRelationIngressNetworks(s.State).Save("wordpress:db mysql:server", false, []string{"192.168.1.0/16"})
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	relationNetwork := model.RelationNetworks()
	c.Assert(relationNetwork, gc.HasLen, 1)

	rin := relationNetwork[0]
	c.Assert(rin.RelationKey(), gc.Equals, "wordpress:db mysql:server")
	c.Assert(rin.CIDRS(), jc.DeepEquals, []string{"192.168.1.0/16"})
}

func (s *MigrationExportSuite) TestRelations(c *gc.C) {
	wordpress := state.AddTestingApplication(c, s.State, s.objectStore, "wordpress", state.AddTestingCharm(c, s.State, "wordpress"))
	mysql := state.AddTestingApplication(c, s.State, s.objectStore, "mysql", state.AddTestingCharm(c, s.State, "mysql"))
	// InferEndpoints will always return provider, requirer
	eps, err := s.State.InferEndpoints("mysql", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	rel, err := s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)
	msEp, wpEp := eps[0], eps[1]
	wordpress_0 := s.Factory.MakeUnit(c, &factory.UnitParams{Application: wordpress})
	mysql_0 := s.Factory.MakeUnit(c, &factory.UnitParams{Application: mysql})

	ru, err := rel.Unit(wordpress_0)
	c.Assert(err, jc.ErrorIsNil)
	wordpressSettings := map[string]interface{}{
		"name": "wordpress/0",
	}
	err = ru.EnterScope(wordpressSettings)
	c.Assert(err, jc.ErrorIsNil)

	ru, err = rel.Unit(mysql_0)
	c.Assert(err, jc.ErrorIsNil)
	mysqlSettings := map[string]interface{}{
		"name": "mysql/0",
	}
	err = ru.EnterScope(mysqlSettings)
	c.Assert(err, jc.ErrorIsNil)

	wordpressAppSettings := map[string]interface{}{
		"war": "worlds",
	}
	err = rel.UpdateApplicationSettings("wordpress", &fakeToken{}, wordpressAppSettings)
	c.Assert(err, jc.ErrorIsNil)

	mysqlAppSettings := map[string]interface{}{
		"million": "one",
	}
	err = rel.UpdateApplicationSettings("mysql", &fakeToken{}, mysqlAppSettings)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	rels := model.Relations()
	c.Assert(rels, gc.HasLen, 1)

	exRel := rels[0]
	c.Assert(exRel.Id(), gc.Equals, rel.Id())
	c.Assert(exRel.Key(), gc.Equals, rel.String())

	exEps := exRel.Endpoints()
	c.Assert(exEps, gc.HasLen, 2)

	checkEndpoint := func(
		exEndpoint description.Endpoint,
		unitName string,
		ep state.Endpoint,
		settings, appSettings map[string]interface{},
	) {
		c.Logf("%#v", exEndpoint)
		c.Check(exEndpoint.ApplicationName(), gc.Equals, ep.ApplicationName)
		c.Check(exEndpoint.Name(), gc.Equals, ep.Name)
		c.Check(exEndpoint.UnitCount(), gc.Equals, 1)
		c.Check(exEndpoint.Settings(unitName), jc.DeepEquals, settings)
		c.Check(exEndpoint.ApplicationSettings(), jc.DeepEquals, appSettings)
		c.Check(exEndpoint.Role(), gc.Equals, string(ep.Role))
		c.Check(exEndpoint.Interface(), gc.Equals, ep.Interface)
		c.Check(exEndpoint.Optional(), gc.Equals, ep.Optional)
		c.Check(exEndpoint.Limit(), gc.Equals, ep.Limit)
		c.Check(exEndpoint.Scope(), gc.Equals, string(ep.Scope))
	}
	checkEndpoint(exEps[0], mysql_0.Name(), msEp, mysqlSettings, mysqlAppSettings)
	checkEndpoint(exEps[1], wordpress_0.Name(), wpEp, wordpressSettings, wordpressAppSettings)

	// Make sure there is a status.
	status := exRel.Status()
	c.Check(status.Value(), gc.Equals, "joining")
}

func (s *MigrationExportSuite) TestSubordinateRelations(c *gc.C) {
	wordpress := state.AddTestingApplication(c, s.State, s.objectStore, "wordpress", state.AddTestingCharm(c, s.State, "wordpress"))
	mysql := state.AddTestingApplication(c, s.State, s.objectStore, "mysql", state.AddTestingCharm(c, s.State, "mysql"))
	wordpress_0 := s.Factory.MakeUnit(c, &factory.UnitParams{Application: wordpress})
	mysql_0 := s.Factory.MakeUnit(c, &factory.UnitParams{Application: mysql})

	logging := s.AddTestingApplication(c, "logging", s.AddTestingCharm(c, "logging"))

	addSubordinate := func(app *state.Application, unit *state.Unit) {
		eps, err := s.State.InferEndpoints(app.Name(), logging.Name())
		c.Assert(err, jc.ErrorIsNil)
		rel, err := s.State.AddRelation(eps...)
		c.Assert(err, jc.ErrorIsNil)
		pru, err := rel.Unit(unit)
		c.Assert(err, jc.ErrorIsNil)
		err = pru.EnterScope(nil)
		c.Assert(err, jc.ErrorIsNil)
		// Need to reload the doc to get the subordinates.
		err = unit.Refresh()
		c.Assert(err, jc.ErrorIsNil)
		subordinates := unit.SubordinateNames()
		c.Assert(subordinates, gc.HasLen, 1)
		loggingUnit, err := s.State.Unit(subordinates[0])
		c.Assert(err, jc.ErrorIsNil)
		sub, err := rel.Unit(loggingUnit)
		c.Assert(err, jc.ErrorIsNil)
		err = sub.EnterScope(nil)
		c.Assert(err, jc.ErrorIsNil)
	}

	addSubordinate(mysql, mysql_0)
	addSubordinate(wordpress, wordpress_0)

	setTools := func(unit *state.Unit) {
		app, err := unit.Application()
		c.Assert(err, jc.ErrorIsNil)
		agentTools := version.Binary{
			Number:  jujuversion.Current,
			Arch:    arch.HostArch(),
			Release: app.CharmOrigin().Platform.OS,
		}
		err = unit.SetAgentVersion(agentTools)
		c.Assert(err, jc.ErrorIsNil)
	}

	units, err := logging.AllUnits()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(units, gc.HasLen, 2)

	for _, unit := range units {
		setTools(unit)
	}

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	rels := model.Relations()
	c.Assert(rels, gc.HasLen, 2)
}

func (s *MigrationExportSuite) TestLinkLayerDevices(c *gc.C) {
	machine := s.Factory.MakeMachine(c, &factory.MachineParams{
		Constraints: constraints.MustParse("arch=amd64 mem=8G"),
	})
	deviceArgs := state.LinkLayerDeviceArgs{
		Name:            "foo",
		Type:            network.EthernetDevice,
		VirtualPortType: network.OvsPort,
	}
	err := machine.SetLinkLayerDevices(deviceArgs)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	devices := model.LinkLayerDevices()
	c.Assert(devices, gc.HasLen, 1)
	device := devices[0]
	c.Assert(device.Name(), gc.Equals, "foo")
	c.Assert(device.Type(), gc.Equals, string(network.EthernetDevice))
	c.Assert(device.VirtualPortType(), gc.Equals, "openvswitch", gc.Commentf("VirtualPortType was not exported correctly"))
}

func (s *MigrationExportSuite) TestLinkLayerDevicesSkipped(c *gc.C) {
	machine := s.Factory.MakeMachine(c, &factory.MachineParams{
		Constraints: constraints.MustParse("arch=amd64 mem=8G"),
	})
	deviceArgs := state.LinkLayerDeviceArgs{
		Name: "foo",
		Type: network.EthernetDevice,
	}
	err := machine.SetLinkLayerDevices(deviceArgs)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.ExportPartial(state.ExportConfig{
		SkipLinkLayerDevices: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	devices := model.LinkLayerDevices()
	c.Assert(devices, gc.HasLen, 0)
}

func (s *MigrationExportSuite) TestInstanceDataSkipped(c *gc.C) {
	s.Factory.MakeMachine(c, &factory.MachineParams{
		Constraints: constraints.MustParse("arch=amd64 mem=8G"),
	})

	model, err := s.State.ExportPartial(state.ExportConfig{
		SkipInstanceData: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	listMachines := model.Machines()

	instData := listMachines[0].Instance()
	c.Assert(instData, gc.Equals, nil)
}

func (s *MigrationExportSuite) TestMissingInstanceDataIgnored(c *gc.C) {
	_, err := s.State.AddOneMachine(defaultInstancePrechecker, state.MachineTemplate{
		Base: state.UbuntuBase("18.04"),
		Jobs: []state.MachineJob{state.JobManageModel},
	})
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.ExportPartial(state.ExportConfig{
		IgnoreIncompleteModel: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	listMachines := model.Machines()

	instData := listMachines[0].Instance()
	c.Assert(instData, gc.Equals, nil)
}

func (s *MigrationBaseSuite) TestMachineAgentBinariesSkipped(c *gc.C) {
	s.Factory.MakeMachine(c, &factory.MachineParams{
		Constraints: constraints.MustParse("arch=amd64 mem=8G"),
	})

	model, err := s.State.ExportPartial(state.ExportConfig{
		SkipMachineAgentBinaries: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	listMachines := model.Machines()
	tools := listMachines[0].Tools()
	c.Assert(tools, gc.Equals, nil)
}

func (s *MigrationBaseSuite) TestMissingMachineAgentBinariesIgnored(c *gc.C) {
	_, err := s.State.AddOneMachine(defaultInstancePrechecker, state.MachineTemplate{
		Base: state.UbuntuBase("18.04"),
		Jobs: []state.MachineJob{state.JobManageModel},
	})
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.ExportPartial(state.ExportConfig{
		IgnoreIncompleteModel: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	listMachines := model.Machines()
	tools := listMachines[0].Tools()
	c.Assert(tools, gc.Equals, nil)
}

func (s *MigrationBaseSuite) TestUnitAgentBinariesSkipped(c *gc.C) {
	dummyCharm := s.Factory.MakeCharm(c, &factory.CharmParams{Name: "dummy"})
	application := s.Factory.MakeApplication(c, &factory.ApplicationParams{Name: "dummy", Charm: dummyCharm})

	_, err := application.AddUnit(state.AddUnitParams{})
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.ExportPartial(state.ExportConfig{
		SkipUnitAgentBinaries: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	listApplications := model.Applications()
	unit := listApplications[0].Units()
	c.Assert(unit[0].Tools(), gc.Equals, nil)
}

func (s *MigrationBaseSuite) TestMissingUnitAgentBinariesIgnored(c *gc.C) {
	dummyCharm := s.Factory.MakeCharm(c, &factory.CharmParams{Name: "dummy"})
	application := s.Factory.MakeApplication(c, &factory.ApplicationParams{Name: "dummy", Charm: dummyCharm})

	_, err := application.AddUnit(state.AddUnitParams{})
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.ExportPartial(state.ExportConfig{
		IgnoreIncompleteModel: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	listApplications := model.Applications()
	unit := listApplications[0].Units()
	c.Assert(unit[0].Tools(), gc.Equals, nil)
}

func (s *MigrationBaseSuite) TestRelationScopeSkipped(c *gc.C) {
	wordpress := state.AddTestingApplication(c, s.State, s.objectStore, "wordpress", state.AddTestingCharm(c, s.State, "wordpress"))
	mysql := state.AddTestingApplication(c, s.State, s.objectStore, "mysql", state.AddTestingCharm(c, s.State, "mysql"))
	// InferEndpoints will always return provider, requirer
	eps, err := s.State.InferEndpoints("mysql", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	_, err = s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)
	s.Factory.MakeUnit(c, &factory.UnitParams{Application: wordpress})
	s.Factory.MakeUnit(c, &factory.UnitParams{Application: mysql})

	model, err := s.State.ExportPartial(state.ExportConfig{
		SkipRelationData: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(model.Relations(), gc.HasLen, 1)
}

func (s *MigrationBaseSuite) TestMissingRelationScopeIgnored(c *gc.C) {
	wordpress := state.AddTestingApplication(c, s.State, s.objectStore, "wordpress", state.AddTestingCharm(c, s.State, "wordpress"))
	mysql := state.AddTestingApplication(c, s.State, s.objectStore, "mysql", state.AddTestingCharm(c, s.State, "mysql"))
	// InferEndpoints will always return provider, requirer
	eps, err := s.State.InferEndpoints("mysql", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	_, err = s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)
	s.Factory.MakeUnit(c, &factory.UnitParams{Application: wordpress})
	s.Factory.MakeUnit(c, &factory.UnitParams{Application: mysql})

	model, err := s.State.ExportPartial(state.ExportConfig{
		IgnoreIncompleteModel: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(model.Relations(), gc.HasLen, 1)
}

func (s *MigrationExportSuite) TestSSHHostKeys(c *gc.C) {
	machine := s.Factory.MakeMachine(c, &factory.MachineParams{
		Constraints: constraints.MustParse("arch=amd64 mem=8G"),
	})
	err := s.State.SetSSHHostKeys(machine.MachineTag(), []string{"bam", "mam"})
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	keys := model.SSHHostKeys()
	c.Assert(keys, gc.HasLen, 1)
	key := keys[0]
	c.Assert(key.MachineID(), gc.Equals, machine.Id())
	c.Assert(key.Keys(), jc.DeepEquals, []string{"bam", "mam"})
}

func (s *MigrationExportSuite) TestSSHHostKeysSkipped(c *gc.C) {
	machine := s.Factory.MakeMachine(c, &factory.MachineParams{
		Constraints: constraints.MustParse("arch=amd64 mem=8G"),
	})
	err := s.State.SetSSHHostKeys(machine.MachineTag(), []string{"bam", "mam"})
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.ExportPartial(state.ExportConfig{
		SkipSSHHostKeys: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	keys := model.SSHHostKeys()
	c.Assert(keys, gc.HasLen, 0)
}

func (s *MigrationExportSuite) TestCloudImageMetadata(c *gc.C) {
	storageSize := uint64(3)
	attrs := cloudimagemetadata.MetadataAttributes{
		Stream:          "stream",
		Region:          "region-test",
		Version:         "22.04",
		Arch:            "arch",
		VirtType:        "virtType-test",
		RootStorageType: "rootStorageType-test",
		RootStorageSize: &storageSize,
		Source:          "test",
	}
	metadata := []cloudimagemetadata.Metadata{{attrs, 2, "1", 2}}

	err := s.State.CloudImageMetadataStorage.SaveMetadata(metadata)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	images := model.CloudImageMetadata()
	c.Assert(images, gc.HasLen, 1)
	image := images[0]
	c.Check(image.Stream(), gc.Equals, "stream")
	c.Check(image.Region(), gc.Equals, "region-test")
	c.Check(image.Version(), gc.Equals, "22.04")
	c.Check(image.Arch(), gc.Equals, "arch")
	c.Check(image.VirtType(), gc.Equals, "virtType-test")
	c.Check(image.RootStorageType(), gc.Equals, "rootStorageType-test")
	value, ok := image.RootStorageSize()
	c.Assert(ok, jc.IsTrue)
	c.Assert(value, gc.Equals, uint64(3))
	c.Check(image.Source(), gc.Equals, "test")
	c.Check(image.Priority(), gc.Equals, 2)
	c.Check(image.ImageId(), gc.Equals, "1")
	c.Check(image.DateCreated(), gc.Equals, int64(2))
}

func (s *MigrationExportSuite) TestCloudImageMetadataSkipped(c *gc.C) {
	storageSize := uint64(3)
	attrs := cloudimagemetadata.MetadataAttributes{
		Stream:          "stream",
		Region:          "region-test",
		Version:         "22.04",
		Arch:            "arch",
		VirtType:        "virtType-test",
		RootStorageType: "rootStorageType-test",
		RootStorageSize: &storageSize,
		Source:          "test",
	}
	metadata := []cloudimagemetadata.Metadata{{attrs, 2, "1", 2}}

	err := s.State.CloudImageMetadataStorage.SaveMetadata(metadata)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.ExportPartial(state.ExportConfig{
		SkipCloudImageMetadata: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	images := model.CloudImageMetadata()
	c.Assert(images, gc.HasLen, 0)
}

func (s *MigrationExportSuite) TestActions(c *gc.C) {
	machine := s.Factory.MakeMachine(c, &factory.MachineParams{
		Constraints: constraints.MustParse("arch=amd64 mem=8G"),
	})

	m, err := s.State.Model()
	c.Assert(err, jc.ErrorIsNil)

	operationID, err := m.EnqueueOperation("a test", 1)
	c.Assert(err, jc.ErrorIsNil)
	a, err := m.EnqueueAction(operationID, machine.MachineTag(), "foo", nil, true, "group", nil)
	c.Assert(err, jc.ErrorIsNil)
	a, err = a.Begin()
	c.Assert(err, jc.ErrorIsNil)
	err = a.Log("hello")
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)
	actions := model.Actions()
	c.Assert(actions, gc.HasLen, 1)
	action := actions[0]
	c.Check(action.Receiver(), gc.Equals, machine.Id())
	c.Check(action.Name(), gc.Equals, "foo")
	c.Check(action.Operation(), gc.Equals, operationID)
	c.Check(action.Parallel(), jc.IsTrue)
	c.Check(action.ExecutionGroup(), gc.Equals, "group")
	c.Check(action.Status(), gc.Equals, "running")
	c.Check(action.Message(), gc.Equals, "")
	logs := action.Logs()
	c.Assert(logs, gc.HasLen, 1)
	c.Assert(logs[0].Message(), gc.Equals, "hello")
	c.Assert(logs[0].Timestamp().IsZero(), jc.IsFalse)
}

func (s *MigrationExportSuite) TestActionsSkipped(c *gc.C) {
	machine := s.Factory.MakeMachine(c, &factory.MachineParams{
		Constraints: constraints.MustParse("arch=amd64 mem=8G"),
	})

	m, err := s.State.Model()
	c.Assert(err, jc.ErrorIsNil)

	operationID, err := s.Model.EnqueueOperation("a test", 1)
	c.Assert(err, jc.ErrorIsNil)
	_, err = m.EnqueueAction(operationID, machine.MachineTag(), "foo", nil, false, "", nil)
	c.Assert(err, jc.ErrorIsNil)
	model, err := s.State.ExportPartial(state.ExportConfig{
		SkipActions: true,
	}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	actions := model.Actions()
	c.Assert(actions, gc.HasLen, 0)
	operations := model.Operations()
	c.Assert(operations, gc.HasLen, 0)
}

func (s *MigrationExportSuite) TestOperations(c *gc.C) {
	m, err := s.State.Model()
	c.Assert(err, jc.ErrorIsNil)

	operationID, err := m.EnqueueOperation("a test", 2)
	c.Assert(err, jc.ErrorIsNil)
	err = m.FailOperationEnqueuing(operationID, "fail", 1)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)
	operations := model.Operations()
	c.Assert(operations, gc.HasLen, 1)
	op := operations[0]
	c.Check(op.Summary(), gc.Equals, "a test")
	c.Check(op.Fail(), gc.Equals, "fail")
	c.Check(op.Status(), gc.Equals, "pending")
	c.Check(op.SpawnedTaskCount(), gc.Equals, 1)
}

type goodToken struct{}

// Check implements leadership.Token
func (*goodToken) Check() error {
	return nil
}

func (s *MigrationExportSuite) TestVolumeAttachmentPlansLocalDisk(c *gc.C) {
	// Storage attachment plans aim to allow the development of external
	// storage backends like iSCSI to be attachable to machines. These
	// types of storage backends, need extra initialization in userspace
	// before they are usable. But this feature also aims to preserve the
	// old codepath, where no extra initialization is needed, and where the
	// providers are forced to guess the final device name that will appear on
	// the machine agents as a result of attaching a disk.
	// This test will ensure that given a local disk (the way it worked before
	// this feature was added), the information set by the provider in
	// VolumeAttachmentInfo is preserved. Different DeviceTypes may overwrite
	// this information, based on what they discover from the newly attached
	// device, after userspace initialization happens. For example, in the
	// case of an iSCSI device, there is no way to guess the final device name,
	// the WWN or any other kind of information about it, until we actually log
	// into the iSCSI target. That information is later sent by the machine
	// worker using SetVolumeAttachmentPlanBlockInfo.
	machine := s.Factory.MakeMachine(c, &factory.MachineParams{
		Volumes: []state.HostVolumeParams{{
			Volume:     state.VolumeParams{Size: 1234},
			Attachment: state.VolumeAttachmentParams{ReadOnly: true},
		}},
	})
	machineTag := machine.MachineTag()

	sb, err := state.NewStorageBackend(s.State)
	c.Assert(err, jc.ErrorIsNil)
	// We know that the first volume is called "0/0" as it is the first volume
	// (volumes use sequences), and it is bound to machine 0.
	volTag := names.NewVolumeTag("0/0")
	err = sb.SetVolumeInfo(volTag, state.VolumeInfo{
		HardwareId: "magic",
		WWN:        "drbr",
		Size:       1500,
		VolumeId:   "volume id",
		Persistent: true,
	})
	c.Assert(err, jc.ErrorIsNil)

	attachmentPlanInfo := state.VolumeAttachmentPlanInfo{
		DeviceType: storage.DeviceTypeLocal,
	}
	attachmentInfo := state.VolumeAttachmentInfo{
		DeviceName: "device name",
		DeviceLink: "device link",
		BusAddress: "bus address",
		ReadOnly:   true,
		PlanInfo:   &attachmentPlanInfo,
	}
	err = sb.SetVolumeAttachmentInfo(machineTag, volTag, attachmentInfo)
	c.Assert(err, jc.ErrorIsNil)

	err = sb.CreateVolumeAttachmentPlan(machineTag, volTag, attachmentPlanInfo)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	volumes := model.Volumes()
	c.Assert(volumes, gc.HasLen, 1)
	volume := volumes[0]

	c.Check(volume.Tag(), gc.Equals, volTag)
	c.Check(volume.Provisioned(), jc.IsTrue)
	c.Check(volume.Size(), gc.Equals, uint64(1500))
	c.Check(volume.Pool(), gc.Equals, "loop")
	c.Check(volume.HardwareID(), gc.Equals, "magic")
	c.Check(volume.WWN(), gc.Equals, "drbr")
	c.Check(volume.VolumeID(), gc.Equals, "volume id")
	c.Check(volume.Persistent(), jc.IsTrue)
	attachments := volume.Attachments()
	c.Assert(attachments, gc.HasLen, 1)
	attachment := attachments[0]
	c.Check(attachment.Host(), gc.Equals, machineTag)
	c.Check(attachment.Provisioned(), jc.IsTrue)
	c.Check(attachment.ReadOnly(), jc.IsTrue)
	c.Check(attachment.DeviceName(), gc.Equals, "device name")
	c.Check(attachment.DeviceLink(), gc.Equals, "device link")
	c.Check(attachment.BusAddress(), gc.Equals, "bus address")

	attachmentPlans := volume.AttachmentPlans()
	c.Assert(attachmentPlans, gc.HasLen, 1)

	plan := attachmentPlans[0]
	c.Check(plan.Machine(), gc.Equals, machineTag)
	c.Check(plan.VolumePlanInfo(), gc.NotNil)
	c.Check(storage.DeviceType(plan.VolumePlanInfo().DeviceType()), gc.Equals, storage.DeviceTypeLocal)
	c.Check(plan.VolumePlanInfo().DeviceAttributes(), gc.DeepEquals, map[string]string(nil))

	// This should all be empty
	planBlockDeviceInfo := plan.BlockDevice()
	c.Check(planBlockDeviceInfo.Name(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.Label(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.UUID(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.HardwareID(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.WWN(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.BusAddress(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.FilesystemType(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.MountPoint(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.Links(), gc.IsNil)
	c.Check(planBlockDeviceInfo.InUse(), gc.Equals, false)
	c.Check(planBlockDeviceInfo.Size(), gc.Equals, uint64(0))
}

func (s *MigrationExportSuite) TestVolumeAttachmentPlansISCSIDisk(c *gc.C) {
	// An ISCSI disk will also set the plan block info back in state. This means
	// that once the machine agent logs into the target, and a disk appears on
	// the system, the machine agent fetches all relevant info about that disk
	// and sends it back to state. This info will take precedence when identifying
	// the attached disk, as this info is observed on the machine itself, not
	// guessed by the provider.
	machine := s.Factory.MakeMachine(c, &factory.MachineParams{
		Volumes: []state.HostVolumeParams{{
			Volume:     state.VolumeParams{Size: 1234},
			Attachment: state.VolumeAttachmentParams{ReadOnly: true},
		}},
	})
	machineTag := machine.MachineTag()

	sb, err := state.NewStorageBackend(s.State)
	c.Assert(err, jc.ErrorIsNil)
	// We know that the first volume is called "0/0" as it is the first volume
	// (volumes use sequences), and it is bound to machine 0.
	volTag := names.NewVolumeTag("0/0")
	err = sb.SetVolumeInfo(volTag, state.VolumeInfo{
		HardwareId: "magic",
		WWN:        "drbr",
		Size:       1500,
		VolumeId:   "volume id",
		Persistent: true,
	})
	c.Assert(err, jc.ErrorIsNil)

	deviceAttrs := map[string]string{
		"iqn":         "bogusIQN",
		"address":     "192.168.1.1",
		"port":        "9999",
		"chap-user":   "example",
		"chap-secret": "supersecretpassword",
	}

	attachmentPlanInfo := state.VolumeAttachmentPlanInfo{
		DeviceType:       storage.DeviceTypeISCSI,
		DeviceAttributes: deviceAttrs,
	}
	attachmentInfo := state.VolumeAttachmentInfo{
		DeviceName: "device name",
		DeviceLink: "device link",
		BusAddress: "bus address",
		ReadOnly:   true,
		PlanInfo:   &attachmentPlanInfo,
	}
	err = sb.SetVolumeAttachmentInfo(machineTag, volTag, attachmentInfo)
	c.Assert(err, jc.ErrorIsNil)

	err = sb.CreateVolumeAttachmentPlan(machineTag, volTag, attachmentPlanInfo)
	c.Assert(err, jc.ErrorIsNil)

	deviceLinks := []string{"/dev/sdb", "/dev/mapper/testDevice"}

	blockInfo := state.BlockDeviceInfo{
		WWN:         "testWWN",
		DeviceLinks: deviceLinks,
		HardwareId:  "test-id",
	}

	err = sb.SetVolumeAttachmentPlanBlockInfo(machineTag, volTag, blockInfo)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	volumes := model.Volumes()
	c.Assert(volumes, gc.HasLen, 1)
	volume := volumes[0]

	c.Check(volume.Tag(), gc.Equals, volTag)
	c.Check(volume.Provisioned(), jc.IsTrue)
	c.Check(volume.Size(), gc.Equals, uint64(1500))
	c.Check(volume.Pool(), gc.Equals, "loop")
	c.Check(volume.HardwareID(), gc.Equals, "magic")
	c.Check(volume.WWN(), gc.Equals, "drbr")
	c.Check(volume.VolumeID(), gc.Equals, "volume id")
	c.Check(volume.Persistent(), jc.IsTrue)
	attachments := volume.Attachments()
	c.Assert(attachments, gc.HasLen, 1)
	attachment := attachments[0]
	c.Check(attachment.Host(), gc.Equals, machineTag)
	c.Check(attachment.Provisioned(), jc.IsTrue)
	c.Check(attachment.ReadOnly(), jc.IsTrue)
	c.Check(attachment.DeviceName(), gc.Equals, "device name")
	c.Check(attachment.DeviceLink(), gc.Equals, "device link")
	c.Check(attachment.BusAddress(), gc.Equals, "bus address")

	attachmentPlans := volume.AttachmentPlans()
	c.Assert(attachmentPlans, gc.HasLen, 1)

	plan := attachmentPlans[0]
	c.Check(plan.Machine(), gc.Equals, machineTag)
	c.Check(plan.VolumePlanInfo(), gc.NotNil)
	c.Check(storage.DeviceType(plan.VolumePlanInfo().DeviceType()), gc.Equals, storage.DeviceTypeISCSI)
	c.Check(plan.VolumePlanInfo().DeviceAttributes(), gc.DeepEquals, deviceAttrs)

	// This should all be empty
	planBlockDeviceInfo := plan.BlockDevice()
	c.Check(planBlockDeviceInfo.Name(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.Label(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.UUID(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.HardwareID(), gc.Equals, blockInfo.HardwareId)
	c.Check(planBlockDeviceInfo.WWN(), gc.Equals, blockInfo.WWN)
	c.Check(planBlockDeviceInfo.BusAddress(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.FilesystemType(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.MountPoint(), gc.Equals, "")
	c.Check(planBlockDeviceInfo.Links(), gc.DeepEquals, blockInfo.DeviceLinks)
	c.Check(planBlockDeviceInfo.InUse(), gc.Equals, false)
	c.Check(planBlockDeviceInfo.Size(), gc.Equals, uint64(0))

}

func (s *MigrationExportSuite) TestVolumes(c *gc.C) {
	machine := s.Factory.MakeMachine(c, &factory.MachineParams{
		Volumes: []state.HostVolumeParams{{
			Volume:     state.VolumeParams{Size: 1234},
			Attachment: state.VolumeAttachmentParams{ReadOnly: true},
		}, {
			Volume: state.VolumeParams{Size: 4000},
		}},
	})
	machineTag := machine.MachineTag()

	sb, err := state.NewStorageBackend(s.State)
	c.Assert(err, jc.ErrorIsNil)
	// We know that the first volume is called "0/0" as it is the first volume
	// (volumes use sequences), and it is bound to machine 0.
	volTag := names.NewVolumeTag("0/0")
	err = sb.SetVolumeInfo(volTag, state.VolumeInfo{
		HardwareId: "magic",
		WWN:        "drbr",
		Size:       1500,
		VolumeId:   "volume id",
		Persistent: true,
	})
	c.Assert(err, jc.ErrorIsNil)
	err = sb.SetVolumeAttachmentInfo(machineTag, volTag, state.VolumeAttachmentInfo{
		DeviceName: "device name",
		DeviceLink: "device link",
		BusAddress: "bus address",
		ReadOnly:   true,
	})
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	volumes := model.Volumes()
	c.Assert(volumes, gc.HasLen, 2)
	provisioned, notProvisioned := volumes[0], volumes[1]

	c.Check(provisioned.Tag(), gc.Equals, volTag)
	c.Check(provisioned.Provisioned(), jc.IsTrue)
	c.Check(provisioned.Size(), gc.Equals, uint64(1500))
	c.Check(provisioned.Pool(), gc.Equals, "loop")
	c.Check(provisioned.HardwareID(), gc.Equals, "magic")
	c.Check(provisioned.WWN(), gc.Equals, "drbr")
	c.Check(provisioned.VolumeID(), gc.Equals, "volume id")
	c.Check(provisioned.Persistent(), jc.IsTrue)
	attachments := provisioned.Attachments()
	c.Assert(attachments, gc.HasLen, 1)
	attachment := attachments[0]
	c.Check(attachment.Host(), gc.Equals, machineTag)
	c.Check(attachment.Provisioned(), jc.IsTrue)
	c.Check(attachment.ReadOnly(), jc.IsTrue)
	c.Check(attachment.DeviceName(), gc.Equals, "device name")
	c.Check(attachment.DeviceLink(), gc.Equals, "device link")
	c.Check(attachment.BusAddress(), gc.Equals, "bus address")

	attachmentPlans := provisioned.AttachmentPlans()
	c.Assert(attachmentPlans, gc.HasLen, 0)

	c.Check(notProvisioned.Tag(), gc.Equals, names.NewVolumeTag("0/1"))
	c.Check(notProvisioned.Provisioned(), jc.IsFalse)
	c.Check(notProvisioned.Size(), gc.Equals, uint64(4000))
	c.Check(notProvisioned.Pool(), gc.Equals, "loop")
	c.Check(notProvisioned.HardwareID(), gc.Equals, "")
	c.Check(notProvisioned.VolumeID(), gc.Equals, "")
	c.Check(notProvisioned.Persistent(), jc.IsFalse)
	attachments = notProvisioned.Attachments()
	c.Assert(attachments, gc.HasLen, 1)
	attachment = attachments[0]
	c.Check(attachment.Host(), gc.Equals, machineTag)
	c.Check(attachment.Provisioned(), jc.IsFalse)
	c.Check(attachment.ReadOnly(), jc.IsFalse)
	c.Check(attachment.DeviceName(), gc.Equals, "")
	c.Check(attachment.DeviceLink(), gc.Equals, "")
	c.Check(attachment.BusAddress(), gc.Equals, "")

	// Make sure there is a status.
	status := provisioned.Status()
	c.Check(status.Value(), gc.Equals, "pending")
}

func (s *MigrationExportSuite) TestFilesystems(c *gc.C) {
	machine := s.Factory.MakeMachine(c, &factory.MachineParams{
		Filesystems: []state.HostFilesystemParams{{
			Filesystem: state.FilesystemParams{Size: 1234},
			Attachment: state.FilesystemAttachmentParams{
				Location: "location",
				ReadOnly: true},
		}, {
			Filesystem: state.FilesystemParams{Size: 4000},
		}},
	})
	machineTag := machine.MachineTag()

	// We know that the first filesystem is called "0/0" as it is the first
	// filesystem (filesystems use sequences), and it is bound to machine 0.
	fsTag := names.NewFilesystemTag("0/0")
	sb, err := state.NewStorageBackend(s.State)
	c.Assert(err, jc.ErrorIsNil)
	err = sb.SetFilesystemInfo(fsTag, state.FilesystemInfo{
		Size:         1500,
		FilesystemId: "filesystem id",
	})
	c.Assert(err, jc.ErrorIsNil)
	err = sb.SetFilesystemAttachmentInfo(machineTag, fsTag, state.FilesystemAttachmentInfo{
		MountPoint: "/mnt/foo",
		ReadOnly:   true,
	})
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	filesystems := model.Filesystems()
	c.Assert(filesystems, gc.HasLen, 2)
	provisioned, notProvisioned := filesystems[0], filesystems[1]

	c.Check(provisioned.Tag(), gc.Equals, fsTag)
	c.Check(provisioned.Volume(), gc.Equals, names.VolumeTag{})
	c.Check(provisioned.Storage(), gc.Equals, names.StorageTag{})
	c.Check(provisioned.Provisioned(), jc.IsTrue)
	c.Check(provisioned.Size(), gc.Equals, uint64(1500))
	c.Check(provisioned.Pool(), gc.Equals, "rootfs")
	c.Check(provisioned.FilesystemID(), gc.Equals, "filesystem id")
	attachments := provisioned.Attachments()
	c.Assert(attachments, gc.HasLen, 1)
	attachment := attachments[0]
	c.Check(attachment.Host(), gc.Equals, machineTag)
	c.Check(attachment.Provisioned(), jc.IsTrue)
	c.Check(attachment.ReadOnly(), jc.IsTrue)
	c.Check(attachment.MountPoint(), gc.Equals, "/mnt/foo")

	c.Check(notProvisioned.Tag(), gc.Equals, names.NewFilesystemTag("0/1"))
	c.Check(notProvisioned.Volume(), gc.Equals, names.VolumeTag{})
	c.Check(notProvisioned.Storage(), gc.Equals, names.StorageTag{})
	c.Check(notProvisioned.Provisioned(), jc.IsFalse)
	c.Check(notProvisioned.Size(), gc.Equals, uint64(4000))
	c.Check(notProvisioned.Pool(), gc.Equals, "rootfs")
	c.Check(notProvisioned.FilesystemID(), gc.Equals, "")
	attachments = notProvisioned.Attachments()
	c.Assert(attachments, gc.HasLen, 1)
	attachment = attachments[0]
	c.Check(attachment.Host(), gc.Equals, machineTag)
	c.Check(attachment.Provisioned(), jc.IsFalse)
	c.Check(attachment.ReadOnly(), jc.IsFalse)
	c.Check(attachment.MountPoint(), gc.Equals, "")

	// Make sure there is a status.
	status := provisioned.Status()
	c.Check(status.Value(), gc.Equals, "pending")
}

func (s *MigrationExportSuite) TestStorage(c *gc.C) {
	_, u, storageTag := s.makeUnitWithStorage(c)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	apps := model.Applications()
	c.Assert(apps, gc.HasLen, 1)
	constraints := apps[0].StorageDirectives()
	c.Assert(constraints, gc.HasLen, 2)
	cons, found := constraints["data"]
	c.Assert(found, jc.IsTrue)
	c.Check(cons.Pool(), gc.Equals, "modelscoped")
	c.Check(cons.Size(), gc.Equals, uint64(0x400))
	c.Check(cons.Count(), gc.Equals, uint64(1))
	cons, found = constraints["allecto"]
	c.Assert(found, jc.IsTrue)
	c.Check(cons.Pool(), gc.Equals, "loop")
	c.Check(cons.Size(), gc.Equals, uint64(0x400))
	c.Check(cons.Count(), gc.Equals, uint64(0))

	storages := model.Storages()
	c.Assert(storages, gc.HasLen, 1)

	storage := storages[0]

	c.Check(storage.Tag(), gc.Equals, storageTag)
	c.Check(storage.Kind(), gc.Equals, "block")
	owner, err := storage.Owner()
	c.Check(err, jc.ErrorIsNil)
	c.Check(owner, gc.Equals, u.Tag())
	c.Check(storage.Name(), gc.Equals, "data")
	c.Check(storage.Attachments(), jc.DeepEquals, []names.UnitTag{
		u.UnitTag(),
	})
}

func (s *MigrationExportSuite) TestPayloads(c *gc.C) {
	unit := s.Factory.MakeUnit(c, nil)
	up, err := s.State.UnitPayloads(unit)
	c.Assert(err, jc.ErrorIsNil)
	original := payloads.Payload{
		PayloadClass: charm.PayloadClass{
			Name: "something",
			Type: "special",
		},
		ID:     "42",
		Status: "running",
		Labels: []string{"foo", "bar"},
	}
	err = up.Track(original)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	applications := model.Applications()
	c.Assert(applications, gc.HasLen, 1)

	units := applications[0].Units()
	c.Assert(units, gc.HasLen, 1)

	payloads := units[0].Payloads()
	c.Assert(payloads, gc.HasLen, 1)

	payload := payloads[0]
	c.Check(payload.Name(), gc.Equals, original.Name)
	c.Check(payload.Type(), gc.Equals, original.Type)
	c.Check(payload.RawID(), gc.Equals, original.ID)
	c.Check(payload.State(), gc.Equals, original.Status)
	c.Check(payload.Labels(), jc.DeepEquals, original.Labels)
}

func (s *MigrationExportSuite) TestResources(c *gc.C) {
	app := s.Factory.MakeApplication(c, nil)
	unit1 := s.Factory.MakeUnit(c, &factory.UnitParams{
		Application: app,
	})
	unit2 := s.Factory.MakeUnit(c, &factory.UnitParams{
		Application: app,
	})

	st := s.State.Resources(state.NewObjectStore(c, s.State.ModelUUID()))

	setUnitResource := func(u *state.Unit) {
		_, reader, err := st.OpenResourceForUniter(u.Name(), "spam")
		c.Assert(err, jc.ErrorIsNil)
		defer reader.Close()
		_, err = io.ReadAll(reader) // Need to read the content to set the resource for the unit.
		c.Assert(err, jc.ErrorIsNil)
	}

	const body = "ham"
	const bodySize = int64(len(body))

	// Initially set revision 1 for the application.
	res1 := s.newResource(c, app.Name(), "spam", 1, body)
	res1, err := st.SetResource(app.Name(), res1.Username, res1.Resource, bytes.NewBufferString(body), state.IncrementCharmModifiedVersion)
	c.Assert(err, jc.ErrorIsNil)

	// Unit 1 gets revision 1.
	setUnitResource(unit1)

	// Now set revision 2 for the application.
	res2 := s.newResource(c, app.Name(), "spam", 2, body)
	res2, err = st.SetResource(app.Name(), res2.Username, res2.Resource, bytes.NewBufferString(body), state.IncrementCharmModifiedVersion)
	c.Assert(err, jc.ErrorIsNil)

	// Unit 2 gets revision 2.
	setUnitResource(unit2)

	// Revision 3 is in the charmstore.
	res3 := resourcetesting.NewCharmResource(c, "spam", body)
	res3.Revision = 3
	err = st.SetCharmStoreResources(app.Name(), []charmresource.Resource{res3}, time.Now())
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	applications := model.Applications()
	c.Assert(applications, gc.HasLen, 1)
	exApp := applications[0]

	exResources := exApp.Resources()
	c.Assert(exResources, gc.HasLen, 1)

	exResource := exResources[0]
	c.Check(exResource.Name(), gc.Equals, "spam")

	checkExRevBase := func(exRev description.ResourceRevision, res charmresource.Resource) {
		c.Check(exRev.Revision(), gc.Equals, res.Revision)
		c.Check(exRev.Type(), gc.Equals, res.Type.String())
		c.Check(exRev.Path(), gc.Equals, res.Path)
		c.Check(exRev.Description(), gc.Equals, res.Description)
		c.Check(exRev.Origin(), gc.Equals, res.Origin.String())
		c.Check(exRev.FingerprintHex(), gc.Equals, res.Fingerprint.Hex())
		c.Check(exRev.Size(), gc.Equals, bodySize)
	}

	checkExRev := func(exRev description.ResourceRevision, res resources.Resource) {
		checkExRevBase(exRev, res.Resource)
		c.Check(exRev.Timestamp().UTC(), gc.Equals, truncateDBTime(res.Timestamp))
		c.Check(exRev.Username(), gc.Equals, res.Username)
	}

	checkExRev(exResource.ApplicationRevision(), res2)

	csRev := exResource.CharmStoreRevision()
	checkExRevBase(csRev, res3)
	// These shouldn't be set for charmstore only revisions.
	c.Check(csRev.Timestamp(), gc.Equals, time.Time{})
	c.Check(csRev.Username(), gc.Equals, "")

	// Units
	units := exApp.Units()
	c.Assert(units, gc.HasLen, 2)

	checkUnitRes := func(exUnit description.Unit, unit *state.Unit, res resources.Resource) {
		c.Check(exUnit.Name(), gc.Equals, unit.Name())
		exResources := exUnit.Resources()
		c.Assert(exResources, gc.HasLen, 1)
		exRes := exResources[0]
		c.Check(exRes.Name(), gc.Equals, "spam")
		checkExRev(exRes.Revision(), res)
	}
	checkUnitRes(units[0], unit1, res1)
	checkUnitRes(units[1], unit2, res2)
}

func (s *MigrationExportSuite) newResource(c *gc.C, appName, name string, revision int, body string) resources.Resource {
	opened := resourcetesting.NewResource(c, nil, name, appName, body)
	res := opened.Resource
	res.Revision = revision
	return res
}

func (s *MigrationExportSuite) TestRemoteApplications(c *gc.C) {
	mac, err := newMacaroon("apimac")
	c.Assert(err, gc.IsNil)
	_, err = s.State.AddRemoteApplication(state.AddRemoteApplicationParams{
		Name:        "gravy-rainbow",
		URL:         "me/model.rainbow",
		SourceModel: s.Model.ModelTag(),
		Token:       "charisma",
		OfferUUID:   "offer-uuid",
		Endpoints: []charm.Relation{{
			Interface: "mysql",
			Name:      "db",
			Role:      charm.RoleProvider,
			Scope:     charm.ScopeGlobal,
		}, {
			Interface: "mysql-root",
			Name:      "db-admin",
			Limit:     5,
			Role:      charm.RoleProvider,
			Scope:     charm.ScopeGlobal,
		}, {
			Interface: "logging",
			Name:      "logging",
			Role:      charm.RoleProvider,
			Scope:     charm.ScopeGlobal,
		}},
		// Macaroon not exported.
		Macaroon: mac,
	})
	c.Assert(err, jc.ErrorIsNil)
	state.AddTestingApplication(c, s.State, s.objectStore, "wordpress", state.AddTestingCharm(c, s.State, "wordpress"))
	eps, err := s.State.InferEndpoints("gravy-rainbow", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	_, err = s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(model.RemoteApplications(), gc.HasLen, 1)
	app := model.RemoteApplications()[0]
	c.Check(app.Tag(), gc.Equals, names.NewApplicationTag("gravy-rainbow"))
	c.Check(app.Name(), gc.Equals, "gravy-rainbow")
	c.Check(app.OfferUUID(), gc.Equals, "offer-uuid")
	c.Check(app.URL(), gc.Equals, "me/model.rainbow")
	c.Check(app.SourceModelTag(), gc.Equals, s.Model.ModelTag())
	c.Check(app.IsConsumerProxy(), jc.IsFalse)

	c.Assert(app.Endpoints(), gc.HasLen, 3)
	ep := app.Endpoints()[0]
	c.Check(ep.Name(), gc.Equals, "db")
	c.Check(ep.Interface(), gc.Equals, "mysql")
	c.Check(ep.Role(), gc.Equals, "provider")
	ep = app.Endpoints()[1]
	c.Check(ep.Name(), gc.Equals, "db-admin")
	c.Check(ep.Interface(), gc.Equals, "mysql-root")
	c.Check(ep.Role(), gc.Equals, "provider")
	ep = app.Endpoints()[2]
	c.Check(ep.Name(), gc.Equals, "logging")
	c.Check(ep.Interface(), gc.Equals, "logging")
	c.Check(ep.Role(), gc.Equals, "provider")

	c.Assert(model.Relations(), gc.HasLen, 1)
	rel := model.Relations()[0]
	c.Assert(rel.Key(), gc.Equals, "wordpress:db gravy-rainbow:db")
}

func (s *MigrationExportSuite) TestModelStatus(c *gc.C) {
	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	c.Check(model.Status().Value(), gc.Equals, "available")
	c.Check(model.StatusHistory(), gc.HasLen, 1)
}

func (s *MigrationExportSuite) TestTooManyStatusHistories(c *gc.C) {
	// Check that we cap the history entries at 20.
	machine := s.Factory.MakeMachine(c, nil)
	s.primeStatusHistory(c, machine, status.Started, 21)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(model.Machines(), gc.HasLen, 1)
	history := model.Machines()[0].StatusHistory()
	c.Assert(history, gc.HasLen, 20)
	s.checkStatusHistory(c, history, status.Started)
}

func (s *MigrationExportSuite) TestRelationWithNoStatus(c *gc.C) {
	// Importing from a model from before relations had status will
	// mean that there's no status to export - don't fail to export if
	// there isn't a status for a relation.
	wordpress := state.AddTestingApplication(c, s.State, s.objectStore, "wordpress", state.AddTestingCharm(c, s.State, "wordpress"))
	mysql := state.AddTestingApplication(c, s.State, s.objectStore, "mysql", state.AddTestingCharm(c, s.State, "mysql"))
	// InferEndpoints will always return provider, requirer
	eps, err := s.State.InferEndpoints("mysql", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	rel, err := s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)
	wordpress0 := s.Factory.MakeUnit(c, &factory.UnitParams{Application: wordpress})
	mysql0 := s.Factory.MakeUnit(c, &factory.UnitParams{Application: mysql})

	ru, err := rel.Unit(wordpress0)
	c.Assert(err, jc.ErrorIsNil)
	wordpressSettings := map[string]interface{}{
		"name": "wordpress/0",
	}
	err = ru.EnterScope(wordpressSettings)
	c.Assert(err, jc.ErrorIsNil)

	ru, err = rel.Unit(mysql0)
	c.Assert(err, jc.ErrorIsNil)
	mysqlSettings := map[string]interface{}{
		"name": "mysql/0",
	}
	err = ru.EnterScope(mysqlSettings)
	c.Assert(err, jc.ErrorIsNil)

	state.RemoveRelationStatus(c, rel)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	rels := model.Relations()
	c.Assert(rels, gc.HasLen, 1)
	c.Assert(rels[0].Status(), gc.IsNil)
}

func (s *MigrationExportSuite) TestRemoteRelationSettingsForUnitsInCMR(c *gc.C) {
	mac, err := newMacaroon("apimac")
	c.Assert(err, gc.IsNil)

	_, err = s.State.AddRemoteApplication(state.AddRemoteApplicationParams{
		Name:        "gravy-rainbow",
		URL:         "me/model.rainbow",
		SourceModel: s.Model.ModelTag(),
		Token:       "charisma",
		OfferUUID:   "offer-uuid",
		Endpoints: []charm.Relation{{
			Interface: "mysql",
			Name:      "db",
			Role:      charm.RoleProvider,
			Scope:     charm.ScopeGlobal,
		}},
		// Macaroon not exported.
		Macaroon: mac,
	})
	c.Assert(err, jc.ErrorIsNil)

	wordpress := state.AddTestingApplication(c, s.State, s.objectStore, "wordpress", state.AddTestingCharm(c, s.State, "wordpress"))
	eps, err := s.State.InferEndpoints("gravy-rainbow", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	rel, err := s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)

	wordpress0 := s.Factory.MakeUnit(c, &factory.UnitParams{Application: wordpress})
	localRU, err := rel.Unit(wordpress0)
	c.Assert(err, jc.ErrorIsNil)

	wordpressSettings := map[string]interface{}{"name": "wordpress/0"}
	err = localRU.EnterScope(wordpressSettings)
	c.Assert(err, jc.ErrorIsNil)

	remoteRU, err := rel.RemoteUnit("gravy-rainbow/0")
	c.Assert(err, jc.ErrorIsNil)

	gravySettings := map[string]interface{}{"name": "gravy-rainbow/0"}
	err = remoteRU.EnterScope(gravySettings)
	c.Assert(err, jc.ErrorIsNil)

	model, err := s.State.Export(map[string]string{}, state.NewObjectStore(c, s.State.ModelUUID()))
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(model.Relations(), gc.HasLen, 1)
	exRel := model.Relations()[0]
	c.Assert(exRel.Key(), gc.Equals, "wordpress:db gravy-rainbow:db")
	c.Assert(exRel.Endpoints(), gc.HasLen, 2)

	for _, exEp := range exRel.Endpoints() {
		if exEp.ApplicationName() == "wordpress" {
			c.Check(exEp.Settings(wordpress0.Name()), jc.DeepEquals, wordpressSettings)
		} else {
			c.Check(exEp.ApplicationName(), gc.Equals, "gravy-rainbow")
			c.Check(exEp.Settings("gravy-rainbow/0"), jc.DeepEquals, gravySettings)
		}
	}
}
