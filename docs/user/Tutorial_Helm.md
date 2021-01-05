
```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2020 Intel Corporation
```

**NOTE**: The helmcharts for CI are located in the `deployments/helm/emcoOpenNESS` folder and only this version is referenced in the top-level Makefile. The helmcharts in the `deployments/helm/emcoCI` folder are not actively maintained and are included for legacy purposes.

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
## Set proxy, docker registry and log level
Set Docker registry parameter before creating helm package
```
export EMCODOCKERREPO=${container_registry_url}/

```
## Create Helm Package
Run `make deploy` from root EMCO folder to create the helmchart tar package under the `EMCO/bin/helm` folder.

## Installing EMCO on the cluster
EMCO can be installed and deployed using the provided Helm chart included in the build artifacts to a Kubernetes cluster.

To install EMCO , navigate to the ```EMCO/bin/helm``` directory. The ```emco-openness-helm-install.sh``` script is used to deploy EMCO to the target cluster.

```
   ./emco-openness-helm-install.sh <optional: -k <path to kubeconfig file>> <optional: -p [enable | disable]> [install | uninstall]

```
The -p option enables or disables persistence for the etcd and mongo storage.  The default value is `disable`.

The -s option with a parameter and value may be supplied multiple times to override multiple Helm values.

In the example below, we install to a Kubernetes cluster using the kube-config associated with that cluster.

```
./emco-openness-helm-install.sh -k /home/test/EMCO/deployments/kubernetes/config_north install
Creating namespace emco
namespace/emco created
Installing EMCO DB. Please wait...
WARNING: Kubernetes configuration file is group-readable. This is insecure. Location: /home/test/EMCO/deployments/kubernetes/config_north
WARNING: Kubernetes configuration file is world-readable. This is insecure. Location: /home/test/EMCO/deployments/kubernetes/config_north
NAME: emco-db
LAST DEPLOYED: Thu Feb  4 17:05:12 2021
NAMESPACE: emco
STATUS: deployed
REVISION: 1
TEST SUITE: None
Done
Installing EMCO Services. Please wait...
WARNING: Kubernetes configuration file is group-readable. This is insecure. Location: /home/test/EMCO/deployments/kubernetes/config_north
WARNING: Kubernetes configuration file is world-readable. This is insecure. Location: /home/test/EMCO/deployments/kubernetes/config_north
NAME: emco-services
LAST DEPLOYED: Thu Feb  4 17:05:30 2021
NAMESPACE: emco
STATUS: deployed
REVISION: 1
TEST SUITE: None
Done
Installing EMCO Tools. Please wait...
WARNING: Kubernetes configuration file is group-readable. This is insecure. Location: /home/test/EMCO/deployments/kubernetes/config_north
WARNING: Kubernetes configuration file is world-readable. This is insecure. Location: /home/test/EMCO/deployments/kubernetes/config_north
NAME: emco-tools
LAST DEPLOYED: Thu Feb  4 17:06:28 2021
NAMESPACE: emco
STATUS: deployed
REVISION: 1
TEST SUITE: None
Done

```

EMCO is installed into the ```emco``` namespace. To verify that all the services are up and running kubectl can be used to list the services running in that namespace.

