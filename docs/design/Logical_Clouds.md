```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2020 Intel Corporation
```
# Getting Started
This document describes Logical Clouds in terms of EMCO's worldview and how to make use of them.

Logical Clouds are an abstraction provided by the Distributed Cloud Manager (DCM), a key EMCO microservice.


# Distributed Cloud Manager
The Distributed Cloud Manager (DCM) provides the Logical Cloud abstraction and effectively completes the concept of "multi-cloud." One Logical Cloud is a grouping of one or many clusters, each with their own control plane, specific configurations and geo-location, which get partitioned for a particular EMCO project. This partitioning is made via the creation of distinct, isolated namespaces in each of the Kubernetes\* clusters that thus make up the Logical Cloud.

A Logical Cloud is the overall target of a Deployment Intent Group and is a mandatory parameter (the specific applications under it further refine what gets run and in which location.) A Logical Cloud must be explicitly created and instantiated before a Deployment Intent Group can be instantiated.

Due to the close relationship with Clusters, which are provided by Cluster Registration (`clm`) above, it is important to understand the mapping between the two. A Logical Cloud groups many Clusters together but a Cluster may also be divided among multiple Logical Clouds, effectively turning the cluster multi-tenant. The partitioning/multi-tenancy of a particular Cluster into different Logical Clouds is done today at the namespace level. Different Logical Clouds access different namespace names, and the name is consistent across the multiple clusters of the Logical Cloud.

![Mapping between Logical Clouds and Clusters](images/emco-lccl.png)

_Figure 1 - Mapping between Logical Clouds and Clusters_

## Types of Logical Clouds

### Standard Logical Clouds
Logical Clouds were introduced to group and partition clusters in a multi-tenant way and across boundaries, improving flexibility and scalability. A Standard Logical Cloud is the default type of Logical Cloud providing just that much. When projects request a Logical Cloud to be created, they provide which permissions are available, the resource quotas and clusters that compose it. The Distributed Cloud Manager, alongside the Resource Synchronizer, sets up all the clusters accordingly, with the necessary credentials, namespace/resources, and finally generating the kubeconfig files used to authenticate/reach each of those clusters in the context of the Logical Cloud.

### Admin Logical Clouds
In some use cases, and in the administrative domains where it makes sense, a project may want to access raw, unmodified, administrator-level clusters. For such cases, no namespaces need to be created and no new users need to be created or authenticated in the API. To solve this, the Distributed Cloud Manager introduces Admin Logical Clouds, which offer the same consistent interface as Standard Logical Clouds to the Distributed Application Scheduler. Being of type Admin means this is a Logical Cloud at the Administrator level. As such, no changes will be made to the clusters themselves. Instead, the only operation that takes place is the reuse of credentials already provided via the Cluster Registration API for the clusters assigned to the Logical Cloud (instead of generating new credentials, namespace/resources and kubeconfig files.)

### Privileged Logical Clouds
This type of Logical Cloud provides most of the capabilities that an Admin Logical Cloud provides but at the user-level like a Standard Logical Cloud. New namespaces are created, with new user and kubeconfig files. However, the EMCO project can now request an enhanced set of permissions/privileges, including targeting cluster-wide Kubernetes resources.

## Lifecycle Operations
Prerequisites to using Logical Clouds:
* With the project-less Cluster Registration API, create the cluster providers, clusters and optionally cluster labels.
* With the Distributed Application Scheduler API, create a project which acts as a tenant in EMCO.

The basic flow of lifecycle operations to get a Logical Cloud up and running via the Distributed Cloud Manager API is:
* Create a Logical Cloud specifying the following attributes:
  - Level: For Standard/Privileged Logical Clouds, set to 1. For Admin Logical Clouds, set to 0.
  - (*for Standard/Privileged only*) Namespace name - the namespace to use in all of the Clusters of the Logical Cloud.
  - (*for Standard/Privileged only*) User name - the name of the user that will be authenticating to the Kubernetes* APIs to access the namespaces created.
* (*for Standard/Privileged only*) User permissions - permissions that the user specified will have in the namespace specified, in all of the clusters.
* (*for Standard/Privileged only*) Create resource quotas and assign them to the Logical Cloud created: this specifies what quotas/limits the user will face in the Logical Cloud, for each of the Clusters.
* Assign the Clusters previously created with the project-less Cluster Registration API to the newly-created Logical Cloud.
* Instantiate the Logical Cloud. All of the clusters assigned to the Logical Cloud are automatically set up to join the Logical Cloud. Once this operation is complete, the Distributed Application Scheduler's lifecycle operations can be followed to deploy applications on top of the Logical Cloud.

