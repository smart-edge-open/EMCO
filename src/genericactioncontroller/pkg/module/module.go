package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

// Client for using the services in the ncm
type Client struct {

	// Add Clients for API's here

	GenericK8sIntent *GenericK8sIntentClient
	BaseResource     *ResourceClient
	Customization    *CustomizationClient
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.GenericK8sIntent = NewGenericK8sIntentClient()
	c.BaseResource = NewResourceClient()
	c.Customization = NewCustomizationClient()
	log.Info("Setting the client!", log.Fields{})
	// Add Client API handlers here
	return c
}
