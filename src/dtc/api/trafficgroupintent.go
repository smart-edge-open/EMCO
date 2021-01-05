// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/open-ness/EMCO/src/dtc/pkg/module"
	orcmod "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"
	pkgerrors "github.com/pkg/errors"
)

var TrGroupIntJSONFile string = "json-schemas/metadata.json"

type trafficgroupintentHandler struct {
	client module.TrafficGroupIntentManager
}

// Check for valid format of input parameters
func validateTrafficGroupIntentInputs(tgi module.TrafficGroupIntent) error {
	// validate metadata
	err := module.IsValidMetadata(tgi.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid traffic group intent metadata")
	}
	return nil
}

func (h trafficgroupintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var tgi module.TrafficGroupIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deploymentIntentGroupName := vars["deployment-intent-group-name"]

	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		log.Error(":: Error validating traffic group POST parameters::", log.Fields{"Error": err})
		http.Error(w, "Invalid parameters", http.StatusNotFound)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&tgi)
	switch {
	case err == io.EOF:
		log.Error(":: Empty traffic group POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding traffic group POST body ::", log.Fields{"Error": err})

		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err, httpError := validation.ValidateJsonSchemaData(TrGroupIntJSONFile, tgi)
	if err != nil {
		log.Error(":: Error validating traffic group POST data ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if tgi.Metadata.Name == "" {
		log.Error(":: Missing name in traffic group POST request ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateTrafficGroupIntentInputs(tgi)
	if err != nil {
		log.Error(":: Invalid create traffic group body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateTrafficGroupIntent(tgi, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, false)
	if err != nil {
		log.Error(":: Error creating traffic group ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "already exists") {
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
		log.Error(":: Error encoding create traffic group response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
func (h trafficgroupintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var tgi module.TrafficGroupIntent
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]

	err := json.NewDecoder(r.Body).Decode(&tgi)

	switch {
	case err == io.EOF:
		log.Error(":: Empty traffic group PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding traffic group PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if tgi.Metadata.Name == "" {
		log.Error(":: Missing name in traffic group PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if tgi.Metadata.Name != name {
		log.Error(":: Mismatched name in traffic group PUT request ::", log.Fields{})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateTrafficGroupIntentInputs(tgi)
	if err != nil {
		log.Error(":: Invalid traffic group PUT inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateTrafficGroupIntent(tgi, project, compositeApp, compositeAppVersion, deployIntentGroup, true)
	if err != nil {
		log.Error(":: Error updating traffic group ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "already exists") {
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
		log.Error(":: Error encoding traffic group update response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h trafficgroupintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetTrafficGroupIntents(project, compositeApp, compositeAppVersion, deployIntentGroup)
		if err != nil {
			log.Error(":: Error getting traffic group intents ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		ret, err = h.client.GetTrafficGroupIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup)
		if err != nil {
			log.Error(":: Error getting traffic group intent ::", log.Fields{"Error": err})
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
		log.Error(":: Error encoding get traffic group response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h trafficgroupintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]

	err := h.client.DeleteTrafficGroupIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup)
	if err != nil {
		log.Error(":: Error deleting traffic group ::", log.Fields{"Error": err, "Name": name})
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
