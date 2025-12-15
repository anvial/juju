// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state

import (
	"context"
	"database/sql"
	"testing"

	"github.com/canonical/sqlair"
	"github.com/juju/tc"

	"github.com/juju/juju/domain/unitstate"
)

type commitHookSuite struct {
	baseSuite
}

func TestCommitHookSuite(t *testing.T) {
	tc.Run(t, &commitHookSuite{})
}

func (s *commitHookSuite) TestCommitHookChanges(c *tc.C) {
	// Arrange
	arg := unitstate.CommitHookChangesArg{
		UnitName:           s.unitName,
		UnitUUID:           s.unitUUID,
		UpdateNetworkInfo:  true,
		RelationSettings:   nil,
		OpenPorts:          nil,
		ClosePorts:         nil,
		CharmState:         nil,
		SecretCreates:      nil,
		TrackLatestSecrets: nil,
		SecretUpdates:      nil,
		SecretGrants:       nil,
		SecretRevokes:      nil,
		SecretDeletes:      nil,
	}

	// Act
	err := s.state.CommitHookChanges(c.Context(), arg)

	// Assert
	c.Assert(err, tc.IsNil)
}

func (s *commitHookSuite) TestUpdateCharmState(c *tc.C) {
	ctx := c.Context()

	// Arrange
	// Set some initial state. This should be overwritten.
	s.addUnitStateCharm(c, "one-key", "one-val")

	expState := map[string]string{
		"two-key":   "two-val",
		"three-key": "three-val",
	}

	// Act
	err := s.TxnRunner().Txn(ctx, func(ctx context.Context, tx *sqlair.TX) error {
		return s.state.updateCharmState(ctx, tx, s.unitUUID, &expState)
	})
	c.Assert(err, tc.ErrorIsNil)

	// Assert
	gotState := make(map[string]string)
	err = s.TxnRunner().StdTxn(ctx, func(ctx context.Context, tx *sql.Tx) error {
		q := "SELECT key, value FROM unit_state_charm WHERE unit_uuid = ?"
		rows, err := tx.QueryContext(ctx, q, s.unitUUID)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var k, v string
			if err := rows.Scan(&k, &v); err != nil {
				return err
			}
			gotState[k] = v
		}
		return rows.Err()
	})
	c.Assert(err, tc.ErrorIsNil)

	c.Check(gotState, tc.DeepEquals, expState)
}

func (s *commitHookSuite) TestUpdateCharmStateEmpty(c *tc.C) {
	ctx := c.Context()

	// Act - use a bad unit uuid to ensure the test fails if setUnitStateCharm
	// is called.
	err := s.TxnRunner().Txn(ctx, func(ctx context.Context, tx *sqlair.TX) error {
		return s.state.updateCharmState(ctx, tx, "bad-unit-uuid", nil)
	})

	// Assert
	c.Assert(err, tc.ErrorIsNil)
}
