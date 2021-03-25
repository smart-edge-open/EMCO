// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package context

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/resourcestatus"
	. "github.com/open-ness/EMCO/src/rsync/pkg/types"
	pkgerrors "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Check status of AppContext against the event to see if it is valid
func (c *Context) checkStateChange(e RsyncEvent) (StateChange, error) {
	var supported bool = false
	var err error
	var dState, cState appcontext.AppContextStatus
	var event StateChange

	utils := &AppContextUtils{ac: c.ac}
	// Check Stop flag return error no processing desired
	sFlag, err := utils.GetAppContextFlag(StopFlagKey)
	if err != nil {
		return StateChange{}, pkgerrors.Errorf("AppContext Error: %s:", err)
	}
	if sFlag {
		return StateChange{}, pkgerrors.Errorf("Stop flag set for context: %s", c.acID)
	}
	// Check PendingTerminate Flag
	tFlag, err := utils.GetAppContextFlag(PendingTerminateFlagKey)
	if err != nil {
		return StateChange{}, pkgerrors.Errorf("AppContext Error: %s:", err)
	}

	if tFlag && e != TerminateEvent {
		return StateChange{}, pkgerrors.Errorf("Terminate Flag is set, Ignoring event: %s:", e)
	}
	// Update the desired state of the AppContext based on this event
	event, ok := StateChanges[e]
	if !ok {
		return StateChange{}, pkgerrors.Errorf("Invalid Event %s:", e)
	}
	state, err := utils.GetAppContextStatus(CurrentStateKey)
	if err != nil {
		return StateChange{}, err
	}
	for _, s := range event.SState {
		if s == state.Status {
			supported = true
			break
		}
	}
	if !supported {
		return StateChange{}, pkgerrors.Errorf("Invalid Source state %s for the Event %s:", state, e)
	} else {
		dState.Status = event.DState
		cState.Status = event.CState
	}
	// Event is supported. Update Desired state and current state
	err = utils.UpdateAppContextStatus(DesiredStateKey, dState)
	if err != nil {
		return StateChange{}, err
	}
	err = utils.UpdateAppContextStatus(CurrentStateKey, cState)
	if err != nil {
		return StateChange{}, err
	}
	err = utils.UpdateAppContextStatus(StatusKey, cState)
	if err != nil {
		return StateChange{}, err
	}
	return event, nil
}

// UpdateQStatus updates status of an element in the queue
func (c *Context) UpdateQStatus(index int, status string) error {
	qUtils := &AppContextQueueUtils{ac: c.ac}
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if err := qUtils.UpdateStatus(index, status); err != nil {
		return err
	}
	return nil
}

// If terminate event recieved set flag to ignore other events
func (c *Context) terminateContextRoutine() {
	utils := &AppContextUtils{ac: c.ac}
	// Set Terminate Flag to Pending
	if err := utils.UpdateAppContextFlag(PendingTerminateFlagKey, true); err != nil {
		return
	}
	// Make all waiting goroutines to stop waiting
	if c.cancel != nil {
		c.cancel()
	}
}

// Start Main Thread for handling
func (c *Context) startMainThread(a interface{}, con Connector) error {
	acID := fmt.Sprintf("%v", a)
	ac := appcontext.AppContext{}
	_, err := ac.LoadAppContext(acID)
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}
	utils := &AppContextUtils{ac: ac}
	// Read AppContext into CompositeApp structure
	c.ca, err = ReadAppContext(a)
	if err != nil {
		log.Error("Fatal! error reading appContext", log.Fields{"err": err})
		return err
	}
	c.acID = acID
	c.con = con
	c.ac = ac
	// Wait for 2 secs
	c.waitTime = 2
	c.maxRetry = getMaxRetries()
	// Check flags in AppContext to create if they don't exist and add default values
	_, err = utils.GetAppContextStatus(CurrentStateKey)
	// If CurrentStateKey doesn't exist assuming this is the very first event for the appcontext
	if err != nil {
		as := appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Created}
		if err := utils.UpdateAppContextStatus(CurrentStateKey, as); err != nil {
			return err
		}
	}
	_, err = utils.GetAppContextFlag(StopFlagKey)
	// Assume doesn't exist and add
	if err != nil {
		if err := utils.UpdateAppContextFlag(StopFlagKey, false); err != nil {
			return err
		}
	}
	_, err = utils.GetAppContextFlag(PendingTerminateFlagKey)
	// Assume doesn't exist and add
	if err != nil {
		if err := utils.UpdateAppContextFlag(PendingTerminateFlagKey, false); err != nil {
			return err
		}
	}
	_, err = utils.GetAppContextStatus(StatusKey)
	// If CurrentStateKey doesn't exist assuming this is the very first event for the appcontext
	if err != nil {
		as := appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Created}
		if err := utils.UpdateAppContextStatus(StatusKey, as); err != nil {
			return err
		}
	}
	// Read the statusAcID to use with status
	c.statusAcID, err = utils.GetStatusAppContext(StatusAppContextIDKey)
	if err != nil {
		// Use appcontext as status appcontext also
		c.statusAcID = c.acID
		c.sc = c.ac
	} else {
		sc := appcontext.AppContext{}
		_, err = sc.LoadAppContext(c.statusAcID)
		if err != nil {
			log.Error("", log.Fields{"err": err})
			return err
		}
		c.sc = sc
	}
	// Start Routine to handle AppContext
	go c.appContextRoutine()
	c.Running = true
	return nil
}

