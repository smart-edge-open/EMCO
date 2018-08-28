// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"
	moduleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

var gpiJSONFile string = "json-schemas/generic-placement-intent.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type genericPlacementIntentHandler struct {
	client moduleLib.GenericPlacementIntentManager
}

// createGenericPlacementIntentHandler handles the create operation of intent
func (h genericPlacementIntentHandler) createGenericPlacementIntentHandler(w http.ResponseWriter, r *http.Request) {

	var g moduleLib.GenericPlacementIntent

	err := json.NewDecoder(r.Body).Decode(&g)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(gpiJSONFile, g)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["project-name"]
	compositeAppName := vars["composite-app-name"]
	version := vars["composite-app-version"]
	digName := vars["deployment-intent-group-name"]

	gPIntent, createErr := h.client.CreateGenericPlacementIntent(g, projectName, compositeAppName, version, digName)
	if createErr != nil {
		log.Error(createErr.Error(), log.Fields{})
		if strings.Contains(createErr.Error(), "Unable to find the project") {
			http.Error(w, createErr.Error(), http.StatusNotFound)
		} else if strings.Contains(createErr.Error(), "Unable to find the composite-app") {
			http.Error(w, createErr.Error(), http.StatusNotFound)
		} else if strings.Contains(createErr.Error(), "Unable to find the deployment-intent-group-name") {
			http.Error(w, createErr.Error(), http.StatusNotFound)
		} else if strings.Contains(createErr.Error(), "Intent already exists") {
			http.Error(w, createErr.Error(), http.StatusConflict)
		} else {
			http.Error(w, createErr.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(gPIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getGenericPlacementHandler handles the GET operations on intent
func (h genericPlacementIntentHandler) getGenericPlacementHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	intentName := vars["intent-name"]
	if intentName == "" {
		log.Error("Missing genericPlacementIntentName in GET request", log.Fields{})
		http.Error(w, "Missing genericPlacementIntentName in GET request", http.StatusBadRequest)
		return
	}
	projectName := vars["project-name"]
	if projectName == "" {
		log.Error("Missing projectName in GET request", log.Fields{})
		http.Error(w, "Missing projectName in GET request", http.StatusBadRequest)
		return
	}
	compositeAppName := vars["composite-app-name"]
	if compositeAppName == "" {
		log.Error("Missing compositeAppName in GET request", log.Fields{})
		http.Error(w, "Missing compositeAppName in GET request", http.StatusBadRequest)
		return
	}

	version := vars["composite-app-version"]
	if version == "" {
		log.Error("Missing version in GET request", log.Fields{})
		http.Error(w, "Missing version in GET request", http.StatusBadRequest)
		return
	}

	dig := vars["deployment-intent-group-name"]
	if dig == "" {
		log.Error("Missing deploymentIntentGroupName in GET request", log.Fields{})
		http.Error(w, "Missing deploymentIntentGroupName in GET request", http.StatusBadRequest)
		return
	}

	gPIntent, err := h.client.GetGenericPlacementIntent(intentName, projectName, compositeAppName, version, dig)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "db Find error") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(gPIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h genericPlacementIntentHandler) getAllGenericPlacementIntentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pList := []string{"project-name", "composite-app-name", "composite-app-version"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	digName := vars["deployment-intent-group-name"]

	gpList, err := h.client.GetAllGenericPlacementIntents(p, ca, v, digName)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Unable to find") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "db Find error") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(gpList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// deleteGenericPlacementHandler handles the delete operations on intent
func (h genericPlacementIntentHandler) deleteGenericPlacementHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	i := vars["intent-name"]
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	digName := vars["deployment-intent-group-name"]

	err := h.client.DeleteGenericPlacementIntent(i, p, ca, v, digName)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
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
