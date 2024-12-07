// Copyright 2024 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package lease

import "github.com/juju/juju/core/model"

// LeaseCheckerWaiter is an interface that checks and waits if a lease is held
// by a holder.
type LeaseCheckerWaiter interface {
	Waiter
	Checker
}

// LeaseManagerGetter is an interface that provides a method to get a lease
// manager for a given lease using its UUID. The lease namespace could be a
// model or an application.
type LeaseManagerGetter interface {
	// GetLeaseManager returns a lease manager for the given model UUID.
	GetLeaseManager(model.UUID) (LeaseCheckerWaiter, error)
}

// ModelLeaseManagerGetter is an interface that provides a method to
// get a lease manager in the scope of a model.
type ModelLeaseManagerGetter interface {
	// GetLeaseManager returns a lease manager for the given model UUID.
	GetLeaseManager() (LeaseCheckerWaiter, error)
}