// Handle AppContext
func (c *Context) appContextRoutine() {
	var lctx context.Context
	var l context.Context
	var lGroup *errgroup.Group
	var lDone context.CancelFunc
	var op RsyncOperation

	utils := &AppContextUtils{ac: c.ac}
	qUtils := &AppContextQueueUtils{ac: c.ac}
	// Create context for the running threads
	ctx, done := context.WithCancel(context.Background())
	gGroup, gctx := errgroup.WithContext(ctx)
	// Stop all running goroutines
	defer done()
	// Start thread to watch for external stop flag
	gGroup.Go(func() error {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				flag, err := utils.GetAppContextFlag(StopFlagKey)
				if err != nil {
					done()
				} else if flag == true {
					log.Info("Forced stop context", log.Fields{})
					// Forced Stop from outside
					done()
				}
			case <-gctx.Done():
				log.Info("Context done", log.Fields{})
				return gctx.Err()
			}
		}
	})
	// Go over all messages
	for {
		// Get first event to process
		c.Lock.Lock()
		index, ele := qUtils.FindFirstPending()
		if index >= 0 {
			c.Lock.Unlock()
			e := ele.Event
			state, err := c.checkStateChange(e)
			// Event is not valid event for the current state of AppContext
			if err != nil {
				log.Error("State Change Error", log.Fields{"error": err})
				if err := c.UpdateQStatus(index, "Skip"); err != nil {
					break
				}
				// Update status with error
				err = utils.UpdateAppContextStatus(StatusKey, appcontext.AppContextStatus{Status: state.ErrState})
				if err != nil {
					break
				}
				// Update Current status with error
				err = utils.UpdateAppContextStatus(CurrentStateKey, appcontext.AppContextStatus{Status: state.ErrState})
				if err != nil {
					break
				}
				// Continue to process more events
				continue
			}
			// Create a derived context
			l, lDone = context.WithCancel(ctx)
			lGroup, lctx = errgroup.WithContext(l)
			c.Lock.Lock()
			c.cancel = lDone
			c.Lock.Unlock()
			switch e {
			case InstantiateEvent:
				op = OpApply
			case TerminateEvent:
				op = OpDelete
			case ReadEvent:
				op = OpRead
			case UpdateEvent:
				// Update AppContext to decide what needs update
				if err := c.updateDeletePhase(ele); err != nil {
					break
				}
				op = OpDelete
				// Enqueue Modify Phase for the AppContext that is being updated to
				go HandleAppContext(ele.UCID, c.acID, UpdateModifyEvent, c.con)
			case UpdateModifyEvent:
				// In Modify Phase find out resources that need to be modified and
				// set skip to be true for those that match
				// This is done to avoid applying resources that have no differences
				if err := c.updateModifyPhase(ele); err != nil {
					break
				}
				op = OpApply
			case AddChildContextEvent:
				log.Error("Not Implemented", log.Fields{"event": e})
				if err := c.UpdateQStatus(index, "Skip"); err != nil {
					break
				}
				continue
			}
			lGroup.Go(func() error {
				return c.run(lctx, lGroup, op)
			})
			// Wait for all subtasks to complete
			log.Info("Wait for all subtasks to complete", log.Fields{})
			if err := lGroup.Wait(); err != nil {
				log.Error("Failed run", log.Fields{"error": err})
				// Mark the event in Queue
				if err := c.UpdateQStatus(index, "Error"); err != nil {
					break
				}
				// Update failed status
				err = utils.UpdateAppContextStatus(StatusKey, appcontext.AppContextStatus{Status: state.ErrState})
				if err != nil {
					break
				}
				// Update Current status with error
				err = utils.UpdateAppContextStatus(CurrentStateKey, appcontext.AppContextStatus{Status: state.ErrState})
				if err != nil {
					break
				}
				continue
			}
			log.Info("Success all subtasks completed", log.Fields{})
			// Mark the event in Queue
			if err := c.UpdateQStatus(index, "Done"); err != nil {
				break
			}

			// Success - Update Status for the AppContext to match the Desired State
			ds, _ := utils.GetAppContextStatus(DesiredStateKey)
			err = utils.UpdateAppContextStatus(StatusKey, ds)
			err = utils.UpdateAppContextStatus(CurrentStateKey, ds)

		} else {
			// Done Processing all elements in queue
			log.Info("Done Processing - no new messages", log.Fields{"context": c.acID})
			// Set the TerminatePending Flag to false before exiting
			_ = utils.UpdateAppContextFlag(PendingTerminateFlagKey, false)
			// release the active contextIDs
			ok, err := DeleteActiveContextRecord(c.acID)
			if !ok {
				log.Info("Deleting activeContextID failed", log.Fields{"context": c.acID, "error": err})
			}
			c.Running = false
			c.Lock.Unlock()
			return
		}
	}
	// Any error in reading/updating appContext is considered
	// fatal and all processing stopped for the AppContext
	// Set running flag to false before exiting
	c.Lock.Lock()
	// release the active contextIDs
	ok, err := DeleteActiveContextRecord(c.acID)
	if !ok {
		log.Info("Deleting activeContextID failed", log.Fields{"context": c.acID, "error": err})
	}
	c.Running = false
	c.Lock.Unlock()
}

