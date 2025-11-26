// Copyright 2021 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package series

// These methods are used only in various tests.

// ESMSupportedJujuSeries returns a slice of just juju extended security
// maintenance supported ubuntu series.
func ESMSupportedJujuSeries() []string {
	var series []string
	for s, version := range ubuntuSeries {
		if !version.ESMSupported {
			continue
		}
		series = append(series, string(s))
	}
	return series
}

// SupportedJujuWorkloadSeries returns a slice of juju supported series that
// target a workload (deploying a charm).
func SupportedJujuWorkloadSeries() []string {
	var series []string
	for s, version := range allSeriesVersions {
		if !version.Supported || version.WorkloadType == UnsupportedWorkloadType {
			continue
		}
		series = append(series, string(s))
	}
	return series
}
