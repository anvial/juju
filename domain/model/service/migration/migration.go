// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package migration

import (
	"context"

	"github.com/juju/juju/core/logger"
	coremodel "github.com/juju/juju/core/model"
	"github.com/juju/juju/core/trace"
	"github.com/juju/juju/domain/model"
	modelerrors "github.com/juju/juju/domain/model/errors"
	"github.com/juju/juju/domain/model/service"
	secretbackenderrors "github.com/juju/juju/domain/secretbackend/errors"
	"github.com/juju/juju/internal/errors"
	jujusecrets "github.com/juju/juju/internal/secrets/provider/juju"
	kubernetessecrets "github.com/juju/juju/internal/secrets/provider/kubernetes"
)

// ModelDeleter is an interface for deleting models.
type ModelDeleter interface {
	// DeleteDB is responsible for removing a model from Juju and all of it's
	// associated metadata.
	DeleteDB(string) error
}

// State is the combined state required by the migration service.
type State interface {
	service.CreateModelState
	service.DeleteModelState
}

// MigrationService defines a service for interacting with the underlying state based
// information of a model.
type MigrationService struct {
	st           State
	modelDeleter ModelDeleter
	logger       logger.Logger
}

// NewMigrationService returns a new MigrationService for interacting with a models state.
func NewMigrationService(
	st State,
	modelDeleter ModelDeleter,
	logger logger.Logger,
) *MigrationService {
	return &MigrationService{
		st:           st,
		modelDeleter: modelDeleter,
		logger:       logger,
	}
}

// ImportModel is responsible for importing an existing model into this Juju
// controller by creating the model record in the controller database and marking
// it as importing.
//
// The following error types can be expected to be returned:
// - [modelerrors.AlreadyExists]: When the model uuid is already in use or a
// model with the same name and owner already exists.
// - [errors.NotFound]: When the cloud, cloud region, or credential do not
// exist.
// - [github.com/juju/juju/domain/access/errors.NotFound]: When the owner of the
// model can not be found.
// - [secretbackenderrors.NotFound] When the secret backend for the model
// cannot be found.
func (s *MigrationService) ImportModel(
	ctx context.Context,
	args model.ModelImportArgs,
) (func(context.Context) error, error) {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	if err := args.Validate(); err != nil {
		return nil, errors.Errorf(
			"cannot validate model import args: %w", err,
		)
	}

	modelType, err := service.ModelTypeForCloud(ctx, s.st, args.Cloud)
	if err != nil {
		return nil, errors.Errorf(
			"determining model type when importing model %q: %w",
			args.Name, err,
		)
	}

	if args.SecretBackend == "" {
		switch modelType {
		case coremodel.CAAS:
			args.SecretBackend = kubernetessecrets.BackendName
		case coremodel.IAAS:
			args.SecretBackend = jujusecrets.BackendName
		default:
			return nil, errors.Errorf(
				"%w for model type %q when creating model with name %q",
				secretbackenderrors.NotFound,
				modelType,
				args.Name,
			)
		}
	}

	if err := s.st.ImportModel(ctx, args.UUID, modelType, args.GlobalModelCreationArgs); err != nil {
		return nil, err
	}

	// Return an activator function that marks the model as alive/active.
	// This is separate from the importing/migrating status which is tracked
	// in the migration tables.
	activator := func(ctx context.Context) error {
		return s.st.Activate(ctx, args.UUID)
	}

	return activator, nil
}

// DeleteModel is responsible for removing a model from Juju and all of it's
// associated metadata.
// - errors.NotValid: When the model uuid is not valid.
// - modelerrors.NotFound: When the model does not exist.
func (s *MigrationService) DeleteModel(
	ctx context.Context,
	uuid coremodel.UUID,
) error {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	if err := uuid.Validate(); err != nil {
		return errors.Errorf("delete model, uuid: %w", err)
	}

	// Delete common items from the model. This helps to ensure that the
	// model is cleaned up correctly.
	if err := s.st.Delete(ctx, uuid); err != nil && !errors.Is(err, modelerrors.NotFound) {
		return errors.Errorf("delete model: %w", err)
	}

	// Delete the db completely from the system. Currently, this will remove
	// the db from the dbaccessor, but it will not drop the db (currently not
	// supported in dqlite). For now we do a best effort to remove all items
	// with in the db.
	if err := s.modelDeleter.DeleteDB(uuid.String()); err != nil {
		return errors.Errorf("delete model: %w", err)
	}

	return nil
}
