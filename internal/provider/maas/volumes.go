// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package maas

import (
	"context"
	"strings"
	"unicode"

	"github.com/dustin/go-humanize"
	"github.com/juju/collections/set"
	"github.com/juju/errors"
	"github.com/juju/gomaasapi/v2"
	"github.com/juju/names/v6"
	"github.com/juju/schema"

	"github.com/juju/juju/core/constraints"
	"github.com/juju/juju/internal/provider/common"
	"github.com/juju/juju/internal/storage"
)

// maasStorageProvider provides a [storage.Provider] implementation that is not
// capable of provisioning any block devices or filesystems within a Juju model
// on behalf of charm storage requirements.
//
// This provider solely exists to support the provisioning of root disk volumes
// for new machines being provisioned within MAAS.
type maasStorageProvider struct{}

const (
	// maasStorageProviderType is the name of the storage provider
	// used to specify storage when acquiring MAAS nodes.
	maasStorageProviderType = storage.ProviderType("maas")

	// rootDiskLabel is the label recognised by MAAS as being for
	// the root disk.
	rootDiskLabel = "root"

	// tagsAttribute is the name of the pool attribute used
	// to specify tag values for requested volumes.
	tagsAttribute = "tags"
)

// RecommendedPoolForKind returns the recommended storage pool to use for
// the given storage kind. If no pool can be recommended nil is returned. For
// the MAAS environ only builtin IAAS pools are returned.
//
// The [maasStorageProvider] is never recommended as it can only ever be used
// for provisioning root disks on new machines.
//
// Implements [storage.ProviderRegistry] interface.
func (*maasEnviron) RecommendedPoolForKind(
	kind storage.StorageKind,
) *storage.Config {
	return common.GetCommonRecommendedIAASPoolForKind(kind)
}

// StorageProviderTypes returns the set of provider types supported for
// provisioning storage on behalf of this environ.
//
// Implements [storage.ProviderRegistry] interface.
func (*maasEnviron) StorageProviderTypes() ([]storage.ProviderType, error) {
	return append(
		common.CommonIAASStorageProviderTypes(),
		maasStorageProviderType,
	), nil
}

// StorageProvider returns the implementation of [storage.Provider] for the
// given storage provider type. See [maasEnviron.StorageProviderTypes] for the
// set of supported provider types.
//
// Implements [storage.ProviderRegistry] interface.
//
// The following errors may be returned:
// - [github.com/juju/juju/core/errors.NotFound] when no provider exists for the
// supplied [storage.ProviderType].
func (*maasEnviron) StorageProvider(t storage.ProviderType) (storage.Provider, error) {
	switch t {
	case maasStorageProviderType:
		return maasStorageProvider{}, nil
	default:
		return common.GetCommonIAASStorageProvider(t)
	}
}

var storageConfigFields = schema.Fields{
	tagsAttribute: schema.OneOf(
		schema.List(schema.String()),
		schema.String(),
	),
}

var storageConfigChecker = schema.FieldMap(
	storageConfigFields,
	schema.Defaults{
		tagsAttribute: schema.Omit,
	},
)

type storageConfig struct {
	tags []string
}

func newStorageConfig(attrs map[string]interface{}) (*storageConfig, error) {
	out, err := storageConfigChecker.Coerce(attrs, nil)
	if err != nil {
		return nil, errors.Annotate(err, "validating MAAS storage config")
	}
	coerced := out.(map[string]interface{})
	var tags []string
	switch v := coerced[tagsAttribute].(type) {
	case []string:
		tags = v
	case string:
		fields := strings.Split(v, ",")
		for _, f := range fields {
			f = strings.TrimSpace(f)
			if len(f) == 0 {
				continue
			}
			if i := strings.IndexFunc(f, unicode.IsSpace); i >= 0 {
				return nil, errors.Errorf("tags may not contain whitespace: %q", f)
			}
			tags = append(tags, f)
		}
	}
	return &storageConfig{tags: tags}, nil
}

func (maasStorageProvider) ValidateForK8s(map[string]any) error {
	// no validation required
	return nil
}

