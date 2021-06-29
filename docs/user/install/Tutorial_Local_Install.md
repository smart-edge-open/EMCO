```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2020 Intel Corporation
```

# Getting Started
This document describes how to efficiently get started with EMCO install locally

- Build all docker images
- Run EMCO Microservices locally using docker-compose
- Deploy an Application

## Requirements
- docker (v18.09.6 or later)
- docker-compose (v1.26.2 or later)
- kubectl (v1.19.0 or later)

## Login to the EMCO Harbor registry
Remember to login to the EMCO Harbor registry by running `docker login <docker repo>` on the build machine and all cluster nodes.

Sometimes, the login fails on some Linux distributions with `Login did not succeed, error: Error response from daemon: Get <docker repo>: x509: certificate signed by unknown authority`.

To resolve this, get the latest CA chain from the organization and install it system-wide for all users with:

```
# sudo su as required
apt-get -y -qq install unzip
http_proxy='' &&\
  curl http://path/to/SHA2RootChain-Base64.zip > /tmp/SHA2RootChain-Base64.zip
unzip /tmp/SHA2RootChain-Base64.zip -d /usr/local/share/ca-certificates/
rm /tmp/SHA2RootChain-Base64.zip
update-ca-certificates
# Restart Docker - this will stop all containers
service docker restart
```

## Set proxy and docker registry
Set Docker registry parameter before building EMCO artifacts
```
export EMCODOCKERREPO=${container_registry_url}/
```

## Build docker images locally
Run ```make all``` from root EMCO folder.

## Running EMCO locally
EMCO can be deployed locally using the docker-compose file found in the ```deployments/docker``` folder.

```
docker-compose up -V
```

All running services can be verified using `docker ps`.

```
test@R90H99AZ:~$ docker ps
CONTAINER ID        IMAGE                                              COMMAND                  CREATED             STATUS              PORTS                              NAMES
417dbe7edfab        emco-ovn:latest                                    "./ovnaction"            17 seconds ago      Up 12 seconds       0.0.0.0:9051->9051/tcp             docker_ovnaction_1
a908787b4889        emco-dcm:latest                                    "./dcm"                  17 seconds ago      Up 14 seconds       0.0.0.0:9077->9077/tcp             docker_dcm_1
08f54195c92f        emco-rsync:latest                                  "./rsync"                17 seconds ago      Up 10 seconds       0.0.0.0:9031->9031/tcp             docker_rsync_1
adac112eb251        emco-ncm:latest                                    "./ncm"                  17 seconds ago      Up 13 seconds       0.0.0.0:9081->9081/tcp             docker_ncm_1
090cc2bdc1da        emco-dtc:latest                                    "./dtc"                  17 seconds ago      Up 8 seconds        0.0.0.0:9018->9018/tcp             docker_dtc_1
0b471301a944        emco-clm:latest                                    "./clm"                  17 seconds ago      Up 11 seconds       0.0.0.0:9061->9061/tcp             docker_clm_1
dcb74a0358c4        emco-orch:latest                                   "./orchestrator"         17 seconds ago      Up 15 seconds       0.0.0.0:9015->9015/tcp             docker_orchestrator_1
a01b783eed1e        emco-gac:latest                                    "./genericactioncont…"   17 seconds ago      Up 14 seconds       0.0.0.0:9020->9020/tcp             docker_genericactioncontroller_1
3cfb3f0e9a5c        <docker repo>/emco/mongo:4.4.1   "docker-entrypoint.s…"   20 seconds ago      Up 17 seconds       0.0.0.0:27017->27017/tcp           docker_mongo_1
f190d3e37912        <docker repo>/emco/etcd:3        "/entrypoint.sh etcd"    20 seconds ago      Up 18 seconds       0.0.0.0:2379-2380->2379-2380/tcp   docker_etcd_1

```

### Deploying an Application
The release artifacts includes a sample prometheus and collectd applications that can be deployed. In this section we will demonstrate how to deploy the application.

* Release artifacts for prometheus and collectd can be created as below:

```
test@R90H99AZ:~/EMCO/kud/tests$ tar -czf collectd.tar.gz -C ./vnfs/comp-app/collection/app1/helm .
test@R90H99AZ:~/EMCO/kud/tests$ tar -czf prometheus-operator.tar.gz -C ./vnfs/comp-app/collection/app2/helm .
test@R90H99AZ:~/EMCO/kud/tests$ tar -czf collectd_profile.tar.gz -C ./vnfs/comp-app/collection/app1/profile .
test@R90H99AZ:~/EMCO/kud/tests$ tar -czf prometheus-operator_profile.tar.gz -C ./vnfs/comp-app/collection/app2/profile .
```

* The emco-cfg.yaml is used to configure the hosts and ports to be used for all emco services on the local machine. The contents of emco-cfg.yaml (located in the `src/tools/emcoctl/examples` folder) should be identical to the below:

