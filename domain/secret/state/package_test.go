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

func (s *baseSuite) getApplicationUUID(ctx context.Context, appName string) (coreapplication.UUID, error) {
	var uuid coreapplication.UUID
	err := s.state.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		var err error
		uuid, err = s.state.GetApplicationUUID(ctx, appName)
		return err
	})
	return uuid, err
}

func (s *baseSuite) getUnitUUID(ctx context.Context, unitName coreunit.Name) (coreunit.UUID, error) {
	var uuid coreunit.UUID
	err := s.state.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		var err error
		uuid, err = s.state.GetUnitUUID(ctx, unitName)
		return err
	})
	return uuid, err
}

func (s *baseSuite) checkUserSecretLabelExists(ctx context.Context, label string) (bool, error) {
	var exists bool
	err := s.state.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		var err error
		exists, err = s.state.CheckUserSecretLabelExists(ctx, label)
		return err
	})
	return exists, err
}

func (s *baseSuite) checkApplicationSecretLabelExists(ctx context.Context, appUUID coreapplication.UUID,
	label string) (bool, error) {
	var exists bool
	err := s.state.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		var err error
		exists, err = s.state.CheckApplicationSecretLabelExists(ctx, appUUID, label)
		return err
	})
	return exists, err
}

func (s *baseSuite) checkUnitSecretLabelExists(ctx context.Context, unitUUID coreunit.UUID, label string) (bool, error) {
	var exists bool
	err := s.state.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		var err error
		exists, err = s.state.CheckUnitSecretLabelExists(ctx, unitUUID, label)
		return err
	})
	return exists, err
}

func (s *baseSuite) createUserSecret(ctx context.Context, version int, uri *coresecrets.URI, secret domainsecret.UpsertSecretParams) error {
	return s.state.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		return s.state.CreateUserSecret(ctx, version, uri, secret)
	})
}

func (s *baseSuite) createCharmApplicationSecret(ctx context.Context, version int, uri *coresecrets.URI, appName string, secret domainsecret.UpsertSecretParams) error {
	return s.state.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		appUUID, err := s.state.GetApplicationUUID(ctx, appName)
		if err != nil {
			return err
		}
		return s.state.CreateCharmApplicationSecret(ctx, version, uri, appUUID, secret)
	})
}

func (s *baseSuite) createCharmUnitSecret(ctx context.Context, version int, uri *coresecrets.URI, unitName coreunit.Name,
	secret domainsecret.UpsertSecretParams) error {
	return s.state.RunAtomic(ctx, func(ctx domain.AtomicContext) error {
		unitUUID, err := s.state.GetUnitUUID(ctx, unitName)
		if err != nil {
			return err
		}
		return s.state.CreateCharmUnitSecret(ctx, version, uri, unitUUID, secret)
	})
}
