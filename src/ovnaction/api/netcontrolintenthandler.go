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

var netCntIntJSONFile string = "json-schemas/metadata.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type netcontrolintentHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client moduleLib.NetControlIntentManager
}

// Check for valid format of input parameters
func validateNetControlIntentInputs(nci moduleLib.NetControlIntent) error {
	// validate metadata
	err := moduleLib.IsValidMetadata(nci.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid network controller intent metadata")
	}
	return nil
}

// Create handles creation of the NetControlIntent entry in the database
func (h netcontrolintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var nci moduleLib.NetControlIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]

	err := json.NewDecoder(r.Body).Decode(&nci)

	switch {
	case err == io.EOF:
		log.Error(":: Empty network control intent POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding network control intent POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(netCntIntJSONFile, nci)
	if err != nil {
		log.Error(":: Invalid network control intent body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if nci.Metadata.Name == "" {
		log.Error(":: Missing name in network control intent POST request ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateNetControlIntentInputs(nci)
	if err != nil {
		log.Error(":: Invalid network control intent body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateNetControlIntent(nci, project, compositeApp, compositeAppVersion, deployIntentGroup, false)
	if err != nil {
		log.Error(":: Error creating network control intent ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "NetControlIntent already exists") {
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
		log.Error(":: Error encoding create network control intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles creation/update of the NetControlIntent entry in the database
func (h netcontrolintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var nci moduleLib.NetControlIntent
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]

	err := json.NewDecoder(r.Body).Decode(&nci)

	switch {
	case err == io.EOF:
		log.Error(":: Empty network control intent PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding network control intent PUT body ::", log.Fields{"Error": err, "Body": nci})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if nci.Metadata.Name == "" {
		log.Error(":: Missing network control intent name in PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if nci.Metadata.Name != name {
		log.Error(":: Mismatched network control intent name in PUT request ::", log.Fields{"URL name": name, "Metadata name": nci.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateNetControlIntentInputs(nci)
	if err != nil {
		log.Error(":: Invalid network control intent inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateNetControlIntent(nci, project, compositeApp, compositeAppVersion, deployIntentGroup, true)
	if err != nil {
		log.Error(":: Error updating network control intent ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "NetControlIntent already exists") {
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
		log.Error(":: Error encoding update network control intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular NetControlIntent Name
// Returns a NetControlIntent
func (h netcontrolintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetNetControlIntents(project, compositeApp, compositeAppVersion, deployIntentGroup)
		if err != nil {
			log.Error(":: Error getting network control intents ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		ret, err = h.client.GetNetControlIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup)
		if err != nil {
			log.Error(":: Error getting network control intent ::", log.Fields{"Error": err})
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
		log.Error(":: Error encoding get network control intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular NetControlIntent  Name
func (h netcontrolintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]

	err := h.client.DeleteNetControlIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup)
	if err != nil {
		log.Error(":: Error deleting network control intent ::", log.Fields{"Error": err, "Name": name})
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
