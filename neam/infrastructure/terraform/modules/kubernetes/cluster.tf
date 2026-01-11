# Kubernetes Cluster Module
# Configures Kubernetes cluster with security hardening and government compliance

variable "environment" {
  type = string
}

variable "deployment_model" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "region" {
  type = string
}

variable "network" {
  type = any
}

# Control plane configuration
variable "control_plane_count" {
  description = "Number of control plane nodes"
  type        = number
  default     = 3
}

variable "worker_count" {
  description = "Number of worker nodes"
  type        = number
  default     = 6
}

# Node pool configurations
variable "control_plane_nodes" {
  description = "Control plane node specifications"
  type = list(object({
    instance_type = string
    cpu           = number
    memory        = number
    disk_size     = number
    labels        = map(string)
  }))
  default = [
    {
      instance_type = "m5.xlarge"
      cpu           = 4
      memory        = 16
      disk_size     = 100
      labels = {
        "node-role.kubernetes.io/control-plane" = ""
        "node-type"                            = "control-plane"
      }
    }
  ]
}

variable "worker_nodes" {
  description = "Worker node pool specifications"
  type = list(object({
    name            = string
    instance_type   = string
    cpu             = number
    memory          = number
    disk_size       = number
    min_count       = number
    max_count       = number
    labels          = map(string)
    taints          = list(object({ key, value, effect }))
  }))
  default = [
    {
      name          = "database"
      instance_type = "m5.2xlarge"
      cpu           = 8
      memory        = 32
      disk_size     = 200
      min_count     = 3
      max_count     = 6
      labels = {
        "workload-type" = "database"
        "node-type"     = "database"
      }
      taints = [
        { key = "database", value = "enabled", effect = "NoSchedule" }
      ]
    },
    {
      name          = "general"
      instance_type = "m5.xlarge"
      cpu           = 4
      memory        = 16
      disk_size     = 100
      min_count     = 2
      max_count     = 10
      labels = {
        "workload-type" = "general"
        "node-type"     = "general"
      }
      taints = []
    },
    {
      name          = "analytics"
      instance_type = "m5.4xlarge"
      cpu           = 16
      memory        = 64
      disk_size     = 500
      min_count     = 2
      max_count     = 4
      labels = {
        "workload-type" = "analytics"
        "node-type"     = "analytics"
      }
      taints = [
        { key = "analytics", value = "enabled", effect = "NoSchedule" }
      ]
    },
    {
      name          = "storage"
      instance_type = "m5.2xlarge"
      cpu           = 8
      memory        = 32
      disk_size     = 500
      min_count     = 2
      max_count     = 4
      labels = {
        "workload-type" = "storage"
        "node-type"     = "storage"
      }
      taints = [
        { key = "storage", value = "bulk", effect = "NoSchedule" }
      ]
    }
  ]
}

# Kubernetes version
variable "kubernetes_version" {
  description = "Kubernetes version to deploy"
  type        = string
  default     = "1.28"
}

# Network configuration
variable "pod_network_cidr" {
  description = "Pod network CIDR"
  type        = string
  default     = "10.244.0.0/16"
}

variable "service_cidr" {
  description = "Service network CIDR"
  type        = string
  default     = "10.96.0.0/12"
}

# Security configuration
variable "enable_mtls" {
  description = "Enable mutual TLS for service mesh"
  type        = bool
  default     = true
}

variable "enable_network_policies" {
  description = "Enable network policies by default"
  type        = bool
  default     = true
}

resource "null_resource" "kubernetes_cluster" {
  # This module creates the Kubernetes cluster
  # For different deployment models, different approaches are used:

  # For RKE2 (recommended for government):
  # - Uses RKE2 as the Kubernetes distribution
  # - Built-in etcd for high availability
  # - Enhanced security features

  # For AKS (private cloud):
  # - Uses Azure Kubernetes Service with private cluster
  # - Azure Active Directory integration
  # - Azure Policy integration

  # For air-gapped:
  # - Pre-configured RKE2 installation
  # - Local container registry

  triggers = {
    cluster_name      = var.cluster_name
    deployment_model  = var.deployment_model
    kubernetes_version = var.kubernetes_version
    pod_network_cidr  = var.pod_network_cidr
    service_cidr      = var.service_cidr
  }

  # For production deployment, use RKE2 or K3s for on-prem
  # This is a placeholder for the actual cluster creation
  provisioner "local-exec" {
    command = <<-EOT
      echo "Creating ${var.deployment_model} Kubernetes cluster: ${var.cluster_name}"
      echo "Kubernetes version: ${var.kubernetes_version}"
      echo "Pod CIDR: ${var.pod_network_cidr}"
      echo "Service CIDR: ${var.service_cidr}"
    EOT
  }
}

# Kubernetes configuration output
output "endpoint" {
  description = "Kubernetes API server endpoint"
  value       = "https://${var.cluster_name}.${var.deployment_model}.neam.gov.eg:6443"
  sensitive   = true
}

output "ca_certificate" {
  description = "Cluster CA certificate for authentication"
  value       = var.deployment_model == "on-prem" ? file(pathexpand("~/.kube/${var.cluster_name}/ca.crt")) : ""
  sensitive   = true
}

output "kubeconfig" {
  description = "Path to kubeconfig file"
  value       = pathexpand("~/.kube/${var.cluster_name}/kubeconfig.yaml")
}

output "node_pools" {
  description = "Configured node pools"
  value = {
    control_plane = var.control_plane_nodes
    worker        = var.worker_nodes
  }
}

output "network_config" {
  description = "Network configuration"
  value = {
    pod_cidr    = var.pod_network_cidr
    service_cidr = var.service_cidr
  }
}
