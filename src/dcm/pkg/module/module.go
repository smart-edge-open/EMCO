// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

// Client for using the services in the orchestrator
type Client struct {
	LogicalCloud   *LogicalCloudClient
	Cluster        *ClusterClient
	Quota          *QuotaClient
	UserPermission *UserPermissionClient
	KeyValue       *KeyValueClient
	// Add Clients for API's here
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.LogicalCloud = NewLogicalCloudClient()
	c.Cluster = NewClusterClient()
	c.Quota = NewQuotaClient()
	c.UserPermission = NewUserPermissionClient()
	c.KeyValue = NewKeyValueClient()
	// Add Client API handlers here
	return c
}
