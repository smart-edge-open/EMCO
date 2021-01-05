```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2021 Intel Corporation
```
# Release Notes
This document provides high level features, fixes, and known issues and limitations for the Edge Multi-Cluster Orchestration (EMCO) project.

# Release History

1. EMCO - 21.03
1. EMCO - 20.12

# Features for Release

1. **EMCO - 21.03**
	- Support for Helm v3 charts in composite applications.
	- Service Discovery for Deployment Intent Groups. See [Service Discovery Design](docs/developer/service-discovery-design.md).
	- `Put` support added to the `emcoctl` tool.
	- Simple EMCO deployment Helm charts have been replaced with fuller function Helm charts with sub-charts per EMCO microservice.
	- The Cluster Manager ( `clm` ) has been extended to support the invocation of registered plugin controllers when clusters are created or deleted.
	- Ability for `rsync` microservice to read (get) Kubernetes resources has been added.
	- Additional query parameters added to the Deployment Intent Group status query to allow for querying the list of apps, the clusters by app and resources by app.  See the status query section of [Resource Lifecycle](docs/user/Resource_Lifecycle.md).


1. **EMCO - 20.12**
	- This is the first release of the Edge Multi-Cluster Orchestration (EMCO) project.  EMCO supports the automated deployment of  geo-distributed applications to multiple clusters via an intent driven API.
	-   EMCO is composed of a number of central microservices:
		-   Cluster Manager (clm) : onboard clusters into EMCO
		-   Network Configuration Manager (ncm) : define and apply provider and virtual network intents to clusters which required additional network interfaces for workloads, such as Virtual Network Functions.  Support for OVN4NVF networks is present.
		-   Distributed Cloud Manager (dcm) : define and instantiate logical clouds which provide a common namespace across a set of clusters to which applications may be deployed
		-   Distributed Application Scheduler (orchestrator) : supports creation of composite applications via onboarding of Helm charts and customization and automation of deployment via support for placement and action controllers.
		-   OVN Action Controller (ovnaction) : action controller which supports creation of network interface intents which automates the addition of OVN4NFV network interfaces connected to provider or vitual networks  to specified applications during deployment.
		-   Traffic Controller (dtc) : action controller which supports creation of network policy intents which will deploy network policy resources to the specified clusters of the application.
		-   Generic Action Controller (gac) : action controller which supports creation of intents which allow for the creation of additional Kubernetes resources for some or all of the clusters where an application is deployed.  Also it supports intents to modify Kubernetes objects which are already part of the application.
		-   Resource Synchronizer (rsync) : handles instantiation, termination and status collection of the resources prepared by the other EMCO microservices to the remote clusters.

	-   EMCO provides a microservice for the remote clusters:
		-   Monitor (monitor) : collects and aggregates status of supported Kubernetes resources that have been deployed by EMCO.  EMCO rsync watches for updates and collects the status information.

	-   EMCO provides a CLI tool (emcoctl) which may be used to interact with the EMCO REST APIs.
	-   Authorization and Authentication may be provided for EMCO by utilizing Istio. See [Emco Integrity Access Management](docs/user/Emco_Integrity_Access_Management.md) for more details.

# Fixes for Release

1. **EMCO - 21.03**

	- Emcoctl get with token has been fixed.
	- Fixes in many microservices to align the data, REST API return codes with the EMCO OpenAPI documentation.
	- REST PUT support added for many of the EMCO APIs.
	- Additional unit test coverage in many packages has been added.
	- Format of the cluster network intent status query response has been simplified to remove inapplicable and redundant `apps` and `clusters` lists.

# Known Issues and Limitations

- **EMCO 21.03**
	- If the `monitor` pod is restarted on an edge cluster, the `rsync` connection will fail because it continues to listen on the previous (now removed) connection.
	- Username / password authentication is enabled by default for EMCO mongo and etcd services.  If persistence is also enabled, then the same passwords should be used across install cycles.
          Installation via the `emco-openness-helm-install.sh` script disables persistence by default.  Installation using the default Helm charts and values ( `deployments/helm/emcoOpenNESS` ) has persistence enabled by default.
		- Refer to [Helm Tutorial](docs/user/Tutorial_Helm.md) and the [Helm Chart README](deployments/helm/emcoOpenNESS/README.md) for more information.
	- REST PUT (update) is not yet supported for `Cluster` resources  and `Deployment Intent Group` resources or sub-resources (i.e. intents) managed by the `orchestrator` microservice.
	- A REST GET of a composite application app or app profile without specifying an appropriate Accept header causes the `orchestrator` microservice to panic.
	- REST GETs of various intent resources of the Traffic Controller microservice `dtc` return incorrect HTTP return codes (something other than 404) when the parent resources in the URI do not exist.

- **EMCO 20.12**
	- EMCO provides a simple Helm chart to deploy EMCO microservices under `deployments/helm/emcoCI`.   This Helm chart supports limited scoped user authentication to the EMCO Mongo and etcd databases.  The comprehensive Helm charts under `deployments/helm/emcoOpenNESS` are still a work in progress and will include the authentication and full integration with EMCO microservices in a future release.
	- Many of the EMCO microservice REST APIs do not support the PUT API for providing modifications to resources after initial creation.
	- The `emcoctl` command line tool does not support a `put` operation at all.
	- In some cases, EMCO does not prevent deletion of API resources which are depended on by other resources.  For example, a Cluster resource might be deleted while a Deployment Intent Group is instantiated and has resources deployed to the Cluster.  Until this issue is addressed in the next release, the best method is to ensure that resources are deleted in the reverse order from their creation.
	- EMCO does not provide for encryption-at-rest for the database storage of the Mongo and etcd databases. EMCO plans to provide support for encryption of critical database resources in an upcoming release. 
	- The example virtual firewall composite application needs to be deployed to a Kubernetes cluster which has Multus, OVN4K8S CNI and virtlet support installed.  Refer to [KUD](https://github.com/onap/multicloud-k8s/tree/master/kud) for an example cluster that which supports the requirement needed by the virtual firewall example.
	- The monitor microservice is only able to monitor the status of a limited set of Kubernetes resource Kinds:  pod, service, configmap, deployment, secret, deamonset, ingress, jobs, statefulset, csrstatus
	- Emcoctl get with token doesn't work. That is because of a bug in the code. Solution to the issue is to remove line 25 from the EMCO/src/emcoctl/cmd/get.go and rebuild emcoctl code.

# Software Compatibility

- **EMCO 21.03**
	- EMCO has been tested with Kubernetes v1.16.8, v1.18.9, v1.19 and v1.20.0

- **EMCO 20.12**
	- EMCO has been tested with Kubernetes v1.18.9 and v1.19.
