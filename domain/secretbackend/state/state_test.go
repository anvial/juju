// Copyright 2024 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/cloud"
	coremodel "github.com/juju/juju/core/model"
	"github.com/juju/juju/core/permission"
	coresecrets "github.com/juju/juju/core/secrets"
	"github.com/juju/juju/core/user"
	"github.com/juju/juju/domain"
	cloudstate "github.com/juju/juju/domain/cloud/state"
	schematesting "github.com/juju/juju/domain/schema/testing"
	"github.com/juju/juju/domain/secretbackend"
	backenderrors "github.com/juju/juju/domain/secretbackend/errors"
	userstate "github.com/juju/juju/domain/user/state"
	"github.com/juju/juju/internal/uuid"
	jujutesting "github.com/juju/juju/testing"
)

type stateSuite struct {
	schematesting.ControllerSuite
	state *State
}

var _ = gc.Suite(&stateSuite{})

func (s *stateSuite) SetUpTest(c *gc.C) {
	s.ControllerSuite.SetUpTest(c)

	s.state = NewState(s.TxnRunnerFactory(), jujutesting.NewCheckLogger(c))
}

func (s *stateSuite) createModel(c *gc.C) coremodel.UUID {
	// We need to generate a user in the database so that we can set the model
	// owner.
	userName := "test-user"
	userUUID, err := user.NewUUID()
	c.Assert(err, jc.ErrorIsNil)
	userState := userstate.NewState(s.TxnRunnerFactory())
	err = userState.AddUser(
		context.Background(),
		userUUID,
		userName,
		userName,
		userUUID,
		// TODO (stickupkid): This should be AdminAccess, but we don't have
		// a model to set the user as the owner of.
		permission.ControllerForAccess(permission.SuperuserAccess),
	)
	c.Assert(err, jc.ErrorIsNil)

	cloudstate := cloudstate.NewState(s.TxnRunnerFactory())
	err = cloudstate.UpsertCloud(context.Background(), cloud.Cloud{
		Name:      "my-cloud",
		Type:      "ec2",
		AuthTypes: cloud.AuthTypes{cloud.AccessKeyAuthType, cloud.UserPassAuthType},
		Regions: []cloud.Region{
			{Name: "my-region"},
		},
	})
	c.Assert(err, jc.ErrorIsNil)

	db := s.DB()
	var cloud_uuid string
	q := `SELECT uuid FROM cloud WHERE name = ?`
	err = db.QueryRowContext(context.Background(), q, "my-cloud").Scan(&cloud_uuid)
	c.Assert(err, jc.ErrorIsNil)

	modelUUID := coremodel.UUID(uuid.MustNewUUID().String())
	q = `
INSERT INTO model_list (uuid) VALUES (?)`[1:]
	_, err = db.ExecContext(context.Background(), q, modelUUID)
	c.Assert(err, jc.ErrorIsNil)
	q = `
INSERT INTO model_metadata
	(model_uuid, cloud_uuid, model_type_id, name, owner_uuid)
VALUES (?, ?, ?, ?, ?);`[1:]
	_, err = db.ExecContext(
		context.Background(), q,
		modelUUID, cloud_uuid, 0, "my-model", userUUID,
	)
	c.Assert(err, jc.ErrorIsNil)
	return modelUUID
}

