#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

source _functions.sh
k8s_path="$(git rev-parse --show-toplevel)"
#
# Start from compiled binaries to foreground. This is usable for development use.
#
source /etc/environment
opath="$(git rev-parse --show-toplevel)"/src/orchestrator

stop_all
start_mongo
start_etcd

echo "Compiling source code"
pushd $opath
generate_config
make all
./orchestrator
popd
