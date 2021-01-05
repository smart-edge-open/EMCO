// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package controller

import (
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	rpc "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/rpc"
	pkgerrors "github.com/pkg/errors"

	clmModel "github.com/open-ness/EMCO/src/clm/pkg/model"
)

// ControllerManager is an interface exposes the Controller functionality
type ControllerManager interface {
	CreateController(ms clmModel.Controller, mayExist bool) (clmModel.Controller, error)
	GetController(name string) (clmModel.Controller, error)
	GetControllers() ([]clmModel.Controller, error)
	InitControllers()
	DeleteController(name string) error
}

// ControllerClient implements the Manager
// It will also be used to maintain some localized state
type ControllerClient struct {
	collectionName string
	tagMeta        string
}

// NewControllerClient returns an instance of the ControllerClient
// which implements the Manager
func NewControllerClient() *ControllerClient {
	return &ControllerClient{
		collectionName: "controller",
		tagMeta:        "controllermetadata",
	}
}

// CreateController a new collection based on the Controller
func (mc *ControllerClient) CreateController(m clmModel.Controller, mayExist bool) (clmModel.Controller, error) {

	log.Info("CLM CreateController .. start", log.Fields{"Controller": m, "exists": mayExist})

	//Construct the composite key to select the entry
	key := clmModel.ControllerKey{
		ControllerName: m.Metadata.Name,
	}

	//Check if this Controller already exists
	_, err := mc.GetController(m.Metadata.Name)
	if err == nil && !mayExist {
		return clmModel.Controller{}, pkgerrors.New("ClmController already exists")
	}

	err = db.DBconn.Insert(mc.collectionName, key, nil, mc.tagMeta, m)
	if err != nil {
		return clmModel.Controller{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	// send message to create/update the  rpc connection
	rpc.UpdateRpcConn(m.Metadata.Name, m.Spec.Host, m.Spec.Port)

	log.Info("CLM CreateController .. end", log.Fields{"Controller": m, "exists": mayExist})
	return m, nil
}

// GetController returns the Controller for corresponding name
func (mc *ControllerClient) GetController(name string) (clmModel.Controller, error) {

	//Construct the composite key to select the entry
	key := clmModel.ControllerKey{
		ControllerName: name,
	}
	value, err := db.DBconn.Find(mc.collectionName, key, mc.tagMeta)
	if err != nil {
		return clmModel.Controller{}, pkgerrors.Wrap(err, "db Find error")
	}

	if value != nil {
		microserv := clmModel.Controller{}
		err = db.DBconn.Unmarshal(value[0], &microserv)
		if err != nil {
			return clmModel.Controller{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return microserv, nil
	}

	return clmModel.Controller{}, pkgerrors.New("Error getting Controller")
}

// GetControllers returns all the  Controllers that are registered
func (mc *ControllerClient) GetControllers() ([]clmModel.Controller, error) {

	//Construct the composite key to select the entry
	key := clmModel.ControllerKey{
		ControllerName: "",
	}

	var resp []clmModel.Controller
	values, err := db.DBconn.Find(mc.collectionName, key, mc.tagMeta)
	if err != nil {
		return []clmModel.Controller{}, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		microserv := clmModel.Controller{}
		err = db.DBconn.Unmarshal(value, &microserv)
		if err != nil {
			return []clmModel.Controller{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}

		resp = append(resp, microserv)
	}

	return resp, nil
}

// DeleteController the  Controller from database
func (mc *ControllerClient) DeleteController(name string) error {

	//Construct the composite key to select the entry
	key := clmModel.ControllerKey{
		ControllerName: name,
	}
	err := db.DBconn.Remove(mc.collectionName, key)
	if err != nil {
		if strings.Contains(err.Error(), "Error finding:") {
			return pkgerrors.Wrap(err, "db Remove error - not found")
		} else if strings.Contains(err.Error(), "Can't delete parent without deleting child") {
			return pkgerrors.Wrap(err, "db Remove error - conflict")
		} else {
			return pkgerrors.Wrap(err, "db Remove error - general")
		}
	}

	// send message to close rpc connection
	rpc.RemoveRpcConn(name)

	return nil
}

// InitControllers initializes connctions for Controllers in the DB
func (mc *ControllerClient) InitControllers() {
	vals, _ := mc.GetControllers()
	for _, v := range vals {
		log.Info("Initializing RPC connection for Controller", log.Fields{
			"Controller": v.Metadata.Name,
		})
		rpc.UpdateRpcConn(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
	}
}
