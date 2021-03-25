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
const default_port = 9039
const default_sds_name = "sds"
const ENV_SD_NAME = "SDS"

func GetServerHostPort() (string, int) {

	// expect name of this sds program to be in env variable "SD_NAME" - e.g. SD_NAME="sds"
	serviceName := os.Getenv(ENV_SD_NAME)
	if serviceName == "" {
		serviceName = default_sds_name
		log.Info("Using default name for SD service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. SD_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = default_host
		log.Info("Using default host for sds gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. SD_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = default_port
		log.Info("Using default port for sds gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}
