# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

#################################################################
# EMCOCTL - CLI for EMCO
#################################################################

Emoctl is command line tool for interacting with EMCO.
All commands take input a file. An input file can contain one or more resources.


### Syntax for describing a resource

```
version: <domain-name>/<api-version>
resourceContext:
  anchor: <URI>
Metadata :
   Name: <name>
   Description: <text>
   userData1: <text>
   userData2: <text>
Spec:
  <key>: <value>
```

### Example resource file

```
version: emco/v2
resourceContext:
  anchor: projects
Metadata :
   Name: proj1
   Description: test
   userData1: test1
   userData2: test2

---
version: emco/v2
resourceContext:
  anchor: projects/proj1/composite-apps
Metadata :
  name: vFw-demo
  description: test
  userData1: test1
  userData2: test2
spec:
  version: v1
```

### EMCO CLI Commands

1. Create Emco Resources

This command will apply the resources in the file. The user is responsible to ensuring the hierarchy of the resources.

`$ emcoctl apply -f filename.yaml`

For applying resources that don't have a json body anchor can be provided as an arguement

`$ emcoctl apply <anchor>`

`$ emcoctl apply projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group/instantiate`


2. Get Emco Resources

Get the resources in the input file. This command will use the metadata name in each of the resources in the file to get information about the resource.

`$ emcoctl get -f filename.yaml`

For getting information for one resource anchor can be provided as an arguement

`$ emcoctl get <anchor>`

`$ emcoctl get projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group`

3. Delete Emco Resources

Delete resources in the file. The emcoctl will start deleting resources in the reverse order than given in the file to maintain hierarchy. This command will use the metadata name in each of the resources in the file to delete the resource..

`$ emcoctl delete -f filename.yaml`

For deleting one resource anchor can be provided as an arguement

`$ emcoctl delete <anchor>`

`$ emcoctl delete projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group`

4. Update Emco Resources

This command will call update (PUT) for the resources in the file.

`$ emcoctl update -f filename.yaml`

## Using helm charts through emcoctl

When you need to use emcoctl for deploying helm
charts the following steps are required.

1. Make sure that the composite app which you are planning to deploy, the tree structure is as below

```

$  tree collection/app1/
collection/app1/
├── helm
│   └── collectd
│       ├── Chart.yaml
│       ├── resources
│       │   └── collectd.conf
│       ├── templates
│       │   ├── configmap.yaml
│       │   ├── daemonset.yaml
│       │   ├── _helpers.tpl
│       │   ├── NOTES.txt
│       │   └── service.yaml
│       └── values.yaml
└── profile
    ├── manifest.yaml
    └── override_values.yaml

5 directories, 10 files

$  tree collection/m3db/
collection/m3db/
├── helm
│   └── m3db
│       ├── Chart.yaml
│       ├── del.yaml
│       ├── templates
│       │   └── m3dbcluster.yaml
│       └── values.yaml
└── profile
    ├── manifest.yaml
    └── override_values.yaml

4 directories, 6 files

```

### NOTE
```
* In the above example, we have a composite app : collection
The collection composite-app shown has two apps : app1(collectd)
and m3db
* Each app has two dirs : a. HELM and b. PROFILE.
* Helm dir shall have the real helm charts of the app.
* profile shall have the two files - manifest.yaml and override_values.yaml for creating the customized profile.
```

### Commands for making the tar files from helm.

```
    tar -czf collectd.tar.gz -C $test_folder/vnfs/comp-app/collection/app1/helm .
    tar -czf collectd_profile.tar.gz -C $test_folder/vnfs/comp-app/collection/app1/profile .
    ----------------------------------------
    tar -czf m3db.tar.gz -C $test_folder/vnfs/comp-app/collection/m3db/helm .
    tar -czf m3db_profile.tar.gz -C $test_folder/vnfs/comp-app/collection/m3db/profile .
```

Once you have generated the tar files, you need to give the path in file which you are applying using the emcoctl. For eg:

```
#adding collectd app to the composite app
version: emco/v2
resourceContext:
  anchor: projects/proj1/composite-apps/collection-composite-app/v1/apps
metadata :
  name: collectd
  description: "description for app"
  userData1: test1
  userData2: test2
file:
  /opt/csar/cb009bfe-bbee-11e8-9766-525400435678/collectd.tar.gz

```

```
#adding collectd app profiles to the composite profile
version: emco/v2
resourceContext:
  anchor: projects/proj1/composite-apps/collection-composite-app/v1/composite-profiles/collection-composite-profile/profiles
metadata :
  name: collectd-profile
  description: test
  userData1: test1
  userData2: test2
spec:
  app-name: collectd
file:
  /opt/csar/cb009bfe-bbee-11e8-9766-525400435678/collectd_profile.tar.gz

```

### Running the emcoctl

```
* Make sure that the emcoctl is build.You can build it by issuing the 'make' command.
Dir : $MULTICLOUD-K8s_HOME/src/tools/emcoctl
```
* Then run the emcoctl by command:
```
./emcoctl --config ./examples/emco-cfg.yaml apply -f ./examples/test.yaml

```
In case you have a separate <b>values.yaml</b> which specifies all the values used :

