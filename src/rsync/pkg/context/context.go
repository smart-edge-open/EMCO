// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package context

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/config"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/rsync/pkg/connector"
	. "github.com/open-ness/EMCO/src/rsync/pkg/types"
)

// Context is Per AppContext struct
type Context struct {
	Lock    *sync.Mutex
	Running bool
	Channel chan RsyncEvent
	// AppContext ID
	acID string
	// AppContext handle
	ac appcontext.AppContext
	// Status AppContext ID
	statusAcID string
	// Status AppContext handle
	sc appcontext.AppContext
	// Connector interface
	con Connector
	// Function to cancel running threads on terminate
	cancel context.CancelFunc
	// APP and Res Dependency
	dependency []ResourceDependency
	// Max Retries for cluster reachability
	maxRetry int
	// wait time (seconds) between trying again for cluster reachability
	waitTime int
	// Structure to hold CompositeApp Information
	ca CompositeApp
}

// AppContextData struct
type AppContextData struct {
	Data map[string]*Context
	sync.Mutex
}

var appContextData = AppContextData{
	Data: map[string]*Context{},
}
// HandleAppContext adds event to queue and starts main thread
func HandleAppContext(a interface{}, ucid interface{}, e RsyncEvent, con Connector) error {

	acID := fmt.Sprintf("%v", a)
	// Create AppContext data if not already created
	_, c := CreateAppContextData(acID)
	// Add event to queue
	err := c.EnqueueToAppContext(a, ucid, e)
	if err != nil {
		return err
	}
	// If main thread is not running start it
	// Acquire Mutex
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if c.Running {
		if e == TerminateEvent {
			c.terminateContextRoutine()
		}
	} else  {
		err = c.startMainThread(a, con)
	}
	return err
}
// EnqueueToAppContext adds the event to the appContext Queue
func (c *Context)EnqueueToAppContext(a interface{}, ucid interface{}, e RsyncEvent) error {
	acID := fmt.Sprintf("%v", a)
	ac := appcontext.AppContext{}
	_, err := ac.LoadAppContext(acID)
	if err != nil {
		logutils.Error("", logutils.Fields{"err": err})
		return err
	}
	qUtils := AppContextQueueUtils{ac: ac}
	var elem AppContextQueueElement
	// Store UpdateID
	if ucid != nil {
		ucID := fmt.Sprintf("%v", ucid)
		elem = AppContextQueueElement{Event: e, Status: "Pending", UCID: ucID}
	} else {
		elem = AppContextQueueElement{Event: e, Status: "Pending"}
	}
	// Acquire Mutex before adding to queue
	c.Lock.Lock()
	// Push the appContext to ActiveContext space of etcD
	ok, err := RecordActiveContext(acID)
	if !ok {
		logutils.Info("Already in active context", logutils.Fields{"AppContextID":acID, "err": err})
	}
	// Enqueue event
	qUtils.Enqueue(elem)
	c.Lock.Unlock()
	return nil
}
// RestartAppContext called in Restart scenario to handle an AppContext
func RestartAppContext(a interface{}, con Connector) error {
	var err error
	acID := fmt.Sprintf("%v", a)
	// Create AppContext data if not already created
	_, c := CreateAppContextData(acID)

	// Acquire Mutex
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if c.Running == false {
		err = c.startMainThread(a, con)
	}
	return err
}
// Create per AppContext thread data
func CreateAppContextData(key string) (bool, *Context) {
	appContextData.Lock()
	defer appContextData.Unlock()
	_, ok := appContextData.Data[key]
	// Create if doesn't exist
	if !ok {
		appContextData.Data[key] = &Context{}
		appContextData.Data[key].Lock = &sync.Mutex{}
		appContextData.Data[key].Running = false
		// Created appContext data (return true)
		return true, appContextData.Data[key]
	}
	// Didn't create appContext data (return false)
	return false, appContextData.Data[key]
}
// Delete per AppContext thread data
func DeleteAppContextData(key string) (error) {
	appContextData.Lock()
	defer appContextData.Unlock()
	_, ok := appContextData.Data[key]
	if ok {
		delete(appContextData.Data, key)
	}
	return nil
}
// Read Max retries from configuration
func getMaxRetries() int {
	s := config.GetConfiguration().MaxRetries
	if s == "" {
		return -1
	}
	maxRetries, err := strconv.Atoi(s)
	if err != nil {
		return -1
	} else {
		if maxRetries < 0 {
			return -1
		}
	}
	return maxRetries
}
// CompositeAppContext represents composite app
type CompositeAppContext struct {
	cid interface{}
}
// InstantiateComApp Instantiatep Aps in Composite App
func (instca *CompositeAppContext) InstantiateComApp(cid interface{}) error {
	instca.cid = cid
	con := connector.Connection{}
	con.Init(instca.cid)
	return HandleAppContext(instca.cid, nil, InstantiateEvent, &con)
}
// TerminateComApp Terminates Apps in Composite App
func (instca *CompositeAppContext) TerminateComApp(cid interface{}) error {
	instca.cid = cid
	con := connector.Connection{}
	con.Init(instca.cid)
	return HandleAppContext(instca.cid, nil, TerminateEvent, &con)
}

// UpdateComApp Updates Apps in Composite App
func (instca *CompositeAppContext) UpdateComApp(cid interface{}, ucid interface{}) error {
	instca.cid = cid
	con := connector.Connection{}
	con.Init(instca.cid)
	return HandleAppContext(instca.cid, ucid, UpdateEvent, &con)
}

// ReadComApp Reads resources in AppContext
func (instca *CompositeAppContext) ReadComApp(cid interface{}) error {
	instca.cid = cid
	con := connector.Connection{}
	con.Init(instca.cid)
	return HandleAppContext(instca.cid, nil, ReadEvent, &con)
}