// Iterate over the appcontext to mark apps/cluster/resources that doesn't need to be deleted
func (c *Context) updateDeletePhase(e AppContextQueueElement) error {

	// Read Update AppContext into CompositeApp structure
	uca, err := ReadAppContext(e.UCID)
	if err != nil {
		log.Error("Fatal! error reading appContext", log.Fields{"err": err})
		return err
	}
	// Iterate over all the subapps and mark all apps, clusters and resources
	// that shouldn't be deleted
	for _, app := range c.ca.Apps {
		foundApp := FindApp(uca, app.Name)
		// If app not found that will be deleted (skip false)
		if foundApp {
			// Check if any clusters are deleted
			for _, cluster := range app.Clusters {
				foundCluster := FindCluster(uca, app.Name, cluster.Name)
				if foundCluster {
					// Check if any resources are deleted
					var resCnt int = 0
					for _, res := range cluster.Resources {
						foundRes := FindResource(uca, app.Name, cluster.Name, res.Name)
						if foundRes {
							// If resource found in both appContext don't delete it
							res.Skip = true
						} else {
							// Resource found to be deleted
							resCnt++
						}
					}
					// No resources marked for deletion, mark this cluster for not deleting
					if resCnt == 0 {
						cluster.Skip = true
					}
				}
			}
		}
	}
	return nil
}

