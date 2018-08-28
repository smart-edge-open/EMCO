
```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2020 Intel Corporation
```

**NOTE**: The helmcharts for CI are located in the `deployments/helm/emcoCI` folder and only this version is referenced in the top-level Makefile. The helmcharts in the `deployments/helm/emcoOpenNESS` folder are not actively maintained and are included for legacy purposes.

# Getting Started
This document describes how to efficiently get started with EMCO install using Helm Package.

- Create Helm package
- Installing and configuring EMCO Microservices using Helm
- Deploy an Application

## Requirements
- docker (v18.09.6 or later)
- helm (v3.3.4 or later)
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
Set Docker registry parameter before creating helm package
```
export EMCODOCKERREPO=${container_registry_url}/

```
## Create Helm Package
Run `make deploy` from root EMCO folder to create the helmchart tar package under the `EMCO/bin/helm` folder.

## Installing EMCO on the cluster
EMCO can be installed and deployed using the provided Helm chart included in the build artifacts to a Kubernetes cluster.

To install EMCO , navigate to the ```EMCO/bin/helm``` directory. The ```emco-helm-install.sh``` script is used to deploy EMCO to the target cluster.

```
   ./emco-helm-install.sh <optional: -s parameter=value> <optional: -k <path to kubeconfig file>> [install | uninstall]

```
The -s option with a parameter and value may be supplied multiple times to override multiple Helm values.

In the example below, we install to a Kubernetes cluster using the kube-config associated with that cluster.
Note that by default, authentication for the EMCO Mongo and Etcd databases is enabled, so values for passwords must be provided.

```
./emco-helm-install.sh -s db.rootPassword=<pw> -s db.emcoPassword=<pw> -s contextdb.rootPassword=<pw> -s contextdb.emcoPassword=<pw> -k /home/test/EMCO/deployments/kubernetes/config_north install
Creating namespace emco
namespace/emco created
Installing EMCO. Please wait...
WARNING: Kubernetes configuration file is group-readable. This is insecure. Location: /home/test/EMCO/deployments/kubernetes/config_north
WARNING: Kubernetes configuration file is world-readable. This is insecure. Location: /home/test/EMCO/deployments/kubernetes/config_north
NAME: emco
LAST DEPLOYED: Fri Oct  2 12:42:23 2020
NAMESPACE: emco
STATUS: deployed
REVISION: 1
TEST SUITE: None
Done

```

EMCO is installed into the ```emco``` namespace. To verify that all the services are up and running kubectl can be used to list the services running in that namespace.

```
[root@nb-cluster7-dccf1 ~]# kubectl get all -n emco
NAME                               READY   STATUS    RESTARTS   AGE
pod/clm-6c8dd6966b-7xf27           1/1     Running   0          7m48s
pod/dcm-79b7877dff-hvmqm           1/1     Running   0          7m48s
pod/dtc-86bf678fdb-ds7x8           1/1     Running   0          7m48s
pod/etcd-ff4bc67d8-g8k92           1/1     Running   0          7m48s
pod/gac-74955676fb-s8tq2           1/1     Running   0          7m48s
pod/mongo-646d44db67-p9fv2         1/1     Running   0          7m48s
pod/ncm-6cf6647cf6-pfmmb           1/1     Running   0          7m48s
pod/orchestrator-8df787485-b9scr   1/1     Running   0          7m48s
pod/ovnaction-6dc486b44d-mgljs     1/1     Running   0          7m48s
pod/rsync-84f5f6c876-4tqgl         1/1     Running   0          7m48s

NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
service/clm            NodePort    10.233.2.191    <none>        9061:31856/TCP                  7m48s
service/dcm            NodePort    10.233.13.243   <none>        9077:31877/TCP                  7m48s
service/dtc            NodePort    10.233.52.127   <none>        9053:32656/TCP,9018:31182/TCP   7m48s
service/etcd           ClusterIP   10.233.62.178   <none>        2379/TCP,2380/TCP               7m48s
service/gac            NodePort    10.233.15.19    <none>        9021:30907/TCP,9020:31280/TCP   7m48s
service/mongo          ClusterIP   10.233.45.165   <none>        27017/TCP                       7m48s
service/ncm            NodePort    10.233.55.75    <none>        9081:32737/TCP                  7m48s
service/orchestrator   NodePort    10.233.1.14     <none>        9015:31298/TCP                  7m48s
service/ovnaction      NodePort    10.233.2.185    <none>        9032:31307/TCP,9051:31181/TCP   7m48s
service/rsync          NodePort    10.233.42.242   <none>        9031:32651/TCP                  7m48s

NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/clm            1/1     1            1           7m48s
deployment.apps/dcm            1/1     1            1           7m48s
deployment.apps/dtc            1/1     1            1           7m48s
deployment.apps/etcd           1/1     1            1           7m48s
deployment.apps/gac            1/1     1            1           7m48s
deployment.apps/mongo          1/1     1            1           7m48s
deployment.apps/ncm            1/1     1            1           7m48s
deployment.apps/orchestrator   1/1     1            1           7m48s
deployment.apps/ovnaction      1/1     1            1           7m48s
deployment.apps/rsync          1/1     1            1           7m48s

NAME                                     DESIRED   CURRENT   READY   AGE
replicaset.apps/clm-6c8dd6966b           1         1         1       7m48s
replicaset.apps/dcm-79b7877dff           1         1         1       7m48s
replicaset.apps/dtc-86bf678fdb           1         1         1       7m48s
replicaset.apps/etcd-ff4bc67d8           1         1         1       7m48s
replicaset.apps/gac-74955676fb           1         1         1       7m48s
replicaset.apps/mongo-646d44db67         1         1         1       7m48s
replicaset.apps/ncm-6cf6647cf6           1         1         1       7m48s
replicaset.apps/orchestrator-8df787485   1         1         1       7m48s
replicaset.apps/ovnaction-6dc486b44d     1         1         1       7m48s
replicaset.apps/rsync-84f5f6c876         1         1         1       7m48s

```