func (s *stateSuite) assertSecretBackend(
	c *gc.C, expectedSecretBackend coresecrets.SecretBackend, expectedNextRotationTime *time.Time,
) {
	db := s.DB()
	row := db.QueryRow(`
SELECT uuid, name, backend_type, token_rotate_interval
FROM secret_backend
WHERE uuid = ?`[1:], expectedSecretBackend.ID)
	c.Assert(row.Err(), gc.IsNil)

	var (
		actual              coresecrets.SecretBackend
		tokenRotateInterval domain.NullableDuration
	)
	err := row.Scan(&actual.ID, &actual.Name, &actual.BackendType, &tokenRotateInterval)
	c.Assert(err, gc.IsNil)

	if tokenRotateInterval.Valid {
		actual.TokenRotateInterval = &tokenRotateInterval.Duration
	}
	if expectedNextRotationTime != nil {
		var actualNextRotationTime sql.NullTime
		row = db.QueryRow(`
SELECT next_rotation_time
FROM secret_backend_rotation
WHERE backend_uuid = ?`[1:], expectedSecretBackend.ID)
		c.Assert(row.Err(), gc.IsNil)
		err = row.Scan(&actualNextRotationTime)
		c.Assert(err, gc.IsNil)
		c.Assert(actualNextRotationTime.Valid, jc.IsTrue)
		c.Assert(actualNextRotationTime.Time.Equal(*expectedNextRotationTime), jc.IsTrue)
	} else {
		row = db.QueryRow(`
SELECT COUNT(*)
FROM secret_backend_rotation
WHERE backend_uuid = ?`[1:], expectedSecretBackend.ID)
		var count int
		err = row.Scan(&count)
		c.Assert(err, gc.IsNil)
		c.Assert(count, gc.Equals, 0)
	}

	if len(expectedSecretBackend.Config) > 0 {
		actual.Config = map[string]interface{}{}
		rows, err := db.Query(`
SELECT name, content
FROM secret_backend_config
WHERE backend_uuid = ?`[1:], expectedSecretBackend.ID)
		c.Assert(err, gc.IsNil)
		c.Assert(rows.Err(), gc.IsNil)
		defer rows.Close()
		for rows.Next() {
			var k, v string
			err = rows.Scan(&k, &v)
			c.Assert(err, gc.IsNil)
			actual.Config[k] = v
		}
	} else {
		var count int
		row = db.QueryRow(`
SELECT COUNT(*)
FROM secret_backend_config
WHERE backend_uuid = ?`[1:], expectedSecretBackend.ID)
		err = row.Scan(&count)
		c.Assert(err, gc.IsNil)
		c.Assert(count, gc.Equals, 0)
	}
	c.Assert(actual, gc.DeepEquals, expectedSecretBackend)
}

func (s *stateSuite) TestCreateSecretBackendFailed(c *gc.C) {
	backendID := uuid.MustNewUUID().String()
	rotateInternal := 24 * time.Hour
	nextRotateTime := time.Now().Add(rotateInternal)
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		NextRotateTime:      &nextRotateTime,
		Config: map[string]interface{}{
			"key1": "",
		},
	})
	c.Assert(err, jc.ErrorIs, backenderrors.NotValid)
	c.Assert(err, gc.ErrorMatches, `secret backend not valid: empty config value for "`+backendID+`"`)

	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		NextRotateTime:      &nextRotateTime,
		Config: map[string]interface{}{
			"": "value1",
		},
	})
	c.Assert(err, jc.ErrorIs, backenderrors.NotValid)
	c.Assert(err, gc.ErrorMatches, `secret backend not valid: empty config key for "`+backendID+`"`)
}

func (s *stateSuite) TestCreateSecretBackend(c *gc.C) {
	backendID := uuid.MustNewUUID().String()
	rotateInternal := 24 * time.Hour
	nextRotateTime := time.Now().Add(rotateInternal)
	result, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		NextRotateTime:      &nextRotateTime,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	c.Assert(err, gc.IsNil)
	c.Assert(result, gc.Equals, backendID)

	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}, &nextRotateTime)
}

func (s *stateSuite) TestCreateSecretBackendWithNoRotateNoConfig(c *gc.C) {
	backendID := uuid.MustNewUUID().String()
	result, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:          backendID,
		Name:        "my-backend",
		BackendType: "vault",
	})
	c.Assert(err, gc.IsNil)
	c.Assert(result, gc.Equals, backendID)

	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:          backendID,
		Name:        "my-backend",
		BackendType: "vault",
	}, nil)
}

