// Copyright 2019 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package upgradesteps_test

import (
	"context"

	"github.com/juju/names/v5"
	jc "github.com/juju/testing/checkers"
	"go.uber.org/mock/gomock"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/api/agent/upgradesteps"
	"github.com/juju/juju/api/base/mocks"
	jujutesting "github.com/juju/juju/internal/testing"
	"github.com/juju/juju/rpc/params"
)

type upgradeStepsSuite struct {
	jujutesting.BaseSuite

	fCaller *mocks.MockFacadeCaller
}

var _ = gc.Suite(&upgradeStepsSuite{})

func (s *upgradeStepsSuite) SetUpTest(c *gc.C) {
	s.BaseSuite.SetUpTest(c)
}

func (s *upgradeStepsSuite) TestWriteAgentState(c *gc.C) {
	defer s.setupMocks(c).Finish()

	uTag0 := names.NewUnitTag("test/0")
	uTag1 := names.NewUnitTag("test/1")
	str0 := "foo"
	str1 := "bar"
	args := params.SetUnitStateArgs{[]params.SetUnitStateArg{
		{Tag: uTag0.String(), UniterState: &str0},
		{Tag: uTag1.String(), UniterState: &str1},
	}}
	s.expectWriteAgentStateSuccess(c, args)

	client := upgradesteps.NewClientFromFacade(s.fCaller)
	err := client.WriteAgentState(context.Background(), []params.SetUnitStateArg{
		{Tag: uTag0.String(), UniterState: &str0},
		{Tag: uTag1.String(), UniterState: &str1},
	})
	c.Assert(err, jc.ErrorIsNil)
}

func (s *upgradeStepsSuite) TestWriteAgentStateError(c *gc.C) {
	defer s.setupMocks(c).Finish()

	uTag0 := names.NewUnitTag("test/0")
	str0 := "foo"
	args := params.SetUnitStateArgs{[]params.SetUnitStateArg{
		{Tag: uTag0.String(), UniterState: &str0},
	}}
	s.expectWriteAgentStateError(c, args)

	client := upgradesteps.NewClientFromFacade(s.fCaller)
	err := client.WriteAgentState(context.Background(), []params.SetUnitStateArg{
		{Tag: uTag0.String(), UniterState: &str0},
	})
	c.Assert(err, gc.ErrorMatches, "did not find")
}

func (s *upgradeStepsSuite) setupMocks(c *gc.C) *gomock.Controller {
	ctrl := gomock.NewController(c)
	s.fCaller = mocks.NewMockFacadeCaller(ctrl)
	return ctrl
}

func (s *upgradeStepsSuite) expectWriteAgentStateSuccess(c *gc.C, args params.SetUnitStateArgs) {
	fExp := s.fCaller.EXPECT()
	resultSource := params.ErrorResults{}
	fExp.FacadeCall(gomock.Any(), "WriteAgentState", unitStateMatcher{c, args}, gomock.Any()).SetArg(3, resultSource)
}

func (s *upgradeStepsSuite) expectWriteAgentStateError(c *gc.C, args params.SetUnitStateArgs) {
	fExp := s.fCaller.EXPECT()
	resultSource := params.ErrorResults{Results: []params.ErrorResult{{
		Error: &params.Error{
			Code:    params.CodeNotFound,
			Message: "did not find",
		},
	}}}
	fExp.FacadeCall(gomock.Any(), "WriteAgentState", unitStateMatcher{c, args}, gomock.Any()).SetArg(3, resultSource)
}

type unitStateMatcher struct {
	c        *gc.C
	expected params.SetUnitStateArgs
}

func (m unitStateMatcher) Matches(x interface{}) bool {
	obtained, ok := x.(params.SetUnitStateArgs)
	if !ok {
		return false
	}

	m.c.Assert(obtained.Args, gc.HasLen, len(m.expected.Args))

	for _, obt := range obtained.Args {
		var found bool
		for _, exp := range m.expected.Args {
			if obt.Tag == exp.Tag {
				m.c.Assert(obt, jc.DeepEquals, exp)
				found = true
			}
		}
		m.c.Assert(found, jc.IsTrue, gc.Commentf("obtained tag %s, not found in expected data", obt.Tag))
	}

	return true
}

func (m unitStateMatcher) String() string {
	return "Match the contents of the UniterState pointer in params.SetUnitStateArg"
}
