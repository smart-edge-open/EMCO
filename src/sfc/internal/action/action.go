// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package action

import (
	"encoding/json"
	"strings"

	nodus "github.com/akraino-edge-stack/icn-nodus/pkg/apis/k8s/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	cacontext "github.com/open-ness/EMCO/src/rsync/pkg/context"
	catypes "github.com/open-ness/EMCO/src/rsync/pkg/types"
	"github.com/open-ness/EMCO/src/sfc/pkg/model"
	sfc "github.com/open-ness/EMCO/src/sfc/pkg/module"
	pkgerrors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// Action applies the supplied intent against the given AppContext ID
func UpdateAppContext(intentName, appContextId string) error {

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextId)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error loading AppContext with Id: %v", appContextId)
	}
	cahandle, err := ac.GetCompositeAppHandle()
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

	// Look up all SFC Intents
	sfcIntents, err := sfc.NewSfcIntentClient().GetAllSfcIntents(pr, ca, caver, dig, intentName)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting SFC Intents for Network Control Intent: %v", intentName)
	}

	if len(sfcIntents) == 0 {
		return pkgerrors.Errorf("No SFC Intents are defined for the Network Control Intent: %v", intentName)
	}

	// For each SFC Intent prepare a NetworkChaining resource and add to the AppContext
	for i, sfcInt := range sfcIntents {
		// Lookup all SFC Client Selector Intents
		sfcClientSelectorIntents, err := sfc.NewSfcClientSelectorIntentClient().GetAllSfcClientSelectorIntents(pr, ca, caver, dig, intentName, sfcInt.Metadata.Name)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error getting SFC Client Selector intents for SFC Intent: %v", sfcInt.Metadata.Name)
		}

		// Lookup all SFC Provider Network Intents
		sfcProviderNetworkIntents, err := sfc.NewSfcProviderNetworkIntentClient().GetAllSfcProviderNetworkIntents(pr, ca, caver, dig, intentName, sfcInt.Metadata.Name)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error getting SFC Provider Network intents for SFC Intent: %v", sfcInt.Metadata.Name)
		}

		// Start preparing the networkchainings CR structure
		// REVISIT - for now, the code will expect 1 occurrence of a client selector and provider network on each side of the
		// chain.  The first occurrence found will be used.  Extras will be ignored.
		// Revisit this code once the chaining CR behavior is verified complete.
		var leftRoutingNetwork nodus.RoutingNetwork
		var rightRoutingNetwork nodus.RoutingNetwork

		leftClient := false
		rightClient := false
		for _, sfcClientSelectorInt := range sfcClientSelectorIntents {
			if sfcClientSelectorInt.Spec.ChainEnd == model.LeftChainEnd && !leftClient {
				leftRoutingNetwork.PodSelector = sfcClientSelectorInt.Spec.PodSelector
				leftRoutingNetwork.NamespaceSelector = sfcClientSelectorInt.Spec.NamespaceSelector
				leftClient = true
			} else if sfcClientSelectorInt.Spec.ChainEnd == model.RightChainEnd && !rightClient {
				rightRoutingNetwork.PodSelector = sfcClientSelectorInt.Spec.PodSelector
				rightRoutingNetwork.NamespaceSelector = sfcClientSelectorInt.Spec.NamespaceSelector
				rightClient = true
			}
			if leftClient && rightClient {
				break
			}
		}
		if !leftClient && !rightClient {
			return pkgerrors.New("Missing left and right client selector intents")
		}
		if !leftClient {
			return pkgerrors.New("Missing left client selector intent")
		}
		if !rightClient {
			return pkgerrors.New("Missing right client selector intent")
		}

		leftNet := false
		rightNet := false
		for _, sfcProviderNetInt := range sfcProviderNetworkIntents {
			if sfcProviderNetInt.Spec.ChainEnd == model.LeftChainEnd && !leftNet {
				leftRoutingNetwork.NetworkName = sfcProviderNetInt.Spec.NetworkName
				leftRoutingNetwork.GatewayIP = sfcProviderNetInt.Spec.GatewayIp
				leftRoutingNetwork.Subnet = sfcProviderNetInt.Spec.Subnet
				leftNet = true
			} else if sfcProviderNetInt.Spec.ChainEnd == model.RightChainEnd && !rightNet {
				rightRoutingNetwork.NetworkName = sfcProviderNetInt.Spec.NetworkName
				rightRoutingNetwork.GatewayIP = sfcProviderNetInt.Spec.GatewayIp
				rightRoutingNetwork.Subnet = sfcProviderNetInt.Spec.Subnet
				rightNet = true
			}
			if leftNet && rightNet {
				break
			}
		}
		if !leftNet && !rightNet {
			return pkgerrors.New("Missing left and right provider network intents")
		}
		if !leftNet {
			return pkgerrors.New("Missing left provider network intent")
		}
		if !rightNet {
			return pkgerrors.New("Missing right provider network intent")
		}

		leftNetwork := make([]nodus.RoutingNetwork, 0)
		rightNetwork := make([]nodus.RoutingNetwork, 0)
		leftNetwork = append(leftNetwork, leftRoutingNetwork)
		rightNetwork = append(rightNetwork, rightRoutingNetwork)
		chain := nodus.NetworkChaining{
			TypeMeta: metav1.TypeMeta{
				APIVersion: model.ChainingAPIVersion,
				Kind:       model.ChainingKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: sfcInt.Metadata.Name,
			},
			Spec: nodus.NetworkChainingSpec{
				ChainType: sfcInt.Spec.ChainType,
				RoutingSpec: nodus.RouteSpec{
					Namespace:    sfcInt.Spec.Namespace,
					NetworkChain: sfcInt.Spec.NetworkChain,
					LeftNetwork:  leftNetwork,
					RightNetwork: rightNetwork,
				},
			},
		}
		chainYaml, err := yaml.Marshal(&chain)
		if err != nil {
			return pkgerrors.Wrapf(err, "Failed to marshal NetworkChaining CR: %v", sfcInt.Metadata.Name)
		}

		// Get the list of apps in the network chain
		// Assumption: the 'app=appname' elements in the network chain are assumed to be app names in the AppContext
		chainApps, err := getChainApps(sfcInt.Spec.NetworkChain)
		if err != nil {
			return err
		}

		// Get the clusters which should get the NetworkChaining resource
		clusters := chainClusters(chainApps, appContext)
		if len(clusters) == 0 {
			return pkgerrors.Errorf("There are no clusters with all the apps for the Network Chain: %v", sfcInt.Spec.NetworkChain)
		}

		// Add the network intents chaining app to the AppContext
		var apphandle interface{}
		if i == 0 {
			apphandle, err = ac.AddApp(cahandle, model.ChainingApp)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding ChainingApp to AppContext: %v", sfcInt.Metadata.Name)
			}

			// need to update the app order instruction
			apporder, err := ac.GetAppInstruction(appcontext.OrderInstruction)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error getting order instruction while adding ChainingApp to AppContext: %v", sfcInt.Metadata.Name)
			}
			aov := make(map[string][]string)
			json.Unmarshal([]byte(apporder.(string)), &aov)
			aov["apporder"] = append(aov["apporder"], model.ChainingApp)
			jappord, _ := json.Marshal(aov)

			_, err = ac.AddInstruction(cahandle, appcontext.AppLevel, appcontext.OrderInstruction, string(jappord))
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding ChainingApp to order instruction: %v", sfcInt.Metadata.Name)
			}
		} else {
			apphandle, err = ac.GetAppHandle(model.ChainingApp)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error getting ChainingApp handle from AppContext: %v", sfcInt.Metadata.Name)
			}
		}

		// Add each cluster to the chaining app and the chaining CR resource to each cluster
		for cluster, _ := range clusters {
			clusterhandle, err := ac.AddCluster(apphandle, cluster)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding cluster to ChainingApp: %v", cluster)
			}

			resName := sfcInt.Metadata.Name + appcontext.Separator + model.ChainingKind
			_, err = ac.AddResource(clusterhandle, resName, string(chainYaml))
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding Network Chain resource: %v", sfcInt.Metadata.Name)
			}

			// add (first time) or update the resource order instruction
			aov := make(map[string][]string)
			resorder, err := ac.GetResourceInstruction(model.ChainingApp, cluster, appcontext.OrderInstruction)
			if err != nil {
				// instruction not found - create it
				aov["resorder"] = []string{resName}
			} else {
				json.Unmarshal([]byte(resorder.(string)), &aov)
				aov["resorder"] = append(aov["resorder"], resName)
			}
			jresord, _ := json.Marshal(aov)

			_, err = ac.AddInstruction(clusterhandle, appcontext.ResourceLevel, appcontext.OrderInstruction, string(jresord))
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding Network Chain to resource order instruction: %v", sfcInt.Metadata.Name)
			}
		}
	}
	return nil
}
