// Copyright 2024 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package service

import (
	"testing"

	"github.com/juju/schema"
	"github.com/juju/tc"
	gomock "go.uber.org/mock/gomock"

	coreerrors "github.com/juju/juju/core/errors"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/internal/errors"
)

type providerServiceSuite struct {
	mockState               *MockProviderState
	mockModelConfigProvider *MockModelConfigProvider
}

func TestProviderServiceSuite(t *testing.T) {
	tc.Run(t, &providerServiceSuite{})
}

func (s *providerServiceSuite) setupMocks(c *tc.C) *gomock.Controller {
	ctrl := gomock.NewController(c)
	s.mockState = NewMockProviderState(ctrl)
	s.mockModelConfigProvider = NewMockModelConfigProvider(ctrl)
	return ctrl
}

func (s *providerServiceSuite) modelConfigProviderFunc(cloudType string) ModelConfigProviderFunc {
	return func(ct string) (environs.ModelConfigProvider, error) {
		if ct != cloudType {
			return nil, errors.Errorf("unknown cloud type %q", ct).Add(coreerrors.NotFound)
		}
		return s.mockModelConfigProvider, nil
	}
}

func (s *providerServiceSuite) TestModelConfig(c *tc.C) {
	defer s.setupMocks(c).Finish()

	s.mockState.EXPECT().ModelConfig(gomock.Any()).Return(
		map[string]string{
			"name": "wallyworld",
			"uuid": "a677bdfd-3c96-46b2-912f-38e25faceaf7",
			"type": "sometype",
		},
		nil,
	)

	svc := NewProviderService(s.mockState, nil)
	cfg, err := svc.ModelConfig(c.Context())
	c.Check(err, tc.ErrorIsNil)
	c.Check(cfg.AllAttrs(), tc.DeepEquals, map[string]any{
		"name":           "wallyworld",
		"uuid":           "a677bdfd-3c96-46b2-912f-38e25faceaf7",
		"type":           "sometype",
		"logging-config": "<root>=INFO",
	})
}

// TestModelConfigWithProviderSchemaCoercion checks that provider-specific
// config attributes are coerced from strings to their proper types based on
// the provider's schema.
func (s *providerServiceSuite) TestModelConfigWithProviderSchemaCoercion(c *tc.C) {
	defer s.setupMocks(c).Finish()

	s.mockState.EXPECT().ModelConfig(gomock.Any()).Return(
		map[string]string{
			"name":           "wallyworld",
			"uuid":           "a677bdfd-3c96-46b2-912f-38e25faceaf7",
			"type":           "testprovider",
			"provider-bool":  "true",
			"provider-int":   "42",
			"regular-string": "value",
		},
		nil,
	)

	s.mockModelConfigProvider.EXPECT().ConfigSchema().Return(
		schema.Fields{
			"provider-bool": schema.Bool(),
			"provider-int":  schema.Int(),
		},
	)

	providerGetter := s.modelConfigProviderFunc("testprovider")

	svc := NewProviderService(s.mockState, providerGetter)
	cfg, err := svc.ModelConfig(c.Context())
	c.Check(err, tc.ErrorIsNil)

	attrs := cfg.AllAttrs()
	c.Check(attrs["name"], tc.Equals, "wallyworld")
	c.Check(attrs["uuid"], tc.Equals, "a677bdfd-3c96-46b2-912f-38e25faceaf7")
	c.Check(attrs["type"], tc.Equals, "testprovider")
	c.Check(attrs["provider-bool"], tc.Equals, true)
	c.Check(attrs["provider-int"], tc.Equals, int64(42))
	c.Check(attrs["regular-string"], tc.Equals, "value")
}

// TestModelConfigWithoutProviderGetter checks that ModelConfig works correctly
// when no provider getter is supplied (graceful degradation).
func (s *providerServiceSuite) TestModelConfigWithoutProviderGetter(c *tc.C) {
	defer s.setupMocks(c).Finish()

	s.mockState.EXPECT().ModelConfig(gomock.Any()).Return(
		map[string]string{
			"name": "wallyworld",
			"uuid": "a677bdfd-3c96-46b2-912f-38e25faceaf7",
			"type": "sometype",
		},
		nil,
	)

	svc := NewProviderService(s.mockState, nil)
	cfg, err := svc.ModelConfig(c.Context())
	c.Check(err, tc.ErrorIsNil)
	c.Check(cfg.Name(), tc.Equals, "wallyworld")
}

// TestModelConfigWithProviderNotFound checks that ModelConfig gracefully
// handles the case where the provider is not found.
func (s *providerServiceSuite) TestModelConfigWithProviderNotFound(c *tc.C) {
	defer s.setupMocks(c).Finish()

	s.mockState.EXPECT().ModelConfig(gomock.Any()).Return(
		map[string]string{
			"name": "wallyworld",
			"uuid": "a677bdfd-3c96-46b2-912f-38e25faceaf7",
			"type": "unknown",
		},
		nil,
	)

	providerGetter := s.modelConfigProviderFunc("testprovider")

	svc := NewProviderService(s.mockState, providerGetter)
	cfg, err := svc.ModelConfig(c.Context())
	c.Check(err, tc.ErrorIsNil)
	c.Check(cfg.Name(), tc.Equals, "wallyworld")
}

// TestModelConfigWithProviderNoSchema checks that ModelConfig gracefully
// handles the case where the provider has no schema.
func (s *providerServiceSuite) TestModelConfigWithProviderNoSchema(c *tc.C) {
	defer s.setupMocks(c).Finish()

	s.mockState.EXPECT().ModelConfig(gomock.Any()).Return(
		map[string]string{
			"name": "wallyworld",
			"uuid": "a677bdfd-3c96-46b2-912f-38e25faceaf7",
			"type": "testprovider",
		},
		nil,
	)

	s.mockModelConfigProvider.EXPECT().ConfigSchema().Return(nil)

	providerGetter := s.modelConfigProviderFunc("testprovider")

	svc := NewProviderService(s.mockState, providerGetter)
	cfg, err := svc.ModelConfig(c.Context())
	c.Check(err, tc.ErrorIsNil)
	c.Check(cfg.Name(), tc.Equals, "wallyworld")
}

// TestModelConfigStateError checks that errors from the state layer are
// properly propagated.
func (s *providerServiceSuite) TestModelConfigStateError(c *tc.C) {
	defer s.setupMocks(c).Finish()

	s.mockState.EXPECT().ModelConfig(gomock.Any()).Return(
		nil,
		errors.Errorf("database error"),
	)

	svc := NewProviderService(s.mockState, nil)
	_, err := svc.ModelConfig(c.Context())
	c.Check(err, tc.ErrorMatches, "getting model config from state:.*database error")
}