func (s *stateSuite) TestUpsertSecretBackendInvalidArg(c *gc.C) {
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{})
	c.Assert(err, gc.ErrorMatches, `secret backend not valid: ID is missing`)

	backendID := uuid.MustNewUUID().String()
	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID: backendID,
	})
	c.Assert(err, gc.ErrorMatches, `secret backend not valid: name is missing`)

	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:   backendID,
		Name: "my-backend",
	})
	c.Assert(err, gc.ErrorMatches, `secret backend not valid: type is missing`)
}

func (s *stateSuite) TestUpdateSecretBackend(c *gc.C) {
	backendID := uuid.MustNewUUID().String()
	rotateInternal := 24 * time.Hour
	nextRotateTime := time.Now().Add(rotateInternal)
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		NextRotateTime:      &nextRotateTime,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}, &nextRotateTime)

	newRotateInternal := 48 * time.Hour
	newNextRotateTime := time.Now().Add(newRotateInternal)
	nameChange := "my-backend-updated"
	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID,
		Name:                nameChange,
		TokenRotateInterval: &newRotateInternal,
		NextRotateTime:      &newNextRotateTime,
		Config: map[string]interface{}{
			"key1": "value1-updated",
			"key3": "value3",
		},
	})
	c.Assert(err, gc.IsNil)

	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend-updated",
		BackendType:         "vault",
		TokenRotateInterval: &newRotateInternal,
		Config: map[string]interface{}{
			"key1": "value1-updated",
			"key3": "value3",
		},
	}, &newNextRotateTime)
}

func (s *stateSuite) TestUpdateSecretBackendWithNoRotateNoConfig(c *gc.C) {
	backendID := uuid.MustNewUUID().String()
	rotateInternal := 24 * time.Hour
	nextRotateTime := time.Now().Add(rotateInternal)
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		NextRotateTime:      &nextRotateTime,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}, &nextRotateTime)

	nameChange := "my-backend-updated"
	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:   backendID,
		Name: nameChange,
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend-updated",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}, &nextRotateTime)
}

func (s *stateSuite) TestUpdateSecretBackendFailed(c *gc.C) {
	backendID1 := uuid.MustNewUUID().String()
	rotateInternal := 24 * time.Hour
	nextRotateTime := time.Now().Add(rotateInternal)
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID1,
		Name:                "my-backend1",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		NextRotateTime:      &nextRotateTime,
	})
	c.Assert(err, gc.IsNil)

	backendID2 := uuid.MustNewUUID().String()
	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID2,
		Name:                "my-backend2",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		NextRotateTime:      &nextRotateTime,
	})
	c.Assert(err, gc.IsNil)

	nameChange := "my-backend1"
	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:   backendID2,
		Name: nameChange,
	})
	c.Assert(err, jc.ErrorIs, backenderrors.AlreadyExists)
	c.Assert(err, gc.ErrorMatches, `secret backend already exists: name "my-backend1"`)

	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID: backendID2,
		Config: map[string]interface{}{
			"key1": "",
		},
	})
	c.Assert(err, jc.ErrorIs, backenderrors.NotValid)
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf(`secret backend not valid: empty config value for %q`, backendID2))

	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID: backendID2,
		Config: map[string]interface{}{
			"": "value1",
		},
	})
	c.Assert(err, jc.ErrorIs, backenderrors.NotValid)
	c.Assert(err, gc.ErrorMatches, fmt.Sprintf(`secret backend not valid: empty config key for %q`, backendID2))
}

