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

var inClientsIntJSONFile string = "json-schemas/inbound-clients.json"

type inboundclientsintentHandler struct {
	client module.InboundClientsIntentManager
}

// Check for valid format of input parameters
func validateInboundClientsIntentInputs(ici module.InboundClientsIntent) error {
	// validate metadata
	err := module.IsValidMetadata(ici.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid inbound clients intent metadata")
	}
	return nil
}

func (h inboundclientsintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var ici module.InboundClientsIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deploymentIntentGroupName := vars["deployment-intent-group-name"]
	trafficIntentGroupName := vars["traffic-group-intent-name"]
	inboundIntentName := vars["intent-name"]
	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		log.Error(":: Error validating inbound clients POST parameters::", log.Fields{"Error": err})
		http.Error(w, "DeploymentIntentGroup does not exist", http.StatusNotFound)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&ici)

	switch {
	case err == io.EOF:
		log.Error(":: Empty inbound clients POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding inbound clients POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(inClientsIntJSONFile, ici)
	if err != nil {
		log.Error(":: Error validating inbound clients POST data ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if ici.Metadata.Name == "" {
		log.Error(":: Missing name in inbound clients POST request ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateInboundClientsIntentInputs(ici)
	if err != nil {
		log.Error(":: Invalid create inbound clients body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClientsInboundIntent(ici, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName, false)
	if err != nil {
		log.Error(":: Error creating inboud clients ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "already exists") {
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
		log.Error(":: Error encoding create inbound clients response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
func (h inboundclientsintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var ici module.InboundClientsIntent
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deploymentIntentGroupName := vars["deployment-intent-group-name"]
	trafficIntentGroupName := vars["traffic-group-intent-name"]
	inboundIntentName := vars["intent-name"]

	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		log.Error(":: Error validating inbound clients PUT parameters::", log.Fields{"Error": err})
		http.Error(w, "DeploymentIntentGroup does not exist", http.StatusNotFound)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&ici)

	switch {
	case err == io.EOF:
		log.Error(":: Empty inbound clients PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding inbound clients PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if ici.Metadata.Name == "" {
		log.Error(":: Missing name in inbound clients PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if ici.Metadata.Name != name {
		log.Error(":: Mismatched name in inbound clients PUT request ::", log.Fields{})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateInboundClientsIntentInputs(ici)
	if err != nil {
		log.Error(":: Invalid inbound clients PUT inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClientsInboundIntent(ici, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName, true)
	if err != nil {
		log.Error(":: Error updating inbound clients ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "already exists") {
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
		log.Error(":: Error encoding inbound clients update response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h inboundclientsintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deploymentIntentGroupName := vars["deployment-intent-group-name"]
	trafficIntentGroupName := vars["traffic-group-intent-name"]
	inboundIntentName := vars["intent-name"]

	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetClientsInboundIntents(project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName)
		if err != nil {
			log.Error(":: Error getting inbound clients intents ::", log.Fields{"Error": err})
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
	} else {
		ret, err = h.client.GetClientsInboundIntent(name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName)
		if err != nil {
			log.Error(":: Error getting inbound clients intent ::", log.Fields{"Error": err})
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
		log.Error(":: Error encoding get inbound clients response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h inboundclientsintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deploymentIntentGroupName := vars["deployment-intent-group-name"]
	trafficIntentGroupName := vars["traffic-group-intent-name"]
	inboundIntentName := vars["intent-name"]

	err := h.client.DeleteClientsInboundIntent(name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName)
	if err != nil {
		log.Error(":: Error deleting inbound clients ::", log.Fields{"Error": err, "Name": name})
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
