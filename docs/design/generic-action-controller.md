```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2020 Intel Corporation
```
# GENERIC ACTION CONTROLLER

The generic action controller microservice is an action controller which is  registered with the central orchestrator. It manages the following usecases:

- <b>Create a new Kubernetes\* object</b> and deploy it along with a specific application which is part of the composite application. There are two variations here:

  - Default: Apply the new object to every instance of the app in every cluster where the app is deployed.
  - Cluster-Specific: Apply the new object only where the app is deployed to a specific cluster, denoted by a cluster-name or a list of clusters denoted by a cluster-label

- <b>Modify an existing Kubernetes object</b> which may have been deployed using the helm chart for an app or may have been newly created by the above mentioned usecase. Modification may correspond to specific fields in the YAML definition of the object.

To acheive both the usecases, the controller exposes REST APIs to create, update and delete the following :

- Resource - Specifies the newly defined object or an existing object.
- Customization - Specifies the modifications(using JSON Patching) to be applied on the objects.

* The <b>outline of this document</b> :

Internally the doc is divided into 2 sections:

  - <b>Examples</b> - showing how to register the controller, and shows configuration examples of 3 supported usecases.
    - Creation of a new k8s resource.
    - Creation of configMaps and secrets using data from multiple external files.
    - Modifying an existing resource.

  - <b>REST API definition</b> - The detailed definition of REST APIs exposed by the controllers and the options supported by each API.


### Controller registration with the orchestrator:

```
#creating controller entries
version: emco/v2
resourceContext:
  anchor: controllers
metadata :
   name: gac
spec:
  host:  {{.HostIP}}
  port: {{.GacPort}}
  type: "action"
  priority: 1
```

### Creating a new Resource

After registration of the controller, we should create Generic-k8s-intents.

* ##### Add the GAC intent
```
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-k8s-intents
metadata:
  name: {{.GacIntent}}
```

For example, Create a new resource which a new <b>NetworkPolicy</b>
Once the intent is registered, we could create the new resource in the following way:


* ##### Add resources to GAC intent
```
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-k8s-intents/{{.GacIntent}}/resources
metadata:
  name: resourceNetworkPolicyApp
spec:
  appName: {{.App}}
  newObject: "true"
  resourceGVK:
    apiVersion: networking.k8s.io/v1
    kind: NetworkPolicy
    name: MyNetworkPolicy
file:
  {{.NetworkPolicyYAML}} # This is the raw manifest YAML definition of networkPolicy.
```

Example of network policy YAML definition
```
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: MyNetworkPolicy
  namespace: default
spec:
  podSelector: {}
  ingress:
    - from:
      - namespaceSelector:
          matchLabels:
            name: backend
```

Next thing might be, you may want to go for a cluster specific customization.
Lets say you want the network policy to be deployed on clusters which have
label as "label_a".

* ##### Add customization to resource
```
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-k8s-intents/{{.GacIntent}}/resources/collectd-resources/customizations
metadata:
  name: collectd-customizations
spec:
  clusterSpecific: "true"
  clusterInfo:
    scope: label
    clusterProvider: {{.ClusterProvider}}
    clusterName: ""
    clusterLabel: "label_a"
    mode: allow
```

### Creating ConfigMaps and Secrets

GAC supports the creation of ConfigMaps and secrets using external files.
For example, we want to deploy a cluster specific configmap and the config
data for the configMap comes from an external JSON file like, sensor.json.
We can also have multiple external files as config Data.

Similar to configMaps, you can also create cluster specific secrets with
external files.
Both the flow for configMap and Secret creation remain identical.
An example of configmap creation is shown below :


* ##### Add the GAC intent
```
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-k8s-intents
metadata:
  name: {{.GacIntent}}
```

