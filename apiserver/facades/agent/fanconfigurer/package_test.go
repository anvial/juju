// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fanconfigurer_test

import (
	stdtesting "testing"

	"github.com/juju/juju/internal/testing"
)

func TestAll(t *stdtesting.T) {
	testing.MgoTestPackage(t)
}
