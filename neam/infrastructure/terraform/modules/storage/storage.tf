# Storage Module for NEAM Platform
# Configures storage classes and persistent volume claims

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

# Storage configuration
variable "storage_classes" {
  description = "Storage class configurations"
  type = list(object({
    name          = string
    provisioner   = string
    reclaim_policy = string
    volume_binding_mode = string
    parameters    = map(string)
  }))
  default = [
    {
      name               = "nvme-ssd"
      provisioner        = "openebs.io/local"
      reclaim_policy     = "Retain"
      volume_binding_mode = "WaitForFirstConsumer"
      parameters = {
        type      = "nvme"
        caching   = "all"
        fsType    = "ext4"
      }
    },
    {
      name               = "hdd-bulk"
      provisioner        = "openebs.io/local"
      reclaim_policy     = "Retain"
      volume_binding_mode = "WaitForFirstConsumer"
      parameters = {
        type      = "hdd"
        caching   = "none"
        fsType    = "ext4"
      }
    },
    {
      name               = "network-ssd"
      provisioner        = "kubernetes.io/aws-ebs"
      reclaim_policy     = "Retain"
      volume_binding_mode = "WaitForFirstConsumer"
      parameters = {
        type      = "gp3"
        encrypted = "true"
        iops      = "3000"
        throughput = "125"
      }
    }
  ]
}

# Helm chart for local path provisioner (for on-prem/air-gapped)
resource "helm_release" "openebs" {
  name       = "openebs"
  repository = "https://openebs.github.io/charts"
  chart      = "openebs"
  version    = "3.9.0"
  namespace  = "openebs"

  values = [
    <<-EOT
    # OpenEBS configuration for NEAM Platform
    # Local Path Provisioner for persistent storage

    ndm:
      enabled: true
      sensors:
        enabled: true
        pathFilters:
          - /dev/sd*
          - /dev/nvme*
        deviceFilters:
          - ^sd[a-z]
          - ^nvme[0-9]n[0-9]

    localpv:
      enabled: true
      localpath:
        enabled: true
        image:
          repository: openebs/localpv-provisioner
          tag: 3.9.0
        resources:
          requests:
            cpu: 500m
            memory: 256Mi
          limits:
            cpu: 1
            memory: 512Mi
        # Storage directories
        hostpathClass:
          name: nvme-ssd
          isDefaultClass: false
          hostPath: /data/nvme
          capacity:
            default: 100Gi
            storageClass: 500Gi
        blockdeviceClass:
          name: block-lvm
          isDefaultClass: false

    # Resource configuration
    resources:
      limits:
        cpu: 500m
        memory: 512Mi

    # Security context
    securityContext:
      runAsNonRoot: true
      runAsUser: 1000
      fsGroup: 1000
    EOT
  ]
}

# Ceph storage for production (optional)
resource "helm_release" "rook_ceph" {
  count    = var.deployment_model == "production" ? 1 : 0
  name     = "rook-ceph"
  repository = "https://charts.rook.io/release"
  chart    = "rook-ceph"
  version  = "1.12.0"
  namespace = "rook-ceph"

  values = [
    <<-EOT
    # Rook Ceph configuration for production
    # Provides distributed storage with erasure coding

    operator:
      resources:
        requests:
          cpu: 500m
          memory: 256Mi
        limits:
          cpu: 1
          memory: 512Mi

    cephCluster:
      enabled: true
      storage:
        useAllNodes: false
        useAllDevices: false
        config:
          osdsPerDevice: "1"
          databaseSizeMB: "10240"
          bluestoreBlockSize: "4194304"
          osdDownScaleLimit: "1"
        directories:
          - path: /data/ceph
        resources:
          osd:
            requests:
              cpu: "1"
              memory: "4Gi"
            limits:
              cpu: "2"
              memory: "8Gi"
      resources:
        mgr:
          requests:
            cpu: 500m
            memory: 512Mi
        mon:
          requests:
            cpu: 500m
            memory: 1Gi

    cephObjectStore:
      enabled: true
      name: neam-object-store
      spec:
        gateway:
          port: 80
          sslCertificateRef: ""
        dataPool:
          erasureCoded:
            dataChunks: 4
            codingChunks: 2
        metadataPool:
          replicated:
            size: 3
        storage:
          useAllNodes: false
          nodes:
            - name: "storage-node-1"
              config:
                storeType: filestore
            - name: "storage-node-2"
              config:
                storeType: filestore
    EOT
  ]

  depends_on = [helm_release.openebs]
}

# Storage class resources
resource "kubernetes_storage_class" "this" {
  for_each = { for sc in var.storage_classes : sc.name => sc }

  metadata {
    name = each.value.name
    labels = {
      environment    = var.environment
      cluster_name   = var.cluster_name
      managed_by     = "terraform"
    }
  }

  provisioner   = each.value.provisioner
  reclaim_policy = each.value.reclaim_policy

  volume_binding_mode = each.value.volume_binding_mode

  dynamic "parameters" {
    for_each = each.value.parameters
    content {
      parameters[0] = parameters[1]
    }
  }
}

# Storage quotas
resource "kubernetes_resource_quota" "storage_quota" {
  metadata {
    name      = "storage-quota"
    namespace = "neam"
  }

  spec = {
    hard = {
      "persistentvolumeclaims" = "100"
      "storageclass.storage.k8s.io/requests.storage" = "50Ti"
      "storageclass.storage.k8s.io/persistentvolumeclaims" = "50"
    }
  }
}

# Output configuration
output "storage_classes" {
  description = "Available storage classes"
  value       = { for sc in var.storage_classes : sc.name => sc.name }
}

output "minio_endpoint" {
  description = "MinIO S3-compatible endpoint"
  value       = "http://minio.neam-storage:9000"
}

output "ceph_endpoint" {
  description = "Ceph object store endpoint"
  value       = var.deployment_model == "production" ? "http://rook-ceph-rgw-rook-ceph.rook-ceph:80" : ""
}
