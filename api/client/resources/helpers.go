// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package resources

import (
	"github.com/juju/errors"
	"github.com/juju/names/v6"

	apiservererrors "github.com/juju/juju/apiserver/errors"
	"github.com/juju/juju/core/resource"
	"github.com/juju/juju/core/unit"
	charmresource "github.com/juju/juju/internal/charm/resource"
	"github.com/juju/juju/rpc/params"
)

// Resource2API converts a resource.Resource into
// a Resource struct.
func Resource2API(res resource.Resource) params.Resource {
	return params.Resource{
		CharmResource:   CharmResource2API(res.Resource),
		UUID:            res.UUID.String(),
		ApplicationName: res.ApplicationName,
		Username:        res.RetrievedBy,
		Timestamp:       res.Timestamp,
	}
}

// apiResult2ApplicationResources converts a ResourcesResult into a resources.ApplicationResources.
func apiResult2ApplicationResources(apiResult params.ResourcesResult) (resource.ApplicationResources, error) {
	var result resource.ApplicationResources

	if apiResult.Error != nil {
		// TODO(ericsnow) Return the resources too?
		err := apiservererrors.RestoreError(apiResult.Error)
		return resource.ApplicationResources{}, errors.Trace(err)
	}

	for _, apiRes := range apiResult.Resources {
		res, err := API2Resource(apiRes)
		if err != nil {
			// This could happen if the server is misbehaving
			// or non-conforming.
			// TODO(ericsnow) Aggregate errors?
			return resource.ApplicationResources{}, errors.Annotate(err, "got bad data from server")
		}
		result.Resources = append(result.Resources, res)
	}

	for _, unitRes := range apiResult.UnitResources {
		tag, err := names.ParseUnitTag(unitRes.Tag)
		if err != nil {
			return resource.ApplicationResources{}, errors.Annotate(err, "got bad data from server")
		}
		resNames := map[string]bool{}
		unitName, err := unit.NewName(tag.Id())
		if err != nil {
			return resource.ApplicationResources{}, errors.Annotate(err, "got bad data from server")
		}
		unitResources := resource.UnitResources{Name: unitName}
		for _, apiRes := range unitRes.Resources {
			res, err := API2Resource(apiRes)
			if err != nil {
				return resource.ApplicationResources{}, errors.Annotate(err, "got bad data from server")
			}
			resNames[res.Name] = true
			unitResources.Resources = append(unitResources.Resources, res)
		}
		result.UnitResources = append(result.UnitResources, unitResources)
	}

	for _, chRes := range apiResult.CharmStoreResources {
		res, err := API2CharmResource(chRes)
		if err != nil {
			return resource.ApplicationResources{}, errors.Annotate(err, "got bad data from server")
		}
		result.RepositoryResources = append(result.RepositoryResources, res)
	}

	return result, nil
}

// API2Resource converts an API Resource struct into
// a resource.Resource.
func API2Resource(apiRes params.Resource) (resource.Resource, error) {
	var res resource.Resource

	charmRes, err := API2CharmResource(apiRes.CharmResource)
	if err != nil {
		return res, errors.Trace(err)
	}

	uuid, err := resource.ParseUUID(apiRes.UUID)
	if err != nil {
		return res, errors.Trace(err)
	}

	res = resource.Resource{
		Resource:        charmRes,
		UUID:            uuid,
		ApplicationName: apiRes.ApplicationName,
		RetrievedBy:     apiRes.Username,
		Timestamp:       apiRes.Timestamp,
	}

	if err := res.Validate(); err != nil {
		return res, errors.Trace(err)
	}

	return res, nil
}

// CharmResource2API converts a charm resource into
// a CharmResource struct.
func CharmResource2API(res charmresource.Resource) params.CharmResource {
	return params.CharmResource{
		Name:        res.Name,
		Type:        res.Type.String(),
		Path:        res.Path,
		Description: res.Description,
		Origin:      res.Origin.String(),
		Revision:    res.Revision,
		Fingerprint: res.Fingerprint.Bytes(),
		Size:        res.Size,
	}
}

// API2CharmResource converts an API CharmResource struct into
// a charm resource.
func API2CharmResource(apiInfo params.CharmResource) (charmresource.Resource, error) {
	var res charmresource.Resource

	rtype, err := charmresource.ParseType(apiInfo.Type)
	if err != nil {
		return res, errors.Trace(err)
	}

	origin, err := charmresource.ParseOrigin(apiInfo.Origin)
	if err != nil {
		return res, errors.Trace(err)
	}

	fp, err := resource.DeserializeFingerprint(apiInfo.Fingerprint)
	if err != nil {
		return res, errors.Trace(err)
	}

	res = charmresource.Resource{
		Meta: charmresource.Meta{
			Name:        apiInfo.Name,
			Type:        rtype,
			Path:        apiInfo.Path,
			Description: apiInfo.Description,
		},
		Origin:      origin,
		Revision:    apiInfo.Revision,
		Fingerprint: fp,
		Size:        apiInfo.Size,
	}

	if err := res.Validate(); err != nil {
		return res, errors.Trace(err)
	}
	return res, nil
}
