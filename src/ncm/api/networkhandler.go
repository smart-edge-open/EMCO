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

var vnJSONFile string = "json-schemas/virtual-network.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type networkHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client netintents.NetworkManager
}

// Check for valid format of input parameters
func validateNetworkInputs(p netintents.Network) error {
	// validate name
	errs := validation.IsValidName(p.Metadata.Name)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid network name - name=[%v], errors: %v", p.Metadata.Name, errs)
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

	subnets := p.Spec.Ipv4Subnets
	for _, subnet := range subnets {
		err := nettypes.ValidateSubnet(subnet)
		if err != nil {
			return pkgerrors.Wrap(err, "invalid subnet")
		}
	}
	return nil
}

// Create handles creation of the Network entry in the database
func (h networkHandler) createNetworkHandler(w http.ResponseWriter, r *http.Request) {
	var p netintents.Network
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]

	err := json.NewDecoder(r.Body).Decode(&p)

	switch {
	case err == io.EOF:
		log.Error(":: Empty network POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding network POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(vnJSONFile, p)
	if err != nil {
		log.Error(":: Invalid network POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing name in network POST request ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateNetworkInputs(p)
	if err != nil {
		log.Error(":: Invalid network body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateNetwork(p, clusterProvider, cluster, false)
	if err != nil {
		log.Error(":: Error creating network ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "Unable to find the cluster") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "Existing cluster network intents must be terminated before creating") {
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
		log.Error(":: Error encoding create network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles creation/update of the Network entry in the database
func (h networkHandler) putNetworkHandler(w http.ResponseWriter, r *http.Request) {
	var p netintents.Network
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]
	name := vars["name"]

	err := json.NewDecoder(r.Body).Decode(&p)

	switch {
	case err == io.EOF:
		log.Error(":: Empty network PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding network PUT body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing network name in PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if p.Metadata.Name != name {
		log.Error(":: Mismatched network name in PUT request ::", log.Fields{"URL name": name, "Metadata name": p.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateNetworkInputs(p)
	if err != nil {
		log.Error(":: Invalid network PUT inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateNetwork(p, clusterProvider, cluster, true)
	if err != nil {
		log.Error(":: Error updating network ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "Unable to find the cluster") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "Existing cluster network intents must be terminated before creating") {
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
		log.Error(":: Error encoding network update response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Network Name
// Returns a Network
func (h networkHandler) getNetworkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]
	name := vars["name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetNetworks(clusterProvider, cluster)
		if err != nil {
			log.Error(":: Error getting networks ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		ret, err = h.client.GetNetwork(name, clusterProvider, cluster)
		if err != nil {
			log.Error(":: Error getting network ::", log.Fields{"Error": err})
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
		log.Error(":: Error encoding get network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular Network  Name
func (h networkHandler) deleteNetworkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]
	name := vars["name"]

	err := h.client.DeleteNetwork(name, clusterProvider, cluster)
	if err != nil {
		log.Error(":: Error deleting network ::", log.Fields{"Error": err, "Name": name})
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
