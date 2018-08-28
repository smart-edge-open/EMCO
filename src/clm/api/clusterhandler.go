// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	clusterPkg "github.com/open-ness/EMCO/src/clm/pkg/cluster"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"

	"github.com/gorilla/mux"
)

var cpJSONFile string = "json-schemas/metadata.json"
var ckvJSONFile string = "json-schemas/cluster-kv.json"
var clJSONFile string = "json-schemas/cluster-label.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type clusterHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client clusterPkg.ClusterManager
}

// Create handles creation of the ClusterProvider entry in the database
func (h clusterHandler) createClusterProviderHandler(w http.ResponseWriter, r *http.Request) {
	var p clusterPkg.ClusterProvider

	err := json.NewDecoder(r.Body).Decode(&p)

	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster provider POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster provider POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(cpJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster provider POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing name in cluster provider POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterProvider(p)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "ClusterProvider already exists") {
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
		log.Error(":: Error encoding create cluster provider response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular ClusterProvider Name
// Returns a ClusterProvider
func (h clusterHandler) getClusterProviderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetClusterProviders()
		if err != nil {
			log.Error(":: Error getting cluster providers ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		ret, err = h.client.GetClusterProvider(name)
		if err != nil {
			log.Error(":: Error getting cluster provider ::", log.Fields{"Error": err, "Name": name})
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
		log.Error(":: Error encoding get cluster provider response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular ClusterProvider  Name
func (h clusterHandler) deleteClusterProviderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	err := h.client.DeleteClusterProvider(name)
	if err != nil {
		log.Error(":: Error deleting cluster provider ::", log.Fields{"Error": err, "Name": name})
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

// Create handles creation of the Cluster entry in the database
func (h clusterHandler) createClusterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	var p clusterPkg.Cluster
	var q clusterPkg.ClusterContent

	// Implemenation using multipart form
	// Review and enable/remove at a later date
	// Set Max size to 16mb here
	err := r.ParseMultipartForm(16777216)
	if err != nil {
		log.Error(":: Error parsing cluster multipart form ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	err = json.NewDecoder(jsn).Decode(&p)
	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(cpJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	//Read the file section and ignore the header
	file, _, err := r.FormFile("file")
	if err != nil {
		log.Error(":: Error getting file section ::", log.Fields{"Error": err})
		http.Error(w, "Unable to process file", http.StatusUnprocessableEntity)
		return
	}

	defer file.Close()

	//Convert the file content to base64 for storage
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error(":: Error reading file section ::", log.Fields{"Error": err})
		http.Error(w, "Unable to read file", http.StatusUnprocessableEntity)
		return
	}

	q.Kubeconfig = base64.StdEncoding.EncodeToString(content)

	// Name is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing name in cluster POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateCluster(provider, p, q)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "ClusterProvider does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "Cluster already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	//	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create cluster response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Cluster Name
// Returns a Cluster
func (h clusterHandler) getClusterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	name := vars["name"]

	label := r.URL.Query().Get("label")
	if len(label) != 0 {
		ret, err := h.client.GetClustersWithLabel(provider, label)
		if err != nil {
			log.Error(":: Error getting clusters by label ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
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
			log.Error(":: Error encoding get clusters by label response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// handle the get all clusters case - return a list of only the json parts
	if len(name) == 0 {
		var retList []clusterPkg.Cluster

		ret, err := h.client.GetClusters(provider)
		if err != nil {
			log.Error(":: Error getting clusters ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		for _, cl := range ret {
			retList = append(retList, clusterPkg.Cluster{Metadata: cl.Metadata})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(retList)
		if err != nil {
			log.Error(":: Error encoding get clusters ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	accepted, _, err := mime.ParseMediaType(r.Header.Get("Accept"))
	if err != nil {
		log.Error(":: Missing Accept header in get cluster request ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	retCluster, err := h.client.GetCluster(provider, name)
	if err != nil {
		log.Error(":: Error getting cluster ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "db Find error") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	retKubeconfig, err := h.client.GetClusterContent(provider, name)
	if err != nil {
		log.Error(":: Error getting cluster content ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	switch accepted {
	case "multipart/form-data":
		mpw := multipart.NewWriter(w)
		w.Header().Set("Content-Type", mpw.FormDataContentType())
		w.WriteHeader(http.StatusOK)
		pw, err := mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/json"}, "Content-Disposition": {"form-data; name=metadata"}})
		if err != nil {
			log.Error(":: Error creating metadata part of cluster response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(pw).Encode(retCluster); err != nil {
			log.Error(":: Error encoding cluster response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pw, err = mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/octet-stream"}, "Content-Disposition": {"form-data; name=file; filename=kubeconfig"}})
		if err != nil {
			log.Error(":: Error creating file part of cluster response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		kcBytes, err := base64.StdEncoding.DecodeString(retKubeconfig.Kubeconfig)
		if err != nil {
			log.Error(":: Error encoding file part of cluster response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = pw.Write(kcBytes)
		if err != nil {
			log.Error(":: Error writing multipart cluster response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(retCluster)
		if err != nil {
			log.Error(":: Error encoding cluster response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/octet-stream":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		kcBytes, err := base64.StdEncoding.DecodeString(retKubeconfig.Kubeconfig)
		if err != nil {
			log.Error(":: Error encoding file part of cluster response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(kcBytes)
		if err != nil {
			log.Error(":: Error writing cluster response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		log.Error(":: Missing Accept header for get cluster ::", log.Fields{"Error": err})
		http.Error(w, "set Accept: multipart/form-data, application/json or application/octet-stream", http.StatusMultipleChoices)
		return
	}
}

// Delete handles DELETE operations on a particular Cluster Name
func (h clusterHandler) deleteClusterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	name := vars["name"]

	err := h.client.DeleteCluster(provider, name)
	if err != nil {
		log.Error(":: Error deleting cluster ::", log.Fields{"Error": err})
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

// Create handles creation of the ClusterLabel entry in the database
func (h clusterHandler) createClusterLabelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	var p clusterPkg.ClusterLabel

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster label POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster label POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(clJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster label POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// LabelName is required.
	if p.LabelName == "" {
		log.Error(":: Missing cluster label name in POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing label name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterLabel(provider, cluster, p)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Cluster does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "Cluster Label already exists") {
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
		log.Error(":: Error encoding cluster label response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Cluster Label
// Returns a ClusterLabel
func (h clusterHandler) getClusterLabelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	label := vars["label"]

	var ret interface{}
	var err error

	if len(label) == 0 {
		ret, err = h.client.GetClusterLabels(provider, cluster)
		if err != nil {
			log.Error(":: Error getting cluster labels ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		ret, err = h.client.GetClusterLabel(provider, cluster, label)
		if err != nil {
			log.Error(":: Error getting cluster label ::", log.Fields{"Error": err})
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
		log.Error(":: Error encoding cluster label response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular ClusterLabel Name
func (h clusterHandler) deleteClusterLabelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	label := vars["label"]

	err := h.client.DeleteClusterLabel(provider, cluster, label)
	if err != nil {
		log.Error(":: Error deleting cluster label ::", log.Fields{"Error": err})
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

// Create handles creation of the ClusterKvPairs entry in the database
func (h clusterHandler) createClusterKvPairsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	var p clusterPkg.ClusterKvPairs

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster kv pair POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster kv pair POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(ckvJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster kv pair POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// KvPairsName is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing cluster kv pair name in POST body ::", log.Fields{"Error": err})
		http.Error(w, "Missing Key Value pair name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterKvPairs(provider, cluster, p)
	if err != nil {
		log.Error(":: Error creating cluster kv pair ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "Cluster does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "Cluster KV Pair already exists") {
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
		log.Error(":: Error encoding cluster kv pair ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Cluster Key Value Pair
// Returns a ClusterKvPairs
func (h clusterHandler) getClusterKvPairsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	kvpair := vars["kvpair"]

	var ret interface{}
	var err error

	if len(kvpair) == 0 {
		ret, err = h.client.GetAllClusterKvPairs(provider, cluster)
		if err != nil {
			log.Error(":: Error getting cluster kv pairs ::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		ret, err = h.client.GetClusterKvPairs(provider, cluster, kvpair)
		if err != nil {
			log.Error(":: Error getting cluster kv pair ::", log.Fields{"Error": err})
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
		log.Error(":: Error encoding cluster kv pair response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular Cluster Name
func (h clusterHandler) deleteClusterKvPairsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	kvpair := vars["kvpair"]

	err := h.client.DeleteClusterKvPairs(provider, cluster, kvpair)
	if err != nil {
		log.Error(":: Error deleting cluster kv pair ::", log.Fields{"Error": err})
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
