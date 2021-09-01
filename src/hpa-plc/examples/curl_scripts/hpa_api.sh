# // Copyright (c) 2020 Intel Corporation

#!/bin/bash

# change the following according to environment:
hpa_placement_addr="http://localhost:9091"

description="test-description"
description_2="test-description-2"
project="proj1"
composite_app="collection-composite-app"
di_group="collection-deployment-intent-group"
app_nam1_1="http-client"
hpa_intent_name_1="hpa-intent-1"

# hpa consumer data
hpa_consumer_name_1="hpa-consumer-1"
k8s_api_version="1.19"
k8s_kind="Deployment"
k8s_res_name="HttpClient"
k8s_container_name="HttpClientContainer"

# hpa resource data
hpa_alloc_resource_name_1="hpa-alloc-resource-1"
hpa_non_alloc_resource_name_1="hpa-non-alloc-resource-1"
hpa_allocatable=false
hpa_mandatory=true
hpa_weight=1
#hpa_resource_info_1="{"key":"gpu", "value":"yes"}"
hpa_resource_info_1="{"key":"nvidia.com/cuda.runtime.major", "value":"10"}"

# endpoints
hpa_intent_url="$hpa_placement_addr/v2/projects/${project}/composite-apps/${composite_app}/v1/deployment-intent-groups/${di_group}/hpa-intents"
hpa_consumer_url="$hpa_placement_addr/v2/projects/${project}/composite-apps/${composite_app}/v1/deployment-intent-groups/${di_group}/hpa-intents/${hpa_intent_name_1}/hpa-resource-consumers"
hpa_resource_url="$hpa_placement_addr/v2/projects/${project}/composite-apps/${composite_app}/v1/deployment-intent-groups/${di_group}/hpa-intents/${hpa_intent_name_1}/hpa-resource-consumers/${hpa_consumer_name_1}/resource-requirements"

hpa_bad_intent_url="$hpa_placement_addr/v2/projects/${project}/composite-apps/collection-composite-app2/v1/deployment-intent-groups/${di_group}/hpa-intents"

# data
hpa_intent_data="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_intent_name_1}",
    "description": "${description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "app-name" : "${app_nam1_1}"
 }
}
EOF
)"

hpa_intent_update_data="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_intent_name_1}",
    "description": "${description_2}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "app-name" : "${app_nam1_1}"
 }
}
EOF
)"

hpa_consumer_data="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_consumer_name_1}",
    "description": "${description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "api-version" : "${k8s_api_version}",
    "kind" : "${k8s_kind}",
    "name" : "${k8s_res_name}",
    "container-name" : "${k8s_container_name}"
 }
}
EOF
)"

hpa_consumer_update_data="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_consumer_name_1}",
    "description": "${description_2}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "api-version" : "${k8s_api_version}",
    "kind" : "${k8s_kind}",
    "name" : "${k8s_res_name}",
    "container-name" : "${k8s_container_name}"
 }
}
EOF
)"

hpa_non_alloc_resource_data="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_non_alloc_resource_name_1}",
    "description": "${description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "allocatable" : false,
    "mandatory" : true,
    "weight" : 1,
    "resource" : {"key":"nvidia.com/cuda.runtime.major", "value":"10"}
 }
}
EOF
)"

hpa_non_alloc_resource_update_data="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_non_alloc_resource_name_1}",
    "description": "${description_2}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "allocatable" : false,
    "mandatory" : true,
    "weight" : 1,
    "resource" : {"key":"nvidia.com/cuda.runtime.major/updated", "value":"10"}
 }
}
EOF
)"

hpa_non_alloc_positive="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_non_alloc_resource_name_1}",
    "description": "${description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "allocatable" : false,
    "mandatory" : true,
    "weight" : 1,
    "resource" : {"key":"nvidia.com/cuda.runtime.major", "value":"10"}
 }
}
EOF
)"

hpa_non_alloc_negative="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_non_alloc_resource_name_1}",
    "description": "${description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "allocatable" : false,
    "mandatory" : true,
    "weight" : 1,
    "resource" : {"key":"nvidia.com/cuda.runtime.major1", "value":"10"}
 }
}
EOF
)"

hpa_alloc_resource_data="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_alloc_resource_name_1}",
    "description": "${description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "allocatable" : true,
    "mandatory" : true,
    "weight" : 1,
    "resource" : {"name":"cpu", "requests":10,"limits":10}
 }
}
EOF
)"

