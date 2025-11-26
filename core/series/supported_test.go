// Copyright 2020 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package series

import (
	"sort"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type SupportedSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&SupportedSuite{})

func (s *SupportedSuite) TestControllerSeries(c *gc.C) {
	info := map[SeriesName]seriesVersion{
		"supported": {
			WorkloadType: ControllerWorkloadType,
			Version:      "1.1.1",
			Supported:    true,
		},
		"ignored": {
			WorkloadType: ControllerWorkloadType,
			Version:      "1.1.1",
			Supported:    false,
		},
	}

	ctrlSeries := controllerSeries(info)
	sort.Strings(ctrlSeries)

	c.Assert(ctrlSeries, jc.DeepEquals, []string{"supported"})
}

func (s *SupportedSuite) TestWorkloadSeries(c *gc.C) {
	info := map[SeriesName]seriesVersion{
		"ctrl-supported": {
			WorkloadType: ControllerWorkloadType,
			Version:      "1.1.1",
			Supported:    true,
		},
		"ctrl-not-updated": {
			WorkloadType: ControllerWorkloadType,
			Version:      "1.1.1",
			Supported:    false,
		},
		"ctrl-ignored": {
			WorkloadType: ControllerWorkloadType,
			Version:      "1.1.1",
			Supported:    false,
		},
		"work-supported": {
			WorkloadType: OtherWorkloadType,
			Version:      "1.1.1",
			Supported:    true,
		},
		"work-ignored": {
			WorkloadType: OtherWorkloadType,
			Version:      "1.1.1",
			Supported:    false,
		},
	}

	workSeries := workloadSeries(info, false)
	sort.Strings(workSeries)

	c.Assert(workSeries, jc.DeepEquals, []string{"ctrl-supported", "work-supported"})

	// Double check that controller series doesn't change when we have workload
	// types.
	ctrlSeries := controllerSeries(info)
	sort.Strings(ctrlSeries)

	c.Assert(ctrlSeries, jc.DeepEquals, []string{"ctrl-supported"})
}
