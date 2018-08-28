// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package ovncontroller

import (
	"encoding/json"

	netintents "github.com/open-ness/EMCO/src/ncm/pkg/networkintents"
	nettypes "github.com/open-ness/EMCO/src/ncm/pkg/networkintents/types"
	appcontext "github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"gopkg.in/yaml.v2"

	pkgerrors "github.com/pkg/errors"
)

// controller takes an appcontext as input
//   finds the cluster(s) associated with the context
//   queries the network intents and adds resources to the context
func Apply(ctxVal interface{}, clusterProvider, cluster string) error {
	type resource struct {
		name  string
		value string
	}
	var resources []resource

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(ctxVal)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting AppContext with Id: %v for %v/%v", ctxVal, clusterProvider, cluster)
	}

	// Find all Network Intents for this cluster
	networkIntents, err := netintents.NewNetworkClient().GetNetworks(clusterProvider, cluster)
	if err != nil {
		return pkgerrors.Wrap(err, "Error finding Network Intents")
	}
	for _, intent := range networkIntents {
		var crNetwork = netintents.CrNetwork{
			ApiVersion: netintents.NETWORK_APIVERSION,
			Kind:       netintents.NETWORK_KIND,
		}
		crNetwork.Network = intent
		// Produce the yaml CR document for each intent
		y, err := yaml.Marshal(&crNetwork)
		if err != nil {
			log.Error("Error marshalling network intent to yaml", log.Fields{
				"error":  err,
				"intent": intent,
			})
			continue
		}
		resources = append(resources, resource{
			name:  intent.Metadata.Name + nettypes.SEPARATOR + netintents.NETWORK_KIND,
			value: string(y),
		})
	}

	// Find all Provider Network Intents for this cluster
	providerNetworkIntents, err := netintents.NewProviderNetClient().GetProviderNets(clusterProvider, cluster)
	if err != nil {
		return pkgerrors.Wrap(err, "Error finding Provider Network Intents")
	}
	for _, intent := range providerNetworkIntents {
		var crProviderNet = netintents.CrProviderNet{
			ApiVersion: netintents.PROVIDER_NETWORK_APIVERSION,
			Kind:       netintents.PROVIDER_NETWORK_KIND,
		}
		crProviderNet.ProviderNet = intent
		// Produce the yaml CR document for each intent
		y, err := yaml.Marshal(&crProviderNet)
		if err != nil {
			log.Error("Error marshalling provider network intent to yaml", log.Fields{
				"error":  err,
				"intent": intent,
			})
			continue
		}
		resources = append(resources, resource{
			name:  intent.Metadata.Name + nettypes.SEPARATOR + netintents.PROVIDER_NETWORK_KIND,
			value: string(y),
		})
	}

	if len(resources) == 0 {
		return nil
	}

	acCluster := clusterProvider + nettypes.SEPARATOR + cluster
	clusterhandle, err := ac.GetClusterHandle(nettypes.CONTEXT_CLUSTER_APP, acCluster)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting cluster handle")
	}

	var orderinstr struct {
		Resorder []string `json:"resorder"`
	}
	var depinstr struct {
		Resdep map[string]string `json:"resdependency"`
	}
	resdep := make(map[string]string)
	for _, resource := range resources {
		orderinstr.Resorder = append(orderinstr.Resorder, resource.name)
		resdep[resource.name] = "go"
		_, err := ac.AddResource(clusterhandle, resource.name, resource.value)
		if err != nil {
			return pkgerrors.Wrap(err, "Error adding Resource to AppContext")
		}
	}
	jresord, err := json.Marshal(orderinstr)
	if err != nil {
		return pkgerrors.Wrap(err, "Error marshalling resource order instruction")
	}
	depinstr.Resdep = resdep
	jresdep, err := json.Marshal(depinstr)
	if err != nil {
		return pkgerrors.Wrap(err, "Error marshalling resource dependency instruction")
	}
	_, err = ac.AddInstruction(clusterhandle, "resource", "order", string(jresord))
	if err != nil {
		return pkgerrors.Wrap(err, "Error adding resource order instruction")
	}
	_, err = ac.AddInstruction(clusterhandle, "resource", "dependency", string(jresdep))
	if err != nil {
		return pkgerrors.Wrap(err, "Error adding resource dependency instruction")
	}

	return nil
}
