// Copyright 2026 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package storage

// VolumeUUID represents the unique id for a storage volume instance.
type VolumeUUID baseUUID

// NewVolumeUUID creates a new, valid storage volume identifier.
func NewVolumeUUID() (VolumeUUID, error) {
	u, err := newUUID()
	return VolumeUUID(u), err
}

// String returns the string representation of this volume uuid. This function
// satisfies the [fmt.Stringer] interface.
func (u VolumeUUID) String() string {
	return baseUUID(u).String()
}

// Validate returns an error if the [VolumeUUID] is not valid.
func (u VolumeUUID) Validate() error {
	return baseUUID(u).validate()
}
