// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package grpc

import (
	"os"
	"strconv"
	"strings"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

const default_host = "localhost"
const default_port = 9038
const default_nps_name = "nps"
const ENV_NP_NAME = "NPS"

func GetServerHostPort() (string, int) {

	// expect name of this nps program to be in env variable "NP_NAME" - e.g. NP_NAME="nps"
	serviceName := os.Getenv(ENV_NP_NAME)
	if serviceName == "" {
		serviceName = default_nps_name
		log.Info("Using default name for NP service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. NP_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = default_host
		log.Info("Using default host for nps gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. NP_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = default_port
		log.Info("Using default port for nps gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}
