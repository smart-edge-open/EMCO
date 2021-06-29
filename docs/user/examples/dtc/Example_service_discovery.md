```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2021 Intel Corporation
```
<!-- omit in toc -->
# Sample application to demonstrate service discovery feature 
This document describes how to deploy an example application with network policy of traffic controller. The deployment consists of server and client pods, once deployed the client sends the request to the server every five seconds. The client pod logs the message "Hello from http-server" to indicate successful connectivity with the server. 

- Requirements
- Install EMCO and emcoctl
- Prepare the edge cluster
- Configure
- Install the client/server application
- Verify network policy resource instantiation
- Verify service entry resource instantiation
- Uninstall the client/server application
- Sample log from the client pod

## Requirements
- The edge cluster where the application is installed should support network policy

## Install EMCO and emcoctl
Install EMCO and emcoctl as described in the tutorial.

## Prepare the edge cluster
Install the Kubernetes edge cluster and make sure it supports network policy. Note down the kubeconfig for the edge cluster which is required later during configuration.

## Testing scenarios

(a) communication between two private clusters (logical cloud level 0)

    (1) Copy the config file
    ```shell
    $ cp src/tools/emcoctl/examples/emco-cfg-remote.yaml src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l0_logical_cloud/emco-cfg-dtc.yaml
    ```
    (2) Modify src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l0_logical_cloud/emco-dtc-multiple-cluster-l0.yaml and src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l0_logical_cloud/emco-cfg-dtc.yaml files to change host name, port number and kubeconfig path.

    (3) Compress the profile and helm files

    Create tar.gz of profiles
    ```shell
    $ cd kud/tests/helm_charts/dtc/service_discovery_overrides/private_cluster/http-server-profile
    $ tar -czvf ../../../../../../../src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l0_logical_cloud/http-server-profile.tar.gz .
    $ cd ../http-client-profile
    $ tar -czvf ../../../../../../../src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l0_logical_cloud/http-client-profile.tar.gz .
    ```
    Create and copy .tgz of application helm charts
    ```shell
    $ cd ../../..
    $ tar -czvf http-server.tgz http-server/
    $ tar -czvf http-client.tgz http-client/
    $ cp *.tgz ../../../../src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l0_logical_cloud/
    ```

    ## Install the client/server app
    Install the app using the commands:
    ```shell
    $ cd ../../../../src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l0_logical_cloud/
    $ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-multiple-cluster-l0.yaml
    ```

    ## Verify network policy resource instantiation
    ```shell
    $ kubectl get networkpolicy
    NAME               POD-SELECTOR      AGE
    testdtc-serverin   app=http-server   28s
    ```
    ## Verify service entry created on the cluster where the client app is running
    ```shell
    $ kubectl get svc
    NAME                    TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
    service/http-service    ClusterIP   10.233.0.1   <none>        443/TCP   1d
    ```

    ## Uninstall the application
    Uninstall the app using the commands:
    ```shell
    $ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-terminate.yaml
    $ emcoctl --config emco-cfg-dtc.yaml delete -f emco-dtc-multiple-cluster-l0.yaml
    ```

    ## Sample log from the client pod

    ```shell
    $ kubectl logs pod/r1-http-client-54568d6c9-ftmr7
    get:
    2020-12-09 00:21:07 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:12 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:17 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:22 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd 
    ```

(b) communication between two private clusters (logical cloud level 1)

    (1) Copy the config file
    ```shell
    $ cp src/tools/emcoctl/examples/emco-cfg-remote.yaml src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l1_logical_cloud/emco-cfg-dtc.yaml
    ```
    (2) Modify src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l1_logical_cloud/emco-dtc-multiple-cluster-l1-step1.yaml, src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l1_logical_cloud/emco-dtc-multiple-cluster-l1-step2.yaml and src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l1_logical_cloud/emco-cfg-dtc.yaml files to change host name, port number and kubeconfig path.

    (3) Compress the profile and helm files

    Create tar.gz of profiles
    ```shell
    $ cd kud/tests/helm_charts/dtc/service_discovery_overrides/private_cluster/http-server-profile
    $ tar -czvf ../../../../../../../src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l1_logical_cloud/http-server-profile.tar.gz .
    $ cd ../http-client-profile
    $ tar -czvf ../../../../../../../src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l1_logical_cloud/http-client-profile.tar.gz .
    ```
    Create and copy .tgz of application helm charts
    ```shell
    $ cd ../../..
    $ tar -czvf http-server.tgz http-server/
    $ tar -czvf http-client.tgz http-client/
    $ cp *.tgz ../../../../src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l1_logical_cloud/
    ```

    ## Install the client/server app
    Install the app using the commands:
    ```shell
    $ cd ../../../../src/tools/emcoctl/examples/dtc/service_discovery/private_cluster/l1_logical_cloud/
    $ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-multiple-cluster-l1-step1.yaml
    $ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-multiple-cluster-l1-step2.yaml
    ```

    ## Verify network policy resource instantiation
    ```shell
    $ kubectl -n ns1 get networkpolicy
    NAME               POD-SELECTOR      AGE
    testdtc-serverin   app=http-server   28s
    ```
    ## Verify service entry created on the cluster where the client app is running
    ```shell
    $ kubectl -n ns1 get svc
    NAME                    TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
    service/http-service    ClusterIP   10.233.0.1   <none>        443/TCP   1d
    ```

    ## Uninstall the application
    Uninstall the app using the commands:
    ```shell
    $ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-terminate-l1.yaml
    $ emcoctl --config emco-cfg-dtc.yaml delete -f emco-dtc-multiple-cluster-l1-step1.yaml
    $ emcoctl --config emco-cfg-dtc.yaml delete -f emco-dtc-multiple-cluster-l1-step2.yaml
    $ emcoctl --config emco-cfg-dtc.yaml delete -f emco-dtc-multiple-cluster-l1-step1.yaml
    ```

    ## Sample log from the client pod

    ```shell
    $ kubectl -n ns1 logs pod/r1-http-client-54568d6c9-ftmr7
    get:
    2020-12-09 00:21:07 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:12 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:17 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:22 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd 
    ```