func (s *stateSuite) TestDeleteSecretBackend(c *gc.C) {
	db := s.DB()
	backendID := uuid.MustNewUUID().String()

	rotateInternal := 24 * time.Hour
	nextRotateTime := time.Now().Add(rotateInternal)
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		NextRotateTime:      &nextRotateTime,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}, &nextRotateTime)

	modelUUID := s.createModel(c)
	_, err = db.Exec(`
UPDATE model_metadata
SET secret_backend_uuid = ?
WHERE model_uuid = ?`, backendID, modelUUID)
	c.Assert(err, jc.ErrorIsNil)

	row := db.QueryRow(`
SELECT secret_backend_uuid
FROM model_metadata
WHERE model_uuid = ?`[1:], modelUUID)
	var configuredBackendUUID sql.NullString
	err = row.Scan(&configuredBackendUUID)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(configuredBackendUUID.Valid, jc.IsTrue)
	c.Assert(configuredBackendUUID.String, gc.Equals, backendID)

	err = s.state.DeleteSecretBackend(context.Background(), backendID, false)
	c.Assert(err, gc.IsNil)

	row = db.QueryRow(`
SELECT COUNT(*)
FROM secret_backend
WHERE uuid = ?`[1:], backendID)
	var count int
	err = row.Scan(&count)
	c.Assert(err, gc.IsNil)
	c.Assert(count, gc.Equals, 0)

	row = db.QueryRow(`
SELECT COUNT(*)
FROM secret_backend_config
WHERE backend_uuid = ?`[1:], backendID)
	err = row.Scan(&count)
	c.Assert(err, gc.IsNil)
	c.Assert(count, gc.Equals, 0)

	row = db.QueryRow(`
SELECT COUNT(*)
FROM secret_backend_rotation
WHERE backend_uuid = ?`[1:], backendID)
	err = row.Scan(&count)
	c.Assert(err, gc.IsNil)
	c.Assert(count, gc.Equals, 0)

	row = db.QueryRow(`
SELECT secret_backend_uuid
FROM model_metadata
WHERE model_uuid = ?`[1:], modelUUID)
	err = row.Scan(&configuredBackendUUID)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(configuredBackendUUID.Valid, jc.IsFalse)
}

func (s *stateSuite) TestDeleteSecretBackendWithNoConfigNoNextRotationTime(c *gc.C) {
	backendID := uuid.MustNewUUID().String()
	rotateInternal := 24 * time.Hour
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
	}, nil)

	err = s.state.DeleteSecretBackend(context.Background(), backendID, false)
	c.Assert(err, gc.IsNil)

	db := s.DB()
	row := db.QueryRow(`
SELECT COUNT(*)
FROM secret_backend
WHERE uuid = ?`[1:], backendID)
	var count int
	err = row.Scan(&count)
	c.Assert(err, gc.IsNil)
	c.Assert(count, gc.Equals, 0)

	row = db.QueryRow(`
SELECT COUNT(*)
FROM secret_backend_config
WHERE backend_uuid = ?`[1:], backendID)
	err = row.Scan(&count)
	c.Assert(err, gc.IsNil)
	c.Assert(count, gc.Equals, 0)

	row = db.QueryRow(`
SELECT COUNT(*)
FROM secret_backend_rotation
WHERE backend_uuid = ?`[1:], backendID)
	err = row.Scan(&count)
	c.Assert(err, gc.IsNil)
	c.Assert(count, gc.Equals, 0)
}

func (s *stateSuite) TestDeleteSecretBackendInUseFail(c *gc.C) {
	c.Skip("TODO: wait for secret DqLite support")
}

func (s *stateSuite) TestDeleteSecretBackendInUseWithForce(c *gc.C) {
	c.Skip("TODO: wait for secret DqLite support")
}

