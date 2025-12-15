// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package unitstate

import (
	"testing"

	"github.com/juju/tc"

	"github.com/juju/juju/core/network"
	"github.com/juju/juju/domain/secret"
)

type commitHookChangesArgSuite struct{}

func TestCommitHookChangesArgSuite(t *testing.T) {
	tc.Run(t, &commitHookChangesArgSuite{})
}

func (s *commitHookChangesArgSuite) TestValidateNoChanges(c *tc.C) {
	hasChanges, err := CommitHookChangesArg{
		UnitName: "testing/0",
		UnitUUID: "unit-uuid",
	}.Validate()

	c.Check(err, tc.ErrorIsNil)
	c.Check(hasChanges, tc.Equals, false)
}

func (s *commitHookChangesArgSuite) TestValidateCreateSecret(c *tc.C) {
	hasChanges, err := CommitHookChangesArg{
		UnitName:      "testing/0",
		UnitUUID:      "unit-uuid",
		SecretCreates: []CreateSecretArg{{CreateCharmSecretParams: secret.CreateCharmSecretParams{}}},
	}.Validate()

	c.Check(err, tc.ErrorIsNil)
	c.Check(hasChanges, tc.Equals, true)
}

func (s *commitHookChangesArgSuite) TestValidateErrorNoUnitName(c *tc.C) {
	hasChanges, err := CommitHookChangesArg{
		UnitUUID: "unit-uuid",
	}.Validate()

	c.Check(err, tc.ErrorMatches, "unit name is required")
	c.Check(hasChanges, tc.Equals, false)
}

func (s *commitHookChangesArgSuite) TestValidateErrorNoUnitUUID(c *tc.C) {
	hasChanges, err := CommitHookChangesArg{
		UnitName: "testing/0",
	}.Validate()

	c.Check(err, tc.ErrorMatches, "unit uuid is required")
	c.Check(hasChanges, tc.Equals, false)
}

func (s *commitHookChangesArgSuite) TestValidateErrorInvalidOpenPort(c *tc.C) {
	hasChanges, err := CommitHookChangesArg{
		UnitName: "testing/0",
		UnitUUID: "unit-uuid",
		OpenPorts: map[string][]network.PortRange{
			"endpoint": {{Protocol: "failme"}},
		},
	}.Validate()

	c.Check(err, tc.ErrorMatches, ".*open port is invalid.*")
	c.Check(hasChanges, tc.Equals, true)
}

func (s *commitHookChangesArgSuite) TestValidateErrorInvalidClosePort(c *tc.C) {
	hasChanges, err := CommitHookChangesArg{
		UnitName: "testing/0",
		UnitUUID: "unit-uuid",
		ClosePorts: map[string][]network.PortRange{
			"endpoint": {{Protocol: "failme"}},
		},
	}.Validate()

	c.Check(err, tc.ErrorMatches, ".*close port is invalid.*")
	c.Check(hasChanges, tc.Equals, true)
}

func (s *commitHookChangesArgSuite) TestRequiresLeadershipTrueCreateSecret(c *tc.C) {
	requiresLeadership := CommitHookChangesArg{
		UnitName:      "testing/0",
		UnitUUID:      "unit-uuid",
		SecretCreates: []CreateSecretArg{{CreateCharmSecretParams: secret.CreateCharmSecretParams{}}},
	}.RequiresLeadership()

	c.Check(requiresLeadership, tc.Equals, true)
}

func (s *commitHookChangesArgSuite) TestRequiresLeadershipTrueApplicationSettings(c *tc.C) {
	requiresLeadership := CommitHookChangesArg{
		UnitName: "testing/0",
		UnitUUID: "unit-uuid",
		RelationSettings: []RelationSettings{{
			ApplicationSettings: map[string]string{"key": "value"},
		}},
	}.RequiresLeadership()

	c.Check(requiresLeadership, tc.Equals, true)
}

func (s *commitHookChangesArgSuite) TestRequiresLeadershipFalseUnitSettings(c *tc.C) {
	requiresLeadership := CommitHookChangesArg{
		UnitName: "testing/0",
		UnitUUID: "unit-uuid",
		RelationSettings: []RelationSettings{{
			Settings: map[string]string{"key": "value"},
		}},
	}.RequiresLeadership()

	c.Check(requiresLeadership, tc.Equals, false)
}

func (s *commitHookChangesArgSuite) TestRequiresLeadershipFalseOpenPorts(c *tc.C) {
	requiresLeadership := CommitHookChangesArg{
		UnitName: "testing/0",
		UnitUUID: "unit-uuid",
		OpenPorts: map[string][]network.PortRange{
			"endpoint": {{Protocol: "failme"}},
		},
	}.RequiresLeadership()

	c.Check(requiresLeadership, tc.Equals, false)
}
