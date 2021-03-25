// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"github.com/open-ness/EMCO/src/sfcclient/pkg/model"
)

// ClientDbInfo structure for storing info about SFC DB
type ClientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
	tagContext string // attribute key name for context object in App Context
}

// SfcIntentManager is an interface exposing the SFC Intent functionality
type SfcManager interface {
	// SFC Intent functions
	CreateSfcClientIntent(sfc model.SfcClientIntent, pr, ca, caver, dig, netctrlint string, exists bool) (model.SfcClientIntent, error)
	GetSfcClientIntent(name, pr, ca, caver, dig, netctrlint string) (model.SfcClientIntent, error)
	GetAllSfcClientIntents(pr, ca, caver, dig, netctrlint string) ([]model.SfcClientIntent, error)
	DeleteSfcClientIntent(name, pr, ca, caver, dig, netctrlint string) error
}

// SfcClient implements the Manager
// It will also be used to maintain some localized state
type SfcClient struct {
	db ClientDbInfo
}

// NewSfcClient returns an instance of the SfcClient
// which implements the Manager
func NewSfcClient() *SfcClient {
	return &SfcClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "sfcclientmetadata",
		},
	}
}