hpa_alloc_resource_update_data="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_alloc_resource_name_1}",
    "description": "${description_2}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "allocatable" : true,
    "mandatory" : true,
    "weight" : 1,
    "resource" : {"name":"cpu", "requests":80,"limits":80}
 }
}
EOF
)"

hpa_alloc_positive="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_alloc_resource_name_1}",
    "description": "${description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "allocatable" : true,
    "mandatory" : true,
    "weight" : 1,
    "resource" : {"name":"cpu", "requests":1,"limits":1}
 }
}
EOF
)"

hpa_alloc_negative="$(cat << EOF
{
 "metadata" : {
    "name": "${hpa_alloc_resource_name_1}",
    "description": "${description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "allocatable" : true,
    "mandatory" : true,
    "weight" : 1,
    "resource" : {"name":"cpu", "requests":50,"limits":50}
 }
}
EOF
)"

echo .
echo -e "\n\n EXECUTE HPA CURL SCRIPTS .. START \n\n"

if [ "$1" == "create" ]; then

   printf "\n\n====================================================\n\n"
   printf "\nCreating hpa-intent...\n\n"
   curl -d "${hpa_intent_data}" -X POST ${hpa_intent_url}

   printf "\nGetting hpa-intent...\n\n"
   curl -X GET ${hpa_intent_url}/${hpa_intent_name_1}

   printf "\nGetting hpa-intent by query...\n\n"
   curl -X QUERY ${hpa_intent_url}?intent=${hpa_intent_name_1}

   printf "\n\n====================================================\n\n"
   printf "\nCreating hpa-consumer...\n\n"
   curl -d "${hpa_consumer_data}" -X POST ${hpa_consumer_url}

   printf "\nGetting hpa-consumer...\n\n"
   curl -X GET ${hpa_consumer_url}/${hpa_consumer_name_1}

   printf "\nGetting hpa-consumer by query...\n\n"
   curl -X QUERY ${hpa_consumer_url}?consumer=${hpa_consumer_name_1}

   printf "\n\n====================================================\n\n"
   printf "\nCreating non-allocatable hpa-resource...\n\n"
   curl -d "${hpa_non_alloc_resource_data}" -X POST ${hpa_resource_url}

   printf "\nGetting non-allocatable hpa-resource...\n\n"
   curl -X GET ${hpa_resource_url}/${hpa_non_alloc_resource_name_1}

   printf "\nGetting non-allocatable hpa-resource by query...\n\n"
   curl -X QUERY ${hpa_resource_url}?resource=${hpa_non_alloc_resource_name_1}

   printf "\n\n====================================================\n\n"
   printf "\nCreating allocatable hpa-resource...\n\n"
   curl -d "${hpa_alloc_resource_data}" -X POST ${hpa_resource_url}

   printf "\nGetting allocatable hpa-resource...\n\n"
   curl -X GET ${hpa_resource_url}/${hpa_alloc_resource_name_1}

   printf "\nGetting allocatable hpa-resource by query...\n\n"
   curl -X QUERY ${hpa_resource_url}?resource=${hpa_alloc_resource_name_1}

   printf "\n\n====================================================\n\n"

elif [ "$1" == "create-negative" ]; then

   printf "\n\n====================================================\n\n"
   printf "\nCreating hpa-intent...\n\n"
   curl -d "${hpa_intent_data}" -X POST ${hpa_intent_url}

   printf "\nGetting hpa-intent...\n\n"
   curl -X GET ${hpa_intent_url}/${hpa_intent_name_1}

   printf "\nCreating hpa-intent with wrong compositeapp name ...\n\n"
   curl -d "${hpa_intent_data}" -X POST ${hpa_bad_intent_url}

   printf "\nGetting hpa-intent...\n\n"
   curl -X GET ${hpa_intent_url}/${hpa_bad_intent_url}

