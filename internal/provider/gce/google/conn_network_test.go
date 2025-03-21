// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package google_test

import (
	"regexp"
	"sort"

	"github.com/juju/collections/set"
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	"google.golang.org/api/compute/v1"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/core/network"
	corefirewall "github.com/juju/juju/core/network/firewall"
	"github.com/juju/juju/internal/provider/gce/google"
)

func (s *connSuite) TestConnectionIngressRules(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:         "spam",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"10.0.0.0/24", "192.168.1.0/24"},
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      []string{"80-81", "92"},
			}, {
				IPProtocol: "udp",
				Ports:      []string{"443", "100-120"},
			},
		},
	}}

	ports, err := s.Conn.IngressRules("spam")
	c.Assert(err, jc.ErrorIsNil)
	c.Check(
		ports, jc.DeepEquals,
		corefirewall.IngressRules{
			corefirewall.NewIngressRule(network.MustParsePortRange("80-81/tcp"), "10.0.0.0/24", "192.168.1.0/24"),
			corefirewall.NewIngressRule(network.MustParsePortRange("92/tcp"), "10.0.0.0/24", "192.168.1.0/24"),
			corefirewall.NewIngressRule(network.MustParsePortRange("100-120/udp"), "10.0.0.0/24", "192.168.1.0/24"),
			corefirewall.NewIngressRule(network.MustParsePortRange("443/udp"), "10.0.0.0/24", "192.168.1.0/24"),
		},
	)
}

func (s *connSuite) TestConnectionIngressRulesCollapse(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:         "spam",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"10.0.0.0/24", "192.168.1.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"81"},
		}, {
			IPProtocol: "tcp",
			Ports:      []string{"82"},
		}, {
			IPProtocol: "tcp",
			Ports:      []string{"80"},
		}, {
			IPProtocol: "tcp",
			Ports:      []string{"83"},
		}, {
			IPProtocol: "tcp",
			Ports:      []string{"92"},
		}},
	}}

	ports, err := s.Conn.IngressRules("spam")
	c.Assert(err, jc.ErrorIsNil)
	c.Check(
		ports, jc.DeepEquals,
		corefirewall.IngressRules{
			corefirewall.NewIngressRule(network.MustParsePortRange("80-83/tcp"), "10.0.0.0/24", "192.168.1.0/24"),
			corefirewall.NewIngressRule(network.MustParsePortRange("92/tcp"), "10.0.0.0/24", "192.168.1.0/24"),
		},
	)
}

func (s *connSuite) TestConnectionIngressRulesDefaultCIDR(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:       "spam",
		TargetTags: []string{"spam"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-81", "92"},
		}},
	}}

	ports, err := s.Conn.IngressRules("spam")
	c.Assert(err, jc.ErrorIsNil)
	c.Check(
		ports, jc.DeepEquals,
		corefirewall.IngressRules{
			corefirewall.NewIngressRule(network.MustParsePortRange("80-81/tcp"), corefirewall.AllNetworksIPV4CIDR),
			corefirewall.NewIngressRule(network.MustParsePortRange("92/tcp"), corefirewall.AllNetworksIPV4CIDR),
		},
	)
}

func (s *connSuite) TestConnectionPortsAPI(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:         "spam",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"0.0.0.0/0"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-81"},
		}},
	}}

	_, err := s.Conn.IngressRules("eggs")
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.FakeConn.Calls, gc.HasLen, 1)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "GetFirewalls")
	c.Check(s.FakeConn.Calls[0].ProjectID, gc.Equals, "spam")
	c.Check(s.FakeConn.Calls[0].Name, gc.Equals, "eggs")
}

