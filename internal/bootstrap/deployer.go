// Copyright 2023 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package bootstrap

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/juju/clock"

	"github.com/juju/juju/api"
	"github.com/juju/juju/controller"
	corearch "github.com/juju/juju/core/arch"
	corebase "github.com/juju/juju/core/base"
	corecharm "github.com/juju/juju/core/charm"
	"github.com/juju/juju/core/config"
	"github.com/juju/juju/core/constraints"
	coreerrors "github.com/juju/juju/core/errors"
	"github.com/juju/juju/core/logger"
	"github.com/juju/juju/core/network"
	"github.com/juju/juju/core/objectstore"
	domainapplication "github.com/juju/juju/domain/application"
	applicationcharm "github.com/juju/juju/domain/application/charm"
	"github.com/juju/juju/environs/bootstrap"
	"github.com/juju/juju/internal/charm"
	"github.com/juju/juju/internal/charm/charmdownloader"
	"github.com/juju/juju/internal/charm/repository"
	"github.com/juju/juju/internal/errors"
)

// DeployCharmResult holds the result of deploying a charm.
type DeployCharmInfo struct {
	URL             *charm.URL
	Charm           charm.Charm
	Origin          *corecharm.Origin
	DownloadInfo    *corecharm.DownloadInfo
	ArchivePath     string
	ObjectStoreUUID objectstore.UUID
}

// Validate validates the DeployCharmInfo.
func (d DeployCharmInfo) Validate() error {
	if d.URL == nil {
		return errors.Errorf("URL is nil").Add(coreerrors.NotValid)
	}
	if d.Charm == nil {
		return errors.Errorf("Charm is nil").Add(coreerrors.NotValid)
	}
	if d.Origin == nil {
		return errors.Errorf("Origin is nil").Add(coreerrors.NotValid)
	}
	if err := d.Origin.Validate(); err != nil {
		return errors.Errorf("Origin: %w", err)
	}
	return nil
}

// ControllerCharmDeployer is the interface that is used to deploy the
// controller charm.
type ControllerCharmDeployer interface {
	// DeployLocalCharm deploys the controller charm from the local charm
	// store.
	DeployLocalCharm(context.Context, string, corebase.Base) (DeployCharmInfo, error)

	// DeployCharmhubCharm deploys the controller charm from charm hub.
	DeployCharmhubCharm(context.Context, string, corebase.Base) (DeployCharmInfo, error)

	// AddIAASControllerApplication adds the controller application.
	AddIAASControllerApplication(context.Context, DeployCharmInfo) error

	// AddCAASControllerApplication adds the controller application.
	AddCAASControllerApplication(context.Context, DeployCharmInfo) error

	// ControllerCharmBase returns the base used for deploying the controller
	// charm.
	ControllerCharmBase() (corebase.Base, error)

	// ControllerCharmArch returns the architecture used for deploying the
	// controller charm.
	ControllerCharmArch() string

	// CompleteCAASProcess is called when the bootstrap process is complete.
	CompleteCAASProcess(context.Context) error
}

// HTTPClient is the interface that is used to make HTTP requests.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// CharmHubRepoFunc is the function that is used to create a charm repository.
type CharmHubRepoFunc func(repository.CharmHubRepositoryConfig) (corecharm.Repository, error)

// Downloader defines an API for downloading and storing charms.
type Downloader interface {
	// Download looks up the requested charm using the appropriate store, downloads
	// it to a temporary file and passes it to the configured storage API so it can
	// be persisted.
	//
	// The resulting charm is verified to be the right hash. It expected that the
	// origin will always have the correct hash following this call.
	//
	// Returns [ErrInvalidHash] if the hash of the downloaded charm does not match
	// the expected hash.
	Download(ctx context.Context, url *url.URL, hash string) (*charmdownloader.DownloadResult, error)
}

// CharmDownloaderFunc is the function that is used to create a charm
// downloader.
type CharmDownloaderFunc func(HTTPClient, logger.Logger) Downloader

