// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"
	moduleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
)

var appIntentJSONFile string = "json-schemas/generic-placement-intent-app.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type appIntentHandler struct {
	client moduleLib.AppIntentManager
}

// createAppIntentHandler handles the create operation of intent
func (h appIntentHandler) createAppIntentHandler(w http.ResponseWriter, r *http.Request) {

	var a moduleLib.AppIntent

	err := json.NewDecoder(r.Body).Decode(&a)
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
	err, httpError := validation.ValidateJsonSchemaData(appIntentJSONFile, a)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["project-name"]
	compositeAppName := vars["composite-app-name"]
	version := vars["composite-app-version"]
	intent := vars["intent-name"]
	digName := vars["deployment-intent-group-name"]

	appIntent, createErr := h.client.CreateAppIntent(a, projectName, compositeAppName, version, intent, digName)
	if createErr != nil {
		log.Error(createErr.Error(), log.Fields{})
		if strings.Contains(createErr.Error(), "Unable to find the project") {
			http.Error(w, createErr.Error(), http.StatusNotFound)
		} else if strings.Contains(createErr.Error(), "Unable to find the composite-app") {
			http.Error(w, createErr.Error(), http.StatusNotFound)
		} else if strings.Contains(createErr.Error(), "Unable to find the intent") {
			http.Error(w, createErr.Error(), http.StatusNotFound)
		} else if strings.Contains(createErr.Error(), "Unable to find the deployment-intent-group-name") {
			http.Error(w, createErr.Error(), http.StatusNotFound)
		} else if strings.Contains(createErr.Error(), "AppIntent already exists") {
			http.Error(w, createErr.Error(), http.StatusConflict)
		} else {
			http.Error(w, createErr.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(appIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h appIntentHandler) getAppIntentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	p := vars["project-name"]
	if p == "" {
		log.Error("Missing projectName in GET request", log.Fields{})
		http.Error(w, "Missing projectName in GET request", http.StatusBadRequest)
		return
	}
	ca := vars["composite-app-name"]
	if ca == "" {
		log.Error("Missing compositeAppName in GET request", log.Fields{})
		http.Error(w, "Missing compositeAppName in GET request", http.StatusBadRequest)
		return
	}

	v := vars["composite-app-version"]
	if v == "" {
		log.Error("Missing version of compositeApp in GET request", log.Fields{})
		http.Error(w, "Missing version of compositeApp in GET request", http.StatusBadRequest)
		return
	}

	i := vars["intent-name"]
	if i == "" {
		log.Error("Missing genericPlacementIntentName in GET request", log.Fields{})
		http.Error(w, "Missing genericPlacementIntentName in GET request", http.StatusBadRequest)
		return
	}

	dig := vars["deployment-intent-group-name"]
	if dig == "" {
		log.Error("Missing deploymentIntentGroupName in GET request", log.Fields{})
		http.Error(w, "Missing deploymentIntentGroupName in GET request", http.StatusBadRequest)
		return
	}

	ai := vars["app-intent-name"]
	if ai == "" {
		log.Error("Missing appIntentName in GET request", log.Fields{})
		http.Error(w, "Missing appIntentName in GET request", http.StatusBadRequest)
		return
	}

	appIntent, err := h.client.GetAppIntent(ai, p, ca, v, i, dig)
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
	err = json.NewEncoder(w).Encode(appIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

/*
getAllIntentsByAppHandler handles the URL:
/v2/project/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/generic-placement-intent/{intent-name}/app-intents?app-name=<app-name>
*/
func (h appIntentHandler) getAllIntentsByAppHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	pList := []string{"project-name", "composite-app-name", "composite-app-version", "intent-name"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	i := vars["intent-name"]
	digName := vars["deployment-intent-group-name"]

	aN := r.URL.Query().Get("app-name")
	if aN == "" {
		log.Error("Missing appName in GET request", log.Fields{})
		http.Error(w, "Missing appName in GET request", http.StatusBadRequest)
		return
	}

	specData, err := h.client.GetAllIntentsByApp(aN, p, ca, v, i, digName)
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
	err = json.NewEncoder(w).Encode(specData)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return

}

/*
getAllAppIntentsHandler handles the URL:
/v2/project/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/generic-placement-intent/{intent-name}/app-intents
*/
func (h appIntentHandler) getAllAppIntentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pList := []string{"project-name", "composite-app-name", "composite-app-version", "intent-name"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	i := vars["intent-name"]
	digName := vars["deployment-intent-group-name"]

	applicationsAndClusterInfo, err := h.client.GetAllAppIntents(p, ca, v, i, digName)
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
	err = json.NewEncoder(w).Encode(applicationsAndClusterInfo)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return

}

func (h appIntentHandler) deleteAppIntentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	i := vars["intent-name"]
	ai := vars["app-intent-name"]
	digName := vars["deployment-intent-group-name"]

	err := h.client.DeleteAppIntent(ai, p, ca, v, i, digName)
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
