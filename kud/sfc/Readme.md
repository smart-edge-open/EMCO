#### SPDX-License-Identifier: Apache-2.0
#### Copyright (c) 2021 Intel Corporation

# Running EMCO testcases with emcoctl

This folder contains an example which uses EMCO to deploy
Service Function Chaining (SFC).  The SFC functionality is
provided by edge clusters which use the Nodus CNI (aka OVN4NFV)
https://github.com/akraino-edge-stack/icn-nodus

EMCO provides two action controllers

1. sfc controller - which takes SFC intents 
as part of an EMCO composite application to deploy the
service functions to create an SFC.

2. sfc client controller - takes SFC client intents to
as part of EMCO composite applications which will connect
to either end of the SFC.

This initial integration of EMCO and Nodus SFC is a technical
preview of this functionality.  Modifications of APIs and
behavior will occur as development of this work continues in
upcoming releases.

## Setup required

1. An Edge cluster which support Nodus SFC must be prepared.
   - As an example, the instructions found here https://github.com/akraino-edge-stack/icn-nodus/tree/master/demo/calico-nodus-secondary-sfc-setup
     were used to set up a working edge cluster.
   - Ensure the EMCO `monitor` is installed.
   - Apply the `NetworkAttachmentDefinition` used by EMCO to interact with Nodus in clusters (for applying `Networks`, `ProviderNetworks` and `NetworkChaining` CRs).

```
$ cat ovn-networkobj.yaml
apiVersion: k8s.cni.cncf.io/v1
kind: NetworkAttachmentDefinition
metadata:
  name: ovn-networkobj
  namespace: default
spec:
  config: '{ "cniVersion": "0.3.1", "name": "ovn4nfv-k8s-plugin", "type": "ovn4nfvk8s-cni"
    }'
```

## Description of SFC example using EMCO

This EMCO SFC example takes the same example described at
https://github.com/akraino-edge-stack/icn-nodus/tree/master/demo/calico-nodus-secondary-sfc-setup#demo-setup
and illustrates how SFC can be deployed and used with EMCO.

Note: that testing so far has just focused on the attachment and connection of the left and right nginx applications.
Provider networking attachments to the SFC when deployed with EMCO have not been tested yet.


### SFC Composite Application

An EMCO composite application is created for deploying the SFC.  It is comprised
of the example service functions (SLB, NGFW and SDEWAN) and an SFC intent which defines
the following items:

- the structure of the chain - i.e. the sequence of the service functions and connecting
  networks
- the namespace and pod label client selector information used to connect client pods to and end of
  the SFC
- the provider networks that attach to an end of the SFC

Note: the APIs for these SFC intents have been designed to support multiple pod and/or
provider network connections on an end of the SFC.  The initial implementation requires
one pod and one provider network be provided for each end and ignores any extra that may
be input via the API.

When this SFC composite application is deployed by EMCO, the SFC action controller will
create the appropriate Nodus networkchaining CR and deploy it to the edge cluster(s) along
with the service functions.

Note: The virtual networks that link together the functions in the SFC (see `virtual network1`,
`dync-net1`, `dync-net2` and `virtual network 2` in the demo picture at the link above) currently
must be created explicitly.  In this EMCO demo, there is a separate EMCO composite application
that has been created to deploy these networks.  It is expected that Nodus SFC functionality
will be enhanced in an upcoming release and eliminate the need for these virtual networks to be created by EMCO.

### SFC Client Composite Applications

Two EMCO composite applications have been created to attach to the ends of the SFC.

Each of these client composite applications have the following:

- use a  logical cloud a Kubernetes namespace to be used - each client uses a different logical cloud/namespace
- the client workload - which in this case - deploys a replicaset of pods
- an SFC client intent which identifies the SFC and which end to connect with

## Running the SFC example

All scripts and files for the SFC demo are located within this directory and sub-directories.

### Setup the demo

```
./setup.sh cleanup
./setup.sh create
../../bin/emcoctl/emcoctl --config emco-cfg.yaml apply -f prerequisites.yaml -v values.yaml
```

These steps clean up any previous demo executions and create the emco-cfg.yamld, values.yaml
and the helm charts used by the demo.

### Manual step (temporary workaround)

The Nodus SFC feature needs Kubernetes namespaces to be labeled (for the client pod matching).
Currently, EMCO logical clouds do not label the namespaces they create.

