{{/*
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation
*/}}
{{- define "common.servicemco" -}}
{{- $common := dict "Values" .Values.common -}}
{{- $noCommon := omit .Values "common" -}}
{{- $overrides := dict "Values" $noCommon -}}
{{- $noValues := omit . "Values" -}}
{{- with merge $noValues $overrides $common -}}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "common.servicename" . }}
  namespace: {{ include "common.namespace" . }}
  labels:
    app: {{ include "common.fullname" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
spec:
  type: {{ .Values.service.type }}
  ports:
  - name: {{ .Values.service.PortName }}
    {{if eq .Values.service.type "NodePort" -}}
    port: {{ .Values.service.internalPort }}
    nodePort: {{ .Values.global.nodePortPrefixExt | default "302" }}{{ .Values.service.nodePort }}
    {{- else -}}
    port: {{ .Values.service.externalPort }}
    targetPort: {{ .Values.service.internalPort }}
    {{- end}}
    protocol: TCP
  selector:
    app: {{ include "common.name" . }}
    release: {{ .Release.Name }}
{{- end -}}
{{- end -}}