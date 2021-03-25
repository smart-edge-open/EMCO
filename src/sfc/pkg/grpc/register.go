// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package grpc

import (
	"os"
	"strconv"
	"strings"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

const default_host = "localhost"
const default_port = 9056
const default_sfc_name = "sfc"
const ENV_SFC_NAME = "SFC_NAME"

func GetServerHostPort() (string, int) {

	// expect name of this sfc program to be in env variable "SFC_NAME" - e.g. SFC_NAME="sfc"
	serviceName := os.Getenv(ENV_SFC_NAME)
	if serviceName == "" {
		serviceName = default_sfc_name
		log.Info("Using default name for SFC service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. SFC_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = default_host
		log.Info("Using default host for sfc gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. SFC_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = default_port
		log.Info("Using default port for sfc gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}
