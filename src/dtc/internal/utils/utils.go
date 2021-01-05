// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package utils

import (
	"encoding/json"

	pkgerrors "github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	rb "github.com/open-ness/EMCO/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

const rsyncName = "rsync"

// CheckDeploymentStatus will check the deployment resource status of the app
func CheckDeploymentStatus(appContextID string, serverApp string) (bool, error) {

	deploymentState := false

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextID)
	if err != nil {
		log.Error("Error loading AppContext", log.Fields{
			"error": err,
		})
		return deploymentState, pkgerrors.Wrapf(err, "Error getting AppContext with Id: %v", appContextID)
	}

	// Get the clusters in the appcontext for this app
	clusters, err := ac.GetClusterNames(serverApp)
	if err != nil {
		log.Error("Unable to get the cluster names",
			log.Fields{"AppName": serverApp, "Error": err})
		return deploymentState, pkgerrors.Wrapf(err, "Unable to get the cluster names for the app - %v", serverApp)
	}

	for _, cluster := range clusters {

		rbValue, err := GetClusterResources(appContextID, serverApp, cluster)
		if err != nil {
			log.Error("Unable to get the cluster resources",
				log.Fields{"Cluster": cluster, "AppName": serverApp, "Error": err})
			deploymentState = false
			continue
		}

		// Get the parent composite app meta
		m, err := ac.GetCompositeAppMeta()
		if err != nil {
			log.Error("Error getting CompositeAppMeta",
				log.Fields{"Cluster": cluster, "AppName": serverApp, "Error": err})
			deploymentState = false
			continue
		}

		// Append the deployment intent group release version
		deploymentName := m.Release + "-" + serverApp

		deploymentState, err = getClusterDeploymentStatus(rbValue, deploymentName)
		if err != nil {
			log.Error("Error gathering cluster deployment status",
				log.Fields{"Cluster": cluster, "AppName": serverApp, "Error": err})
			deploymentState = false
			continue
		}
	}

	return deploymentState, nil
}

// GetClusterResources will retrieve the cluster resources
func GetClusterResources(appContextID string, app string, cluster string) (*rb.ResourceBundleStatus, error) {

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextID)
	if err != nil {
		log.Error("Error loading AppContext", log.Fields{
			"error": err,
		})
		return nil, err
	}

	csh, err := ac.GetClusterStatusHandle(app, cluster)
	if err != nil {
		log.Error("No cluster status handle for cluster, app",
			log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
		return nil, err
	}
	clusterRbValue, err := ac.GetValue(csh)
	if err != nil {
		log.Error("No cluster status value for cluster, app",
			log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
		return nil, err
	}
	var rbValue rb.ResourceBundleStatus
	err = json.Unmarshal([]byte(clusterRbValue.(string)), &rbValue)
	if err != nil {
		log.Error("Error unmarshaling cluster status value for cluster, app",
			log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
		return nil, err
	}

	return &rbValue, nil
}

// getClusterDeploymentStatus takes in a ResourceBundleStatus CR and returns the status of the deployment resource
func getClusterDeploymentStatus(rbData *rb.ResourceBundleStatus, deploymentName string) (bool, error) {

	deploymentStatus := false
	for _, d := range rbData.DeploymentStatuses {
		if !CompareResource(d.Name, deploymentName) {
			continue
		}

		for _, condition := range d.Status.Conditions {
			if (condition.Type == appsv1.DeploymentAvailable || condition.Type == appsv1.DeploymentProgressing) &&
				condition.Status == corev1.ConditionTrue {
				// If the deployment is in active/ready state then continue to check other deployments
				deploymentStatus = true
				break
			} else {
				// deployment is not in active/ready state
				deploymentStatus = false
				return deploymentStatus, nil
			}
		}

	}

	return deploymentStatus, nil
}

// CompareResource compares the input resource name with the monitor resource name
func CompareResource(r string, qResource string) bool {

	if r == qResource {
		return true
	}
	return false
}

// CleanupCompositeApp will delete the app context
func CleanupCompositeApp(context appcontext.AppContext, err error, reason string, details []string) error {
	cleanuperr := context.DeleteCompositeApp()
	newerr := pkgerrors.Wrap(err, reason)
	if cleanuperr != nil {
		log.Warn("Error cleaning AppContext, ", log.Fields{
			"Related details": details,
		})
		return pkgerrors.Wrap(err, "After previous error, cleaning the AppContext also failed.")
	}
	return newerr
}

// RemoveChildCtx removes the child context ID in the parent's meta
func RemoveChildCtx(childContexts []string, childContextID string) {
	for i := range childContexts {
		if childContexts[i] == childContextID {
			childContexts[i] = childContexts[len(childContexts)-1]
			childContexts[len(childContexts)-1] = ""
			childContexts = childContexts[:len(childContexts)-1]
		}
	}
}
