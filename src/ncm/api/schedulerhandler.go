// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/open-ness/EMCO/src/ncm/pkg/scheduler"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"

	"github.com/gorilla/mux"
)

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type schedulerHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client scheduler.SchedulerManager
}

//  applyClusterHandler handles requests to apply network intents for a cluster
func (h schedulerHandler) applySchedulerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["cluster-provider"]
	cluster := vars["cluster"]

	err := h.client.ApplyNetworkIntents(provider, cluster)
	if err != nil {
		log.Error(":: Error applying network intents ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

//  terminateSchedulerHandler handles requests to terminate network intents for a cluster
func (h schedulerHandler) terminateSchedulerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["cluster-provider"]
	cluster := vars["cluster"]

	err := h.client.TerminateNetworkIntents(provider, cluster)
	if err != nil {
		log.Error(":: Error terminating network intents ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

//  stopSchedulerHandler handles requests to stop instantiation or termination network intents for a cluster
func (h schedulerHandler) stopSchedulerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["cluster-provider"]
	cluster := vars["cluster"]

	err := h.client.StopNetworkIntents(provider, cluster)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

//  statusSchedulerHandler handles requests to query status of network intents for a cluster
func (h schedulerHandler) statusSchedulerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["cluster-provider"]
	cluster := vars["cluster"]

	qParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Error(":: Error parsing network status query parameters ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var queryInstance string
	if i, found := qParams["instance"]; found {
		queryInstance = i[0]
	} else {
		queryInstance = "" // default type
	}

	var queryType string
	if t, found := qParams["type"]; found {
		queryType = t[0]
		if queryType != "cluster" && queryType != "rsync" {
			log.Error(":: Invalid network status type query ::", log.Fields{})
			http.Error(w, "Invalid query type", http.StatusBadRequest)
			return
		}
	} else {
		queryType = "rsync" // default type
	}

	var queryOutput string
	if o, found := qParams["output"]; found {
		queryOutput = o[0]
		if queryOutput != "summary" && queryOutput != "all" && queryOutput != "detail" {
			log.Error(":: Invalid network status output query ::", log.Fields{})
			http.Error(w, "Invalid query output", http.StatusBadRequest)
			return
		}
	} else {
		queryOutput = "all" // default output format
	}

	var filterApps []string
	if a, found := qParams["app"]; found {
		filterApps = a
		for _, app := range filterApps {
			errs := validation.IsValidName(app)
			if len(errs) > 0 {
				log.Error(":: Invalid network status app query name ::", log.Fields{})
				http.Error(w, "Invalid app query", http.StatusBadRequest)
				return
			}
		}
	} else {
		filterApps = make([]string, 0)
	}

	var filterClusters []string
	if c, found := qParams["cluster"]; found {
		filterClusters = c
		for _, cl := range filterClusters {
			parts := strings.Split(cl, "+")
			if len(parts) != 2 {
				log.Error(":: Invalid network status cluster query format ::", log.Fields{})
				http.Error(w, "Invalid cluster query", http.StatusBadRequest)
				return
			}
			for _, p := range parts {
				errs := validation.IsValidName(p)
				if len(errs) > 0 {
					log.Error(":: Invalid network status cluster query name ::", log.Fields{"Error": errs})
					http.Error(w, "Invalid cluster query", http.StatusBadRequest)
					return
				}
			}
		}
	} else {
		filterClusters = make([]string, 0)
	}

	var filterResources []string
	if r, found := qParams["resource"]; found {
		filterResources = r
		for _, res := range filterResources {
			errs := validation.IsValidName(res)
			if len(errs) > 0 {
				log.Error(":: Invalid network status resource query name ::", log.Fields{"Error": errs})
				http.Error(w, "Invalid resources query", http.StatusBadRequest)
				return
			}
		}
	} else {
		filterResources = make([]string, 0)
	}

	status, iErr := h.client.NetworkIntentsStatus(provider, cluster, queryInstance, queryType, queryOutput, filterApps, filterClusters, filterResources)
	if iErr != nil {
		log.Error(":: Error getting network intent status ::", log.Fields{"Error": iErr})
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	iErr = json.NewEncoder(w).Encode(status)
	if iErr != nil {
		log.Error(":: Error encoding network intent status response ::", log.Fields{"Error": iErr})
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
}
