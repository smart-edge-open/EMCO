#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

export ovnport=`ip a | grep ': ovn4nfv0' | cut -d ':' -f 2`
kubectl -n kube-system exec $(kubectl get pods -lapp=ovn-controller -n kube-system  --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') -it -- ovs-vsctl del-port br-int $ovnport
sleep 5
kubectl -n kube-system exec $(kubectl get pods -lapp=ovn-control-plane -n kube-system  --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') -it -- ovn-nbctl lsp-del $ovnport
sleep 5

kubectl delete -f /home/vagrant/git/icn-nodus/deploy/ovn4nfv-k8s-plugin.yaml
sleep 60
# wait for above to completely terminate
kubectl delete -f /home/vagrant/git/icn-nodus/deploy/ovn-daemonset.yaml
# wait for above to completely terminate

sleep 60
# after above two are completely terminated, re-apply them
kubectl apply -f /home/vagrant/git/icn-nodus/deploy/ovn-daemonset.yaml
# wait for ovn to come up
sleep 30
kubectl apply -f /home/vagrant/git/icn-nodus/deploy/ovn4nfv-k8s-plugin.yaml

