# NEAM Terraform Variables
# Variable definitions for infrastructure deployment

variable "environment" {
  description = "Deployment environment (development, staging, production)"
  type        = string
  default     = "development"
  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be one of: development, staging, production"
  }
}

variable "kubernetes_context" {
  description = "Kubernetes cluster context for deployment"
  type        = string
  default     = "neam-cluster"
}

variable "kubeconfig_path" {
  description = "Path to kubeconfig file for Kubernetes access"
  type        = string
  default     = "~/.kube/config"
}

variable "docker_registry" {
  description = "Docker registry URL for NEAM images"
  type        = string
  default     = "neamacr.azurecr.io"
}

variable "image_tag" {
  description = "Docker image tag to deploy"
  type        = string
  default     = "v1.0.0"
}

variable "cluster_name" {
  description = "Name of the Kubernetes cluster"
  type        = string
  default     = "neam-cluster"
}

variable "location" {
  description = "Azure region for resource deployment"
  type        = string
  default     = "eastus"
}

variable "resource_group_name" {
  description = "Azure Resource Group name"
  type        = string
  default     = "neam-rg"
}

variable "azure_location" {
  description = "Azure location for state storage"
  type        = string
  default     = "eastus"
}

# Database sizing
variable "postgres_storage_gb" {
  description = "PostgreSQL storage size in GB"
  type        = number
  default     = 100
}

variable "clickhouse_storage_gb" {
  description = "ClickHouse storage size in GB"
  type        = number
  default     = 500
}

# High availability
variable "enable_ha" {
  description = "Enable high availability deployment"
  type        = bool
  default     = false
}

# Node pool configuration
variable "node_pool_sku" {
  description = "VM SKU for worker nodes"
  type        = string
  default     = "Standard_D4s_v3"
}

variable "node_pool_count" {
  description = "Number of worker nodes"
  type        = number
  default     = 3
}

# Monitoring
variable "enable_monitoring" {
  description = "Enable monitoring and observability"
  type        = bool
  default     = true
}

variable "retention_days" {
  description = "Log retention period in days"
  type        = number
  default     = 90
}

# Network configuration
variable "enable_private_endpoints" {
  description = "Enable private endpoints for databases"
  type        = bool
  default     = true
}

variable "allowed_ip_ranges" {
  description = "IP ranges allowed to access the platform"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

# TLS configuration
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

# Feature flags
variable "enable_python_engine" {
  description = "Enable Python ML engine"
  type        = bool
  default     = true
}

variable "enable_temporal" {
  description = "Enable Temporal workflow engine"
  type        = bool
  default     = true
}

variable "enable_air_gapped" {
  description = "Enable air-gapped deployment mode"
  type        = bool
  default     = false
}