```
  orchestrator:
    host: localhost
    port: 9015
  clm:
    host: localhost
    port: 9061
  ncm:
    host: localhost
    port: 9081
  ovnaction:
    host: localhost
    port: 9051
  dcm:
    host: localhost
    port: 9077
  gac:
    host: localhost
    port: 9020
  dtc:
    host: localhost
    port: 9048

```

* The test.yaml has details on the workloads to be deployed and the clusters to be used. Modify the following in the test.yaml (located in the `src/tools/emcoctl/examples` folder):
  * Update the kubeconfig path for the `cluster1` resource.
  * Update the directories where the workload helm charts (i.e. collectd.tar.gz, etc.) are located for the `prometheus-operator`, `collectd`, `prometheus-profile` & `collectd-profile` resources.


* Then deploy the workload using emcoctl.

```
EMCO/bin/emcoctl$ ./emcoctl --config ../../src/tools/emcoctl/examples/emco-cfg.yaml apply -f ../../src/tools/emcoctl/examples/test.yaml

```

* To check the status of deployment of application use below command

```
./emcoctl --config ../../src/tools/emcoctl/examples/emco-cfg.yaml get projects/proj1/composite-apps/collection-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/status
```

Status should look like this

```
Using config file: ../../src/tools/emcoctl/examples/emco-cfg.yaml
http://0.0.0.0:9015/v2URL: projects/proj1/composite-apps/collection-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/status Response Code: 200 Response: {"project":"proj1","composite-app-name":"collection-composite-app","composite-app-version":"v1","composite-profile-name":"collection-composite-profile","name":"collection-deployment-intent-group","states":{"actions":[{"state":"Created","instance":"","time":"2020-11-05T21:57:29.731Z"},{"state":"Approved","instance":"","time":"2020-11-05T21:57:29.774Z"},{"state":"Instantiated","instance":"466646771139515941","time":"2020-11-05T21:57:33.057Z"}]},"status":"Instantiated","rsync-status":{"Applied":97},"apps":[{"name":"prometheus-operator","clusters":[{"cluster-provider":"provider1","cluster":"cluster1","resources":[{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"alertmanagers.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"podmonitors.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"prometheuses.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"prometheusrules.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"apps","Version":"v1","Kind":"Deployment"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"policy","Version":"v1beta1","Kind":"PodSecurityPolicy"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1beta1","Kind":"Role"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1beta1","Kind":"RoleBinding"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Secret"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ServiceAccount"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-grafana-clusterrole","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-grafana-clusterrolebinding","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-grafana-config-dashboards","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-alertmanager.rules","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-apiserver","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-apiserver","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-cluster-total","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-controller-manager","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-coredns","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-coredns","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-etcd","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-etcd","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-general.rules","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-grafana-datasource","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-coredns","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-cluster","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-namespace","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-node","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-pod","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-workload","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-workloads-namespace","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-k8s.rules","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kube-apiserver.rules","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-controller-manager","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-controller-manager","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-etcd","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-etcd","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kube-prometheus-general.rules","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kube-prometheus-node-recording.rules","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-proxy","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-proxy","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-scheduler","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-scheduler","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kube-scheduler.rules","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kube-state-metrics","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-kubelet","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kubelet","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubelet.rules","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-resources","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-storage","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-system","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-system-apiserver","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-system-controller-manager","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-system-kubelet","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-system-scheduler","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-namespace-by-pod","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-namespace-by-workload","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-node-network","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-nodes","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"apps","Version":"v1","Kind":"Deployment"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"policy","Version":"v1beta1","Kind":"PodSecurityPolicy"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ServiceAccount"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"batch","Version":"v1","Kind":"Job"},"name":"r1-prometheus-operator-operator-cleanup","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-operator-psp","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-operator-psp","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-persistentvolumesusage","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-pod-total","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"policy","Version":"v1beta1","Kind":"PodSecurityPolicy"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"Prometheus"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ServiceAccount"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-prometheus-operator","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-prometheus-psp","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-prometheus-psp","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-prometheus-remote-write","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Secret"},"name":"r1-prometheus-operator-prometheus-scrape-confg","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-proxy","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-scheduler","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-statefulset","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-workload-total","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"service-monitor-collectd","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"servicemonitors.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"thanosrulers.monitoring.coreos.com","rsync-status":"Applied"}]}]},{"name":"collectd","clusters":[{"cluster-provider":"provider1","cluster":"cluster1","resources":[{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"collectd","rsync-status":"Applied"},{"GVK":{"Group":"apps","Version":"v1","Kind":"DaemonSet"},"name":"r1-collectd","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-collectd-config","rsync-status":"Applied"}]}]}]}
```

* Check the status on kubernetes cluster
```
kubectl --kubeconfig=/path/to/kubeconfig get all -A | grep -i collectd
```

```
default                 pod/r1-collectd-g9brg                                  1/1     Running   0          51s
default                 pod/r1-collectd-lm8mg                                  1/1     Running   0          6s
default                 service/collectd                                         ClusterIP      10.233.5.90     <none>        9104/TCP                                                                                                                                     13d
default       daemonset.apps/r1-collectd    2         2         2       2            2           <none>                        13d
```
