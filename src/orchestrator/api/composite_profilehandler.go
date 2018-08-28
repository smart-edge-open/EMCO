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
	moduleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
)

var caprofileJSONFile string = "json-schemas/metadata.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type compositeProfileHandler struct {
	client moduleLib.CompositeProfileManager
}

// createCompositeProfileHandler handles the create operation of intent
func (h compositeProfileHandler) createHandler(w http.ResponseWriter, r *http.Request) {

	var cpf moduleLib.CompositeProfile

	err := json.NewDecoder(r.Body).Decode(&cpf)
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
	err, httpError := validation.ValidateJsonSchemaData(caprofileJSONFile, cpf)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["project-name"]
	compositeAppName := vars["composite-app-name"]
	version := vars["composite-app-version"]

	cProf, createErr := h.client.CreateCompositeProfile(cpf, projectName, compositeAppName, version)
	if createErr != nil {
		log.Error(createErr.Error(), log.Fields{})
		if strings.Contains(createErr.Error(), "Unable to find the project") {
			http.Error(w, createErr.Error(), http.StatusNotFound)
		} else if strings.Contains(createErr.Error(), "Unable to find the composite-app") {
			http.Error(w, createErr.Error(), http.StatusNotFound)
		} else if strings.Contains(createErr.Error(), "CompositeProfile already exists") {
			http.Error(w, createErr.Error(), http.StatusConflict)
		} else {
			http.Error(w, createErr.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(cProf)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler handles the GET operations on CompositeProfile
func (h compositeProfileHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cProfName := vars["composite-profile-name"]

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

	// handle the get all composite profile case
	if len(cProfName) == 0 {
		var retList []moduleLib.CompositeProfile

		ret, err := h.client.GetCompositeProfiles(projectName, compositeAppName, version)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		for _, cl := range ret {
			retList = append(retList, moduleLib.CompositeProfile{Metadata: cl.Metadata})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(retList)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	cProf, err := h.client.GetCompositeProfile(cProfName, projectName, compositeAppName, version)
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
	err = json.NewEncoder(w).Encode(cProf)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// deleteHandler handles the delete operations on CompostiteProfile
func (h compositeProfileHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	c := vars["composite-profile-name"]
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]

	err := h.client.DeleteCompositeProfile(c, p, ca, v)
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
