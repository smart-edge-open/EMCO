# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation
export GO111MODULE=on
export EMCOBUILDROOT=$(shell pwd)
export CONFIG := $(wildcard config/*.txt)

ifndef MODS
MODS=clm dcm dtc nps sds genericactioncontroller monitor ncm orchestrator ovnaction rsync tools/emcoctl sfc sfcclient hpa-plc hpa-ac
endif

all: check-env docker-reg build

check-env:
	@echo "Check for environment parameters"
ifndef EMCODOCKERREPO
	$(error EMCODOCKERREPO env variable needs to be set)
endif

ifndef BRANCH
export BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
endif

ifeq ($(BUILD_CAUSE), RELEASE)
 ifndef TAG
  export TAG=$(shell git tag --points-at HEAD | awk 'NR==1 {print $1}')
  ifndef TAG
  export TAG=${BRANCH}-daily-`date +"%m%d%y"`
  endif
 endif
endif

docker-reg:
	@echo "Setting up docker Registry with base image"
export BUILD_BASE_IMAGE_NAME := $(shell cat $(CONFIG) | grep 'BUILD_BASE_IMAGE_NAME' | cut -d'=' -f2)
export BUILD_BASE_IMAGE_VERSION := $(shell cat $(CONFIG) | grep 'BUILD_BASE_IMAGE_VERSION' | cut -d'=' -f2)
export SERVICE_BASE_IMAGE_NAME := $(shell cat $(CONFIG) | grep 'SERVICE_BASE_IMAGE_NAME' | cut -d'=' -f2)
export SERVICE_BASE_IMAGE_VERSION := $(shell cat $(CONFIG) | grep 'SERVICE_BASE_IMAGE_VERSION' | cut -d'=' -f2)

clean:
	@echo "Cleaning artifacts"
	@for m in $(MODS); do \
	    $(MAKE) -C ./src/$$m clean; \
	 done
	@rm -rf bin
	@echo "    Done."

pre-compile: clean
	@echo "Setting up pre-requisites"
	@for m in $(MODS); do \
	    mkdir -p bin/$$m;  \
	    ARGS=""; CJ="src/$$m/config.json"; JS="src/$$m/json-schemas"; \
	    [[ -f $$CJ ]] && ARGS="$$ARGS $$CJ"; \
	    [[ -d $$JS ]] && ARGS="$$ARGS $$JS"; \
	    [[ -z "$$ARGS" ]] || cp -r $$ARGS bin/$$m; \
	 done
	@echo "    Done."

compile-container: pre-compile
	@echo "Building artifacts"
	@for m in $(MODS); do \
	    $(MAKE) -C ./src/$$m all; \
	 done
	@echo "    Done."

compile: check-env docker-reg
	@echo "Building microservices within Docker build container"
	docker run --rm --user `id -u`:`id -g` --env MODS="${MODS}" --env GO111MODULE --env XDG_CACHE_HOME=/tmp/.cache --env BRANCH=${BRANCH} --env TAG=${TAG} -v `pwd`:/repo ${EMCODOCKERREPO}${BUILD_BASE_IMAGE_NAME}${BUILD_BASE_IMAGE_VERSION} /bin/sh -c "cd /repo; make compile-container"
	@echo "    Done."

# Modules that follow naming conventions are done in a loop, rest later
build: compile
	@echo "Packaging microservices "
	@export ARGS="--build-arg EMCODOCKERREPO=${EMCODOCKERREPO} --build-arg          SERVICE_BASE_IMAGE_NAME=${SERVICE_BASE_IMAGE_NAME} --build-arg                          SERVICE_BASE_IMAGE_VERSION=${SERVICE_BASE_IMAGE_VERSION}"; \
	 for m in $(MODS); do \
	    case $$m in \
	      "tools/emcoctl") continue;; \
	      "ovnaction") d="ovn"; n=$$d;; \
	      "genericactioncontroller") d="gac"; n=$$d;; \
	      "orchestrator") d=$$m; n="orch";; \
	      *) d=$$m; n=$$m;; \
	    esac; \
	    echo "Packaging $$m"; \
	    docker build $$ARGS --rm -t emco-$$n -f ./build/docker/Dockerfile.$$d ./bin/$$m; \
	 done
	@echo "    Done."

deploy: check-env docker-reg build
	@echo "Creating helm charts. Pushing microservices to registry & copying docker-compose files if BUILD_CAUSE set to DEV_TEST"
	@docker run --env USER=${USER} --env EMCODOCKERREPO=${EMCODOCKERREPO} --env BUILD_CAUSE=${BUILD_CAUSE} --env BRANCH=${BRANCH} --env TAG=${TAG} --env EMCOSRV_RELEASE_TAG=${EMCOSRV_RELEASE_TAG} --rm --user `id -u`:`id -g` --env GO111MODULE --env XDG_CACHE_HOME=/tmp/.cache -v `pwd`:/repo ${EMCODOCKERREPO}${BUILD_BASE_IMAGE_NAME}${BUILD_BASE_IMAGE_VERSION} /bin/sh -c "cd /repo/scripts ; sh deploy_emco_openness.sh"
	@MODS=`echo ${MODS} | sed 's/ovnaction/ovn/;s/genericactioncontroller/gac/;s/orchestrator/orch/;'` ./scripts/push_to_registry.sh
	@echo "    Done."

test:
	@echo "Running tests"
	@for m in $(MODS); do \
	    $(MAKE) -C ./src/$$m test; \
	 done
	@echo "    Done."

tidy:
	@echo "Cleaning up dependencies"
	@for m in $(MODS); do \
	    cd src/$$m; go mod tidy; cd - > /dev/null; \
	 done
	@echo "    Done."

build-base:
	@echo "Building base images and pushing to Harbor"
	./scripts/build-base-images.sh
