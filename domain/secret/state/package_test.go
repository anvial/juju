// Copyright 2024 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state

import (
	"context"
	"database/sql"

	"github.com/canonical/sqlair"
	"github.com/juju/tc"

	coreapplication "github.com/juju/juju/core/application"
	coresecrets "github.com/juju/juju/core/secrets"
	coreunit "github.com/juju/juju/core/unit"
	"github.com/juju/juju/domain"
	schematesting "github.com/juju/juju/domain/schema/testing"
	domainsecret "github.com/juju/juju/domain/secret"
	loggertesting "github.com/juju/juju/internal/logger/testing"
)

type baseSuite struct {
	schematesting.ModelSuite
	state *State
}

func (s *baseSuite) SetUpTest(c *tc.C) {
	s.ModelSuite.SetUpTest(c)
	s.state = NewState(s.TxnRunnerFactory(), loggertesting.WrapCheckLog(c))

	c.Cleanup(func() {
		s.state = nil
	})
}

// txn executes a transactional function within a database context,
// ensuring proper error handling and assertion.
func (s *baseSuite) txn(c *tc.C, fn func(ctx context.Context, tx *sqlair.TX) error) error {
	return tc.Must1(c, s.state.DB, c.Context()).Txn(c.Context(), fn)
}

// queryRows returns rows as a slice of maps for the given query.
// This is intended to be used with SELECT statements for assertions.
func (s *baseSuite) queryRows(c *tc.C, query string, args ...interface{}) []map[string]interface{} {
	var results []map[string]interface{}
	err := s.TxnRunner().StdTxn(c.Context(), func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return err
		}

		for rows.Next() {
			values := make([]interface{}, len(cols))
			valuePtrs := make([]interface{}, len(cols))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return err
			}

			row := make(map[string]interface{})
			for i, col := range cols {
				row[col] = values[i]
			}
			results = append(results, row)
		}
		return rows.Err()
	})
	c.Assert(err, tc.IsNil, tc.Commentf("querying rows with query %q", query))
	return results
}

func getApplicationUUID(ctx context.Context, st *State, appName string) (coreapplication.UUID, error) {
	var uuid coreapplication.UUID
	err := st.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		var err error
		uuid, err = st.GetApplicationUUID(ctx, appName)
		return err
	})
	return uuid, err
}

func getUnitUUID(ctx context.Context, st *State, unitName coreunit.Name) (coreunit.UUID, error) {
	var uuid coreunit.UUID
	err := st.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		var err error
		uuid, err = st.GetUnitUUID(ctx, unitName)
		return err
	})
	return uuid, err
}

func checkUserSecretLabelExists(ctx context.Context, st *State, label string) (bool, error) {
	var exists bool
	err := st.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		var err error
		exists, err = st.CheckUserSecretLabelExists(ctx, label)
		return err
	})
	return exists, err
}

func checkApplicationSecretLabelExists(ctx context.Context, st *State, appUUID coreapplication.UUID, label string) (bool, error) {
	var exists bool
	err := st.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		var err error
		exists, err = st.CheckApplicationSecretLabelExists(ctx, appUUID, label)
		return err
	})
	return exists, err
}

func checkUnitSecretLabelExists(ctx context.Context, st *State, unitUUID coreunit.UUID, label string) (bool, error) {
	var exists bool
	err := st.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		var err error
		exists, err = st.CheckUnitSecretLabelExists(ctx, unitUUID, label)
		return err
	})
	return exists, err
}

func createUserSecret(ctx context.Context, st *State, version int, uri *coresecrets.URI, secret domainsecret.UpsertSecretParams) error {
	return st.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		return st.CreateUserSecret(ctx, version, uri, secret)
	})
}

func createCharmApplicationSecret(ctx context.Context, st *State, version int, uri *coresecrets.URI, appName string, secret domainsecret.UpsertSecretParams) error {
	return st.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		appUUID, err := st.GetApplicationUUID(ctx, appName)
		if err != nil {
			return err
		}
		return st.CreateCharmApplicationSecret(ctx, version, uri, appUUID, secret)
	})
}

func createCharmUnitSecret(ctx context.Context, st *State, version int, uri *coresecrets.URI, unitName coreunit.Name, secret domainsecret.UpsertSecretParams) error {
	return st.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		unitUUID, err := st.GetUnitUUID(ctx, unitName)
		if err != nil {
			return err
		}
		return st.CreateCharmUnitSecret(ctx, version, uri, unitUUID, secret)
	})
}
