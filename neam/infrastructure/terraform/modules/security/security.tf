# Security Module for NEAM Platform
# Implements security hardening, RBAC, and compliance controls

variable "environment" {
  type = string
}

variable "deployment_model" {
  type = string
}

variable "cluster_name" {
  type = string
}

# Security context constraints for Pod Security Standards
resource "kubernetes_manifest" "pss_restricted" {
  manifest = {
    apiVersion     = "security.openshift.io/v1"
    kind           = "SecurityContextConstraints"
    metadata = {
      name = "neam-restricted"
      annotations = {
        "kubernetes.io/description" = "NEAM Platform Restricted Security Context Constraint"
      }
    }
    allowPrivilegeEscalation = false
    allowHostDirVolumePlugin = false
    allowHostNetwork         = false
    allowHostPID             = false
    allowHostPorts           = false
    allowPrivilegeContainer  = false
    allowedCapabilities      = []
    defaultAllowPrivilegeEscalation = false
    fsGroup = {
      type = "MustRunAs"
      ranges = [
        {
          min = 1000
          max = 65535
        }
      ]
    }
    runAsUser = {
      type = "MustRunAs"
      uid  = 1000
    }
    seLinuxContext = {
      type = "MustRunAs"
    }
    supplementalGroups = {
      type = "MustRunAs"
      ranges = [
        {
          min = 1000
          max = 65535
        }
      ]
    }
    volumes = [
      "configMap",
      "emptyDir",
      "persistentVolumeClaim",
      "secret"
    ]
    readOnlyRootFilesystem = true
  }
}

# Network Policies for namespace isolation
resource "kubernetes_manifest" "default_deny_network_policy" {
  for_each = toset([
    "neam",
    "neam-database",
    "neam-analytics",
    "neam-cache",
    "neam-search",
    "neam-storage",
    "neam-monitoring"
  ])

  manifest = {
    apiVersion = "networking.k8s.io/v1"
    kind       = "NetworkPolicy"
    metadata = {
      name      = "default-deny-all"
      namespace = each.value
    }
    spec = {
      podSelector = {}
      policyTypes = ["Ingress", "Egress"]
      ingress = [
        {
          from = [
            {
              namespaceSelector = {
                matchLabels = {
                  name = each.value
                }
              }
            }
          ]
        }
      ]
      egress = [
        {
          to = [
            {
              namespaceSelector = {
                matchLabels = {
                  name = "kube-system"
                }
              }
            }
          ]
        }
      ]
    }
  }
}

# RBAC Configuration
resource "kubernetes_manifest" "neam_admin_role" {
  manifest = {
    apiVersion = "rbac.authorization.k8s.io/v1"
    kind       = "ClusterRole"
    metadata = {
      name = "neam-admin"
      labels = {
        app.kubernetes.io/name = "neam-platform"
      }
    }
    rules = [
      {
      apiGroups = [""]
      resources = ["*"]
      verbs     = ["*"]
      },
      {
      apiGroups = ["apps", "networking.k8s.io"]
      resources = ["*"]
      verbs     = ["*"]
      },
      {
      apiGroups = ["postgresql.cnpg.io", "clickhouse.altinity.com", "redis-operator.k8s", "opensearch.opster.io"]
      resources = ["*"]
      verbs     = ["*"]
      }
    ]
  }
}

resource "kubernetes_manifest" "neam_viewer_role" {
  manifest = {
    apiVersion = "rbac.authorization.k8s.io/v1"
    kind       = "ClusterRole"
    metadata = {
      name = "neam-viewer"
      labels = {
        app.kubernetes.io/name = "neam-platform"
      }
    }
    rules = [
      {
      apiGroups = [""]
      resources = ["configmaps", "pods", "secrets", "services", "endpoints"]
      verbs     = ["get", "list", "watch"]
      },
      {
      apiGroups = ["apps"]
      resources = ["deployments", "statefulsets", "daemonsets"]
      verbs     = ["get", "list", "watch"]
      },
      {
      apiGroups = ["networking.k8s.io"]
      resources = ["ingresses", "networkpolicies"]
      verbs     = ["get", "list", "watch"]
      }
    ]
  }
}

# Service accounts for each namespace
resource "kubernetes_manifest" "neam_service_account" {
  for_each = toset([
    "neam",
    "neam-database",
    "neam-analytics",
    "neam-cache",
    "neam-search",
    "neam-storage"
  ])

  manifest = {
    apiVersion = "v1"
    kind       = "ServiceAccount"
    metadata = {
      name      = "neam-service"
      namespace = each.value
      labels = {
        app.kubernetes.io/name = "neam-platform"
      }
    }
    automountServiceAccountToken = true
  }
}

