// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"fmt"
	"strconv"
	"time"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/state"
	pkgerrors "github.com/pkg/errors"
)

// MigrateJson contains metadata and spec for migrate API
type MigrateJson struct {
	MetaData UpdateMetadata `json:"metadata,omitempty"`
	Spec     MigrateSpec    `json:"spec"`
}

// RollbackJson contains metadata and spec for rollback API
type RollbackJson struct {
	MetaData UpdateMetadata `json:"metadata,omitempty"`
	Spec     RollbackSpec   `json:"spec"`
}

type UpdateMetadata struct {
	Description string `json:"description"`
}

type MigrateSpec struct {
	TargetCompositeAppVersion string `json:"target-composite-app-version"`
	TargetDigName             string `json:"target-dig-name"`
}

type RollbackSpec struct {
	Revison string `json:"revision"`
}

/*
Migrate methods takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName, targetCompositeAppVersion and targetDeploymentIntentName.
This method is responsible for creation and saving of context for saving into etcd
and ensuring sourceDeploymentIntent gets migrated to targetDeploymentIntent.
*/
func (c InstantiationClient) Migrate(p string, ca string, v string, tCav string, di string, tDi string) error {
	log.Info("Migrate API", log.Fields{"profile": p, "compositeapp": ca, "version": v, "targetcompositeappversion": tCav,
		"sourcedeploymentintentgroup": di, "targetdeploymentintentgroup": tDi})

	// Fetch source DIG context ID
	ss, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info: "+di)
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(ss)
	if err != nil {
		return pkgerrors.Errorf("Error getting current state from DeploymentIntentGroup stateInfo: " + di)
	}

	if stateVal != state.StateEnum.Instantiated && stateVal != state.StateEnum.InstantiateStopped {
		return pkgerrors.Errorf("DeploymentIntentGroup is not instantiated :" + di)
	}

	sourceCtxId := state.GetLastContextIdFromStateInfo(ss)

	// in case of migrate dig comes from JSON body
	dIGrp, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(tDi, p, ca, tCav)
	if err != nil {
		return pkgerrors.Wrap(err, "Not finding the deploymentIntentGroup")
	}

	// BEGIN : Make app context
	instantiator := Instantiator{p, ca, tCav, tDi, dIGrp}
	cca, err := instantiator.MakeAppContext()
	if err != nil {
		return pkgerrors.Wrap(err, "Error in making AppContext")
	}
	// END : Make app context

	// BEGIN : callScheduler
	err = callScheduler(cca.context, cca.ctxval, p, ca, tCav, tDi)
	if err != nil {
		return pkgerrors.Wrap(err, "Error in callScheduler")
	}
	// END : callScheduler

	// Fetch target DIG context ID
	ts, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(tDi, p, ca, tCav)
	if err != nil {
		return pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info: "+tDi)
	}

	stateVal, err = state.GetCurrentStateFromStateInfo(ts)
	if err != nil {
		return pkgerrors.Errorf("Error getting current state from DeploymentIntentGroup stateInfo: " + tDi)
	}

	if stateVal != state.StateEnum.Approved {
		return pkgerrors.Errorf("DeploymentIntentGroup is not approved :" + tDi)
	}

	targetCtxId := fmt.Sprintf("%v", cca.ctxval)

	log.Info("sourceCtxId", log.Fields{"sourceCtxId": sourceCtxId})
	log.Info("targetCtxId", log.Fields{"targetCtxId": targetCtxId})

	// Read the Status ContextID from source
	statusID := state.GetStatusContextIdFromStateInfo(ss)
	// Update Status context id
	err = state.UpdateAppContextStatusContextID(targetCtxId, statusID)
	if err != nil {
		return err
	}

	err = callRsyncUpdate(sourceCtxId, targetCtxId)
	if err != nil {
		return err
	}

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	a := state.ActionEntry{
		State:     state.StateEnum.Updated,
		ContextId: sourceCtxId,
		TimeStamp: time.Now(),
	}
	ss.Actions = append(ss.Actions, a)

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, ss)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+di)
	}

	key = DeploymentIntentGroupKey{
		Name:         tDi,
		Project:      p,
		CompositeApp: ca,
		Version:      tCav,
	}

	a = state.ActionEntry{
		State:     state.StateEnum.Instantiated,
		ContextId: targetCtxId,
		TimeStamp: time.Now(),
	}
	ts.Actions = append(ts.Actions, a)
	// Update the status context ID to match source
	ts.StatusContextId = statusID

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, ts)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+tDi)
	}
	return nil
}