// BaseDeployerConfig holds the configuration for a baseDeployer.
type BaseDeployerConfig struct {
	DataDir              string
	ApplicationService   ApplicationService
	AgentPasswordService AgentPasswordService
	ModelConfigService   ModelConfigService
	ObjectStore          objectstore.ObjectStore
	Constraints          constraints.Value
	BootstrapAddresses   network.ProviderAddresses
	ControllerConfig     controller.Config
	NewCharmHubRepo      CharmHubRepoFunc
	NewCharmDownloader   CharmDownloaderFunc
	CharmhubHTTPClient   HTTPClient
	ControllerCharmName  string
	Channel              charm.Channel
	Logger               logger.Logger
	Clock                clock.Clock
}

// Validate validates the configuration.
func (c BaseDeployerConfig) Validate() error {
	if c.DataDir == "" {
		return errors.Errorf("DataDir").Add(coreerrors.NotValid)
	}
	if c.AgentPasswordService == nil {
		return errors.Errorf("AgentPasswordService").Add(coreerrors.NotValid)
	}
	if c.ModelConfigService == nil {
		return errors.Errorf("ModelConfigService").Add(coreerrors.NotValid)
	}
	if c.ObjectStore == nil {
		return errors.Errorf("ObjectStore").Add(coreerrors.NotValid)
	}
	if c.BootstrapAddresses == nil {
		return errors.Errorf("BootstrapAddresses").Add(coreerrors.NotValid)
	}
	if c.ControllerConfig == nil {
		return errors.Errorf("ControllerConfig").Add(coreerrors.NotValid)
	}
	if c.NewCharmHubRepo == nil {
		return errors.Errorf("NewCharmHubRepo").Add(coreerrors.NotValid)
	}
	if c.NewCharmDownloader == nil {
		return errors.Errorf("NewCharmDownloader").Add(coreerrors.NotValid)
	}
	if c.CharmhubHTTPClient == nil {
		return errors.Errorf("CharmhubHTTPClient").Add(coreerrors.NotValid)
	}
	if c.Logger == nil {
		return errors.Errorf("Logger").Add(coreerrors.NotValid)
	}
	if c.Clock == nil {
		return errors.Errorf("Clock").Add(coreerrors.NotValid)
	}
	return nil
}

type baseDeployer struct {
	dataDir             string
	applicationService  ApplicationService
	passwordService     AgentPasswordService
	modelConfigService  ModelConfigService
	objectStore         objectstore.ObjectStore
	bootstrapAddresses  network.ProviderAddresses
	constraints         constraints.Value
	controllerConfig    controller.Config
	newCharmHubRepo     CharmHubRepoFunc
	charmhubHTTPClient  HTTPClient
	charmDownloader     CharmDownloaderFunc
	controllerCharmName string
	channel             charm.Channel
	logger              logger.Logger
	clock               clock.Clock
}

func makeBaseDeployer(config BaseDeployerConfig) baseDeployer {
	return baseDeployer{
		dataDir:             config.DataDir,
		applicationService:  config.ApplicationService,
		passwordService:     config.AgentPasswordService,
		modelConfigService:  config.ModelConfigService,
		objectStore:         config.ObjectStore,
		bootstrapAddresses:  config.BootstrapAddresses,
		constraints:         config.Constraints,
		controllerConfig:    config.ControllerConfig,
		newCharmHubRepo:     config.NewCharmHubRepo,
		charmhubHTTPClient:  config.CharmhubHTTPClient,
		charmDownloader:     config.NewCharmDownloader,
		controllerCharmName: config.ControllerCharmName,
		channel:             config.Channel,
		logger:              config.Logger,
		clock:               config.Clock,
	}
}

// ControllerCharmArch returns the architecture used for deploying the
// controller charm.
func (b *baseDeployer) ControllerCharmArch() string {
	arch := corearch.DefaultArchitecture
	if b.constraints.HasArch() {
		arch = *b.constraints.Arch
	}
	return arch
}

