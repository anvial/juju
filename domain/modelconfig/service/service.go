// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package service

import (
	"context"

	"github.com/juju/collections/transform"
	"github.com/juju/schema"

	"github.com/juju/juju/core/changestream"
	coreerrors "github.com/juju/juju/core/errors"
	"github.com/juju/juju/core/trace"
	"github.com/juju/juju/core/watcher"
	"github.com/juju/juju/core/watcher/eventsource"
	"github.com/juju/juju/domain/modelconfig/validators"
	"github.com/juju/juju/domain/modeldefaults"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/environs/config"
	"github.com/juju/juju/internal/errors"
)

// ModelDefaultsProvider is responsible for providing the default config values
// for a model.
type ModelDefaultsProvider interface {
	// ModelDefaults will return the default config values to be used for a model
	// and its config.
	ModelDefaults(context.Context) (modeldefaults.Defaults, error)
}

// ModelConfigProviderFunc describes a type that is able to return a
// [environs.ModelConfigProvider] for the model it is scoped to. The function
// internally determines the cloud type for the model. If no model config
// provider exists for the model's cloud type then a [coreerrors.NotFound]
// error is returned. If the cloud type provider does not support model config
// then a [coreerrors.NotSupported] error is returned.
type ModelConfigProviderFunc func() (environs.ModelConfigProvider, error)

// State represents the state entity for accessing and setting per
// model configuration values.
type State interface {
	ProviderState
	SpaceValidatorState

	// GetModelAgentVersionAndStream returns the current model's set agent
	// version and stream.
	// The following errors can be expected:
	// - [github.com/juju/juju/core/errors.NotFound] if no agent version or
	// stream has been set.
	GetModelAgentVersionAndStream(context.Context) (ver string, stream string, err error)

	// ModelConfigHasAttributes returns the set of attributes that model config
	// currently has set out of the list supplied.
	ModelConfigHasAttributes(context.Context, []string) ([]string, error)

	// SetModelConfig is responsible for setting the current model config and
	// overwriting all previously set values even if the config supplied is
	// empty or nil.
	SetModelConfig(context.Context, map[string]string) error

	// UpdateModelConfig is responsible for both inserting, updating and
	// removing model config values for the current model.
	UpdateModelConfig(context.Context, map[string]string, []string) error

	// NamespaceForWatchModelConfig returns the namespace identifier used for
	// watching model configuration changes.
	NamespaceForWatchModelConfig() string
}

// SpaceValidatorState represents the state entity for validating space-related
// model config.
type SpaceValidatorState interface {
	// SpaceExists checks if the space identified by the given space name exists.
	SpaceExists(ctx context.Context, spaceName string) (bool, error)
}

// WatcherFactory describes methods for creating watchers.
type WatcherFactory interface {
	// NewNamespaceWatcher returns a new watcher that filters changes from the
	// input base watcher's db/queue. Change-log events will be emitted only if
	// the filter accepts them, and dispatching the notifications via the
	// Changes channel. A filter option is required, though additional filter
	// options can be provided.
	NewNamespaceWatcher(
		ctx context.Context,
		initialQuery eventsource.NamespaceQuery,
		summary string,
		filterOption eventsource.FilterOption, filterOptions ...eventsource.FilterOption,
	) (watcher.StringsWatcher, error)
}

// Service defines the service for interacting with ModelConfig.
type Service struct {
	defaultsProvider              ModelDefaultsProvider
	modelValidator                config.Validator
	modelConfigProviderGetterFunc ModelConfigProviderFunc
	st                            State
}

// NewService creates a new ModelConfig service.
func NewService(
	defaultsProvider ModelDefaultsProvider,
	modelValidator config.Validator,
	modelConfigProviderGetterFunc ModelConfigProviderFunc,
	st State,
) *Service {
	return &Service{
		defaultsProvider:              defaultsProvider,
		modelValidator:                modelValidator,
		modelConfigProviderGetterFunc: modelConfigProviderGetterFunc,
		st:                            st,
	}
}

// ModelConfig returns the current config for the model.
func (s *Service) ModelConfig(ctx context.Context) (*config.Config, error) {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	stConfig, err := s.st.ModelConfig(ctx)
	if err != nil {
		return nil, errors.Errorf("getting model config from state: %w", err)
	}

	agentVersion, agentStream, err := s.st.GetModelAgentVersionAndStream(ctx)
	if err != nil {
		return nil, errors.Errorf("getting agent version and stream for model config: %w", err)
	}

	// Coerce provider-specific attributes from string to their proper types.
	altConfig, err := s.deserializeMap(stConfig)
	if err != nil {
		return nil, errors.Errorf("coercing provider config attributes: %w", err)
	}

	// We add the agent version and stream to model config here. Over time we need
	// to remove uses of agent version and stream from model config. We prefer
	// to augment config with this value on read rather then persisting on
	// writing.
	altConfig[config.AgentVersionKey] = agentVersion
	altConfig[config.AgentStreamKey] = agentStream
	return config.New(config.NoDefaults, altConfig)
}