func (s *stateSuite) TestListSecretBackends(c *gc.C) {
	backendID1 := uuid.MustNewUUID().String()
	rotateInternal1 := 24 * time.Hour
	nextRotateTime1 := time.Now().Add(rotateInternal1)
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID1,
		Name:                "my-backend1",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal1,
		NextRotateTime:      &nextRotateTime1,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID1,
		Name:                "my-backend1",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal1,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}, &nextRotateTime1)

	backendID2 := uuid.MustNewUUID().String()
	rotateInternal2 := 48 * time.Hour
	nextRotateTime2 := time.Now().Add(rotateInternal2)
	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID2,
		Name:                "my-backend2",
		BackendType:         "kubernetes",
		TokenRotateInterval: &rotateInternal2,
		NextRotateTime:      &nextRotateTime2,
		Config: map[string]interface{}{
			"key3": "value3",
			"key4": "value4",
		},
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID2,
		Name:                "my-backend2",
		BackendType:         "kubernetes",
		TokenRotateInterval: &rotateInternal2,
		Config: map[string]interface{}{
			"key3": "value3",
			"key4": "value4",
		},
	}, &nextRotateTime2)

	backends, err := s.state.ListSecretBackends(context.Background())
	c.Assert(err, gc.IsNil)
	c.Assert(backends, gc.HasLen, 2)
	c.Assert(backends, gc.DeepEquals, []*coresecrets.SecretBackend{
		{
			ID:                  backendID1,
			Name:                "my-backend1",
			BackendType:         "vault",
			TokenRotateInterval: &rotateInternal1,
			Config: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			ID:                  backendID2,
			Name:                "my-backend2",
			BackendType:         "kubernetes",
			TokenRotateInterval: &rotateInternal2,
			Config: map[string]interface{}{
				"key3": "value3",
				"key4": "value4",
			},
		},
	})
}

func (s *stateSuite) TestListSecretBackendsEmpty(c *gc.C) {
	backends, err := s.state.ListSecretBackends(context.Background())
	c.Assert(err, gc.IsNil)
	c.Assert(backends, gc.IsNil)
}

func (s *stateSuite) TestGetSecretBackendByName(c *gc.C) {
	backendID := uuid.MustNewUUID().String()
	rotateInternal := 24 * time.Hour
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:          backendID,
		Name:        "my-backend",
		BackendType: "vault",
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:          backendID,
		Name:        "my-backend",
		BackendType: "vault",
	}, nil)

	backend, err := s.state.GetSecretBackendByName(context.Background(), "my-backend")
	c.Assert(err, gc.IsNil)
	c.Assert(backend, gc.DeepEquals, &coresecrets.SecretBackend{
		ID:          backendID,
		Name:        "my-backend",
		BackendType: "vault",
	})

	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID,
		TokenRotateInterval: &rotateInternal,
	})
	c.Assert(err, gc.IsNil)
	backend, err = s.state.GetSecretBackendByName(context.Background(), "my-backend")
	c.Assert(err, gc.IsNil)
	c.Assert(backend, gc.DeepEquals, &coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
	})

	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID: backendID,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	c.Assert(err, gc.IsNil)
	backend, err = s.state.GetSecretBackendByName(context.Background(), "my-backend")
	c.Assert(err, gc.IsNil)
	c.Assert(backend, gc.DeepEquals, &coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
}

func (s *stateSuite) TestGetSecretBackendByNameNotFound(c *gc.C) {
	backend, err := s.state.GetSecretBackendByName(context.Background(), "my-backend")
	c.Assert(err, jc.ErrorIs, backenderrors.NotFound)
	c.Assert(err, gc.ErrorMatches, `secret backend not found: "my-backend"`)
	c.Assert(backend, gc.IsNil)
}

func (s *stateSuite) TestGetSecretBackend(c *gc.C) {
	backendID := uuid.MustNewUUID().String()
	rotateInternal := 24 * time.Hour
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:          backendID,
		Name:        "my-backend",
		BackendType: "vault",
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:          backendID,
		Name:        "my-backend",
		BackendType: "vault",
	}, nil)

	backend, err := s.state.GetSecretBackend(context.Background(), backendID)
	c.Assert(err, gc.IsNil)
	c.Assert(backend, gc.DeepEquals, &coresecrets.SecretBackend{
		ID:          backendID,
		Name:        "my-backend",
		BackendType: "vault",
	})

	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID,
		TokenRotateInterval: &rotateInternal,
	})
	c.Assert(err, gc.IsNil)
	backend, err = s.state.GetSecretBackend(context.Background(), backendID)
	c.Assert(err, gc.IsNil)
	c.Assert(backend, gc.DeepEquals, &coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
	})

	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID: backendID,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	c.Assert(err, gc.IsNil)
	backend, err = s.state.GetSecretBackend(context.Background(), backendID)
	c.Assert(err, gc.IsNil)
	c.Assert(backend, gc.DeepEquals, &coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
}

