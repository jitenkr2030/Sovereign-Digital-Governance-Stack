{{- define "library.configmap" -}}
{{- if .Values.configmap.create }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "library.fullname" . }}
  labels:
    {{- include "library.labels" . | nindent 4 }}
  {{- if or .Values.configmap.annotations .Values.global.labels }}
  annotations:
    {{- if .Values.configmap.annotations }}
    {{- toYaml .Values.configmap.annotations | nindent 4 }}
    {{- end }}
    {{- if .Values.global.labels }}
    {{- toYaml .Values.global.labels | nindent 4 }}
    {{- end }}
  {{- end }}
data:
  {{- range $key, $value := .Values.configmap.data }}
  {{ $key }}: |
    {{- if kindIs "string" $value }}
      {{- $value | nindent 4 }}
    {{- else }}
      {{- toYaml $value | nindent 4 }}
    {{- end }}
  {{- end }}
{{- end }}
{{- end }}

{{- define "library.configmapFromFile" -}}
{{- if .Values.configmapFromFile.create }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "library.fullname" . }}-files
  labels:
    {{- include "library.labels" . | nindent 4 }}
  {{- if or .Values.configmapFromFile.annotations .Values.global.labels }}
  annotations:
    {{- if .Values.configmapFromFile.annotations }}
    {{- toYaml .Values.configmapFromFile.annotations | nindent 4 }}
    {{- end }}
    {{- if .Values.global.labels }}
    {{- toYaml .Values.global.labels | nindent 4 }}
    {{- end }}
  {{- end }}
binaryData:
  {{- range $key, $value := .Values.configmapFromFile.binaryData }}
  {{ $key }}: {{ $value | b64enc | quote }}
  {{- end }}
data:
  {{- range $key, $value := .Values.configmapFromFile.data }}
  {{ $key }}: |
    {{- $value | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}
