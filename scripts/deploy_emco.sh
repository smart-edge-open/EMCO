#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

REGISTRY=${EMCODOCKERREPO}
BRANCH=`git rev-parse --abbrev-ref HEAD`
BIN_PATH=${EMCOBUILDROOT}/bin
TAG=$1

push_to_registry() {
    M_IMAGE=$1
    M_TAG=$2
    echo "Pushing ${M_IMAGE} to ${REGISTRY}${M_IMAGE}:${M_TAG}..."
    docker tag ${M_IMAGE}:latest ${REGISTRY}${M_IMAGE}:${M_TAG}
    docker push ${REGISTRY}${M_IMAGE}:${M_TAG}
}

create_helm_chart() {
  echo "Creating helm chart"
  mkdir -p ${BIN_PATH}/helm
  cp -rf ${EMCOBUILDROOT}/deployments/helm/emcoCI ${EMCOBUILDROOT}/deployments/helm/monitor ${BIN_PATH}/helm/
  cat > ${BIN_PATH}/helm/emcoCI/values.yaml <<EOF
registryPrefix: ${REGISTRY}
tag: ${TAG}
noProxyHosts: localhost,127.0.0.1,0.0.0.0
enableDbAuth: true
db:
  rootUser: admin
#  rootPassword: <provide password as an override value>
  emcoUser: emco
#  emcoPassword: <provide password as an override value>
contextdb:
  rootUser: root
#  rootPassword: <provide password as an override value>
  emcoUser: emco
#  emcoPassword: <provide password as an override value>
EOF
  cat > ${BIN_PATH}/helm/monitor/values.yaml <<EOF
registryPrefix: ${REGISTRY}
tag: ${TAG}
EOF
  cat > ${BIN_PATH}/helm/helm_value_overrides.yaml <<EOF
#update proxies
noProxyHosts: localhost,127.0.0.1,0.0.0.0
#update and uncomment if build tag to be changed
#tag: latest
#update and uncomment to override registry
#registryPrefix: registry.docker.com/
#enableDbAuth: true
#db:
#  rootUser: admin
#  rootPassword: <provide password as an override value>
#  emcoUser: emco
#  emcoPassword: <provide password as an override value>
#contextdb:
#  rootUser: root
#  rootPassword: <provide password as an override value>
#  emcoUser: emco
#  emcoPassword: <provide password as an override value>
EOF
  
  # emcoCI
  cp ${EMCOBUILDROOT}/deployments/helm/emco-helm-install.sh ${BIN_PATH}/helm/install_template
  cat ${BIN_PATH}/helm/install_template | sed "s/emco-helm.tgz/emco-helm-${TAG}.tgz/" > ${BIN_PATH}/helm/emco-helm-install.sh
  chmod +x ${BIN_PATH}/helm/emco-helm-install.sh
  rm -f ${BIN_PATH}/helm/install_template
  tar -cvzf  ${BIN_PATH}/helm/emco-helm-${TAG}.tgz -C ${BIN_PATH}/helm/ emcoCI
  rm -rf ${BIN_PATH}/helm/emcoCI

  # monitor
  tar -cvzf  ${BIN_PATH}/helm/monitor-helm-${TAG}.tgz -C ${BIN_PATH}/helm/ monitor
  rm -rf ${BIN_PATH}/helm/monitor
}


if [ -z ${TAG} ]; then
  TAG=${BRANCH}-daily-`date +"%m%d%y"`
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
    TAG=`git tag --points-at HEAD | awk 'NR==1 {print $1}'`
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

[[ -z "$MODS" ]] && export MODS="clm dcm dtc nps gac monitor ncm orch ovn rsync"
MODS=$(echo $MODS | sed 's;tools/emcoctl;;')
for m in $MODS; do
    push_to_registry emco-$m ${TAG}
done

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
