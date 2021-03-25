// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"-"`
	UserData1   string `json:"userData1" yaml:"-"`
	UserData2   string `json:"userData2" yaml:"-"`
}

// SfcIntentKey is the key structure that is used in the database
type SfcIntentKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	DigName             string `json:"deploymentintentgroup"`
	NetControlIntent    string `json:"netcontrolintent"`
	SfcIntent           string `json:"sfcintent"`
}

// SfcIntent defines the high level structure of a network chain document
type SfcIntent struct {
	Metadata Metadata      `json:"metadata" yaml:"metadata"`
	Spec     SfcIntentSpec `json:"spec" yaml:"spec"`
}

// SfcIntentSpec contains the specification of a network chain
type SfcIntentSpec struct {
	ChainType    string `json:"chainType"`
	NetworkChain string `json:"networkChain"`
	Namespace    string `json:"namespace"`
}

// RouteSpec contains the routing specificaiton of a network chain
type RouteSpec struct {
	LeftNetwork  []RoutingNetwork `json:"leftNetwork"`
	RightNetwork []RoutingNetwork `json:"rightNetwork"`
	NetworkChain string           `json:"networkChain"`
	Namespace    string           `json:"namespace"`
}

// RoutingNetwork contains the route networkroute network details for en element of a network chain
type RoutingNetwork struct {
	NetworkName string `json:"networkName"`
	GatewayIP   string `json:"gatewayIp"`
	Subnet      string `json:"subnet"`
}

// CrChain is the structure for the Network Chain Custom Resource
type CrChain struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	//	Chain      Chain
}

// RoutingChainType is currently only defined chaining type
const RoutingChainType = "Routing"

// Name of 'app' to create for Network Chaining CRs in the EMCO AppContext
const ChainingApp = "network-chain-intents"

// ChainingAPIVersion is the kubernetes version of a network chain custom resource
const ChainingAPIVersion = "k8s.plugin.opnfv.org/v1alpha1"

// ChainingKind is the Kind string for a network chain
const ChainingKind = "NetworkChaining"

// Chain constants
const LeftChainEnd = "left"
const RightChainEnd = "right"

// SfcClientSelectorIntent defines the high level structure of a network chain document
type SfcClientSelectorIntent struct {
	Metadata Metadata                    `json:"metadata" yaml:"metadata"`
	Spec     SfcClientSelectorIntentSpec `json:"spec" yaml:"spec"`
}

// SfcClientSelectorIntentSpec contains the specification of a network chain
type SfcClientSelectorIntentSpec struct {
	ChainEnd          string               `json:"chainEnd"`
	PodSelector       metav1.LabelSelector `json:"podSelector"`
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector"`
}

// SfcClientSelectorIntentKey is the key structure that is used in the database
type SfcClientSelectorIntentKey struct {
	Project                 string `json:"project"`
	CompositeApp            string `json:"compositeapp"`
	CompositeAppVersion     string `json:"compositeappversion"`
	DigName                 string `json:"deploymentintentgroup"`
	NetControlIntent        string `json:"netcontrolintent"`
	SfcIntent               string `json:"sfcintent"`
	SfcClientSelectorIntent string `json:"sfcclientselectorintent"`
}

// SfcClientSelectorIntentByEndKey is the key structure that is used in the database
type SfcClientSelectorIntentByEndKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	DigName             string `json:"deploymentintentgroup"`
	NetControlIntent    string `json:"netcontrolintent"`
	SfcIntent           string `json:"sfcintent"`
	ChainEnd            string `json:"chainEnd"`
}

// SfcProviderNetworkIntent defines the high level structure of a network chain document
type SfcProviderNetworkIntent struct {
	Metadata Metadata                     `json:"metadata" yaml:"metadata"`
	Spec     SfcProviderNetworkIntentSpec `json:"spec" yaml:"spec"`
}

// SfcProviderNetworkIntentSpec contains the specification of a network chain
type SfcProviderNetworkIntentSpec struct {
	ChainEnd    string `json:"chainEnd"`
	NetworkName string `json:"networkName"`
	GatewayIp   string `json:"gatewayIp"`
	Subnet      string `json:"subnet"`
}

// SfcProviderNetworkIntentKey is the key structure that is used in the database
type SfcProviderNetworkIntentKey struct {
	Project                  string `json:"project"`
	CompositeApp             string `json:"compositeapp"`
	CompositeAppVersion      string `json:"compositeappversion"`
	DigName                  string `json:"deploymentintentgroup"`
	NetControlIntent         string `json:"netcontrolintent"`
	SfcIntent                string `json:"sfcintent"`
	SfcProviderNetworkIntent string `json:"sfcprovidernetworkintent"`
}

// SfcProviderNetworkIntentByEndKey is the key structure that is used in the database
type SfcProviderNetworkIntentByEndKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	DigName             string `json:"deploymentintentgroup"`
	NetControlIntent    string `json:"netcontrolintent"`
	SfcIntent           string `json:"sfcintent"`
	ChainEnd            string `json:"chainEnd"`
}

// SfcEndKey can be added to resources for searching SFC intents by ChainEnd value
type SfcEndKey struct {
	ChainEnd string `json:"chainEnd"`
}
