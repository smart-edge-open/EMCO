```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2021 Intel Corporation
```

# Subcontroller support for dtc

## Background

The traffic controller provides network policy and service discovery action submodules and are currently internal to the dtc microservice. These can be subcontrollers to dtc and can be independent microservices. Architecturally, this would facilitate easy addition of new sub controllers through api registration.

## gRPC flow

The gRPC calls from the orchestrator will be multiplexed and subcontrollers will be called based on their priority. The subcontrollers will run gRPC server and the dtc will establish client connection to issue calls.

## Modularity and code reuse

The `rpc`, `controller` and `rsyncclient` packages will be reused from the existing code and modularized to store the data based on the tag and name. The dtc will add new controller registration API's to register the sub controllers.

## Supported API's

Subcontroller registration API's POST, GET, PUT, DELETE will be added.

## Controller data store

The controller package client initialization code will be updated to take the name and tag.

    func NewControllerClient(name, tag string) *ControllerClient {
        return &ControllerClient{
           collectionName: name,
           tagMeta:        tag,
        }
    }
