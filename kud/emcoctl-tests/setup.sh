#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH=${KUBE_PATH:-"oops"}
# tar files
test_folder=../tests/
demo_folder=../demo/
function create {
    mkdir -p output
    tar -czf output/collectd.tar.gz -C $test_folder/vnfs/comp-app/collection/app1/helm .
    tar -czf output/collectd_profile.tar.gz -C $test_folder/vnfs/comp-app/collection/app1/profile .
    tar -czf output/prometheus-operator.tar.gz -C $test_folder/vnfs/comp-app/collection/app2/helm .
    tar -czf output/prometheus-operator_profile.tar.gz -C $test_folder/vnfs/comp-app/collection/app2/profile .
    tar -czf output/operator.tar.gz -C $test_folder/vnfs/comp-app/collection/operators-latest/helm .
    tar -czf output/operator_profile.tar.gz -C $test_folder/vnfs/comp-app/collection/operators-latest/profile .
    tar -czf output/m3db.tar.gz -C $test_folder/vnfs/comp-app/collection/m3db/helm .
    tar -czf output/m3db_profile.tar.gz -C $test_folder/vnfs/comp-app/collection/m3db/profile .
    tar -czf output/http-client.tar.gz -C $test_folder/helm_charts/dtc http-client
    tar -czf output/http-server.tar.gz -C $test_folder/helm_charts/dtc http-server
    tar -czf output/http-server-profile.tar.gz -C $test_folder/helm_charts/dtc/network_policy_overrides/http-server-profile .
    tar -czf output/http-client-profile.tar.gz -C $test_folder/helm_charts/dtc/network_policy_overrides/http-client-profile .
    tar -czf output/firewall.tar.gz -C $demo_folder/composite-firewall firewall
    tar -czf output/packetgen.tar.gz -C $demo_folder/composite-firewall packetgen
    tar -czf output/sink.tar.gz -C $demo_folder/composite-firewall sink
    tar -czf output/profile.tar.gz -C $demo_folder/composite-firewall manifest.yaml override_values.yaml

        cat << NET > values.yaml
    ProjectName: proj1
    ClusterProvider: provider1
    Cluster1: cluster1
    ClusterLabel: edge-cluster
    ClusterLabelNetworkPolicy: networkpolicy-supported
    AdminCloud: default
    CompositeApp: prometheus-collectd-composite-app
    App1: prometheus-operator
    App2: collectd
    App3: operator
    App4: http-client
    App5: http-server
    KubeConfig: $KUBE_PATH
    HelmApp1: output/prometheus-operator.tar.gz
    HelmApp2: output/collectd.tar.gz
    HelmApp3: output/operator.tar.gz
    HelmApp4: output/http-client.tar.gz
    HelmApp5: output/http-server.tar.gz
    HelmAppFirewall: output/firewall.tar.gz
    HelmAppPacketgen: output/packetgen.tar.gz
    HelmAppSink: output/sink.tar.gz
    ProfileFw: output/profile.tar.gz
    ProfileApp1: output/prometheus-operator_profile.tar.gz
    ProfileApp2: output/collectd_profile.tar.gz
    ProfileApp3: output/operator_profile.tar.gz
    ProfileApp4: output/http-client-profile.tar.gz
    ProfileApp5: output/http-server-profile.tar.gz
    CompositeProfile: collection-composite-profile
    GenericPlacementIntent: collection-placement-intent
    DeploymentIntent: collection-deployment-intent-group
    RsyncPort: 30441
    CompositeAppGac: gac-composite-app
    GacIntent: collectd-gac-intent
    CompositeAppDtc: dtc-composite-app
    DtcIntent: collectd-dtc-intent
    ConfigmapFile: info.json
    GacPort: 30493
    OvnPort: 30473
    DtcPort: 30483
    HostIP: $HOST_IP

NET
cat << NET > emco-cfg.yaml
  orchestrator:
    host: $HOST_IP
    port: 30415
  clm:
    host: $HOST_IP
    port: 30461
  ncm:
    host: $HOST_IP
    port: 30431
  ovnaction:
    host: $HOST_IP
    port: 30471
  dcm:
    host: $HOST_IP
    port: 30477
  gac:
    host: $HOST_IP
    port: 30491
  dtc:
   host: $HOST_IP
   port: 30481
NET

}

function usage {
    echo "Usage: $0  create|cleanup"
}

function cleanup {
    rm -f *.tar.gz
    rm -f values.yaml
    rm -f emco-cfg.yaml
    rm -rf output
}

if [ "$#" -lt 1 ] ; then
    usage
    exit
fi

case "$1" in
    "create" )
        if [ "${HOST_IP}" == "oops" ] || [ "${KUBE_PATH}" == "oops" ] ; then
            echo -e "ERROR - HOST_IP & KUBE_PATH environment variable needs to be set"
        else
            create
        fi
        ;;
    "cleanup" )
        cleanup
    ;;
    *)
        usage ;;
esac
