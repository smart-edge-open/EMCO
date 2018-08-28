// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

// Client for using the services in the ncm
type Client struct {
	TrafficGroupIntent *TrafficGroupIntentDbClient
	ServerInboundIntent *InboundServerIntentDbClient
	ClientsInboundIntent *InboundClientsIntentDbClient
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.TrafficGroupIntent = NewTrafficGroupIntentClient()
	c.ServerInboundIntent = NewServerInboundIntentClient()
	c.ClientsInboundIntent = NewClientsInboundIntentClient()
	return c
}