elif [ "$1" == "update" ]; then

   printf "\n\n====================================================\n\n"
   printf "\nGetting hpa-intent...\n\n"
   curl -X GET ${hpa_intent_url}/${hpa_intent_name_1}

   printf "\nUpdating hpa-intent...\n\n"
   curl -d "${hpa_intent_update_data}" -X PUT ${hpa_intent_url}/${hpa_intent_name_1}

   printf "\nGetting hpa-intent...\n\n"
   curl -X GET ${hpa_intent_url}/${hpa_intent_name_1}

   printf "\n\n====================================================\n\n"
   printf "\nGetting hpa-consumer...\n\n"
   curl -X GET ${hpa_consumer_url}/${hpa_consumer_name_1}

   printf "\nUpdating hpa-consumer...\n\n"
   curl -d "${hpa_consumer_update_data}" -X PUT ${hpa_consumer_url}/${hpa_consumer_name_1}

   printf "\nGetting hpa-consumer...\n\n"
   curl -X GET ${hpa_consumer_url}/${hpa_consumer_name_1}

   printf "\n\n====================================================\n\n"
   printf "\nGetting non-allocatable hpa-resource...\n\n"
   curl -X GET ${hpa_resource_url}/${hpa_non_alloc_resource_name_1}

   printf "\nUpdating non-allocatable hpa-resource...\n\n"
   curl -d "${hpa_non_alloc_resource_update_data}" -X PUT ${hpa_resource_url}/${hpa_non_alloc_resource_name_1}

   printf "\nGetting non-allocatable hpa-resource...\n\n"
   curl -X GET ${hpa_resource_url}/${hpa_non_alloc_resource_name_1}

   printf "\n\n====================================================\n\n"
   printf "\nGetting allocatable hpa-resource...\n\n"
   curl -X GET ${hpa_resource_url}/${hpa_alloc_resource_name_1}

   printf "\nUpdating allocatable hpa-resource...\n\n"
   curl -d "${hpa_alloc_resource_update_data}" -X PUT ${hpa_resource_url}/${hpa_alloc_resource_name_1}

   printf "\nGetting allocatable hpa-resource...\n\n"
   curl -X GET ${hpa_resource_url}/${hpa_alloc_resource_name_1}

   printf "\n\n====================================================\n\n"

elif [ "$1" == "getall" ]; then
   printf "\n\n====================================================\n\n"
   printf "\nGetting all hpa-intent...\n\n"
   curl -X GET ${hpa_intent_url}

   printf "\nGetting all hpa-consumer...\n\n"
   curl -X GET ${hpa_consumer_url}

   printf "\nGetting all hpa-resource...\n\n"
   curl -X GET ${hpa_resource_url}

   printf "\n\n====================================================\n\n"

elif [ "$1" == "delete" ]; then
   printf "\n\n====================================================\n\n"

   printf "\nDeleting non-alloc hpa-resource...\n\n"
   curl -X DELETE ${hpa_resource_url}/${hpa_non_alloc_resource_name_1}

   printf "\nDeleting alloc hpa-resource...\n\n"
   curl -X DELETE ${hpa_resource_url}/${hpa_alloc_resource_name_1}

   printf "\nDeleting hpa-consumer...\n\n"
   curl -X DELETE ${hpa_consumer_url}/${hpa_consumer_name_1}

   printf "\nDeleting hpa-intent...\n\n"
   curl -X DELETE ${hpa_intent_url}/${hpa_intent_name_1}
   
   printf "\nGetting all hpa-intent...\n\n"
   curl -X GET ${hpa_intent_url}

   printf "\nGetting all hpa-consumer...\n\n"
   curl -X GET ${hpa_consumer_url}

   printf "\nGetting all hpa-resource...\n\n"
   curl -X GET ${hpa_resource_url}

   printf "\n\n====================================================\n\n"

elif [ "$1" == "deleteall" ]; then
   printf "\n\n====================================================\n\n"

   printf "\nDeleting all hpa-resources...\n\n"
   curl -X DELETE ${hpa_resource_url}

   printf "\nDeleting all hpa-consumers...\n\n"
   curl -X DELETE ${hpa_consumer_url}

   printf "\nDeleting all hpa-intents...\n\n"
   curl -X DELETE ${hpa_intent_url}
   
   printf "\nGetting all hpa-intent...\n\n"
   curl -X GET ${hpa_intent_url}

   printf "\nGetting all hpa-consumer...\n\n"
   curl -X GET ${hpa_consumer_url}

   printf "\nGetting all hpa-resource...\n\n"
   curl -X GET ${hpa_resource_url}

   printf "\n\n====================================================\n\n"

else
   printf "\nUnknown input...\n\n"
fi

echo .
echo -e "\n\n EXECUTE HPA CURL SCRIPTS .. END \n\n"
