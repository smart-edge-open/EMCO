// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package grpc

import (
	"os"
	"strconv"
	"strings"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

const defaultHost = "localhost"
const defaultPort = 9042
const defaultHpaactionName = "hpaaction"
const ENV_HPAACTION_NAME = "HPAACTION_NAME"

// GetServerHostPort ..
func GetServerHostPort() (string, int) {

	// expect name of this ncm program to be in env variable "ENV_HPAACTION_NAME" - e.g. ENV_HPAACTION_NAME="hpa"
	serviceName := os.Getenv(ENV_HPAACTION_NAME)
	if serviceName == "" {
		serviceName = defaultHpaactionName
		log.Info("Using default name for HPAACTION service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. HPAACTION_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = defaultHost
		log.Info("Using default host for hpaaction gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. HPAACTION_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = defaultPort
		log.Info("Using default port for hpaaction gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}
