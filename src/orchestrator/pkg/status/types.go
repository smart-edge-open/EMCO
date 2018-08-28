// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package status

import (
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/state"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// StatusQueryParam defines the type of the query parameter
type StatusQueryParam = string
type queryparams struct {
	Instance StatusQueryParam // identify which AppContext to use - default is latest
	Summary  StatusQueryParam // only show high level summary
	All      StatusQueryParam // include basic resource information
	Detail   StatusQueryParam // show resource details
	Rsync    StatusQueryParam // select rsync (appcontext) data as source for query
	App      StatusQueryParam // filter results by specified app(s)
	Cluster  StatusQueryParam // filter results by specified cluster(s)
	Resource StatusQueryParam // filter results by specified resource(s)
}

// StatusQueryEnum defines the set of valid query parameter strings
var StatusQueryEnum = &queryparams{
	Instance: "instance",
	Summary:  "summary",
	All:      "all",
	Detail:   "detail",
	Rsync:    "rsync",
	App:      "app",
	Cluster:  "cluster",
	Resource: "resource",
}

type StatusResult struct {
	Name          string                 `json:"name,omitempty,inline"`
	State         state.StateInfo        `json:"states,omitempty,inline"`
	Status        appcontext.StatusValue `json:"status,omitempty,inline"`
	RsyncStatus   map[string]int         `json:"rsync-status,omitempty,inline"`
	ClusterStatus map[string]int         `json:"cluster-status,omitempty,inline"`
	Apps          []AppStatus            `json:"apps,omitempty,inline"`
}

type AppStatus struct {
	Name     string          `json:"name,omitempty"`
	Clusters []ClusterStatus `json:"clusters,omitempty"`
}

type ClusterStatus struct {
	ClusterProvider string           `json:"cluster-provider,omitempty"`
	Cluster         string           `json:"cluster,omitempty"`
	Resources       []ResourceStatus `json:"resources,omitempty"`
}

type ResourceStatus struct {
	Gvk           schema.GroupVersionKind `json:"GVK,omitempty"`
	Name          string                  `json:"name,omitempty"`
	Detail        interface{}             `json:"detail,omitempty"`
	RsyncStatus   string                  `json:"rsync-status,omitempty"`
	ClusterStatus string                  `json:"cluster-status,omitempty"`
}
