// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package operation_test

import (
	stdcontext "context"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/internal/charm/hooks"
	loggertesting "github.com/juju/juju/internal/logger/testing"
	"github.com/juju/juju/internal/worker/uniter/hook"
	"github.com/juju/juju/internal/worker/uniter/operation"
)

type LeaderSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&LeaderSuite{})

func (s *LeaderSuite) newFactory(c *gc.C) operation.Factory {
	return operation.NewFactory(operation.FactoryParams{
		Logger: loggertesting.WrapCheckLog(c),
	})
}

func (s *LeaderSuite) TestAcceptLeadership_Prepare_BadState(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewAcceptLeadership()
	c.Assert(err, jc.ErrorIsNil)

	newState, err := op.Prepare(stdcontext.Background(), operation.State{})
	c.Check(newState, gc.IsNil)
	// accept is only valid in Continue mode, when we're sure nothing is queued
	// or in progress.
	c.Check(err, gc.Equals, operation.ErrCannotAcceptLeadership)
}

func (s *LeaderSuite) TestAcceptLeadership_Prepare_NotLeader(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewAcceptLeadership()
	c.Assert(err, jc.ErrorIsNil)

	newState, err := op.Prepare(stdcontext.Background(), operation.State{Kind: operation.Continue})
	c.Check(newState, gc.IsNil)
	// *execute* is currently just a no-op -- all the meat happens in commit.
	c.Check(err, gc.Equals, operation.ErrSkipExecute)
}

func (s *LeaderSuite) TestAcceptLeadership_Prepare_AlreadyLeader(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewAcceptLeadership()
	c.Assert(err, jc.ErrorIsNil)

	newState, err := op.Prepare(stdcontext.Background(), operation.State{
		Kind:   operation.Continue,
		Leader: true,
	})
	c.Check(newState, gc.IsNil)
	// *execute* is currently just a no-op -- all the meat happens in commit.
	c.Check(err, gc.Equals, operation.ErrSkipExecute)
}

func (s *LeaderSuite) TestAcceptLeadership_Commit_NotLeader_BlankSlate(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewAcceptLeadership()
	c.Assert(err, jc.ErrorIsNil)
	_, err = op.Prepare(stdcontext.Background(), operation.State{Kind: operation.Continue})
	c.Check(err, gc.Equals, operation.ErrSkipExecute)

	newState, err := op.Commit(stdcontext.Background(), operation.State{
		Kind: operation.Continue,
	})
	c.Check(err, jc.ErrorIsNil)
	c.Check(newState, gc.DeepEquals, &operation.State{
		Kind:   operation.RunHook,
		Step:   operation.Queued,
		Hook:   &hook.Info{Kind: hooks.LeaderElected},
		Leader: true,
	})
}

func (s *LeaderSuite) TestAcceptLeadership_Commit_NotLeader_Preserve(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewAcceptLeadership()
	c.Assert(err, jc.ErrorIsNil)
	_, err = op.Prepare(stdcontext.Background(), operation.State{Kind: operation.Continue})
	c.Check(err, gc.Equals, operation.ErrSkipExecute)

	newState, err := op.Commit(stdcontext.Background(), operation.State{
		Kind:    operation.Continue,
		Started: true,
		Hook:    &hook.Info{Kind: hooks.Install},
	})
	c.Check(err, jc.ErrorIsNil)
	c.Check(newState, gc.DeepEquals, &operation.State{
		Kind:    operation.RunHook,
		Step:    operation.Queued,
		Hook:    &hook.Info{Kind: hooks.LeaderElected},
		Leader:  true,
		Started: true,
	})
}

func (s *LeaderSuite) TestAcceptLeadership_Commit_AlreadyLeader(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewAcceptLeadership()
	c.Assert(err, jc.ErrorIsNil)
	_, err = op.Prepare(stdcontext.Background(), operation.State{Kind: operation.Continue})
	c.Check(err, gc.Equals, operation.ErrSkipExecute)

	newState, err := op.Commit(stdcontext.Background(), operation.State{
		Kind:   operation.Continue,
		Leader: true,
	})
	c.Check(newState, gc.IsNil)
	c.Check(err, jc.ErrorIsNil)
}

func (s *LeaderSuite) TestAcceptLeadership_DoesNotNeedGlobalMachineLock(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewAcceptLeadership()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(op.NeedsGlobalMachineLock(), jc.IsFalse)
}

func (s *LeaderSuite) TestResignLeadership_Prepare_Leader(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewResignLeadership()
	c.Assert(err, jc.ErrorIsNil)

	newState, err := op.Prepare(stdcontext.Background(), operation.State{Leader: true})
	c.Check(newState, gc.IsNil)
	c.Check(err, jc.ErrorIsNil)
}

func (s *LeaderSuite) TestResignLeadership_Prepare_NotLeader(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewResignLeadership()
	c.Assert(err, jc.ErrorIsNil)

	newState, err := op.Prepare(stdcontext.Background(), operation.State{})
	c.Check(newState, gc.IsNil)
	c.Check(err, gc.Equals, operation.ErrSkipExecute)
}

func (s *LeaderSuite) TestResignLeadership_Execute(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewResignLeadership()
	c.Assert(err, jc.ErrorIsNil)

	_, err = op.Prepare(stdcontext.Background(), operation.State{Leader: true})
	c.Check(err, jc.ErrorIsNil)

	// Execute is a no-op (which logs that we should run leader-deposed)
	newState, err := op.Execute(stdcontext.Background(), operation.State{})
	c.Check(newState, gc.IsNil)
	c.Check(err, jc.ErrorIsNil)
}

func (s *LeaderSuite) TestResignLeadership_Commit_ClearLeader(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewResignLeadership()
	c.Assert(err, jc.ErrorIsNil)

	newState, err := op.Commit(stdcontext.Background(), operation.State{Leader: true})
	c.Check(newState, gc.DeepEquals, &operation.State{})
	c.Check(err, jc.ErrorIsNil)
}

func (s *LeaderSuite) TestResignLeadership_Commit_PreserveOthers(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewResignLeadership()
	c.Assert(err, jc.ErrorIsNil)

	newState, err := op.Commit(stdcontext.Background(), overwriteState)
	c.Check(newState, gc.DeepEquals, &overwriteState)
	c.Check(err, jc.ErrorIsNil)
}

func (s *LeaderSuite) TestResignLeadership_Commit_All(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewResignLeadership()
	c.Assert(err, jc.ErrorIsNil)

	leaderState := overwriteState
	leaderState.Leader = true
	newState, err := op.Commit(stdcontext.Background(), leaderState)
	c.Check(newState, gc.DeepEquals, &overwriteState)
	c.Check(err, jc.ErrorIsNil)
}

func (s *LeaderSuite) TestResignLeadership_DoesNotNeedGlobalMachineLock(c *gc.C) {
	factory := s.newFactory(c)
	op, err := factory.NewResignLeadership()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(op.NeedsGlobalMachineLock(), jc.IsFalse)
}