Apart from the creation/instantiation of Logical Clouds, the following operations are also available:
* Terminate a Logical Cloud - this removes all of the Logical Cloud -related resources from all of the respective Clusters.
* Delete a Logical Cloud - this eliminates all traces of the Logical Cloud in EMCO.

## Deploying Logical Clouds

EMCO comes bundled with many `emcoctl` example files designed to make it easy to understand the multiple abstractions and how to deploy them in order. EMCOctl is the preferred client to interact with the EMCO APIs.

Logical Clouds have to exist before the instantiation of apps / Deployment Intent Groups can take place (as mentioned above).

Let's take a look at the examples bundled with EMCO, which can be seen in `main/src/tools/emcoctl/examples/`. There are two Logical Cloud related folders there:
* `l1`, containing `emcoctl` yaml files that deploy a sample 2-cluster Standard Logical Cloud.
  * `l0`, containing `emcoctl` yaml files that deploy a sample 2-cluster Admin Logical Cloud.

For consistency, both folders contain 3 identically-named files (but they are not the same):
* `1-logical-cloud-prerequisites.yaml`, defines all Logical Cloud related resources that need to be created on a clean deployment
* `2-logical-cloud-instantiate.yaml`, defines a single operation, **instantiate**, for the Logical Cloud
* `values.yaml`, templates the variables to be used in the yaml files above.

Additionally, it's important to be aware that `emcoctl` also requires an EMCO configuration file, such as the example `emco-cfg.yaml` in the same directory as the `l0` and `l1` folders. This file should be modified according to your environment.

These instructions assume the `emcoctl` command has been loaded to the `$PATH`.

### Standard Logical Clouds

#### Create

To deploy one Standard Logical Cloud with two clusters according to the example files above, execute the following (from within `src/tools/emcoctl/examples/l1`):

```
emcoctl --config ../emco-cfg.yaml -v values.yaml -f 1-logical-cloud-prerequisites.yaml apply
emcoctl --config ../emco-cfg.yaml -v values.yaml -f 2-logical-cloud-instantiate.yaml apply
```

#### Delete

To delete the Logical Cloud created, execute the following (from within `src/tools/emcoctl/examples/l1`):

```
emcoctl --config ../emco-cfg.yaml -v values.yaml -f 2-logical-cloud-instantiate.yaml delete
emcoctl --config ../emco-cfg.yaml -v values.yaml -f 1-logical-cloud-prerequisites.yaml delete
```

(Notice the reversed order.)

#### Customize

Using the examples is straightforward, but what about customizing the Standard Logical Cloud?

Everything that can be customized is located in the first file, `1-logical-cloud-prerequisites.yaml`. Edit it to add your customizations.

Different clusters can be created and assigned to the Logical Cloud. Here's one such cluster being assigned to the Logical Cloud, starting from the end of the example yaml file:

    ---
    #add cluster reference to logical cloud
    version: emco/v2
    resourceContext:
      anchor: projects/ns1/logical-clouds/lc1/cluster-references
    metadata:
      name: lc-cl-1
    spec:
      cluster-provider: {{.ClusterProvider}}
      cluster-name: {{.Cluster1}}
      loadbalancer-ip: "0.0.0.0"

The **name** given to this resource is not important, but it represents the name of the cluster as associated to the Logical Cloud. The most important parts are assigning the right `cluster-provider` and `cluster-name` fields, which should match previously created clusters and the cluster provider. And, of course, the `anchor` should reflect the right project name and logical cloud, also previously created.

What about customizing the Logical Cloud itself? I.e., modifying what the cloud project is allowed to do in such Logical Cloud. This is divided in two parts:

**First**, the creation of the Logical Cloud itself, including the specification of permissions (which will get translated to Kubernetes roles and role bindings via EMCO):

    ---
    #create logical cloud without admin permissions
    version: emco/v2
    resourceContext:
      anchor: projects/proj1/logical-clouds
    metadata:
      name: lc1
    spec:
      namespace: ns1
      user:
        user-name: user-1
        type: certificate

The namespace can also be renamed. This will be the namespace name that EMCO will create in every cluster associated with this Logical Cloud.
Logical Clouds that don't specify a `level` field will automatically default to Standard and consequently expect the namespace to be provided.

