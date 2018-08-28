```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2020 Intel Corporation
```
<!-- omit in toc -->
# Sample application to demonstrate network policy of traffic controller 
This document describes how to deploy an example application with network policy of traffic controller. The deployment consists of server and client pods, once deployed the client sends the request to the server every five seconds. The client pod logs the message "Hello from http-server" to indicate successful connectivity with the server. 

- Requirements
- Install EMCO and emcoctl
- Prepare the edge cluster
- Configure
- Install the client/server application
- Verify network policy resource instantiation
- Uninstall the client/server application
- Sample log from the client pod

## Requirements
- The edge cluster where the application is installed should support network policy
- Application pods should have the label app=&lt;app name&gt;

## Install EMCO and emcoctl
Install EMCO and emcoctl as described in the tutorial.

## Prepare the edge cluster
Install the Kubernetes edge cluster and make sure it supports network policy. Note down the kubeconfig for the edge cluster which is required later during configuration.

## Configure
(1) Copy the config file
```shell
$ cp src/tools/emcoctl/examples/emco-cfg-remote.yaml src/tools/emcoctl/examples/dtc/emco-cfg-dtc.yaml
```
(2) Modify src/tools/emcoctl/examples/dtc/emco-dtc-single-cluster.yaml and src/tools/emcoctl/examples/dtc/emco-cfg-dtc.yaml files to change host name, port number and kubeconfig path.

(3) Compress the profile and helm files

Create tar.gz of profiles
```shell
$ cd kud/tests/helm_charts/dtc/http-server-profile
$ tar -czvf ../../../../../src/tools/emcoctl/examples/dtc/http-server-profile.tar.gz .
$ cd ../http-client-profile
$ tar -czvf ../../../../../src/tools/emcoctl/examples/dtc/http-client-profile.tar.gz .
```
Create and copy .tgz of application helm charts
```shell
$ cd ../
$ tar -czvf http-server.tgz http-server/
$ tar -czvf http-client.tgz http-client/
$ cp *.tgz ../../../../src/tools/emcoctl/examples/dtc/
```

## Install the client/server app
Install the app using the commands:
```shell
$ cd ../../../../src/tools/emcoctl/examples/dtc/
$ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-single-cluster.yaml
```

## Verify network policy resource instantiation
```shell
$ kubectl get networkpolicy
  NAME               POD-SELECTOR      AGE
  testdtc-serverin   app=http-server   28s
```


## Uninstall the application
Uninstall the app using the commands:
```shell
$ emcoctl --config emco-cfg-dtc.yaml apply -f emco-dtc-terminate.yaml
$ emcoctl --config emco-cfg-dtc.yaml delete -f emco-dtc-single-cluster.yaml
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
