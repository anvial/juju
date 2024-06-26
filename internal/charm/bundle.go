// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package charm

import (
	"os"
	"path/filepath"
	"strings"
)

// The Bundle interface is implemented by any type that
// may be handled as a bundle. It encapsulates all
// the data of a bundle.
type Bundle interface {
	// Data returns the contents of the bundle's bundle.yaml file.
	Data() *BundleData
	// BundleBytes returns the raw bytes content of a bundle
	BundleBytes() []byte
	// ReadMe returns the contents of the bundle's README.md file.
	ReadMe() string
	// ContainsOverlays returns true if the bundle contains any overlays.
	ContainsOverlays() bool
}

// ReadBundle reads a Bundle from path, which can point to either a
// bundle archive or a bundle directory.
func ReadBundle(path string) (Bundle, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return ReadBundleDir(path)
	}
	return ReadBundleArchive(path)
}

// IsValidLocalCharmOrBundlePath returns true if path is valid for reading a
// local charm or bundle.
func IsValidLocalCharmOrBundlePath(path string) bool {
	return strings.HasPrefix(path, ".") || filepath.IsAbs(path)
}
