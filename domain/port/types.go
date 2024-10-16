// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package port

import (
	"sort"

	"github.com/juju/juju/core/network"
)

// UnitPortRange represents a range of ports for a given protocol for a
// given unit.
type UnitEndpointPortRange struct {
	UnitName  string
	Endpoint  string
	PortRange network.PortRange
}

func (u UnitEndpointPortRange) LessThan(other UnitEndpointPortRange) bool {
	if u.UnitName != other.UnitName {
		return u.UnitName < other.UnitName
	}
	if u.Endpoint != other.Endpoint {
		return u.Endpoint < other.Endpoint
	}
	return u.PortRange.LessThan(other.PortRange)
}

func SortUnitEndpointPortRanges(portRanges UnitEndpointPortRanges) {
	sort.Slice(portRanges, func(i, j int) bool {
		return portRanges[i].LessThan(portRanges[j])
	})
}

type UnitEndpointPortRanges []UnitEndpointPortRange

func (prs UnitEndpointPortRanges) ByUnitByEndpoint() map[string]network.GroupedPortRanges {
	byUnitByEndpoint := make(map[string]network.GroupedPortRanges)
	for _, unitEnpointPortRange := range prs {
		unitUUID := unitEnpointPortRange.UnitName
		endpoint := unitEnpointPortRange.Endpoint
		if _, ok := byUnitByEndpoint[unitUUID]; !ok {
			byUnitByEndpoint[unitUUID] = network.GroupedPortRanges{}
		}
		byUnitByEndpoint[unitUUID][endpoint] = append(byUnitByEndpoint[unitUUID][endpoint], unitEnpointPortRange.PortRange)
	}
	return byUnitByEndpoint
}
