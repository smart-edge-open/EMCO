// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"github.com/open-ness/EMCO/src/ncm/pkg/networkintents"
	"github.com/open-ness/EMCO/src/ncm/pkg/scheduler"
)

// Client for using the services in the ncm
type Client struct {
	Network     *networkintents.NetworkClient
	ProviderNet *networkintents.ProviderNetClient
	Scheduler   *scheduler.SchedulerClient
	// Add Clients for API's here
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.Network = networkintents.NewNetworkClient()
	c.ProviderNet = networkintents.NewProviderNetClient()
	c.Scheduler = scheduler.NewSchedulerClient()
	// Add Client API handlers here
	return c
}