/*
Update methods takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName.
This method is responsible for creation and saving of context into etcd and ensuring new intents are applied on DeploymentIntentGroup.
*/
func (c InstantiationClient) Update(p string, ca string, v string, di string) (int64, error) {

	log.Info("Update API", log.Fields{"profile": p, "compositeapp": ca, "version": v, "deploymentintentgroup": di})

	// Fetch source DIG context ID
	ss, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(di, p, ca, v)
	if err != nil {
		return -1, pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info: "+di)
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(ss)
	if err != nil {
		return -1, pkgerrors.Errorf("Error getting current state from DeploymentIntentGroup stateInfo: " + di)
	}

	if stateVal != state.StateEnum.Instantiated && stateVal != state.StateEnum.InstantiateStopped {
		return -1, pkgerrors.Errorf("DeploymentIntentGroup is not instantiated :" + di)
	}

	sourceCtxId := state.GetLastContextIdFromStateInfo(ss)
	lastRevision, err := state.GetLatestRevisionFromStateInfo(ss)
	if err != nil {
		return -1, pkgerrors.Wrap(err, "Latest revision not found "+di)
	}

	dIGrp, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(di, p, ca, v)
	if err != nil {
		return -1, pkgerrors.Wrap(err, "Not finding the deploymentIntentGroup")
	}

	// BEGIN : Make app context
	instantiator := Instantiator{p, ca, v, di, dIGrp}
	cca, err := instantiator.MakeAppContext()
	if err != nil {
		return -1, pkgerrors.Wrap(err, "Error in making AppContext")
	}
	// END : Make app context

	// BEGIN : callScheduler
	err = callScheduler(cca.context, cca.ctxval, p, ca, v, di)
	if err != nil {
		return -1, pkgerrors.Wrap(err, "Error in callScheduler")
	}
	// END : callScheduler

	targetCtxId := fmt.Sprintf("%v", cca.ctxval)

	// Update Status Context ID in AppContext
	statusID := state.GetStatusContextIdFromStateInfo(ss)
	// Update Status context id to be source status collected in source
	err = state.UpdateAppContextStatusContextID(targetCtxId, statusID)
	if err != nil {
		return -1, err
	}
	err = callRsyncUpdate(sourceCtxId, targetCtxId)
	if err != nil {
		return -1, err
	}

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	// Updating the previous state
	a := state.ActionEntry{
		State:     state.StateEnum.Updated,
		ContextId: sourceCtxId,
		TimeStamp: time.Now(),
		Revision:  lastRevision,
	}
	ss.Actions = append(ss.Actions, a)

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, ss)
	if err != nil {
		return -1, pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+di)
	}

	// TODO : Atomicity check
	latestRevision := lastRevision + 1

	// Instantiating the current state
	a = state.ActionEntry{
		State:     state.StateEnum.Instantiated,
		ContextId: targetCtxId,
		TimeStamp: time.Now(),
		Revision:  latestRevision,
	}
	ss.Actions = append(ss.Actions, a)

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, ss)
	if err != nil {
		return -1, pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+di)
	}

	log.Info("Updated revisionID", log.Fields{"Updated to revisionID": latestRevision})

	return latestRevision, nil

}

/*
Rollback methods takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName and revision.
This method is responsible for creation and saving of context for saving into etcd
and ensuring DeploymentIntentGroup is rollback to given revision.
*/
func (c InstantiationClient) Rollback(p string, ca string, v string, di string, rbRev string) error {
	log.Info("Rollback API", log.Fields{"profile": p, "compositeapp": ca, "version": v, "deploymentintentgroup": di,
		"rbRev": rbRev})

	ss, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info: "+di)
	}

	prevRevisionID, err := state.GetLatestRevisionFromStateInfo(ss)
	if err != nil {
		return pkgerrors.Wrap(err, "Failed to get previous RevisionID")
	}

	sourceCtxId := state.GetLastContextIdFromStateInfo(ss)

	rID, err := strconv.ParseInt(rbRev, 10, 64)
	if err != nil {
		return pkgerrors.Wrap(err, "Parsing error "+rbRev)
	}
	targetCtxId, err := state.GetMatchingContextIDforRevision(ss, rID)
	if err != nil {
		return pkgerrors.Wrap(err, "GetMatchingContextIDforRevision error "+rbRev)
	}

	err = callRsyncUpdate(sourceCtxId, targetCtxId)
	if err != nil {
		return err
	}

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	a := state.ActionEntry{
		State:     state.StateEnum.Updated,
		ContextId: sourceCtxId,
		TimeStamp: time.Now(),
		Revision:  prevRevisionID,
	}
	ss.Actions = append(ss.Actions, a)

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, ss)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+di)
	}

	// TODO : Atomicity check
	latestRevision := prevRevisionID + 1

	// Instantiating the current state
	a = state.ActionEntry{
		State:     state.StateEnum.Instantiated,
		ContextId: targetCtxId,
		TimeStamp: time.Now(),
		Revision:  latestRevision,
	}
	ss.Actions = append(ss.Actions, a)

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, ss)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+di)
	}

	log.Info("Rollback Completed", log.Fields{"Rollback revisionID": latestRevision})

	return nil
}
