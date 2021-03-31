# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation
export GO111MODULE=on
export EMCOBUILDROOT=$(shell pwd)
export CONFIG := $(wildcard config/*.txt)

all: check-env docker-reg build

check-env:
	@echo "Check for environment parameters"
ifndef EMCODOCKERREPO
	$(error EMCODOCKERREPO env variable needs to be set)
endif

ifndef BRANCH
export BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
endif

docker-reg:
	@echo "Setting up docker Registry with base image"
export BUILD_BASE_IMAGE_NAME := $(shell cat $(CONFIG) | grep 'BUILD_BASE_IMAGE_NAME' | cut -d'=' -f2)
export BUILD_BASE_IMAGE_VERSION := $(shell cat $(CONFIG) | grep 'BUILD_BASE_IMAGE_VERSION' | cut -d'=' -f2)
export SERVICE_BASE_IMAGE_NAME := $(shell cat $(CONFIG) | grep 'SERVICE_BASE_IMAGE_NAME' | cut -d'=' -f2)
export SERVICE_BASE_IMAGE_VERSION := $(shell cat $(CONFIG) | grep 'SERVICE_BASE_IMAGE_VERSION' | cut -d'=' -f2)

clean:
	@echo "Cleaning artifacts"
	$(MAKE) -C ./src/clm clean
	$(MAKE) -C ./src/monitor clean
	$(MAKE) -C ./src/ncm clean
	$(MAKE) -C ./src/orchestrator clean
	$(MAKE) -C ./src/ovnaction clean
	$(MAKE) -C ./src/dtc clean
	$(MAKE) -C ./src/rsync clean
	$(MAKE) -C ./src/dcm clean
	$(MAKE) -C ./src/genericactioncontroller clean
	$(MAKE) -C ./src/tools/emcoctl clean
	@rm -rf bin
	@echo "    Done."

pre-compile: clean
	@echo "Setting up pre-requisites"
	@mkdir -p bin/clm bin/monitor bin/ncm bin/orchestrator bin/ovnaction bin/dtc bin/rsync bin/dcm bin/genericactioncontroller bin/emcoctl
	@cp -r src/clm/config.json src/clm/json-schemas bin/clm
	@cp -r src/ncm/config.json src/ncm/json-schemas bin/ncm
	@cp -r src/orchestrator/config.json src/orchestrator/json-schemas bin/orchestrator
	@cp -r src/ovnaction/config.json src/ovnaction/json-schemas bin/ovnaction
	@cp -r src/genericactioncontroller/config.json src/genericactioncontroller/json-schemas bin/genericactioncontroller
	@cp -r src/dtc/config.json src/dtc/json-schemas bin/dtc
	@cp -r src/rsync/config.json bin/rsync
	@cp -r src/dcm/config.json bin/dcm
	@echo "    Done."

compile-container: pre-compile
	@echo "Building artifacts"
	$(MAKE) -C ./src/clm all
	$(MAKE) -C ./src/monitor all
	$(MAKE) -C ./src/ncm all
	$(MAKE) -C ./src/orchestrator all
	$(MAKE) -C ./src/ovnaction all
	$(MAKE) -C ./src/dtc all
	$(MAKE) -C ./src/rsync all
	$(MAKE) -C ./src/dcm all
	$(MAKE) -C ./src/genericactioncontroller all
	$(MAKE) -C ./src/tools/emcoctl all
	@echo "    Done."

compile: check-env docker-reg
	@echo "Building microservices within Docker build container"
	docker run --rm --user `id -u`:`id -g` --env GO111MODULE --env XDG_CACHE_HOME=/tmp/.cache --env BRANCH=${BRANCH} -v `pwd`:/repo ${EMCODOCKERREPO}${BUILD_BASE_IMAGE_NAME}${BUILD_BASE_IMAGE_VERSION} /bin/sh -c "cd /repo; make compile-container"
	@echo "    Done."

build: compile
	@echo "Packaging microservices "
	@echo "Packaging CLM"
	@docker build --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} --rm -t emco-clm -f ./build/docker/Dockerfile.clm ./bin/clm
	@echo "Packaging NCM"
	@docker build --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} --rm -t emco-ncm -f ./build/docker/Dockerfile.ncm ./bin/ncm
	@echo "Packaging Orchestrator"
	@docker build --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} --rm -t emco-orch -f ./build/docker/Dockerfile.orchestrator ./bin/orchestrator
	@echo "Packaging OvnAction"
	@docker build --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} --rm -t emco-ovn -f ./build/docker/Dockerfile.ovn ./bin/ovnaction
	@echo "Packaging GenericActionController"
	@docker build --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} --rm -t emco-gac -f ./build/docker/Dockerfile.gac ./bin/genericactioncontroller
	@echo "Packaging DTC"
	@docker build --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} --rm -t emco-dtc -f ./build/docker/Dockerfile.dtc ./bin/dtc
	@echo "Packaging RSync"
	@docker build --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} --rm -t emco-rsync -f ./build/docker/Dockerfile.rsync ./bin/rsync
	@echo "Packing DCM"
	@docker build --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} --rm -t emco-dcm -f ./build/docker/Dockerfile.dcm ./bin/dcm
	@echo "Packing Monitor"
	@docker build --build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION} --rm -t emco-monitor -f ./build/docker/Dockerfile.monitor ./bin/monitor
	@echo "    Done."

deploy: check-env docker-reg build
	@echo "Creating helm charts. Pushing microservices to registry & copying docker-compose files if BUILD_CAUSE set to DEV_TEST"
	@docker run --env USER=${USER} --env EMCODOCKERREPO=${EMCODOCKERREPO} --env BUILD_CAUSE=${BUILD_CAUSE} --env BRANCH=${BRANCH} --rm --user `id -u`:`id -g` --env GO111MODULE --env XDG_CACHE_HOME=/tmp/.cache -v `pwd`:/repo ${EMCODOCKERREPO}${BUILD_BASE_IMAGE_NAME}${BUILD_BASE_IMAGE_VERSION} /bin/sh -c "cd /repo/scripts ; sh deploy_emco_openness.sh"
	./scripts/push_to_registry.sh
	@echo "    Done."

test:
	@echo "Running tests"
	$(MAKE) -C ./src/clm test
	$(MAKE) -C ./src/dcm test
	$(MAKE) -C ./src/dtc test
	$(MAKE) -C ./src/genericactioncontroller test
	$(MAKE) -C ./src/monitor test
	$(MAKE) -C ./src/ncm test
	$(MAKE) -C ./src/orchestrator test
	$(MAKE) -C ./src/ovnaction test
	$(MAKE) -C ./src/rsync test
	$(MAKE) -C ./src/tools/emcoctl test
	@echo "    Done."

tidy:
	@echo "Cleaning up dependencies"
	@cd src/clm; go mod tidy
	@cd src/dcm; go mod tidy
	@cd src/monitor; go mod tidy
	@cd src/ncm; go mod tidy
	@cd src/orchestrator; go mod tidy
	@cd src/ovnaction; go mod tidy
	@cd src/genericactioncontroller; go mod tidy
	@cd src/dtc; go mod tidy
	@cd src/rsync; go mod tidy
	@cd src/tools/emcoctl; go mod tidy
	@echo "    Done."

build-base:
	@echo "Building base images and pushing to Harbor"
	./scripts/build-base-images.sh
