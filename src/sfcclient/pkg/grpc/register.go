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
const default_port = 9058
const default_sfcclient_name = "sfcclient"
const ENV_SFCCLIENT_NAME = "SFCCLIENT_NAME"

func GetServerHostPort() (string, int) {

	// expect name of this sfcclient program to be in env variable "SFCCLIENT_NAME" - e.g. SFCCLIENT_NAME="sfcclient"
	serviceName := os.Getenv(ENV_SFCCLIENT_NAME)
	if serviceName == "" {
		serviceName = default_sfcclient_name
		log.Info("Using default name for SFCCLIENT service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. SFCCLIENT_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = default_host
		log.Info("Using default host for sfcclient gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. SFCCLIENT_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = default_port
		log.Info("Using default port for sfcclient gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}
