// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package grpc

import (
	"os"
	"strconv"
	"strings"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

// PlacementControllerName .. name of the controller
const PlacementControllerName = "EmcoHpaPlacementController"
const defaultHost = "localhost"
const defaultPort = 9099
const defaultHpaplacementName = "hpaplacement"
const ENV_HPAPLACEMENT_NAME = "HPAPLACEMENT_NAME"

// GetServerHostPort ..
func GetServerHostPort() (string, int) {
	// expect name of this hpaplacement program to be in env variable "HPAPLACEMENT_NAME" - e.g. HPAPLACEMENT_NAME="hpaplacement"
	serviceName := os.Getenv(ENV_HPAPLACEMENT_NAME)
	if serviceName == "" {
		serviceName = defaultHpaplacementName
		log.Info("Using default name for HPAPLACEMENT service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. HPAPLACEMENT_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = defaultHost
		log.Info("Using default host for hpaplacement gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. HPAPLACEMENT_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = defaultPort
		log.Info("Using default port for hpaplacement gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}
