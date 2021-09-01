```
Copyright (c) 2021 Intel Corporation
```

# PROBLEM STATEMENT

 1. **Problem**: EMCO runs Placement controllers in sequence, daisy-chaining
    the output of one into the input of another. This is not ideal for several
    reasons:

    * The admin needs to spell out the order in which the controllers need to
      be run.

    * The set of solutions chosen by the first controller in the chain may
      severely constrain the range of solutions available to the next
      controller, resulting in a non-  optimal or even non-existent solution.

    * The scheduler, which calls the placement controllers, can only place a
      combined large timeout on the daisy chain.

    **Solution**: Adopt a concurrent model, where the EMCO scheduler calls
    all known Placement controllers concurrently. Each controller must return a
    list of partition candidates (i.e., potential mappings from apps to
    clusters (or apps to apps) that satisfy the specific Placement intent which
    that controller is meant to handle). The EMCO scheduler must then combine
    the lists of partition candidates and send a subset of them for deployment.
    However, this solution can lead to Problem 2.

    Note: The concurrency is not an important deliverable for the first release.

 1. **Problem**: Since the number of clusters could be large, the list of
    partition candidates returned by each could get potentially large. We
    need a compact representation of a large set of partition candidates.

    **Solution**: See Section [Representation of Partition Candidates](#Representation-of-Partition-Candidates)

 1. **Problem**: EMCO needs a Hardware Platform Awareness (HPA)-based
    Placement Intent. To implement this intent, it may take a Placement
    Controller and an Action Controller. It should be possible for the app
    developer to state that a certain microservice in a specific app needs a
    specific list of resources, and the implementation must pass that
    requirement to each appropriate K8s cluster so that the K8s scheduler can
    place that microservice on a node that has that specific list of
    resources.

    **Solution**: See Section [HPA Intent Overview](#HPA-Intent-Overview).

# SOLUTION: ARCHITECTURE AND API

## Representation of Partition Candidates

Each Placement controller must return a list of partition candidates that
satisfy that intent type. The scheduler would then combine them to produce a
single list of partition candidates that is passed to the Deployment service.
To handle clusters at scale, the list of partition candidates must be
represented efficiently. To do that, we note that there are two kinds of
intent types:

 * Intent types that specify what type of cluster a given app can map to,
  based on the cluster's resources, location, etc. E.g. HPA intent. For this
  type, potentially a large number of clusters may match a given app. To
  handle scaling, the list of clusters for a app is best expressed as a
  allow/deny list of the form:

  ```
  { app-name-1: replication-count: allow | deny cluster-list-1,
    app-name-2: replication-count: allow | deny cluster-list-2,
    �
  }

  ```

  For each app, this structure specifies not only the list of clusters it is
  compatible with but also the replication count: how many clusters that app
  should be deployed on. The intent handler may choose either the allow form
  or deny form for each app, based on whatever is most compact. When two apps
  cannot be placed together due to resource contention, the handler must
  ensure that their respective cluster lists are disjoint.

 * Intent types that specify which set of apps can be placed together.
   E.g. affinity or anti-affinity rules. For this type, each partition
   candidate is best expressed as a allow/deny list of compatible apps:

   ```
   { app-name-1: allow | deny app-list-1,
     app-name-2: allow | deny app-list-2,
     �
   }

   ```

   This form will not be supported in the first implementation; it is a
   placeholder for the future.

## HPA Intent Overview

The HPA-based Intent has two parts: determining a suitable cluster based on
the hardware requirements for each microservice and modifying the Kubernetes
objects corresponding to the app or microservice, so that the Kubernetes
controller in the target cluster can satisfy those requirements. The former
is handled by a placement controller and the latter by an action controller.

The HPA Intent and its controllers are aimed to work with Kubernetes clusters.
Kubernetes has a specific way of expressing hardware requirements and Quality
of Service for pods and objects containing pods such as Deployments.

The HPA Intent will use the same syntax and semantics as Kubernetes for
expressing [resource requirements](https://
kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/). In particular:

 * There are two kinds of resources:

   * Allocatable: resources which can be quantified and allocated to
     containers in specific quantities, such as `cpu` and `memory`. These
     need to be assigned to specific containers within a pod.
   * Non-allocatable: resources which are properties of CPUs, hosts, etc.
     such as the availablity of a specifc instruction set like AVX512).
     These are assigned to a pod, rather than a container. However, for the
     sake of uniformity, the consumer is expressed as a container, if both
     allocatable and non-allocatable resources are assigned to a consumer.

 * Each resource requirement in the intent shall be stated using the same name
   as in Kubernetes, such as `cpu`, `memory`, and `intel.com/gpu`, for both
   allocatable and non-allocatable resources.

 * For allocatable resources, the Kubernetes notion of `requests` (minimum
   needed amount of a resource) and `limits` (maximum amount of that resource
   that can be allocated) shall be followed.

   * Note that the `requests` field is of relevance to the HPA Placement
     Controller, because it needs to find a cluster where at least one node has
     that resource in that quantity.
   * In particular, if the HPA intent specifies a `limits` without a `requests`
     field, [Kubernetes would set `requests` to be same as `limits`](https://kubernetes.io/docs/tasks/configure-pod-container/assign-cpu-resource/#if-you-do-not-specify-a-cpu-limit).
     So, the HPA Placement controller must treat `limits` as same as `requests`
     while selcting clusters.
   * If `requests` and `limits` are the same for `cpu` and `memory`, Kubernetes
     treats that pod differently. See [relevant Kubernetes documentation](https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/).

 * Kubernetes (from v1.12) supports two different CPU manager policies:
   `none` (default) and `static`, which allows pods with certain resource
   characteristics to be granted increased CPU affinity and exclusivity on
   the node. Only containers that are both part of a Guaranteed pod and have
   integer CPU requests are assigned exclusive CPUs.

   To ensure that a pod gets exclusive access to CPU cores, the following
   steps need to be done:

   * The cluster admin must prepare the Linux kernel with `--isolcpus x-y` parameter.
   * The cluster admin must start the kubelet with `--cpu-manager-policy static` and
     `--reserved-cpus` (since v1.17) set outside the isolated CPUs range.
   * The cluster admin may set CPU affinity for `systemd` daemon (which spawns most
     processes in a Linux system).
   * The EMCO admin must create the HPA intent as follows:
     * The `requests` field for `cpu` resource is an integer.
     * The `requests` and `limits` fields are the same for both `cpu` and
       `memory` resources (which makes Kubernetes put that pod in the
       Guaranteed QoS class).

 * The above applies to the default CPU manager that ships with Kubernetes.
   There is an alternative named [CPU Manager for
   Kubernetes](https://github.com/intel/CPU-Manager-for-Kubernetes) which is
   not part of the standard Kubernetes. It enables CPU pinning to specific
   cpusets or specific cores, and other advanced features. However, since it
   is not part of the standard Kubernetes, it is not supported by EMCO.

## HPA Intent API

An HPA intent describes the hardware resource requirements for each
microservice within an application. Each microservice in this context is a
HPA resource consumer and it may need one or both of two kinds of resources:
allocatable and non-allocatable.

A microservice is usually implemented as a Kubernetes Deployment or
StatefulSet but could also be a DaemonSet or a Job. Accordingly, a HPA
resource consumer is expressed in terms of the relevant Kubernetes object to
which the needed HPA resources need to be bound.

### HPA Intents

At the highest level, the APIs allow for CRUD operations on HPA intents.
Note that the  double-quotes inside `body/raw` field are not backslah-protected for readability.

Refer to
[POSTMAN Json Schema](https://schema.getpostman.com/json/collection/v2.1.0/collection.json) for
the syntax used here.

#### Create an intent
```
{
  "name": "Add HPA intent for an application",
  "request": {
    "method": "POST",
    "url": {
       "raw": "{{baseUrl}}/projects/:project-name/composite-apps/:composite-app-name/:composite-app-version/deployment-intent-groups/:deployment-intent-group-name/hpa-intents",
       ...
    },
    "header": [
      {
        "key": "Content-Type",
        "value": "application/json"
      }
    ],
    "body": {
      "mode": "raw",
      "raw": "{
         "metadata": {
            "name": "<string>",
            "description": "<string>",
            "userData1": "<string>",
            "userData2": "<string>"
         },
         "spec": {
            "appName": "<string>",
         }
      }"
    }
  }
}
```

#### GET all HPA intents for an application

```
{
  "name": "Get all HPA intents for an application",
  "request": {
    "method": "GET",
    "url": {
       "raw": "{{baseUrl}}/projects/:project-name/composite-apps/:composite-app-name/:composite-app-version/ deployment-intent-groups/:deployment-intent-group-name/hpa-intents",
       ...
    },
    "header": [],
  },
  "response": [
    {
      "name": "Success",
      "originalRequest": {
         ...
      },
      "status": "OK",
      "code": 200,
      "_postman_previewlanguage": "json",
      "cookie": [],
      "header": [
        {
          "key": "Content-Type",
          "value": "application/json"
        }
      ],
      "body": "[
         "metadata": {
            "name": "<string>",
            "description": "<string>",
            "userData1": "<string>",
            "userData2": "<string>"
         },
         "spec": {
            "appName": "<string>",
            "resourceConsumers": [
               ...
            ]
         }
      ]"
    },
    ...
  ]
}
```

### Get a single HPA intent by name

```
{
  "name": "Get one HPA intent by name",
  "request": {
    "method": "GET",
    "url": {
       "raw": "{{baseUrl}}/projects/:project-name/composite-apps/:composite-app-name/:composite-app-version/ deployment-intent-groups/:deployment-intent-group-name/hpa-intents/:hpa-intent-name",
       ...
    },
    "header": [],
  },
  "response": [
    {
      "name": "Success",
      "originalRequest": {
         ...
      },
      "status": "OK",
      "code": 200,
      "_postman_previewlanguage": "json",
      "cookie": [],
      "header": [
        {
          "key": "Content-Type",
          "value": "application/json"
        }
      ],
      "body": "{
         "metadata": {
            "name": "<string>",
            "description": "<string>",
            "userData1": "<string>",
            "userData2": "<string>"
         },
         "spec": {
            "appName": "<string>"
            "resourceConsumers": [
               ...
            ]
         }
      }"
    },
    ...
  ]
}
```

#### PUT a single HPA intent by name
A PUT is not allowed on an HPA Intent if it has one or more resource
consumers. The resource consumers must be deleted first before the PUT is
allowed.

#### DELETE a single HPA intent by name
A DELETE is not allowed on an HPA Intent if it has one or more resource
consumers. The resource consumers must be deleted first before the HPA intent
itself is deleted.

### HPA Resource Consumers

A resource consumer for an allocatable resource is a container within a pod.
For uniformity, if a microservice needs both allocatable and non-allocatable
resources, the resource consumer is expressed as a container. However, if the
microservice needs only non-allocatable resources, the container field is not
required.

Since a microservice may be a Kubernetes Deployment, StatefulSet, or some
object that contains a pod, the resource consumer is expressed in terms of
any of these Kubernetes objects.

#### Create a resource consumer

```
{
  "name": "Add a resource consumer to a HPA intent",
  "request": {
    "method": "POST",
    "url": {
       "raw": "{{baseUrl}}/projects/:project-name/composite-apps/:composite-app-name/:composite-app-version/ deployment-intent-groups/:deployment-intent-group-name/hpa-intents/:hpa-intent-name/hpa-resource-consumers",
       ...
    },
    "header": [
      {
        "key": "Content-Type",
        "value": "application/json"
      }
    ],
    "body": {
      "mode": "raw",
      "raw": "{
         "metadata": {
            "name": "<string>",
            "description": "<string>",
            "userData1": "<string>",
            "userData2": "<string>"
         },
         "spec": {
            "consumer": {
              "apiVersion": "<string>",
              "kind": "<string>",
              "name": "<string>",
              "container-name": "<string>",
            }
         }
      }"
    }
  }
}
```

In the request body above:

 * `apiVersion`: K8s version. E.g. apps/v1
 * `kind`: Type of object. E.g. Deployment.
 * `name`: metadata/name field of the consumer object
 * `container-name`: required if the consumer needs any allocatable resource,
   optional otherwise.

#### Other operations on resource consumers

One can get all the resource consumers for a given HPA intent with:

```
GET /projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{hpa-intent-name}/hpa-resource-consumers
```

One can get, update or delete a single resource consumer by name,
respectively, with `GET`, `PUT` or `DELETE` on the following URL:

```
/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-      groups/{deployment-intent-group-name}/hpa-intents/{hpa-intent-name}/hpa-resource-consumers/{consumer-name}
```

The `PUT` and `DELETE` operations are not allowed on a resource consumer when it has one or
more resource requirements. The resource requirements must be deleted first
before these operations are allowed.

### HPA Resource Requirements

#### Create a resource requirement for a resource consumer

```
{
  "name": "Add a resource requirement to a resource consumer",
  "request": {
    "method": "POST",
    "url": {
       "raw": "{{baseUrl}}/projects/:project-name/composite-apps/:composite-app-name/:composite-app-version/ deployment-intent-groups/:deployment-intent-group-name/hpa-intents/:hpa-intent-name/hpa-resource-consumers/:consumer-name/resource-requirements",
       ...
    },
    "header": [
      {
        "key": "Content-Type",
        "value": "application/json"
      }
    ],
    "body": {
      "type": "object",
      "properties": {
         "item": {
            "type": "object",
            "name": "metadata",
            "$ref": "#/definitions/metadata"
         },
         "item": {
            "type": "object",
            "name": "spec",
            "items": {
               "item": {
                  "name": "mandatory",
                  "type": "boolean"
               },
               "item": {
                  "name": "weight",
                  "type": "integer"
               },
               "item": {
                  "name": "allocatable",
                  "type": "boolean"
               },
               "oneOf": [
                  { "$ref": "#/definitions/allocatable-resource" },
                  { "$ref": "#/definitions/non-allocatable-resource" },
               ]
            }
         }
      }
    }
  }
}
```

The definition of `allocatable-resource` is as below:

```
{
   "name": "<string>",
   "requests": "<integer>",
   "limits": "<integer>",
   "units": "<string>"
}
```

The `non-allocatable-resource` definition is as below:

```
{
   "name": "<string>",
   "value": "<string>"
}
```

The field names above have the following meaning and semantics:

 * `mandatory`: True if resource is required for a successful deployment.
 * `weight`: An integer in the range `1..10`. if the resource is not
   mandatory, this specifies the priority to be given to this resource. 10 is
   highest. See [Scope of First Implementation](#Scope-of-First-Implementation).
 * `allocatable`: True if this is an allocatable resource, false otherwise.
    The fields of `allocatable-resource` are required if this is true,
    forbidden if false. Conversely, the fields of `non-allocatable-
    resource` are required if this is true, forbidden if false.
 * `name`: Name of the resource in Kubernetes terms. E.g. `cpu`,
   `nvidia.com/gpu` for allocatable resources;
   `feature.node.kubernetes.io/cpu-pstate.turbo` for non-allocatable
   resources.
 * `requests`: Minimum quantity of the resource needed.
 * `limits`: Maximum quantity, to be enforced by Kubernetes.
 * `units`: The unit of the resource. E.g. MB. See
   [Scope of First Implementation](#Scope-of-First-Implementation).
 * `value`: The label value for a non-allocatable resource. E.g. "true".

#### Other operations on resource requirements

One can get all resource requirements for a given resource consumer with:

```
GET /projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{hpa-intent-name}/hpa-resource-consumers/{consumer-name}/resource-requirements
```

One can get, update or delete a single resource consumer by name,
respectively, with `GET`, `PUT` or `DELETE` on the following URL:

```
/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{hpa-intent-name}/hpa-resource-consumers/{consumer-name}/resource-requirements/{requirement-name}
```

### Usage Example

Consider a scenario with a composite application `app1`, which has two
applications `fwall` and `analytics`. The `fwall` app has 2 microservices:
  1. `core-fwall`, which needs 4 exclusive CPU cores and 2 GB memory.
      It can use a GPU if available, but does not need it.
  2. `api-server`, which needs at least 1 CPU core (not exclusive), at most 2 CPU cores and 200 MB memory.

The `analytics` app also has two microservices:
  1. `ai-model`, which needs at least 2 CPU cores (no upper limit).
  2. `logger`, which needs a Quick Assist accelerator.

We assume that the cluster admin has prepared the relevant clusters to enable
exclusive CPU access as [discussed before](#HPA-Intent-Overview).

The set of EMCO API calls to create a HPA intent for this scenario is presented here. We assume that the project, composite app, apps, profiles and deployment intent groups have already been created.

#### HPA intent for fwall app

The following creates an HPA intent for the `fwall` app.
```
/projects/my-proj1/composite-apps/my-capp1/v1.0/deployment-intent-groups/my-dig1/hpa-intents

POST:
   metadata:
      name: fwall-hpa-intent
      description: Assign resurces to core-fwall and api-server
      userData1: foo
      userData2: bar
   spec:
      app: fwall
```

Then we create two resource consumers, one for each microservice:
```
/projects/my-proj1/composite-apps/my-capp1/v1.0/deployment-intent-groups/my-dig1/hpa-intents/fwall-hpa-intent/hpa-resource-consumers

POST:
   metadata:
      name: core-fwall
      description: core-fwall microservice
   spec:
      consumer:
         apiVersion: apps/v1
         kind: StatefulSet
         name: core-fwall
         container-name: palo-alto-e3000

POST:
   metadata:
      name: api-server
      description: api-server microservice
   spec:
      consumer:
         apiVersion: apps/v1
         kind: Deployment
         name: api-server
         container-name: nginx
```

Then we assign resource requirements to the `core-fwall` microservice. To ensure exclusive CPU access, we first force Kubernetes to give Guaranteed QoS (by setting `requests` == `limits` for both `cpu` and `memory`) and by asking for integer requests for CPU.

```
/projects/my-proj1/composite-apps/my-capp1/v1.0/deployment-intent-groups/my-dig1/hpa-intents/fwall-hpa-intent/hpa-resource-consumers/core-fwall/resource-requirements

POST:
   metadata:
      name: core-fwall-cpu-requirement
   spec:
      mandatory: true
      allocatable: true
      name: cpu
      requests: 4
      limits: 4

POST:
   metadata:
      name: core-fwall-RAM-needs
   spec:
      mandatory: true
      allocatable: true
      name: memory
      requests: 1000
      limits: 1000
      units: MB

POST:
   metadata:
      name: core-fwall-gpu-needs
   spec:
      mandatory: false
      weight: 9
      allocatable: true
      name: nvidia.com/gpu
      limits: 1
```

Finally we assign resource requirements to the `api-server` microservice:
```
/projects/my-proj1/composite-apps/my-capp1/v1.0/deployment-intent-groups/my-dig1/hpa-intents/fwall-hpa-intent/hpa-resource-consumers/api-server/resource-   requirements

POST:
   metadata:
      name: api-server-cpu-needs
   spec:
      mandatory: true
      allocatable: true
      name: cpu
      requests: 1
      limits: 2

POST:
   metadata:
      name: api-server-ram-needs
   spec:
      mandatory: true
      allocatable: true
      name: memory
      requests: 200
      units: MB
```

#### HPA intent for analytics app

The following creates an HPA intent for the `analytics` app.
```
/projects/my-proj1/composite-apps/my-capp1/v1.0/deployment-intent-groups/my-dig1/hpa-intents

POST:
   metadata:
      name: analytics-hpa-intent
      description: Assign resurces to ai-model and logger
      userData1: foo
      userData2: bar
   spec:
      app: analytics
```

Then we create two resource consumers, one for each microservice:
```
/projects/my-proj1/composite-apps/my-capp1/v1.0/deployment-intent-groups/my-dig1/hpa-intents/analytics-hpa-intent/hpa-resource-consumers

POST:
   metadata:
      name: ai-model
      description: ai-model microservice
   spec:
      consumer:
         apiVersion: apps/v1
         kind: StatefulSet
         name: ai-model
         container-name: openvino

POST:
   metadata:
      name: logger
      description: logger microservice
   spec:
      consumer:
         apiVersion: apps/v1
         kind: Deployment
         name: logger
         container-name: logger
```

Then we assign resource requirements to the `ai-model` microservice.

```
/projects/my-proj1/composite-apps/my-capp1/v1.0/deployment-intent-groups/my-dig1/hpa-intents/analytics-hpa-intent/hpa-resource-consumers/ai-model/resource-requirements

POST:
   metadata:
      name: ai-model-cpu-requirement
   spec:
      mandatory: true
      allocatable: true
      name: cpu
      requests: 2
   ```

Finally we assign resource requirements to the `logger` microservice:
```
/projects/my-proj1/composite-apps/my-capp1/v1.0/deployment-intent-groups/my-dig1/hpa-intents/analytics-hpa-intent/hpa-resource-consumers/logger/resource-   requirements

POST:
   metadata:
      name: logger-cpu-needs
   spec:
      mandatory: true
      allocatable: true
      name: intel.com/qat
      requests: 1
      limits: 1
```

# DESIGN

## Scope of First Implementation

The principal aim of the first implementation is to get to  functional parity
with the DCC-F functionality for HPA Intents. Accordingly, the resources of
interest for the first implementation are CPU, memory and GPUs; though the
architecture and design are generic across resources, validation will focus
only on these resources.

For the same reason, the following are not priorities for the first
implementation:

  * Fractional CPU specifications (e.g. `cpu: 500m` or `cpu: 1.5`)
  * Resource units other than default: The default CPU units are integer
    cores and memory units are MB. (This mainly helps in unit testing and
    validation.)
  * Weights for non-mandatory resources.
  * Replicas per deployment other than 1.

