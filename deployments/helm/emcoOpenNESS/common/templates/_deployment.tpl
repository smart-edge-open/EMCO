{{/*
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation
*/}}
{{- define "common.deployment" -}}
{{- $common := dict "Values" .Values.common -}}
{{- $noCommon := omit .Values "common" -}}
{{- $overrides := dict "Values" $noCommon -}}
{{- $noValues := omit . "Values" -}}
{{- with merge $noValues $overrides $common -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ include "common.namespace" . }}
  labels:
    app: {{ include "common.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
spec:
  selector:
    matchLabels:
      app: {{ include "common.name" . }}
      release: {{ .Release.Name }}
  replicas: {{ .Values.replicaCount }}
  template:
    metadata:
      labels:
        app: {{ include "common.name" . }}
        release: {{ .Release.Name }}
    spec:
      containers:
      - image: "{{ include "common.repository" . }}{{ .Values.image }}:{{ .Values.imageTag }}"
        imagePullPolicy: {{ .Values.global.pullPolicy | default .Values.pullPolicy }}
        name: {{ include "common.name" . }}
        env:
        {{- $userProxy := .Values | default dict }}
        {{- if $userProxy.noProxyHosts }}
        - name: NO_PROXY
          value: {{ $userProxy.noProxyHosts }}
        - name: no_proxy
          value: {{ $userProxy.noProxyHosts }}
        {{- end}}
        {{- if $userProxy.httpProxy }}
        - name: HTTP_PROXY
          value: {{ .Values.httpProxy }}
        - name: http_proxy
          value: {{ .Values.httpProxy }}
        {{- end}}
        {{- if $userProxy.httpsProxy }}
        - name: HTTPS_PROXY
          value: {{ .Values.httpsProxy }}
        - name: https_proxy
          value: {{ .Values.httpsProxy }}
        {{- end}}
        {{- if eq (empty .Values.global.disableDbAuth) true }}
        - name: DB_EMCO_USERNAME
          value: emco
        - name: DB_EMCO_PASSWORD
          valueFrom:
            secretKeyRef:
              name: emco-mongo
              key: userPassword
        - name: CONTEXTDB_EMCO_USERNAME
          value: "root"
        - name: CONTEXTDB_EMCO_PASSWORD
          valueFrom:
            secretKeyRef:
              name: emco-etcd
              key: etcd-root-password
        {{- end }}
        command: [{{ .Values.command }}]
        args: [{{ .Values.args }}]
        workingDir: {{ .Values.workingDir }}
        ports:
        - containerPort: {{ .Values.service.internalPort }}
        {{- $si := .Values.serviceInternal | default dict }}
        {{- if $si.internalPort }}
        - containerPort: {{ $si.internalPort }}
        {{- end }}
        {{- if eq .Values.liveness.enabled true }}
        livenessProbe:
          tcpSocket:
            port: {{ .Values.service.internalPort }}
          initialDelaySeconds: {{ .Values.liveness.initialDelaySeconds }}
          periodSeconds: {{ .Values.liveness.periodSeconds }}
        {{ end }}

        readinessProbe:
          tcpSocket:
            port: {{ .Values.service.internalPort }}
          initialDelaySeconds: {{ .Values.readiness.initialDelaySeconds }}
          periodSeconds: {{ .Values.readiness.periodSeconds }}
        volumeMounts:
          - mountPath: /etc/localtime
            name: localtime
            readOnly: true
          - mountPath: {{ .Values.workingDir }}/config.json
            name: {{ include "common.name" .}}
            subPath: config.json
        resources:
{{ include "common.resources" .  }}
        {{- if .Values.nodeSelector }}
        nodeSelector:
{{ toYaml .Values.nodeSelector  }}
        {{- end -}}
        {{- if .Values.affinity }}
        affinity:
{{ toYaml .Values.affinity  }}
        {{- end }}
      volumes:
      - name: localtime
        hostPath:
          path: /etc/localtime
      - name : {{ include "common.name" . }}
        configMap:
          name: {{ include "common.fullname" . }}
      imagePullSecrets:
      - name: "{{ include "common.namespace" . }}-docker-registry-key"
{{- end -}}
{{- end -}}
