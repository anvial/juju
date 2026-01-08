// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package service

import (
	"context"

	"github.com/juju/juju/core/logger"
	"github.com/juju/juju/core/trace"
	"github.com/juju/juju/domain/application/charm"
	"github.com/juju/juju/domain/crossmodelrelation"
	"github.com/juju/juju/internal/errors"
)

// ModelMigrationState describes persistence methods for migration of cross
// model relations in the model database.
type ModelMigrationState interface {
	// ImportOffers adds offers being migrated to the current model.
	ImportOffers(context.Context, []crossmodelrelation.OfferImport) error

	// ImportRemoteApplications adds remote application offerers being migrated
	// to the current model.
	ImportRemoteApplications(context.Context, []crossmodelrelation.RemoteApplicationImport) error
}

// MigrationService provides the API for model migration actions within
// the cross model relation domain.
type MigrationService struct {
	modelState ModelMigrationState
	logger     logger.Logger
}

// MigrationService returns a new service reference wrapping the input state
// for migration.
func NewMigrationService(
	modelState ModelMigrationState,
	logger logger.Logger,
) *MigrationService {
	return &MigrationService{
		modelState: modelState,
		logger:     logger,
	}
}

// ImportOffers adds offers being migrated to the current model.
func (s *MigrationService) ImportOffers(ctx context.Context, imports []crossmodelrelation.OfferImport) error {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	return errors.Capture(s.modelState.ImportOffers(ctx, imports))
}

// ImportRemoteApplications adds remote application offerers being migrated to
// the current model. These are applications that this model is consuming from
// other models.
func (s *MigrationService) ImportRemoteApplications(ctx context.Context, imports []crossmodelrelation.RemoteApplicationImport) error {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	// Build synthetic charms for each remote application in the service layer.
	importsWithCharms := make([]crossmodelrelation.RemoteApplicationImport, len(imports))
	for i, imp := range imports {
		importsWithCharms[i] = imp
		importsWithCharms[i].SyntheticCharm = buildSyntheticCharm(imp.Name, imp.Endpoints)
	}

	return errors.Capture(s.modelState.ImportRemoteApplications(ctx, importsWithCharms))
}

// BuildSyntheticCharmForTest creates a synthetic charm from the remote application's
// endpoints. This is used during migration to recreate the charm that
// represents a remote application.
// This is exported for testing purposes.
func BuildSyntheticCharmForTest(appName string, endpoints []crossmodelrelation.RemoteApplicationEndpoint) charm.Charm {
	return buildSyntheticCharm(appName, endpoints)
}

// buildSyntheticCharm creates a synthetic charm from the remote application's
// endpoints. This is used during migration to recreate the charm that
// represents a remote application.
func buildSyntheticCharm(appName string, endpoints []crossmodelrelation.RemoteApplicationEndpoint) charm.Charm {
	provides := make(map[string]charm.Relation)
	requires := make(map[string]charm.Relation)

	for _, ep := range endpoints {
		rel := charm.Relation{
			Name:      ep.Name,
			Role:      ep.Role,
			Interface: ep.Interface,
			Scope:     charm.ScopeGlobal,
		}
		switch ep.Role {
		case charm.RoleProvider:
			provides[ep.Name] = rel
		case charm.RoleRequirer:
			requires[ep.Name] = rel
		}
	}

	return charm.Charm{
		Metadata: charm.Metadata{
			Name:     appName,
			Provides: provides,
			Requires: requires,
		},
		Source:        charm.CMRSource,
		ReferenceName: appName,
	}
}
