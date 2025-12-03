// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package controllerpresence

import (
	"context"
	"testing"

	"github.com/juju/clock"
	"github.com/juju/tc"
	"github.com/juju/worker/v4/workertest"
	"go.uber.org/goleak"
	gomock "go.uber.org/mock/gomock"

	"github.com/juju/juju/api"
	coreerrors "github.com/juju/juju/core/errors"
	loggertesting "github.com/juju/juju/internal/logger/testing"
	"github.com/juju/juju/internal/testhelpers"
	apiremotecaller "github.com/juju/juju/internal/worker/apiremotecaller"
)

func TestWorker(t *testing.T) {
	defer goleak.VerifyNone(t)
	tc.Run(t, &WorkerSuite{})
}

type WorkerSuite struct {
	testhelpers.IsolationSuite

	statusService       *MockStatusService
	apiRemoteSubscriber *MockAPIRemoteSubscriber
	subscription        *MockSubscription
}

func (s *WorkerSuite) TestValidate(c *tc.C) {
	defer s.setupMocks(c).Finish()

	err := s.newConfig(c).Validate()
	c.Assert(err, tc.IsNil)

	config := s.newConfig(c)
	config.StatusService = nil
	err = config.Validate()
	c.Assert(err, tc.ErrorIs, coreerrors.NotValid)

	config = s.newConfig(c)
	config.APIRemoteSubscriber = nil
	err = config.Validate()
	c.Assert(err, tc.ErrorIs, coreerrors.NotValid)

	config = s.newConfig(c)
	config.Logger = nil
	err = config.Validate()
	c.Assert(err, tc.ErrorIs, coreerrors.NotValid)

	config = s.newConfig(c)
	config.Clock = nil
	err = config.Validate()
	c.Assert(err, tc.ErrorIs, coreerrors.NotValid)
}

func (s *WorkerSuite) TestWorkerNoInitialRemotes(c *tc.C) {
	defer s.setupMocks(c).Finish()

	done := make(chan struct{})
	s.apiRemoteSubscriber.EXPECT().Subscribe().Return(s.subscription, nil)
	s.apiRemoteSubscriber.EXPECT().GetAPIRemotes().DoAndReturn(func() ([]apiremotecaller.RemoteConnection, error) {
		return nil, nil
	})
	s.subscription.EXPECT().Changes().DoAndReturn(func() <-chan struct{} {
		close(done)
		return nil
	})
	s.subscription.EXPECT().Close()

	w := s.newWorker(c)
	defer workertest.DirtyKill(c, w)

	select {
	case <-done:
	case <-c.Context().Done():
		c.Fatal("worker did not start")
	}

	workertest.CleanKill(c, w)
}

func (s *WorkerSuite) TestWorkerInitialRemotes(c *tc.C) {
	defer s.setupMocks(c).Finish()

	// Wait for the worker to create the initial connection.

	done := make(chan struct{})
	wait := make(chan struct{})
	s.apiRemoteSubscriber.EXPECT().Subscribe().Return(s.subscription, nil)
	s.apiRemoteSubscriber.EXPECT().GetAPIRemotes().DoAndReturn(func() ([]apiremotecaller.RemoteConnection, error) {
		return []apiremotecaller.RemoteConnection{remoteConnection{
			controllerID: "0",
			fn: func() error {
				close(done)
				<-wait
				return nil
			},
		}}, nil
	})
	s.subscription.EXPECT().Changes().Return(make(<-chan struct{}))
	s.subscription.EXPECT().Close()

	w := s.newWorker(c)
	defer workertest.DirtyKill(c, w)

	select {
	case <-done:
	case <-c.Context().Done():
		c.Fatal("worker did not start")
	}

	c.Assert(w.runner.WorkerNames(), tc.DeepEquals, []string{"controller-0"})

	close(wait)

	workertest.CleanKill(c, w)
}

func (s *WorkerSuite) TestWorkerRemotesSubscription(c *tc.C) {
	defer s.setupMocks(c).Finish()

	// Wait for the worker to create the initial connection.

	first := make(chan struct{})
	done := make(chan struct{})
	ch := make(chan struct{})

	wait0 := make(chan struct{})
	wait1 := make(chan struct{})

	s.apiRemoteSubscriber.EXPECT().Subscribe().Return(s.subscription, nil)
	s.subscription.EXPECT().Close()

	gomock.InOrder(
		s.apiRemoteSubscriber.EXPECT().GetAPIRemotes().DoAndReturn(func() ([]apiremotecaller.RemoteConnection, error) {
			return []apiremotecaller.RemoteConnection{remoteConnection{
				controllerID: "0",
				fn: func() error {
					close(first)
					<-wait0
					return nil
				},
			}}, nil
		}),
		s.subscription.EXPECT().Changes().Return(ch),
		s.apiRemoteSubscriber.EXPECT().GetAPIRemotes().DoAndReturn(func() ([]apiremotecaller.RemoteConnection, error) {
			return []apiremotecaller.RemoteConnection{remoteConnection{
				controllerID: "1",
				fn: func() error {
					close(done)
					<-wait1
					return nil
				},
			}}, nil
		}),
		s.subscription.EXPECT().Changes().Return(ch),
	)

	w := s.newWorker(c)
	defer workertest.DirtyKill(c, w)

	select {
	case <-first:
	case <-c.Context().Done():
		c.Fatal("worker did not start")
	}

	c.Assert(w.runner.WorkerNames(), tc.DeepEquals, []string{"controller-0"})

	close(wait0)

	select {
	case ch <- struct{}{}:
	case <-c.Context().Done():
		c.Fatal("could not send change")
	}

	select {
	case <-done:
	case <-c.Context().Done():
		c.Fatal("worker did not start")
	}

	c.Assert(w.runner.WorkerNames(), tc.DeepEquals, []string{"controller-1"})

	close(wait1)

	workertest.CleanKill(c, w)
}

func (s *WorkerSuite) newConfig(c *tc.C) WorkerConfig {
	return WorkerConfig{
		StatusService:       s.statusService,
		APIRemoteSubscriber: s.apiRemoteSubscriber,
		Clock:               clock.WallClock,
		Logger:              loggertesting.WrapCheckLog(c),
	}
}

func (s *WorkerSuite) newWorker(c *tc.C) *controllerWorker {
	worker, err := newWorker(s.newConfig(c))
	c.Assert(err, tc.IsNil)
	return worker.(*controllerWorker)
}

func (s *WorkerSuite) setupMocks(c *tc.C) *gomock.Controller {
	mockCtrl := gomock.NewController(c)
	s.statusService = NewMockStatusService(mockCtrl)
	s.apiRemoteSubscriber = NewMockAPIRemoteSubscriber(mockCtrl)
	s.subscription = NewMockSubscription(mockCtrl)

	c.Cleanup(func() {
		s.statusService = nil
		s.apiRemoteSubscriber = nil
		s.subscription = nil
	})

	return mockCtrl
}

type remoteConnection struct {
	controllerID string
	fn           func() error
}

func (r remoteConnection) ControllerID() string {
	return r.controllerID
}

func (r remoteConnection) Connection(ctx context.Context, fn func(ctx context.Context, c api.Connection) error) error {
	return r.fn()
}
