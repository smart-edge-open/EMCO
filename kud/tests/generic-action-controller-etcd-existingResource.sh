#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation



set -o errexit
set -o nounset
set -o pipefail


source _functions.sh
source _common.sh

base_url_orchestrator=${base_url_orchestrator:-"http://172.25.55.56:9015/v2"}
base_url_clm=${base_url_clm:-"http://172.25.55.56:9017/v2"}
base_url_gac=${base_url_gac:-"http://172.25.55.56:9020/v2"}
base_url_dcm=${base_url_dcm:-"http://172.25.55.56:9081/v2"}

# base_url_clm=${base_url_clm:-"http://192.168.121.29:30073/v2"}
# base_url_ncm=${base_url_ncm:-"http://192.168.121.29:31955/v2"}
# base_url_orchestrator=${base_url_orchestrator:-"http://192.168.121.29:32447/v2"}
# base_url_rysnc=${base_url_orchestrator:-"http://192.168.121.29:32002/v2"}


CSAR_DIR="/opt/csar"
csar_id="operators-cb009bfe-bbee-11e8-9766-525400435678"


app1_helm_path="$CSAR_DIR/$csar_id/operator.tar.gz"
app1_profile_path="$CSAR_DIR/$csar_id/operator_profile.tar.gz"



# ---------BEGIN: SET CLM DATA---------------

clusterprovidername="collection-cluster1-provider"
clusterproviderdata="$(cat<<EOF
{
  "metadata": {
    "name": "$clusterprovidername",
    "description": "description of $clusterprovidername",
    "userData1": "$clusterprovidername user data 1",
    "userData2": "$clusterprovidername user data 2"
  }
}
EOF
)"

clustername="cluster1"
clusterdata="$(cat<<EOF
{
  "metadata": {
    "name": "$clustername",
    "description": "description of $clustername",
    "userData1": "$clustername user data 1",
    "userData2": "$clustername user data 2"
  }
}
EOF
)"

kubeconfigcluster1="/opt/kud/multi-cluster/cluster1/artifacts/admin.conf"

labelname="LabelCluster1"
labeldata="$(cat<<EOF
{"label-name": "$labelname"}
EOF
)"

# add logical cloud

admin_logical_cloud_name="lcadmin"
admin_logical_cloud_data="$(cat << EOF
{
 "metadata" : {
    "name": "${admin_logical_cloud_name}",
    "description": "logical cloud description",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "level": "0"
  }
 }
}
EOF
)"

# skipping fancy cluster reference generation for now - just 1 logical cloud cluster reference
lc_cluster_1_name="lc1-c1"
cluster_1_data="$(cat << EOF
{
 "metadata" : {
    "name": "${lc_cluster_1_name}",
    "description": "logical cloud cluster 1 description",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },

 "spec" : {
    "cluster-provider": "${clusterprovidername}",
    "cluster-name": "${clustername}",
    "loadbalancer-ip" : "0.0.0.0"
  }
}
EOF
)"




# add the rsync controller entry
rsynccontrollername="rsync"
rsync_service_host="172.25.55.56"
rsync_service_port=9031
rsynccontrollerdata="$(cat<<EOF
{
  "metadata": {
    "name": "rsync",
    "description": "description of $rsynccontrollername controller",
    "userData1": "user data 1 for $rsynccontrollername",
    "userData2": "user data 2 for $rsynccontrollername"
  },
  "spec": {
    "host": "${rsync_service_host}",
    "port": ${rsync_service_port}
  }
}
EOF
)"



# add the genericaction controller entry
genericactioncontrollername="genericaction"
genericaction_service_host="172.25.55.56"
genericaction_service_port=9033
genericactioncontrollerdata="$(cat<<EOF
{
  "metadata": {
    "name": "$genericactioncontrollername",
    "description": "description of $genericactioncontrollername controller",
    "userData1": "user data 1 for $genericactioncontrollername",
    "userData2": "user data 2 for $genericactioncontrollername"
  },
  "spec": {
    "host": "${genericaction_service_host}",
    "type": "action",
    "priority": 1,
    "port": ${genericaction_service_port}
  }
}
EOF
)"




# ------------END: SET CLM DATA--------------


#-------------BEGIN:SET ORCH DATA------------------

# define a project
projectname="OperatorsProjectCluster1"
projectdata="$(cat<<EOF
{
  "metadata": {
    "name": "$projectname",
    "description": "description of $projectname controller",
    "userData1": "$projectname user data 1",
    "userData2": "$projectname user data 2"
  }
}
EOF
)"

