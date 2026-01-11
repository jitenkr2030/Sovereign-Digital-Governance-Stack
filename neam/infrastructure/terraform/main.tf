# Terraform Root Configuration for NEAM Platform Infrastructure
# Supports on-prem, private cloud, and air-gapped deployments

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
    null = {
      source  = "hashicorp/null"
      version = "~> 3.2"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.4"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }

  # Remote state configuration for team collaboration
  backend "s3" {
    bucket         = "neam-terraform-state"
    key            = "neam-platform/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "neam-terraform-locks"
  }
}

# Provider configurations
provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.kube_context
}

provider "helm" {
  kubernetes {
    config_path    = var.kubeconfig_path
    config_context = var.kube_context
  }
}

# Variable definitions
variable "environment" {
  description = "Deployment environment (production, staging, development)"
  type        = string
  default     = "production"
}

variable "kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = "~/.kube/config"
}

variable "kube_context" {
  description = "Kubernetes context to use"
  type        = string
  default     = ""
}

variable "deployment_model" {
  description = "Deployment model: on-prem, private-cloud, air-gapped, hybrid"
  type        = string
  default     = "on-prem"
}

variable "cluster_name" {
  description = "Name of the Kubernetes cluster"
  type        = string
  default     = "neam-cluster"
}

variable "region" {
  description = "Region for cloud resources (if applicable)"
  type        = string
  default     = "me-central-1"
}

# Resource group for Azure or tags for AWS
variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default = {
    Project     = "NEAM"
    Platform    = "National Economy Administration"
    Environment = "production"
    ManagedBy   = "Terraform"
    Compliance  = "Government Standard"
  }
}

# Module instantiation
module "network" {
  source = "./modules/network"

  environment     = var.environment
  deployment_model = var.deployment_model
  cluster_name    = var.cluster_name
  region          = var.region

  depends_on = []
}

module "kubernetes_cluster" {
  source = "./modules/kubernetes"

  environment      = var.environment
  deployment_model = var.deployment_model
  cluster_name     = var.cluster_name
  region           = var.region

  network = module.network

  depends_on = [module.network]
}

module "storage" {
  source = "./modules/storage"

  environment      = var.environment
  deployment_model = var.deployment_model
  cluster_name     = var.cluster_name

  kubernetes_provider = module.kubernetes_cluster

  depends_on = [module.kubernetes_cluster]
}

module "databases" {
  source = "./modules/databases"

  environment      = var.environment
  deployment_model = var.deployment_model
  cluster_name     = var.cluster_name

  kubernetes_provider = module.kubernetes_cluster
  storage            = module.storage

  depends_on = [module.kubernetes_cluster, module.storage]
}

module "monitoring" {
  source = "./modules/monitoring"

  environment      = var.environment
  deployment_model = var.deployment_model
  cluster_name     = var.cluster_name

  kubernetes_provider = module.kubernetes_cluster

  depends_on = [module.kubernetes_cluster]
}

module "security" {
  source = "./modules/security"

  environment      = var.environment
  deployment_model = var.deployment_model
  cluster_name     = var.cluster_name

  depends_on = [module.kubernetes_cluster]
}

# Outputs
output "cluster_endpoint" {
  description = "Kubernetes cluster API server endpoint"
  value       = module.kubernetes_cluster.endpoint
  sensitive   = true
}

output "cluster_ca_certificate" {
  description = "Kubernetes cluster CA certificate"
  value       = module.kubernetes_cluster.ca_certificate
  sensitive   = true
}

output "storage_endpoint" {
  description = "Storage endpoint for object storage"
  value       = module.storage.minio_endpoint
}

output "database_endpoints" {
  description = "Database connection endpoints"
  value       = module.databases.endpoints
}

output "monitoring_endpoints" {
  description = "Monitoring and observability endpoints"
  value       = module.monitoring.endpoints
}
