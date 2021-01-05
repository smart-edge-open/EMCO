// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"
	moduleLib "github.com/open-ness/EMCO/src/ovnaction/pkg/module"
	pkgerrors "github.com/pkg/errors"

	"github.com/gorilla/mux"
)

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type chainHandler struct {
	// Interface that implements workload intent operations
	// We will set this variable with a mock interface for testing
	client moduleLib.ChainManager
}

func validateRoutingNetwork(r moduleLib.RoutingNetwork) error {
	errs := validation.IsValidName(r.NetworkName)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid routing network name: %v", errs)
	}

	err := validation.IsIpv4Cidr(r.Subnet)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid routing network subnet")
	}

	err = validation.IsIpv4(r.GatewayIP)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid routing network gateway IP")
	}

	return nil
}

// validateNetworkChain checks that the network chain string input follows the
// generic format:  "app=app1,net1,app=app2,net2, ..... ,netN-1,app=appN"
// assume "app=app1" can conform to validation.IsValidLabel() with an "="
func validateNetworkChain(chain string) error {
	elems := strings.Split(chain, ",")

	// chain needs at least two apps and a network
	if len(elems) < 3 {
		return pkgerrors.Errorf("Network chain is too short")
	}

	// chain needs to have an odd number of elements
	if len(elems)%2 == 0 {
		return pkgerrors.Errorf("Invalid network chain - even number of elements")
	}

	for i, s := range elems {
		// allows whitespace in comma separated elements
		t := strings.TrimSpace(s)
		// if even element, verify a=b format
		if i%2 == 0 {
			if strings.Index(t, "=") < 1 {
				return pkgerrors.Errorf("Invalid deployment label element of network chain")
			}
			errs := validation.IsValidLabel(t)
			if len(errs) > 0 {
				return pkgerrors.Errorf("Invalid deployment label element: %v", errs)
			}
		} else {
			errs := validation.IsValidName(t)
			if len(errs) > 0 {
				return pkgerrors.Errorf("Invalid network element of network chain: %v", errs)
			}
		}
	}
	return nil
}

// Check for valid format of input parameters
func validateChainInputs(ch moduleLib.Chain) error {
	if strings.ToLower(ch.Spec.ChainType) != moduleLib.RoutingChainType {
		return pkgerrors.New("Invalid network chain type")
	}

	for _, r := range ch.Spec.RoutingSpec.LeftNetwork {
		err := validateRoutingNetwork(r)
		if err != nil {
			return err
		}
	}

	for _, r := range ch.Spec.RoutingSpec.RightNetwork {
		err := validateRoutingNetwork(r)
		if err != nil {
			return err
		}
	}

	err := validateNetworkChain(ch.Spec.RoutingSpec.NetworkChain)
	if err != nil {
		return err
	}

	errs := validation.IsValidName(ch.Spec.RoutingSpec.Namespace)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid network chain route spec namespace: %v", errs)
	}

	return nil
}

// Create handles creation of the Chain entry in the database
func (h chainHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var ch moduleLib.Chain
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := json.NewDecoder(r.Body).Decode(&ch)

	switch {
	case err == io.EOF:
		log.Error(":: Empty network chain POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding network chain POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if ch.Metadata.Name == "" {
		log.Error(":: Missing name in network chain POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateChainInputs(ch)
	if err != nil {
		log.Error(":: Invalid network chain body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateChain(ch, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, false)
	if err != nil {
		log.Error(":: Error creating network chain ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "Chain already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create network chain response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles creation/update of the Chain entry in the database
func (h chainHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var ch moduleLib.Chain
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := json.NewDecoder(r.Body).Decode(&ch)

	switch {
	case err == io.EOF:
		log.Error(":: Empty network chain PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding network chain PUT body ::", log.Fields{"Error": err, "Body": ch})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if ch.Metadata.Name == "" {
		log.Error(":: Missing network chain name in PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if ch.Metadata.Name != name {
		log.Error(":: Mismatched network chain name in PUT request ::", log.Fields{"URL name": name, "Metadata name": ch.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateChainInputs(ch)
	if err != nil {
		log.Error(":: Invalid network chain inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateChain(ch, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, true)
	if err != nil {
		log.Error(":: Error updating network chain ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "Chain already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding update network chain response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Chain Name
// Returns a Chain
func (h chainHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetChains(project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
		if err != nil {
			log.Error(":: Error getting network chains ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		ret, err = h.client.GetChain(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
		if err != nil {
			log.Error(":: Error getting network chain ::", log.Fields{"Error": err})
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
		log.Error(":: Error encoding get network chain response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular Chain
func (h chainHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := h.client.DeleteChain(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
	if err != nil {
		log.Error(":: Error deleting network chain ::", log.Fields{"Error": err, "Name": name})
		if strings.Contains(err.Error(), "not found") {
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
