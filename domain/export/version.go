// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package export

// ExportVersions maps export version formats to the first Juju version for
// which the export version was generated.
// This drives the automated generation of types for serialisable controller
// and model formats.
// To add a new format, add the next version sequence integer mapped to the
// current Juju version in string form, then run `go generate` in the
// generate/export directory.
var ExportVersions = map[uint64]string{
	0: "4.0.1",
}
