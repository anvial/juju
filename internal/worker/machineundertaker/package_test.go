// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package machineundertaker_test

import (
	"context"
	stdtesting "testing"

	gc "gopkg.in/check.v1"
)

func TestPackage(t *stdtesting.T) {
	gc.TestingT(t)
}

type fakeCredentialAPI struct{}

func (*fakeCredentialAPI) InvalidateModelCredential(_ context.Context, reason string) error {
	return nil
}
