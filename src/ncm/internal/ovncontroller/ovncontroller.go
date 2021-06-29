// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package ovncontroller

import (
	"encoding/json"

	otheryaml "github.com/ghodss/yaml"
	nad "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	netintents "github.com/open-ness/EMCO/src/ncm/pkg/networkintents"
	nettypes "github.com/open-ness/EMCO/src/ncm/pkg/networkintents/types"
	appcontext "github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgerrors "github.com/pkg/errors"
)

// makeNetworkAttachmentDefinition makes a network attachment definition and
// returns it as a string ready to be added to the resources list for the
// appcontext
func makeNetworkAttachmentDefinition(name string) (string, error) {
	nad := nad.NetworkAttachmentDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "k8s.cni.cncf.io/v1",
			Kind:       "NetworkAttachmentDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: nad.NetworkAttachmentDefinitionSpec{
			Config: `{
                  "cniVersion": "0.3.1",
                  "type": "ovn4nfvk8s-cni",
                  "nfn-network": "` + name + `"
                }`,
		},
	}
	y, err := otheryaml.Marshal(&nad)
	if err != nil {
		log.Error("Error marshalling netwrok attachment definition to yaml", log.Fields{
			"error": err,
			"name":  name,
		})
		return "", err
	}
	return string(y), nil
}

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
		crNetwork.MetaData.Labels = make(map[string]string)
		crNetwork.MetaData.Labels[nettypes.NetLabel] = intent.Metadata.Name
		crNetwork.MetaData.Name = intent.Metadata.Name
		crNetwork.NetworkSpec = intent.Spec
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

		// make the network attachment definition CR for this network
		nadRes, err := makeNetworkAttachmentDefinition(intent.Metadata.Name)
		if err != nil {
			// TODO - probably should error out instead (and above too)
			continue
		}
		resources = append(resources, resource{
			name:  intent.Metadata.Name + nettypes.SEPARATOR + "NetworkAttachmentDefinition",
			value: nadRes,
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
		//crProviderNet.MetaData.Labels = make(map[string]string)
		//crProviderNet.MetaData.Labels[nettypes.NetLabel] = intent.Metadata.Name
		crProviderNet.MetaData.Name = intent.Metadata.Name
		crProviderNet.ProviderNetSpec = intent.Spec
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

		// make the network attachment definition CR for this network
		nadRes, err := makeNetworkAttachmentDefinition(intent.Metadata.Name)
		if err != nil {
			// TODO - probably should error out instead (and above too)
			continue
		}
		resources = append(resources, resource{
			name:  intent.Metadata.Name + nettypes.SEPARATOR + "NetworkAttachmentDefinition",
			value: nadRes,
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
