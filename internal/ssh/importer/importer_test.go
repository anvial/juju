// Copyright 2024 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package importer

import (
	"context"
	"net/url"
	"slices"

	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gomock "go.uber.org/mock/gomock"
	gc "gopkg.in/check.v1"

	importererrors "github.com/juju/juju/internal/ssh/importer/errors"
)

type importerSuite struct {
	resolver *MockResolver
}

var (
	_ = gc.Suite(&importerSuite{})
)

func (i *importerSuite) setupMocks(c *gc.C) *gomock.Controller {
	ctrl := gomock.NewController(c)
	i.resolver = NewMockResolver(ctrl)
	return ctrl
}

// TestInvalidURI is testing that if we don't supply a subject in the URI we get
// back an error that satisfies [errors.NotValid]
func (i *importerSuite) TestInvalidURI(c *gc.C) {
	defer i.setupMocks(c).Finish()

	uri, err := url.Parse("gh:")
	c.Assert(err, jc.ErrorIsNil)
	importer := Importer{
		resolvers: map[string]Resolver{
			"gh": i.resolver,
		},
	}
	_, err = importer.FetchPublicKeysForSubject(context.Background(), uri)
	c.Check(err, jc.ErrorIs, errors.NotValid)
}

// TestNoResolver is testing that if we ask for a subjects public key using a
// resolver that dosn't exist we get back a [importererrors.NoResolver] error.
func (i *importerSuite) TestNoResolver(c *gc.C) {
	defer i.setupMocks(c).Finish()

	uri, err := url.Parse("lp:~tlm")
	c.Assert(err, jc.ErrorIsNil)
	importer := Importer{
		resolvers: map[string]Resolver{
			"gh": i.resolver,
		},
	}
	_, err = importer.FetchPublicKeysForSubject(context.Background(), uri)
	c.Check(err, jc.ErrorIs, importererrors.NoResolver)
}

// TestSubjectNotFound is testing that if the resolver tells a subject does not
// exist we return a [importererrors.SubjectNotFound] error.
func (i *importerSuite) TestSubjectNotFound(c *gc.C) {
	defer i.setupMocks(c).Finish()

	i.resolver.EXPECT().PublicKeysForSubject(gomock.Any(), "tlm").
		Return(nil, importererrors.SubjectNotFound)

	uri, err := url.Parse("gh:tlm")
	c.Assert(err, jc.ErrorIsNil)
	importer := Importer{
		resolvers: map[string]Resolver{
			"gh": i.resolver,
		},
	}
	_, err = importer.FetchPublicKeysForSubject(context.Background(), uri)
	c.Check(err, jc.ErrorIs, importererrors.SubjectNotFound)
}

// TestFetchPublicKeysForSubject is testing the happy path for
// [Importer.FetchPublicKeysForSubject].
func (i *importerSuite) TestFetchPublicKeysForSubject(c *gc.C) {
	defer i.setupMocks(c).Finish()

	i.resolver.EXPECT().PublicKeysForSubject(gomock.Any(), "tlm").Return(
		[]string{
			"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII4GpCvqUUYUJlx6d1kpUO9k/t4VhSYsf0yE0/QTqDzC key1",
			"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJQJ9wv0uC3yytXM3d2sJJWvZLuISKo7ZHwafHVviwVe key2",
		}, nil,
	)

	uri, err := url.Parse("gh:tlm")
	c.Assert(err, jc.ErrorIsNil)
	importer := Importer{
		resolvers: map[string]Resolver{
			"gh": i.resolver,
		},
	}
	keys, err := importer.FetchPublicKeysForSubject(context.Background(), uri)
	c.Check(err, jc.ErrorIsNil)

	expected := []string{
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII4GpCvqUUYUJlx6d1kpUO9k/t4VhSYsf0yE0/QTqDzC key1",
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJQJ9wv0uC3yytXM3d2sJJWvZLuISKo7ZHwafHVviwVe key2",
	}

	slices.Sort(keys)
	slices.Sort(expected)
	c.Check(keys, jc.DeepEquals, expected)
}
