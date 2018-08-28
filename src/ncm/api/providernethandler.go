// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	netintents "github.com/open-ness/EMCO/src/ncm/pkg/networkintents"
	nettypes "github.com/open-ness/EMCO/src/ncm/pkg/networkintents/types"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"
	pkgerrors "github.com/pkg/errors"

	"github.com/gorilla/mux"
)

var pnetJSONFile string = "json-schemas/provider-network.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type providernetHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client netintents.ProviderNetManager
}

// Check for valid format of input parameters
func validateProviderNetInputs(p netintents.ProviderNet) error {
	// validate name
	errs := validation.IsValidName(p.Metadata.Name)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid provider network name=[%v], errors: %v", p.Metadata.Name, errs)
	}

	// validate cni type
	found := false
	for _, val := range nettypes.CNI_TYPES {
		if p.Spec.CniType == val {
			found = true
			break
		}
	}
	if !found {
		return pkgerrors.Errorf("Invalid cni type: %v", p.Spec.CniType)
	}

	// validate the provider network type
	found = false
	for _, val := range nettypes.PROVIDER_NET_TYPES {
		if strings.ToUpper(p.Spec.ProviderNetType) == val {
			found = true
			break
		}
	}
	if !found {
		return pkgerrors.Errorf("Invalid provider network type: %v", p.Spec.ProviderNetType)
	}

	// validate the subnets
	subnets := p.Spec.Ipv4Subnets
	for _, subnet := range subnets {
		err := nettypes.ValidateSubnet(subnet)
		if err != nil {
			return pkgerrors.Wrap(err, "invalid subnet")
		}
	}

	// validate the VLAN ID
	errs = validation.IsValidNumberStr(p.Spec.Vlan.VlanId, 0, 4095)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid VlAN ID %v - error: %v", p.Spec.Vlan.VlanId, errs)
	}

	// validate the VLAN Node Selector value
	expectLabels := false
	found = false
	for _, val := range nettypes.VLAN_NODE_SELECTORS {
		if strings.ToLower(p.Spec.Vlan.VlanNodeSelector) == val {
			found = true
			if val == nettypes.VLAN_NODE_SPECIFIC {
				expectLabels = true
			}
			break
		}
	}
	if !found {
		return pkgerrors.Errorf("Invalid VlAN Node Selector %v", p.Spec.Vlan.VlanNodeSelector)
	}

	// validate the node label list
	gotLabels := false
	for _, label := range p.Spec.Vlan.NodeLabelList {
		errs = validation.IsValidLabel(label)
		if len(errs) > 0 {
			return pkgerrors.Errorf("Invalid Label=%v - errors: %v", label, errs)
		}
		gotLabels = true
	}

	// Need at least one label if node selector value was "specific"
	// (if selector is "any" - don't care if labels were supplied or not
	if expectLabels && !gotLabels {
		return pkgerrors.Errorf("Node Labels required for VlAN node selector \"%v\"", nettypes.VLAN_NODE_SPECIFIC)
	}

	return nil
}

// Create handles creation of the ProviderNet entry in the database
func (h providernetHandler) createProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	var p netintents.ProviderNet
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]

	err := json.NewDecoder(r.Body).Decode(&p)

	switch {
	case err == io.EOF:
		log.Error(":: Empty provider network POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding provider network POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(pnetJSONFile, p)
	if err != nil {
		log.Error(":: Invalid provider network POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing provider network name in POST body ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateProviderNetInputs(p)
	if err != nil {
		log.Error(":: Invalid provider network body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateProviderNet(p, clusterProvider, cluster, false)
	if err != nil {
		log.Error(":: Error creating provider network ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "Unable to find the cluster") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "Existing cluster provider network intents must be terminated before creating") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create provider network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles creation/update of the ProviderNet entry in the database
func (h providernetHandler) putProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	var p netintents.ProviderNet
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]
	name := vars["name"]

	err := json.NewDecoder(r.Body).Decode(&p)

	switch {
	case err == io.EOF:
		log.Error(":: Empty provider network PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing provider network name in PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if p.Metadata.Name != name {
		log.Error(":: Mismatched provider network name in PUT request ::", log.Fields{"URL name": name, "Metadata name": p.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateProviderNetInputs(p)
	if err != nil {
		log.Error(":: Invalid provider network PUT inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateProviderNet(p, clusterProvider, cluster, true)
	if err != nil {
		log.Error(":: Error updating provider network ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "Unable to find the cluster") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "Existing cluster provider network intents must be terminated before creating") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding provider network update response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular ProviderNet Name
// Returns a ProviderNet
func (h providernetHandler) getProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]
	name := vars["name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetProviderNets(clusterProvider, cluster)
		if err != nil {
			log.Error(":: Error getting provider networks ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		ret, err = h.client.GetProviderNet(name, clusterProvider, cluster)
		if err != nil {
			log.Error(":: Error getting provider network ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding get provider network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular ProviderNet  Name
func (h providernetHandler) deleteProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]
	name := vars["name"]

	err := h.client.DeleteProviderNet(name, clusterProvider, cluster)
	if err != nil {
		log.Error(":: Error deleting provider network ::", log.Fields{"Error": err, "Name": name})
		if strings.Contains(err.Error(), "Unable to find") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "conflict") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
