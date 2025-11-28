// Copyright 2020 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package series

import (
	"sort"
	"strconv"
)

type namedSeriesVersion struct {
	Name          SeriesName
	SeriesVersion seriesVersion
	Version       float64
}

func makeNamedSeries(values map[SeriesName]seriesVersion) []namedSeriesVersion {
	res := make([]namedSeriesVersion, 0, len(values))
	for name, series := range values {
		ver, err := strconv.ParseFloat(series.Version, 10)
		if err != nil {
			ver = -1
		}

		res = append(res, namedSeriesVersion{
			Name:          name,
			SeriesVersion: series,
			Version:       ver,
		})
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Version > res[j].Version {
			return true
		}
		if res[i].Version < res[j].Version {
			return false
		}
		return res[i].Name < res[j].Name
	})

	return res
}

// controller.
func controllerSeries(series map[SeriesName]seriesVersion) []string {
	namedSeries := makeNamedSeries(series)
	var result []string
	for _, namedSeries := range namedSeries {
		version := namedSeries.SeriesVersion
		if version.WorkloadType != ControllerWorkloadType {
			continue
		}

		if version.ESMSupported || version.Supported {
			result = append(result, namedSeries.Name.String())
		}
	}
	return result
}

// workloadSeries returns a slice of series that are supported to run on a
// target workload (charm).
// Note: workload series will also include controller workload types, as they
// can also be used for workloads.
func workloadSeries(series map[SeriesName]seriesVersion, includeUnsupported bool) []string {
	namedSeries := makeNamedSeries(series)
	var result []string
	for _, namedSeries := range namedSeries {
		version := namedSeries.SeriesVersion
		if version.WorkloadType == UnsupportedWorkloadType {
			continue
		}
		if includeUnsupported || version.ESMSupported || version.Supported {
			result = append(result, namedSeries.Name.String())
		}
	}
	return result
}

// WorkloadType defines what type of workload the series is aimed at.
// Controllers only support Ubuntu systems.
type WorkloadType int

const (
	// ControllerWorkloadType defines a workload type that is for controllers
	// only.
	ControllerWorkloadType WorkloadType = iota

	// OtherWorkloadType workload type is for everything else.
	// In the future we might want to differentiate this.
	OtherWorkloadType

	// UnsupportedWorkloadType is used where the series does not support
	// running Juju agents.
	UnsupportedWorkloadType
)

// seriesVersion represents a ubuntu series that includes the version, if the
// series is an LTS and the supported defines if Juju supports the series
// version.
type seriesVersion struct {
	// WorkloadType defines what type the series version is intended to work
	// against.
	WorkloadType WorkloadType

	// Version represents the version of the series.
	Version string

	// LTS provides a lookup for a LTS series.  Like seriesVersions,
	// the values here are current at the time of writing.
	LTS bool

	// Supported defines if Juju classifies the series as officially supported.
	Supported bool

	// Extended security maintenance for customers, extends the supported bool
	// for how Juju classifies the series.
	ESMSupported bool
}

// setSupported updates a series map based on the series name.
func setSupported(series map[SeriesName]seriesVersion, name string) bool {
	if version, ok := series[SeriesName(name)]; ok {
		version.Supported = true
		series[SeriesName(name)] = version
		return true
	}
	return false
}

// SeriesName represents a series name for distros
type SeriesName string

func (s SeriesName) String() string {
	return string(s)
}

const (
	Precise SeriesName = "precise"
	Quantal SeriesName = "quantal"
	Raring  SeriesName = "raring"
	Saucy   SeriesName = "saucy"
	Trusty  SeriesName = "trusty"
	Utopic  SeriesName = "utopic"
	Vivid   SeriesName = "vivid"
	Wily    SeriesName = "wily"
	Xenial  SeriesName = "xenial"
	Yakkety SeriesName = "yakkety"
	Zesty   SeriesName = "zesty"
	Artful  SeriesName = "artful"
	Bionic  SeriesName = "bionic"
	Cosmic  SeriesName = "cosmic"
	Disco   SeriesName = "disco"
	Eoan    SeriesName = "eoan"
	Focal   SeriesName = "focal"
	Groovy  SeriesName = "groovy"
	Hirsute SeriesName = "hirsute"
	Impish  SeriesName = "impish"
	Jammy   SeriesName = "jammy"
	Kinetic SeriesName = "kinetic"
	Lunar   SeriesName = "lunar"
	Mantic  SeriesName = "mantic"
	Noble   SeriesName = "noble"
)