func (s *stateSuite) TestGetSecretBackendNotFound(c *gc.C) {
	backendID := uuid.MustNewUUID().String()
	backend, err := s.state.GetSecretBackend(context.Background(), backendID)
	c.Assert(err, jc.ErrorIs, backenderrors.NotFound)
	c.Assert(err, gc.ErrorMatches, `secret backend not found: "`+backendID+`"`)
	c.Assert(backend, gc.IsNil)
}

func (s *stateSuite) TestSecretBackendRotated(c *gc.C) {
	backendID := uuid.MustNewUUID().String()
	rotateInternal := 24 * time.Hour
	nextRotateTime := time.Now().Add(rotateInternal)
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
		NextRotateTime:      &nextRotateTime,
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
	}, &nextRotateTime)

	newNextRotateTime := time.Now().Add(2 * rotateInternal)
	err = s.state.SecretBackendRotated(context.Background(), backendID, newNextRotateTime)
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
	},
		// No ops because the new next rotation time is after the current one.
		&nextRotateTime,
	)

	newNextRotateTime = time.Now().Add(rotateInternal / 2)
	err = s.state.SecretBackendRotated(context.Background(), backendID, newNextRotateTime)
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID,
		Name:                "my-backend",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal,
	},
		// The next rotation time is updated.
		&newNextRotateTime,
	)

	nonExistBackendID := uuid.MustNewUUID().String()
	newNextRotateTime = time.Now().Add(rotateInternal / 4)
	err = s.state.SecretBackendRotated(context.Background(), nonExistBackendID, newNextRotateTime)
	c.Assert(err, jc.ErrorIs, backenderrors.NotFound)
	c.Assert(err, gc.ErrorMatches, `secret backend not found: "`+nonExistBackendID+`"`)
}

func (s *stateSuite) TestGetSecretBackendRotateChanges(c *gc.C) {
	backendID1 := uuid.MustNewUUID().String()
	rotateInternal1 := 24 * time.Hour
	nextRotateTime1 := time.Now().Add(rotateInternal1)
	_, err := s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID1,
		Name:                "my-backend1",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal1,
		NextRotateTime:      &nextRotateTime1,
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID1,
		Name:                "my-backend1",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal1,
	}, &nextRotateTime1)

	backendID2 := uuid.MustNewUUID().String()
	rotateInternal2 := 48 * time.Hour
	nextRotateTime2 := time.Now().Add(rotateInternal2)
	_, err = s.state.UpsertSecretBackend(context.Background(), secretbackend.UpsertSecretBackendParams{
		ID:                  backendID2,
		Name:                "my-backend2",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal2,
		NextRotateTime:      &nextRotateTime2,
	})
	c.Assert(err, gc.IsNil)
	s.assertSecretBackend(c, coresecrets.SecretBackend{
		ID:                  backendID2,
		Name:                "my-backend2",
		BackendType:         "vault",
		TokenRotateInterval: &rotateInternal2,
	}, &nextRotateTime2)

	changes, err := s.state.GetSecretBackendRotateChanges(context.Background(), backendID1, backendID2)
	c.Assert(err, gc.IsNil)
	c.Assert(changes, gc.HasLen, 2)
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Name < changes[j].Name
	})
	c.Assert(changes[0].ID, gc.Equals, backendID1)
	c.Assert(changes[0].Name, gc.Equals, "my-backend1")
	c.Assert(changes[0].NextTriggerTime.Equal(nextRotateTime1), jc.IsTrue)
	c.Assert(changes[1].ID, gc.Equals, backendID2)
	c.Assert(changes[1].Name, gc.Equals, "my-backend2")
	c.Assert(changes[1].NextTriggerTime.Equal(nextRotateTime2), jc.IsTrue)
}