**Second**, the creation of User Permissions (which get translated to Role/RoleBindings or Cluster/ClusterRoleBindings in Kubernetes):

    ---
    #create logical cloud without admin permissions
    version: emco/v2
    resourceContext:
      anchor: projects/proj1/logical-clouds
    metadata:
      name: permission-1
    spec:
      namespace: ns1
      apiGroups:
      - ""
      - "apps"
      - "k8splugin.io"
      resources:
      - secrets
      - pods
      - configmaps
      - services
      - deployments
      - resourcebundlestates
      verbs:
      - get
      - watch
      - list
      - create

It's important to define `apiGroups`, `resource` and `verbs` according to the intended goal. They are scoped to the namespace specified in the User Permission.
If the namespace is empty, i.e. `""`, then the scope of the three variables above is the cluster, i.e. cluster-wide. **This is what determines that the Logical Cloud will be of Privileged type**.

**Third**, the creation of Cluster Quotas (which get translated to resource quotas in Kubernetes):

    ---
    #create cluster quotas
    version: emco/v2
    resourceContext:
      anchor: projects/proj1/logical-clouds/lc1/cluster-quotas
    metadata:
        name: quota1
    spec:
        limits.cpu: '400'
        limits.memory: 1000Gi
        requests.cpu: '300'
        requests.memory: 900Gi
        requests.storage: 500Gi
        requests.ephemeral-storage: '500'
        limits.ephemeral-storage: '500'
        persistentvolumeclaims: '500'
        pods: '500'
        configmaps: '1000'
        replicationcontrollers: '500'
        resourcequotas: '500'
        services: '500'
        services.loadbalancers: '500'
        services.nodeports: '500'
        secrets: '500'
        count/replicationcontrollers: '500'
        count/deployments.apps: '500'
        count/replicasets.apps: '500'
        count/statefulsets.apps: '500'
        count/jobs.batch: '500'
        count/cronjobs.batch: '500'
        count/deployments.extensions: '500'

Many resource quota types can be defined. The semantics match Kubernetes' semantics for Resource Quotas. Make sure the Logical Cloud referenced in the `anchor` matches the Logical Cloud previously created - the one where permissions were specified.

### Admin Logical Clouds

#### Create

To deploy one Admin Logical Cloud with two clusters according to the example files above, execute the following (from within `src/tools/emcoctl/examples/l0`):

```
emcoctl --config ../emco-cfg.yaml -v values.yaml -f 1-logical-cloud-prerequisites.yaml apply
emcoctl --config ../emco-cfg.yaml -v values.yaml -f 2-logical-cloud-instantiate.yaml apply
```

#### Delete

To delete the Logical Cloud created, execute the following (from within `src/tools/emcoctl/examples/l1`):

```
emcoctl --config ../emco-cfg.yaml -v values.yaml -f 2-logical-cloud-instantiate.yaml delete
emcoctl --config ../emco-cfg.yaml -v values.yaml -f 1-logical-cloud-prerequisites.yaml delete
```

(Notice the reversed order.)

#### Customize

Besides associating clusters, there is no other real customization that can be made to Admin Logical Clouds, as they are designed to retrieve and reuse the administrator access to the associated clusters (i.e. it doesn't make sense to change permissions, quotas or namespace.)

But how to set a Logical Cloud to Admin instead of Standard?

Instead of defining namespace and user permissions in the `spec` portion of the Logical Cloud, simply set the `level` field to `"0"` (Logical Clouds that don't specify a level will automatically default to Standard.) Here's what an Admin Logical Cloud definition in an emcoctl yaml file looks like:

    ---
    #create default logical cloud with admin permissions
    version: emco/v2
    resourceContext:
      anchor: projects/ns1/logical-clouds
    metadata:
      name: lc1
    spec:
      level: "0"

Notice that `level: "0"` is the only element within `spec`.

### Instantiation and Termination

Instantiation of Logical Clouds is mandatory before any non-DCM instantiation operation that requires Logical Clouds can be executed. It tells EMCO to go ahead and create the resources in the right clusters and set any relevant entity to the right state. This requirement applies to both Admin, Standard and Privileged Logical Clouds.

Both of the instantiation example files (`2-logical-cloud-instantiate.yaml`) contain the following yaml resource:

    ---
    #instantiate logical cloud
    version: emco/v2
    resourceContext:
      anchor: projects/ns1/logical-clouds/lc1/instantiate

Which is a simple `emcoctl` request that asks the DCM API to begin instantiating the referenced Logical Cloud.

With regards to the reverse operation, **terminate**, it's simply a matter of calling emcoctl with the `delete` command. The yaml above will be converted to a terminate operation in runtime. Same with the yaml resources that create EMCO resources - they will get converted to delete operations in runtime when emcoctl is called with the `delete` command.
