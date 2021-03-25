// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"github.com/open-ness/EMCO/src/sfc/pkg/model"
)

// ClientDbInfo structure for storing info about SFC DB
type ClientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
	tagContext string // attribute key name for context object in App Context
}

// SfcIntentManager is an interface exposing the SFC Intent functionality
type SfcIntentManager interface {
	// SFC Intent functions
	CreateSfcIntent(sfc model.SfcIntent, pr, ca, caver, dig, netctrlint string, exists bool) (model.SfcIntent, error)
	GetSfcIntent(name, pr, ca, caver, dig, netctrlint string) (model.SfcIntent, error)
	GetAllSfcIntents(pr, ca, caver, dig, netctrlint string) ([]model.SfcIntent, error)
	DeleteSfcIntent(name, pr, ca, caver, dig, netctrlint string) error
}

// SfcIntentManager is an interface exposing the SFC Intent functionality
type SfcClientSelectorIntentManager interface {
	// SFC Client Selector Intent functions
	CreateSfcClientSelectorIntent(sfc model.SfcClientSelectorIntent, pr, ca, caver, dig, netctrlint, sfcIntent string, exists bool) (model.SfcClientSelectorIntent, error)
	GetSfcClientSelectorIntent(name, pr, ca, caver, dig, netctrlint, sfcIntent string) (model.SfcClientSelectorIntent, error)
	GetAllSfcClientSelectorIntents(pr, ca, caver, dig, netctrlint, sfcIntent string) ([]model.SfcClientSelectorIntent, error)
	GetSfcClientSelectorIntentsByEnd(pr, ca, caver, dig, netctrlint, sfcIntent, chainEnd string) ([]model.SfcClientSelectorIntent, error)
	DeleteSfcClientSelectorIntent(name, pr, ca, caver, dig, netctrlint, sfcIntent string) error
}

// SfcIntentManager is an interface exposing the SFC Intent functionality
type SfcProviderNetworkIntentManager interface {
	// SFC Network Provider Intent functions
	CreateSfcProviderNetworkIntent(sfc model.SfcProviderNetworkIntent, pr, ca, caver, dig, netctrlint, sfcIntent string, exists bool) (model.SfcProviderNetworkIntent, error)
	GetSfcProviderNetworkIntent(name, pr, ca, caver, dig, netctrlint, sfcIntent string) (model.SfcProviderNetworkIntent, error)
	GetAllSfcProviderNetworkIntents(pr, ca, caver, dig, netctrlint, sfcIntent string) ([]model.SfcProviderNetworkIntent, error)
	GetSfcProviderNetworkIntentsByEnd(pr, ca, caver, dig, netctrlint, sfcIntent, chainEnd string) ([]model.SfcProviderNetworkIntent, error)
	DeleteSfcProviderNetworkIntent(name, pr, ca, caver, dig, netctrlint, sfcIntent string) error
}

// Client for using the services in the ncm
type Client struct {
	SfcIntent                *SfcIntentClient
	SfcClientSelectorIntent  *SfcClientSelectorIntentClient
	SfcProviderNetworkIntent *SfcProviderNetworkIntentClient
	// Add Clients for API's here
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.SfcIntent = NewSfcIntentClient()
	c.SfcClientSelectorIntent = NewSfcClientSelectorIntentClient()
	c.SfcProviderNetworkIntent = NewSfcProviderNetworkIntentClient()
	// Add Client API handlers here
	return c
}

// SfcIntentClient implements the SfcIntentManager
// It will also be used to maintain some localized state
type SfcIntentClient struct {
	db ClientDbInfo
}

// NewSfcIntentClient returns an instance of the SfcIntentClient
// which implements the Manager for SFC Intents
func NewSfcIntentClient() *SfcIntentClient {
	return &SfcIntentClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "sfcmetadata",
		},
	}
}

// SfcClientSelectorIntentClient implements the SfcClientSelectorIntentManager
// It will also be used to maintain some localized state
type SfcClientSelectorIntentClient struct {
	db ClientDbInfo
}

// NewSfcClientSelectorIntentClient returns an instance of the SfcIntentClient
// which implements the Manager for SFC Client Selector Intents
func NewSfcClientSelectorIntentClient() *SfcClientSelectorIntentClient {
	return &SfcClientSelectorIntentClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "sfcclientselectormetadata",
		},
	}
}

// SfcProviderNetworkIntentClient implements the SfcProviderNetworkIntentManager
// It will also be used to maintain some localized state
type SfcProviderNetworkIntentClient struct {
	db ClientDbInfo
}

// NewSfcProviderNetworkIntentClient returns an instance of the SfcIntentClient
// which implements the Manager for SFC Provider Network Intents
func NewSfcProviderNetworkIntentClient() *SfcProviderNetworkIntentClient {
	return &SfcProviderNetworkIntentClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "sfcprovidernetworkmetadata",
		},
	}
}
