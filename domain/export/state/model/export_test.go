// Copyright 2025 Canonical Ltd. All rights reserved.
// Licensed under the AGPLv3, see LICENCE file for details.

package model

import (
	"testing"

	schematesting "github.com/juju/juju/domain/schema/testing"
	"github.com/juju/tc"
)

type exportStateSuiteV0 struct {
	schematesting.ModelSuite
}

func TestExportStateSuiteV0(t *testing.T) {
	tc.Run(t, &exportStateSuiteV0{})
}

func (s *exportStateSuiteV0) TestExportRuns(c *tc.C) {
	st := NewState(s.TxnRunnerFactory())
	_, err := st.ExportV0(c.Context())
	c.Assert(err, tc.ErrorIsNil)
}