var ubuntuSeries = map[SeriesName]seriesVersion{
	Precise: {
		WorkloadType: ControllerWorkloadType,
		Version:      "12.04",
	},
	Quantal: {
		WorkloadType: ControllerWorkloadType,
		Version:      "12.10",
	},
	Raring: {
		WorkloadType: ControllerWorkloadType,
		Version:      "13.04",
	},
	Saucy: {
		WorkloadType: ControllerWorkloadType,
		Version:      "13.10",
	},
	Trusty: {
		WorkloadType: ControllerWorkloadType,
		Version:      "14.04",
		LTS:          true,
		ESMSupported: true,
	},
	Utopic: {
		WorkloadType: ControllerWorkloadType,
		Version:      "14.10",
	},
	Vivid: {
		WorkloadType: ControllerWorkloadType,
		Version:      "15.04",
	},
	Wily: {
		WorkloadType: ControllerWorkloadType,
		Version:      "15.10",
	},
	Xenial: {
		WorkloadType: ControllerWorkloadType,
		Version:      "16.04",
		LTS:          true,
		ESMSupported: true,
	},
	Yakkety: {
		WorkloadType: ControllerWorkloadType,
		Version:      "16.10",
	},
	Zesty: {
		WorkloadType: ControllerWorkloadType,
		Version:      "17.04",
	},
	Artful: {
		WorkloadType: ControllerWorkloadType,
		Version:      "17.10",
	},
	Bionic: {
		WorkloadType: ControllerWorkloadType,
		Version:      "18.04",
		LTS:          true,
		ESMSupported: true,
	},
	Cosmic: {
		WorkloadType: ControllerWorkloadType,
		Version:      "18.10",
	},
	Disco: {
		WorkloadType: ControllerWorkloadType,
		Version:      "19.04",
	},
	Eoan: {
		WorkloadType: ControllerWorkloadType,
		Version:      "19.10",
	},
	Focal: {
		WorkloadType: ControllerWorkloadType,
		Version:      "20.04",
		LTS:          true,
		Supported:    true,
		ESMSupported: true,
	},
	Groovy: {
		WorkloadType: ControllerWorkloadType,
		Version:      "20.10",
	},
	Hirsute: {
		WorkloadType: ControllerWorkloadType,
		Version:      "21.04",
	},
	Impish: {
		WorkloadType: ControllerWorkloadType,
		Version:      "21.10",
	},
	Jammy: {
		WorkloadType: ControllerWorkloadType,
		Version:      "22.04",
		LTS:          true,
		Supported:    true,
		ESMSupported: true,
	},
	Kinetic: {
		WorkloadType: ControllerWorkloadType,
		Version:      "22.10",
	},
	Lunar: {
		WorkloadType: ControllerWorkloadType,
		Version:      "23.04",
	},
	Mantic: {
		WorkloadType: ControllerWorkloadType,
		Version:      "23.10",
	},
	Noble: {
		WorkloadType: ControllerWorkloadType,
		Version:      "24.04",
		LTS:          true,
		Supported:    true,
		ESMSupported: true,
	},
}

const (
	Centos7      SeriesName = "centos7"
	Centos8      SeriesName = "centos8"
	Centos9      SeriesName = "centos9"
	OpenSUSELeap SeriesName = "opensuseleap"
	Kubernetes   SeriesName = "kubernetes"
)

var windowsVersions = map[string]seriesVersion{
	"Windows Server 2008 R2": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2008r2",
		Supported:    true,
	},
	"Hyper-V Server 2012 R2": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2012hvr2",
		Supported:    true,
	},
	"Hyper-V Server 2012": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2012hv",
		Supported:    true,
	},
	"Windows Server 2012 R2": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2012r2",
		Supported:    true,
	},
	"Windows Storage Server 2012 R2": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2012r2",
		Supported:    true,
	},
	"Windows Server 2012": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2012",
		Supported:    true,
	},
	"Windows Storage Server 2012": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2012",
		Supported:    true,
	},
	"Windows Server 2016": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2016",
		Supported:    true,
	},
	"Windows Storage Server 2016": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2016",
		Supported:    true,
	},
	"Hyper-V Server 2016": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2016hv",
		Supported:    true,
	},
	"Windows Server 2019": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2019",
		Supported:    true,
	},
	"Windows Storage Server 2019": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2019",
		Supported:    true,
	},
	"Windows 7": {
		WorkloadType: OtherWorkloadType,
		Version:      "win7",
		Supported:    true,
	},
	"Windows 8": {
		WorkloadType: OtherWorkloadType,
		Version:      "win8",
		Supported:    true,
	},
	"Windows 8.1": {
		WorkloadType: OtherWorkloadType,
		Version:      "win81",
		Supported:    true,
	},
	"Windows 10": {
		WorkloadType: OtherWorkloadType,
		Version:      "win10",
		Supported:    true,
	},
}

var windowsNanoVersions = map[SeriesName]seriesVersion{
	"Windows Server 2016": {
		WorkloadType: OtherWorkloadType,
		Version:      "win2016nano",
		Supported:    true,
	},
}

var centosSeries = map[SeriesName]seriesVersion{
	Centos7: {
		WorkloadType: OtherWorkloadType,
		Version:      "centos7",
		Supported:    true,
	},
	Centos8: {
		WorkloadType: OtherWorkloadType,
		Version:      "centos8",
		Supported:    true,
	},
	Centos9: {
		WorkloadType: OtherWorkloadType,
		Version:      "centos9",
		Supported:    true,
	},
}

var opensuseSeries = map[SeriesName]seriesVersion{
	OpenSUSELeap: {
		WorkloadType: OtherWorkloadType,
		Version:      "opensuse42",
		Supported:    true,
	},
}

var kubernetesSeries = map[SeriesName]seriesVersion{
	Kubernetes: {
		WorkloadType: OtherWorkloadType,
		Version:      "kubernetes",
		Supported:    true,
	},
}

var macOSXSeries = map[SeriesName]seriesVersion{
	"catalina": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "19",
		Supported:    true,
	},
	"mojave": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "18",
		Supported:    true,
	},
	"highsierra": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "17",
		Supported:    true,
	},
	"sierra": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "16",
		Supported:    true,
	},
	"elcapitan": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "15",
		Supported:    true,
	},
	"yosemite": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "14",
		Supported:    true,
	},
	"mavericks": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "13",
		Supported:    true,
	},
	"mountainlion": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "12",
		Supported:    true,
	},
	"lion": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "11",
		Supported:    true,
	},
	"snowleopard": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "10",
		Supported:    true,
	},
	"leopard": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "9",
		Supported:    true,
	},
	"tiger": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "8",
		Supported:    true,
	},
	"panther": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "7",
		Supported:    true,
	},
	"jaguar": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "6",
		Supported:    true,
	},
	"puma": {
		WorkloadType: UnsupportedWorkloadType,
		Version:      "5",
		Supported:    true,
	},
}
