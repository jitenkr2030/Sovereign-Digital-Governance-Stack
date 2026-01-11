# NEAM Terraform Infrastructure Configuration
# Main infrastructure deployment for NEAM platform

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.24"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }

  backend "azurerm" {
    resource_group_name  = "neam-tfstate-rg"
    storage_account_name = "neamtfstate"
    container_name       = "tfstate"
    key                  = "neam-platform.tfstate"
  }
}

# Provider configuration
provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.kubernetes_context
}

provider "helm" {
  kubernetes {
    config_path    = var.kubeconfig_path
    config_context = var.kubernetes_context
  }
}

# Variables
variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "development"
}

variable "kubernetes_context" {
  description = "Kubernetes cluster context"
  type        = string
  default     = "neam-cluster"
}

variable "kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = "~/.kube/config"
}

variable "docker_registry" {
  description = "Docker registry URL"
  type        = string
  default     = "neamacr.azurecr.io"
}

variable "image_tag" {
  description = "Docker image tag"
  type        = string
  default     = "v1.0.0"
}

variable "cluster_name" {
  description = "Kubernetes cluster name"
  type        = string
  default     = "neam-cluster"
}

variable "location" {
  description = "Azure region for resource deployment"
  type        = string
  default     = "eastus"
}

# Local values
locals {
  common_labels = {
    app         = "neam"
    environment = var.environment
    managed_by  = "terraform"
    version     = var.image_tag
  }

  namespace          = "neam-system"
  monitoring_ns      = "neam-monitoring"
  data_ns            = "neam-data"

  service_ports = {
    api-gateway     = 8080
    sensing         = 8082
    intelligence    = 8083
    intervention    = 8084
    reporting       = 8085
    security        = 8086
    frontend        = 80
  }
}

# Random identifiers
resource "random_id" "cluster_prefix" {
  byte_length = 4
}

# Kubernetes namespace
resource "kubernetes_namespace" "neam_system" {
  metadata {
    name = local.namespace
    labels = local.common_labels
  }
}

resource "kubernetes_namespace" "neam_monitoring" {
  metadata {
    name = local.monitoring_ns
    labels = local.common_labels
  }
}

resource "kubernetes_namespace" "neam_data" {
  metadata {
    name = local.data_ns
    labels = local.common_labels
  }
}

# ConfigMap for common configuration
resource "kubernetes_config_map" "neam_config" {
  metadata {
    name      = "neam-config"
    namespace = local.namespace
    labels    = local.common_labels
  }

  data = {
    "environment"            = var.environment
    "docker_registry"        = var.docker_registry
    "image_tag"              = var.image_tag
    "kafka_prefix"           = "neam"
    "data_retention_days"    = "90"
    "log_level"              = "info"
  }
}

# Secret for sensitive configuration
resource "kubernetes_secret" "neam_secrets" {
  metadata {
    name      = "neam-secrets"
    namespace = local.namespace
    labels    = local.common_labels
  }

  type = "Opaque"

  data = {
    "jwt_secret"           = random_id.jwt_secret.hex
    "postgres_password"    = random_id.db_password.hex
    "clickhouse_password"  = random_id.db_password.hex
    "redis_password"       = random_id.redis_password.hex
    "api_key"              = random_id.api_key.hex
  }
}

resource "random_id" "jwt_secret" {
  byte_length = 32
}

resource "random_id" "db_password" {
  byte_length = 16
}

resource "random_id" "redis_password" {
  byte_length = 16
}

resource "random_id" "api_key" {
  byte_length = 32
}

# Service accounts
resource "kubernetes_service_account" "neam" {
  metadata {
    name      = "neam-service-account"
    namespace = local.namespace
    labels    = local.common_labels
  }
}

# Role-based access control
resource "kubernetes_cluster_role" "neam" {
  metadata {
    name = "neam-cluster-role"
    labels = local.common_labels
  }

  rule {
    api_groups = [""]
    resources  = ["pods", "services", "configmaps", "secrets", "endpoints"]
    verbs      = ["get", "list", "watch", "create", "update", "patch", "delete"]
  }

  rule {
    api_groups = ["apps"]
    resources  = ["deployments", "replicasets", "statefulsets"]
    verbs      = ["get", "list", "watch", "create", "update", "patch", "delete"]
  }
}

resource "kubernetes_cluster_role_binding" "neam" {
  metadata {
    name = "neam-cluster-role-binding"
    labels = local.common_labels
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "neam-cluster-role"
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.neam.metadata[0].name
    namespace = local.namespace
  }
}

# Network policies
resource "kubernetes_network_policy" "neam_deny_all" {
  metadata {
    name      = "neam-deny-all"
    namespace = local.namespace
  }

  spec {
    pod_selector {}

    policy_types = ["Ingress", "Egress"]

    ingress {
      from {
        namespace_selector {
          match_labels = {
            name = local.namespace
          }
        }
      }
    }

    egress {
      to {
        namespace_selector {
          match_labels = {
            name = local.namespace
          }
        }
      }

      to {
        pod_selector {
          match_labels = {
            k8s-app = "kube-dns"
          }
        }
      }
    }
  }
}