# define a composite application
operators_compositeapp_name="OperatorsCompositeApp"
compositeapp_version="v1"
compositeapp_data="$(cat <<EOF
{
  "metadata": {
    "name": "${operators_compositeapp_name}",
    "description": "description of ${operators_compositeapp_name}",
    "userData1": "user data 1 for ${operators_compositeapp_name}",
    "userData2": "user data 2 for ${operators_compositeapp_name}"
   },
   "spec":{
      "version":"${compositeapp_version}"
   }
}
EOF
)"

# add operator into operators compositeApp


operator_app_name="operator"
operator_helm_chart=${app1_helm_path}

operator_app_data="$(cat <<EOF
{
  "metadata": {
    "name": "${operator_app_name}",
    "description": "description for app ${operator_app_name}",
    "userData1": "user data 2 for ${operator_app_name}",
    "userData2": "user data 2 for ${operator_app_name}"
   }
}
EOF
)"


# Add the composite profile
operators_composite_profile_name="operators_composite-profile"
operators_composite_profile_data="$(cat <<EOF
{
   "metadata":{
      "name":"${operators_composite_profile_name}",
      "description":"description of ${operators_composite_profile_name}",
      "userData1":"user data 1 for ${operators_composite_profile_name}",
      "userData2":"user data 2 for ${operators_composite_profile_name}"
   }
}
EOF
)"

# Add the operator profile data into operators composite profile data
operator_profile_name="operator-profile"
operator_profile_file=$app1_profile_path
operator_profile_data="$(cat <<EOF
{
   "metadata":{
      "name":"${operator_profile_name}",
      "description":"description of ${operator_profile_name}",
      "userData1":"user data 1 for ${operator_profile_name}",
      "userData2":"user data 2 for ${operator_profile_name}"
   },
   "spec":{
      "app-name":  "${operator_app_name}"
   }
}
EOF
)"



# define the generic placement intent
generic_placement_intent_name="Operators-generic-placement-intent"
generic_placement_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${generic_placement_intent_name}",
      "description":"${generic_placement_intent_name}",
      "userData1":"${generic_placement_intent_name}",
      "userData2":"${generic_placement_intent_name}"
   }
}
EOF
)"

# define placement intent for operator sub-app
operator_placement_intent_name="operator-placement-intent"
operator_placement_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${operator_placement_intent_name}",
      "description":"description of ${operator_placement_intent_name}",
      "userData1":"user data 1 for ${operator_placement_intent_name}",
      "userData2":"user data 2 for ${operator_placement_intent_name}"
   },
   "spec":{
      "app-name":"${operator_app_name}",
      "intent":{
         "allOf":[
            {  "provider-name":"${clusterprovidername}",
               "cluster-label-name":"${labelname}"
            }
         ]
      }
   }
}
EOF
)"


# define a deployment intent group
release="operators"
deployment_intent_group_name="operators_deployment_intent_group"
deployment_intent_group_data="$(cat <<EOF
{
   "metadata":{
      "name":"${deployment_intent_group_name}",
      "description":"descriptiont of ${deployment_intent_group_name}",
      "userData1":"user data 1 for ${deployment_intent_group_name}",
      "userData2":"user data 2 for ${deployment_intent_group_name}"
   },
   "spec":{
      "profile":"${operators_composite_profile_name}",
      "version":"${release}",
      "logical-cloud":"${admin_logical_cloud_name}",
      "override-values":[]
   }
}
EOF
)"


# define the generic-k8s-intent
generick8s_intent_name="generick8s_intent"
generick8s_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${generick8s_intent_name}",
      "description":"descriptionf of ${generick8s_intent_name}",
      "userData1":"user data 1 for ${generick8s_intent_name}",
      "userData2":"user data 2 for ${generick8s_intent_name}"
   }
}
EOF
)"

# define the resource
resource_name="resourceETCD"
appName="${operator_app_name}"
kind="StatefulSet"
k8sResourceName="etcd"
resource_data="$(cat <<EOF
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
EOF
)"


# define the customization
customization_name="customizeETCD"
#customization_data_file="/opt/kud/multi-cluster/cluster1/artifacts/sensor.json"
#customization_data_file2="/opt/kud/multi-cluster/cluster1/artifacts/sensor2.json"

customization_data="$(cat <<EOF
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
EOF
)"

# define the intents to be used by the group
deployment_intents_in_group_name="operators_deploy_intents"
deployment_intents_in_group_data="$(cat <<EOF
{
   "metadata":{
      "name":"${deployment_intents_in_group_name}",
      "description":"descriptionf of ${deployment_intents_in_group_name}",
      "userData1":"user data 1 for ${deployment_intents_in_group_name}",
      "userData2":"user data 2 for ${deployment_intents_in_group_name}"
   },
   "spec":{
      "intent":{
         "genericPlacementIntent":"${generic_placement_intent_name}",
         "genericaction":"${generick8s_intent_name}"
      }
   }
}
EOF
)"