// ValidateConfig is defined on the Provider interface.
func (maasStorageProvider) ValidateConfig(cfg *storage.Config) error {
	_, err := newStorageConfig(cfg.Attrs())
	return errors.Trace(err)
}

// Supports returns true or false to the caller indicating of the given
// [storage.StorageKind] is supported by this provider for provisioning.Supports
//
// [maasStorageProvider] always returns false for all [storage.StorageKind]
// values. This is because the provider can only provision root disks for newly
// created machines in MAAS.
//
// This current implementation should be considered a quirk of the storage
// provider interface as it only supports asking broad provisioning questions
// and not specifics about the context the provisioning is occuring. We
// understand that this is safe to do for the moment as the provisioning of
// machines doesn't interrogate the capabilities of the storage provider.
//
// Implements the [storage.Provider] interface.
func (maasStorageProvider) Supports(k storage.StorageKind) bool {
	return false
}

// Scope returns the [storage.Scope] for which provisioning of storage occurs.
// For MAAS storage is always provisioned via the API and so is always
// considered to be [storage.ScopeEnviron] as the API calls originate from
// within environ.
//
// Implements the [storage.Provider] interface.
func (maasStorageProvider) Scope() storage.Scope {
	return storage.ScopeEnviron
}

// Dynamic indicates to the caller if this provider supports the provisioning of
// dynamic storage. The answer to this question for MAAS is always false.
//
// Implements the [storage.Provider] interface.
func (maasStorageProvider) Dynamic() bool {
	return false
}

// Releasable indicates to the caller if this provider supports releasing of
// storage from it's attached entity. The answer to this question for MAAS is
// always false.
func (maasStorageProvider) Releasable() bool {
	return false
}

// DefaultPools returns the default pools available through the maas provider.
// By default a pool by the same name as the provider is offered. The reason
// this provider returns a default pool that cannot be used for provisioning of
// storage beside machine root disks is so that a model user can set tag
// attributes on the pool which are then used on the root disk in MAAS.
//
// It should be noted that while a default pool is offered, it is never
// recommended. See [maasEnviron.RecommendedPoolForKind].
//
// Implements [storage.Provider] interface.
func (maasStorageProvider) DefaultPools() []*storage.Config {
	// The error for constructing a new storage pool config is disgarded as the
	// [storage.Provider] interface does not support a error return. Because
	// the configuration is static in nature we assume that unit testing will
	// suffice.
	defaultPool, _ := storage.NewConfig(
		maasStorageProviderType.String(),
		maasStorageProviderType,
		storage.Attrs{},
	)

	return []*storage.Config{defaultPool}
}

// VolumeSource is responsible for returning a [storage.VolumeSource] capable of
// provisioning volumes in the model for this storage provider. The
// [maasStorageProvider] does not support the provisioning of volumes outside of
// root disk for a new machines. Because of this an error satisfying
// [github.com/juju/juju/core/storage] is always returned.
//
// Implements [storage.Provider] interface.
func (maasStorageProvider) VolumeSource(_ *storage.Config) (
	storage.VolumeSource, error,
) {
	return nil, errors.NotSupportedf("volumes")
}

// FilesystemSource is responsible for returning a [storage.FilesystemSource]
// capable of provisioning filesystems in the model for this storage provider.
// The [maasStorageProvider] does not support the provisioning of file systems
// outside of root disk volumes for  new machines. Because of this an error
// satisfying [github.com/juju/juju/core/storage] is always returned.
//
// Implements [storage.Provider] interface.
func (maasStorageProvider) FilesystemSource(_ *storage.Config) (
	storage.FilesystemSource, error,
) {
	return nil, errors.NotSupportedf("filesystems")
}

type volumeInfo struct {
	name     string
	sizeInGB uint64
	tags     []string
}

// mibToGB converts the value in MiB to GB.
// Juju works in MiB, MAAS expects GB.
func mibToGb(m uint64) uint64 {
	return common.MiBToGiB(m) * (humanize.GiByte / humanize.GByte)
}