### Deploying an Application
The release artifacts includes a sample promethues and collectd applications that can be deployed. In this section we will demonstrate how to deploy the application.

* Release artifacts for prometheus and collectd can be created as below:

```
test@R90H99AZ:~/EMCO/kud/tests$ tar -czf collectd.tar.gz -C ./vnfs/comp-app/collection/app1/helm .
test@R90H99AZ:~/EMCO/kud/tests$ tar -czf prometheus-operator.tar.gz -C ./vnfs/comp-app/collection/app2/helm .
test@R90H99AZ:~/EMCO/kud/tests$ tar -czf collectd_profile.tar.gz -C ./vnfs/comp-app/collection/app1/profile .
test@R90H99AZ:~/EMCO/kud/tests$ tar -czf prometheus-operator_profile.tar.gz -C ./vnfs/comp-app/collection/app2/profile .
```

* The emco-cfg-remote.yaml is used to configure the hosts and ports to be used for all emco services on the remote cluster. The file (located in the `src/tools/emcoctl/examples` folder) will look like below and needs to be updated to reflect one's config:

```
  orchestrator:
    host: 10.23.208.71
    port: 42298
  clm:
    host: 10.23.208.71
    port: 42856
  ncm:
    host: 10.23.208.71
    port: 42737
  ovnaction:
    host: 10.23.208.71
    port: 42181
  dcm:
    host: 10.23.208.71
    port: 42877

```

* The test.yaml has details on the workloads to be deployed and the clusters to be used. Modify the following in the test.yaml (located in the `src/tools/emcoctl/examples` folder):
  * Update the kubeconfig path for the `cluster1` resource.
  * Update the directories where the workload helm charts (i.e. collectd.tar.gz, etc.) are located for the `prometheus-operator`, `collectd`, `prometheus-profile` & `collectd-profile` resources.

* Then deploy the workload using emcoctl.

```
EMCO/bin/emcoctl$ ./emcoctl --config ../../src/tools/emcoctl/examples/emco-cfg-remote.yaml apply -f ../../src/tools/emcoctl/examples/test.yaml

```

* To check the status of deployment of application use below command

```
./emcoctl --config ../../src/tools/emcoctl/examples/emco-cfg-remote.yaml get projects/proj1/composite-apps/collection-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/status
```

Status should look like this