```
./emcoctl --config ./examples/emco-cfg.yaml apply -f ./examples/test.yaml -v ./examples/templates/values.yaml
```

Here, emco-cfg.yaml contains the config/port details of each of the microservices you are using.
A sample configuration is :

```
  orchestrator:
    host: localhost
    port: 9015
  clm:
    host: localhost
    port: 9019
  ncm:
    host: localhost
    port: 9016
  ovnaction:
    host: localhost
    port: 9051
```
### Running the emcoctl with template file

```
* Emcoctl supports template values in the input file. The example input file with this feature is
examples/test_template.yaml. This file can be used with examples/values.yaml like below.
```
* Then run the emcoctl with values file:
```
emcoctl --config ./examples/emco-cfg.yaml apply -f ./examples/test_template.yaml -v ./examples/values.yaml

```
### Running the emcoctl with token

```
* Emcoctl supports JWT tokens for interacting with EMCO when EMCO services are running with Istio Ingress and OAuth2 server.
```
* Then run the emcoctl with values file:
```
emcoctl --config ./examples/emco-cfg.yaml apply -f ./examples/test_template.yaml -t "<token>"

```
### Status queries with emcoctl
Some resources, like the Deployment Intent Group support status queries.  The status query provides information about the
resource and any apps which have deployed to remote clusters.  The status query supports a variety of paramters which can
be used to modify the output.

The following query parameters are available:

`type`=< `rsync` | `cluster` >
* default type is 'rsync'
* `rsync`: gathers status based on the rsync resources.
* `cluster`: gathers status based on cluster resource information received in the ResourceBundleState CRs received from the cluster(s)

`output`=< `summary` | `all` | `detail` >
* default output value is: 'all'
* `summary`: will just show the top level EMCO resource state and status along with aggregated resource statuses but no resource detail information
  any filters added will affect the aggregated resource status results, although resource details will not be displayed
* `all`: will include a list of resources, organized by App and Cluster, showing basic resource identification (Group Version Kind) and resource statuses
* `detail`: includes in the resource list the metadata, spec, and status of the resource as received in the ResourceBundleState CR


The following query parameters filter the results returned.  Aggregated status results at the top level are relative to the filter parameters supplied
These parameters can be supplied multiple times in a given status query.

`app`=< `appname` >
* default is all apps
* This will filter the results of the query to show results only for the resources of the specified App(s).

`cluster=< `cluster` >
* default is all clusters
* This will filter the results of the query to show results only for the specified cluster(s)

`resource`=< `resource name` >
* default is all resources
* This will filter the results of the query to show results only for the specified resource(s)

The following query parameters may be included in status queries for `Deployment Intent Groups`.  If one of these parameters is present, then the status
query will make the corresponding query.  Any other query parameters that are not appropriate will be ignored.

`apps`
* Return a list of all of the apps for this app context.
* This parameter takes precedence over `clusters` and `resources` query parameters.
* The `instance` query parameter may be provided.

`clusters`
* Returns a list of clusters to which this `Deployment Intent Group` will be deployed
* This parameter takes precedence over the `resources` query parameter.
* The `app` query filter may be included to filter the response to just the clusters to which the supplied app(s) are deployed.
* The `instance` query parameter may be provided.

`resources`
* Returns a list of resources for this `Deployment Intent Group`,
* The `app` query filter may be included to filter the response to just the resources for the supplied app(s).
* The `instance` query parameter may be provided.
* The `type` parameter may be supplied to return results for either `rsync` or `cluster` resources.
* If `type`=`cluster` is provided, then the `cluster` query filter may also be provided to filter results for the suppplied cluster(s).


Refer to the status query section of  [Resource Lifecycle](../../../docs/user/Resource_Lifecycle.md) to see more examples of the output of various status queries.

#### Status query examples
This section illustrates how to provide status query parameters to the `emcoctl` tool.

Basic status query.  By default, all apps, clusters and resources will be displayed.  The default query type is `rsync`, so the status returned indicates
the status of whether or not EMCO has successfully applied or terminated the resources (not the actual resource status in the cluster).
```
emcoctl --config emco-cfg.yaml get projects/proj1/composite-apps/collection-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/status

```

Query showing just the summary information of the Deployment Intent Group.
```
emcoctl --config emco-cfg.yaml get projects/proj1/composite-apps/collection-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/status\?output=summary

```
Query showing the detailed status of two resources in a given cluster.
Note that the cluster is specified as the `clusterprovider+cluster`.  The `+` is represented in ascii notation `%2B`.
```
emcoctl --config emco-cfg.yaml get projects/proj1/composite-apps/collection-composite-app/v1/deployment-intent-groups/collection-deployment-intent-group/status\?resource=alertmanagers.monitoring.coreos.com\&resource=servicemonitors.monitoring.coreos.com\&output=all\&cluster=provider1\%2Bcluster1\&output=detail

```
