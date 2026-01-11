{{- define "library.deployment" -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "library.fullname" . }}
  labels:
    {{- include "library.labels" . | nindent 4 }}
  {{- if or .Values.deployment.annotations .Values.global.labels }}
  annotations:
    {{- if .Values.deployment.annotations }}
    {{- toYaml .Values.deployment.annotations | nindent 4 }}
    {{- end }}
    {{- if .Values.global.labels }}
    {{- toYaml .Values.global.labels | nindent 4 }}
    {{- end }}
  {{- end }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "library.selectorLabels" . | nindent 6 }}
  {{- if .Values.deployment.strategy }}
  strategy:
    {{- toYaml .Values.deployment.strategy | nindent 4 }}
  {{- end }}
  template:
    metadata:
      {{- include "library.podMetadata" . | nindent 6 }}
    spec:
      {{- include "library.imagePullSecrets" . | nindent 6 }}
      serviceAccountName: {{ include "library.serviceAccountName" . }}
      {{- if .Values.podSecurityContext.enabled }}
      securityContext:
        runAsUser: {{ .Values.podSecurityContext.runAsUser }}
        runAsGroup: {{ .Values.podSecurityContext.runAsGroup }}
        fsGroup: {{ .Values.podSecurityContext.fsGroup }}
        {{- if .Values.podSecurityContext.supplementalGroups }}
        supplementalGroups:
          {{- toYaml .Values.podSecurityContext.supplementalGroups | nindent 8 }}
        {{- end }}
        {{- if .Values.podSecurityContext.seccompProfile }}
        seccompProfile:
          {{- toYaml .Values.podSecurityContext.seccompProfile | nindent 8 }}
        {{- end }}
      {{- end }}
      {{- if .Values.topologySpreadConstraints }}
      topologySpreadConstraints:
        {{- toYaml .Values.topologySpreadConstraints | nindent 6 }}
      {{- end }}
      {{- if .Values.terminationGracePeriodSeconds }}
      terminationGracePeriodSeconds: {{ .Values.terminationGracePeriodSeconds }}
      {{- end }}
      {{- if .Values.initContainers }}
      initContainers:
        {{- toYaml .Values.initContainers | nindent 6 }}
      {{- end }}
      containers:
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ include "library.imageTag" . }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          {{- if .Values.command }}
          command:
            {{- toYaml .Values.command | nindent 8 }}
          {{- end }}
          {{- if .Values.args }}
          args:
            {{- toYaml .Values.args | nindent 8 }}
          {{- end }}
          ports:
            {{- range $key, $value := .Values.containerPorts }}
            - name: {{ $key }}
              containerPort: {{ $value }}
              protocol: {{ default "TCP" $.Values.containerPortProtocol }}
            {{- end }}
          {{- if .Values.env }}
          env:
            {{- toYaml .Values.env | nindent 8 }}
          {{- end }}
          {{- if .Values.envFrom }}
          envFrom:
            {{- toYaml .Values.envFrom | nindent 8 }}
          {{- end }}
          {{- if .Values.lifecycle }}
          lifecycle:
            {{- toYaml .Values.lifecycle | nindent 8 }}
          {{- end }}
          {{- if .Values.livenessProbe.enabled }}
          livenessProbe:
            httpGet:
              path: {{ .Values.livenessProbe.path }}
              port: {{ .Values.livenessProbe.port }}
            initialDelaySeconds: {{ .Values.livenessProbe.initialDelaySeconds }}
            periodSeconds: {{ .Values.livenessProbe.periodSeconds }}
            timeoutSeconds: {{ .Values.livenessProbe.timeoutSeconds }}
            failureThreshold: {{ .Values.livenessProbe.failureThreshold }}
            successThreshold: {{ .Values.livenessProbe.successThreshold }}
          {{- end }}
          {{- if .Values.readinessProbe.enabled }}
          readinessProbe:
            httpGet:
              path: {{ .Values.readinessProbe.path }}
              port: {{ .Values.readinessProbe.port }}
            initialDelaySeconds: {{ .Values.readinessProbe.initialDelaySeconds }}
            periodSeconds: {{ .Values.readinessProbe.periodSeconds }}
            timeoutSeconds: {{ .Values.readinessProbe.timeoutSeconds }}
            failureThreshold: {{ .Values.readinessProbe.failureThreshold }}
            successThreshold: {{ .Values.readinessProbe.successThreshold }}
          {{- end }}
          {{- if .Values.startupProbe.enabled }}
          startupProbe:
            httpGet:
              path: {{ .Values.startupProbe.path }}
              port: {{ .Values.startupProbe.port }}
            initialDelaySeconds: {{ .Values.startupProbe.initialDelaySeconds }}
            periodSeconds: {{ .Values.startupProbe.periodSeconds }}
            timeoutSeconds: {{ .Values.startupProbe.timeoutSeconds }}
            failureThreshold: {{ .Values.startupProbe.failureThreshold }}
            {{- if .Values.startupProbe.successThreshold }}
            successThreshold: {{ .Values.startupProbe.successThreshold }}
            {{- end }}
          {{- end }}
          resources:
            {{- toYaml .Values.resources | nindent 10 }}
          {{- if .Values.containerSecurityContext.enabled }}
          securityContext:
            allowPrivilegeEscalation: {{ .Values.containerSecurityContext.allowPrivilegeEscalation }}
            capabilities:
              drop:
                {{- toYaml .Values.containerSecurityContext.capabilities.drop | nindent 12 }}
              {{- if .Values.containerSecurityContext.capabilities.add }}
              add:
                {{- toYaml .Values.containerSecurityContext.capabilities.add | nindent 12 }}
              {{- end }}
            {{- if .Values.containerSecurityContext.procMount }}
            procMount: {{ .Values.containerSecurityContext.procMount }}
            {{- end }}
            readOnlyRootFilesystem: {{ .Values.containerSecurityContext.readOnlyRootFilesystem }}
            runAsGroup: {{ .Values.containerSecurityContext.runAsGroup }}
            runAsNonRoot: {{ .Values.containerSecurityContext.runAsNonRoot }}
            runAsUser: {{ .Values.containerSecurityContext.runAsUser }}
            seccompProfile:
              {{- toYaml .Values.containerSecurityContext.seccompProfile | nindent 12 }}
            {{- if .Values.containerSecurityContext.windowsOptions }}
            windowsOptions:
              {{- toYaml .Values.containerSecurityContext.windowsOptions | nindent 12 }}
            {{- end }}
          {{- end }}
          {{- if .Values.volumeMounts }}
          volumeMounts:
            {{- toYaml .Values.volumeMounts | nindent 8 }}
          {{- end }}
      {{- if .Values.volumes }}
      volumes:
        {{- toYaml .Values.volumes | nindent 6 }}
      {{- end }}
      {{- if .Values.hostAliases }}
      hostAliases:
        {{- toYaml .Values.hostAliases | nindent 6 }}
      {{- end }}
      {{- if .Values.dnsPolicy }}
      dnsPolicy: {{ .Values.dnsPolicy }}
      {{- end }}
      {{- if .Values.dnsConfig }}
      dnsConfig:
        {{- toYaml .Values.dnsConfig | nindent 6 }}
      {{- end }}
      {{- if .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml .Values.nodeSelector | nindent 6 }}
      {{- end }}
      {{- if .Values.affinity }}
      affinity:
        {{- toYaml .Values.affinity | nindent 6 }}
      {{- end }}
      {{- if .Values.tolerations }}
      tolerations:
        {{- toYaml .Values.tolerations | nindent 6 }}
      {{- end }}
      {{- if .Values.priorityClassName }}
      priorityClassName: {{ .Values.priorityClassName }}
      {{- end }}
      {{- if .Values.schedulerName }}
      schedulerName: {{ .Values.schedulerName }}
      {{- end }}
{{- end }}
