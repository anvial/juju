// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/sqlair"
	"github.com/juju/errors"

	coreDB "github.com/juju/juju/core/database"
	"github.com/juju/juju/core/logger"
	"github.com/juju/juju/core/network"
	"github.com/juju/juju/domain"
	networkerrors "github.com/juju/juju/domain/network/errors"
	"github.com/juju/juju/internal/database"
)

// State represents a type for interacting with the underlying state.
type State struct {
	*domain.StateBase
	logger logger.Logger
}

// NewState returns a new State for interacting with the underlying state.
func NewState(factory coreDB.TxnRunnerFactory, logger logger.Logger) *State {
	return &State{
		StateBase: domain.NewStateBase(factory),
		logger:    logger,
	}
}

// AddSpace creates and returns a new space.
func (st *State) AddSpace(
	ctx context.Context,
	uuid string,
	name string,
	providerID network.Id,
	subnetIDs []string,
) error {
	db, err := st.DB()
	if err != nil {
		return errors.Trace(domain.CoerceError(err))
	}

	insertSpaceStmt, err := sqlair.Prepare(`
INSERT INTO space (uuid, name) 
VALUES ($Space.uuid, $Space.name)`, Space{})
	if err != nil {
		return errors.Trace(err)
	}

	insertProviderStmt, err := sqlair.Prepare(`
INSERT INTO provider_space (provider_id, space_uuid)
VALUES ($ProviderSpace.provider_id, $ProviderSpace.space_uuid)`, ProviderSpace{})
	if err != nil {
		return errors.Trace(err)
	}

	subnetIDsInS := sqlair.S{}
	for _, sid := range subnetIDs {
		subnetIDsInS = append(subnetIDsInS, sid)
	}

	err = db.Txn(ctx, func(ctx context.Context, tx *sqlair.TX) error {
		if err := tx.Query(ctx, insertSpaceStmt, Space{UUID: uuid, Name: name}).Run(); err != nil {
			if database.IsErrConstraintUnique(err) {
				return fmt.Errorf("inserting space uuid %q into space table: %w with err: %w", uuid, networkerrors.ErrSpaceAlreadyExists, err)
			}
			return errors.Annotatef(err, "inserting space uuid %q into space table", uuid)
		}
		if providerID != "" {
			if err := tx.Query(ctx, insertProviderStmt, ProviderSpace{ProviderID: providerID, SpaceUUID: uuid}).Run(); err != nil {
				return errors.Annotatef(err, "inserting provider id %q into provider_space table", providerID)
			}
		}

		// Update all subnets (including their fan overlays) to include
		// the space uuid.
		for _, subnetID := range subnetIDs {
			if err := st.updateSubnetSpaceID(ctx, tx, subnetID, uuid); err != nil {
				return errors.Annotatef(err, "updating subnet %q using space uuid %q", subnetID, uuid)
			}
		}
		return nil
	})
	return errors.Trace(domain.CoerceError(err))
}

const retrieveSpacesStmt = `
SELECT     
    space.uuid                           AS &SpaceSubnetRow.uuid,
    space.name                           AS &SpaceSubnetRow.name,
    provider_space.provider_id           AS &SpaceSubnetRow.provider_id,
    subnet.uuid                          AS &SpaceSubnetRow.subnet_uuid,
    subnet.cidr                          AS &SpaceSubnetRow.subnet_cidr,
    subnet.vlan_tag                      AS &SpaceSubnetRow.subnet_vlan_tag,
    provider_subnet.provider_id          AS &SpaceSubnetRow.subnet_provider_id,
    provider_network.provider_network_id AS &SpaceSubnetRow.subnet_provider_network_id,
    availability_zone.name               AS &SpaceSubnetRow.subnet_az
FROM space 
    LEFT JOIN provider_space
    ON space.uuid = provider_space.space_uuid
    LEFT JOIN subnet   
    ON space.uuid = subnet.space_uuid
    LEFT JOIN provider_subnet
    ON subnet.uuid = provider_subnet.subnet_uuid
    LEFT JOIN provider_network_subnet
    ON subnet.uuid = provider_network_subnet.subnet_uuid
    LEFT JOIN provider_network
    ON provider_network_subnet.provider_network_uuid = provider_network.uuid
    LEFT JOIN availability_zone_subnet
    ON availability_zone_subnet.subnet_uuid = subnet.uuid
    LEFT JOIN availability_zone
    ON availability_zone_subnet.availability_zone_uuid = availability_zone.uuid`

