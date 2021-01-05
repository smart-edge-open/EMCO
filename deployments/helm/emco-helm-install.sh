#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation
#-------- DO NOT EDIT THIS FILE ---------#
usage() {
	echo "USAGE:" $0 "<optional: -s parameter=value> <optional: -k <path to kubeconfig file>> [install | uninstall]"
	echo "NOTES: multiple instances of <-s parameter=value> may be provided"
}

EMCO_HELM_FILE="emco-helm.tgz"

install() {
	echo "Creating namespace emco"
	kubectl ${KUBCFG} create ns emco
	echo "Installing EMCO. Please wait..."
	helm ${KUBCFG} install --namespace=emco -f helm_value_overrides.yaml ${HELMSETVALUES} --wait emco ${EMCO_HELM_FILE}
	if [ "$?" -ne "0" ]; then
	    echo "Deleting namespace emco"
		kubectl ${KUBCFG} delete ns emco
	fi
	echo "Done"
}

uninstall() {
	echo "Removing EMCO..."
	helm ${KUBCFG} uninstall --namespace=emco emco
	echo "Deleting namespace emco"
	kubectl ${KUBCFG} delete ns emco
	echo "Done"
}

KUBCFG=""
HELMSETVALUES=""
while getopts "hs:k:" opt; do
	case $opt in
	s)
		HELMSETVALUES=${HELMSETVALUES}" --set $OPTARG"
		;;
	k)
		KUBCFG="--kubeconfig="$OPTARG
		;;
	h)
		usage
		exit 0
		;;
	\?)
		echo "Invalid option: -$OPTARG" >&2
		usage
		exit 1
		;;
	:)
		echo "Option -$OPTARG requires an argument." >&2
		usage
		exit 1
		;;
	esac
done

shift $((OPTIND -1))

if [ "$1" = "install" ]; then
	install
elif [ "$1" = "uninstall" ]; then
	uninstall
else
	echo "Not a valid command: "$2
	exit 2
fi
exit 0
