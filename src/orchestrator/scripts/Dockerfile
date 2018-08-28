# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

FROM ubuntu:18.04

ARG HTTP_PROXY=${HTTP_PROXY}
ARG HTTPS_PROXY=${HTTPS_PROXY}

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY
ENV no_proxy $NO_PROXY

EXPOSE 9015

RUN groupadd -r onap && useradd -r -g onap onap

WORKDIR /opt/multicloud/k8s/orchestrator
RUN chown onap:onap /opt/multicloud/k8s/orchestrator -R

ADD --chown=onap ./orchestrator ./

USER onap

CMD ["./orchestrator"]