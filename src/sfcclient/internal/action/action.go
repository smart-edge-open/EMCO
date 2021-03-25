// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package action

import (
	"encoding/json"
	"strings"

	jyaml "github.com/ghodss/yaml"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	cacontext "github.com/open-ness/EMCO/src/rsync/pkg/context"
	catypes "github.com/open-ness/EMCO/src/rsync/pkg/types"
	sfc "github.com/open-ness/EMCO/src/sfc/pkg/module"
	sfcclient "github.com/open-ness/EMCO/src/sfcclient/pkg/module"
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// getChainApps will return the list of applications that are present in
// the provided string which follows the format of the NetworkChain field.
// "net=virutal-net1,app=slb,dync-net1,app=ngfw,dync-net2,app=sdewan,net=virutal-net2"
func getChainApps(networkChain string) ([]string, error) {
	netsAndApps := strings.Split(networkChain, ",")
	apps := make([]string, 0)
	for _, netOrApp := range netsAndApps {
		elem := strings.Split(netOrApp, "=")
		if len(elem) != 2 {
			return []string{}, pkgerrors.Errorf("Invalid network chain format: %v", networkChain)
		}
		if elem[0] == "app" {
			apps = append(apps, elem[1])
		}
	}
	return apps, nil
}

// chainClusters returns the list of clusters to which the Network Chain needs to be
// deployed.  To qualify, a cluster must be present for each app in the apps list.
func chainClusters(apps []string, ac catypes.CompositeApp) map[string]struct{} {
	clusters := make(map[string]struct{}, 0)
	for i, a := range apps {
		// an app in the chain is not in the AppContext, so the clusters list is empty
		if _, ok := ac.Apps[a]; !ok {
			return make(map[string]struct{}, 0)
		}

		// for first app, the list of that apps clusters in the AppContext is the starting cluster list
		if i == 0 {
			for k, _ := range ac.Apps[a].Clusters {
				clusters[k] = struct{}{}
			}
		} else {
			// for the rest of the apps, whittle down the clusters list to find the
			// common intersection for all apps in the chain
			for k, _ := range clusters {
				if _, ok := ac.Apps[a].Clusters[k]; !ok {
					delete(clusters, k)
				}
			}
		}
	}
	return clusters
}

// UpdateAppContext applies the supplied intent against the given AppContext ID
// The SFC Client controller will handle all SFC Client intents that are found for the
// supplied Network Control Intent (intentName).
func UpdateAppContext(intentName, appContextId string) error {

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextId)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error loading AppContext with Id: %v", appContextId)
	}
	//cahandle, err := ac.GetCompositeAppHandle()
	_, err = ac.GetCompositeAppHandle()
	if err != nil {
		return err
	}

	appContext, err := cacontext.ReadAppContext(appContextId)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error reading AppContext with Id: %v", appContextId)
	}

	pr := appContext.CompMetadata.Project
	ca := appContext.CompMetadata.CompositeApp
	caver := appContext.CompMetadata.Version
	dig := appContext.CompMetadata.DeploymentIntentGroup

	// Look up all SFC Client Intents
	sfcClientIntents, err := sfcclient.NewSfcClient().GetAllSfcClientIntents(pr, ca, caver, dig, intentName)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting SFC Client Intents for Network Control Intent: %v", intentName)
	}

	if len(sfcClientIntents) == 0 {
		return pkgerrors.Errorf("No SFC Client Intents are defined for the Network Control Intent: %v", intentName)
	}

	// For each SFC Client Intent ...
	for _, sfcClientInt := range sfcClientIntents {

		// query all SFC Client Selector Intents that match the chainEnd of this SFC Client Intent
		sfcClientSelectorIntents, err := sfc.NewSfcClientSelectorIntentClient().GetSfcClientSelectorIntentsByEnd(
			pr,
			sfcClientInt.Spec.ChainCompositeApp,
			sfcClientInt.Spec.ChainCompositeAppVersion,
			sfcClientInt.Spec.ChainDeploymentIntentGroup,
			sfcClientInt.Spec.ChainNetControlIntent,
			sfcClientInt.Spec.ChainName,
			sfcClientInt.Spec.ChainEnd)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error finding SFC Client Selector Intents for %v by end: %v", sfcClientInt.Spec.ChainName, sfcClientInt.Spec.ChainEnd)
		}
		if len(sfcClientSelectorIntents) == 0 {
			return pkgerrors.Errorf("No SFC Client Selector Intents found for %v by end: %v", sfcClientInt.Spec.ChainName, sfcClientInt.Spec.ChainEnd)
		}

		labels := make(map[string]string)
		for i, sfcClientSelectorIntent := range sfcClientSelectorIntents {
			if i > 0 {
				// only handle one intent (for now)
				log.Warn("More than one SFC Client Selector Intent was found for chain end.",
					log.Fields{"Chain": sfcClientInt.Spec.ChainName, "End": sfcClientInt.Spec.ChainEnd})
				break
			}
			// copy the labels (necessary?)
			for k, v := range sfcClientSelectorIntent.Spec.PodSelector.MatchLabels {
				labels[k] = v
			}
		}

		// Get all clusters for the current App from the AppContext
		clusters, err := ac.GetClusterNames(sfcClientInt.Spec.AppName)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error getting clusters for app: %v", sfcClientInt.Spec.AppName)
		}
		for _, c := range clusters {
			rh, err := ac.GetResourceHandle(sfcClientInt.Spec.AppName, c,
				strings.Join([]string{sfcClientInt.Spec.WorkloadResource,
					sfcClientInt.Spec.ResourceType}, "+"))
			if err != nil {
				log.Error("App Context resource handle not found", log.Fields{
					"project":                 pr,
					"composite app":           ca,
					"composite app version":   caver,
					"deployment intent group": dig,
					"network control intent":  intentName,
					"sfc client":              sfcClientInt.Metadata.Name,
					"app":                     sfcClientInt.Spec.AppName,
					"resource":                sfcClientInt.Spec.WorkloadResource,
					"resource type":           sfcClientInt.Spec.ResourceType,
				})
				return pkgerrors.Wrapf(err, "Error getting resource handle [%v] for SFC client [%v] from cluster [%v]",
					strings.Join([]string{sfcClientInt.Spec.WorkloadResource,
						sfcClientInt.Spec.ResourceType}, "+"),
					sfcClientInt.Metadata.Name, c)
			}
			r, err := ac.GetValue(rh)
			if err != nil {
				log.Error("Error retrieving resource from App Context", log.Fields{
					"error":           err,
					"resource handle": rh,
				})
				return pkgerrors.Wrapf(err, "Error getting resource value [%v] for SFC client [%v] from cluster [%v]",
					strings.Join([]string{sfcClientInt.Spec.WorkloadResource,
						sfcClientInt.Spec.ResourceType}, "+"),
					sfcClientInt.Metadata.Name, c)
			}

			// Unmarshal resource to K8S object
			robj, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(r.(string)))
			if err != nil {
				return pkgerrors.Wrapf(err, "Error decoding resource: %v", sfcClientInt.Spec.WorkloadResource)
			}

			// add labels to resource
			AddLabelsToPodTemplates(robj, labels)

			// Marshal object back to yaml format (via json - seems to eliminate most clutter)
			j, err := json.Marshal(robj)
			if err != nil {
				log.Error("Error marshalling resource to JSON", log.Fields{
					"error": err,
				})
				return pkgerrors.Wrapf(err,
					"Error marshalling to JSON resource value [%v] for SFC client [%v] from cluster [%v]",
					strings.Join([]string{sfcClientInt.Spec.WorkloadResource,
						sfcClientInt.Spec.ResourceType}, "+"),
					sfcClientInt.Metadata.Name,
					c)
			}
			y, err := jyaml.JSONToYAML(j)
			if err != nil {
				log.Error("Error marshalling resource to YAML", log.Fields{
					"error": err,
				})
				return pkgerrors.Wrapf(err,
					"Error marshalling to YAML resource value [%v] for SFC client [%v] from cluster [%v]",
					strings.Join([]string{sfcClientInt.Spec.WorkloadResource,
						sfcClientInt.Spec.ResourceType}, "+"),
					sfcClientInt.Metadata.Name, c)
			}

			// Update resource in AppContext
			err = ac.UpdateResourceValue(rh, string(y))
			if err != nil {
				log.Error("Network updating app context resource handle", log.Fields{
					"error":           err,
					"resource handle": rh,
				})
			}
		}
	}
	return nil
}
