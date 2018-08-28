```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2020 Intel Corporation
```

# EMCO

## Overview

The Edge Multi-Cluster Orchestrator (EMCO) is a software framework for
intent-based deployment of cloud-native applications to a set of Kubernetes
clusters, spanning enterprise data centers, multiple cloud service providers
and numerous edge locations. It is architected to be flexible, modular and
highly scalable. It is aimed at various verticals, including telecommunication
service providers.

## Build and Deploy EMCO

### Set up the environment

Set up the following environment variables. Note that the value for the
container registry URL must end with a `/`.

```
export EMCODOCKERREPO=${container_registry_url}/
export HTTP_PROXY=${http_proxy}
export HTTPS_PROXY=${https_proxy}

```
### Update the base container images, if needed

The external dependencies for EMCO are captured partly in the environment
variables above and partly in a configuration file `config/config.txt`. 

The configuration file specifies two important parameters:
  * `BUILD_BASE_IMAGE`: The name and version of the base image used for
    building EMCO components themselves.
  * `SERVICE_BASE_IMAGE`: The name and version of the base image used to
    deploy the microservices that constitute EMCO.

By default, `config.txt` has the following content:
```
BUILD_BASE_IMAGE_NAME=emco-service-build-base
BUILD_BASE_IMAGE_VERSION=:1.1
SERVICE_BASE_IMAGE_NAME=alpine
SERVICE_BASE_IMAGE_VERSION=:3.12
```

By default, `emco-service-build-base` is built from `golang:1.14.1-alpine`, with the `make` utility added.

You may want to review and possibly update the base image names and versions.

Note: The build base image should be based on a Linux distribution that uses `apt` for package management, such as Alpine, Debian or Ubuntu. It should also provide Go language version 1.14.

### Populate the EMCODOCKERREPO registry

Populate the EMCODOCKERREPO registry with base images listed in `config/config.txt`, along with `mongodb` and `etcd` images.

The base images and versions that have been validated are as below:
  1.	Alpine:3.12 (for deploying EMCO components)
  2.	golang:1.14.1-alpine (for building EMCO components)
  3.	mongo:4.4.1
  4.	etcd:3

### Create the build base image in the EMCODOCKERREPO registry

EMCO does not assume that the base build image, such as `golang:1.14.1-alpine`, has the necessary utilities such as `make`.

Run the following to create the final build container image and populate that
in the `EMCODOCKERREPO` registry.

```
make build-base
```

### Deploy EMCO locally
You can build and deploy the EMCO components in your local environment (and
use them to deploy your workload in a set of remote Kubernetes clusters).

This is done in two stages:

 * Build the EMCO components: 
   ```make all```
   This spawns a build container that generates the needed EMCO binaries and
   container images.
 * Deploy EMCO components locally: 
   ```docker-compose up```
   using `deployments/docker/docker-compose.yml`. This spawns a set of
   containers, each running one EMCO component.

See [this tutorial](docs/user/Tutorial_Local_Install.md) for further details. 

### Deploy EMCO in a Kubernetes cluster
Alternatively, you can build EMCO locally and deploy EMCO components in a
Kubernetes cluster using Helm charts (and use them to deploy your workload in
another set of Kubernetes clusters).

Do the following steps:

 * Set up the environment:
   ```export BUILD_CAUSE=DEV_TEST```
   This triggers the following steps to push the locally generated EMCO images
   to the `EMCODOCKERREPO` container registry with appropriate tags.
 * Set up the Helm charts: Be sure to reference those image names and tags in
   your Helm charts.
 * Build and deploy EMCO: 
   ```make deploy```

See [this tutorial](docs/user/Tutorial_Helm.md) for further details. 
