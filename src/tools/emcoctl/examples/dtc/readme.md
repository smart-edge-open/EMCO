# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

### Steps to run emcoctl to test network policy and service discovery feature

(1) Modify the ../emco-cfg-remote.yaml to assign the correct node IPs and nodeports of each microservices

(2) Modify the emco-dtc-np-single-cluster.yaml to assign the correct kubeconfig path of the clusters and assign the correct rsync node port

(3) Compress the sample helm charts provided in the folder "kud/tests/helm_charts_dtc" in .tgz format and also compress the http-server-profile and http-client-profile to tar.gz 

(4) Run emcoctl to deploy the http-server and http-client apps 

emcoctl --config ../emco-cfg-remote.yaml apply -f emco-dtc-single-cluster.yaml

### Steps to terminate

(1) Terminate the app installed on the clusters.
emcoctl --config ../emco-cfg-remote.yaml apply -f emco-dtc-terminate.yaml

(2) Delete the mongodb and etcd resources
emcoctl --config ../emco-cfg-remote.yaml delete -f emco-dtc-single-cluster.yaml