// DeployLocalCharm deploys the controller charm from the local charm
// store.
func (b *baseDeployer) DeployLocalCharm(ctx context.Context, arch string, base corebase.Base) (DeployCharmInfo, error) {
	path := filepath.Join(b.dataDir, "charms", bootstrap.ControllerCharmArchive)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return DeployCharmInfo{}, errors.Errorf("opening charm with path %q", path).Add(coreerrors.NotFound)
	} else if err != nil {
		return DeployCharmInfo{}, errors.Capture(err)
	}

	sha256, sha384, err := b.calculateLocalCharmHashes(path, info.Size())
	if err != nil {
		return DeployCharmInfo{}, errors.Errorf("calculating hashes for %q: %w", path, err)
	}

	result, err := b.applicationService.ResolveControllerCharmDownload(ctx, domainapplication.ResolveControllerCharmDownload{
		SHA256: sha256,
		SHA384: sha384,
		Path:   path,
		Size:   info.Size(),
	})
	if err != nil {
		return DeployCharmInfo{}, errors.Errorf("resolving controller charm download: %w", err)
	}

	b.logger.Debugf(ctx, "Successfully deployed local Juju controller charm")

	// The revision of the charm will always be zero during bootstrap, so
	// there is no need to have additional logic to determine the revision.
	// Also there is no architecture, because that prevents the URL from
	// being parsed.

	origin := corecharm.Origin{
		Source:   corecharm.Local,
		Type:     "charm",
		Hash:     sha256,
		Revision: ptr(0),
		Platform: corecharm.Platform{
			Architecture: arch,
			OS:           base.OS,
			Channel:      base.Channel.String(),
		},
	}

	return DeployCharmInfo{
		URL: &charm.URL{
			Schema:   charm.Local.String(),
			Name:     bootstrap.ControllerCharmName,
			Revision: 0,
		},
		Charm:           result.Charm,
		Origin:          &origin,
		ArchivePath:     result.ArchivePath,
		ObjectStoreUUID: result.ObjectStoreUUID,
	}, nil
}

// DeployCharmhubCharm deploys the controller charm from charm hub.
func (b *baseDeployer) DeployCharmhubCharm(ctx context.Context, arch string, base corebase.Base) (DeployCharmInfo, error) {
	modelCfg, err := b.modelConfigService.ModelConfig(ctx)
	if err != nil {
		return DeployCharmInfo{}, errors.Capture(err)
	}
	charmhubURL, _ := modelCfg.CharmHubURL()

	charmRepo, err := b.newCharmHubRepo(repository.CharmHubRepositoryConfig{
		Logger:             b.logger,
		CharmhubURL:        charmhubURL,
		CharmhubHTTPClient: b.charmhubHTTPClient,
	})
	if err != nil {
		return DeployCharmInfo{}, errors.Capture(err)
	}

	var curl *charm.URL
	if b.controllerCharmName == "" {
		curl = charm.MustParseURL(controllerCharmURL)
	} else {
		curl = charm.MustParseURL(b.controllerCharmName)
	}
	origin := corecharm.Origin{
		Source:  corecharm.CharmHub,
		Type:    "charm",
		Channel: &b.channel,
		Platform: corecharm.Platform{
			Architecture: arch,
			OS:           base.OS,
			Channel:      base.Channel.Track,
		},
	}

	// Since we're running on the machine to which the controller charm will be
	// deployed, we know the exact platform to ask for, no need to review the
	// supported base.
	//
	// We prefer the latest LTS bases, if the current base is not one,
	// charmRepo.ResolveWithPreferredChannel, will return an origin with the
	// latest LTS based on data provided by charmhub in the revision-not-found
	// error response.
	//
	// The controller charm doesn't have any base specific code.
	resolved, err := charmRepo.ResolveWithPreferredChannel(ctx, curl.Name, origin)
	if err != nil {
		return DeployCharmInfo{}, errors.Errorf("resolving %q: %w", controllerCharmURL, err)
	}

	downloadInfo := resolved.EssentialMetadata.DownloadInfo

	downloadURL, err := url.Parse(downloadInfo.DownloadURL)
	if err != nil {
		return DeployCharmInfo{}, errors.Errorf("parsing download URL %q: %w", downloadInfo.DownloadURL, err)
	}

	charmDownloader := b.charmDownloader(b.charmhubHTTPClient, b.logger)
	downloadResult, err := charmDownloader.Download(ctx, downloadURL, resolved.Origin.Hash)
	if err != nil {
		return DeployCharmInfo{}, errors.Errorf("downloading %q: %w", downloadURL, err)
	}

	// We can pass the computed SHA384 because we've ensured that the download
	// SHA256 was correct.

	result, err := b.applicationService.ResolveControllerCharmDownload(ctx, domainapplication.ResolveControllerCharmDownload{
		SHA256: resolved.Origin.Hash,
		SHA384: downloadResult.SHA384,
		Path:   downloadResult.Path,
		Size:   downloadResult.Size,
	})
	if err != nil {
		return DeployCharmInfo{}, errors.Errorf("resolving controller charm download: %w", err)
	}

	if resolved.Origin.Revision == nil {
		return DeployCharmInfo{}, errors.Errorf("resolved charm %q has no revision", resolved.URL)
	}

	b.logger.Debugf(ctx, "Successfully deployed charmhub Juju controller charm")

	curl = curl.
		WithRevision(*resolved.Origin.Revision).
		WithArchitecture(resolved.Origin.Platform.Architecture)

	return DeployCharmInfo{
		URL:             curl,
		Charm:           result.Charm,
		Origin:          &resolved.Origin,
		DownloadInfo:    &downloadInfo,
		ArchivePath:     result.ArchivePath,
		ObjectStoreUUID: result.ObjectStoreUUID,
	}, nil
}

