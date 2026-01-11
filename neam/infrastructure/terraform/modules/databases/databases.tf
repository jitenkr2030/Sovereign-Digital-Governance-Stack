# Database Module for NEAM Platform
# Deploys and configures all database systems

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

variable "storage" {
  type = any
}

# PostgreSQL configuration
variable "postgres_config" {
  description = "PostgreSQL configuration"
  type = object({
    instances       = number
    storage_size    = string
    storage_class   = string
    cpu_request     = string
    cpu_limit       = string
    memory_request  = string
    memory_limit    = string
  })
  default = {
    instances      = 3
    storage_size   = "100Gi"
    storage_class  = "nvme-ssd"
    cpu_request    = "2"
    cpu_limit      = "4"
    memory_request = "4Gi"
    memory_limit   = "8Gi"
  }
}

# TimescaleDB configuration
variable "timescaledb_config" {
  description = "TimescaleDB configuration"
  type = object({
    instances      = number
    storage_size   = string
    storage_class  = string
    cpu_request    = string
    cpu_limit      = string
    memory_request = string
    memory_limit   = string
  })
  default = {
    instances      = 3
    storage_size   = "500Gi"
    storage_class  = "nvme-ssd"
    cpu_request    = "4"
    cpu_limit      = "8"
    memory_request = "16Gi"
    memory_limit   = "32Gi"
  }
}

# ClickHouse configuration
variable "clickhouse_config" {
  description = "ClickHouse configuration"
  type = object({
    replicas_per_shard = number
    shards             = number
    storage_size       = string
    storage_class      = string
    cpu_request        = string
    cpu_limit          = string
    memory_request     = string
    memory_limit       = string
  })
  default = {
    replicas_per_shard = 2
    shards             = 2
    storage_size       = "200Gi"
    storage_class      = "nvme-ssd"
    cpu_request        = "4"
    cpu_limit          = "8"
    memory_request     = "16Gi"
    memory_limit       = "32Gi"
  }
}

# Helm chart for PostgreSQL Operator (CloudNativePG)
resource "helm_release" "postgresql" {
  name       = "neam-postgresql"
  repository = "https://cloudnative-pg.github.io/charts"
  chart      = "cloudnative-pg"
  version    = "0.20.0"
  namespace  = "neam-database"

  values = [
    <<-EOT
    # CloudNativePG configuration for NEAM

    image:
      repository: ghcr.io/cloudnative-pg/postgresql
      tag: "15.5"

    # crds:
    #   create: true

    controller:
      resources:
        requests:
          cpu: 100m
          memory: 128Mi
        limits:
          cpu: 500m
          memory: 512Mi

    # Monitoring
    monitoring:
      enabled: true
      podMonitor:
        enabled: true
        namespace: monitoring
    EOT
  ]
}

# Helm chart for ClickHouse Operator
resource "helm_release" "clickhouse" {
  name       = "clickhouse-operator"
  repository = "https://docs.clickhouse.com/operator/charts"
  chart      = "clickhouse-operator"
  version    = "0.22.0"
  namespace  = "neam-analytics"

  values = [
    <<-EOT
    clickhouse:
      image:
        repository: clickhouse/clickhouse-server
        tag: "23.12"

    config:
      profiles:
        default/max_memory_usage: 10737418240

      users:
        neam_admin/password: ${var.deployment_model == "production" ? "placeholder" : "neam_dev"}
        readonly/password: "readonly"

      quotas:
        default/intervals/0/duration: 3600

    metrics:
      enabled: true
      serviceMonitor:
        enabled: true
        namespace: monitoring

    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 512Mi
    EOT
  ]
}

