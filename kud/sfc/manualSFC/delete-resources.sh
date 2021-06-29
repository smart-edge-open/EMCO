#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

function delay {
	echo "waiting for $1 $2"
	
	for i in $(eval echo {$1..1})
	do
		echo -en "\r$i    "
		sleep 1
	done
	echo "    "
}


kubectl delete -f sfc-with-virtual-and-provider-network.yaml
delay 5 "before deleting nginx-right"
kubectl delete -f nginx-right-deployment.yaml
delay 5 "before deleting nginx-left"
kubectl delete -f nginx-left-deployment.yaml
delay 5 "before deleting nginx-sdewan"
kubectl delete -f sdewan-multiple-network.yaml
delay 5 "before deleting ngfw"
kubectl delete -f ngfw.yaml 
delay 5 "before deleting slb"
kubectl delete -f slb-multiple-network.yaml 
delay 5 "before deleting networks"
kubectl delete -f sfc-virtual-network.yaml 
delay 5 "before deleting left ns"
kubectl delete -f namespace-left.yaml
kubectl delete -f namespace-right.yaml