* ##### Add resources to GAC intent
```
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-k8s-intents/{{.GacIntent}}/resources
metadata:
  name: collectd-resources
spec:
  appName: {{.App}}
  newObject: "true"
  resourceGVK:
    apiVersion: v1
    kind: configMap
    name: sensor-info-script
file:
  {{.HelmApp}} # File upload not required in case of configMap and secret as part of resource creation
```

* ##### Customize using the data file
```
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-k8s-intents/{{.GacIntent}}/resources/collectd-resources/customizations
metadata:
  name: collectd-customizations
spec:
  clusterSpecific: "true"
  clusterInfo:
    scope: label
    clusterProvider: {{.ClusterProvider}}
    clusterName: ""
    clusterLabel: {{.ClusterLabel}}
    mode: allow
files:
  - {{.ConfigmapFile}} # this might be data file like sensor.json which is the
  data to be loaded into configMap.
```

### Modifying an existing resource

GAC supports modifying an existing resource. For example, you have an etcd-cluster in your composite app which is currently configured for
replica count 3, and you want the replica count to be 6.
This can be done by the following steps:

* ##### Add the GAC intent
```
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-k8s-intents
metadata:
  name: {{.GacIntent}}
```


* ##### Add resources to GAC intent
```
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-k8s-intents/{{.GacIntent}}/resources
metadata:
  name: resourceETCD
spec:
  appName: {{.App}}
  newObject: "true"
  resourceGVK:
    apiVersion: "apps/v1"
    kind: "etcd"
    name: "MyEtcdName"
```

* ##### Add customization to increase the replica count
```
version: emco/v2
resourceContext:
  anchor: projects/{{.ProjectName}}/composite-apps/{{.CompositeApp}}/v1/deployment-intent-groups/{{.DeploymentIntent}}/generic-k8s-intents/{{.GacIntent}}/resources/collectd-resources/customizations
metadata:
  name: customizeETCD
spec:
  clusterSpecific: "true"
  clusterInfo:
    scope: label
    clusterProvider: {{.ClusterProvider}}
    clusterName: ""
    clusterLabel: {{.ClusterLabel}}
    mode: allow
  patchType: "json",
    patchJson: [
      {
        "op": "replace",
        "path": "/spec/replicas",
        "value": 6
      }
    ]
```

### REST APIs

 - The proposed API is structured in two parts: one defines the base resource
object of interest(which shall be called 'resource' in the APIs), and the other defines customizations to be applied on the base resource.
- The base resource may be a newly defined object (the first use
case above) or an existing object defined in the Helm chart of some app (the second use case).
- The customization specifies the JSON patches to be applied
on the base resource and, optionally, the cluster(s) to which the app must be mapped.
- If no clusters are specified, the customization is applied
independent of which cluster(s) the app is mapped to.

- When customizations have been defined for a resource, only the customized variants shall be applied.
- The base resource is applied only when no
customizations exist.

## Definition of a Generic K8s Intent

The following APIs allow for the creation and retrieval of Generic K8s Intents.

- Meta data body
```
{
   "metadata":{
      "name":"${generick8s_intent_name}",
      "description":"descriptionf of ${generick8s_intent_name}",
      "userData1":"user data 1 for ${generick8s_intent_name}",
      "userData2":"user data 2 for ${generick8s_intent_name}"
   }
}
```
- Operations supported
```
CREATE - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents
```
```
GET ALL - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents
```
```
GET SPECIFIC - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{name}
```
```
UPDATE - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{name}
```
```
DELETE - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{name}
```

## Definition of a Resource

The following APIs allow for the creation and retrieval of Resource.

- Metadata Body of resource
```
{
  "metadata":{
    "name": "${resource_name}",
    "description": "description for ${resource_name}",
    "userData1": "user data 1 for ${resource_name}",
    "userData2": "user data 2 for ${resource_name}"
  },
  "spec":{
     "appName": "${appName}",
     "newObject": "false",
     "resourceGVK":{
       "apiVersion": "apps/v1",
       "kind": "${kind}",
       "name": "${k8sResourceName}"
      }
   }
}
```



