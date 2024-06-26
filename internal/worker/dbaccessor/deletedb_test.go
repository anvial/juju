// Copyright 2024 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package dbaccessor

import (
	"context"
	"database/sql"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/domain/schema"
	"github.com/juju/juju/internal/database"
	databasetesting "github.com/juju/juju/internal/database/testing"
	loggertesting "github.com/juju/juju/internal/logger/testing"
)

type deleteDBSuite struct {
	databasetesting.DqliteSuite
}

var _ = gc.Suite(&deleteDBSuite{})

func (s *deleteDBSuite) TestDeleteDBContentsOnEmptyDB(c *gc.C) {
	runner := s.TxnRunner()

	err := runner.StdTxn(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		return deleteDBContents(ctx, tx, loggertesting.WrapCheckLog(c))
	})
	c.Assert(err, gc.IsNil)
}

func (s *deleteDBSuite) TestDeleteDBContentsOnControllerDB(c *gc.C) {
	runner, db := s.OpenDBForNamespace(c, "controller-foo", false)
	logger := loggertesting.WrapCheckLog(c)

	// This test isn't necessarily, as you can't delete the controller database
	// contents, but adds more validation to the function.

	err := database.NewDBMigration(
		runner, logger, schema.ControllerDDL()).Apply(context.Background())
	c.Assert(err, jc.ErrorIsNil)

	err = runner.StdTxn(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		return deleteDBContents(ctx, tx, logger)
	})
	c.Assert(err, gc.IsNil)

	s.ensureEmpty(c, db)
}

func (s *deleteDBSuite) TestDeleteDBContentsOnModelDB(c *gc.C) {
	runner, db := s.OpenDBForNamespace(c, "model-foo", false)

	logger := loggertesting.WrapCheckLog(c)

	err := database.NewDBMigration(
		runner, logger, schema.ModelDDL()).Apply(context.Background())
	c.Assert(err, jc.ErrorIsNil)

	err = runner.StdTxn(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		return deleteDBContents(ctx, tx, logger)
	})
	c.Assert(err, gc.IsNil)

	s.ensureEmpty(c, db)
}

func (s *deleteDBSuite) ensureEmpty(c *gc.C, db *sql.DB) {
	schemaStmt := `SELECT COUNT(*) FROM sqlite_master WHERE name NOT LIKE 'sqlite_%';`
	var count int
	err := db.QueryRow(schemaStmt).Scan(&count)
	c.Assert(err, jc.ErrorIsNil)
	c.Check(count, gc.Equals, 0)
}