// deserializeMap converts a map[string]string to map[string]any and coerces
// any provider-specific values that are found in the provider's schema.
func (s *Service) deserializeMap(m map[string]string) (map[string]any, error) {
	if s.modelConfigProviderGetterFunc == nil {
		return nil, errors.New("no model config provider getter")
	}

	cloudType, ok := m[config.TypeKey]
	if !ok || cloudType == "" {
		// No cloud type - just convert without coercion.
		return stringMapToAny(m), nil
	}

	provider, err := s.modelConfigProviderGetterFunc()
	if err != nil && !errors.Is(err, coreerrors.NotSupported) {
		return nil, errors.Capture(err)
	} else if provider == nil {
		// Provider not found or doesn't support config schema.
		return nil, errors.New("provider not found or doesn't support config schema")
	}

	fields := provider.ConfigSchema()

	// We are building a set of defaults here for each key that exists in
	// the provider's schema set to [schema.Omit]. The reason for this is
	// that [schema.FieldMap.Coerce] will try and apply defaults for keys
	// that don't exist in the input.
	//
	// We don't want this to happen here. The purpose of this function is to
	// fundamentally coerce the type we store the value in at a state level
	// to that of the schema only if and when the key exists in the input.
	omitDefaults := make(schema.Defaults, len(fields))
	for k := range fields {
		omitDefaults[k] = schema.Omit
	}

	providerFieldMap := schema.FieldMap(fields, omitDefaults)
	coercedCfg, err := providerFieldMap.Coerce(m, nil)
	if err != nil {
		return nil, errors.Errorf("unable to coerce provider config: %w", err)
	}

	providerResult := coercedCfg.(map[string]any)

	// Build final result: coerced provider attrs + uncoerced non-provider attrs
	result := stringMapToAny(m)
	for key, val := range providerResult {
		result[key] = val
	}

	return result, nil
}

func stringMapToAny(m map[string]string) map[string]any {
	return transform.Map(m, func(k, v string) (string, any) { return k, v })
}

// ModelConfigValues returns the config values for the model and the source of
// the value.
func (s *Service) ModelConfigValues(
	ctx context.Context,
) (config.ConfigValues, error) {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	cfg, err := s.ModelConfig(ctx)
	if err != nil {
		return config.ConfigValues{}, err
	}

	defaults, err := s.defaultsProvider.ModelDefaults(ctx)
	if err != nil {
		return config.ConfigValues{}, errors.Errorf("getting model defaults: %w", err)
	}

	allAttrs := cfg.AllAttrs()
	if len(allAttrs) == 0 {
		allAttrs = map[string]any{}
		for k, v := range defaults {
			allAttrs[k] = v.Value()
		}
	}

	result := make(config.ConfigValues, len(allAttrs))
	for attr, val := range allAttrs {
		isDefault, source := defaults[attr].ValueSource(val)
		if !isDefault {
			source = config.JujuModelConfigSource
		}

		result[attr] = config.ConfigValue{
			Value:  val,
			Source: source,
		}
	}

	return result, nil
}

// buildUpdatedModelConfig is responsible for taking the currently set
// ModelConfig and applying in memory update operations.
func (s *Service) buildUpdatedModelConfig(
	ctx context.Context,
	updates map[string]any,
	removeAttrs []string,
) (*config.Config, *config.Config, error) {
	current, err := s.ModelConfig(ctx)
	if err != nil {
		return nil, current, err
	}

	newConf, err := current.Remove(removeAttrs)
	if err != nil {
		return newConf, current, errors.Errorf("building new model config with removed attributes: %w", err)
	}

	newConf, err = newConf.Apply(updates)
	if err != nil {
		return newConf, current, errors.Errorf("building new model config with removed attributes: %w", err)
	}

	return newConf, current, nil
}