// buildMAASVolumeParameters creates the MAAS volume information to include
// in a request to acquire a MAAS node, based on the supplied storage parameters.
func buildMAASVolumeParameters(args []storage.VolumeParams, cons constraints.Value) ([]volumeInfo, error) {
	if len(args) == 0 && cons.RootDisk == nil {
		return nil, nil
	}
	volumes := make([]volumeInfo, len(args)+1)
	rootVolume := volumeInfo{name: rootDiskLabel}
	if cons.RootDisk != nil {
		rootVolume.sizeInGB = mibToGb(*cons.RootDisk)
	}
	volumes[0] = rootVolume
	for i, v := range args {
		cfg, err := newStorageConfig(v.Attributes)
		if err != nil {
			return nil, errors.Trace(err)
		}
		info := volumeInfo{
			name:     v.Tag.Id(),
			sizeInGB: mibToGb(v.Size),
			tags:     cfg.tags,
		}
		volumes[i+1] = info
	}
	return volumes, nil
}

func (mi *maasInstance) volumes(
	ctx context.Context,
	mTag names.MachineTag, requestedVolumes []names.VolumeTag,
) (
	[]storage.Volume, []storage.VolumeAttachment, error,
) {
	if mi.constraintMatches.Storage == nil {
		return nil, nil, errors.NotFoundf("constraint storage mapping")
	}

	var volumes []storage.Volume
	var attachments []storage.VolumeAttachment

	// Set up a collection of volumes tags which
	// we specifically asked for when the node was acquired.
	validVolumes := set.NewStrings()
	for _, v := range requestedVolumes {
		validVolumes.Add(v.Id())
	}

	for label, devices := range mi.constraintMatches.Storage {
		// We don't explicitly allow the root volume to be specified yet.
		if label == rootDiskLabel {
			continue
		}

		// We only care about the volumes we specifically asked for.
		if !validVolumes.Contains(label) {
			continue
		}

		// There should be exactly one block device per label.
		if len(devices) == 0 {
			continue
		} else if len(devices) > 1 {
			// This should never happen, as we only request one block
			// device per label. If it does happen, we'll just report
			// the first block device and log this warning.
			logger.Warningf(ctx,
				"expected 1 block device for label %s, received %d",
				label, len(devices),
			)
		}

		device := devices[0]
		volumeTag := names.NewVolumeTag(label)
		vol := storage.Volume{
			volumeTag,
			storage.VolumeInfo{
				VolumeId:   volumeTag.String(),
				Size:       device.Size() / humanize.MiByte,
				Persistent: false,
			},
		}
		attachment := storage.VolumeAttachment{
			volumeTag,
			mTag,
			storage.VolumeAttachmentInfo{
				ReadOnly: false,
			},
		}

		const devDiskByIdPrefix = "/dev/disk/by-id/"
		const devPrefix = "/dev/"

		if blockDev, ok := device.(gomaasapi.BlockDevice); ok {
			// Handle a block device specifically that way the path used
			// by Juju will always be a persistent path.
			idPath := blockDev.IDPath()
			if idPath == devPrefix+blockDev.Name() {
				// On vMAAS (i.e. with virtio), the device name
				// will be stable, and is what is used to form
				// id_path.
				deviceName := idPath[len(devPrefix):]
				attachment.DeviceName = deviceName
			} else if strings.HasPrefix(idPath, devDiskByIdPrefix) {
				const wwnPrefix = "wwn-"
				id := idPath[len(devDiskByIdPrefix):]
				if strings.HasPrefix(id, wwnPrefix) {
					vol.WWN = id[len(wwnPrefix):]
				} else {
					vol.HardwareId = id
				}
			} else {
				// It's neither /dev/<name> nor /dev/disk/by-id/<hardware-id>,
				// so set it as the device link and hope for
				// the best. At worst, the path won't exist
				// and the storage will remain pending.
				attachment.DeviceLink = idPath
			}
		} else {
			// Handle all other storage devices using the path MAAS provided.
			// In the case of partitions the path is always stable because its
			// based on the GUID of the partition using the dname path.
			attachment.DeviceLink = device.Path()
		}

		volumes = append(volumes, vol)
		attachments = append(attachments, attachment)
	}
	return volumes, attachments, nil
}
