// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"github.com/open-ness/EMCO/src/clm/pkg/cluster"
	"github.com/open-ness/EMCO/src/clm/pkg/controller"
)

// Client for using the services in the ncm
type Client struct {
	Cluster    *cluster.ClusterClient
	Controller *controller.ControllerClient
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.Cluster = cluster.NewClusterClient()
	// Add Client API handlers here
	return c
}

// NewController creates a new controlller for using the services
func NewController() *Client {
	c := &Client{}
	c.Controller = controller.NewControllerClient()
	// Add Client API handlers here
	return c
}
