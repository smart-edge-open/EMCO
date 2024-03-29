// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package controller

import (
	"encoding/json"
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	rpc "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/rpc"
	mtypes "github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"
	pkgerrors "github.com/pkg/errors"
)

// Controller contains the parameters needed for Controllers
// It implements the interface for managing the Controllers
type Controller struct {
	Metadata mtypes.Metadata `json:"metadata"`
	Spec     ControllerSpec  `json:"spec"`
}

type ControllerSpec struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Type     string `json:"type"`
	Priority int    `json:"priority"`
}

const MinControllerPriority = 1
const MaxControllerPriority = 1000000
const CONTROLLER_TYPE_ACTION string = "action"
const CONTROLLER_TYPE_PLACEMENT string = "placement"

var CONTROLLER_TYPES = [...]string{CONTROLLER_TYPE_ACTION, CONTROLLER_TYPE_PLACEMENT}

// ControllerKey is the key structure that is used in the database
type ControllerKey struct {
	ControllerName string `json:"controller-name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (mk ControllerKey) String() string {
	out, err := json.Marshal(mk)
	if err != nil {
		return ""
	}

	return string(out)
}

// ControllerManager is an interface exposes the Controller functionality
type ControllerManager interface {
	CreateController(ms Controller, mayExist bool) (Controller, error)
	GetController(name string) (Controller, error)
	GetControllers() ([]Controller, error)
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
func NewControllerClient(name, tag string) *ControllerClient {
	return &ControllerClient{
		collectionName: name,
		tagMeta:        tag,
	}
}

// CreateController a new collection based on the Controller
func (mc *ControllerClient) CreateController(m Controller, mayExist bool) (Controller, error) {

	log.Info("CreateController .. start", log.Fields{"Controller": m, "exists": mayExist})

	//Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName: m.Metadata.Name,
	}

	//Check if this Controller already exists
	_, err := mc.GetController(m.Metadata.Name)
	if err == nil && !mayExist {
		return Controller{}, pkgerrors.New("Controller already exists")
	}

	err = db.DBconn.Insert(mc.collectionName, key, nil, mc.tagMeta, m)
	if err != nil {
		return Controller{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	// send message to create/update the  rpc connection
	rpc.UpdateRpcConn(m.Metadata.Name, m.Spec.Host, m.Spec.Port)

	log.Info("CreateController .. end", log.Fields{"Controller": m, "exists": mayExist})
	return m, nil
}

// GetController returns the Controller for corresponding name
func (mc *ControllerClient) GetController(name string) (Controller, error) {

	//Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName: name,
	}
	value, err := db.DBconn.Find(mc.collectionName, key, mc.tagMeta)
	if err != nil {
		return Controller{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return Controller{}, pkgerrors.New("Controller not found")
	}

	if value != nil {
		microserv := Controller{}
		err = db.DBconn.Unmarshal(value[0], &microserv)
		if err != nil {
			return Controller{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return microserv, nil
	}

	return Controller{}, pkgerrors.New("Error getting Controller")
}

// GetControllers returns all the  Controllers that are registered
func (mc *ControllerClient) GetControllers() ([]Controller, error) {

	//Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName: "",
	}

	var resp []Controller
	values, err := db.DBconn.Find(mc.collectionName, key, mc.tagMeta)
	if err != nil {
		return []Controller{}, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		microserv := Controller{}
		err = db.DBconn.Unmarshal(value, &microserv)
		if err != nil {
			return []Controller{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}

		resp = append(resp, microserv)
	}

	return resp, nil
}

// DeleteController the  Controller from database
func (mc *ControllerClient) DeleteController(name string) error {

	//Construct the composite key to select the entry
	key := ControllerKey{
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

// InitControllers initializes connctions for controllers in the DB
func (mc *ControllerClient) InitControllers() {
	vals, _ := mc.GetControllers()
	for _, v := range vals {
		log.Info("Initializing RPC connection for controller", log.Fields{
			"Controller": v.Metadata.Name,
		})
		rpc.UpdateRpcConn(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
	}
}
