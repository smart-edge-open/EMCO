```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2020 Intel Corporation
```
# OVERVIEW

When applications are deployed to multiple clusters by EMCO, it is important to enable microservices within those applications to discover, name-resolve, and communicate with one another, across applications and clusters, subject to network policies.  The range of scenarios where discovery, name resolution and communication must be enabled includes:

 * Intra-cluster: microservices of different apps in the same cluster.
 * Inter-cluster: microservices of different apps in different clusters.
 * External clients: A client not managed by EMCO needs to be able to discover, name-resolve and communicate with an EMCO-managed microservice, which may be exposed as a NodePort service or LoadBalancer service in Kubernetes\*.

The focus of this document is on service discovery and communication in the Inter-cluster and External Client scenarios.

EMCO users declare the traffic communication policies using a single traffic intent, which encompasses network policies and QoS. The [traffic intent API](https://wiki.onap.org/display/DW/L7+Proxy+Service+Mesh+Controller+API%27s) has been documented elsewhere. The service discovery feature would have to take into account the traffic intent for that deployment and also the specific partition of apps onto clusters, as specified by the EMCO generic scheduler (based on Placement intents).

The service discovery feature is expected to insert virtual services and DNS entries in each cluster as needed. To do that, it must wait for the applications to be deployed, gather the network endpoints (IP addresses and ports) of relevant services, and deploy additional Kubernetes objects in each cluster as needed. The additional Kubernetes objects would depend on the specific solution adopted; for example, they may include virtual services and DNS records that map microservice/application names to network endpoints.

For the external client scenario, the EMCO-managed microservice needs to be exposed as a NodePort or Load Balance service. For a Load Balance service in a cloud (such as AWS\* or Azure\*), the configuration of the external DNS is typically handled by the cloud provider. At any rate, configuring external DNS is outside the scope of EMCO, at least for the 21.03 release.

# SOLUTION

## Overview

The solution involves a new action controller that handles all traffic intents. This traffic controller handles traffic communication, security, and potentially QoS in future releases. To start, it will only handle network policies and service discovery. Check the Release Notes to find out if new features have been added.

The traffic controller will have a modular construction with multiple sub-controllers, each of which handles one functional feature, such as network policies or service discovery. The controller shall invoke the sub-controllers in sequence -- there is no expectation of concurrency among the sub-controllers.

The intended sequence of operations for a new deployment is as follows:

 * After placement decisions are made, the orchestrator calls the action controllers in sequence, including the newly added traffic controller.
 * The traffic controller shall invoke the network policy sub-controller (NPS) and then the service discovery sub-controller (SDS).
* The SDS creates a nested appcontext within the deployment's appcontext to keep the status and data related to service discovery for that deployment.
* When both the sub-controllers are done, the traffic controller sends a response back to the orchestrator. However, the SDS continues to run asynchronously.
 * The orchestrator calls rsync, to deploy the application elements and also the CRs for the cluster monitor agents to act upon.
 * It is expected that the monitor agent in each cluster will return back the network endpoint data (IP addresses and ports) of every relevant microservice (of types NodePort, LoadBalancer, etc.)
 * The asynchronously running SDS monitors the deployment status in rsync. When the deployment is complete and the network endpoint data are available, the SDS creates Kubernetes objects related to virtual services and DNS entries in the nexted appcontext.
* The SDS then triggers rsync to push the objects in the nested appcontext to the clusters.

When a deployment is terminated, the rsync needs to delete the cluster resources associated with the nested appcontext before deleting the resources associated with the parent context.

Concurrent deployments would not cause additional complexities since:
  a. The traffic controller's sub-controllers would keep their data in distinct nodes under different deployment IDs.
  b. It is conceivable that the traffic controller serializes requests, i.e., runs each request to completion before taking the next request.

Further details, including timeouts etc., are addressed in later sections.

## Design

The following design is divided into three sub-categories.

1.  **Instantiate**:

  During the instantiation of an app, a gRPC call will be initiated from the Orchestrator to the network policy controller. The SDS will receive this request with the reference of intent name and parent app context ID. The SDS will read the network inbound client sets from the MongoDB\* and get the list of clients required for service to service communication. This info will be helpful in deploying the virtual service entries onto the clusters where the client will be deployed. SDS will create a new child app context and will be linked with the parent's app context.

For Example: The parent's appcontext looks something like this `/context/2572109775535238771/…` and the child's appcontext is `/context/1112111432272854425/…` .

In order to associate these, one idea is to use the metadata key already in the AppContext:
```
/context/2572109775535238771/meta/
Current Example contents:  {"Project":"testvfw","CompositeApp":"compositevfw","Version":"v1","Release":"fw0",
"DeploymentIntentGroup":"vfw_deployment_intent_group"}
```
 We could extend this to stash additional metadata, like a list of nested appcontexts or a primary app context.

Next, once the child's app-contexts are created, we will call rsync to deploy this child app-context and spawn a go-routine to poll for the service specs of the deployed service. The follwing details are polled continously every 5 seconds: loadbalancer/Node IP, node ports, and protocols. Once the following spec details are obtained, we call rsync to update the service entries, set the child's app-context as "instantiated" and finally exit the go routine. We assign a timeout of 120 seconds for the polling logic; if timeout expires then we set the status of the child's app context as failed/timedout and exit the go-routine.

Every time SDS polls for the service specs, it will monitor the status of the parent's app context. If the parent's app context is instantiating/instantiated then the polling will continue. In other cases, the polling will be terminated abruptly.

There will be certain corner cases to cover during termination, which will be discussed below.

2. **Terminate:**   

   When the user triggers terminate, the orchestrator begins initiating the uninstall of resources of the parent's app context. The orchestrator needs to check the state of the child's appcontext. If the child's app context is "instantiating/instantiated", then it calls rsync to terminate the child's appcontext resources. If the child's app context is failed, then there is no need to call rsync to uninstall for the child's app context because the polling routine would have already deleted the resources deployed on the clusters. After all resources in the child's app context have been terminated, it will terminate the parent's app context.

3. **Status:**   

  To get the status of each child app context, we populate the list of nested child app contexts to the deployment intent group status structure as shown below.
```
/deployment-intent-group/vfw_deployment_intent_group/status
		output:
			{
  "project": "testvfw",
  "composite-app-name": "compositevfw",
  "composite-app-version": "v1",
  "composite-profile-name": "vfw_composite-profile",
  "name": "vfw_deployment_intent_group",
  “nested” : [
        1112111432272854425
  ],
  "states": {
    "actions": [
      {
        "state": "Created",
        "instance": "",
        "time": "2020-10-14T23:13:40.932Z"
      },
      {
        "state": "Approved",
        "instance": "",
        "time": "2020-10-14T23:15:54.239Z"
      },
      {
        "state": "Instantiated",
        "instance": "2572109775535238771",
        "time": "2020-10-14T23:15:54.442Z"
      }
    ]
  },
  "status": "Instantiated"
     <etc, etc>
}
```

Now, query status of the nested appcontext:
`curl    <parent app context>/deployment-intent-group/vfw_deployment_intent_group/status?instance=c<child app context>`    

This will provide the status of the child app context.
