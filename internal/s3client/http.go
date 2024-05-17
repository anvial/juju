// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package s3client

import (
	jujuhttp "github.com/juju/http/v2"

	"github.com/juju/juju/core/logger"
)

// DefaultHTTPClient returns the default http client used to access the object
// store.
func DefaultHTTPClient(logger logger.Logger) HTTPClient {
	return jujuhttp.NewClient(
		jujuhttp.WithLogger(httpLogger{
			Logger: logger,
		}),
	)
}

type httpLogger struct {
	logger.Logger
}

func (l httpLogger) IsTraceEnabled() bool {
	return l.IsLevelEnabled(logger.TRACE)
}