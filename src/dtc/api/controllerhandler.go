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
	controller "github.com/open-ness/EMCO/src/orchestrator/pkg/module/controller"
)

var controllerJSONFile string = "json-schemas/controller.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type controllerHandler struct {
	// Interface that implements controller operations
	// We will set this variable with a mock interface for testing
	client controller.ControllerManager
}

// Create handles creation of the controller entry in the database
func (h controllerHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var m controller.Controller

	err := json.NewDecoder(r.Body).Decode(&m)
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
	err, httpError := validation.ValidateJsonSchemaData(controllerJSONFile, m)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateController(m, false)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusConflict)
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

// Put handles creation or update of the controller entry in the database
func (h controllerHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var m controller.Controller
	vars := mux.Vars(r)
	name := vars["controller-name"]

	err := json.NewDecoder(r.Body).Decode(&m)
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
	err, httpError := validation.ValidateJsonSchemaData(controllerJSONFile, m)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// name in URL should match name in body
	if m.Metadata.Name != name {
		log.Error("Mismatched name in PUT request", log.Fields{})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateController(m, true)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusConflict)
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

// Get handles GET operations on a particular controller Name
// Returns a controller
func (h controllerHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["controller-name"]
	var ret interface{}
	var err error

	// handle the get all controllers case
	if len(name) == 0 {
		ret, err = h.client.GetControllers()
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {

		ret, err = h.client.GetController(name)
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

// Delete handles DELETE operations on a particular controller Name
func (h controllerHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["controller-name"]

	err := h.client.DeleteController(name)
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
