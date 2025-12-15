// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state

import (
	"context"

	"github.com/canonical/sqlair"

	"github.com/juju/juju/core/network"
	"github.com/juju/juju/domain/unitstate"
	"github.com/juju/juju/internal/errors"
)

// CommitHookChanges persists a set of changes after a hook successfully
// completes and executes them in a single transaction.
func (st *State) CommitHookChanges(ctx context.Context, arg unitstate.CommitHookChangesArg) error {
	db, err := st.DB(ctx)
	if err != nil {
		return errors.Capture(err)
	}

	return db.Txn(ctx, func(ctx context.Context, tx *sqlair.TX) error {
		if err := st.updateNetworkInfo(ctx, tx, arg.UpdateNetworkInfo); err != nil {
			return errors.Errorf("update network info: %v", err)
		}

		if err := st.updateRelationSettings(ctx, tx, arg.RelationSettings); err != nil {
			return errors.Errorf("update relation settings: %v", err)
		}

		if err := st.updatePorts(ctx, tx, arg.OpenPorts, arg.ClosePorts); err != nil {
			return errors.Errorf("update ports: %v", err)
		}

		if err := st.updateCharmState(ctx, tx, arg.UnitUUID, arg.CharmState); err != nil {
			return errors.Errorf("update charm state: %v", err)
		}

		if err := st.createSecrets(ctx, tx, arg.SecretCreates); err != nil {
			return errors.Errorf("create secrets: %v", err)
		}

		if err := st.updateSecrets(ctx, tx, arg.SecretUpdates); err != nil {
			return errors.Errorf("update secrets: %v", err)
		}

		if err := st.grantSecretsAccess(ctx, tx, arg.SecretGrants); err != nil {
			return errors.Errorf("grant secrets access: %v", err)
		}

		if err := st.revokeSecretsAccess(ctx, tx, arg.SecretRevokes); err != nil {
			return errors.Errorf("revoke secrets access: %v", err)
		}

		if err := st.deleteSecrets(ctx, tx, arg.SecretDeletes); err != nil {
			return errors.Errorf("delete secrets: %v", err)
		}

		if err := st.trackSecrets(ctx, tx, arg.TrackLatestSecrets); err != nil {
			return errors.Errorf("track latest secrets: %v", err)
		}

		// TODO: (hml) 10-Dec-2025
		// Implement storage
		return nil
	})
}

func (st *State) updateNetworkInfo(ctx context.Context, tx *sqlair.TX, info bool) error {
	return nil
}

func (st *State) updateRelationSettings(ctx context.Context, tx *sqlair.TX, settings []unitstate.RelationSettings) error {
	return nil
}

func (st *State) updatePorts(ctx context.Context, tx *sqlair.TX, openPorts network.GroupedPortRanges, closePorts network.GroupedPortRanges) error {
	return nil
}

func (st *State) updateCharmState(ctx context.Context, tx *sqlair.TX, unit string, charmState *map[string]string) error {
	if charmState == nil {
		return nil
	}
	st.logger.Criticalf(ctx, "charm state: %v", charmState)
	return st.setUnitStateCharm(ctx, tx, unitUUID{UUID: unit}, *charmState)
}

func (st *State) createSecrets(ctx context.Context, tx *sqlair.TX, creates []unitstate.CreateSecretArg) error {
	return nil
}

func (st *State) updateSecrets(ctx context.Context, tx *sqlair.TX, updates []unitstate.UpdateSecretArg) error {
	return nil
}

func (st *State) grantSecretsAccess(ctx context.Context, tx *sqlair.TX, grants []unitstate.GrantRevokeSecretArg) error {
	return nil
}

func (st *State) revokeSecretsAccess(ctx context.Context, tx *sqlair.TX, revokes []unitstate.GrantRevokeSecretArg) error {
	return nil
}

func (st *State) deleteSecrets(ctx context.Context, tx *sqlair.TX, deletes []unitstate.DeleteSecretArg) error {
	return nil
}

func (st *State) trackSecrets(ctx context.Context, tx *sqlair.TX, secrets []string) error {
	return nil
}