The fields in the `spec` object in the POST request body are as follows:

 * `appName`: Name of the application of interest.
 * `newObject`: Flag indicating whether this intent defines a new object.
   If `true`, the `file` must be uploaded along with the API creation[unless it is configMap or secret]; The `file` shall contain the resource definition in YAML format. In case of configMap/Secret, the `file` neeed not be present. The data file in case of configMap/secret shall be uploaded through customization.

 * `resourceGVK`: A reference to the object being created or existing.
   (`apiVersion` and `kind` in K8s terms) and a name (`metadata/name` field in K8s).

Only one object of a given `apiVersion`, `kind` and `name` can exist as a
base resource without customizations.

The following operations are supported on individual base resources:

```
CREATE - projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{intent-name}/resources
```
```
GET ALL - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{intent-name}/resources
```
```
GET SPECIFIC - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{intent-name}/resources/{name}
```
```
PUT - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{intent-name}/resources/{name}
```
```
DELETE - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{intent-name}/resources/{name}
```

## Definition of Customization

Customization allows us to customize an existing resource or a newly created resource.
Let's say you want a cluster specific deployment, then this can help.
Customization also supports uploading of multiple data `files`
which can be used for the `data` field in case of configMaps and
secrets.
It can also help modifying an existing k8s resource through JSON
patching.

- Metadata Body of Customization
```
{
  "metadata": {
    "name": "${customization_name}",
   "description": "description for ${customization_name}",
    "userData1": "user data 1 for ${customization_name}",
    "userData2": "user data 2 for ${customization_name}"
  },
  "spec": {
    "clusterSpecific": "true",
    "clusterInfo": {
      "scope": "label",
      "clusterProvider": "${clusterprovidername}",
      "clusterName": "",
      "clusterLabel": "${labelname}",
      "mode": "allow"
    },
    "patchType": "json",
    "patchJson": [
      {
        "op": "replace",
        "path": "/spec/replicas",
        "value": 1
      }
    ]
  }
}
```

- Operations supported
```
CREATE - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{intent-name}/resources/{resource-name}/customizations
```
```
GET ALL - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{intent-name}/resources/{resource-name}/customizations
```
```
GET SPECIFIC - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{intent-name}/resources/{resource-name}/customizations/{name}
```
```
UPDATE - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{intent-name}/resources/{resource-name}/customizations/{name}
```
```
DELETE - /projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-k8s-intents/{intent-name}/resources/{resource-name}/customizations/{name}
```


The fields in the `spec` object in the POST request body are as follows:

 * `clusterSpecific`: A flag that the variant is meant only for a single (set
   of) clusters. If `true`, the `clusterInfo` field must be present.
 * `clusterInfo`: The set of clusters to which this customization applies.
   * `scope`: Set to `label` if a cluster label is provided, `name`
     otherwise.
   * `mode`: If set to `allow`, the customizations are applied only to the
     specified set of clusters. If set to `deny`, the customizations are
     applied to all relevant clusters except those specified.
   * `clusterProvider`: Name of the service provider hosting the cluster.
   * `clusterName`: Cluster name. Required and relevant only if `scope` is `name`.
   * `clusterLabel`: A label set on the cluster. Required and relevant only if `scope` is `label`.
    * `patchType`: Specifies the type of patch. Required and relevant only if you want to modify an existing resource.
    * `patchJson`: This consists of :
      *  `op` : stands for the operation to be performed as patch. For example, "replace". For other operations in case of JSON patching , refer : https://github.com/evanphx/json-patch
      * `path` : denotes the path of the item in the YAML definition of the resource. For eg, you can find the 'replica' count at `/spec/replicas`
      * `value` : modified value of the item.


> **NOTE**: Updating or deleting a customization when the customized resource has been deployed will affect only the database entry but not the actual deployments in any cluster.
