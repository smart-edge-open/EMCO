// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/open-ness/EMCO/src/dcm/pkg/module"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

// quotaHandler is used to store backend implementations objects
type quotaHandler struct {
	client module.QuotaManager
}

// CreateHandler handles creation of the quota entry in the database
func (h quotaHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	logicalCloud := vars["logical-cloud-name"]
	var v module.Quota

	err := json.NewDecoder(r.Body).Decode(&v)
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

	// Quota Name is required.
	if v.MetaData.QuotaName == "" {
		msg := "Missing name in POST request"
		log.Error(msg, log.Fields{})
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateQuota(project, logicalCloud, v)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Logical Cloud does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "Cluster Quotas not allowed for Logical Cloud Level 0") {
			http.Error(w, err.Error(), http.StatusBadRequest)
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

// getHandler handles GET operations over quotas
// Returns a list of Quotas
func (h quotaHandler) getAllHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	logicalCloud := vars["logical-cloud-name"]
	var ret interface{}
	var err error

	ret, err = h.client.GetAllQuotas(project, logicalCloud)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// getHandler handle GET operations on a particular name
// Returns a Quota
func (h quotaHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	logicalCloud := vars["logical-cloud-name"]
	name := vars["quota-name"]
	var ret interface{}
	var err error

	ret, err = h.client.GetQuota(project, logicalCloud, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if err.Error() == "Cluster Quota does not exist" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// UpdateHandler handles Update operations on a particular quota
func (h quotaHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	var v module.Quota
	vars := mux.Vars(r)
	project := vars["project-name"]
	logicalCloud := vars["logical-cloud-name"]
	name := vars["quota-name"]

	err := json.NewDecoder(r.Body).Decode(&v)
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

	// Name is required.
	if v.MetaData.QuotaName == "" {
		log.Error("API: Missing name in PUT request", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.UpdateQuota(project, logicalCloud, name, v)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if err.Error() == "Cluster Quota does not exist" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
}

//deleteHandler handles DELETE operations on a particular record
func (h quotaHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	logicalCloud := vars["logical-cloud-name"]
	name := vars["quota-name"]

	err := h.client.DeleteQuota(project, logicalCloud, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
