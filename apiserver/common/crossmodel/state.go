// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package crossmodel

import (
	"context"

	"github.com/juju/errors"
	"github.com/juju/names/v6"

	"github.com/juju/juju/core/crossmodel"
	"github.com/juju/juju/internal/relation"
	"github.com/juju/juju/state"
)

// GetBackend wraps a State to provide a Backend interface implementation.
func GetBackend(st *state.State) stateShim {
	model, err := st.Model()
	if err != nil {
		logger.Errorf(context.TODO(), "called GetBackend on a State with no Model.")
		return stateShim{}
	}
	return stateShim{State: st, Model: model}
}

// TODO - CAAS(ericclaudejones): This should contain state alone, model will be
// removed once all relevant methods are moved from state to model.
type stateShim struct {
	*state.State
	*state.Model
}

func (st stateShim) KeyRelation(key string) (Relation, error) {
	r, err := st.State.KeyRelation(key)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return relationShim{r, st.State}, nil
}

func (st stateShim) OfferConnectionForRelation(relationKey string) (OfferConnection, error) {
	return st.State.OfferConnectionForRelation(relationKey)
}

// ControllerTag returns the tag of the controller in which we are operating.
// This is a temporary transitional step. Eventually code using
// crossmodel.Backend will only need to be passed a state.Model.
func (st stateShim) ControllerTag() names.ControllerTag {
	return st.Model.ControllerTag()
}

// ModelTag returns the tag of the model in which we are operating.
// This is a temporary transitional step.
func (st stateShim) ModelTag() names.ModelTag {
	return st.Model.ModelTag()
}

type applicationShim struct {
	*state.Application
}

func (a applicationShim) EndpointBindings() (Bindings, error) {
	return a.Application.EndpointBindings()
}

func (st stateShim) Application(name string) (Application, error) {
	a, err := st.State.Application(name)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return applicationShim{a}, nil
}

type remoteApplicationShim struct {
	*state.RemoteApplication
}

func (a remoteApplicationShim) DestroyOperation(force bool) state.ModelOperation {
	return a.RemoteApplication.DestroyOperation(force)
}

func (st stateShim) RemoteApplication(name string) (RemoteApplication, error) {
	a, err := st.State.RemoteApplication(name)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &remoteApplicationShim{a}, nil
}

func (st stateShim) AddRelation(eps ...relation.Endpoint) (Relation, error) {
	r, err := st.State.AddRelation(eps...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return relationShim{r, st.State}, nil
}

func (st stateShim) EndpointsRelation(eps ...relation.Endpoint) (Relation, error) {
	r, err := st.State.EndpointsRelation(eps...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return relationShim{r, st.State}, nil
}

func (st stateShim) AddRemoteApplication(args state.AddRemoteApplicationParams) (RemoteApplication, error) {
	a, err := st.State.AddRemoteApplication(args)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return remoteApplicationShim{a}, nil
}

func (st stateShim) OfferUUIDForRelation(key string) (string, error) {
	oc, err := st.State.OfferConnectionForRelation(key)
	if err != nil {
		return "", errors.Trace(err)
	}
	return oc.OfferUUID(), nil
}

func (st stateShim) GetRemoteEntity(token string) (names.Tag, error) {
	r := st.State.RemoteEntities()
	return r.GetRemoteEntity(token)
}

func (st stateShim) GetToken(entity names.Tag) (string, error) {
	r := st.State.RemoteEntities()
	return r.GetToken(entity)
}

func (st stateShim) ExportLocalEntity(entity names.Tag) (string, error) {
	r := st.State.RemoteEntities()
	return r.ExportLocalEntity(entity)
}

func (st stateShim) ImportRemoteEntity(entity names.Tag, token string) error {
	r := st.State.RemoteEntities()
	return r.ImportRemoteEntity(entity, token)
}

func (st stateShim) ApplicationOfferForUUID(offerUUID string) (*crossmodel.ApplicationOffer, error) {
	return state.NewApplicationOffers(st.State).ApplicationOfferForUUID(offerUUID)
}

func (s stateShim) SaveIngressNetworks(relationKey string, cidrs []string) (state.RelationNetworks, error) {
	api := state.NewRelationIngressNetworks(s.State)
	return api.Save(relationKey, false, cidrs)
}

func (s stateShim) IngressNetworks(relationKey string) (state.RelationNetworks, error) {
	api := state.NewRelationIngressNetworks(s.State)
	return api.Networks(relationKey)
}

type relationShim struct {
	*state.Relation
	st *state.State
}

func (r relationShim) RemoteUnit(unitId string) (RelationUnit, error) {
	ru, err := r.Relation.RemoteUnit(unitId)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return relationUnitShim{ru}, nil
}

func (r relationShim) AllRemoteUnits(appName string) ([]RelationUnit, error) {
	all, err := r.Relation.AllRemoteUnits(appName)
	if err != nil {
		return nil, errors.Trace(err)
	}
	result := make([]RelationUnit, len(all))
	for i, ru := range all {
		result[i] = relationUnitShim{ru}
	}
	return result, nil
}

func (r relationShim) Unit(unitId string) (RelationUnit, error) {
	unit, err := r.st.Unit(unitId)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ru, err := r.Relation.Unit(unit)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return relationUnitShim{ru}, nil
}

func (r relationShim) ReplaceApplicationSettings(appName string, values map[string]interface{}) error {
	currentSettings, err := r.ApplicationSettings(appName)
	if err != nil {
		return errors.Trace(err)
	}
	// This is a replace rather than an update so make the update
	// remove any settings missing from the new values.
	for key := range currentSettings {
		if _, found := values[key]; !found {
			values[key] = ""
		}
	}
	// We're replicating changes from another controller so we need to
	// trust them that the leadership was managed correctly - we can't
	// check it here.
	return errors.Trace(r.UpdateApplicationSettings(appName, &successfulToken{}, values))
}

type successfulToken struct{}

// Check is all of the lease.Token interface.
func (t successfulToken) Check() error {
	return nil
}

type relationUnitShim struct {
	*state.RelationUnit
}

func (r relationUnitShim) Settings() (map[string]interface{}, error) {
	settings, err := r.RelationUnit.Settings()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return settings.Map(), nil
}

func (r relationUnitShim) ReplaceSettings(s map[string]interface{}) error {
	settings, err := r.RelationUnit.Settings()
	if err != nil {
		return errors.Trace(err)
	}
	settings.Update(s)
	for _, key := range settings.Keys() {
		if _, ok := s[key]; ok {
			continue
		}
		settings.Delete(key)
	}
	_, err = settings.Write()
	return errors.Trace(err)
}
