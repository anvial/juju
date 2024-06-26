// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package schema

import (
	"context"
	"database/sql"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"go.uber.org/mock/gomock"
	gc "gopkg.in/check.v1"

	databasetesting "github.com/juju/juju/internal/database/testing"
)

type patchSuite struct {
	testing.IsolationSuite

	tx *MockTx
}

var _ = gc.Suite(&patchSuite{})

func (s *patchSuite) TestPatchHash(c *gc.C) {
	defer s.setupMocks(c).Finish()

	patch := MakePatch("SELECT 1")
	c.Assert(patch, gc.NotNil)
	c.Assert(patch.hash, gc.Equals, "4ATr1bVTKkuFmEpi+K1IqBqjRgwcoHcB84YTXXLN7PU=")
}

func (s *patchSuite) TestPatchHashWithSpaces(c *gc.C) {
	defer s.setupMocks(c).Finish()

	patch := MakePatch(`
                SELECT 1
`)
	c.Assert(patch, gc.NotNil)
	c.Assert(patch.hash, gc.Equals, "4ATr1bVTKkuFmEpi+K1IqBqjRgwcoHcB84YTXXLN7PU=")
}

func (s *patchSuite) TestPatchRun(c *gc.C) {
	defer s.setupMocks(c).Finish()

	patch := MakePatch("SELECT * FROM schema_master", 1, 2, "a")

	s.tx.EXPECT().ExecContext(gomock.Any(), "SELECT * FROM schema_master", 1, 2, "a").Return(nil, nil)

	patch.run(context.Background(), s.tx)
}

func (s *patchSuite) setupMocks(c *gc.C) *gomock.Controller {
	ctrl := gomock.NewController(c)

	s.tx = NewMockTx(ctrl)

	return ctrl
}

type schemaSuite struct {
	databasetesting.DqliteSuite
}

var _ = gc.Suite(&schemaSuite{})

func (s *schemaSuite) TestSchemaAdd(c *gc.C) {
	schema := New(
		MakePatch("SELECT 1"),
		MakePatch("SELECT 2"),
	)
	c.Check(schema.Len(), gc.Equals, 2)

	schema.Add(MakePatch("SELECT 3"))
	c.Check(schema.Len(), gc.Equals, 3)
	schema.Add(MakePatch("SELECT 4"))
	c.Check(schema.Len(), gc.Equals, 4)
}

func (s *schemaSuite) TestEnsureWithNoPatches(c *gc.C) {
	schema := New()
	current, err := schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 0, Post: 0})
}

func (s *schemaSuite) TestSchemaRunMultipleTimes(c *gc.C) {
	schema := New(
		MakePatch("CREATE TEMP TABLE foo (id INTEGER PRIMARY KEY);"),
		MakePatch("CREATE TEMP TABLE bar (id INTEGER PRIMARY KEY);"),
	)
	current, err := schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 0, Post: 2})

	schema = New(
		MakePatch("CREATE TEMP TABLE foo (id INTEGER PRIMARY KEY);"),
		MakePatch("CREATE TEMP TABLE bar (id INTEGER PRIMARY KEY);"),
	)
	current, err = schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 2, Post: 2})
}

func (s *schemaSuite) TestSchemaRunMultipleTimesWithAdditions(c *gc.C) {
	schema := New(
		MakePatch("CREATE TEMP TABLE foo (id INTEGER PRIMARY KEY);"),
		MakePatch("CREATE TEMP TABLE bar (id INTEGER PRIMARY KEY);"),
	)
	current, err := schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 0, Post: 2})

	schema = New(
		MakePatch("CREATE TEMP TABLE foo (id INTEGER PRIMARY KEY);"),
		MakePatch("CREATE TEMP TABLE bar (id INTEGER PRIMARY KEY);"),
		MakePatch("CREATE TEMP TABLE baz (id INTEGER PRIMARY KEY);"),
	)
	current, err = schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 2, Post: 3})
}

func (s *schemaSuite) TestEnsure(c *gc.C) {
	schema := New(
		MakePatch("CREATE TEMP TABLE foo (id INTEGER PRIMARY KEY);"),
		MakePatch("CREATE TEMP TABLE bar (id INTEGER PRIMARY KEY);"),
	)
	current, err := schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 0, Post: 2})
}

func (s *schemaSuite) TestEnsureIdempotent(c *gc.C) {
	schema := New(
		MakePatch("CREATE TEMP TABLE foo (id INTEGER PRIMARY KEY);"),
		MakePatch("CREATE TEMP TABLE bar (id INTEGER PRIMARY KEY);"),
	)
	current, err := schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 0, Post: 2})

	current, err = schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 2, Post: 2})
}

func (s *schemaSuite) TestEnsureTwiceWithAdditionalChanges(c *gc.C) {
	schema := New(
		MakePatch("CREATE TEMP TABLE foo (id INTEGER PRIMARY KEY);"),
		MakePatch("CREATE TEMP TABLE bar (id INTEGER PRIMARY KEY);"),
	)
	current, err := schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 0, Post: 2})

	schema.Add(MakePatch("CREATE TEMP TABLE baz (id INTEGER PRIMARY KEY);"))

	current, err = schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 2, Post: 3})

	schema.Add(MakePatch("CREATE TEMP TABLE alice (id INTEGER PRIMARY KEY);"))

	current, err = schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 3, Post: 4})
}

func (s *schemaSuite) TestEnsureHashBreaks(c *gc.C) {
	schema := New(
		MakePatch("CREATE TEMP TABLE foo (id INTEGER PRIMARY KEY);"),
		MakePatch("CREATE TEMP TABLE bar (id INTEGER PRIMARY KEY);"),
	)
	current, err := schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.IsNil)
	c.Assert(current, gc.DeepEquals, ChangeSet{Current: 0, Post: 2})

	err = s.TxnRunner().StdTxn(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "UPDATE schema SET hash = 'blah' WHERE version=2;")
		return err
	})
	c.Assert(err, jc.ErrorIsNil)

	schema.Add(MakePatch("CREATE TEMP TABLE baz (id INTEGER PRIMARY KEY);"))

	_, err = schema.Ensure(context.Background(), s.TxnRunner())
	c.Assert(err, gc.ErrorMatches, `failed to query current schema version: hash mismatch for version 2`)
}