// reconcileRemovedAttributes will take a set of attributes to remove from the
// model config and figure out if there exists a model default for the
// attribute. If a model default exists then a set of updates will be provided
// to instead change the attribute to the model default. This function will also
// check that the removed attributes first exist in the model's config otherwise
// we risk bringing in model defaults that were not previously set for the model
// config.
func (s *Service) reconcileRemovedAttributes(
	ctx context.Context,
	removeAttrs []string,
) (map[string]any, error) {
	if len(removeAttrs) == 0 {
		return map[string]any{}, nil
	}

	updates := map[string]any{}
	hasAttrs, err := s.st.ModelConfigHasAttributes(ctx, removeAttrs)
	if err != nil {
		return updates, errors.Errorf("determining model config has attributes: %w", err)
	}

	defaults, err := s.defaultsProvider.ModelDefaults(ctx)
	if err != nil {
		return updates, errors.Errorf("getting model defaults for config attribute removal: %w", err)
	}

	for _, attr := range hasAttrs {
		if val := defaults[attr].Value(); val != nil {
			updates[attr] = val
		}
	}

	return updates, nil
}

// SetModelConfig will remove any existing model config for the model and
// replace with the new config provided. The new config will also be hydrated
// with any model default attributes that have not been set on the config.
func (s *Service) SetModelConfig(
	ctx context.Context,
	cfg map[string]any,
) error {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	defaults, err := s.defaultsProvider.ModelDefaults(ctx)
	if err != nil {
		return errors.Errorf("getting model defaults: %w", err)
	}

	// We want to make a copy of cfg so that we don't modify the users input.
	cfgCopy := make(map[string]any, len(cfg))
	for k, v := range cfg {
		cfgCopy[k] = v
	}

	for k, v := range defaults {
		applyVal := v.ApplyStrategy(cfgCopy[k])
		if applyVal != nil {
			cfgCopy[k] = applyVal
		}
	}

	setCfg, err := config.New(config.NoDefaults, cfgCopy)
	if err != nil {
		return errors.Errorf("constructing new model config with model defaults: %w", err)
	}

	_, err = s.validatorForSetModelConfig().Validate(ctx, setCfg, nil)
	if err != nil {
		return errors.Errorf("validating model config to set for model: %w", err)
	}

	rawCfg, err := CoerceConfigForStorage(setCfg.AllAttrs())
	if err != nil {
		return errors.Errorf("coercing model config for storage: %w", err)
	}

	return s.st.SetModelConfig(ctx, rawCfg)
}

// UpdateModelConfig takes a set of updated and removed attributes to apply.
// Removed attributes are replaced with their model default values should they
// exist. All model config updates are validated against the currently set
// model config. The model config is ran through several validation steps before
// being persisted. If an error occurs during validation then a
// config.ValidationError is returned. The caller can also optionally pass in
// additional config.Validators to be run.
//
// The following validations on model config are run by default:
// - Agent version is not change between updates.
// - Agent stream is not changed between updates.
// - Charmhub url is not changed between updates.
// - The networking space chosen is valid and can be used.
// - The secret backend is valid and can be used.
// - Authorized keys are not changed.
// - Container networking method is not being changed.
func (s *Service) UpdateModelConfig(
	ctx context.Context,
	updateAttrs map[string]any,
	removeAttrs []string,
	additionalValidators ...config.Validator,
) error {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	// noop with no updates or removals to perform.
	if len(updateAttrs) == 0 && len(removeAttrs) == 0 {
		return nil
	}

	updates, err := s.reconcileRemovedAttributes(ctx, removeAttrs)
	if err != nil {
		return errors.Capture(err)
	}

	// It's important here that we apply the user updates over the top of the
	// calculated ones. This way we always take the user's supplied key value
	// over defaults.
	for k, v := range updateAttrs {
		updates[k] = v
	}

	newCfg, currCfg, err := s.buildUpdatedModelConfig(ctx, updates, removeAttrs)
	if err != nil {
		return errors.Errorf("making updated model configuration: %w", err)
	}

	validatedCfg, err := s.validatorForUpdateModelConfig().Validate(ctx, newCfg, currCfg)
	if err != nil {
		return errors.Errorf("validating updated model configuration: %w", err)
	}

	// We need to walk through all of the updates and potentially find any
	// changes that were made by the validators.
	validatedUpdates := make(map[string]any, len(updates))
	validatedCfgAttrs := validatedCfg.AllAttrs()
	for k := range updates {
		validatedCfgVal, exists := validatedCfgAttrs[k]
		if !exists {
			continue
		}
		validatedUpdates[k] = validatedCfgVal
	}

	rawCfgUpdate, err := CoerceConfigForStorage(validatedUpdates)
	if err != nil {
		return errors.Errorf("coercing new configuration for persistence: %w", err)
	}

	err = s.st.UpdateModelConfig(ctx, rawCfgUpdate, removeAttrs)
	if err != nil {
		return errors.Errorf("updating model config: %w", err)
	}
	return nil
}

// spaceValidator implements validators.SpaceProvider.
type spaceValidator struct {
	st SpaceValidatorState
}

