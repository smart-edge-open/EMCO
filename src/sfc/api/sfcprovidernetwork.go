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
	"github.com/open-ness/EMCO/src/sfc/pkg/model"

	"github.com/gorilla/mux"
)

var sfcProviderNetworkJSONFile string = "json-schemas/sfc-provider-network.json"

// Create handles creation of the SFC Provider Network entry in the database
func (h sfcProviderNetworkIntentHandler) createProviderNetworkHandler(w http.ResponseWriter, r *http.Request) {
	var sfcProviderNetwork model.SfcProviderNetworkIntent
	vars := mux.Vars(r)
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	sfcIntent := vars["sfc-name"]

	err := json.NewDecoder(r.Body).Decode(&sfcProviderNetwork)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Provider Network POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Provider Network POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcProviderNetworkJSONFile, sfcProviderNetwork)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateSfcProviderNetworkIntent(sfcProviderNetwork, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, sfcIntent, false)
	if err != nil {
		log.Error(":: Error creating SFC Provider Network ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "SFC Provider Network already exists") {
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
		log.Error(":: Error encoding create SFC Provider Network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles update of the SFC Provider Network entry in the database
func (h sfcProviderNetworkIntentHandler) putProviderNetworkHandler(w http.ResponseWriter, r *http.Request) {
	var sfcProviderNetwork model.SfcProviderNetworkIntent
	vars := mux.Vars(r)
	name := vars["sfc-provider-network"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	sfcIntent := vars["sfc-name"]

	err := json.NewDecoder(r.Body).Decode(&sfcProviderNetwork)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Provider Network PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Provider Network PUT body ::", log.Fields{"Error": err, "Body": sfcProviderNetwork})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcProviderNetworkJSONFile, sfcProviderNetwork)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name in URL should match name in body
	if sfcProviderNetwork.Metadata.Name != name {
		log.Error(":: Mismatched SFC Provider Network name in PUT request ::", log.Fields{"URL name": name, "Metadata name": sfcProviderNetwork.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateSfcProviderNetworkIntent(sfcProviderNetwork, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, sfcIntent, true)
	if err != nil {
		log.Error(":: Error updating SFC Provider Network ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "SFC Provider Network already exists") {
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
		log.Error(":: Error encoding update SFC Provider Network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular SFC Provider Network Name
// Returns a SFC Provider Network
func (h sfcProviderNetworkIntentHandler) getProviderNetworkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["sfc-provider-network"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	sfcIntent := vars["sfc-name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetAllSfcProviderNetworkIntents(project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, sfcIntent)
		if err != nil {
			log.Error(":: Error getting SFC Provider Network ::", log.Fields{"Error": err})
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
		ret, err = h.client.GetSfcProviderNetworkIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, sfcIntent)
		if err != nil {
			log.Error(":: Error getting SFC Provider Network ::", log.Fields{"Error": err})
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
		log.Error(":: Error encoding get SFC Provider Network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular SfcProviderNetwork
func (h sfcProviderNetworkIntentHandler) deleteProviderNetworkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["sfc-provider-network"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	sfcIntent := vars["sfc-name"]

	err := h.client.DeleteSfcProviderNetworkIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, sfcIntent)
	if err != nil {
		log.Error(":: Error deleting SFC Provider Network ::", log.Fields{"Error": err, "Name": name})
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
