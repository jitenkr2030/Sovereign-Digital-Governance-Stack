{{- define "library.horizontalPodAutoscaler" -}}
{{- if .Values.autoscaling.enabled }}
{{- if semverCompare ">=1.18-0" .Capabilities.KubeVersion.GitVersion }}
apiVersion: autoscaling/v2
{{- else }}
apiVersion: autoscaling/v2beta2
{{- end }}
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "library.fullname" . }}
  labels:
    {{- include "library.labels" . | nindent 4 }}
  {{- if or .Values.autoscaling.annotations .Values.global.labels }}
  annotations:
    {{- if .Values.autoscaling.annotations }}
    {{- toYaml .Values.autoscaling.annotations | nindent 4 }}
    {{- end }}
    {{- if .Values.global.labels }}
    {{- toYaml .Values.global.labels | nindent 4 }}
    {{- end }}
  {{- end }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "library.fullname" . }}
  {{- if .Values.autoscaling.minReplicas }}
  minReplicas: {{ .Values.autoscaling.minReplicas }}
  {{- end }}
  {{- if .Values.autoscaling.maxReplicas }}
  maxReplicas: {{ .Values.autoscaling.maxReplicas }}
  {{- end }}
  metrics:
    {{- toYaml .Values.autoscaling.metrics | nindent 4 }}
  {{- if .Values.autoscaling.behavior }}
  behavior:
    {{- toYaml .Values.autoscaling.behavior | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}