func (s *connSuite) TestConnectionOpenPortsAdd(c *gc.C) {
	s.FakeConn.Err = errors.NotFoundf("spam")

	rules := corefirewall.IngressRules{
		corefirewall.NewIngressRule(network.MustParsePortRange("80-81/tcp")), // leave out CIDR to check default
		corefirewall.NewIngressRule(network.MustParsePortRange("80-81/udp"), corefirewall.AllNetworksIPV4CIDR),
		corefirewall.NewIngressRule(network.MustParsePortRange("100-120/tcp"), "192.168.1.0/24", "10.0.0.0/24"),
		corefirewall.NewIngressRule(network.MustParsePortRange("67/udp"), "10.0.0.0/24"),
	}
	err := s.Conn.OpenPortsWithNamer("spam", google.HashSuffixNamer, rules)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.FakeConn.Calls, gc.HasLen, 4)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "GetFirewalls")
	c.Check(s.FakeConn.Calls[1].FuncName, gc.Equals, "AddFirewall")
	c.Check(s.FakeConn.Calls[1].Firewall, jc.DeepEquals, &compute.Firewall{
		Name:         "spam-4eebe8d7a9",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"10.0.0.0/24", "192.168.1.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"100-120"},
		}},
	})
	c.Check(s.FakeConn.Calls[2].FuncName, gc.Equals, "AddFirewall")
	c.Check(s.FakeConn.Calls[2].Firewall, jc.DeepEquals, &compute.Firewall{
		Name:         "spam-a34d80f7b6",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"10.0.0.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "udp",
			Ports:      []string{"67"},
		}},
	})
	c.Check(s.FakeConn.Calls[3].FuncName, gc.Equals, "AddFirewall")
	c.Check(s.FakeConn.Calls[3].Firewall, jc.DeepEquals, &compute.Firewall{
		Name:         "spam",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"0.0.0.0/0"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-81"},
		}, {
			IPProtocol: "udp",
			Ports:      []string{"80-81"},
		}},
	})
}

func (s *connSuite) TestConnectionOpenPortsUpdateSameCIDR(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:         "spam-ad7554",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"192.168.1.0/24", "10.0.0.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-81"},
		}},
	}}

	rules := corefirewall.IngressRules{
		corefirewall.NewIngressRule(network.MustParsePortRange("443/tcp"), "192.168.1.0/24", "10.0.0.0/24"),
	}
	err := s.Conn.OpenPortsWithNamer("spam", google.HashSuffixNamer, rules)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.FakeConn.Calls, gc.HasLen, 2)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "GetFirewalls")
	c.Check(s.FakeConn.Calls[1].FuncName, gc.Equals, "UpdateFirewall")
	sort.Strings(s.FakeConn.Calls[1].Firewall.Allowed[0].Ports)
	c.Check(s.FakeConn.Calls[1].Firewall, jc.DeepEquals, &compute.Firewall{
		Name:         "spam-ad7554",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"10.0.0.0/24", "192.168.1.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"443", "80-81"},
		}},
	})
}

func (s *connSuite) TestConnectionOpenPortsUpdateAddCIDR(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:         "spam-arbitrary-name",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"192.168.1.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-81"},
		}},
	}}

	rules := corefirewall.IngressRules{
		corefirewall.NewIngressRule(network.MustParsePortRange("80-81/tcp"), "10.0.0.0/24"),
	}
	err := s.Conn.OpenPortsWithNamer("spam", google.HashSuffixNamer, rules)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.FakeConn.Calls, gc.HasLen, 2)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "GetFirewalls")
	c.Check(s.FakeConn.Calls[1].FuncName, gc.Equals, "UpdateFirewall")
	sort.Strings(s.FakeConn.Calls[1].Firewall.Allowed[0].Ports)
	c.Check(s.FakeConn.Calls[1].Name, gc.Equals, "spam-arbitrary-name")
	c.Check(s.FakeConn.Calls[1].Firewall, jc.DeepEquals, &compute.Firewall{
		Name:         "spam-arbitrary-name",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"10.0.0.0/24", "192.168.1.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-81"},
		}},
	})
}