The workaround until this is addressed in the next EMCO release is to manually label
the namespaces created by the logical clouds in the `prerequistes.yaml` resource.

```
<on edge cluster> kubectl label ns sfc-tail sfc=tail
<on edge cluster> kubectl label ns sfc-head sfc=head
```

### Deploy the SFC

As noted above, currently the SFC networks need to be created.  In this step,apply the
SFC network and then the SFC composite app.

```
../../bin/emcoctl/emcoctl --config emco-cfg.yaml apply -f sfc-networks.yaml -v values.yaml
../../bin/emcoctl/emcoctl --config emco-cfg.yaml apply -f sfc-ca.yaml -v values.yaml
```

### Deploy the SFC Client applications

Deploy the client applications as follows:

```
../../bin/emcoctl/emcoctl --config emco-cfg.yaml apply -f sfc-left-client.yaml -v values.yaml
../../bin/emcoctl/emcoctl --config emco-cfg.yaml apply -f sfc-right-client.yaml -v values.yaml
```

*NOTE:* - during testing and development of SFC with EMCO, it has been found that Nodus SFC
appears to prefer to apply the SFC CR **last** - i.e. after the client applications.
While this issue is being investigated further, the current implementation of EMCO SFC
provides a 100 second delay before it applies the CR.
In other words, after deploying the SFC composite application, apply the two SFC client
applications within 100 seconds.

### Testing the SFC

After the SFC composite applications have been deployed as described above, the SFC can be tested by
running a traceroute command from a left side client pod to a right side client pod as follows:

```
kubectl -n sfc-head exec `kubectl get pods -lsfc=head -n sfc-head  --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | head -1` -- traceroute -n -q 1 -I 172.30.22.4

traceroute to 172.30.22.4 (172.30.22.4), 30 hops max, 46 byte packets
 1  172.30.11.3  2.298 ms
 2  172.30.33.3  1.433 ms
 3  172.30.44.2  0.669 ms
 4  172.30.22.4  0.731 ms

```

## How to clean up the edge cluster

Delete the emcoctl demo files in reverse order.  Each one typically needs to be deleted twice to fully clean up the resources.
Take your time, no rush (and probably do each step twice).

```
../../bin/emcoctl/emcoctl --config emco-cfg.yaml apply -f sfc-right-client.yaml -v values.yaml
../../bin/emcoctl/emcoctl --config emco-cfg.yaml apply -f sfc-left-client.yaml -v values.yaml
../../bin/emcoctl/emcoctl --config emco-cfg.yaml apply -f sfc-ca.yaml -v values.yaml
../../bin/emcoctl/emcoctl --config emco-cfg.yaml apply -f sfc-networks.yaml -v values.yaml
../../bin/emcoctl/emcoctl --config emco-cfg.yaml apply -f prerequisites.yaml -v values.yaml
```


## More advanced clean up

It may be helpful to clean up and reinstall the Nodus CNI.  The following steps may help accomplish this.

Note: in the examples below, the Nodus repository has been pulling into `/home/vagrant/git`.

```
export ovnport=`ip a | grep ': ovn4nfv0' | cut -d ':' -f 2`
kubectl -n kube-system exec $(kubectl get pods -lapp=ovn-controller -n kube-system  --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') -it -- ovs-vsctl del-port br-int $ovnport
kubectl -n kube-system exec $(kubectl get pods -lapp=ovn-control-plane -n kube-system  --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') -it -- ovn-nbctl lsp-del $ovnport

kubectl delete -f /home/vagrant/git/icn-nodus/deploy/ovn4nfv-k8s-plugin.yaml
# wait for above to completely terminate
kubectl delete -f /home/vagrant/git/icn-nodus/deploy/ovn-daemonset.yaml
# wait for above to completely terminate

# after above two are completely terminated, re-apply them
kubectl apply -f /home/vagrant/git/icn-nodus/deploy/ovn-daemonset.yaml
# wait for ovn to come up
kubectl apply -f /home/vagrant/git/icn-nodus/deploy/ovn4nfv-k8s-plugin.yaml
```

## Debugging

The `collect-info.sh` script can be used to collect logs and information about the SFC demo.

The SFC demo can also be deployed manually (not using EMCO) on the target edge cluster using the instructions
here https://github.com/akraino-edge-stack/icn-nodus/tree/master/demo/calico-nodus-secondary-sfc-setup
if needed to verify the edge cluster has been set up correctly.