# Pod disruption budgets for high availability
resource "kubernetes_manifest" "postgres_pdb" {
  manifest = {
    apiVersion = "policy/v1"
    kind       = "PodDisruptionBudget"
    metadata = {
      name      = "neam-postgres-pdb"
      namespace = "neam-database"
    }
    spec = {
      maxUnavailable = 1
      selector = {
        matchLabels = {
          app.kubernetes.io/name = "neam-postgres"
        }
      }
    }
  }
}

resource "kubernetes_manifest" "redis_pdb" {
  manifest = {
    apiVersion = "policy/v1"
    kind       = "PodDisruptionBudget"
    metadata = {
      name      = "neam-redis-pdb"
      namespace = "neam-cache"
    }
    spec = {
      maxUnavailable = 1
      selector = {
        matchLabels = {
          app.kubernetes.io/name = "neam-redis"
        }
      }
    }
  }
}

resource "kubernetes_manifest" "opensearch_pdb" {
  manifest = {
    apiVersion = "policy/v1"
    kind       = "PodDisruptionBudget"
    metadata = {
      name      = "neam-opensearch-pdb"
      namespace = "neam-search"
    }
    spec = {
      maxUnavailable = 1
      selector = {
        matchLabels = {
          app.kubernetes.io/name = "neam-opensearch"
        }
      }
    }
  }
}

# Resource quotas per namespace
resource "kubernetes_manifest" "neam_quota" {
  for_each = toset([
    "neam",
    "neam-database",
    "neam-analytics",
    "neam-cache",
    "neam-search",
    "neam-storage"
  ])

  manifest = {
    apiVersion = "v1"
    kind       = "ResourceQuota"
    metadata = {
      name      = "neam-resource-quota"
      namespace = each.value
    }
    spec = {
      hard = {
        "cpu"    = "100"
        "memory" = "200Gi"
        "pods"   = "100"
        "services" = "50"
        "secrets" = "100"
        "configmaps" = "100"
        "persistentvolumeclaims" = "50"
      }
    }
  }
}

# Limit ranges
resource "kubernetes_manifest" "neam_limits" {
  for_each = toset([
    "neam",
    "neam-database",
    "neam-analytics",
    "neam-cache",
    "neam-search",
    "neam-storage"
  ])

  manifest = {
    apiVersion = "v1"
    kind       = "LimitRange"
    metadata = {
      name      = "neam-limit-range"
      namespace = each.value
    }
    spec = {
      limits = [
        {
          type = "Container"
          default = {
            cpu    = "500m"
            memory = "256Mi"
          }
          defaultRequest = {
            cpu    = "100m"
            memory = "128Mi"
          }
          max = {
            cpu    = "16"
            memory = "64Gi"
          }
          min = {
            cpu    = "10m"
            memory = "16Mi"
          }
        },
        {
          type = "Pod"
          max = {
            cpu    = "32"
            memory = "128Gi"
          }
          min = {
            cpu    = "10m"
            memory = "16Mi"
          }
        },
        {
          type = "PersistentVolumeClaim"
          max = {
            storage = "10Ti"
          }
          min = {
            storage = "1Gi"
          }
        }
      ]
    }
  }
}

# Audit logging policy
resource "kubernetes_manifest" "audit_policy" {
  manifest = {
    apiVersion = "auditregistration.k8s.io/v1alpha1"
    kind       = "AuditPolicy"
    metadata = {
      name      = "neam-audit-policy"
      namespace = "kube-system"
    }
    spec = {
      profile = "Baseline"
      rules = [
        {
          level = "RequestResponse"
          users = ["system:kube-proxy", "system:services"]
          verbs = ["get", "list", "watch"]
        },
        {
          level = "Metadata"
          resources = [
            {
              group      = ""
              resources  = ["pods", "services", "endpoints", "configmaps"]
              omitStages = ["RequestReceived"]
            }
          ]
        }
      ]
    }
  }
}

# Security annotations for pods
output "security_context_requirements" {
  description = "Required security context for all pods"
  value = {
    runAsNonRoot          = true
    runAsUser             = 1000
    runAsGroup            = 1000
    fsGroup               = 1000
    allowPrivilegeEscalation = false
    readOnlyRootFilesystem = true
    capabilities = {
      drop = ["ALL"]
    }
  }
}

output "network_policy_template" {
  description = "Network policy template for namespace isolation"
  value = {
    default_deny = {
      ingress = "Deny all by default"
      egress  = "Allow only to kube-system for DNS"
    }
    explicit_allow = [
      "neam namespace can access neam-database:5432",
      "neam namespace can access neam-cache:6379",
      "neam namespace can access neam-search:9200",
      "All namespaces can access monitoring for metrics"
    ]
  }
}
