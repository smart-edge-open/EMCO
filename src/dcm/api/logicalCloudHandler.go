// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	dcm "github.com/open-ness/EMCO/src/dcm/pkg/module"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	orch "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
)

// logicalCloudHandler is used to store backend implementations objects
type logicalCloudHandler struct {
	client               dcm.LogicalCloudManager
	clusterClient        dcm.ClusterManager
	quotaClient          dcm.QuotaManager
	userPermissionClient dcm.UserPermissionManager
}

// CreateHandler handles the creation of a logical cloud
func (h logicalCloudHandler) createHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	project := vars["project-name"]
	var v dcm.LogicalCloud

	err := json.NewDecoder(r.Body).Decode(&v)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Logical Cloud Name is required.
	if v.MetaData.LogicalCloudName == "" {
		msg := "Missing name in POST request"
		log.Error(msg, log.Fields{})
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Validate that the specified Project exists
	// before associating a Logical Cloud with it
	p := orch.NewProjectClient()
	_, err = p.GetProject(project)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	ret, err := h.client.Create(project, v)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Unable to find the project") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "Logical Cloud already exists") {
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

// getAllHandler handles GET operations over logical clouds
// Returns a list of Logical Clouds
func (h logicalCloudHandler) getAllHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	var ret interface{}
	var err error

	ret, err = h.client.GetAll(project)
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

// getHandler handles GET operations on a particular name
// Returns a Logical Cloud
func (h logicalCloudHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	name := vars["logical-cloud-name"]
	var ret interface{}
	var err error

	ret, err = h.client.Get(project, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Logical Cloud does not exist") {
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

// updateHandler handles Update operations on a particular logical cloud
func (h logicalCloudHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	var v dcm.LogicalCloud
	vars := mux.Vars(r)
	project := vars["project-name"]
	name := vars["logical-cloud-name"]

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

	if v.MetaData.LogicalCloudName == "" {
		log.Error("API: Missing name in PUT request", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.Update(project, name, v)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if err.Error() == "Logical Cloud does not exist" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// deleteHandler handles Delete operations on a particular logical cloud
func (h logicalCloudHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	name := vars["logical-cloud-name"]

	err := h.client.Delete(project, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Logical Cloud does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "The Logical Cloud can't be deleted yet, it is being terminated") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "The Logical Cloud is instantiated, please terminate first") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "The Logical Cloud is instantiating, please wait and then terminate") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "The Logical Cloud has failed instantiating, for safety please terminate and try again") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// instantiateHandler handles instantiateing a particular logical cloud
func (h logicalCloudHandler) instantiateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	name := vars["logical-cloud-name"]

	// Get logical cloud
	lc, err := h.client.Get(project, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Logical Cloud does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get Clusters
	clusters, err := h.clusterClient.GetAllClusters(project, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "No Cluster References associated") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get Quotas
	quotas, err := h.quotaClient.GetAllQuotas(project, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userPermissions, err := h.userPermissionClient.GetAllUserPerms(project, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Instantiate the Logical Cloud
	err = dcm.Instantiate(project, lc, clusters, quotas, userPermissions)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "The Logical Cloud can't be re-instantiated yet, it is being terminated") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "The Logical Cloud is already instantiated") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "The Logical Cloud is already instantiating") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "The Logical Cloud has failed instantiating before, please terminate and try again") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "The Logical Cloud has failed terminating, please try to terminate again or delete the Logical Cloud") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "The clusters associated to this L0 logical cloud don't all share the same namespace name") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "Level-1 Logical Clouds require a Quota to be associated first") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "Level-1 Logical Clouds require a User Permission assigned to its primary namespace") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if strings.Contains(err.Error(), "It looks like the cluster provided as reference does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// terminateHandler handles terminating a particular logical cloud
func (h logicalCloudHandler) terminateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	name := vars["logical-cloud-name"]

	// Get logical cloud
	lc, err := h.client.Get(project, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Logical Cloud does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get Clusters
	clusters, err := h.clusterClient.GetAllClusters(project, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "No Cluster References associated") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get Quotas
	quotas, err := h.quotaClient.GetAllQuotas(project, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Terminate the Logical Cloud
	err = dcm.Terminate(project, lc, clusters, quotas)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "The Logical Cloud is not instantiated") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "The Logical Cloud has already been terminated") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "The Logical Cloud is already being terminated") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if strings.Contains(err.Error(), "The Logical Cloud is still instantiating") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// stopHandler handles aborting the pending instantiation or termination of a logical cloud
func (h logicalCloudHandler) stopHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	name := vars["logical-cloud-name"]

	// Get logical cloud
	lc, err := h.client.Get(project, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Logical Cloud does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Stop the instantiation
	err = dcm.Stop(project, lc)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if err.Error() == "Logical Clouds can't be stopped" {
			http.Error(w, err.Error(), http.StatusNotImplemented)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
