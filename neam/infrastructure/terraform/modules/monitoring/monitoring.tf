# Monitoring Module for NEAM Platform
# Deploys Prometheus, Grafana, and related observability tools

variable "environment" {
  type = string
}

variable "deployment_model" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "kubernetes_provider" {
  type = any
}

# Helm chart for Kube Prometheus Stack
resource "helm_release" "kube_prometheus" {
  name       = "neam-monitoring"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = "55.5.0"
  namespace  = "monitoring"

  values = [
    <<-EOT
    # NEAM Platform Monitoring Configuration

    ## Global settings
    global:
      imageRegistry: ""
      imagePullSecrets: []
      prometheus:
        retention: "30d"
        retentionSize: "100GB"
      alertmanager:
        retention: "120h"

    ## Prometheus configuration
    prometheus:
      enabled: true
      prometheusSpec:
        replicas: 2
        retention: "30d"
        retentionSize: "100GB"
        storageSpec:
          volumeClaimTemplate:
            spec:
              storageClassName: nvme-ssd
              resources:
                requests:
                  storage: 200Gi
        resources:
          requests:
            cpu: 500m
            memory: 2Gi
          limits:
            cpu: 2
            memory: 8Gi
        ruleSelector:
          matchLabels:
            prometheus: neam-monitoring
        serviceMonitorSelector:
          matchLabels:
            release: neam-monitoring
        podMonitorSelector:
          matchLabels:
            release: neam-monitoring
        scrapeConfigSelector:
          matchLabels:
            release: neam-monitoring
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          fsGroup: 2000

    ## Alertmanager configuration
    alertmanager:
      enabled: true
      alertmanagerSpec:
        replicas: 3
        storage:
          volumeClaimTemplate:
            spec:
              storageClassName: nvme-ssd
              resources:
                requests:
                  storage: 10Gi
        resources:
          requests:
            cpu: 100m
            memory: 200Mi
          limits:
            cpu: 500m
            memory: 512Mi
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000

    ## Grafana configuration
    grafana:
      enabled: true
      replicas: 2
      persistence:
        enabled: true
        storageClassName: nvme-ssd
        size: 50Gi
      resources:
        requests:
          cpu: 100m
          memory: 256Mi
        limits:
          cpu: 1
          memory: 1Gi
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
      adminUser: admin
      adminPassword: "${GRAFANA_ADMIN_PASSWORD}"
      config:
        dashboardproviders: |
          dashboardproviders.yaml:
            path:
              dashboards: /var/lib/grafana/dashboards/neam
        datasources: |
          datasources.yaml:
            apiVersion: 1
            datasources:
              - name: Prometheus
                type: prometheus
                url: http://neam-monitoring-kube-prometheus:9090
                access: proxy
                isDefault: true
              - name: OpenSearch
                type: opensearch
                url: https://neam-opensearch.neam-search:9200
                access: proxy
                jsonData:
                  tlsSkipVerify: true
                  database: neam-logs
              - name: ClickHouse
                type: clickhouse
                url: http://neam-clickhouse.neam-analytics:8123
                access: proxy
              - name: PostgreSQL
                type: postgres
                url: postgresql://neam-postgres.neam-database:5432/neam_core
                access: proxy
                jsonData:
                  sslmode: require

    ## Node Exporter
    nodeExporter:
      enabled: true
      resources:
        requests:
          cpu: 100m
          memory: 256Mi
        limits:
          cpu: 500m
          memory: 512Mi

    ## Kube State Metrics
    kubeStateMetrics:
      enabled: true
      resources:
        requests:
          cpu: 100m
          memory: 256Mi
        limits:
          cpu: 500m
          memory: 512Mi

    ## Additional Service Monitors
    additionalServiceMonitors:
      - name: neam-postgres
        selector:
          matchLabels:
            app.kubernetes.io/name: neam-postgres
        namespaceSelector:
          matchNames:
            - neam-database
        endpoints:
          - port: metrics
            interval: 30s
      - name: neam-clickhouse
        selector:
          matchLabels:
            clickhouse.altinity.com/namespace: neam-analytics
        namespaceSelector:
          matchNames:
            - neam-analytics
        endpoints:
          - port: metrics
            interval: 30s
      - name: neam-redis
        selector:
          matchLabels:
            app.kubernetes.io/name: neam-redis
        namespaceSelector:
          matchNames:
            - neam-cache
        endpoints:
          - port: redis
            interval: 30s

    ## Additional Dashboards
    grafanaDashboards:
      default:
        neam-overview:
          gnetId: 13332
          revision: 1
          datasource: Prometheus
        postgres-overview:
          gnetId: 9628
          revision: 1
          datasource: PostgreSQL
        redis-dashboard:
          gnetId: 11835
          revision: 1
          datasource: Prometheus
    EOT
  ]
}

