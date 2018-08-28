// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"github.com/open-ness/EMCO/src/clm/pkg/cluster"
)

// Client for using the services in the ncm
type Client struct {
	Cluster *cluster.ClusterClient
	// Add Clients for API's here
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.Cluster = cluster.NewClusterClient()
	// Add Client API handlers here
	return c
}
