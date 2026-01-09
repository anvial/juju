// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package modelmigration

import (
	"testing"

	"github.com/juju/clock"
	"github.com/juju/description/v11"
	"github.com/juju/tc"
	"go.uber.org/mock/gomock"

	"github.com/juju/juju/domain/application/charm"
	"github.com/juju/juju/domain/crossmodelrelation"
	loggertesting "github.com/juju/juju/internal/logger/testing"
	"github.com/juju/juju/internal/uuid"
)

type importSuite struct {
	importService *MockImportService
}

func TestImportSuite(t *testing.T) {
	tc.Run(t, &importSuite{})
}

func (s *importSuite) TestImportOffers(c *tc.C) {
	defer s.setupMocks(c).Finish()

	// Arrange
	model := description.NewModel(description.ModelArgs{})
	app := model.AddApplication(description.ApplicationArgs{})
	offerUUID := uuid.MustNewUUID()
	offerArgs := description.ApplicationOfferArgs{
		OfferUUID:       offerUUID.String(),
		OfferName:       "test",
		Endpoints:       map[string]string{"db-admin": "db-admin"},
		ApplicationName: "test",
	}
	app.AddOffer(offerArgs)
	offerUUID2 := uuid.MustNewUUID()
	offerArgs2 := description.ApplicationOfferArgs{
		OfferUUID:       offerUUID2.String(),
		OfferName:       "second",
		Endpoints:       map[string]string{"identity": "identity"},
		ApplicationName: "apple",
	}
	app.AddOffer(offerArgs2)
	input := []crossmodelrelation.OfferImport{
		{
			UUID:            offerUUID,
			Name:            "test",
			ApplicationName: "test",
			Endpoints:       []string{"db-admin"},
		}, {
			UUID:            offerUUID2,
			Name:            "second",
			ApplicationName: "apple",
			Endpoints:       []string{"identity"},
		},
	}
	s.importService.EXPECT().ImportOffers(
		gomock.Any(),
		input,
	).Return(nil)

	// Act
	err := s.newImportOperation(c).importOffers(c.Context(), []description.Application{app})

	// Assert
	c.Assert(err, tc.ErrorIsNil)
}

func (s *importSuite) TestImportRemoteApplications(c *tc.C) {
	defer s.setupMocks(c).Finish()

	// Arrange
	model := description.NewModel(description.ModelArgs{})
	remoteApp := model.AddRemoteApplication(description.RemoteApplicationArgs{
		Name:            "remote-mysql",
		OfferUUID:       "offer-uuid-1234",
		URL:             "ctrl:admin/model.mysql",
		SourceModelUUID: "source-model-uuid",
		Macaroon:        "macaroon-data",
		Bindings:        map[string]string{"db": "alpha"},
	})
	remoteApp.AddEndpoint(description.RemoteEndpointArgs{
		Name:      "db",
		Role:      "provider",
		Interface: "mysql",
	})

	expected := []crossmodelrelation.RemoteApplicationImport{
		{
			Name:            "remote-mysql",
			OfferUUID:       "offer-uuid-1234",
			URL:             "ctrl:admin/model.mysql",
			SourceModelUUID: "source-model-uuid",
			Macaroon:        "macaroon-data",
			Endpoints: []crossmodelrelation.RemoteApplicationEndpoint{
				{
					Name:      "db",
					Role:      charm.RoleProvider,
					Interface: "mysql",
				},
			},
			Bindings:        map[string]string{"db": "alpha"},
			IsConsumerProxy: false,
		},
	}
	s.importService.EXPECT().ImportRemoteApplications(
		gomock.Any(),
		expected,
	).Return(nil)

	// Act
	err := s.newImportOperation(c).importRemoteApplications(c.Context(), model.RemoteApplications())

	// Assert
	c.Assert(err, tc.ErrorIsNil)
}

func (s *importSuite) TestImportRemoteApplicationsEmpty(c *tc.C) {
	defer s.setupMocks(c).Finish()

	// Arrange
	model := description.NewModel(description.ModelArgs{})

	// Act - no remote applications, no mock expectations needed
	err := s.newImportOperation(c).importRemoteApplications(c.Context(), model.RemoteApplications())

	// Assert
	c.Assert(err, tc.ErrorIsNil)
}

func (s *importSuite) TestImportRemoteApplicationsMultiple(c *tc.C) {
	defer s.setupMocks(c).Finish()

	// Arrange
	model := description.NewModel(description.ModelArgs{})

	remoteApp1 := model.AddRemoteApplication(description.RemoteApplicationArgs{
		Name:            "remote-mysql",
		OfferUUID:       "offer-uuid-1",
		URL:             "ctrl:admin/model.mysql",
		SourceModelUUID: "source-model-uuid-1",
	})
	remoteApp1.AddEndpoint(description.RemoteEndpointArgs{
		Name:      "db",
		Role:      "provider",
		Interface: "mysql",
	})

	remoteApp2 := model.AddRemoteApplication(description.RemoteApplicationArgs{
		Name:            "remote-postgresql",
		OfferUUID:       "offer-uuid-2",
		URL:             "ctrl:admin/model.postgresql",
		SourceModelUUID: "source-model-uuid-2",
	})
	remoteApp2.AddEndpoint(description.RemoteEndpointArgs{
		Name:      "db",
		Role:      "provider",
		Interface: "pgsql",
	})
	remoteApp2.AddEndpoint(description.RemoteEndpointArgs{
		Name:      "admin",
		Role:      "requirer",
		Interface: "admin",
	})

	expected := []crossmodelrelation.RemoteApplicationImport{
		{
			Name:            "remote-mysql",
			OfferUUID:       "offer-uuid-1",
			URL:             "ctrl:admin/model.mysql",
			SourceModelUUID: "source-model-uuid-1",
			Endpoints: []crossmodelrelation.RemoteApplicationEndpoint{
				{
					Name:      "db",
					Role:      charm.RoleProvider,
					Interface: "mysql",
				},
			},
		},
		{
			Name:            "remote-postgresql",
			OfferUUID:       "offer-uuid-2",
			URL:             "ctrl:admin/model.postgresql",
			SourceModelUUID: "source-model-uuid-2",
			Endpoints: []crossmodelrelation.RemoteApplicationEndpoint{
				{
					Name:      "db",
					Role:      charm.RoleProvider,
					Interface: "pgsql",
				},
				{
					Name:      "admin",
					Role:      charm.RoleRequirer,
					Interface: "admin",
				},
			},
		},
	}
	s.importService.EXPECT().ImportRemoteApplications(
		gomock.Any(),
		expected,
	).Return(nil)

	// Act
	err := s.newImportOperation(c).importRemoteApplications(c.Context(), model.RemoteApplications())

	// Assert
	c.Assert(err, tc.ErrorIsNil)
}

func (s *importSuite) setupMocks(c *tc.C) *gomock.Controller {
	ctrl := gomock.NewController(c)

	s.importService = NewMockImportService(ctrl)

	c.Cleanup(func() {
		s.importService = nil
	})

	return ctrl
}

func (s *importSuite) newImportOperation(c *tc.C) *importOperation {
	return &importOperation{
		importService: s.importService,
		clock:         clock.WallClock,
		logger:        loggertesting.WrapCheckLog(c),
	}
}
