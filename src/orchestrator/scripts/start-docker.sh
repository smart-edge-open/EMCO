#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

source _functions.sh

#
# Start from containers. build.sh should be run prior this script.
#
stop_all
start_mongo
start_etcd
generate_config
start_all