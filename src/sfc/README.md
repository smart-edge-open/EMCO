```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2021 Intel Corporation
```

# Overview of Service Function Chaining (SFC) - aka Network Chains

EMCO provides two action controllers for handling the input and application of SFC  intents.

- The `sfc` action controller handles the input of intents which will be used to generate an SFC CR which will be applied to clusters along with the applications that comprise the functions in the SFC.
- `sfcclient` action controller handles the intents which will be used to attach client applications to the left or right ends of the SFC.

At this time, SFC is supported on clusters where the `ovn4nfv` CNI is deployed.

See:  https://github.com/akraino-edge-stack/icn-ovn4nfv-k8s-network-controller

# SFC Controller

The `sfc` action controller is responsible for taking in via API the intents for network chains.

The `sfc` action controller intents are added to a deployment intent group which is created to deploy `Network Chains` to a set of
clusters.  These intents will be used to create `NetworkChaining` CR.  The left and right ends of a network chain are connected to one or more `Provider Networks` and/or `Pods`.

`Provider Networks` may be deployed to cluster by EMCO using `ncm`.

Other `Deployment Intent Groups` which include `sfclclient` action controller intents that refer to a `Network Chain` may specify that `Pods` in specified applicatoin resources  be *connected* to an end of a `Network Chain` via label matching.

## Network Chain:

URL:

```
/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/network-chains
```

Body:

```
{
  "metadata": {
    "name": "sfc1",
    "description": "virtual/provider network to virtual/provider network",
    "userData1": "data 1",
    "userData2": "data 2"
  },
  "spec": {
    "chainType": "Routing",
    "namespace": "default",
    "networkChain": "net=vnet1,app=a10,net=dync-net1,app=vfw,net=dync-net2,app=sdewan,net=vnet2"
  }
}
```

### Routing Spec Information:

#### Network Chain Client Selectors

Defines the labels and namespaces that will match with pod template specs to attach the pod to the specified end of the SFC.

URL:

```
/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/network-chains/{sfc}/client-selectors
```

Body:

```
{
  "metadata": {
    "name": "client1",
    "description": "pod client of an SFC",
    "userData1": "data 1",
    "userData2": "data 2"
  },
  "spec": {
    "chainEnd": "< left | right >",
    "podSelector": {
      "matchLabels": {
	"app": "vbng"
      }
    },
    "namespaceSelector": {
      "matchLabels": {
	"app": "vbng"
      }
    }
  }
}
```

#### Network Chain Provider Networks

Defines the provider networks that may be attached to a specified end of the SFC.

URL:

```
/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/network-chains/{sfc}/provider-networks

```

Body:

```
{
  "metadata": {
    "name": "chain-providernetwork",
    "description": "chain-providernetwork",
    "userData1": "data 1",
    "userData2": "data 2"
  },
  "spec": {
    "chainEnd": "< left | right >",
    "networkName": "provider-network-1",
    "gatewayIp": "172.30.10.3",
 	    "networkRepresentor": {
      "gatewayip": "<ipaddress>",
      "subnet": "<ipaddress>"
    }
  }
}
```

# Network Chain Client Controller

The `sfcclient` controller is responsible for taking in via API the intents that will match up applications to specified endpoints of a Network Chain.

When this intent is included in the set of intents of a `Deployment Intent Group` and the `sfcclient` action is executed, the identified chain will be queried and the `podSelector` and `namespaceSelector` match labels of the specified end of the chain (`left` or `right`) will be added to the pod template of the specified application workload resource.

## Network Chain Client Intents

URL:

```
/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/sfc-clients
```

Body:

```
{
  "metadata": {
    "name": "chain1 client1",
    "description": "chain1 client1 information",
    "userData1": "blah blah",
    "userData2": "abc xyz"
  },
  "spec": {
    "appName": "app1",
    "workloadResource": "app1Deployment",
    "resourceType": "deployment"
    "chainEnd": "<left | right>",
    "chainName": "chain1",
    "chainCompositeApp": "chain1-CA",
    "chainCompositeAppVersion": "chain1-CA-version",
    "chainDeploymentIntentGroup": "chain1-deployment-intent-group",
    "chainNetworkControlIntent": "chain1-net-control-intent"
  }
}
```
