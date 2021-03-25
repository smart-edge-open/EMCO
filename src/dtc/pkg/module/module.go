// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"github.com/open-ness/EMCO/src/orchestrator/pkg/module/controller"
)

// Client for using the services in the ncm
type Client struct {
	TrafficGroupIntent *TrafficGroupIntentDbClient
	ServerInboundIntent *InboundServerIntentDbClient
	ClientsInboundIntent *InboundClientsIntentDbClient
	Controller *controller.ControllerClient
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.TrafficGroupIntent = NewTrafficGroupIntentClient()
	c.ServerInboundIntent = NewServerInboundIntentClient()
	c.ClientsInboundIntent = NewClientsInboundIntentClient()
	c.Controller = controller.NewControllerClient("dtccontroller", "dtccontrollermetadata")
	return c
}