func (s *connSuite) TestConnectionOpenPortsUpdateAndAdd(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:         "spam-d01a82",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"192.168.1.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-81"},
		}},
	}, {
		Name:         "spam-8e65efabcd",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"172.0.0.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"100-120", "443"},
		}},
	}}

	rules := corefirewall.IngressRules{
		corefirewall.NewIngressRule(network.MustParsePortRange("443/tcp"), "192.168.1.0/24"),
		corefirewall.NewIngressRule(network.MustParsePortRange("80-100/tcp"), "10.0.0.0/24"),
		corefirewall.NewIngressRule(network.MustParsePortRange("443/tcp"), "10.0.0.0/24"),
		corefirewall.NewIngressRule(network.MustParsePortRange("67/udp"), "172.0.0.0/24"),
	}
	err := s.Conn.OpenPortsWithNamer("spam", google.HashSuffixNamer, rules)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.FakeConn.Calls, gc.HasLen, 4)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "GetFirewalls")
	c.Check(s.FakeConn.Calls[1].FuncName, gc.Equals, "UpdateFirewall")
	sort.Strings(s.FakeConn.Calls[1].Firewall.Allowed[0].Ports)
	c.Check(s.FakeConn.Calls[1].Name, gc.Equals, "spam-8e65efabcd")
	c.Check(s.FakeConn.Calls[1].Firewall, jc.DeepEquals, &compute.Firewall{
		Name:         "spam-8e65efabcd",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"172.0.0.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"100-120", "443"},
		}, {
			IPProtocol: "udp",
			Ports:      []string{"67"},
		}},
	})
	c.Check(s.FakeConn.Calls[2].FuncName, gc.Equals, "AddFirewall")
	sort.Strings(s.FakeConn.Calls[2].Firewall.Allowed[0].Ports)
	c.Check(s.FakeConn.Calls[2].Firewall, jc.DeepEquals, &compute.Firewall{
		Name:         "spam-a34d80f7b6",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"10.0.0.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"443", "80-100"},
		}},
	})
	c.Check(s.FakeConn.Calls[3].FuncName, gc.Equals, "UpdateFirewall")
	sort.Strings(s.FakeConn.Calls[3].Firewall.Allowed[0].Ports)
	c.Check(s.FakeConn.Calls[3].Name, gc.Equals, "spam-d01a82")
	c.Check(s.FakeConn.Calls[3].Firewall, jc.DeepEquals, &compute.Firewall{
		Name:         "spam-d01a82",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"192.168.1.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"443", "80-81"},
		}},
	})
}

func (s *connSuite) TestConnectionClosePortsRemove(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:         "spam",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"0.0.0.0/0"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"443"},
		}},
	}}

	rules := corefirewall.IngressRules{
		corefirewall.NewIngressRule(network.MustParsePortRange("443/tcp")),
	}
	err := s.Conn.ClosePorts("spam", rules)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.FakeConn.Calls, gc.HasLen, 2)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "GetFirewalls")
	c.Check(s.FakeConn.Calls[1].FuncName, gc.Equals, "RemoveFirewall")
	c.Check(s.FakeConn.Calls[1].Name, gc.Equals, "spam")
}

func (s *connSuite) TestConnectionClosePortsUpdate(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:         "spam",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"0.0.0.0/0"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-81", "443"},
		}},
	}}

	rules := corefirewall.IngressRules{
		corefirewall.NewIngressRule(network.MustParsePortRange("443/tcp")),
	}
	err := s.Conn.ClosePorts("spam", rules)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.FakeConn.Calls, gc.HasLen, 2)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "GetFirewalls")
	c.Check(s.FakeConn.Calls[1].FuncName, gc.Equals, "UpdateFirewall")
	sort.Strings(s.FakeConn.Calls[1].Firewall.Allowed[0].Ports)
	c.Check(s.FakeConn.Calls[1].Firewall, jc.DeepEquals, &compute.Firewall{
		Name:         "spam",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"0.0.0.0/0"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-81"},
		}},
	})
}

func (s *connSuite) TestConnectionClosePortsCollapseUpdate(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:         "spam",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"0.0.0.0/0"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-80", "100-120", "81-81", "82-82"},
		}},
	}}

	rules := corefirewall.IngressRules{
		corefirewall.NewIngressRule(network.MustParsePortRange("80-82/tcp")),
	}
	err := s.Conn.ClosePorts("spam", rules)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.FakeConn.Calls, gc.HasLen, 2)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "GetFirewalls")
	c.Check(s.FakeConn.Calls[1].FuncName, gc.Equals, "UpdateFirewall")
	sort.Strings(s.FakeConn.Calls[1].Firewall.Allowed[0].Ports)
	c.Check(s.FakeConn.Calls[1].Firewall, jc.DeepEquals, &compute.Firewall{
		Name:         "spam",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"0.0.0.0/0"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"100-120"},
		}},
	})
}

