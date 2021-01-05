#### SPDX-License-Identifier: Apache-2.0
#### Copyright (c) 2020 Intel Corporation

#################################################################
# Running EMCO testcases with emcoctl
#################################################################

This folder contains following test cases to run with EMCO. These tests assumes one edge cluster to run all test cases. EMCO needs to be installed on a cluster before running these tests. 

1. Prometheus and Collectd Helm charts
2. vFw
3. Collectd Helm chart and adding configmap during instantiation (using Generic Action Controller)
4. Http Server and Client

## Setup Test Environment to run test cases

1. export environment variables 1) KUBE_PATH where the kubeconfig for edge cluster is located and 2) HOST_IP: IP address of the cluster where EMCO is installed

#### NOTE: For HOST_IP, assuming here that nodeports are used to access all EMCO services both from outside and between the EMCO services.

2. Setup script

    Run Sets up script for creating artifacts needed to test EMCO on one cluster.

    `$ ./setup.sh create`

    Output of this command are 1) values.yaml file for the current environment 2) emco_cfg.yaml for the current environment and 3) Helm chart and profiles tar.gz files for all the usecases.

    Cleanup artifacts generated above with cleanup command

    `$ ./setup.sh cleanup`

## Create Prerequisites to run Tests
1. Apply prerequisites.yaml. This is required for all the tests. This creates controllers, one project, one cluster, default logical cloud. This step is required to be done only once for all usecases:

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f prerequisites.yaml -v values.yaml`

## Running test cases

1. Prometheus and Collectd usecase

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f test-prometheus-collectd.yaml -v values.yaml`

2. Generic action controller testcase

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f test-gac.yaml -v values.yaml`

3. Distributed Traffic Controller testcase

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f test-dtc.yaml -v values.yaml`
4. Firewall testcase

    `$ $bin/emcoctl --config emco-cfg.yaml apply -f test-vfw.yaml -v values.yaml`


## Cleanup

1. Delete Prometheus and Collectd usecase

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f test-prometheus-collectd.yaml -v values.yaml`

2. Delete Generic action controller testcase

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f test-gac.yaml -v values.yaml`

3. Delete Distributed Traffic Controller testcase

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f test-dtc.yaml -v values.yaml`

4. Firewall testcase

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f test-vfw.yaml -v values.yaml`

5. Cleanup prerequisites

    `$ $bin/emcoctl --config emco-cfg.yaml delete -f prerequisites.yaml -v values.yaml`

6. Cleanup generated files

    `$ ./setup.sh cleanup`

#### NOTE: Known issue with the test cases: Deletion of the resources fails sometimes as some resources can't be deleted before others are deleted. This can happen due to timing issue. In that case try deleting again and the deletion should succeed.
