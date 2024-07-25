// Copyright 2024 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state

import (
	"context"

	"github.com/canonical/sqlair"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/domain/application/charm"
	schematesting "github.com/juju/juju/domain/schema/testing"
)

type configSuite struct {
	schematesting.ModelSuite
}

var _ = gc.Suite(&configSuite{})

var configTestCases = [...]struct {
	name   string
	input  []charmConfig
	output charm.Config
}{
	{
		name:  "empty",
		input: []charmConfig{},
		output: charm.Config{
			Options: make(map[string]charm.Option),
		},
	},
	{
		name: "string",
		input: []charmConfig{
			{
				Key:          "string",
				Type:         "string",
				Description:  "description",
				DefaultValue: "default",
			},
		},
		output: charm.Config{
			Options: map[string]charm.Option{
				"string": {
					Type:        charm.OptionString,
					Description: "description",
					Default:     "default",
				},
			},
		},
	},
	{
		name: "secret",
		input: []charmConfig{
			{
				Key:          "secret",
				Type:         "secret",
				Description:  "description",
				DefaultValue: "default",
			},
		},
		output: charm.Config{
			Options: map[string]charm.Option{
				"secret": {
					Type:        charm.OptionSecret,
					Description: "description",
					Default:     "default",
				},
			},
		},
	},
	{
		name: "int",
		input: []charmConfig{
			{
				Key:          "int",
				Type:         "int",
				Description:  "description",
				DefaultValue: "1",
			},
		},
		output: charm.Config{
			Options: map[string]charm.Option{
				"int": {
					Type:        charm.OptionInt,
					Description: "description",
					Default:     1,
				},
			},
		},
	},
	{
		name: "float",
		input: []charmConfig{
			{
				Key:          "float",
				Type:         "float",
				Description:  "description",
				DefaultValue: "4.2",
			},
		},
		output: charm.Config{
			Options: map[string]charm.Option{
				"float": {
					Type:        charm.OptionFloat,
					Description: "description",
					Default:     4.2,
				},
			},
		},
	},
	{
		name: "boolean",
		input: []charmConfig{
			{
				Key:          "boolean",
				Type:         "boolean",
				Description:  "description",
				DefaultValue: "true",
			},
		},
		output: charm.Config{
			Options: map[string]charm.Option{
				"boolean": {
					Type:        charm.OptionBool,
					Description: "description",
					Default:     true,
				},
			},
		},
	},
}

func (s *configSuite) TestDecodeConfig(c *gc.C) {
	for _, tc := range configTestCases {
		c.Logf("Running test case %q", tc.name)

		result, err := decodeConfig(tc.input)
		c.Assert(err, jc.ErrorIsNil)
		c.Check(result, gc.DeepEquals, tc.output)
	}
}

func (s *configSuite) TestDecodeConfigType(c *gc.C) {
	_, err := decodeConfigType("invalid")
	c.Assert(err, gc.ErrorMatches, `unknown config type "invalid"`)
}

func (s *configSuite) TestEncodeConfigType(c *gc.C) {
	_, err := decodeConfigType("invalid")
	c.Assert(err, gc.ErrorMatches, `unknown config type "invalid"`)
}

func (s *configSuite) TestEncodeConfigDefaultValue(c *gc.C) {
	_, err := encodeConfigDefaultValue(int64(0))
	c.Assert(err, gc.ErrorMatches, `unknown config default value type int64`)
}

var configTypeTestCases = [...]struct {
	name   string
	kind   charm.OptionType
	input  string
	output any
}{
	{
		name:   "string",
		kind:   charm.OptionString,
		input:  "deadbeef",
		output: "deadbeef",
	},
	{
		name:   "int",
		kind:   charm.OptionInt,
		input:  "42",
		output: 42,
	},
	{
		name:   "float",
		kind:   charm.OptionFloat,
		input:  "42.3",
		output: 42.3,
	},
	{
		name:   "bool",
		kind:   charm.OptionBool,
		input:  "true",
		output: true,
	},
	{
		name:   "secret",
		kind:   charm.OptionSecret,
		input:  "ssh",
		output: "ssh",
	},
}

func (s *configSuite) TestEncodeThenDecodeDefaultValue(c *gc.C) {
	for _, tc := range configTypeTestCases {
		c.Logf("Running test case %q", tc.name)

		decoded, err := decodeConfigDefaultValue(tc.kind, tc.input)
		c.Assert(err, jc.ErrorIsNil)
		c.Check(decoded, gc.DeepEquals, tc.output)

		encoded, err := encodeConfigDefaultValue(decoded)
		c.Assert(err, jc.ErrorIsNil)
		c.Check(encoded, gc.DeepEquals, tc.input)
	}
}

func (s *configSuite) TestDecodeConfigTypeError(c *gc.C) {
	_, err := decodeConfigDefaultValue(charm.OptionType("invalid"), "")
	c.Assert(err, gc.Not(jc.ErrorIsNil))
}

type configStateSuite struct {
	schematesting.ModelSuite
}

var _ = gc.Suite(&configStateSuite{})

func (s *configStateSuite) TestConfigType(c *gc.C) {
	type charmConfigType struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	stmt := sqlair.MustPrepare(`
SELECT charm_config_type.* AS &charmConfigType.* FROM charm_config_type ORDER BY id;
`, charmConfigType{})

	var results []charmConfigType
	err := s.TxnRunner().Txn(context.Background(), func(ctx context.Context, tx *sqlair.TX) error {
		return tx.Query(ctx, stmt).GetAll(&results)
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, gc.HasLen, 5)

	m := []charm.OptionType{
		charm.OptionString,
		charm.OptionInt,
		charm.OptionFloat,
		charm.OptionBool,
		charm.OptionSecret,
	}

	for i, value := range m {
		c.Logf("result %d: %#v", i, value)
		result, err := encodeConfigType(value)
		c.Assert(err, jc.ErrorIsNil)
		c.Check(result, gc.DeepEquals, results[i].ID)
	}
}