# Deployment for API Gateway
resource "kubernetes_deployment" "api_gateway" {
  metadata {
    name      = "neam-api-gateway"
    namespace = local.namespace
    labels    = merge(local.common_labels, { component = "api-gateway" })
  }

  spec {
    replicas = 2

    selector {
      match_labels = merge(local.common_labels, { component = "api-gateway" })
    }

    template {
      metadata {
        labels = merge(local.common_labels, { component = "api-gateway" })
      }

      spec {
        service_account_name = kubernetes_service_account.neam.metadata[0].name

        container {
          name  = "api-gateway"
          image = "${var.docker_registry}/neam/api-gateway:${var.image_tag}"

          port {
            container_port = local.service_ports["api-gateway"]
            name           = "http"
          }

          env_from {
            config_map_ref {
              name = kubernetes_config_map.neam_config.metadata[0].name
            }
          }

          env_from {
            secret_ref {
              name = kubernetes_secret.neam_secrets.metadata[0].name
            }
          }

          resources {
            requests = {
              cpu    = "500m"
              memory = "512Mi"
            }
            limits = {
              cpu    = "1000m"
              memory = "1Gi"
            }
          }

          liveness_probe {
            http_get {
              path = "/health"
              port = "http"
            }
            initial_delay_seconds = 30
            period_seconds        = 10
          }

          readiness_probe {
            http_get {
              path = "/ready"
              port = "http"
            }
            initial_delay_seconds = 10
            period_seconds        = 5
          }
        }
      }
    }
  }
}

# Service for API Gateway
resource "kubernetes_service" "api_gateway" {
  metadata {
    name      = "neam-api-gateway"
    namespace = local.namespace
    labels    = local.common_labels
  }

  spec {
    selector = merge(local.common_labels, { component = "api-gateway" })

    port {
      name     = "http"
      port     = local.service_ports["api-gateway"]
      target_port = local.service_ports["api-gateway"]
    }
  }
}

# Deployment for Sensing Layer
resource "kubernetes_deployment" "sensing" {
  metadata {
    name      = "neam-sensing"
    namespace = local.namespace
    labels    = merge(local.common_labels, { component = "sensing" })
  }

  spec {
    replicas = 3

    selector {
      match_labels = merge(local.common_labels, { component = "sensing" })
    }

    template {
      metadata {
        labels = merge(local.common_labels, { component = "sensing" })
      }

      spec {
        service_account_name = kubernetes_service_account.neam.metadata[0].name

        container {
          name  = "sensing"
          image = "${var.docker_registry}/neam/sensing:${var.image_tag}"

          port {
            container_port = local.service_ports["sensing"]
            name           = "http"
          }

          env_from {
            config_map_ref {
              name = kubernetes_config_map.neam_config.metadata[0].name
            }
          }

          resources {
            requests = {
              cpu    = "500m"
              memory = "512Mi"
            }
            limits = {
              cpu    = "2000m"
              memory = "2Gi"
            }
          }
        }
      }
    }
  }
}

# Deployment for Intelligence Engine
resource "kubernetes_deployment" "intelligence" {
  metadata {
    name      = "neam-intelligence"
    namespace = local.namespace
    labels    = merge(local.common_labels, { component = "intelligence" })
  }

  spec {
    replicas = 2

    selector {
      match_labels = merge(local.common_labels, { component = "intelligence" })
    }

    template {
      metadata {
        labels = merge(local.common_labels, { component = "intelligence" })
      }

      spec {
        service_account_name = kubernetes_service_account.neam.metadata[0].name

        container {
          name  = "intelligence"
          image = "${var.docker_registry}/neam/intelligence:${var.image_tag}"

          port {
            container_port = local.service_ports["intelligence"]
            name           = "http"
          }

          resources {
            requests = {
              cpu    = "1000m"
              memory = "2Gi"
            }
            limits = {
              cpu    = "4000m"
              memory = "8Gi"
            }
          }
        }
      }
    }
  }
}

# Deployment for Intervention Engine
resource "kubernetes_deployment" "intervention" {
  metadata {
    name      = "neam-intervention"
    namespace = local.namespace
    labels    = merge(local.common_labels, { component = "intervention" })
  }

  spec {
    replicas = 2

    selector {
      match_labels = merge(local.common_labels, { component = "intervention" })
    }

    template {
      metadata {
        labels = merge(local.common_labels, { component = "intervention" })
      }

      spec {
        service_account_name = kubernetes_service_account.neam.metadata[0].name

        container {
          name  = "intervention"
          image = "${var.docker_registry}/neam/intervention:${var.image_tag}"

          port {
            container_port = local.service_ports["intervention"]
            name           = "http"
          }

          resources {
            requests = {
              cpu    = "500m"
              memory = "512Mi"
            }
            limits = {
              cpu    = "2000m"
              memory = "2Gi"
            }
          }
        }
      }
    }
  }
}

