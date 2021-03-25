// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"
	"github.com/open-ness/EMCO/src/sfcclient/pkg/model"
	"github.com/open-ness/EMCO/src/sfcclient/pkg/module"

	"github.com/gorilla/mux"
)

var sfcClientJSONFile string = "json-schemas/sfc-client.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type sfcHandler struct {
	// Interface that implements workload intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcManager
}

// Create handles creation of the SFC Client Intent entry in the database
func (h sfcHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var sfcClient model.SfcClientIntent
	vars := mux.Vars(r)
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := json.NewDecoder(r.Body).Decode(&sfcClient)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Client Intent POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Client Intent POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcClientJSONFile, sfcClient)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateSfcClientIntent(sfcClient, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, false)
	if err != nil {
		log.Error(":: Error creating sfc ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "SFC Client Intent already exists") {
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
		log.Error(":: Error encoding create SFC Client Intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles update of the SFC Client Intent entry in the database
func (h sfcHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var sfcClient model.SfcClientIntent
	vars := mux.Vars(r)
	name := vars["sfc-client-name"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := json.NewDecoder(r.Body).Decode(&sfcClient)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Client Intent PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Client Intent PUT body ::", log.Fields{"Error": err, "Body": sfcClient})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcClientJSONFile, sfcClient)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name in URL should match name in body
	if sfcClient.Metadata.Name != name {
		log.Error(":: Mismatched SFC Client Intent name in PUT request ::", log.Fields{"URL name": name, "Metadata name": sfcClient.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateSfcClientIntent(sfcClient, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, true)
	if err != nil {
		log.Error(":: Error updating SFC Client Intent ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "SFC Client Intent already exists") {
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
		log.Error(":: Error encoding update SFC Client Intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular SFC Client Intent Name
// Returns an SfcIntent
func (h sfcHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["sfc-client-name"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetAllSfcClientIntents(project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
		if err != nil {
			log.Error(":: Error getting SFC Client Intent ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else if strings.Contains(err.Error(), "not found") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		ret, err = h.client.GetSfcClientIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
		if err != nil {
			log.Error(":: Error getting SFC Client Intent ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else if strings.Contains(err.Error(), "not found") {
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
		log.Error(":: Error encoding get SFC Client Intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular SFC Client Intent
func (h sfcHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["sfc-client-name"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := h.client.DeleteSfcClientIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
	if err != nil {
		log.Error(":: Error deleting SFC Client Intent ::", log.Fields{"Error": err, "Name": name})
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
