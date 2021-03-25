```text
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2021 Intel Corporation
```

# Sub controller support for dtc

## Background

The traffic controller provides network policy and service discovery action sub modules and are currently internal to the dtc microservice. These can be sub controllers to dtc and can be independent microservices. Architecturally, this would facilitate easy addition of new sub controllers through api registration.

## Grpc flow

The grpc calls from the orchestrator will be multiplexed and sub controllers will be called based on their priority. The sub controllers will run grpc server and the dtc will establish client connection to issue calls.

## Modularity and code reuse

The rpc, controller and rsyncclient packages will be reused from the existing code and modularized to store the data based on the tag and name. The dtc will add new controller registration API's to register the sub controllers.

## Supported API's

Sub controller registration API's POST, GET, PUT, DELETE will be added.

## Controller data store

The controller package client initialization code will be updated to take the name and tag.

    func NewControllerClient(name, tag string) *ControllerClient {
        return &ControllerClient{
           collectionName: name,
           tagMeta:        tag,
        }
    }