// Iterate over the appcontext to mark apps/cluster/resources that doesn't need to be Modified
func (c *Context) updateModifyPhase(e AppContextQueueElement) error {
	// Read Update from AppContext into CompositeApp structure
	uca, err := ReadAppContext(e.UCID)
	if err != nil {
		log.Error("Fatal! error reading appContext", log.Fields{"err": err})
		return err
	}
	utils := &AppContextUtils{ac: c.ac}
	// Load update appcontext also
	uac := appcontext.AppContext{}
	_, err = uac.LoadAppContext(e.UCID)
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}
	updateUtils := &AppContextUtils{ac: uac}
	// Iterate over all the subapps and mark all apps, clusters and resources
	// that match exactly and shouldn't be changed
	for _, app := range c.ca.Apps {
		foundApp := FindApp(uca, app.Name)
		if foundApp {
			// Check if any clusters are modified
			for _, cluster := range app.Clusters {
				foundCluster := FindCluster(uca, app.Name, cluster.Name)
				if foundCluster {
					diffRes := false
					// Check if any resources are added or modified
					for _, res := range cluster.Resources {
						foundRes := FindResource(uca, app.Name, cluster.Name, res.Name)
						if foundRes {
							// Read the resource from both AppContext and Compare
							cRes, _, err1 := utils.GetRes(res.Name, app.Name, cluster.Name)
							uRes, _, err2 := updateUtils.GetRes(res.Name, app.Name, cluster.Name)
							if err1 != nil || err2 != nil {
								log.Error("Fatal Error: reading resources", log.Fields{"err1": err1, "err2": err2})
								return err1
							}
							if bytes.Equal(cRes, uRes) {
								res.Skip = true
							} else {
								log.Info("Update Resource Diff found::", log.Fields{"resource": res.Name, "cluster": cluster})
								diffRes = true
							}
						} else {
							// Found a new resource that is added to the cluster
							diffRes = true
						}
					}
					// If no resources diff, skip cluster
					if !diffRes {
						cluster.Skip = true
					}
				}
			}
		}
	}
	return nil
}

// Iterate over the appcontext to apply/delete/read resources
func (c *Context) run(ctx context.Context, g *errgroup.Group, op RsyncOperation) error {
	// Iterate over all the subapps and start go Routines per app
	for _, a := range c.ca.AppOrder {
		app := a
		// If marked to skip then no processing needed
		if c.ca.Apps[app].Skip {
			log.Info("Update Skipping App::", log.Fields{"App": app})
			// Reset bit and skip app
			t := c.ca.Apps[app]
			t.Skip = false
			continue
		}
		// TODO - this code should be removed once support for App Dependency
		// instructions which allow a time based delay before applying are supported.
		if app == "network-chain-intents" && op == OpApply {
			log.Info("Chain App - sleep for 100 seconds::", log.Fields{"App": app})
			time.Sleep(100 * time.Second)
			log.Info("Chain App - done sleeping for 100 seconds::", log.Fields{"App": app})
		}
		g.Go(func() error {
			return c.runApp(ctx, g, op, app)
		})
	}
	return nil
}

func (c *Context) runApp(ctx context.Context, g *errgroup.Group, op RsyncOperation, app string) error {
	// Iterate over all clusters
	for _, cluster := range c.ca.Apps[app].Clusters {
		// If marked to skip then no processing needed
		if cluster.Skip {
			log.Info("Update Skipping Cluster::", log.Fields{"App": app, "cluster": cluster})
			// Reset bit and skip cluster
			cluster.Skip = false
			continue
		}
		cluster := cluster.Name
		err := c.con.StartClusterWatcher(cluster)
		if err != nil {
			log.Error("Error starting Cluster Watcher", log.Fields{
				"error":   err,
				"cluster": cluster,
			})
		}
		g.Go(func() error {
			return c.runCluster(ctx, g, op, app, cluster)
		})
	}
	return nil
}

