// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

// This is a sample module - FirstModule

import (
	"log"
)

// FirstModule is a sample module
type FirstModule struct {
	Metadata Metadata `json:"metadata"`
}

// FirstModuleClient is a sample client for firstModule
type FirstModuleClient struct {
	db FirstModuleClientDbInfo
}

// FirstModuleClientDbInfo is a sample client db info
type FirstModuleClientDbInfo struct {
	storeName string
	tagState  string // attribute key name for context object in App Context
}

// FirstModuleManager is an interface representing all the methods that are required for this module
type FirstModuleManager interface {
	PrintAB(a string, b string) error
}

// NewFirstModuleClient returns FirstModuleClient
func NewFirstModuleClient() *FirstModuleClient {
	return &FirstModuleClient{
		db: FirstModuleClientDbInfo{
			storeName: "controller",
			tagState:  "stateInfo",
		},
	}
}

func (c FirstModuleClient) printAB(a string, b string) error {
	log.Println(a)
	log.Println(b)
	return nil
}
