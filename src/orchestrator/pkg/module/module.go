// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"github.com/open-ness/EMCO/src/orchestrator/pkg/module/controller"
)

// Client for using the services in the orchestrator
type Client struct {
	Project                *ProjectClient
	CompositeApp           *CompositeAppClient
	App                    *AppClient
	Controller             *controller.ControllerClient
	GenericPlacementIntent *GenericPlacementIntentClient
	AppIntent              *AppIntentClient
	DeploymentIntentGroup  *DeploymentIntentGroupClient
	Intent                 *IntentClient
	CompositeProfile       *CompositeProfileClient
	AppProfile             *AppProfileClient
	// Add Clients for API's here
	Instantiation *InstantiationClient
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.Project = NewProjectClient()
	c.CompositeApp = NewCompositeAppClient()
	c.App = NewAppClient()
	c.Controller = controller.NewControllerClient("controller", "controllermetadata")
	c.GenericPlacementIntent = NewGenericPlacementIntentClient()
	c.AppIntent = NewAppIntentClient()
	c.DeploymentIntentGroup = NewDeploymentIntentGroupClient()
	c.Intent = NewIntentClient()
	c.CompositeProfile = NewCompositeProfileClient()
	c.AppProfile = NewAppProfileClient()
	// Add Client API handlers here
	c.Instantiation = NewInstantiationClient()
	return c
}
