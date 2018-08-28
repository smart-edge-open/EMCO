package grpc

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"os"
	"strconv"
	"strings"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

const DEFAULTHOST = "localhost"
const DEFAULTPORT = 9033
const DEFAULTGENERICACTIONCONTROLLER_NAME = "genericaction"
const ENV_GENERIC_ACTION_CONTROLLER_NAME = "GENERICACTIONCONTROLLER_NAME"

func GetServerHostPort() (string, int) {

	// expect name of this contoller program to be in env variable "GENERIC_ACTION_CONTROLLER_NAME" - e.g. GENERIC_ACTION_CONTROLLER_NAME="generic_action_controller"
	serviceName := os.Getenv(ENV_GENERIC_ACTION_CONTROLLER_NAME)
	if serviceName == "" {
		serviceName = DEFAULTGENERICACTIONCONTROLLER_NAME
		log.Info("Using default name for generic-action-controller service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. GENERIC_ACTION_CONTROLLER_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = DEFAULTHOST
		log.Info("Using default host for generic-action-controller gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. GENERIC_ACTION_CONTROLLER_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = DEFAULTPORT
		log.Info("Using default port for generic-action-controller gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}
