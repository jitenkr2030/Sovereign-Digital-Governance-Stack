{{- if .Values.serviceAccount.create -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "library.serviceAccountName" . }}
  labels:
    {{- include "library.labels" . | nindent 4 }}
  {{- if or .Values.serviceAccount.annotations .Values.global.labels }}
  annotations:
    {{- if .Values.serviceAccount.annotations }}
    {{- toYaml .Values.serviceAccount.annotations | nindent 4 }}
    {{- end }}
    {{- if .Values.global.labels }}
    {{- toYaml .Values.global.labels | nindent 4 }}
    {{- end }}
  {{- end }}
{{- end }}
