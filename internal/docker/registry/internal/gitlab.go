// Copyright 2021 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package internal

import (
	"net/http"
	"strings"

	"github.com/juju/errors"

	"github.com/juju/juju/internal/docker"
)

type gitlabContainerRegistry struct {
	*baseClient
}

func newGitlabContainerRegistry(repoDetails docker.ImageRepoDetails, transport http.RoundTripper) (RegistryInternal, error) {
	c, err := newBase(repoDetails, transport, normalizeRepoDetailsCommon)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &gitlabContainerRegistry{c}, nil
}

func (c *gitlabContainerRegistry) String() string {
	return "registry.gitlab.com"
}

// Match checks if the repository details matches current provider format.
func (c *gitlabContainerRegistry) Match() bool {
	return strings.Contains(c.repoDetails.ServerAddress, "registry.gitlab.com")
}

func (c *gitlabContainerRegistry) WrapTransport(...TransportWrapper) error {
	return c.baseClient.WrapTransport()
}

// Ping pings the gitlab endpoint.
func (c gitlabContainerRegistry) Ping() error {
	if !c.repoDetails.IsPrivate() {
		// The root gitlab endpoint requires credentials.
		// Anonymous login does not work.
		return nil
	}
	return c.baseClient.Ping()
}
