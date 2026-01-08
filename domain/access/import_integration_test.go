// Copyright 2026 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package access_test

import (
	"github.com/canonical/sqlair"
	"github.com/juju/description/v11"
	"github.com/juju/names/v6"
	"github.com/juju/tc"

	"context"
	"testing"

	"github.com/juju/juju/core/database"
	"github.com/juju/juju/core/model"
	coremodelmigration "github.com/juju/juju/core/modelmigration"
	"github.com/juju/juju/core/permission"
	"github.com/juju/juju/core/user"
	"github.com/juju/juju/domain/access/modelmigration"
	"github.com/juju/juju/domain/access/service"
	"github.com/juju/juju/domain/access/state"
	schematesting "github.com/juju/juju/domain/schema/testing"
	"github.com/juju/juju/internal/errors"
	loggertesting "github.com/juju/juju/internal/logger/testing"
	"github.com/juju/juju/internal/uuid"
)

type importSuite struct {
	schematesting.ControllerSuite

	coordinator *coremodelmigration.Coordinator
	scope       coremodelmigration.Scope
	svc         *service.Service

	modelUUID              model.UUID
	adminUserUUID          user.UUID
	controllerPermissionID permission.ID
}

func TestImportSuite(t *testing.T) {
	tc.Run(t, &importSuite{})
}

func (s *importSuite) SetUpTest(c *tc.C) {
	s.ControllerSuite.SetUpTest(c)
	controllerUUID := s.SeedControllerUUID(c)

	s.controllerPermissionID, _ = permission.ParseTagForID(names.NewControllerTag(controllerUUID))
	s.adminUserUUID = tc.Must(c, user.NewUUID)
	s.modelUUID = tc.Must(c, model.NewUUID)

	s.coordinator = coremodelmigration.NewCoordinator(loggertesting.WrapCheckLog(c))
	modelmigration.RegisterOfferAccessImport(s.coordinator, loggertesting.WrapCheckLog(c))

	controllerFactory := func(context.Context) (database.TxnRunner, error) {
		return s.ControllerTxnRunner(), nil
	}

	s.scope = coremodelmigration.NewScope(controllerFactory, nil, nil, s.modelUUID)
	s.svc = service.NewService(
		state.NewState(controllerFactory, loggertesting.WrapCheckLog(c)),
	)

	adminUserName, _ := user.NewName("admin")
	_, _, _ = s.svc.AddUser(c.Context(), service.AddUserArg{
		UUID:        s.adminUserUUID,
		Name:        adminUserName,
		CreatorUUID: s.adminUserUUID,
		Permission: permission.AccessSpec{
			Target: s.controllerPermissionID,
			Access: permission.SuperuserAccess,
		},
	})

	c.Cleanup(func() {
		s.coordinator = nil
		s.svc = nil
		s.scope = coremodelmigration.Scope{}
		s.modelUUID = ""
		s.adminUserUUID = ""
		s.controllerPermissionID = permission.ID{}
	})
}

func (s *importSuite) TestOfferPermissionImport(c *tc.C) {
	// Arrange: add users on which offer permissions are set.
	joeUserUUID := s.addUserToController(c, "joe", permission.LoginAccess)
	simonUserUUID := s.addUserToController(c, "simon", permission.LoginAccess)

	// Arrange: set up the import data
	desc := description.NewModel(description.ModelArgs{
		Type:   string(model.IAAS),
		Config: map[string]interface{}{"uuid": s.modelUUID.String()},
	})
	appName := "foo"
	app := desc.AddApplication(description.ApplicationArgs{
		Name:     appName,
		CharmURL: "ch:foo-1",
	})
	offerOneUUID := tc.Must(c, uuid.NewUUID).String()
	offerOneName := "foo"
	app.AddOffer(description.ApplicationOfferArgs{
		OfferUUID:       offerOneUUID,
		OfferName:       offerOneName,
		Endpoints:       map[string]string{"db": "db"},
		ApplicationName: appName,
		ACL: map[string]string{
			"admin": "admin",
			"joe":   "consume",
			"simon": "read",
		},
	})
	offerTwoUUID := tc.Must(c, uuid.NewUUID).String()
	offerTwoName := "agent"
	app.AddOffer(description.ApplicationOfferArgs{
		OfferUUID:       offerTwoUUID,
		OfferName:       offerTwoName,
		Endpoints:       map[string]string{"cos-agent": "cos-agent"},
		ApplicationName: appName,
		ACL: map[string]string{
			"simon": "admin",
		},
	})

	// Act
	err := s.coordinator.Perform(c.Context(), s.scope, desc)

	// Assert
	c.Assert(err, tc.ErrorIsNil)

	obtainedOfferPermissions := s.getOfferPermissions(c)
	c.Check(obtainedOfferPermissions, tc.SameContents, []offerAccess{
		{GrantTo: s.adminUserUUID.String(), GrantOn: offerOneUUID, AccessType: "admin"},
		{GrantTo: joeUserUUID, GrantOn: offerOneUUID, AccessType: "consume"},
		{GrantTo: simonUserUUID, GrantOn: offerOneUUID, AccessType: "read"},
		{GrantTo: simonUserUUID, GrantOn: offerTwoUUID, AccessType: "admin"},
	})
}

func (s *importSuite) addUserToController(c *tc.C, name string, access permission.Access) string {
	userName, _ := user.NewName(name)
	userUUID, _, _ := s.svc.AddUser(c.Context(), service.AddUserArg{
		Name:        userName,
		CreatorUUID: s.adminUserUUID,
		Permission: permission.AccessSpec{
			Target: s.controllerPermissionID,
			Access: access,
		},
	})
	return userUUID.String()
}

type offerAccess struct {
	GrantOn    string `db:"grant_on"`
	GrantTo    string `db:"grant_to"`
	AccessType string `db:"access_type"`
}

// getRelationApplicationSettings gets the relation application settings.
func (s *importSuite) getOfferPermissions(c *tc.C) []offerAccess {
	stmt, err := sqlair.Prepare(`
SELECT * AS &offerAccess.*
FROM v_permission_offer
`, offerAccess{})
	c.Assert(err, tc.ErrorIsNil)

	var access []offerAccess

	err = s.TxnRunner().Txn(c.Context(), func(ctx context.Context, tx *sqlair.TX) error {
		return tx.Query(ctx, stmt).GetAll(&access)
	})

	c.Assert(err, tc.ErrorIsNil, tc.Commentf("(Assert) getting offer permissions: %s",
		errors.ErrorStack(err)))
	return access
}
