{{- define "library.persistentVolumeClaim" -}}
{{- if .Values.persistentVolumeClaim.create }}
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ include "library.fullname" . }}
  labels:
    {{- include "library.labels" . | nindent 4 }}
  {{- if or .Values.persistentVolumeClaim.annotations .Values.global.labels }}
  annotations:
    {{- if .Values.persistentVolumeClaim.annotations }}
    {{- toYaml .Values.persistentVolumeClaim.annotations | nindent 4 }}
    {{- end }}
    {{- if .Values.global.labels }}
    {{- toYaml .Values.global.labels | nindent 4 }}
    {{- end }}
  {{- end }}
spec:
  accessModes:
    {{- toYaml .Values.persistentVolumeClaim.accessModes | nindent 4 }}
  resources:
    requests:
      storage: {{ .Values.persistentVolumeClaim.size }}
  {{- if .Values.persistentVolumeClaim.storageClassName }}
  storageClassName: {{ .Values.persistentVolumeClaim.storageClassName }}
  {{- end }}
  {{- if .Values.persistentVolumeClaim.selector }}
  selector:
    {{- toYaml .Values.persistentVolumeClaim.selector | nindent 4 }}
  {{- end }}
  {{- if .Values.persistentVolumeClaim.dataSource }}
  dataSource:
    {{- toYaml .Values.persistentVolumeClaim.dataSource | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}
