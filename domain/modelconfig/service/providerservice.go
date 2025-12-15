// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package service

import (
	"context"

	"github.com/juju/collections/transform"

	"github.com/juju/juju/core/changestream"
	"github.com/juju/juju/core/trace"
	"github.com/juju/juju/core/watcher"
	"github.com/juju/juju/core/watcher/eventsource"
	"github.com/juju/juju/environs/config"
	"github.com/juju/juju/internal/errors"
)

// ProviderState defines the state methods required by the ProviderService.
type ProviderState interface {
	// AllKeysQuery returns a SQL statement that will return all known model config
	// keys.
	AllKeysQuery() string
	// ModelConfig returns the currently set config for the model.
	ModelConfig(context.Context) (map[string]string, error)
	// NamespaceForWatchModelConfig returns the namespace identifier used for
	// watching model configuration changes.
	NamespaceForWatchModelConfig() string
}

// ProviderService defines the service for interacting with ModelConfig.
// The provider service is a subset of the ModelConfig service, and is used by
// the provider package to interact with the ModelConfig service. By not
// exposing the full ModelConfig service, the provider package is not able to
// modify the ModelConfig entities, only read them.
//
// Provider-specific config attributes are stored as strings in the database
// (map[string]string). When reading from the database, if a providerSchema is
// provided, the service will coerce provider-specific attributes from strings
// to their proper types (bool, int, etc.) according to the provider's schema.
// This ensures that provider code can safely type-assert these values without
// panicking.
type ProviderService struct {
	st                            ProviderState
	modelConfigProviderGetterFunc ModelConfigProviderFunc
}

// NewProviderService creates a new ModelConfig service.
func NewProviderService(
	st ProviderState,
	modelConfigProviderGetterFunc ModelConfigProviderFunc,
) *ProviderService {
	return &ProviderService{
		st:                            st,
		modelConfigProviderGetterFunc: modelConfigProviderGetterFunc,
	}
}

// ModelConfig returns the current config for the model.
func (s *ProviderService) ModelConfig(ctx context.Context) (*config.Config, error) {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	stConfig, err := s.st.ModelConfig(ctx)
	if err != nil {
		return nil, errors.Errorf("getting model config from state: %w", err)
	}

	// Coerce provider-specific attributes from string to their proper types.
	coerced, err := s.deserializeMap(stConfig)
	if err != nil {
		return nil, errors.Errorf("coercing provider config attributes: %w", err)
	}

	return config.New(config.NoDefaults, coerced)
}

// deserializeMap converts a map[string]string from the database to map[string]any
// and coerces any provider-specific values that are found in the provider's schema.
// This is necessary because the database stores all config as strings, but provider
// code expects typed values (e.g., bool, int) for provider-specific attributes.
func (s *ProviderService) deserializeMap(m map[string]string) (map[string]any, error) {
	result := make(map[string]any, len(m))

	// If we don't have a model config provider getter, just do basic string->any conversion
	if s.modelConfigProviderGetterFunc == nil {
		return transform.Map(m, func(k, v string) (string, any) { return k, v }), nil
	}

	// Get the cloud type from the config
	cloudType, ok := m[config.TypeKey]
	if !ok || cloudType == "" {
		// No cloud type - just convert without coercion
		return transform.Map(m, func(k, v string) (string, any) { return k, v }), nil
	}

	// Get the provider for this cloud type
	provider, err := s.modelConfigProviderGetterFunc(cloudType)
	if err != nil {
		// Provider not found or doesn't support schema - graceful degradation
		return transform.Map(m, func(k, v string) (string, any) { return k, v }), nil
	}

	if provider == nil {
		// No provider available - just convert without coercion
		return transform.Map(m, func(k, v string) (string, any) { return k, v }), nil
	}

	// Get the schema from the provider
	fields := provider.ConfigSchema()
	if fields == nil {
		// No schema available - just convert without coercion
		return transform.Map(m, func(k, v string) (string, any) { return k, v }), nil
	}

	for key, strVal := range m {
		if field, ok := fields[key]; ok {
			// This is a provider-specific attribute - coerce it to proper type
			coercedVal, err := field.Coerce(strVal, []string{key})
			if err != nil {
				return nil, errors.Errorf("unable to coerce provider config key %q: %w", key, err)
			}
			result[key] = coercedVal
		} else {
			// Not a provider-specific attribute - keep as string
			result[key] = strVal
		}
	}

	return result, nil
}

// WatchableProviderService defines the service for interacting with ModelConfig
// and the ability to create watchers.
type WatchableProviderService struct {
	ProviderService
	watcherFactory WatcherFactory
}

// NewWatchableProviderService creates a new WatchableProviderService for
// interacting with ModelConfig and the ability to create watchers.
func NewWatchableProviderService(
	st ProviderState,
	modelConfigProviderGetterFunc ModelConfigProviderFunc,
	watcherFactory WatcherFactory,
) *WatchableProviderService {
	return &WatchableProviderService{
		ProviderService: ProviderService{
			st:                            st,
			modelConfigProviderGetterFunc: modelConfigProviderGetterFunc,
		},
		watcherFactory: watcherFactory,
	}
}

// Watch returns a watcher that returns keys for any changes to model
// config.
func (s *WatchableProviderService) Watch(ctx context.Context) (watcher.StringsWatcher, error) {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	return s.watcherFactory.NewNamespaceWatcher(
		ctx,
		eventsource.InitialNamespaceChanges(s.st.AllKeysQuery()),
		"provider model config watcher",
		eventsource.NamespaceFilter(s.st.NamespaceForWatchModelConfig(), changestream.All),
	)
}