```
Using config file: ../../src/tools/emcoctl/examples/emco-cfg-remote.yaml
http://10.23.208.71:42298/v2URL: projects/proj1/composite-apps/collection-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/status Response Code: 200 Response: {"project":"proj1","composite-app-name":"collection-composite-app","composite-app-version":"v1","composite-profile-name":"collection-composite-profile","name":"collection-deployment-intent-group","states":{"actions":[{"state":"Created","instance":"","time":"2020-11-05T22:50:04.83Z"},{"state":"Approved","instance":"","time":"2020-11-05T22:50:04.853Z"},{"state":"Instantiated","instance":"2214358778903857349","time":"2020-11-05T22:50:06.306Z"}]},"status":"Instantiated","rsync-status":{"Applied":97},"apps":[{"name":"prometheus-operator","clusters":[{"cluster-provider":"provider1","cluster":"cluster1","resources":[{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"alertmanagers.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"podmonitors.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"prometheuses.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"prometheusrules.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"apps","Version":"v1","Kind":"Deployment"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"policy","Version":"v1beta1","Kind":"PodSecurityPolicy"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1beta1","Kind":"Role"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1beta1","Kind":"RoleBinding"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Secret"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ServiceAccount"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-grafana-clusterrole","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-grafana-clusterrolebinding","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-grafana-config-dashboards","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-alertmanager.rules","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-apiserver","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-apiserver","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-cluster-total","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-controller-manager","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-coredns","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-coredns","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-etcd","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-etcd","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-general.rules","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-grafana-datasource","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-coredns","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-cluster","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-namespace","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-node","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-pod","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-workload","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-k8s-resources-workloads-namespace","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-k8s.rules","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kube-apiserver.rules","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-controller-manager","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-controller-manager","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-etcd","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-etcd","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kube-prometheus-general.rules","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kube-prometheus-node-recording.rules","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-proxy","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-proxy","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-scheduler","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-scheduler","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kube-scheduler.rules","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kube-state-metrics","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-kubelet","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kubelet","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubelet.rules","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-resources","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-storage","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-system","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-system-apiserver","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-system-controller-manager","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-system-kubelet","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-kubernetes-system-scheduler","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-namespace-by-pod","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-namespace-by-workload","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-node-network","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-nodes","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"apps","Version":"v1","Kind":"Deployment"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"policy","Version":"v1beta1","Kind":"PodSecurityPolicy"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ServiceAccount"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"batch","Version":"v1","Kind":"Job"},"name":"r1-prometheus-operator-operator-cleanup","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-operator-psp","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-operator-psp","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-persistentvolumesusage","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-pod-total","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"policy","Version":"v1beta1","Kind":"PodSecurityPolicy"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"Prometheus"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ServiceAccount"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"PrometheusRule"},"name":"r1-prometheus-operator-prometheus-operator","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-prometheus-psp","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-prometheus-psp","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-prometheus-remote-write","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Secret"},"name":"r1-prometheus-operator-prometheus-scrape-confg","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-proxy","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-scheduler","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-statefulset","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-workload-total","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"service-monitor-collectd","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"servicemonitors.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"thanosrulers.monitoring.coreos.com","rsync-status":"Applied"}]}]},{"name":"collectd","clusters":[{"cluster-provider":"provider1","cluster":"cluster1","resources":[{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"collectd","rsync-status":"Applied"},{"GVK":{"Group":"apps","Version":"v1","Kind":"DaemonSet"},"name":"r1-collectd","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-collectd-config","rsync-status":"Applied"}]}]}]}
```

* Check the status on kubernetes cluster
```
kubectl --kubeconfig=/path/to/kubeconfig get all -A | grep -i collectd
```

```
default                 pod/r1-collectd-pc6f7                                 1/1     Running   0          12s
default                 pod/r1-collectd-vhjgf                                 1/1     Running   0          57s
default                 service/collectd                                         ClusterIP      10.233.5.90     <none>        9104/TCP                                                                                                                                     13d
default       daemonset.apps/r1-collectd    2         2         2       2            2           <none>                        13d
```

### Cleanup
To uninstall EMCO use the same script and execute the following command.
```
./emco-helm-install.sh -k /home/test/EMCO/deployments/kubernetes/config_north uninstall
Removing EMCO...
WARNING: Kubernetes configuration file is group-readable. This is insecure. Location: /home/test/EMCO/deployments/kubernetes/config_north
WARNING: Kubernetes configuration file is world-readable. This is insecure. Location: /home/test/EMCO/deployments/kubernetes/config_north
release "emco" uninstalled
Deleting namespace emco
namespace "emco" deleted
Done
```

### Pushing EMCO images to Harbor
It is often required to use developer tags for locally built EMCO images, push to Harbor and reference these custom images in helmcharts for developer testing.

To enable these features, run `export BUILD_CAUSE=DEV_TEST` prior to running `make deploy`. This will tag locally built images as `<username>-latest`, push to Harbor and reference these images in the generated helmcharts.

Developers can then use the ```emco-helm-install.sh``` script from the above sections to install these custom images on their cluster.
