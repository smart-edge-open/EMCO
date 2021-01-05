#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

REGISTRY=${EMCODOCKERREPO}
IMAGE=$1
TAG=$2

if [ -z ${TAG} ]; then
  TAG=master-daily-`date +"%m%d%y"`
  # check if it is a cron scheduled build
  if [ "${BUILD_CAUSE}" != "TIMERTRIGGER" ]; then
    echo "WARNING: this is not a CI build; skipping..."
    exit 0
  fi
fi

echo "Pushing ${IMAGE} to ${REGISTRY}${IMAGE}:${TAG}..."

docker tag ${IMAGE}:latest ${REGISTRY}${IMAGE}:${TAG}
docker push ${REGISTRY}${IMAGE}:${TAG}
