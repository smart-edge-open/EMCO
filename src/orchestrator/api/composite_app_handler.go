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

var caJSONFile string = "json-schemas/composite-app.json"

// compositeAppHandler to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type compositeAppHandler struct {
	// Interface that implements CompositeApp operations
	// We will set this variable with a mock interface for testing
	client moduleLib.CompositeAppManager
}

// createHandler handles creation of the CompositeApp entry in the database
func (h compositeAppHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var c moduleLib.CompositeApp

	err := json.NewDecoder(r.Body).Decode(&c)
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
	err, httpError := validation.ValidateJsonSchemaData(caJSONFile, c)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["project-name"]

	ret, err := h.client.CreateCompositeApp(c, projectName, false)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Unable to find the project") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "CompositeApp already exists") {
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
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler handles GET operations on a particular CompositeApp Name
// Returns a compositeApp
func (h compositeAppHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["composite-app-name"]
	version := vars["version"]
	projectName := vars["project-name"]

	ret, err := h.client.GetCompositeApp(name, version, projectName)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "db Find error") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getAllCompositeAppsHandler handles the GetAllComppositeApps, returns a list of compositeApps under a project
func (h compositeAppHandler) getAllCompositeAppsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pName := vars["project-name"]

	var caList []moduleLib.CompositeApp

	cApps, err := h.client.GetAllCompositeApps(pName)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Unable to find") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "db Find error") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	for _, cApp := range cApps {
		caList = append(caList, moduleLib.CompositeApp{Metadata: cApp.Metadata, Spec: cApp.Spec})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(caList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	return
}

// deleteHandler handles DELETE operations on a particular CompositeApp Name
func (h compositeAppHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["composite-app-name"]
	version := vars["version"]
	projectName := vars["project-name"]

	err := h.client.DeleteCompositeApp(name, version, projectName)
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

func (h compositeAppHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	var c moduleLib.CompositeApp

	err := json.NewDecoder(r.Body).Decode(&c)
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
	err, httpError := validation.ValidateJsonSchemaData(caJSONFile, c)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["project-name"]

	ret, err := h.client.CreateCompositeApp(c, projectName, true)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Unable to find the project") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "CompositeApp already exists") {
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
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
