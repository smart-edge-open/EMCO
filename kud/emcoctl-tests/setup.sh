#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

HOST_IP=${HOST_IP:-"oops"}
KUBE_PATH=${KUBE_PATH:-"oops"}
PRIVILEGED=${2:-"admin"}

# tar files
test_folder=../tests/
demo_folder=../demo/
deployment_folder=../../deployments/
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
    tar -czf output/monitor.tar.gz -C $deployment_folder/helm monitor

        cat << NET > values.yaml
    ProjectName: proj1
    ClusterProvider: provider1
    Cluster1: cluster1
    ClusterLabel: edge-cluster
    ClusterLabelNetworkPolicy: networkpolicy-supported
    Cluster1Ref: cluster1-ref
    AdminCloud: default
    PrivilegedCloud: privileged-cloud
    PrimaryNamespace: ns1
    ClusterQuota: quota1
    StandardPermission: standard-permission
    PrivilegedPermission: privileged-permission
    CompositeApp: prometheus-collectd-composite-app
    App1: prometheus-operator
    App2: collectd
    App3: operator
    App4: http-client
    App5: http-server
    AppMonitor: monitor
    KubeConfig: $KUBE_PATH
    HelmApp1: output/prometheus-operator.tar.gz
    HelmApp2: output/collectd.tar.gz
    HelmApp3: output/operator.tar.gz
    HelmApp4: output/http-client.tar.gz
    HelmApp5: output/http-server.tar.gz
    HelmAppMonitor: output/monitor.tar.gz
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
    CompositeAppMonitor: monitor-composite-app
    ConfigmapFile: info.json
    GacPort: 30493
    OvnPort: 30473
    DtcPort: 30483
    NpsPort: 30485
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

# head of prerequisites.yaml
cat << NET > prerequisites.yaml
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

---
#create project
version: emco/v2
resourceContext:
  anchor: projects
metadata :
   name: {{.ProjectName}}
---
#creating controller entries
version: emco/v2
resourceContext:
  anchor: controllers
metadata :
   name: rsync
spec:
  host:  {{.HostIP}}
  port: {{.RsyncPort}}

---
version: emco/v2
resourceContext:
  anchor: controllers
metadata :
   name: dtc
spec:
  host: {{.HostIP}}
  port: {{.DtcPort}}
  type: "action"
  priority: 1

---
#creating dtc controller entries
version: emco/v2
resourceContext:
  anchor: dtc-controllers
metadata :
   name: nps
spec:
  host:  {{.HostIP}}
  port: {{.NpsPort}}
  type: "action"
  priority: 1

---
#creating cluster provider
version: emco/v2
resourceContext:
  anchor: cluster-providers
metadata :
   name: {{.ClusterProvider}}

---
#creating cluster
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters
metadata :
   name: {{.Cluster1}}
file:
  {{.KubeConfig}}

---
#Add label cluster
version: emco/v2
resourceContext:
  anchor: cluster-providers/{{.ClusterProvider}}/clusters/{{.Cluster1}}/labels
label-name: {{.ClusterLabel}}

NET

if [ "$PRIVILEGED" = "privileged" ]; then
# rest of prerequisites.yaml for a privileged cloud
cat << NET >> prerequisites.yaml
---
#create privileged logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds
metadata:
  name: {{.PrivilegedCloud}}
spec:
  namespace: {{.PrimaryNamespace}}
  user:
    user-name: user-1
    type: certificate

---
#create cluster quotas
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.PrivilegedCloud}}/cluster-quotas
metadata:
    name: {{.ClusterQuota}}
spec:
    limits.cpu: '400'
    limits.memory: 1000Gi
    requests.cpu: '300'
    requests.memory: 900Gi
    requests.storage: 500Gi
    requests.ephemeral-storage: '500'
    limits.ephemeral-storage: '500'
    persistentvolumeclaims: '500'
    pods: '500'
    configmaps: '1000'
    replicationcontrollers: '500'
    resourcequotas: '500'
    services: '500'
    services.loadbalancers: '500'
    services.nodeports: '500'
    secrets: '500'
    count/replicationcontrollers: '500'
    count/deployments.apps: '500'
    count/replicasets.apps: '500'
    count/statefulsets.apps: '500'
    count/jobs.batch: '500'
    count/cronjobs.batch: '500'
    count/deployments.extensions: '500'

---
#add primary user permission
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.PrivilegedCloud}}/user-permissions
metadata:
    name: {{.StandardPermission}}
spec:
    namespace: {{.PrimaryNamespace}}
    apiGroups:
    - ""
    - "apps"
    - "k8splugin.io"
    resources:
    - secrets
    - pods
    - configmaps
    - services
    - deployments
    - resourcebundlestates
    verbs:
    - get
    - watch
    - list
    - create

---
#add privileged cluster-wide user permission
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.PrivilegedCloud}}/user-permissions
metadata:
    name: {{.PrivilegedPermission}}
spec:
    namespace: ""
    apiGroups:
    - "*"
    resources:
    - "*"
    verbs:
    - "*"

---
#add cluster reference to logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.PrivilegedCloud}}/cluster-references
metadata:
  name: {{.Cluster1Ref}}
spec:
  cluster-provider: {{.ClusterProvider}}
  cluster-name: {{.Cluster1}}
  loadbalancer-ip: "0.0.0.0"

NET

# instantiation.yaml specifically to instantiate a privileged logical cloud
cat << NET >> instantiation.yaml
---
#instantiate logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.PrivilegedCloud}}/instantiate

NET

else
# rest of prerequisites.yaml for an admin cloud
cat << NET >> prerequisites.yaml
---
#create admin logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds
metadata:
  name: {{.AdminCloud}}
spec:
  level: "0"

---
#add cluster reference to logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.AdminCloud}}/cluster-references
metadata:
  name: {{.Cluster1Ref}}
spec:
  cluster-provider: {{.ClusterProvider}}
  cluster-name: {{.Cluster1}}
  loadbalancer-ip: "0.0.0.0"

---
#instantiate logical cloud
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/logical-clouds/{{.AdminCloud}}/instantiate

NET

fi

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
