{{- define "library.podDisruptionBudget" -}}
{{- if and .Values.podDisruptionBudget.enabled (semverCompare ">=1.21-0" .Capabilities.KubeVersion.GitVersion) }}
apiVersion: policy/v1
{{- else if semverCompare ">=1.5-0" .Capabilities.KubeVersion.GitVersion }}
apiVersion: policy/v1beta1
{{- else }}
apiVersion: extensions/v1beta1
{{- end }}
kind: PodDisruptionBudget
metadata:
  name: {{ include "library.fullname" . }}
  labels:
    {{- include "library.labels" . | nindent 4 }}
  {{- if or .Values.podDisruptionBudget.annotations .Values.global.labels }}
  annotations:
    {{- if .Values.podDisruptionBudget.annotations }}
    {{- toYaml .Values.podDisruptionBudget.annotations | nindent 4 }}
    {{- end }}
    {{- if .Values.global.labels }}
    {{- toYaml .Values.global.labels | nindent 4 }}
    {{- end }}
  {{- end }}
spec:
  {{- if .Values.podDisruptionBudget.minAvailable }}
  minAvailable: {{ .Values.podDisruptionBudget.minAvailable }}
  {{- end }}
  {{- if .Values.podDisruptionBudget.maxUnavailable }}
  maxUnavailable: {{ .Values.podDisruptionBudget.maxUnavailable }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "library.selectorLabels" . | nindent 6 }}
{{- end }}
{{- end }}
