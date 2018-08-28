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
const default_port = 9048
const default_dtc_name = "dtc"
const ENV_DTC_NAME = "DTC_NAME"

func GetServerHostPort() (string, int) {

	// expect name of this dtc program to be in env variable "DTC_NAME" - e.g. DTC_NAME="dtc"
	serviceName := os.Getenv(ENV_DTC_NAME)
	if serviceName == "" {
		serviceName = default_dtc_name
		log.Info("Using default name for DTC service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. DTC_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = default_host
		log.Info("Using default host for dtc gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. DTC_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = default_port
		log.Info("Using default port for dtc gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}
