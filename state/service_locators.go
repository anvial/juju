// Copyright 2022 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state

import (
	"fmt"
	"github.com/juju/errors"
	"github.com/juju/mgo/v2/bson"
	"github.com/juju/mgo/v2/txn"

	"github.com/juju/juju/rpc/params"
)

//// ServiceLocators describes the state functionality for service locators.
//type ServiceLocators interface {
//	// AllServiceLocators returns the list of all service locators.
//	//AllServiceLocators() ([]*ServiceLocator, error)
//}

// ServiceLocatorsState returns the service locators for the current state.
func (st *State) ServiceLocatorsState() *serviceLocatorPersistence {
	return &serviceLocatorPersistence{
		st: st,
	}
}

var slLogger = logger.Child("service-locator")

// serviceLocatorPersistence provides the persistence
// functionality for service locators.
type serviceLocatorPersistence struct {
	st *State
}

// ServiceLocator represents the state of a juju service locator.
type ServiceLocator struct {
	st  *State
	doc serviceLocatorDoc
}

type serviceLocatorDoc struct {
	DocId              string                 `bson:"_id"`
	Id                 int                    `bson:"id"`
	UnitId             string                 `bson:"unit-id"`
	ConsumerUnitId     string                 `bson:"consumer-unit-id"`
	ConsumerRelationId int                    `bson:"consumer-relation-id"`
	Name               string                 `bson:"name"`
	Type               string                 `bson:"type"`
	Params             map[string]interface{} `bson:"params"`
}

func newServiceLocator(st *State, doc *serviceLocatorDoc) *ServiceLocator {
	serviceLocator := &ServiceLocator{
		st:  st,
		doc: *doc,
	}
	return serviceLocator
}

// Id returns the ID of the service locator.
func (sl *ServiceLocator) Id() int {
	return sl.doc.Id
}

// Name returns the name of the service locator.
func (sl *ServiceLocator) Name() string {
	return sl.doc.Name
}

// Type returns the type of the service locator.
func (sl *ServiceLocator) Type() string {
	return sl.doc.Type
}

// UnitId returns the owner unit ID of the service locator.
func (sl *ServiceLocator) UnitId() string {
	return sl.doc.UnitId
}

// ConsumerUnitId returns the consumer unit ID of the service locator.
func (sl *ServiceLocator) ConsumerUnitId() string {
	return sl.doc.ConsumerUnitId
}

// ConsumerRelationId returns the consumer relation ID of the service locator.
func (sl *ServiceLocator) ConsumerRelationId() int {
	return sl.doc.ConsumerRelationId
}

// Params returns the param list of the service locator.
func (sl *ServiceLocator) Params() map[string]interface{} {
	return sl.doc.Params
}

// AddServiceLocator creates a new service locator record, which records details about a
// network service provided to related units.
func (sp *serviceLocatorPersistence) AddServiceLocator(args params.AddServiceLocatorParams) (*ServiceLocator, error) {
	id, err := sequenceWithMin(sp.st, "service-locator", 1)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer errors.DeferredAnnotatef(&err, "cannot add service locator %q", args.Name)

	model, err := sp.st.Model()
	if err != nil {
		return nil, errors.Trace(err)
	} else if model.Life() != Alive {
		return nil, errors.Errorf("model is no longer alive")
	}

	serviceLocatorDoc := serviceLocatorDoc{
		DocId:              fmt.Sprintf("%s.%s", args.Name, args.UnitId),
		Id:                 id,
		Name:               args.Name,
		Type:               args.Type,
		UnitId:             args.UnitId,
		ConsumerUnitId:     args.ConsumerUnitId,
		ConsumerRelationId: args.ConsumerRelationId,
		Params:             args.Params,
	}
	buildTxn := func(attempt int) ([]txn.Op, error) {
		// If we've tried once already and failed, check that
		// model may have been destroyed.
		if attempt > 0 {
			if err := checkModelActive(sp.st); err != nil {
				return nil, errors.Trace(err)
			}
			return nil, errors.AlreadyExistsf("service locator name: %s unit-id: %s", args.Name, args.UnitId)
		}
		ops := []txn.Op{
			model.assertActiveOp(),
			{
				C:      serviceLocatorsC,
				Id:     serviceLocatorDoc.DocId,
				Assert: txn.DocMissing,
				Insert: &serviceLocatorDoc,
			},
		}
		return ops, nil
	}
	if err = sp.st.db().Run(buildTxn); err != nil {
		return nil, errors.Trace(err)
	}
	return &ServiceLocator{doc: serviceLocatorDoc}, nil
}

// RemoveServiceLocator removes a service locator record
func (sp *serviceLocatorPersistence) RemoveServiceLocator(slId string) []txn.Op {
	op := txn.Op{
		C:      serviceLocatorsC,
		Id:     slId,
		Remove: true,
	}
	return []txn.Op{op}
}

// AllServiceLocators returns all service locators in the model.
func (sp *serviceLocatorPersistence) AllServiceLocators() ([]*ServiceLocator, error) {
	locators, err := sp.serviceLocators(nil)
	return locators, errors.Annotate(err, "getting service locators")
}

// ServiceLocator returns the service locator.
func (sp *serviceLocatorPersistence) ServiceLocator(slId string) ([]*ServiceLocator, error) {
	locators, err := sp.serviceLocators(bson.D{{"id", slId}})
	return locators, errors.Annotatef(err, "getting service locators for %v", slId)
}

// serviceLocators returns the service locators for the input condition
func (sp *serviceLocatorPersistence) serviceLocators(condition bson.D) ([]*ServiceLocator, error) {
	serviceLocatorCollection, closer := sp.st.db().GetCollection(serviceLocatorsC)
	defer closer()

	var locatorDocs []serviceLocatorDoc
	if err := serviceLocatorCollection.Find(condition).All(&locatorDocs); err != nil {
		return nil, errors.Trace(err)
	}

	locators := make([]*ServiceLocator, len(locatorDocs))
	for i, v := range locatorDocs {
		locators[i] = newServiceLocator(sp.st, &v)
	}
	return locators, nil
}