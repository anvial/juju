// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state

import (
	"github.com/juju/clock"
	"github.com/juju/juju/core/database"
	"github.com/juju/juju/core/logger"
	"github.com/juju/juju/domain"
)

// InsertIAASUnitState represents the minium state required to insert
// an IAAS unit. Splitting this out, allows for sharing with the
// relation domain to facilitate subordinate unit creation.
type InsertIAASUnitState struct {
	*domain.StateBase
	clock  clock.Clock
	logger logger.Logger
}

// NewInsertIAASUnitState returns a new insert IAAS unit state reference.
func NewInsertIAASUnitState(factory database.TxnRunnerFactory, clock clock.Clock, logger logger.Logger) *InsertIAASUnitState {
	return &InsertIAASUnitState{
		StateBase: domain.NewStateBase(factory),
		clock:     clock,
		logger:    logger,
	}
}