# Helm chart for OpenTelemetry Operator
resource "helm_release" "opentelemetry" {
  name       = "opentelemetry"
  repository = "https://open-telemetry.github.io/opentelemetry-helm-charts"
  chart      = "opentelemetry-operator"
  version    = "0.47.0"
  namespace  = "monitoring"

  values = [
    <<-EOT
    admissionWebhooks:
      create: true

    manager:
      resources:
        requests:
          cpu: 100m
          memory: 256Mi
        limits:
          cpu: 500m
          memory: 512Mi

    crds:
      create: true
    EOT
  ]
}

# OpenTelemetry Collector configuration
resource "kubernetes_manifest" "opentelemetry_collector" {
  manifest = {
    apiVersion = "opentelemetry.io/v1alpha1"
    kind       = "OpenTelemetryCollector"
    metadata = {
      name      = "neam-collector"
      namespace = "monitoring"
    }
    spec = {
      mode       = "daemonset"
      replicas   = 3
      config = {
        receivers = {
          otlp = {
            protocols = {
              grpc = {}
              http = {}
            }
          }
          prometheus = {
            config = {
              scrape_configs = [
                {
                  job_name = "neam-platform"
                  scrape_interval = "30s"
                }
              ]
            }
          }
        }
        processors = {
          batch = {}
          memory_limiter = {
            limit_mib = 1000
            spike_limit_mib = 200
          }
          resource = {
            attributes = [
              { key = "deployment.environment", value = var.environment }
            ]
          }
        }
        exporters = {
          prometheus = {
            endpoint = "0.0.0.0:8889"
          }
          otlp = {
            endpoint = "jaeger-collector:4317"
          }
        }
        service = {
          pipelines = {
            metrics = {
              receivers  = ["prometheus"]
              processors = ["batch", "memory_limiter"]
              exporters  = ["prometheus"]
            }
            traces = {
              receivers  = ["otlp"]
              processors = ["batch", "memory_limiter"]
              exporters  = ["otlp"]
            }
          }
        }
      }
    }
  }
}

# Helm chart for Jaeger
resource "helm_release" "jaeger" {
  name       = "jaeger"
  repository = "https://jaegertracing.github.io/helm-charts"
  chart      = "jaeger"
  version    = "0.77.0"
  namespace  = "monitoring"

  values = [
    <<-EOT
    provisionDataStore:
      cassandra: false
      elasticsearch: true
      kafka: false

    elasticsearch:
      enabled: true
      host: neam-opensearch.neam-search.svc.cluster.local
      port: 9200

    collector:
      replicaCount: 2
      resources:
        requests:
          cpu: 100m
          memory: 256Mi
        limits:
          cpu: 1
          memory: 1Gi

    query:
      replicaCount: 1
      resources:
        requests:
          cpu: 100m
          memory: 512Mi
        limits:
          cpu: 1
          memory: 2Gi

    agent:
      replicaCount: 1
    EOT
  ]
}

# Output monitoring endpoints
output "endpoints" {
  description = "Monitoring and observability endpoints"
  value = {
    prometheus    = "http://neam-monitoring-kube-prometheus:9090"
    grafana       = "http://neam-monitoring-grafana:3000"
    alertmanager  = "http://neam-monitoring-alertmanager:9093"
    jaeger_query  = "http://jaeger-query:16686"
    otlp_grpc     = "0.0.0.0:4317"
    otlp_http     = "0.0.0.0:4318"
  }
}

output "credentials" {
  description = "Monitoring credentials"
  value = {
    grafana_admin = "admin"
  }
  sensitive = true
}
