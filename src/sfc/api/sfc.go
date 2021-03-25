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

var sfcJSONFile string = "json-schemas/sfc.json"

// Create handles creation of the SFC entry in the database
func (h sfcIntentHandler) createSfcHandler(w http.ResponseWriter, r *http.Request) {
	var sfc model.SfcIntent
	vars := mux.Vars(r)
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := json.NewDecoder(r.Body).Decode(&sfc)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcJSONFile, sfc)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateSfcIntent(sfc, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, false)
	if err != nil {
		log.Error(":: Error creating SFC Intent ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "SFC Intent already exists") {
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
		log.Error(":: Error encoding create SFC Intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles update of the SFC entry in the database
func (h sfcIntentHandler) putSfcHandler(w http.ResponseWriter, r *http.Request) {
	var sfc model.SfcIntent
	vars := mux.Vars(r)
	name := vars["sfc-name"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := json.NewDecoder(r.Body).Decode(&sfc)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC PUT body ::", log.Fields{"Error": err, "Body": sfc})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcJSONFile, sfc)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name in URL should match name in body
	if sfc.Metadata.Name != name {
		log.Error(":: Mismatched SFC name in PUT request ::", log.Fields{"URL name": name, "Metadata name": sfc.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateSfcIntent(sfc, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, true)
	if err != nil {
		log.Error(":: Error updating SFC ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "SFC already exists") {
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
		log.Error(":: Error encoding update SFC response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular SFC Name
// Returns an SfcIntent
func (h sfcIntentHandler) getSfcHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["sfc-name"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetAllSfcIntents(project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
		if err != nil {
			log.Error(":: Error getting SFC Intents ::", log.Fields{"Error": err})
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
		ret, err = h.client.GetSfcIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
		if err != nil {
			log.Error(":: Error getting SFC Intent ::", log.Fields{"Error": err})
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
		log.Error(":: Error encoding get SFC Intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular SFC
func (h sfcIntentHandler) deleteSfcHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["sfc-name"]
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := h.client.DeleteSfcIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
	if err != nil {
		log.Error(":: Error deleting SFC ::", log.Fields{"Error": err, "Name": name})
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