```
[root@nb-cluster7-dccf1 ~]# kubectl get all -n emco
NAME                                              READY   STATUS    RESTARTS   AGE
pod/emco-db-emco-etcd-0                           1/1     Running   0          3m25s
pod/emco-db-emco-mongo-0                          1/1     Running   0          3m25s
pod/emco-services-clm-5654d875b8-cdxb8            1/1     Running   0          3m8s
pod/emco-services-dcm-79c5847bf-krq2r             1/1     Running   0          3m8s
pod/emco-services-dtc-688768587-77899             1/1     Running   0          3m8s
pod/emco-services-gac-57cb4f59b8-rlbxm            1/1     Running   0          3m8s
pod/emco-services-ncm-8459494874-c2lsr            1/1     Running   0          3m8s
pod/emco-services-orchestrator-5c586d7d49-2qq88   1/1     Running   0          3m8s
pod/emco-services-ovnaction-d9d5bb5cb-clpml       1/1     Running   0          3m8s
pod/emco-services-rsync-c94fdbd74-99f9w           1/1     Running   0          3m8s
pod/emco-tools-fluentd-0                          1/1     Running   0          2m10s
pod/emco-tools-fluentd-ddsv7                      1/1     Running   1          2m10s
pod/emco-tools-fluentd-k8hks                      1/1     Running   2          2m10s

NAME                                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
service/clm                             NodePort    10.233.36.44    <none>        9061:30461/TCP                  3m8s
service/dcm                             NodePort    10.233.53.3     <none>        9077:30477/TCP                  3m8s
service/dtc                             NodePort    10.233.36.32    <none>        9048:30483/TCP,9018:30481/TCP   3m8s
service/emco-etcd                       ClusterIP   None            <none>        2380/TCP,2379/TCP               3m25s
service/emco-mongo                      ClusterIP   None            <none>        27017/TCP                       3m25s
service/emco-mongo-read                 ClusterIP   10.233.45.231   <none>        27017/TCP                       3m25s
service/emco-tools-fluentd-aggregator   ClusterIP   10.233.14.213   <none>        24224/TCP                       2m11s
service/emco-tools-fluentd-forwarder    ClusterIP   10.233.21.81    <none>        9880/TCP                        2m11s
service/emco-tools-fluentd-headless     ClusterIP   None            <none>        24224/TCP                       2m11s
service/gac                             NodePort    10.233.16.181   <none>        9033:30493/TCP,9020:30491/TCP   3m8s
service/ncm                             NodePort    10.233.11.199   <none>        9081:30431/TCP                  3m8s
service/orchestrator                    NodePort    10.233.54.64    <none>        9015:30415/TCP                  3m8s
service/ovnaction                       NodePort    10.233.42.62    <none>        9053:30473/TCP,9051:30471/TCP   3m8s
service/rsync                           NodePort    10.233.5.126    <none>        9031:30441/TCP                  3m8s

NAME                                DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/emco-tools-fluentd   2         2         2       2            2           <none>          2m11s

NAME                                         READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/emco-services-clm            1/1     1            1           3m8s
deployment.apps/emco-services-dcm            1/1     1            1           3m8s
deployment.apps/emco-services-dtc            1/1     1            1           3m8s
deployment.apps/emco-services-gac            1/1     1            1           3m8s
deployment.apps/emco-services-ncm            1/1     1            1           3m8s
deployment.apps/emco-services-orchestrator   1/1     1            1           3m8s
deployment.apps/emco-services-ovnaction      1/1     1            1           3m8s
deployment.apps/emco-services-rsync          1/1     1            1           3m8s

NAME                                                    DESIRED   CURRENT   READY   AGE
replicaset.apps/emco-services-clm-5654d875b8            1         1         1       3m8s
replicaset.apps/emco-services-dcm-79c5847bf             1         1         1       3m8s
replicaset.apps/emco-services-dtc-688768587             1         1         1       3m8s
replicaset.apps/emco-services-gac-57cb4f59b8            1         1         1       3m8s
replicaset.apps/emco-services-ncm-8459494874            1         1         1       3m8s
replicaset.apps/emco-services-orchestrator-5c586d7d49   1         1         1       3m8s
replicaset.apps/emco-services-ovnaction-d9d5bb5cb       1         1         1       3m8s
replicaset.apps/emco-services-rsync-c94fdbd74           1         1         1       3m8s

NAME                                  READY   AGE
statefulset.apps/emco-db-emco-etcd    1/1     3m26s
statefulset.apps/emco-db-emco-mongo   1/1     3m26s
statefulset.apps/emco-tools-fluentd   1/1     2m11s
```

### Database Authentication

When the EMCO databases `etcd` and `mongo` are deployed, username / password authentication is enabled by default.  The
EMCO services will be configured with the credentials to access the databases.

If password values are not overridden during installation, the helm install process will create random passwords.
The passwords are stored in the secrets:

```
$ kubectl -n emco get secret
NAME                            TYPE                                  DATA   AGE
emco-etcd                       Opaque                                1      3s
emco-mongo                      Opaque                                2      3s
```

The following values can be provided on installation to override values:

- `global.db.rootPassword` - set to override mongo root password (default is random password)
- `global.db.emcoPassword` - set to override mongo user password (default is random password)
- `global.contextdb.rootPassword` - set to override etcd password (default is random password)
- `global.disableDbAuth` - set to `true` to not use database authentication (default is `false`)

Note: that the previous EMCO release (20.12) provided the following values.  The `emco-openness-helm-install.sh`
script will convert these legacy value names to the new value names to provide backward compatibility.
- `db.rootPassword` - set to override mongo root password (default is random password)
- `db.emcoPassword` - set to override mongo user password (default is random password)
- `contextdb.rootPassword` - set to override etcd password (default is random password)
- `enableDbAuth` - set to `true` to enable database authentication (default is `true`)

Note: the current release only uses `contextdb.rootPassword` for the `contextdb` (i.e. `etcd`).

#### Installation examples with various database authentication options:

Install EMCO with database authentication and default (random) passwords, persistence is disabled:

`./emco-openness-helm-install.sh -k <path to kubeconfig file> install`

Install EMCO with database authentication and override password values, persistence is disabled:

`./emco-openness-helm-install.sh -s global.db.rootPassword=abc -s global.db.emcoPassword=def -s global.contextdb.rootPassword=xyz -k <path to kubeconfig file> install`

Install EMCO with database authentication enabled and override password values using legacy password value names, persistence is disabled:

`./emco-openness-helm-install.sh -s db.rootPassword=abc -s db.emcoPassword=def -s contextdb.rootPassword=xyz -k <path to kubeconfig file> install`

Another example using legacy password value names, persistence is disabled:

`./emco-openness-helm-install.sh -s 'enableDbAuth=true --timeout=30m --set db.rootPassword=abc --set db.emcoPassword=def --set contextdb.rootPassword=xyz --set contextdb.emcoPassword=xyz' -k <path to kubeconfig file> install`

Install EMCO with database authentication and override password values, enable persistence:

`./emco-openness-helm-install.sh -s global.db.rootPassword=abc -s global.db.emcoPassword=def -s global.contextdb.rootPassword=xyz -p enable -k <path to kubeconfig file> install`

#### Known issues with database authentication and persistence enabled

If persistence is enabled, then care needs to be taken with the database authentication passwords.  If a new install of EMCO re-uses persistent data from a previous installation and the database authentication configuration
has changed - e.g. authentication enabled/disabled, or the passwords have changed - then the EMCO services may fail to start up successfully.

Workarounds:

- Uninstall EMCO and then remove the host storage directories for the persistent volumes and then reinstall:

```
    - `sudo rm -r /dockerdata-nfs/emco-db/emco/mongo/data`
    - `sudo rm -r /dockerdata-nfs/emco-db/emco/etcd/data-0`
```