# Helm chart for Redis Operator
resource "helm_release" "redis" {
  name       = "redis-operator"
  repository = "https://ot-container-kit.github.io/container-kit"
  chart      = "redis-operator"
  version    = "0.4.0"
  namespace  = "neam-cache"

  values = [
    <<-EOT
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

# PostgreSQL Cluster
resource "kubernetes_manifest" "postgres_cluster" {
  manifest = {
    apiVersion = "postgresql.cnpg.io/v1"
    kind       = "Cluster"
    metadata = {
      name      = "neam-postgres"
      namespace = "neam-database"
      labels = {
        app.kubernetes.io/name    = "neam-postgres"
        app.kubernetes.io/part-of = "neam"
      }
    }
    spec = {
      instances = var.postgres_config.instances
      imageName = "ghcr.io/cloudnative-pg/postgresql:15.5"
      storage = {
        storageClassName = var.postgres_config.storage_class
        size             = var.postgres_config.storage_size
      }
      resources = {
        requests = {
          cpu    = var.postgres_config.cpu_request
          memory = var.postgres_config.memory_request
        }
        limits = {
          cpu    = var.postgres_config.cpu_limit
          memory = var.postgres_config.memory_limit
        }
      }
      bootstrap = {
        initdb = {
          database = "neam_core"
          owner    = "neam_admin"
          dataChecksums = true
        }
      }
      postgresql = {
        parameters = {
          shared_buffers            = "256MB"
          max_connections           = "500"
          effective_cache_size      = "1GB"
          work_mem                  = "16MB"
          maintenance_work_mem      = "128MB"
          checkpoint_completion_target = "0.9"
        }
      }
      backup = {
        barmanObjectStore = {
          destinationPath = "s3://neam-backups/postgres"
          endpointURL     = "http://minio.neam-storage:9000"
        }
      }
    }
  }
}

# TimescaleDB Cluster (same PostgreSQL operator with TimescaleDB image)
resource "kubernetes_manifest" "timescaledb_cluster" {
  manifest = {
    apiVersion = "postgresql.cnpg.io/v1"
    kind       = "Cluster"
    metadata = {
      name      = "neam-timescaledb"
      namespace = "neam-database"
      labels = {
        app.kubernetes.io/name    = "neam-timescaledb"
        app.kubernetes.io/part-of = "neam"
      }
    }
    spec = {
      instances = var.timescaledb_config.instances
      imageName = "timescale/timescaledb-ha:pg15.5"
      storage = {
        storageClassName = var.timescaledb_config.storage_class
        size             = var.timescaledb_config.storage_size
      }
      resources = {
        requests = {
          cpu    = var.timescaledb_config.cpu_request
          memory = var.timescaledb_config.memory_request
        }
        limits = {
          cpu    = var.timescaledb_config.cpu_limit
          memory = var.timescaledb_config.memory_limit
        }
      }
      bootstrap = {
        initdb = {
          database = "neam_metrics"
          owner    = "neam_metrics_user"
          preInitScript = <<-EOT
            CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
            CREATE EXTENSION IF NOT EXISTS pg_cron;
          EOT
        }
      }
    }
  }
}

# Redis Cluster
resource "kubernetes_manifest" "redis_cluster" {
  manifest = {
    apiVersion = "redis-operator.k8s/v1beta1"
    kind       = "RedisCluster"
    metadata = {
      name      = "neam-redis"
      namespace = "neam-cache"
      labels = {
        app.kubernetes.io/name    = "neam-redis"
        app.kubernetes.io/part-of = "neam"
      }
    }
    spec = {
      redisClusterSize = 6
      image           = "redis:7.2-alpine"
      version         = "7.2"
      storage = {
        volumeClaimTemplate = {
          spec = {
            storageClassName = "nvme-ssd"
            accessModes      = ["ReadWriteOnce"]
            resources = {
              requests = {
                storage = "50Gi"
              }
            }
          }
        }
      }
      redisLeader = {
        replicas = 3
      }
      redisFollower = {
        replicas = 3
      }
    }
  }
}

# Output endpoints
output "endpoints" {
  description = "Database connection endpoints"
  value = {
    postgres = {
      host     = "neam-postgres.neam-database.svc.cluster.local"
      port     = 5432
      database = "neam_core"
    }
    timescaledb = {
      host     = "neam-timescaledb.neam-database.svc.cluster.local"
      port     = 5432
      database = "neam_metrics"
    }
    clickhouse = {
      host     = "neam-clickhouse.neam-analytics.svc.cluster.local"
      port     = 8123
      database = "neam_analytics"
    }
    redis = {
      host = "neam-redis.neam-cache.svc.cluster.local"
      port = 6379
    }
  }
}

output "connection_secrets" {
  description = "Secret names for database connections"
  value = {
    postgres    = "neam-postgres-app"
    timescaledb = "neam-timescaledb-app"
    clickhouse  = "neam-clickhouse-neam-clickhouse"
    redis       = "neam-redis"
  }
}
