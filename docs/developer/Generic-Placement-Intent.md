```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2020 Intel Corporation
```

# Generic Placement Intents

EMCO supports deploying applications to multiple clusters. Generic Placement Intents are used by EMCO to describe the clusters the user desires that the applications to be deployed on.

There are 2 types of intents that are supported.
- AllOf - This intent is used when the application *must* be deployment on all the clusters in the list
- AnyOf - This intent is used when the application has a group of clusters and gives option to EMCO to pick one of the clusters based on decisions made by the Placement controllers. AnyOf means, <b>“any one of the clusters”</b> in the array will be used for deployment. In case of label, only 1 cluster among a group of clusters resolved under the label will be selected.


## Intent specification

Each element of AllOf and AnyOf list consists of following fields:

<b>ProviderName</b> : The name of the provider of the cluster.

<b>ClusterName</b> : The name of the cluster can be given explicitly.

<b>ClusterLabelName</b> :  The cluster label can be given explicitly.

<b>NOTE </b>: Either the ClusterName or ClusterLabelName are required at a time. Specifying both together is an error.</b>

AllOf list can also have AnyOf as part of it as shown in the examples below.

<b>Example Scenario </b> :

Suppose you have to deploy on cluster1, cluster2 and either on cluster3 or cluster4.

In this case,

<b>AllOfArray</b> shall consists of two elements – one each for cluster1 and cluster2.
<b>AnyOfArray</b> shall consists of two elements – one each for cluster3 and cluster4.

There shall be 2 possibilities of where final deployment shall take place:

<b>1.</b> Cluster1, Cluster2 and Cluster3

<b>2.</b> Cluster1, Cluster2 and Cluster4

```
intent:
    allOf:
    - provider-name: clusterprovidername
      cluster-label-name: cluster1
    - provider-name: clusterprovidername
      cluster-label-name: cluster2
    anyOf:
      - provider-name: clusterprovidername
        cluster-label-name: cluster3
      - provider-name: clusterprovidername
        cluster-label-name: cluster4
```
<b>Example Scenario </b> :

Suppose you have a group of clusters under <b>Label1, Label2, Label3 and Label4</b>.

You want to deploy on all the clusters with Label1 and Label2 and you want to deploy on any 1 cluster under Label3 and Label4.

In this case, you can use the combination of <b>AnyOf</b> and <b>AllOf</b>.

<b>NOTE</b> : There can be scenarios when the same topology can be realised by two different intent structures. For example :

```
{
   "intent": {
      "allOf": [
         {
            "provider-name": "clusterprovidername",
            "cluster-label-name": "labelname1"
         },
         {
            "provider-name": "clusterprovidername",
            "cluster-label-name": "labelname2"
         },
         {
            "anyOf": [
               {
                  "provider-name": "clusterprovidername",
                  "cluster-label-name": "labelname3"
               },
               {
                  "provider-name": "clusterprovidername",
                  "cluster-label-name": "labelname4"
               }
            ]
         }
      ]
   }
}
```

The above is equivalent to :

```
{
  "intent": {
    "allOf": [
      {
        "provider-name": "clusterprovidername",
        "cluster-label-name": "labelname1"
      },
      {
        "provider-name": "clusterprovidername",
        "cluster-label-name": "labelname2"
      }
    ],
    "anyOf": [
      {
        "provider-name": "clusterprovidername",
        "cluster-label-name": "labelname3"
      },
      {
        "provider-name": "clusterprovidername",
        "cluster-label-name": "labelname4"
      }
    ]
  }
}
```

Yaml for the same intent

```
intent:
    allOf:
    - provider-name: clusterprovidername
      cluster-label-name: labelname1
    - provider-name: clusterprovidername
      cluster-label-name: labelname2
    - anyOf:
      - provider-name: clusterprovidername
        cluster-label-name: labelname3
      - provider-name: clusterprovidername
        cluster-label-name: labelname4
```
<b>Example Scenario </b> :

Another scenario that requires deploying of either on cluster1 or cluster2 and either on cluster3 or cluster4.

In this case,

<b>AllOf</b> consists of two elements – one each of the AnyOf selection.

<b>AnyOf</b> There are two AnyOf elements – one for cluster1 and cluster2 and one for cluster3 and cluster4.

```
intent:
    allOf:
      - anyOf:
          - provider-name: p
            cluster-name: cluster1
          - provider-name: p
            cluster-name: cluster2
      - anyOf:
          - provider-name: p
            cluster-name: cluster3
          - provider-name: p
            cluster-name: cluster4
```


### The concept of Group Number

<b>Group Number</b>: Group number is an internal concept used by <b>EMCO</b>.

<b>Purpose</b>: This is used by orchestrators to resolve the AllOf and AnyOf intents and convert them to groups that can be used by the Placement controllers to figure out what are the various groups of clusters that placement controllers have a choice.

To better understand, let's take an example, we have a request in <b>AllOf</b> for Label1 and Label2 and there is an <b>AnyOf</b> for clusterLabel Label3.

The <b>Label1</b> resolves to <b>ClusterName1 and ClusterName2.</b>

The <b>Label2</b> resolves to <b>ClusterName3 and ClusterName4.</b>

The <b>Label3</b> resolves to <b>ClusterName5 and ClusterName6.</b>

In this case, the groups shall be assigned as follows.

<b>Group1 - ClusterName1

Group2 – ClusterName2

Group3 – ClusterName3

Group4 – ClusterName4

Group5 – ClusterName5 and ClusterName6
</b>

All the clusters where the deployment is to take place mandatorily are assigned an internal group number.
* So, <b>ClusterName1, ClusterName2, ClusterName3, ClusterName4</b>, have been assigned separate group number.

* Since <b>ClusterName5 and ClusterName6</b> have both been assigned the same group number, only 1 among shall be chosen for deployment. Each Placement controller gets to decide which clusters in a group doesn't fit its constraints and it removes those clusters from the group. 

### Generic Placement Intent Internal Structures

```
type IntentStruc struct {
    AllOfArray []AllOf `json:"allOf,omitempty"`
    AnyOfArray []AnyOf `json:"anyOf,omitempty"`
}
```

This is the structure that we have in place for the Generic Placement Intent.

It consists of two components:

<b>AllOfArray</b>: This is the All-Of Array. It denotes an array of clusters on which the app shall be deployed. AllOf stands for deploy on <b>“all of”</b> them.

Below is the structure of <b>AllOf</b>

```
type AllOf struct {
    ProviderName     string  `json:"provider-name,omitempty"`
    ClusterName      string  `json:"cluster-name,omitempty"`
    ClusterLabelName string  `json:"cluster-label-name,omitempty"`
    AnyOfArray       []AnyOf `json:"anyOf,omitempty"`
}
```

Let’s further explore AnyOf, here is the structure for AnyOf:

```
type AnyOf struct {
    ProviderName     string `json:"provider-name,omitempty"`
    ClusterName      string `json:"cluster-name,omitempty"`
    ClusterLabelName string `json:"cluster-label-name,omitempty"`
}
```