- Or, disable database persistence on installation:
    - Remove the `-p enable` option from installation, or set `-p disable`

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
  gac:
    host: 10.23.208.71
    port: 42887
  dtc:
    host: 10.23.208.71
    port: 42897

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
---
GET  --> URL: http://10.23.208.71:49006/v2/projects/proj1/composite-apps/collection-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/status
Response Code: 200
Response: {"project":"proj1","composite-app-name":"collection-composite-app","composite-app-version":"v1","composite-profile-name":"collection-composite-profile","name":"collection-deployment-intent-group","states":{"actions":[{"state":"Created","instance":"","time":"2021-01-31T02:23:42.361Z"},{"state":"Approved","instance":"","time":"2021-01-31T02:23:42.383Z"},{"state":"Instantiated","instance":"1408427859572459654","time":"2021-01-31T02:23:44.255Z"}]},"status":"Instantiating","rsync-status":{"Applied":18,"Pending":36},"apps":[{"name":"collectd","clusters":[{"cluster-provider":"provider1","cluster":"cluster1","resources":[{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"collectd","rsync-status":"Applied"},{"GVK":{"Group":"apps","Version":"v1","Kind":"DaemonSet"},"name":"r1-collectd","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-collectd-config","rsync-status":"Applied"}]}]},{"name":"prometheus-operator","clusters":[{"cluster-provider":"provider1","cluster":"cluster1","resources":[{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"alertmanagers.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"podmonitors.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"prometheuses.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"prometheusrules.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"apps","Version":"v1","Kind":"Deployment"},"name":"r1-grafana","rsync-status":"Pending"},{"GVK":{"Group":"policy","Version":"v1beta1","Kind":"PodSecurityPolicy"},"name":"r1-grafana","rsync-status":"Pending"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1beta1","Kind":"Role"},"name":"r1-grafana","rsync-status":"Pending"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1beta1","Kind":"RoleBinding"},"name":"r1-grafana","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"Secret"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-grafana","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"ServiceAccount"},"name":"r1-grafana","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-grafana-clusterrole","rsync-status":"Applied"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-grafana-clusterrolebinding","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-grafana-config-dashboards","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-apiserver","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-coredns","rsync-status":"Pending"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-coredns","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"ConfigMap"},"name":"r1-prometheus-operator-grafana-datasource","rsync-status":"Applied"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-controller-manager","rsync-status":"Pending"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-controller-manager","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-etcd","rsync-status":"Pending"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-etcd","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-proxy","rsync-status":"Pending"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-proxy","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-kube-scheduler","rsync-status":"Pending"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kube-scheduler","rsync-status":"Pending"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-kubelet","rsync-status":"Pending"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-operator","rsync-status":"Pending"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-operator","rsync-status":"Pending"},{"GVK":{"Group":"apps","Version":"v1","Kind":"Deployment"},"name":"r1-prometheus-operator-operator","rsync-status":"Pending"},{"GVK":{"Group":"policy","Version":"v1beta1","Kind":"PodSecurityPolicy"},"name":"r1-prometheus-operator-operator","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-operator","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"ServiceAccount"},"name":"r1-prometheus-operator-operator","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-operator","rsync-status":"Pending"},{"GVK":{"Group":"batch","Version":"v1","Kind":"Job"},"name":"r1-prometheus-operator-operator-cleanup","rsync-status":"Pending"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-operator-psp","rsync-status":"Pending"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-operator-psp","rsync-status":"Pending"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Pending"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Pending"},{"GVK":{"Group":"policy","Version":"v1beta1","Kind":"PodSecurityPolicy"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Pending"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"Prometheus"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"Service"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"ServiceAccount"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"r1-prometheus-operator-prometheus","rsync-status":"Pending"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRole"},"name":"r1-prometheus-operator-prometheus-psp","rsync-status":"Pending"},{"GVK":{"Group":"rbac.authorization.k8s.io","Version":"v1","Kind":"ClusterRoleBinding"},"name":"r1-prometheus-operator-prometheus-psp","rsync-status":"Pending"},{"GVK":{"Group":"","Version":"v1","Kind":"Secret"},"name":"r1-prometheus-operator-prometheus-scrape-confg","rsync-status":"Applied"},{"GVK":{"Group":"monitoring.coreos.com","Version":"v1","Kind":"ServiceMonitor"},"name":"service-monitor-collectd","rsync-status":"Pending"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"servicemonitors.monitoring.coreos.com","rsync-status":"Applied"},{"GVK":{"Group":"apiextensions.k8s.io","Version":"v1beta1","Kind":"CustomResourceDefinition"},"name":"thanosrulers.monitoring.coreos.com","rsync-status":"Applied"}]}]}]}
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
./emco-openness-helm-install.sh -k /home/test/EMCO/deployments/kubernetes/config_north uninstall
Removing EMCO...
Deleting namespace emco
namespace "emco" deleted
Deleting emco persistent volumes
persistentvolume "emco-db-emco-etcd-data-0" deleted
persistentvolume "emco-db-emco-mongo-data" deleted
Done
```

### Pushing EMCO images to Harbor
It is often required to use developer tags for locally built EMCO images, push to Harbor and reference these custom images in helmcharts for developer testing.

To enable these features, run `export BUILD_CAUSE=DEV_TEST` prior to running `make deploy`. This will tag locally built images as `<username>-latest`, push to Harbor and reference these images in the generated helmcharts.

Developers can then use the ```emco-openness-helm-install.sh``` script from the above sections to install these custom images on their cluster.

