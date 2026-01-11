{{- define "library.service" -}}
{{- if .Values.service.enabled }}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "library.fullname" . }}
  labels:
    {{- include "library.labels" . | nindent 4 }}
  {{- if or .Values.service.annotations .Values.global.labels }}
  annotations:
    {{- if .Values.service.annotations }}
    {{- toYaml .Values.service.annotations | nindent 4 }}
    {{- end }}
    {{- if .Values.global.labels }}
    {{- toYaml .Values.global.labels | nindent 4 }}
    {{- end }}
  {{- end }}
spec:
  type: {{ .Values.service.type }}
  {{- if .Values.service.clusterIP }}
  clusterIP: {{ .Values.service.clusterIP }}
  {{- end }}
  {{- if .Values.service.sessionAffinity }}
  sessionAffinity: {{ .Values.service.sessionAffinity }}
  {{- end }}
  {{- if .Values.service.sessionAffinityConfig }}
  sessionAffinityConfig:
    {{- toYaml .Values.service.sessionAffinityConfig | nindent 4 }}
  {{- end }}
  {{- if .Values.service.externalTrafficPolicy }}
  externalTrafficPolicy: {{ .Values.service.externalTrafficPolicy }}
  {{- end }}
  {{- if .Values.service.healthCheckNodePort }}
  healthCheckNodePort: {{ .Values.service.healthCheckNodePort }}
  {{- end }}
  ports:
    {{- range $key, $value := .Values.containerPorts }}
    - name: {{ $key }}
      port: {{ $value }}
      targetPort: {{ $value }}
      {{- if $.Values.service.portMappings }}
      {{- $portMapping := index $.Values.service.portMappings $key | default (dict) }}
      {{- if $portMapping.nodePort }}
      nodePort: {{ $portMapping.nodePort }}
      {{- end }}
      {{- if $portMapping.protocol }}
      protocol: {{ $portMapping.protocol }}
      {{- end }}
      {{- end }}
    {{- end }}
  selector:
    {{- include "library.selectorLabels" . | nindent 4 }}
{{- end }}
{{- end }}

{{- define "library.serviceHeadless" -}}
{{- if .Values.serviceHeadless.enabled }}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "library.fullname" . }}-headless
  labels:
    {{- include "library.labels" . | nindent 4 }}
  {{- if .Values.serviceHeadless.annotations }}
  annotations:
    {{- toYaml .Values.serviceHeadless.annotations | nindent 4 }}
  {{- end }}
spec:
  type: ClusterIP
  clusterIP: None
  ports:
    {{- range $key, $value := .Values.containerPorts }}
    - name: {{ $key }}
      port: {{ $value }}
      targetPort: {{ $value }}
    {{- end }}
  selector:
    {{- include "library.selectorLabels" . | nindent 4 }}
{{- end }}
{{- end }}
