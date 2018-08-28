// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation
package types

// TODO - should move to common module types location - e.g. in orchestrator
type ClientDbInfo struct {
	StoreName  string // name of the mongodb collection to use for client documents
	TagMeta    string // attribute key name for the json data of a client document
	TagContent string // attribute key name for the file data of a client document
	TagState   string // attribute key name for context object in App Context
}
