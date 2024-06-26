// Copyright 2020 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package verifycharmprofile_test

import (
	"context"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/core/model"
	loggertesting "github.com/juju/juju/internal/logger/testing"
	"github.com/juju/juju/internal/worker/uniter/operation"
	"github.com/juju/juju/internal/worker/uniter/remotestate"
	"github.com/juju/juju/internal/worker/uniter/resolver"
	"github.com/juju/juju/internal/worker/uniter/verifycharmprofile"
)

type verifySuite struct{}

var _ = gc.Suite(&verifySuite{})

func (s *verifySuite) TestNextOpNotInstallNorUpgrade(c *gc.C) {
	local := resolver.LocalState{
		State: operation.State{Kind: operation.RunAction},
	}
	remote := remotestate.Snapshot{}
	res := newVerifyCharmProfileResolver(c)

	op, err := res.NextOp(context.Background(), local, remote, nil)
	c.Assert(err, gc.Equals, resolver.ErrNoOperation)
	c.Assert(op, gc.IsNil)
}

func (s *verifySuite) TestNextOpInstallProfileNotRequired(c *gc.C) {
	local := resolver.LocalState{
		State: operation.State{Kind: operation.Install},
	}
	remote := remotestate.Snapshot{
		CharmProfileRequired: false,
	}
	res := newVerifyCharmProfileResolver(c)

	op, err := res.NextOp(context.Background(), local, remote, nil)
	c.Assert(err, gc.Equals, resolver.ErrNoOperation)
	c.Assert(op, gc.IsNil)
}

func (s *verifySuite) TestNextOpInstallProfileRequiredEmptyName(c *gc.C) {
	local := resolver.LocalState{
		State: operation.State{Kind: operation.Install},
	}
	remote := remotestate.Snapshot{
		CharmProfileRequired: true,
	}
	res := newVerifyCharmProfileResolver(c)

	op, err := res.NextOp(context.Background(), local, remote, nil)
	c.Assert(err, gc.Equals, resolver.ErrDoNotProceed)
	c.Assert(op, gc.IsNil)
}

func (s *verifySuite) TestNextOpMisMatchCharmRevisions(c *gc.C) {
	local := resolver.LocalState{
		State: operation.State{Kind: operation.Upgrade},
	}
	remote := remotestate.Snapshot{
		CharmProfileRequired: true,
		LXDProfileName:       "juju-wordpress-74",
		CharmURL:             "ch:wordpress-75",
	}
	res := newVerifyCharmProfileResolver(c)

	op, err := res.NextOp(context.Background(), local, remote, nil)
	c.Assert(err, gc.Equals, resolver.ErrDoNotProceed)
	c.Assert(op, gc.IsNil)
}

func (s *verifySuite) TestNextOpMatchingCharmRevisions(c *gc.C) {
	local := resolver.LocalState{
		State: operation.State{Kind: operation.Upgrade},
	}
	remote := remotestate.Snapshot{
		CharmProfileRequired: true,
		LXDProfileName:       "juju-wordpress-75",
		CharmURL:             "ch:wordpress-75",
	}
	res := newVerifyCharmProfileResolver(c)

	op, err := res.NextOp(context.Background(), local, remote, nil)
	c.Assert(err, gc.Equals, resolver.ErrNoOperation)
	c.Assert(op, gc.IsNil)
}

func (s *verifySuite) TestNewResolverCAAS(c *gc.C) {
	r := verifycharmprofile.NewResolver(loggertesting.WrapCheckLog(c), model.CAAS)
	op, err := r.NextOp(context.Background(), resolver.LocalState{}, remotestate.Snapshot{}, nil)
	c.Assert(err, gc.Equals, resolver.ErrNoOperation)
	c.Assert(op, jc.ErrorIsNil)
}

func newVerifyCharmProfileResolver(c *gc.C) resolver.Resolver {
	return verifycharmprofile.NewResolver(loggertesting.WrapCheckLog(c), model.IAAS)
}