// AddIAASControllerApplication adds the IAAS controller application.
func (b *baseDeployer) AddIAASControllerApplication(ctx context.Context, info DeployCharmInfo) error {
	// These are abstract methods that are expected to be implemented by
	// concrete types.
	return errors.Errorf("can not add IAAS controller application").Add(coreerrors.NotImplemented)
}

// AddIAASControllerApplication adds the IAAS controller application.
func (b *baseDeployer) AddCAASControllerApplication(ctx context.Context, info DeployCharmInfo) error {
	// These are abstract methods that are expected to be implemented by
	// concrete types.
	return errors.Errorf("can not add CAAS controller application").Add(coreerrors.NotImplemented)
}

// CompleteCAASProcess is called when the bootstrap process is complete.
func (b *baseDeployer) CompleteCAASProcess(context.Context) error {
	// These are abstract methods that are expected to be implemented by
	// concrete types.
	return errors.Errorf("can not complete CAAS process").Add(coreerrors.NotImplemented)
}

func (b *baseDeployer) calculateLocalCharmHashes(path string, expectedSize int64) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", errors.Errorf("opening %q: %w", path, err)
	}

	hasher256 := sha256.New()
	hasher384 := sha512.New384()

	if size, err := io.Copy(io.MultiWriter(hasher256, hasher384), file); err != nil {
		return "", "", errors.Errorf("hashing %q: %w", path, err)
	} else if size != expectedSize {
		return "", "", errors.Errorf("expected %d bytes, got %d", expectedSize, size)
	}

	sha256 := hex.EncodeToString(hasher256.Sum(nil))
	sha384 := hex.EncodeToString(hasher384.Sum(nil))
	return sha256, sha384, nil
}

func (b *baseDeployer) createCharmSettings() (config.ConfigAttributes, error) {
	cfg := config.ConfigAttributes{
		"is-juju": true,
	}
	cfg["identity-provider-url"] = b.controllerConfig.IdentityURL()

	// Attempt to set the controller URL on to the controller charm config.
	addr := b.controllerConfig.PublicDNSAddress()
	if addr == "" {
		if providerAddress, ok := b.bootstrapAddresses.OneMatchingScope(network.ScopeMatchPublic); ok {
			addr = providerAddress.MachineAddress.Host()
		}
	}
	if addr != "" {
		cfg["controller-url"] = api.ControllerAPIURL(addr, b.controllerConfig.APIPort())
	}
	return cfg, nil
}

func (b *baseDeployer) controllerDownloadInfo(schema string, info *corecharm.DownloadInfo) (*applicationcharm.DownloadInfo, error) {
	// DownloadInfo is not required for local charms, so we only set it if
	// it's not nil.
	if schema == charm.Local.String() && info != nil {
		return nil, errors.New("download info should not be set for local charms")
	}

	if info == nil {
		return nil, nil
	}

	return &applicationcharm.DownloadInfo{
		Provenance:         applicationcharm.ProvenanceBootstrap,
		CharmhubIdentifier: info.CharmhubIdentifier,
		DownloadURL:        info.DownloadURL,
		DownloadSize:       info.DownloadSize,
	}, nil

}
func ptr[T any](v T) *T {
	return &v
}
