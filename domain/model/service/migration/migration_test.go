// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package migration

import (
	"testing"

	"github.com/juju/names/v6"
	"github.com/juju/tc"
	gomock "go.uber.org/mock/gomock"

	"github.com/juju/juju/cloud"
	coremodel "github.com/juju/juju/core/model"
	"github.com/juju/juju/domain/model"
	modelerrors "github.com/juju/juju/domain/model/errors"
	loggertesting "github.com/juju/juju/internal/logger/testing"
	jujusecrets "github.com/juju/juju/internal/secrets/provider/juju"
	kubernetessecrets "github.com/juju/juju/internal/secrets/provider/kubernetes"
	"github.com/juju/juju/internal/testhelpers"
)

type migrationServiceSuite struct {
	testhelpers.IsolationSuite

	state   *MockState
	deleter *MockModelDeleter
}

func TestMigrationServiceSuite(t *testing.T) {
	tc.Run(t, &migrationServiceSuite{})
}

func (s *migrationServiceSuite) newService(c *tc.C) *MigrationService {
	return NewMigrationService(s.state, s.deleter, loggertesting.WrapCheckLog(c))
}

func (s *migrationServiceSuite) TestImportModelIAAS(c *tc.C) {
	defer s.setupMocks(c).Finish()

	uuid := tc.Must(c, coremodel.NewUUID)

	sExp := s.state.EXPECT()
	sExp.CloudType(gomock.Any(), "aws").Return("aws", nil)
	sExp.CloudSupportsAuthType(gomock.Any(), "aws", cloud.EmptyAuthType).Return(true, nil)
	sExp.Create(gomock.Any(), uuid, coremodel.IAAS, model.GlobalModelCreationArgs{
		Name:          "foo",
		Cloud:         "aws",
		Qualifier:     coremodel.QualifierFromUserTag(names.NewUserTag("jim")),
		SecretBackend: jujusecrets.BackendName,
	}).Return(nil)

	svc := s.newService(c)

	fn, err := svc.ImportModel(c.Context(), model.ModelImportArgs{
		UUID: uuid,
		GlobalModelCreationArgs: model.GlobalModelCreationArgs{
			Name:      "foo",
			Cloud:     "aws",
			Qualifier: coremodel.QualifierFromUserTag(names.NewUserTag("jim")),
		},
	})
	c.Assert(err, tc.ErrorIsNil)
	c.Check(fn, tc.Not(tc.IsNil))
}

func (s *migrationServiceSuite) TestImportModelCAAS(c *tc.C) {
	defer s.setupMocks(c).Finish()

	uuid := tc.Must(c, coremodel.NewUUID)

	sExp := s.state.EXPECT()
	sExp.CloudType(gomock.Any(), "k8s").Return(cloud.CloudTypeKubernetes, nil)
	sExp.CloudSupportsAuthType(gomock.Any(), "k8s", cloud.EmptyAuthType).Return(true, nil)
	sExp.Create(gomock.Any(), uuid, coremodel.CAAS, model.GlobalModelCreationArgs{
		Name:          "foo",
		Cloud:         "k8s",
		Qualifier:     coremodel.QualifierFromUserTag(names.NewUserTag("jim")),
		SecretBackend: kubernetessecrets.BackendName,
	}).Return(nil)

	svc := s.newService(c)

	fn, err := svc.ImportModel(c.Context(), model.ModelImportArgs{
		UUID: uuid,
		GlobalModelCreationArgs: model.GlobalModelCreationArgs{
			Name:      "foo",
			Cloud:     "k8s",
			Qualifier: coremodel.QualifierFromUserTag(names.NewUserTag("jim")),
		},
	})
	c.Assert(err, tc.ErrorIsNil)

	c.Check(fn, tc.Not(tc.IsNil))
}

func (s *migrationServiceSuite) TestImportModelActivate(c *tc.C) {
	defer s.setupMocks(c).Finish()

	uuid := tc.Must(c, coremodel.NewUUID)

	sExp := s.state.EXPECT()
	sExp.CloudType(gomock.Any(), gomock.Any()).Return("aws", nil)
	sExp.CloudSupportsAuthType(gomock.Any(), gomock.Any(), cloud.EmptyAuthType).Return(true, nil)
	sExp.Create(gomock.Any(), uuid, gomock.Any(), gomock.Any()).Return(nil)

	sExp.Activate(gomock.Any(), uuid).Return(nil)

	svc := s.newService(c)

	fn, err := svc.ImportModel(c.Context(), model.ModelImportArgs{
		UUID: uuid,
		GlobalModelCreationArgs: model.GlobalModelCreationArgs{
			Name:      "foo",
			Cloud:     "aws",
			Qualifier: coremodel.QualifierFromUserTag(names.NewUserTag("jim")),
		},
	})
	c.Assert(err, tc.ErrorIsNil)
	c.Check(fn, tc.Not(tc.IsNil))

	err = fn(c.Context())
	c.Assert(err, tc.ErrorIsNil)
}

func (s *migrationServiceSuite) TestDeleteModel(c *tc.C) {
	defer s.setupMocks(c).Finish()

	uuid := tc.Must(c, coremodel.NewUUID)

	s.state.EXPECT().Delete(gomock.Any(), uuid).Return(nil)
	s.deleter.EXPECT().DeleteDB(uuid.String()).Return(nil)

	svc := s.newService(c)

	err := svc.DeleteModel(c.Context(), uuid)
	c.Assert(err, tc.ErrorIsNil)
}

func (s *migrationServiceSuite) TestDeleteModelNotFound(c *tc.C) {
	defer s.setupMocks(c).Finish()

	uuid := tc.Must(c, coremodel.NewUUID)

	s.state.EXPECT().Delete(gomock.Any(), uuid).Return(modelerrors.NotFound)
	s.deleter.EXPECT().DeleteDB(uuid.String()).Return(nil)

	svc := s.newService(c)

	err := svc.DeleteModel(c.Context(), uuid)
	c.Assert(err, tc.ErrorIsNil)
}

func (s *migrationServiceSuite) setupMocks(c *tc.C) *gomock.Controller {
	ctrl := gomock.NewController(c)

	s.state = NewMockState(ctrl)
	s.deleter = NewMockModelDeleter(ctrl)

	c.Cleanup(func() {
		s.state = nil
		s.deleter = nil
	})

	return ctrl
}
