// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package apicaller_test

import (
	"context"

	"github.com/juju/errors"
	"github.com/juju/names/v6"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/worker/v4"
	"github.com/juju/worker/v4/dependency"
	dt "github.com/juju/worker/v4/dependency/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/agent"
	"github.com/juju/juju/api"
	"github.com/juju/juju/api/base"
	"github.com/juju/juju/core/logger"
	loggertesting "github.com/juju/juju/internal/logger/testing"
	coretesting "github.com/juju/juju/internal/testing"
	"github.com/juju/juju/internal/worker/apicaller"
)

type ManifoldSuite struct {
	testing.IsolationSuite
	testing.Stub

	manifold       dependency.Manifold
	manifoldConfig apicaller.ManifoldConfig
	agent          *mockAgent
	conn           *mockConn
	getter         dependency.Getter
}

var _ = gc.Suite(&ManifoldSuite{})

func (s *ManifoldSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.Stub = testing.Stub{}
	s.manifoldConfig = apicaller.ManifoldConfig{
		AgentName:            "agent-name",
		APIConfigWatcherName: "api-config-watcher-name",
		APIOpen: func(context.Context, *api.Info, api.DialOpts) (api.Connection, error) {
			panic("just a fake")
		},
		NewConnection: func(_ context.Context, a agent.Agent, apiOpen api.OpenFunc, logger logger.Logger) (api.Connection, error) {
			c.Check(apiOpen, gc.NotNil) // uncomparable
			c.Check(logger, gc.NotNil)  // uncomparable
			s.AddCall("NewConnection", a)
			if err := s.NextErr(); err != nil {
				return nil, err
			}
			return s.conn, nil
		},
		Filter: func(err error) error {
			panic(err)
		},
		Logger: loggertesting.WrapCheckLog(c),
	}
	s.manifold = apicaller.Manifold(s.manifoldConfig)
	checkFilter := func() {
		s.manifold.Filter(errors.New("arrgh"))
	}
	c.Check(checkFilter, gc.PanicMatches, "arrgh")

	s.agent = &mockAgent{
		stub:   &s.Stub,
		model:  coretesting.ModelTag,
		entity: names.NewMachineTag("42"),
	}
	s.getter = dt.StubGetter(map[string]interface{}{
		"agent-name": s.agent,
	})

	// Watch out for this: it uses its own Stub because Close calls
	// are made from the worker's loop goroutine. You should make
	// sure to stop the worker before checking the mock conn's calls.
	s.conn = &mockConn{
		stub:   &testing.Stub{},
		broken: make(chan struct{}),
	}
}

func (s *ManifoldSuite) TestInputsOptionalConfigPropertiesUnset(c *gc.C) {
	s.manifoldConfig.APIConfigWatcherName = ""
	c.Check(apicaller.Manifold(s.manifoldConfig).Inputs, jc.DeepEquals, []string{
		"agent-name",
	})
}

func (s *ManifoldSuite) TestInputs(c *gc.C) {
	c.Check(s.manifold.Inputs, jc.DeepEquals, []string{
		"agent-name",
		"api-config-watcher-name",
	})
}

func (s *ManifoldSuite) TestStartMissingAgent(c *gc.C) {
	getter := dt.StubGetter(map[string]interface{}{
		"agent-name": dependency.ErrMissing,
	})

	worker, err := s.manifold.Start(context.Background(), getter)
	c.Check(worker, gc.IsNil)
	c.Check(err, gc.Equals, dependency.ErrMissing)
	s.CheckCalls(c, nil)
}

func (s *ManifoldSuite) TestStartCannotOpenAPI(c *gc.C) {
	s.SetErrors(errors.New("no api for you"))

	worker, err := s.manifold.Start(context.Background(), s.getter)
	c.Check(worker, gc.IsNil)
	c.Check(err, gc.ErrorMatches, `\[deadbe\] "machine-42" cannot open api: no api for you`)
	s.CheckCalls(c, []testing.StubCall{{
		FuncName: "NewConnection",
		Args:     []interface{}{s.agent},
	}})
}

func (s *ManifoldSuite) TestStartSuccess(c *gc.C) {
	worker, err := s.manifold.Start(context.Background(), s.getter)
	c.Check(err, jc.ErrorIsNil)
	defer assertStop(c, worker)
	s.CheckCalls(c, []testing.StubCall{{
		FuncName: "NewConnection",
		Args:     []interface{}{s.agent},
	}})
}

func (s *ManifoldSuite) setupWorkerTest(c *gc.C) worker.Worker {
	w, err := s.manifold.Start(context.Background(), s.getter)
	c.Assert(err, jc.ErrorIsNil)
	s.AddCleanup(func(c *gc.C) { w.Kill() })
	return w
}

func (s *ManifoldSuite) TestKillWorkerClosesConnection(c *gc.C) {
	worker := s.setupWorkerTest(c)
	assertStop(c, worker)
	s.conn.stub.CheckCalls(c, []testing.StubCall{{
		FuncName: "Close",
	}})
}

func (s *ManifoldSuite) TestKillWorkerReportsCloseErr(c *gc.C) {
	s.conn.stub.SetErrors(errors.New("bad plumbing"))
	worker := s.setupWorkerTest(c)

	assertStopError(c, worker, "bad plumbing")
	s.conn.stub.CheckCalls(c, []testing.StubCall{{
		FuncName: "Close",
	}})
}

func (s *ManifoldSuite) TestBrokenConnectionKillsWorkerWithCloseErr(c *gc.C) {
	s.conn.stub.SetErrors(errors.New("bad plumbing"))
	worker := s.setupWorkerTest(c)

	close(s.conn.broken)
	err := worker.Wait()
	c.Check(err, gc.ErrorMatches, "bad plumbing")
	s.conn.stub.CheckCalls(c, []testing.StubCall{{
		FuncName: "Close",
	}})
}

func (s *ManifoldSuite) TestBrokenConnectionKillsWorkerWithFallbackErr(c *gc.C) {
	worker := s.setupWorkerTest(c)

	close(s.conn.broken)
	err := worker.Wait()
	c.Check(err, gc.ErrorMatches, "api connection broken unexpectedly")
	s.conn.stub.CheckCalls(c, []testing.StubCall{{
		FuncName: "Close",
	}})
}

func (s *ManifoldSuite) TestOutputSuccess(c *gc.C) {
	worker := s.setupWorkerTest(c)

	var apicaller base.APICaller
	err := s.manifold.Output(worker, &apicaller)
	c.Check(err, jc.ErrorIsNil)
	c.Check(apicaller, gc.Equals, s.conn)

	var conn api.Connection
	err = s.manifold.Output(worker, &conn)
	c.Check(err, jc.ErrorIsNil)
	c.Check(conn, gc.Equals, s.conn)
}

func (s *ManifoldSuite) TestOutputBadWorker(c *gc.C) {
	var apicaller base.APICaller
	err := s.manifold.Output(dummyWorker{}, &apicaller)
	c.Check(apicaller, gc.IsNil)
	c.Check(err.Error(), gc.Equals, "in should be a *apicaller.apiConnWorker; got apicaller_test.dummyWorker")
}

func (s *ManifoldSuite) TestOutputBadTarget(c *gc.C) {
	worker := s.setupWorkerTest(c)

	var apicaller interface{}
	err := s.manifold.Output(worker, &apicaller)
	c.Check(apicaller, gc.IsNil)
	c.Check(err.Error(), gc.Equals, "out should be *base.APICaller or *api.Connection; got *interface {}")
}
