package api

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/json"
	"strings"

	"io"
	"net/http"

	moduleLib "github.com/open-ness/EMCO/src/genericactioncontroller/pkg/module"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"
	pkgerrors "github.com/pkg/errors"

	"github.com/gorilla/mux"
)

var genericK8sIntentFile string = "json-schemas/metadata.json"

type generick8sintentHandler struct {
	client moduleLib.GenericK8sIntentManager
}

func validateGenericK8sIntentInputs(gki moduleLib.GenericK8sIntent) error {
	// validate metadata
	err := moduleLib.IsValidMetadata(gki.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid GenericK8sIntent metadata")
	}
	return nil
}

// createHandler handles creation of the GenericK8sIntent entry in the database
func (g generick8sintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var gki moduleLib.GenericK8sIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]

	err := json.NewDecoder(r.Body).Decode(&gki)

	switch {
	case err == io.EOF:
		log.Error(":: Empty generick8sIntent POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding generick8sIntent POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(genericK8sIntentFile, gki)
	if err != nil {
		log.Error(":: JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if gki.Metadata.Name == "" {
		log.Error(":: Missing name in POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing genericK8sIntentName in POST request", http.StatusBadRequest)
		return
	}

	err = validateGenericK8sIntentInputs(gki)
	if err != nil {
		log.Error(":: validateGenericK8sIntentInputs error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := g.client.CreateGenericK8sIntent(gki, project, compositeApp, compositeAppVersion, deployIntentGroup, false)
	if err != nil {
		log.Error(":: CreateGenericK8sIntent error ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "GenericK8sIntent already exists") {
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
		log.Error(":: GenericK8sIntent Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (g generick8sintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var gki moduleLib.GenericK8sIntent
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]

	err := json.NewDecoder(r.Body).Decode(&gki)

	switch {
	case err == io.EOF:
		log.Error(":: Empty genericK8sIntent body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding resource body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is mandatory.
	if gki.Metadata.Name == "" {
		log.Error(":: Missing name in POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if gki.Metadata.Name != name {

		log.Error(":: Mismatched name in PUT request ::", log.Fields{"bodyname": gki.Metadata.Name, "name": name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(genericK8sIntentFile, gki)
	if err != nil {
		log.Error(":: JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if gki.Metadata.Name == "" {
		log.Error(":: Missing genericK8sIntentName in POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing genericK8sIntentName in POST request", http.StatusBadRequest)
		return
	}

	err = validateGenericK8sIntentInputs(gki)
	if err != nil {
		log.Error(":: validateGenericK8sIntentInputs failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := g.client.CreateGenericK8sIntent(gki, project, compositeApp, compositeAppVersion, deployIntentGroup, true)
	if err != nil {
		log.Error(":: CreateGenericK8sIntent failure ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "GenericK8sIntent already exists") {
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
		log.Error(":: GenericK8sIntent encoding failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular GenericK8sIntent Name
// Returns a GenericK8sIntent
func (g generick8sintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = g.client.GetAllGenericK8sIntents(project, compositeApp, compositeAppVersion, deployIntentGroup)
		if err != nil {
			log.Error(":: GetAllGenericK8sIntents failure ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		ret, err = g.client.GetGenericK8sIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup)
		if err != nil {
			log.Error(":: GetGenericK8sIntent failure ::", log.Fields{"Error": err})
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
		log.Error(":: GenericK8sIntent encoding failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (g generick8sintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]

	err := g.client.DeleteGenericK8sIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup)
	if err != nil {
		log.Error(":: DeleteGenericK8sIntent failure ::", log.Fields{"Error": err})
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