// GetSpace returns the space by UUID.
func (st *State) GetSpace(
	ctx context.Context,
	uuid string,
) (*network.SpaceInfo, error) {
	db, err := st.DB()
	if err != nil {
		return nil, errors.Trace(err)
	}

	// Append the space uuid condition to the query only if it's passed to the function.
	q := retrieveSpacesStmt + " WHERE space.uuid = $M.id;"

	spacesStmt, err := sqlair.Prepare(q, SpaceSubnetRow{}, sqlair.M{})
	if err != nil {
		return nil, errors.Annotatef(err, "preparing %q", q)
	}

	var spaceRows SpaceSubnetRows
	if err := db.Txn(ctx, func(ctx context.Context, tx *sqlair.TX) error {
		err := tx.Query(ctx, spacesStmt, sqlair.M{"id": uuid}).GetAll(&spaceRows)
		if err != nil && !errors.Is(err, sqlair.ErrNoRows) {
			return errors.Annotatef(err, "retrieving space %q", uuid)
		}

		return nil
	}); errors.Is(err, sqlair.ErrNoRows) || len(spaceRows) == 0 {
		return nil, fmt.Errorf("space not found with %s: %w", uuid, networkerrors.ErrSpaceNotFound)
	} else if err != nil {
		return nil, errors.Annotate(err, "querying spaces")
	}

	return &spaceRows.ToSpaceInfos()[0], nil
}

// GetSpaceByName returns the space by name.
func (st *State) GetSpaceByName(
	ctx context.Context,
	name string,
) (*network.SpaceInfo, error) {
	db, err := st.DB()
	if err != nil {
		return nil, errors.Trace(err)
	}

	// Append the space.name condition to the query.
	q := retrieveSpacesStmt + " WHERE space.name = $M.name;"

	s, err := sqlair.Prepare(q, SpaceSubnetRow{}, sqlair.M{})
	if err != nil {
		return nil, errors.Annotatef(err, "preparing %q", q)
	}

	var rows SpaceSubnetRows
	if err := db.Txn(ctx, func(ctx context.Context, tx *sqlair.TX) error {
		return errors.Trace(tx.Query(ctx, s, sqlair.M{"name": name}).GetAll(&rows))
	}); errors.Is(err, sqlair.ErrNoRows) || len(rows) == 0 {
		return nil, fmt.Errorf("space not found with %s: %w", name, networkerrors.ErrSpaceNotFound)
	} else if err != nil {
		return nil, errors.Annotate(domain.CoerceError(err), "querying spaces by name")
	}

	return &rows.ToSpaceInfos()[0], nil
}

// GetAllSpaces returns all spaces for the model.
func (st *State) GetAllSpaces(
	ctx context.Context,
) (network.SpaceInfos, error) {

	db, err := st.DB()
	if err != nil {
		return nil, errors.Trace(err)
	}

	s, err := sqlair.Prepare(retrieveSpacesStmt, SpaceSubnetRow{})
	if err != nil {
		return nil, errors.Annotatef(err, "preparing %q", retrieveSpacesStmt)
	}

	var rows SpaceSubnetRows
	if err := db.Txn(ctx, func(ctx context.Context, tx *sqlair.TX) error {
		return errors.Trace(tx.Query(ctx, s).GetAll(&rows))
	}); errors.Is(err, sqlair.ErrNoRows) || len(rows) == 0 {
		return nil, nil
	} else if err != nil {
		st.logger.Errorf("querying all spaces, %v", err)
		return nil, errors.Annotate(domain.CoerceError(err), "querying all spaces")
	}

	return rows.ToSpaceInfos(), nil
}

// UpdateSpace updates the space identified by the passed uuid.
func (st *State) UpdateSpace(
	ctx context.Context,
	uuid string,
	name string,
) error {
	db, err := st.DB()
	if err != nil {
		return errors.Trace(domain.CoerceError(err))
	}

	q := `
UPDATE space
SET    name = ?
WHERE  uuid = ?;`
	err = db.StdTxn(ctx, func(ctx context.Context, tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, q, name, uuid)
		if err != nil {
			return errors.Annotatef(err, "updating space %q with name %q", uuid, name)
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return errors.Trace(err)
		}
		if affected == 0 {
			return fmt.Errorf("space not found with %s: %w", uuid, networkerrors.ErrSpaceNotFound)
		}
		return nil
	})
	return domain.CoerceError(err)
}

// DeleteSpace deletes the space identified by the passed uuid.
func (st *State) DeleteSpace(
	ctx context.Context,
	uuid string,
) error {
	db, err := st.DB()
	if err != nil {
		return errors.Trace(err)
	}

	deleteSpaceStmt := "DELETE FROM space WHERE uuid = ?;"
	deleteProviderSpaceStmt := "DELETE FROM provider_space WHERE space_uuid = ?;"
	updateSubnetSpaceUUIDStmt := "UPDATE subnet SET space_uuid = ? WHERE space_uuid = ?;"

	err = db.StdTxn(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, deleteProviderSpaceStmt, uuid); err != nil {
			return errors.Annotatef(err, "removing space %q from the provider_space table", uuid)
		}

		if _, err := tx.ExecContext(ctx, updateSubnetSpaceUUIDStmt, network.AlphaSpaceId, uuid); err != nil {
			return errors.Annotatef(err, "updating subnet table by removing the space %q", uuid)
		}

		delSpaceResult, err := tx.ExecContext(ctx, deleteSpaceStmt, uuid)
		if err != nil {
			return errors.Annotatef(err, "removing space %q", uuid)
		}
		delSpaceAffected, err := delSpaceResult.RowsAffected()
		if err != nil {
			return errors.Trace(err)
		}
		if delSpaceAffected != 1 {
			return networkerrors.ErrSpaceNotFound
		}

		return nil
	})
	return domain.CoerceError(err)
}
