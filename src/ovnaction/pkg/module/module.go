// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

// Client for using the services in the ncm
type Client struct {
	NetControlIntent *NetControlIntentClient
	WorkloadIntent   *WorkloadIntentClient
	WorkloadIfIntent *WorkloadIfIntentClient
	// Add Clients for API's here
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.NetControlIntent = NewNetControlIntentClient()
	c.WorkloadIntent = NewWorkloadIntentClient()
	c.WorkloadIfIntent = NewWorkloadIfIntentClient()
	// Add Client API handlers here
	return c
}