# Deployment for Reporting Service
resource "kubernetes_deployment" "reporting" {
  metadata {
    name      = "neam-reporting"
    namespace = local.namespace
    labels    = merge(local.common_labels, { component = "reporting" })
  }

  spec {
    replicas = 2

    selector {
      match_labels = merge(local.common_labels, { component = "reporting" })
    }

    template {
      metadata {
        labels = merge(local.common_labels, { component = "reporting" })
      }

      spec {
        service_account_name = kubernetes_service_account.neam.metadata[0].name

        container {
          name  = "reporting"
          image = "${var.docker_registry}/neam/reporting:${var.image_tag}"

          port {
            container_port = local.service_ports["reporting"]
            name           = "http"
          }

          resources {
            requests = {
              cpu    = "500m"
              memory = "512Mi"
            }
            limits = {
              cpu    = "2000m"
              memory = "4Gi"
            }
          }
        }
      }
    }
  }
}

# Deployment for Security Service
resource "kubernetes_deployment" "security" {
  metadata {
    name      = "neam-security"
    namespace = local.namespace
    labels    = merge(local.common_labels, { component = "security" })
  }

  spec {
    replicas = 2

    selector {
      match_labels = merge(local.common_labels, { component = "security" })
    }

    template {
      metadata {
        labels = merge(local.common_labels, { component = "security" })
      }

      spec {
        service_account_name = kubernetes_service_account.neam.metadata[0].name

        container {
          name  = "security"
          image = "${var.docker_registry}/neam/security:${var.image_tag}"

          port {
            container_port = local.service_ports["security"]
            name           = "http"
          }

          resources {
            requests = {
              cpu    = "500m"
              memory = "512Mi"
            }
            limits = {
              cpu    = "2000m"
              memory = "2Gi"
            }
          }
        }
      }
    }
  }
}

# Deployment for Frontend
resource "kubernetes_deployment" "frontend" {
  metadata {
    name      = "neam-frontend"
    namespace = local.namespace
    labels    = merge(local.common_labels, { component = "frontend" })
  }

  spec {
    replicas = 3

    selector {
      match_labels = merge(local.common_labels, { component = "frontend" })
    }

    template {
      metadata {
        labels = merge(local.common_labels, { component = "frontend" })
      }

      spec {
        container {
          name  = "frontend"
          image = "${var.docker_registry}/neam/frontend:${var.image_tag}"

          port {
            container_port = local.service_ports["frontend"]
            name           = "http"
          }

          resources {
            requests = {
              cpu    = "100m"
              memory = "128Mi"
            }
            limits = {
              cpu    = "500m"
              memory = "256Mi"
            }
          }
        }
      }
    }
  }
}

# Ingress configuration
resource "kubernetes_ingress_v1" "neam" {
  metadata {
    name      = "neam-ingress"
    namespace = local.namespace
    annotations = {
      "kubernetes.io/ingress.class"                    = "nginx"
      "nginx.ingress.kubernetes.io/ssl-redirect"       = "true"
      "nginx.ingress.kubernetes.io/proxy-body-size"   = "100m"
      "nginx.ingress.kubernetes.io/proxy-read-timeout" = "300"
      "nginx.ingress.kubernetes.io/proxy-send-timeout" = "300"
    }
  }

  spec {
    tls {
      secret_name = "neam-tls"
      hosts       = ["*.neam.gov", "neam.gov"]
    }

    rule {
      host = "neam.gov"

      http {
        path {
          path      = "/api"
          path_type = "Prefix"

          backend {
            service {
              name = "neam-api-gateway"
              port {
                number = local.service_ports["api-gateway"]
              }
            }
          }
        }

        path {
          path      = "/"
          path_type = "Prefix"

          backend {
            service {
              name = "neam-frontend"
              port {
                number = local.service_ports["frontend"]
              }
            }
          }
        }
      }
    }
  }
}

# TLS Secret
resource "kubernetes_secret" "tls" {
  metadata {
    name      = "neam-tls"
    namespace = local.namespace
  }

  type = "kubernetes.io/tls"

  data = {
    "tls.crt" = var.tls_cert
    "tls.key" = var.tls_key
  }
}

variable "tls_cert" {
  description = "TLS certificate for HTTPS"
  type        = string
  default     = ""
}

variable "tls_key" {
  description = "TLS private key"
  type        = string
  default     = ""
}

# Output values
output "namespace" {
  value       = local.namespace
  description = "Kubernetes namespace for NEAM"
}

output "api_gateway_url" {
  value       = "https://neam.gov/api"
  description = "API Gateway URL"
}

output "frontend_url" {
  value       = "https://neam.gov"
  description = "Frontend URL"
}

output "services" {
  value = {
    api-gateway  = kubernetes_service.api_gateway.metadata[0].name
    sensing      = kubernetes_deployment.sensing.metadata[0].name
    intelligence = kubernetes_deployment.intelligence.metadata[0].name
    intervention = kubernetes_deployment.intervention.metadata[0].name
    reporting    = kubernetes_deployment.reporting.metadata[0].name
    security     = kubernetes_deployment.security.metadata[0].name
    frontend     = kubernetes_deployment.frontend.metadata[0].name
  }
  description = "NEAM service names"
}
