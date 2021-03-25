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

var sfcClientSelectorJSONFile string = "json-schemas/sfc-client-selector.json"

// Create handles creation of the SFC Client Selector entry in the database
func (h sfcClientSelectorIntentHandler) createClientSelectorHandler(w http.ResponseWriter, r *http.Request) {
	var sfcClientSelector model.SfcClientSelectorIntent
	vars := mux.Vars(r)
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	sfcIntent := vars["sfc-name"]

	err := json.NewDecoder(r.Body).Decode(&sfcClientSelector)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Client Selector POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Client Selector POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcClientSelectorJSONFile, sfcClientSelector)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateSfcClientSelectorIntent(sfcClientSelector, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, sfcIntent, false)
	if err != nil {
		log.Error(":: Error creating SFC Client Selector ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "SFC Client Selector already exists") {
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
		log.Error(":: Error encoding create SFC Client Selector response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles update of the SFC Client Selector entry in the database
func (h sfcClientSelectorIntentHandler) putClientSelectorHandler(w http.ResponseWriter, r *http.Request) {
	var sfcClientSelectorIntent model.SfcClientSelectorIntent
	vars := mux.Vars(r)
	name := vars["sfc-client-selector"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	sfcIntent := vars["sfc-name"]

	err := json.NewDecoder(r.Body).Decode(&sfcClientSelectorIntent)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Client Selector PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Client Selector PUT body ::", log.Fields{"Error": err, "Body": sfcClientSelectorIntent})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcClientSelectorJSONFile, sfcClientSelectorIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name in URL should match name in body
	if sfcClientSelectorIntent.Metadata.Name != name {
		log.Error(":: Mismatched SFC Client Selector name in PUT request ::", log.Fields{"URL name": name, "Metadata name": sfcClientSelectorIntent.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateSfcClientSelectorIntent(sfcClientSelectorIntent, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, sfcIntent, true)
	if err != nil {
		log.Error(":: Error updating SFC Client Selector ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "SFC Client Selector already exists") {
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
		log.Error(":: Error encoding update SFC Client Selector response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular SFC Client Selector Name
// Returns a SFC Client Selector
func (h sfcClientSelectorIntentHandler) getClientSelectorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["sfc-client-selector"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	sfcIntent := vars["sfc-name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetAllSfcClientSelectorIntents(project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, sfcIntent)
		if err != nil {
			log.Error(":: Error getting SFC Client Selector ::", log.Fields{"Error": err})
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
		ret, err = h.client.GetSfcClientSelectorIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, sfcIntent)
		if err != nil {
			log.Error(":: Error getting SFC Client Selector ::", log.Fields{"Error": err})
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
		log.Error(":: Error encoding get SFC Client Selector response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular SfcClientSelector
func (h sfcClientSelectorIntentHandler) deleteClientSelectorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["sfc-client-selector"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	sfcIntent := vars["sfc-name"]

	err := h.client.DeleteSfcClientSelectorIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, sfcIntent)
	if err != nil {
		log.Error(":: Error deleting SFC Client Selector Client Selector ::", log.Fields{"Error": err, "Name": name})
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