(c) communication between a private cluster (client app) and public cluster (server app) (logical cloud level 0)

    (1) Copy the config file
    ```shell
    $ cp src/tools/emcoctl/examples/emco-cfg-remote.yaml src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l0_logical_cloud/emco-cfg-dtc.yaml
    ```
    (2) Modify src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l0_logical_cloud/emco-dtc-multiple-cluster-l0.yaml and src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l0_logical_cloud/emco-cfg-dtc.yaml files to change host name, port number and kubeconfig path.

    (3) Compress the profile and helm files

    Create tar.gz of profiles
    ```shell
    $ cd kud/tests/helm_charts/dtc/service_discovery_overrides/public_cluster\(azure\)/http-server-profile
    $ tar -czvf ../../../../../../../src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l0_logical_cloud/http-server-profile.tar.gz .
    $ cd ../http-client-profile
    $ tar -czvf ../../../../../../../src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l0_logical_cloud/http-client-profile.tar.gz .
    ```
    Create and copy .tgz of application helm charts
    ```shell
    $ cd ../../..
    $ tar -czvf http-server.tgz http-server/
    $ tar -czvf http-client.tgz http-client/
    $ cp *.tgz ../../../../src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l0_logical_cloud/
    ```

    ## Install the client/server app
    Install the app using the commands:
    ```shell
    $ cd ../../../../src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l0_logical_cloud/
    $ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-multiple-cluster-l0.yaml
    ```

    ## Verify network policy resource instantiation
    ```shell
    $ kubectl get networkpolicy
    NAME               POD-SELECTOR      AGE
    testdtc-serverin   app=http-server   28s
    ```
    ## Verify service entry created on the cluster where the client app is running
    ```shell
    $ kubectl get svc
    NAME                    TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
    service/http-service    ClusterIP   10.233.0.1   <none>        443/TCP   1d
    ```

    ## Uninstall the application
    Uninstall the app using the commands:
    ```shell
    $ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-terminate.yaml
    $ emcoctl --config emco-cfg-dtc.yaml delete -f emco-dtc-multiple-cluster-l0.yaml
    ```

    ## Sample log from the client pod

    ```shell
    $ kubectl logs pod/r1-http-client-54568d6c9-ftmr7
    get:
    2020-12-09 00:21:07 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:12 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:17 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:22 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd 
    ```
(d) communication between a private cluster (client app) and public cluster (server app) (logical cloud level 1)

    (1) Copy the config file
    ```shell
    $ cp src/tools/emcoctl/examples/emco-cfg-remote.yaml src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l1_logical_cloud/emco-cfg-dtc.yaml
    ```
    (2) Modify src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l1_logical_cloud/emco-dtc-multiple-cluster-l1-step1.yaml, src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l1_logical_cloud/emco-dtc-multiple-cluster-l1-step2.yaml and src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l1_logical_cloud/emco-cfg-dtc.yaml files to change host name, port number and kubeconfig path.

    (3) Compress the profile and helm files

    Create tar.gz of profiles
    ```shell
    $ cd kud/tests/helm_charts/dtc/service_discovery_overrides/public_cluster\(azure\)/http-server-profile
    $ tar -czvf ../../../../../../../src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l1_logical_cloud/http-server-profile.tar.gz .
    $ cd ../http-client-profile
    $ tar -czvf ../../../../../../../src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l1_logical_cloud/http-client-profile.tar.gz .
    ```
    Create and copy .tgz of application helm charts
    ```shell
    $ cd ../../..
    $ tar -czvf http-server.tgz http-server/
    $ tar -czvf http-client.tgz http-client/
    $ cp *.tgz ../../../../src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l1_logical_cloud/
    ```

    ## Install the client/server app
    Install the app using the commands:
    ```shell
    $ cd ../../../../src/tools/emcoctl/examples/dtc/service_discovery/public_cluster\(azure\)/l1_logical_cloud/
    $ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-multiple-cluster-l1-step1.yaml
    $ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-multiple-cluster-l1-step2.yaml
    ```

    ## Verify network policy resource instantiation
    ```shell
    $ kubectl -n ns1 get networkpolicy
    NAME               POD-SELECTOR      AGE
    testdtc-serverin   app=http-server   28s
    ```
    ## Verify service entry created on the cluster where the client app is running
    ```shell
    $ kubectl -n ns1 get svc
    NAME                    TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
    service/http-service    ClusterIP   10.233.0.1   <none>        443/TCP   1d
    ```

    ## Uninstall the application
    Uninstall the app using the commands:
    ```shell
    $ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-terminate-l1.yaml
    $ emcoctl --config emco-cfg-dtc.yaml delete -f emco-dtc-multiple-cluster-l1-step1.yaml
    $ emcoctl --config emco-cfg-dtc.yaml delete -f emco-dtc-multiple-cluster-l1-step2.yaml
    $ emcoctl --config emco-cfg-dtc.yaml delete -f emco-dtc-multiple-cluster-l1-step1.yaml
    ```

    ## Sample log from the client pod

    ```shell
    $ kubectl -n ns1 logs pod/r1-http-client-54568d6c9-ftmr7
    get:
    2020-12-09 00:21:07 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:12 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:17 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd
    get:
    2020-12-09 00:21:22 Hello from http-server with the pod IP - 10.233.120.123 and podname - r1-http-server-7cf7db8d8-7bmsd 
    ```
