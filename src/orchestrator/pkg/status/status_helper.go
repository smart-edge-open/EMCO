// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package status

import (
	"encoding/json"
	"fmt"
	"strings"

	rb "github.com/open-ness/EMCO/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"github.com/open-ness/EMCO/src/monitor/pkg/generated/clientset/versioned/scheme"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/utils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/resourcestatus"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/state"
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// decodeYAML reads a YAMl []byte to extract the Kubernetes object definition
func decodeYAML(y []byte, into runtime.Object) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(y, nil, into)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}

func getUnstruct(y []byte) (unstructured.Unstructured, error) {
	//Decode the yaml file to create a runtime.Object
	unstruct := unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := decodeYAML(y, &unstruct)
	if err != nil {
		log.Info(":: Error decoding YAML ::", log.Fields{"object": y, "error": err})
		return unstructured.Unstructured{}, pkgerrors.Wrap(err, "Decode object error")
	}

	return unstruct, nil
}

// GetClusterResources takes in a ResourceBundleStatus CR and resturns a list of ResourceStatus elments
func GetClusterResources(rbData rb.ResourceBundleStatus, qOutput string, fResources []string,
	resourceList *[]ResourceStatus, cnts map[string]int) (int, error) {

	count := 0

	for _, p := range rbData.PodStatuses {
		if !keepResource(p.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = p.Name
		r.Gvk = (&p.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = p
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, s := range rbData.ServiceStatuses {
		if !keepResource(s.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = s.Name
		r.Gvk = (&s.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = s
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, d := range rbData.DeploymentStatuses {
		if !keepResource(d.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = d.Name
		r.Gvk = (&d.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = d
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, c := range rbData.ConfigMapStatuses {
		if !keepResource(c.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = c.Name
		r.Gvk = (&c.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = c
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, s := range rbData.SecretStatuses {
		if !keepResource(s.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = s.Name
		r.Gvk = (&s.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = s
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, d := range rbData.DaemonSetStatuses {
		if !keepResource(d.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = d.Name
		r.Gvk = (&d.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = d
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, i := range rbData.IngressStatuses {
		if !keepResource(i.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = i.Name
		r.Gvk = (&i.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = i
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, j := range rbData.JobStatuses {
		if !keepResource(j.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = j.Name
		r.Gvk = (&j.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = j
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, s := range rbData.StatefulSetStatuses {
		if !keepResource(s.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = s.Name
		r.Gvk = (&s.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = s
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	return count, nil
}

// isResourceHandle takes a cluster handle and determines if the other handle parameter is a resource handle for this cluster
// handle.  It does this by verifying that the cluster handle is a prefix of the handle and that the remainder of the handle
// is a value that matches to a resource format:  "resource/<name>+<type>/"
// Example cluster handle:
// /context/6385596659306465421/app/network-intents/cluster/vfw-cluster-provider+edge01/
// Example resource handle:
// /context/6385596659306465421/app/network-intents/cluster/vfw-cluster-provider+edge01/resource/emco-private-net+ProviderNetwork/
func isResourceHandle(ch, h interface{}) bool {
	clusterHandle := fmt.Sprintf("%v", ch)
	handle := fmt.Sprintf("%v", h)
	diff := strings.Split(handle, clusterHandle)

	if len(diff) != 2 && diff[0] != "" {
		return false
	}

	parts := strings.Split(diff[1], "/")

	if len(parts) == 3 &&
		parts[0] == "resource" &&
		len(strings.Split(parts[1], "+")) == 2 &&
		parts[2] == "" {
		return true
	} else {
		return false
	}
}

// keepResource keeps a resource if the filter list is empty or if the resource is part of the list
func keepResource(r string, rList []string) bool {
	if len(rList) == 0 {
		return true
	}
	for _, res := range rList {
		if r == res {
			return true
		}
	}
	return false
}

// GetAppContextResources collects the resource status of all resources in an AppContext subject to the filter parameters
func GetAppContextResources(ac appcontext.AppContext, ch interface{}, qOutput string, fResources []string, resourceList *[]ResourceStatus, statusCnts map[string]int) (int, error) {
	count := 0

	// Get all Resources for the Cluster
	hs, err := ac.GetAllHandles(ch)
	if err != nil {
		log.Info(":: Error getting all handles ::", log.Fields{"handles": ch, "error": err})
		return 0, err
	}

	for _, h := range hs {
		// skip any handles that are not resource handles
		if !isResourceHandle(ch, h) {
			continue
		}

		// Get Resource from AppContext
		res, err := ac.GetValue(h)
		if err != nil {
			log.Info(":: Error getting resource value ::", log.Fields{"Handle": h})
			continue
		}

		// Get Resource Status from AppContext
		// Default to "Pending" if this key does not yet exist (or any other error occurs)
		rstatus := resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Pending}
		sh, err := ac.GetLevelHandle(h, "status")
		if err == nil {
			s, err := ac.GetValue(sh)
			if err == nil {
				js, err := json.Marshal(s)
				if err == nil {
					json.Unmarshal(js, &rstatus)
				}
			}
		}

		// Get the unstructured object
		unstruct, err := getUnstruct([]byte(res.(string)))
		if err != nil {
			log.Info(":: Error getting GVK ::", log.Fields{"Resource": res, "error": err})
			continue
		}
		if !keepResource(unstruct.GetName(), fResources) {
			continue
		}

		// Make and fill out a ResourceStatus structure
		r := ResourceStatus{}
		r.Gvk = unstruct.GroupVersionKind()
		r.Name = unstruct.GetName()
		if qOutput == "detail" {
			r.Detail = unstruct.Object
		}
		r.RsyncStatus = fmt.Sprintf("%v", rstatus.Status)
		*resourceList = append(*resourceList, r)
		cnt := statusCnts[rstatus.Status]
		statusCnts[rstatus.Status] = cnt + 1
		count++
	}

	return count, nil
}

// getListOfApps gets the list of apps from the app context
func getListOfApps(appContextId string) []string {
	var ac appcontext.AppContext
	apps := make([]string, 0)

	_, err := ac.LoadAppContext(appContextId)
	if err != nil {
		log.Info(":: Error loading the app context::", log.Fields{"appContextId": appContextId, "error": err})
		return apps
	}

	ch := "/context/" + appContextId + "/"

	// Get all handles
	hs, err := ac.GetAllHandles(ch)
	if err != nil {
		log.Info(":: Error getting all handles ::", log.Fields{"handles": ch, "error": err})
		return apps
	}

	for _, h := range hs {
		contextHandle := fmt.Sprintf("%v", ch)
		handle := fmt.Sprintf("%v", h)
		diff := strings.Split(handle, contextHandle)

		if len(diff) != 2 && diff[0] != "" {
			continue
		}

		parts := strings.Split(diff[1], "/")

		if len(parts) == 3 && parts[0] == "app" {
			apps = append(apps, parts[1])
		}
	}

	return apps
}

// types of status queries
const clusterStatus = "clusterStatus"
const deploymentIntentGroupStatus = "digStatus"

// PrepareClusterStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func PrepareClusterStatusResult(stateInfo state.StateInfo, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (ClusterStatusResult, error) {
	status, err := prepareStatusResult(clusterStatus, stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
	if err != nil {
		return ClusterStatusResult{}, err
	} else {
		rval := ClusterStatusResult{
			Name:          status.Name,
			State:         status.State,
			Status:        status.Status,
			RsyncStatus:   status.RsyncStatus,
			ClusterStatus: status.ClusterStatus,
		}
		if len(status.Apps) > 0 && len(status.Apps[0].Clusters) > 0 {
			rval.Cluster = status.Apps[0].Clusters[0]
		}
		return rval, nil
	}
}

// PrepareStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func PrepareStatusResult(stateInfo state.StateInfo, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (StatusResult, error) {
	return prepareStatusResult(deploymentIntentGroupStatus, stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
}

// covenience fn that ignores the index returned by GetSliceContains
func isNameInList(name string, namesList []string) bool {
	_, ok := utils.GetSliceContains(namesList, name)
	return ok
}

// prepareStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func prepareStatusResult(statusType string, stateInfo state.StateInfo, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (StatusResult, error) {

	statusResult := StatusResult{}

	statusResult.Apps = make([]AppStatus, 0)
	statusResult.State = stateInfo

	var currentCtxId, statusCtxId string
	if qInstance != "" {
		// ToDo: Locate the context id that is current. Different in
		// case update or rollback scenario
		currentCtxId = qInstance
		statusCtxId = qInstance
	} else {
		currentCtxId = state.GetLastContextIdFromStateInfo(stateInfo)
		// For App and cluster level status use status AppContext
		statusCtxId = state.GetStatusContextIdFromStateInfo(stateInfo)
	}

	// If currentCtxId is still an empty string, an AppContext has not yet been
	// created for this resource.  Just return the statusResult with the stateInfo.
	if currentCtxId == "" {
		return statusResult, nil
	}

	ac, err := state.GetAppContextFromId(currentCtxId)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext for status query not found")
	}

	// get the appcontext status value
	h, err := ac.GetCompositeAppHandle()
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext handle not found")
	}
	sh, err := ac.GetLevelHandle(h, "status")
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext status handle not found")
	}
	statusVal, err := ac.GetValue(sh)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext status value not found")
	}
	acStatus := appcontext.AppContextStatus{}
	js, err := json.Marshal(statusVal)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "Invalid AppContext status value format")
	}
	err = json.Unmarshal(js, &acStatus)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "Invalid AppContext status value format")
	}

	statusResult.Status = acStatus.Status
	// For App and cluster level status use status AppContext
	ac, err = state.GetAppContextFromId(statusCtxId)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext for status query not found")
	}
	// Get the composite app meta
	caMeta, err := ac.GetCompositeAppMeta()
	if statusType != clusterStatus {
		if err != nil {
			return StatusResult{}, pkgerrors.Wrap(err, "Error getting CompositeAppMeta")
		}
		if len(caMeta.ChildContextIDs) > 0 {
			// Add the child context IDs to status result
			statusResult.ChildContextIDs = caMeta.ChildContextIDs
		}
	}

	rsyncStatusCnts := make(map[string]int)
	clusterStatusCnts := make(map[string]int)

	// Get the list of apps from the app context
	apps := getListOfApps(currentCtxId)

	// If filter-apps list is provided, ensure that every app to be
	// filtered is part of this composite app
	for _, fApp := range fApps {
		if !isNameInList(fApp, apps) {
			return StatusResult{},
				fmt.Errorf("Filter app %s not in list of apps for composite app %s",
					fApp, caMeta.CompositeApp)
		}
	}

	// Loop through each app and get the status data for each cluster in the app
	for _, app := range apps {
		appCount := 0
		if len(fApps) > 0 && !isNameInList(app, fApps) {
			continue
		}
		// Get the clusters in the appcontext for this app
		clusters, err := ac.GetClusterNames(app)
		if err != nil {
			continue
		}
		var appStatus AppStatus
		appStatus.Name = app
		appStatus.Clusters = make([]ClusterStatus, 0)

		for _, cluster := range clusters {
			clusterCount := 0
			if len(fClusters) > 0 && !isNameInList(cluster, fClusters) {
				continue
			}

			var clusterStatus ClusterStatus
			pc := strings.Split(cluster, "+")
			clusterStatus.ClusterProvider = pc[0]
			clusterStatus.Cluster = pc[1]
			clusterStatus.ReadyStatus = getClusterReadyStatus(ac, app, cluster)

			if qType == "cluster" {
				csh, err := ac.GetClusterStatusHandle(app, cluster)
				if err != nil {
					log.Info(":: No cluster status handle for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
				clusterRbValue, err := ac.GetValue(csh)
				if err != nil {
					log.Info(":: No cluster status value for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
				var rbValue rb.ResourceBundleStatus
				err = json.Unmarshal([]byte(clusterRbValue.(string)), &rbValue)
				if err != nil {
					log.Info(":: Error unmarshaling cluster status value for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}

				clusterStatus.Resources = make([]ResourceStatus, 0)
				cnt, err := GetClusterResources(rbValue, qOutput, fResources, &clusterStatus.Resources, clusterStatusCnts)
				if err != nil {
					log.Info(":: Error gathering cluster resources for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
				appCount += cnt
				clusterCount += cnt
			} else if qType == "rsync" {
				ch, err := ac.GetClusterHandle(app, cluster)
				if err != nil {
					log.Info(":: No handle for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}

				/* code to get status for resources from AppContext */
				clusterStatus.Resources = make([]ResourceStatus, 0)
				cnt, err := GetAppContextResources(ac, ch, qOutput, fResources, &clusterStatus.Resources, rsyncStatusCnts)
				if err != nil {
					log.Info(":: Error gathering appcontext resources for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
				appCount += cnt
				clusterCount += cnt
			} else {
				log.Info(":: Invalid status type ::", log.Fields{"Status Type": qType})
				continue
			}

			if clusterCount > 0 {
				appStatus.Clusters = append(appStatus.Clusters, clusterStatus)
			}
		}
		if appCount > 0 && qOutput != "summary" {
			statusResult.Apps = append(statusResult.Apps, appStatus)
		}
	}
	statusResult.RsyncStatus = rsyncStatusCnts
	statusResult.ClusterStatus = clusterStatusCnts

	return statusResult, nil
}

// PrepareAppsListStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func PrepareAppsListStatusResult(stateInfo state.StateInfo, qInstance string) (AppsListResult, error) {
	statusResult := AppsListResult{}

	var currentCtxId string
	if qInstance != "" {
		currentCtxId = qInstance
	} else {
		currentCtxId = state.GetStatusContextIdFromStateInfo(stateInfo)
	}

	// If currentCtxId is still an empty string, an AppContext has not yet been
	// created for this resource.  So, no Apps list can be returned from the AppContext.
	if currentCtxId == "" {
		statusResult.Apps = make([]string, 0)
		return statusResult, nil
	}

	// Get the list of apps from the app context
	statusResult.Apps = getListOfApps(currentCtxId)

	return statusResult, nil
}

// prepareStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func PrepareClustersByAppStatusResult(stateInfo state.StateInfo, qInstance string, fApps []string) (ClustersByAppResult, error) {
	statusResult := ClustersByAppResult{}

	statusResult.ClustersByApp = make([]ClustersByAppEntry, 0)

	var currentCtxId string
	if qInstance != "" {
		currentCtxId = qInstance
	} else {
		currentCtxId = state.GetStatusContextIdFromStateInfo(stateInfo)
	}

	// If currentCtxId is still an empty string, an AppContext has not yet been
	// created for this resource.  Just return the statusResult with the stateInfo.
	if currentCtxId == "" {
		return statusResult, nil
	}

	ac, err := state.GetAppContextFromId(currentCtxId)
	if err != nil {
		return ClustersByAppResult{}, pkgerrors.Wrap(err, "AppContext for status query not found")
	}

	// Get the list of apps from the app context
	apps := getListOfApps(currentCtxId)

	// Loop through each app and get the clusters for the app
	for _, app := range apps {
		// apply app filter if provided
		if len(fApps) > 0 {
			found := false
			for _, a := range fApps {
				if a == app {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		// add app to output structure
		entry := ClustersByAppEntry{
			App: app,
		}

		// Get the clusters in the appcontext for this app
		entry.Clusters = make([]ClusterEntry, 0)
		clusters, err := ac.GetClusterNames(app)
		if err != nil {
		} else {
			for _, cl := range clusters {
				pc := strings.Split(cl, "+")
				entry.Clusters = append(entry.Clusters, ClusterEntry{
					ClusterProvider: pc[0],
					Cluster:         pc[1],
				})
			}
		}

		statusResult.ClustersByApp = append(statusResult.ClustersByApp, entry)
	}

	return statusResult, nil
}

// PrepareResourcesByAppStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the ResourcesByAppStatusResult structure appropriately from information in the AppContext
func PrepareResourcesByAppStatusResult(stateInfo state.StateInfo, qInstance, qType string, fApps, fClusters []string) (ResourcesByAppResult, error) {
	statusResult := ResourcesByAppResult{}

	statusResult.ResourcesByApp = make([]ResourcesByAppEntry, 0)

	var currentCtxId string
	if qInstance != "" {
		currentCtxId = qInstance
	} else {
		currentCtxId = state.GetStatusContextIdFromStateInfo(stateInfo)
	}

	// If currentCtxId is still an empty string, an AppContext has not yet been
	// created for this resource.  Just return the statusResult with the stateInfo.
	if currentCtxId == "" {
		return statusResult, nil
	}

	ac, err := state.GetAppContextFromId(currentCtxId)
	if err != nil {
		return ResourcesByAppResult{}, pkgerrors.Wrap(err, "AppContext for status query not found")
	}

	// Get the list of apps from the app context
	apps := getListOfApps(currentCtxId)

	// Loop through each app and get the status data for each cluster in the app
	for _, app := range apps {
		if len(fApps) > 0 {
			found := false
			for _, a := range fApps {
				if a == app {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Get the clusters in the appcontext for this app
		clusters, err := ac.GetClusterNames(app)
		if err != nil {
			continue
		}
		var appStatus AppStatus
		appStatus.Name = app
		appStatus.Clusters = make([]ClusterStatus, 0)

		doneOneCluster := false
		for _, cluster := range clusters {

			// if query type is "cluster", then apply the
			// cluster filter
			if qType == "cluster" && len(fClusters) > 0 {
				found := false
				for _, c := range fClusters {
					if c == cluster {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// if query type is not "cluster", then return
			// after collecting the resources for the first cluster.
			// (they are assumed to be the same for all clusters)
			if qType != "cluster" && doneOneCluster {
				break
			}

			rsyncStatusCnts := make(map[string]int)
			clusterStatusCnts := make(map[string]int)

			resourcesByAppEntry := ResourcesByAppEntry{
				App: app,
			}
			resourcesByAppEntry.Resources = make([]ResourceEntry, 0)

			if qType == "cluster" {
				pc := strings.Split(cluster, "+")
				resourcesByAppEntry.ClusterProvider = pc[0]
				resourcesByAppEntry.Cluster = pc[1]

				csh, err := ac.GetClusterStatusHandle(app, cluster)
				if err != nil {
					log.Info(":: No cluster status handle for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
				clusterRbValue, err := ac.GetValue(csh)
				if err != nil {
					log.Info(":: No cluster status value for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
				var rbValue rb.ResourceBundleStatus
				err = json.Unmarshal([]byte(clusterRbValue.(string)), &rbValue)
				if err != nil {
					log.Info(":: Error unmarshaling cluster status value for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}

				resources := make([]ResourceStatus, 0)
				_, err = GetClusterResources(rbValue, "all", make([]string, 0), &resources, clusterStatusCnts)
				if err != nil {
					log.Info(":: Error gathering cluster resources for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}

				for _, r := range resources {
					resourcesByAppEntry.Resources = append(resourcesByAppEntry.Resources, ResourceEntry{
						Name: r.Name,
						Gvk:  r.Gvk,
					})
				}
				doneOneCluster = true
			} else if qType == "rsync" {
				ch, err := ac.GetClusterHandle(app, cluster)
				if err != nil {
					log.Info(":: No handle for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}

				/* code to get status for resources from AppContext */
				resources := make([]ResourceStatus, 0)
				_, err = GetAppContextResources(ac, ch, "all", make([]string, 0), &resources, rsyncStatusCnts)
				if err != nil {
					log.Info(":: Error gathering appcontext resources for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}

				for _, r := range resources {
					resourcesByAppEntry.Resources = append(resourcesByAppEntry.Resources, ResourceEntry{
						Name: r.Name,
						Gvk:  r.Gvk,
					})
				}
				doneOneCluster = true
			} else {
				log.Info(":: Invalid status type ::", log.Fields{"Status Type": qType})
				continue
			}
			statusResult.ResourcesByApp = append(statusResult.ResourcesByApp, resourcesByAppEntry)
		}
	}

	return statusResult, nil
}

func getClusterReadyStatus(ac appcontext.AppContext, app, cluster string) string {
	ch, err := ac.GetClusterHandle(app, cluster)
	if err != nil {
		return string(appcontext.ClusterReadyStatusEnum.Unknown)
	}
	rsh, nil := ac.GetLevelHandle(ch, "readystatus")
	if rsh != nil {
		status, err := ac.GetValue(rsh)
		if err != nil {
			return string(appcontext.ClusterReadyStatusEnum.Unknown)
		}
		return status.(string)
	}

	return string(appcontext.ClusterReadyStatusEnum.Unknown)
}
