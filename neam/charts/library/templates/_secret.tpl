{{- define "library.secret" -}}
{{- if .Values.secret.create }}
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "library.fullname" . }}
  labels:
    {{- include "library.labels" . | nindent 4 }}
  {{- if or .Values.secret.annotations .Values.global.labels }}
  annotations:
    {{- if .Values.secret.annotations }}
    {{- toYaml .Values.secret.annotations | nindent 4 }}
    {{- end }}
    {{- if .Values.global.labels }}
    {{- toYaml .Values.global.labels | nindent 4 }}
    {{- end }}
  {{- end }}
type: {{ .Values.secret.type }}
data:
  {{- range $key, $value := .Values.secret.data }}
  {{ $key }}: {{ $value | b64enc | quote }}
  {{- end }}
  {{- if .Values.secret.stringData }}
stringData:
  {{- range $key, $value := .Values.secret.stringData }}
  {{ $key }}: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- end }}
{{- end }}

{{- define "library.tlsSecret" -}}
{{- if .Values.tlsSecret.create }}
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "library.fullname" . }}-tls
  labels:
    {{- include "library.labels" . | nindent 4 }}
  {{- if .Values.tlsSecret.annotations }}
  annotations:
    {{- toYaml .Values.tlsSecret.annotations | nindent 4 }}
  {{- end }}
type: kubernetes.io/tls
data:
  tls.crt: {{ .Values.tlsSecret.crt | b64enc | quote }}
  tls.key: {{ .Values.tlsSecret.key | b64enc | quote }}
{{- end }}
{{- end }}