// HasSpace implements validators.SpaceProvider. It checks whether the
// given space exists.
func (v *spaceValidator) HasSpace(ctx context.Context, spaceName string) (bool, error) {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	return v.st.SpaceExists(ctx, spaceName)
}

// validatorForSetModelConfig returns a config validator to use on model config
// when it is being set initially. The validator returned will check that:
// - Network space exists.
// - Secret backend exists.
// - Container networking method is valid.
func (s *Service) validatorForSetModelConfig(
	additional ...config.Validator,
) config.Validator {
	agg := &config.AggregateValidator{
		Validators: []config.Validator{
			validators.ContainerNetworkingMethodValue(),
			s.modelValidator,
		},
	}
	agg.Validators = append(agg.Validators, additional...)
	return agg
}

// validatorForUpdateModelConfig returns a config validator to use on model config
// when it is being updated. The validator returned will check that:
// - Agent version is not being changed.
// - CharmhubURL is not being changed.
// - Network space exists.
// - Container networking method is not being changed.
func (s *Service) validatorForUpdateModelConfig(
	additional ...config.Validator,
) config.Validator {
	agg := &config.AggregateValidator{
		Validators: []config.Validator{
			validators.AgentStreamChange(),
			validators.AgentVersionChange(),
			validators.CharmhubURLChange(),
			validators.SpaceChecker(&spaceValidator{
				st: s.st,
			}),
			validators.ContainerNetworkingMethodChange(),
			s.modelValidator,
		},
	}
	agg.Validators = append(agg.Validators, additional...)
	return agg
}

// WatchableService defines the service for interacting with ModelConfig
// and the ability to create watchers.
type WatchableService struct {
	Service
	watcherFactory WatcherFactory
}

// NewWatchableService creates a new WatchableService for interacting with
// ModelConfig and the ability to create watchers.
func NewWatchableService(
	defaultsProvider ModelDefaultsProvider,
	modelValidator config.Validator,
	modelConfigProviderGetterFunc ModelConfigProviderFunc,
	st State,
	watcherFactory WatcherFactory,
) *WatchableService {
	return &WatchableService{
		Service: Service{
			defaultsProvider:              defaultsProvider,
			modelValidator:                modelValidator,
			modelConfigProviderGetterFunc: modelConfigProviderGetterFunc,
			st:                            st,
		},
		watcherFactory: watcherFactory,
	}
}

// ProviderModelConfigGetter returns a ModelConfigProviderFunc that can be used
// to get a ModelConfigProvider for the model. The function internally
// determines the cloud type from the model config and caches the provider for
// the lifetime of the function.
func ProviderModelConfigGetter(ctx context.Context, st State) ModelConfigProviderFunc {
	var (
		cachedProvider environs.ModelConfigProvider
		cached         bool
	)

	return func() (environs.ModelConfigProvider, error) {
		// Return cached provider if available.
		if cached {
			return cachedProvider, nil
		}

		// Get the model config to determine cloud type.
		cfg, err := st.ModelConfig(ctx)
		if err != nil {
			return nil, errors.Errorf("getting model config for provider lookup: %w", err)
		}

		cloudType, ok := cfg[config.TypeKey]
		if !ok || cloudType == "" {
			return nil, errors.Errorf("cloud type not found in model config").Add(coreerrors.NotFound)
		}

		envProvider, err := environs.GlobalProviderRegistry().Provider(cloudType)
		if errors.Is(err, coreerrors.NotFound) {
			return nil, errors.Errorf(
				"no model config provider exists for cloud type %q", cloudType,
			).Add(coreerrors.NotFound)
		}
		if err != nil {
			return nil, errors.Capture(err)
		}

		modelConfigProvider, supports := envProvider.(environs.ModelConfigProvider)
		if !supports {
			return nil, errors.Errorf(
				"model config provider not supported for cloud type %q", cloudType,
			).Add(coreerrors.NotSupported)
		}

		// Cache the provider for subsequent calls.
		cachedProvider = modelConfigProvider
		cached = true

		return modelConfigProvider, nil
	}
}

// Watch returns a watcher that returns keys for any changes to model
// config.
func (s *WatchableService) Watch(ctx context.Context) (watcher.StringsWatcher, error) {
	ctx, span := trace.Start(ctx, trace.NameFromFunc())
	defer span.End()

	return s.watcherFactory.NewNamespaceWatcher(
		ctx,
		eventsource.InitialNamespaceChanges(s.st.AllKeysQuery()),
		"model config watcher",
		eventsource.NamespaceFilter(s.st.NamespaceForWatchModelConfig(), changestream.All),
	)
}
