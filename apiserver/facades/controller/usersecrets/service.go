// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package usersecrets

import (
	"context"

	"github.com/juju/juju/core/secrets"
	"github.com/juju/juju/core/watcher"
)

// SecretService instances provide secret apis.
type SecretService interface {
	GetSecret(ctx context.Context, uri *secrets.URI) (*secrets.SecretMetadata, error)
	DeleteObsoleteUserSecrets(ctx context.Context) error
	WatchObsoleteUserSecrets(ctx context.Context) (watcher.NotifyWatcher, error)
}
