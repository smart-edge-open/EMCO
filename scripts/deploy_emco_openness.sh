#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

REGISTRY=${EMCODOCKERREPO}
#EMCOBUILDROOT is now container's root DIR
EMCOBUILDROOT=/repo
BIN_PATH=${EMCOBUILDROOT}/bin
TAG=${TAG}

create_helm_chart() {
  echo "Creating helm chart"
  mkdir -p ${BIN_PATH}/helm
  cp -rf ${EMCOBUILDROOT}/deployments/helm/emcoOpenNESS ${EMCOBUILDROOT}/deployments/helm/monitor ${BIN_PATH}/helm/
  cat > ${BIN_PATH}/helm/emcoOpenNESS/common/values.yaml <<EOF
repository: ${REGISTRY}
imageTag: ${TAG}
noProxyHosts: ${NO_PROXY}
EOF
  cat > ${BIN_PATH}/helm/monitor/values.yaml <<EOF
registryPrefix: ${REGISTRY}
tag: ${TAG}
EOF
  cat > ${BIN_PATH}/helm/helm_value_overrides.yaml <<EOF
#update proxies
noProxyHosts: ${NO_PROXY}
#update and uncomment if build tag to be changed
#imageTag: latest
#update and uncomment to override registry
#repository: registry.docker.com/
global:
  loglevel: info
EOF

  # Submodules to use evaluated values.yaml via common templates
  cp -rf ${EMCOBUILDROOT}/deployments/helm/emcoOpenNESS/common ${EMCOBUILDROOT}/deployments/helm/
  cat > ${EMCOBUILDROOT}/deployments/helm/common/values.yaml <<EOF
repository: ${REGISTRY}
imageTag: ${TAG}
noProxyHosts: ${NO_PROXY}
EOF

  # emcoOpenNESS
  cp ${EMCOBUILDROOT}/deployments/helm/emco-openness-helm-install.sh ${BIN_PATH}/helm/install_template
  cat ${BIN_PATH}/helm/install_template | sed -e "s/emco-db-0.1.0.tgz/emco-db-${TAG}.tgz/" \
                                              -e "s/emco-services-0.1.0.tgz/emco-services-${TAG}.tgz/" \
                                              -e "s/emco-tools-0.1.0.tgz/emco-tools-${TAG}.tgz/" > ${BIN_PATH}/helm/emco-openness-helm-install.sh
  chmod +x ${BIN_PATH}/helm/emco-openness-helm-install.sh
  rm -f ${BIN_PATH}/helm/install_template

  make -C ${BIN_PATH}/helm/emcoOpenNESS all
  mv ${BIN_PATH}/helm/emcoOpenNESS/dist/packages/emco-db-0.1.0.tgz ${BIN_PATH}/helm/emco-db-${TAG}.tgz
  mv ${BIN_PATH}/helm/emcoOpenNESS/dist/packages/emco-services-0.1.0.tgz ${BIN_PATH}/helm/emco-services-${TAG}.tgz
  mv ${BIN_PATH}/helm/emcoOpenNESS/dist/packages/emco-tools-0.1.0.tgz ${BIN_PATH}/helm/emco-tools-${TAG}.tgz
  rm -rf ${BIN_PATH}/helm/emcoOpenNESS

  # monitor
  tar -cvzf  ${BIN_PATH}/helm/monitor-helm-${TAG}.tgz -C ${BIN_PATH}/helm/ monitor
  rm -rf ${BIN_PATH}/helm/monitor
}

if [ "${BUILD_CAUSE}" != "RELEASE" ];then
  if [ -z ${TAG} ]; then
    TAG=${BRANCH}-daily-`date +"%m%d%y"`
  fi
fi

# check if it is a cron scheduled build
if [ "${BUILD_CAUSE}" != "TIMERTRIGGER" ] && [ "${BUILD_CAUSE}" != "DEV_TEST" ] && [ "${BUILD_CAUSE}" != "RELEASE" ]; then
    echo "WARNING: this is not a CI build; skipping..."
    TAG="latest"
    create_helm_chart
    exit 0
fi

if [ "${BUILD_CAUSE}" == "RELEASE" ]; then
  if [ ! -z ${EMCOSRV_RELEASE_TAG} ]; then
    TAG=${EMCOSRV_RELEASE_TAG}
  else
    TAG=${TAG}
  fi
  if [ -z ${TAG} ]; then
    echo "HEAD has no tag associated with it"
    exit 0
  fi
fi

if [ "${BUILD_CAUSE}" == "DEV_TEST" ]; then
  TAG=${USER}-latest
fi

if [ "${BUILD_CAUSE}" == "TIMERTRIGGER" ] ; then
  if [ -z "${CI_COMMIT_REF_NAME}" ]; then
    CI_COMMIT_REF_NAME=${BRANCH}
  fi
  TAG=${CI_COMMIT_REF_NAME}-daily-`date +"%m%d%y"`
fi

echo "Creating docker deployment - docker-compose.yml"
mkdir -p ${EMCOBUILDROOT}/bin/docker
cp -f ${EMCOBUILDROOT}/deployments/docker/docker-compose.yml ${BIN_PATH}/docker
cat > ${BIN_PATH}/docker/.env <<EOF
REGISTRY_PREFIX=${REGISTRY}
TAG=:${TAG}
NO_PROXY=${NO_PROXY}
HTTP_PROXY=${HTTP_PROXY}
HTTPS_PROXY=${HTTPS_PROXY}
EOF

create_helm_chart