#---------END: SET ORCH DATA--------------------


function createLogicalCloudData {
   print_msg "Creating logical cloud ${admin_logical_cloud_name}"
    call_api -d "${admin_logical_cloud_data}" "${base_url_dcm}/projects/${projectname}/logical-clouds"
    call_api -d "${cluster_1_data}" "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}/cluster-references"
}


function createGenericActionData {
   print_msg "BEGIN :: createGenericActionData"
   call_api -d "${generick8s_intent_data}" \
             "${base_url_gac}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-k8s-intents"

   print_msg "BEGIN :: create resource_data"
   call_api -H "Content-Type: multipart/form-data" -F "metadata=$resource_data" "${base_url_gac}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-k8s-intents/${generick8s_intent_name}/resources"

   print_msg "BEGIN :: create cutomization_data"
   call_api -H "Content-Type: multipart/form-data" -F "metadata=$customization_data" \
             "${base_url_gac}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-k8s-intents/${generick8s_intent_name}/resources/${resource_name}/customizations"

   print_msg "END :: createGenericActionData"

}

function instantiateLogicalCloud {
   call_api -d "{ }" "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}/instantiate"
}

function createOrchestratorData {

    print_msg "creating controller entries"
    call_api -d "${rsynccontrollerdata}" "${base_url_orchestrator}/controllers"

    print_msg "creating controller entries"
    call_api -d "${genericactioncontrollerdata}" "${base_url_orchestrator}/controllers"

    print_msg "creating project entry"
    call_api -d "${projectdata}" "${base_url_orchestrator}/projects"

   createLogicalCloudData
   instantiateLogicalCloud


    print_msg "creating operators composite app entry"
    call_api -d "${compositeapp_data}" "${base_url_orchestrator}/projects/${projectname}/composite-apps"

    print_msg "adding operator sub-app to the composite app"
    call_api -F "metadata=${operator_app_data}" \
             -F "file=@${operator_helm_chart}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/apps"


    print_msg "creating operators composite profile entry"
    call_api -d "${operators_composite_profile_data}" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/composite-profiles"

    print_msg "adding operator sub-app profile to the composite profile"
    call_api -F "metadata=${operator_profile_data}" \
             -F "file=@${operator_profile_file}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/composite-profiles/${operators_composite_profile_name}/profiles"

    print_msg "create the deployment intent group"
    call_api -d "${deployment_intent_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups"
    call_api -d "${deployment_intents_in_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents"

   createGenericActionData

   print_msg "create the generic placement intent"
    call_api -d "${generic_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents"
    print_msg "add the operator app placement intent to the generic placement intent"
    call_api -d "${operator_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents"

}


function deleteGenericActionData {
   print_msg "BEGIN :: deleteGenericActionData"
   delete_resource "${base_url_gac}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-k8s-intents/${generick8s_intent_name}/resources/${resource_name}/customization/${customization_name}"

   delete_resource "${base_url_gac}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-k8s-intents/${generick8s_intent_name}/resources/${resource_name}"

   delete_resource "${base_url_gac}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-k8s-intents/${generick8s_intent_name}"

   print_msg "END :: deleteGenericActionData"

}

function deleteLogicalCloud {
    delete_resource "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}/cluster-references/${lc_cluster_1_name}"
    delete_resource "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}"
}

function deleteOrchestratorData {



   # TODO- delete rsync controller and any other controller
    delete_resource "${base_url_orchestrator}/controllers/${rsynccontrollername}"
    delete_resource "${base_url_orchestrator}/controllers/${genericactioncontrollername}"


    deleteGenericActionData

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents/${deployment_intents_in_group_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${operator_placement_intent_name}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/composite-profiles/${operators_composite_profile_name}/profiles/${operator_profile_name}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/composite-profiles/${operators_composite_profile_name}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/apps/${operator_app_name}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}"
}

function createClmData {
    print_msg "Creating cluster provider and cluster"
    call_api -d "${clusterproviderdata}" "${base_url_clm}/cluster-providers"
    call_api -H "Content-Type: multipart/form-data" -F "metadata=$clusterdata" -F "file=@$kubeconfigcluster1" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters"
    call_api -d "${labeldata}" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels"


}


function deleteClmData {
   deleteLogicalCloud
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels/${labelname}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}"
}
function createData {
    createClmData
    createOrchestratorData
}
function deleteData {
    deleteClmData
    deleteOrchestratorData
}
function instantiate {
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/approve"
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/instantiate"
}


function terminateOrchData {
    #call_api -d "{  }" "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}/terminate"
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${operators_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/terminate"
    }

# Setup
populate_CSAR_operator_helm "$csar_id"


#terminateOrchData
#sleep 20
#deleteData
createData
instantiate