func (c *Context) runCluster(ctx context.Context, g *errgroup.Group, op RsyncOperation, app, cluster string) error {
	log.Info(" runCluster::", log.Fields{"app": app, "cluster": cluster})
	utils := &AppContextUtils{ac: c.ac}
	namespace, level := utils.GetNamespace()
	cl, err := c.con.GetClientInternal(cluster, level, namespace)
	if err != nil {
		log.Error("Error in creating client", log.Fields{"error": err, "cluster": cluster, "app": app})
		return err
	}

	// Keep retrying for reachability
	for {
		// Wait for cluster to be reachable
		err := c.waitForClusterReady(ctx, cl, app, cluster)
		if err != nil {
			return err
		}
		reachable := true
		// Handle all resources in order
		for i, res := range c.ca.Apps[app].Clusters[cluster].ResOrder {
			// If marked to skip then no processing needed
			if c.ca.Apps[app].Clusters[cluster].Resources[res].Skip {
				// Reset bit and skip resource
				log.Info("Update Skipping Resource::", log.Fields{"App": app, "cluster": cluster, "resource": res})
				r := c.ca.Apps[app].Clusters[cluster].Resources[res]
				r.Skip = false
				continue
			}
			// Dependency here
			breakonError, err := c.handleResource(ctx, g, cl, op, app, cluster, res)
			if err != nil {
				log.Error("Error in resource", log.Fields{"error": err, "cluster": cluster, "resource": res})
				// If failure is due to reachability issues start retrying
				if err = cl.IsReachable(); err != nil {
					reachable = false
					break
				}
				if breakonError {
					// handle status tracking before exiting if at least one resource got handled
					if i > 0 {
						serr := c.handleStatusTracking(ctx, g, cl, op, app, cluster)
						if serr != nil {
							log.Info("Error handling status tracker", log.Fields{"error": serr})
						}
					}
					return err
				}
			}
		}
		// Check if the break from loop due to reachabilty issues
		if reachable != false {
			serr := c.handleStatusTracking(ctx, g, cl, op, app, cluster)
			if serr != nil {
				log.Info("Error handling status tracker", log.Fields{"error": serr})
			}
			// Done processing cluster without errors
			return nil
		}
	}
}

func (c *Context) handleResource(ctx context.Context, g *errgroup.Group, cl ClientProvider, op RsyncOperation, app, cluster, res string) (bool, error) {
	log.Info(" handleResource::", log.Fields{"app": app, "cluster": cluster, "res": res})

	switch op {
	case OpApply:
		// Get resource dependency here
		err := c.instantiateResource(cl, res, app, cluster)
		if err != nil {
			// return true for breakon error
			return true, err
		}
	case OpDelete:
		err := c.terminateResource(cl, res, app, cluster)
		if err != nil {
			// return false for breakon error
			return false, err
		}
	case OpRead:
		err := c.readResource(cl, res, app, cluster)
		if err != nil {
			// return false for breakon error
			return false, err
		}
	}
	// return false for breakon error
	return false, nil
}

