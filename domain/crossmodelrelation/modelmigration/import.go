// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package modelmigration

import (
	"context"

	"github.com/juju/clock"
	"github.com/juju/collections/transform"
	"github.com/juju/description/v11"

	"github.com/juju/juju/core/logger"
	"github.com/juju/juju/core/modelmigration"
	"github.com/juju/juju/domain/application/charm"
	"github.com/juju/juju/domain/crossmodelrelation"
	"github.com/juju/juju/domain/crossmodelrelation/service"
	modelstate "github.com/juju/juju/domain/crossmodelrelation/state/model"
	"github.com/juju/juju/internal/errors"
	"github.com/juju/juju/internal/uuid"
)

// Coordinator is the interface that is used to add operations to a migration.
type Coordinator interface {
	// Add adds the given operation to the migration.
	Add(modelmigration.Operation)
}

// RegisterImport registers the import operations with the given coordinator.
func RegisterImport(coordinator Coordinator, clock clock.Clock, logger logger.Logger) {
	coordinator.Add(&importOperation{
		clock:  clock,
		logger: logger,
	})
}

// ImportService provides a subset of the cross model relation domain
// service methods needed for import.
type ImportService interface {
	// ImportOffers adds offers being migrated to the current model.
	ImportOffers(context.Context, []crossmodelrelation.OfferImport) error

	// ImportRemoteApplications adds remote application offerers being migrated
	// to the current model.
	ImportRemoteApplications(context.Context, []crossmodelrelation.RemoteApplicationImport) error
}

type importOperation struct {
	modelmigration.BaseOperation

	importService ImportService

	clock  clock.Clock
	logger logger.Logger
}

// Name returns the name of this operation.
func (i *importOperation) Name() string {
	return "import cross model relations"
}

// Setup implements Operation.
func (i *importOperation) Setup(scope modelmigration.Scope) error {
	i.importService = service.NewMigrationService(
		// TODO(import) - get model UUID and pass it in here
		modelstate.NewState(scope.ModelDB(), "", i.clock, i.logger),
		i.logger,
	)
	return nil
}

// Execute the import of the cross model relations contained in the model.
func (i *importOperation) Execute(ctx context.Context, model description.Model) error {
	if err := i.importOffers(ctx, model.Applications()); err != nil {
		return errors.Errorf("importing offers: %w", err)
	}
	if err := i.importRemoteApplications(ctx, model.RemoteApplications()); err != nil {
		return errors.Errorf("importing remote applications: %w", err)
	}
	return nil
}

func (i *importOperation) importOffers(ctx context.Context, apps []description.Application) error {
	input := make([]crossmodelrelation.OfferImport, 0)
	for _, app := range apps {
		for _, offer := range app.Offers() {
			offerUUID, err := uuid.UUIDFromString(offer.OfferUUID())
			if err != nil {
				return errors.Errorf("validating uuid for offer %q,%q: %w",
					offer.ApplicationName(), offer.OfferName(), err)
			}

			endpoints := transform.MapToSlice(
				offer.Endpoints(),
				func(k, v string) []string {
					return []string{v}
				},
			)
			imp := crossmodelrelation.OfferImport{
				UUID:            offerUUID,
				Name:            offer.OfferName(),
				ApplicationName: offer.ApplicationName(),
				Endpoints:       endpoints,
			}
			input = append(input, imp)
		}
	}
	if len(input) == 0 {
		return nil
	}
	return i.importService.ImportOffers(ctx, input)
}

func (i *importOperation) importRemoteApplications(ctx context.Context, remoteApps []description.RemoteApplication) error {
	input := make([]crossmodelrelation.RemoteApplicationImport, 0, len(remoteApps))
	for _, remoteApp := range remoteApps {
		endpoints := make([]crossmodelrelation.RemoteApplicationEndpoint, 0, len(remoteApp.Endpoints()))
		for _, ep := range remoteApp.Endpoints() {
			role, err := parseRelationRole(ep.Role())
			if err != nil {
				return errors.Errorf("parsing role for endpoint %q on remote app %q: %w",
					ep.Name(), remoteApp.Name(), err)
			}
			endpoints = append(endpoints, crossmodelrelation.RemoteApplicationEndpoint{
				Name:      ep.Name(),
				Role:      role,
				Interface: ep.Interface(),
			})
		}

		imp := crossmodelrelation.RemoteApplicationImport{
			Name:            remoteApp.Name(),
			OfferUUID:       remoteApp.OfferUUID(),
			URL:             remoteApp.URL(),
			SourceModelUUID: remoteApp.SourceModelUUID(),
			Macaroon:        remoteApp.Macaroon(),
			Endpoints:       endpoints,
			Bindings:        remoteApp.Bindings(),
			IsConsumerProxy: remoteApp.IsConsumerProxy(),
		}
		input = append(input, imp)
	}
	if len(input) == 0 {
		return nil
	}
	return i.importService.ImportRemoteApplications(ctx, input)
}

// parseRelationRole parses a string role to a charm.RelationRole.
func parseRelationRole(role string) (charm.RelationRole, error) {
	switch role {
	case "provider":
		return charm.RoleProvider, nil
	case "requirer":
		return charm.RoleRequirer, nil
	case "peer":
		return charm.RolePeer, nil
	default:
		return "", errors.Errorf("unknown relation role %q", role)
	}
}