func (s *connSuite) TestConnectionClosePortsRemoveCIDR(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:         "glass-onion",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"192.168.1.0/24", "10.0.0.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-81", "443"},
		}},
	}}

	rules := corefirewall.IngressRules{
		corefirewall.NewIngressRule(network.MustParsePortRange("443/tcp"), "192.168.1.0/24"),
		corefirewall.NewIngressRule(network.MustParsePortRange("80-81/tcp"), "192.168.1.0/24"),
	}
	err := s.Conn.ClosePorts("spam", rules)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.FakeConn.Calls, gc.HasLen, 2)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "GetFirewalls")
	c.Check(s.FakeConn.Calls[1].FuncName, gc.Equals, "UpdateFirewall")
	sort.Strings(s.FakeConn.Calls[1].Firewall.Allowed[0].Ports)
	c.Check(s.FakeConn.Calls[1].Firewall, jc.DeepEquals, &compute.Firewall{
		Name:         "glass-onion",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"10.0.0.0/24"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"443", "80-81"},
		}},
	})
}

func (s *connSuite) TestRemoveFirewall(c *gc.C) {
	err := s.Conn.RemoveFirewall("glass-onion")
	c.Assert(err, jc.ErrorIsNil)

	c.Check(s.FakeConn.Calls, gc.HasLen, 1)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "RemoveFirewall")
	c.Check(s.FakeConn.Calls[0].ProjectID, gc.Equals, "spam")
	c.Check(s.FakeConn.Calls[0].Name, gc.Equals, "glass-onion")
}

func (s *connSuite) TestConnectionCloseMoMatches(c *gc.C) {
	s.FakeConn.Firewalls = []*compute.Firewall{{
		Name:         "spam",
		TargetTags:   []string{"spam"},
		SourceRanges: []string{"0.0.0.0/0"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"80-81", "443"},
		}},
	}}

	rules := corefirewall.IngressRules{
		corefirewall.NewIngressRule(network.MustParsePortRange("100-110/tcp"), "192.168.0.1/24"),
	}
	err := s.Conn.ClosePorts("spam", rules)
	c.Assert(err, gc.ErrorMatches, regexp.QuoteMeta(`closing port(s) [100-110/tcp from 192.168.0.1/24] over non-matching rules not supported`))

	c.Check(s.FakeConn.Calls, gc.HasLen, 1)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "GetFirewalls")
}

func (s *connSuite) TestNetworks(c *gc.C) {
	s.FakeConn.Networks = []*compute.Network{{
		Name: "kamar-taj",
	}}
	results, err := s.Conn.Networks()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, gc.HasLen, 1)
	c.Assert((*results[0]).Name, gc.Equals, "kamar-taj")

	c.Check(s.FakeConn.Calls, gc.HasLen, 1)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "ListNetworks")
	c.Check(s.FakeConn.Calls[0].ProjectID, gc.Equals, "spam")
}

func (s *connSuite) TestSubnetworks(c *gc.C) {
	s.FakeConn.Subnetworks = []*compute.Subnetwork{{
		Name: "heptapod",
	}}
	results, err := s.Conn.Subnetworks("us-central1")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, gc.HasLen, 1)
	c.Assert((*results[0]).Name, gc.Equals, "heptapod")

	c.Check(s.FakeConn.Calls, gc.HasLen, 1)
	c.Check(s.FakeConn.Calls[0].FuncName, gc.Equals, "ListSubnetworks")
	c.Check(s.FakeConn.Calls[0].ProjectID, gc.Equals, "spam")
	c.Check(s.FakeConn.Calls[0].Region, gc.Equals, "us-central1")
}

func (s *connSuite) TestRandomSuffixNamer(c *gc.C) {
	ruleset := google.NewRuleSetFromRules(corefirewall.IngressRules{
		corefirewall.NewIngressRule(network.MustParsePortRange("80-90/tcp")),
		corefirewall.NewIngressRule(network.MustParsePortRange("80-90/tcp"), "10.0.10.0/24"),
	})
	i := 0
	for _, firewall := range ruleset {
		i++
		c.Logf("%#v", *firewall)
		name, err := google.RandomSuffixNamer(firewall, "mischief", set.NewStrings())
		c.Assert(err, jc.ErrorIsNil)
		if firewall.SourceCIDRs[0] == "0.0.0.0/0" {
			c.Assert(name, gc.Equals, "mischief")
		} else {
			c.Assert(name, gc.Matches, "mischief-[0-9a-f]{8}")
		}
	}
	c.Assert(i, gc.Equals, 2)
}