func (c *Context) handleStatusTracking(ctx context.Context, g *errgroup.Group, cl ClientProvider, op RsyncOperation, app, cluster string) error {
	log.Info(" handleStatusTracking::", log.Fields{"app": app, "cluster": cluster})
	label := c.acID + "-" + app
	switch op {
	case OpApply:
		err := c.addStatusTracker(cl, app, cluster, label)
		if err != nil {
			return err
		}
	case OpDelete:
		err := c.deleteStatusTracker(cl, app, cluster, label)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) instantiateResource(cl ClientProvider, name, app, cluster string) error {
	utils := &AppContextUtils{ac: c.ac}
	res, _, err := utils.GetRes(name, app, cluster)
	if err != nil {
		c.updateResourceStatus(name, app, cluster,
			resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		return err
	}
	// Add the label based on the Status Appcontext ID
	label := c.statusAcID + "-" + app
	log.Info("Tag Label:", log.Fields{"label": label})
	b, err := cl.TagResource(res, label)
	if err != nil {
		return err
	}
	if err := cl.Apply(b); err != nil {
		c.updateResourceStatus(name, app, cluster,
			resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		log.Error("Failed to apply res", log.Fields{
			"error":    err,
			"resource": name,
		})
		return err
	}
	c.updateResourceStatus(name, app, cluster,
		resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Applied})
	log.Info("Installed::", log.Fields{
		"cluster":  cluster,
		"resource": name,
	})
	// Currently only subresource supported is approval
	subres, _, err := utils.GetSubResApprove(name, app, cluster)
	if err == nil {
		result := strings.Split(name, "+")
		if result[0] == "" {
			return pkgerrors.Errorf("Resource name is nil %s:", name)
		}
		log.Info("Approval Subresource::", log.Fields{"cluster": cluster, "resource": result[0], "approval": string(subres)})
		err = cl.Approve(result[0], subres)
		return err
	}
	return nil
}

func (c *Context) terminateResource(cl ClientProvider, name, app, cluster string) error {

	utils := &AppContextUtils{ac: c.ac}
	res, sh, err := utils.GetRes(name, app, cluster)
	if err != nil {
		if sh != nil {
			c.updateResourceStatus(name, app, cluster,
				resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		}
		return err
	}
	if err := cl.Delete(res); err != nil {
		c.updateResourceStatus(name, app, cluster,
			resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		log.Error("Failed to delete res", log.Fields{
			"error":    err,
			"resource": name,
		})
		return err
	}
	c.updateResourceStatus(name, app, cluster,
		resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Deleted})
	log.Info("Deleted::", log.Fields{
		"cluster":  cluster,
		"resource": name,
	})
	return nil
}

func (c *Context) readResource(cl ClientProvider, name, app, cluster string) error {

	utils := &AppContextUtils{ac: c.ac}
	res, _, err := utils.GetRes(name, app, cluster)
	if err != nil {
		c.updateResourceStatus(name, app, cluster,
			resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		return err
	}
	namespace, _ := utils.GetNamespace()

	// Get the resource from the cluster
	b, err := cl.Get(res, namespace)
	if err != nil {
		c.updateResourceStatus(name, app, cluster,
			resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		log.Error("Failed to read res", log.Fields{
			"error":    err,
			"resource": name,
		})
		return err
	}
	// Store result back in AppContext
	utils.PutRes(name, app, cluster, b)
	c.updateResourceStatus(name, app, cluster,
		resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Applied})
	log.Info("Applied::", log.Fields{
		"cluster":  cluster,
		"resource": name,
	})
	return nil
}

func (c *Context) waitForClusterReady(ctx context.Context, cl ClientProvider, app string, cluster string) error {
	utils := &AppContextUtils{ac: c.ac}
	// Check if reachable
	if err := cl.IsReachable(); err == nil {
		utils.SetClusterReadyStatus(app, cluster, appcontext.ClusterReadyStatusEnum.Available)
		return nil
	}
	utils.SetClusterReadyStatus(app, cluster, appcontext.ClusterReadyStatusEnum.Retrying)
	timedOut := false
	retryCnt := 0
	forceDone := false
Loop:
	for {
		select {
		// Wait for wait time before checking cluster ready
		case <-time.After(time.Duration(c.waitTime) * time.Second):
			// Context is canceled
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// If cluster is reachable then done
			if err := cl.IsReachable(); err == nil {
				utils.SetClusterReadyStatus(app, cluster, appcontext.ClusterReadyStatusEnum.Available)
				return nil
			}
			log.Info("Cluster is not reachable - keep trying::", log.Fields{"cluster": cluster, "retry count": retryCnt})
			retryCnt++
			if c.maxRetry >= 0 && retryCnt > c.maxRetry {
				timedOut = true
				break Loop
			}
			break
		// Check if the context is canceled
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if timedOut {
		return pkgerrors.Errorf("Retries exceeded max: " + cluster)
	}
	if forceDone {
		return pkgerrors.Errorf("Termination of rsync cluster retry: " + cluster)
	}
	return nil
}

func (c *Context) addStatusTracker(cl ClientProvider, app string, cluster string, label string) error {

	b, err := c.con.GetStatusCR(label)
	if err != nil {
		log.Error("Failed to get status CR for installing", log.Fields{
			"error": err,
			"label": label,
		})
		return err
	}
	if err = cl.Apply(b); err != nil {
		log.Error("Failed to apply status tracker", log.Fields{
			"error":   err,
			"cluster": cluster,
			"app":     app,
			"label":   label,
		})
		return err
	}
	log.Info("Status tracker installed::", log.Fields{
		"cluster": cluster,
		"app":     app,
		"label":   label,
	})
	return nil
}

func (c *Context) deleteStatusTracker(cl ClientProvider, app string, cluster string, label string) error {
	b, err := c.con.GetStatusCR(label)
	if err != nil {
		log.Error("Failed to get status CR for deleting", log.Fields{
			"error": err,
			"label": label,
		})
		return err
	}
	if err = cl.Delete(b); err != nil {
		log.Error("Failed to delete res", log.Fields{
			"error": err,
			"app":   app,
			"label": label,
		})
		return err
	}
	log.Info("Status tracker deleted::", log.Fields{
		"cluster": cluster,
		"app":     app,
		"label":   label,
	})
	return nil
}

func (c *Context) updateResourceStatus(name, app, cluster string, status interface{}) {
	// Use utils with status appContext
	utils := &AppContextUtils{ac: c.sc}
	_ = utils.AddResourceStatus(name, app, cluster, status)
	// Treating status errors as non fatal
}
