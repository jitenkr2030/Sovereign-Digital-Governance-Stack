# Repository Configuration Guide

This comprehensive guide documents the manifest repository structure, configuration management practices, and operational procedures for implementing GitOps in the NEAM Platform. The guide covers repository organization, manifest templating strategies, image update automation, and sync wave management.

## Repository Structure Overview

The NEAM Platform GitOps repository structure follows principles that enable efficient management of Kubernetes manifests across multiple environments and applications. The structure balances standardization with flexibility, providing clear conventions while accommodating diverse application requirements.

The primary goal of the repository structure is to separate configuration concerns in a way that enables independent evolution of different aspects. Environment-specific configuration is separated from base application configuration, enabling promotion of unchanged base manifests through environments. Team-specific configuration is separated from shared platform configuration, enabling teams to manage their applications without affecting other teams.

The structure also supports parallel development workflows where multiple teams can work on different applications simultaneously without conflict. Clear ownership boundaries and merge strategies enable efficient collaboration while maintaining stability for shared components.

### Root Directory Organization

The repository root contains directories for different aspects of platform configuration. Each top-level directory serves a specific purpose and follows consistent internal organization.

The applications directory contains subdirectories for each application deployed through GitOps. Each application directory contains the base Kubernetes manifests that define the application without environment-specific modifications. Applications may be organized by team, by service type, or by business domain depending on organizational structure.

The environments directory contains subdirectories for each deployment environment. Each environment directory contains environment-specific configuration that overlays the base application manifests. This includes parameter values, resource limits, replica counts, and any environment-specific resources like ingress configurations.

The infrastructure directory contains platform-level configuration that spans multiple applications. This includes cluster-wide resource quotas, network policies, service mesh configuration, and shared monitoring configuration. Infrastructure configuration typically requires elevated permissions to modify.

The common directory contains shared resources used across multiple applications. This includes standard Kubernetes resources like service accounts, RBAC bindings, and network policies that are referenced by multiple applications. Centralization of common resources enables consistent configuration and simplified updates.

### Application Directory Structure

Each application directory follows a consistent internal structure that separates different aspects of the application configuration. This structure enables clear understanding of what each file contains and simplifies navigation for developers and operators.

The base subdirectory contains the primary Kubernetes manifests for the application. These manifests define the application components without environment-specific customization. Base manifests should be portable across environments, containing only the core application definition.

The overlays subdirectory contains environment-specific modifications applied to the base manifests. Each environment has a subdirectory containing Kustomize overlays or Helm values that customize the base manifests for that environment. Overlays should contain only the differences from base manifests, keeping them minimal and focused.

The components subdirectory contains additional resources that support the application but are not part of the core application. This might include monitoring configurations, backup policies, or disaster recovery resources. Components are optional and should be used when additional resources are needed.

A sample application directory structure appears as follows. The structure shows separation between base configuration and environment overlays, with each overlay containing only the differences needed for that environment.

```
applications/myapp/
├── base/
│   ├── kustomization.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   └── configmap.yaml
├── overlays/
│   ├── development/
│   │   ├── kustomization.yaml
│   │   └── replica-patch.yaml
│   ├── staging/
│   │   ├── kustomization.yaml
│   │   └── replica-patch.yaml
│   └── production/
│       ├── kustomization.yaml
│       ├── replica-patch.yaml
│       └── resource-limits.yaml
└── components/
    ├── monitoring/
    │   └── servicemonitor.yaml
    └── backup/
        └── backup-schedule.yaml
```

### Environment Directory Structure

The environments directory contains subdirectories for each deployment environment. Each environment directory aggregates configuration from multiple applications into a coherent deployment configuration.

The common subdirectory within an environment contains resources that apply to the entire environment. This includes environment-wide network policies, resource quotas, and namespace configurations. Environment-common resources are applied before application-specific resources.

The applications subdirectory within an environment contains references to application overlays for that environment. Rather than duplicating overlay content, this directory contains kustomization files that compose the appropriate overlays for the environment. This composition enables consistent ordering and validation across all applications in the environment.

A sample environment directory structure demonstrates the organization pattern. The structure shows how applications are composed into a complete environment deployment.

```
environments/production/
├── common/
│   ├── namespace.yaml
│   ├── resource-quota.yaml
│   └── network-policy.yaml
└── applications/
    ├── kustomization.yaml
    └── apps/
        ├── myapp.yaml
        └── otherapp.yaml
```

## Manifest Templating Strategies

### Kustomize Implementation

Kustomize is the primary templating tool for the NEAM Platform, providing overlay-based configuration management that preserves YAML structure and avoids the complexity of template processing. Kustomize enables environment-specific customization without modifying base manifests.

The base kustomization file declares the resources that comprise the application. Resources are specified as file paths relative to the kustomization file. The kustomization file also declares any common labels, annotations, or patches that should be applied to all resources.

Overlay kustomization files reference the base kustomization and specify patches that modify the base for the specific environment. Patches use the strategic merge patch format, which understands Kubernetes resource structures and handles field merging correctly. This format is more robust than JSON patch for Kubernetes resources.

Kustomize configurations should follow best practices for maintainability. Keep patches small and focused on specific changes. Use common labels to enable resource selection across the application. Leverage generators for configmaps and secrets to include content from files or environment variables.

A sample base kustomization demonstrates the basic structure. The kustomization declares resources and applies common labels for identification.

```yaml
# applications/myapp/base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - deployment.yaml
  - service.yaml
  - configmap.yaml

commonLabels:
  app.kubernetes.io/name: myapp
  app.kubernetes.io/managed-by: argocd
  app.kubernetes.io/part-of: neam-platform
```

A sample overlay kustomization demonstrates environment customization. The overlay references the base and applies patches to modify the deployment for the production environment.

```yaml
# applications/myapp/overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - path: replica-patch.yaml
    target:
      kind: Deployment
  - path: resource-limits.yaml
    target:
      kind: Deployment

commonLabels:
  environment: production
```

### Helm Chart Implementation

Helm charts provide an alternative templating approach for applications with complex configuration requirements. Helm's template processing enables conditional logic, iteration over lists, and sophisticated value transformations that exceed Kustomize's capabilities.

Chart structure follows the standard Helm conventions with templates, values, and chart metadata in defined locations. The templates directory contains templates that render to Kubernetes manifests. The values.yaml file contains default configuration values, and environment-specific values files override defaults.

Template development should follow Helm best practices for maintainability. Use named templates for reusable template fragments. Leverage helper functions to reduce repetition. Use the values file for all configurable parameters rather than hardcoding values in templates.

Values file organization should separate concerns similar to the repository structure. A values.yaml file contains the base configuration for the application. Environment-specific values files contain only the differences for each environment. This organization enables clear comparison between configurations.

A sample Helm values structure demonstrates the organization pattern. The structure shows separation between base values and environment-specific overrides.

```yaml
# charts/myapp/values.yaml
replicaCount: 1

image:
  repository: registry.example.com/myapp
  tag: latest

service:
  port: 80

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

```yaml
# charts/myapp/values-production.yaml
replicaCount: 3

image:
  tag: v1.2.0

resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
```

### Hybrid Approaches

Some applications benefit from hybrid approaches that combine Kustomize and Helm. This approach might use Kustomize for cross-cutting concerns while using Helm for application-specific templating. The hybrid approach requires careful organization to maintain clarity.

Common patterns for hybrid approaches include using Kustomize to manage resources that are shared across multiple applications while using Helm for application-specific resources. Kustomize overlays can reference Helm charts as resources, combining the two approaches within a single Application.

The decision between Kustomize and Helm should consider several factors. Kustomize is preferable when base manifests are stable and only small modifications are needed. Helm is preferable when complex conditional logic or value transformations are required. Consider team familiarity and tooling support when making the decision.

## Image Update Automation

### ArgoCD Image Updater Configuration

ArgoCD Image Updater automates the process of updating container image tags in Git manifests when new images are available. This automation reduces manual intervention in the deployment process while maintaining GitOps principles by committing changes back to the repository.

The Image Updater watches configured container images in the cluster and detects when new images are available. Detection strategies include polling container registries, webhooks from registries, and tag pattern matching. When a new image is detected that matches configured update rules, Image Updater modifies the Git manifests and commits the changes.

Installation of Image Updater adds the Image Updater components to the ArgoCD namespace. The installation includes the Image Updater deployment, service account, and RBAC permissions. Configure Image Updater with the same cluster access as ArgoCD to enable it to read image information from the cluster.

Configuration of Image Updater involves creating Kubernetes Custom Resources that define update policies for specific images. The ImageUpdatePolicy resource specifies the image to update, the update strategy, and the write-back method for committing changes to Git.

### Update Strategies

Image update strategies determine how Image Updater selects the new image version to apply. Different strategies suit different release models and risk tolerances.

The semver strategy updates to the highest version that satisfies a semver constraint. This strategy is appropriate for applications that follow semantic versioning and where major version changes should be reviewed before application. Constraints can specify minimum versions, allowed version ranges, or pinned major versions.

The latest strategy updates to the most recently pushed image tag. This strategy provides the fastest update cycle but carries the highest risk of deploying unstable images. Reserve this strategy for development environments or applications with extensive automated testing.

The glob strategy matches image tags against a pattern and updates to matching tags. This strategy enables custom update rules that do not fit semver or latest patterns. Glob patterns support wildcards and character classes for flexible matching.

A sample ImageUpdatePolicy demonstrates semver strategy configuration. The policy updates to the highest 1.x version when available and commits changes to the main branch.

```yaml
apiVersion: argocdimageupdater.argoproj.io/v1alpha1
kind: ImageUpdatePolicy
metadata:
  name: myapp-update-policy
  namespace: argocd
spec:
  imageName: registry.example.com/myapp
  updateStrategy: semver
  semver:
    range: ">=1.0.0 <2.0.0"
  gitWriteBack:
    branch: main
    commitMessageTemplate: "chore: Update {{.Image}} to {{.NewTag}}"
```

### Write-Back Configuration

The write-back configuration specifies how Image Updater commits image updates to Git. This configuration must balance automation speed with change management requirements.

Branch configuration specifies the target branch for commits. Direct commits to protected branches require bypassing branch protection or using a separate write-back branch with merge workflows. The write-back branch approach enables code review of automated updates before merging to the main branch.

Commit message configuration uses templates to generate commit messages for image updates. Templates can include image name, new tag, and other metadata. Well-crafted commit messages enable tracking of automated changes and simplify rollback decisions.

Pull request configuration can be used instead of direct commits, creating pull requests for image updates that require review before merging. This approach adds review overhead but ensures that all changes are reviewed. Configure pull request labels and reviewers for automated assignment.

### Registry Integration

Container registry integration enables Image Updater to query available image tags and detect new versions. Different registries have different authentication requirements and API characteristics.

Docker Registry v2 integration uses the registry API to list image tags. Authentication is configured through Kubernetes secrets containing registry credentials. The secret must have pull permissions for the target images.

GitHub Container Registry integration uses GitHub's API for tag listing. Authentication uses GitHub personal access tokens with appropriate scopes. Configure the token as a Kubernetes secret and reference it in Image Updater configuration.

Harbor registry integration uses Harbor's API for tag listing. Harbor supports more sophisticated filtering and retention policies. Configure Harbor credentials similarly to other registries.

## Sync Wave Management

### Wave Annotation Configuration

Sync waves use annotations on Kubernetes resources to specify the order of synchronization. Resources are processed in ascending wave number order, with lower-numbered waves completing before higher-numbered waves begin.

The sync wave annotation uses the format argocd.argoproj.io/sync-wave with an integer value. Negative wave numbers are allowed for resources that must be created first. Positive wave numbers are used for most resources, with higher numbers for resources with more dependencies.

Default wave assignment places resources without annotations in wave 0. This is appropriate for most resources where order does not matter. Explicit annotation is recommended when order matters, making the intended order visible in the manifest.

A sample deployment with sync wave annotation demonstrates the syntax. The deployment is assigned to wave 3, ensuring that lower-wave resources are available first.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    argocd.argoproj.io/sync-wave: "3"
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
```

### Wave Planning Best Practices

Effective wave planning ensures that resources are created in an order that satisfies dependencies. Poorly planned waves cause synchronization failures when dependent resources are not yet available.

Namespace resources should occupy the lowest wave numbers, typically wave 0. This ensures that namespaces exist before resources that reference them. CRD resources should also be in low waves to make custom resource types available early.

RBAC resources should follow namespaces and CRDs but precede resources that require the defined permissions. Service accounts, roles, and rolebindings typically occupy wave 1. This ordering ensures that permissions are established before resources that use them.

Storage resources should be created before pods that depend on them. Persistent volume claims, storage classes, and volume snapshot resources occupy wave 2. This ordering ensures that storage is available when pods are scheduled.

Application workloads typically occupy wave 3 or higher. Deployments, statefulsets, and daemon sets are created after their dependencies are available. Complex applications might use multiple waves for different workload components.

### Wave Validation

Validation of sync wave configuration ensures that the intended ordering is achieved during synchronization. Review the sync wave plan before applying changes to verify correct ordering.

The ArgoCD CLI provides a sync command with the --dry-run option that displays the planned synchronization without applying changes. This display includes the wave number for each resource, enabling verification of ordering.

Review the sync plan for resources that should have dependencies but are in the same or lower wave. If dependencies are not satisfied, adjust wave numbers or resource organization. Pay special attention to resources that reference secrets, configmaps, or persistent volume claims.

Check for wave number conflicts that cause resources to be created in unexpected order. Similar wave numbers should contain resources with similar dependency levels. Large gaps between wave numbers indicate potential organizational issues.

## Repository Procedures

### Change Management Workflow

Changes to the manifest repository follow a structured workflow that ensures quality and enables collaboration. This workflow applies to changes from any source including automated updates, developer modifications, and operational changes.

Changes begin as branches in the repository. Feature branches are created for new applications or significant changes. Hotfix branches are created for urgent production fixes. Branch names should follow a consistent convention that indicates the branch purpose.

Pull requests enable code review before changes are merged. Required reviewers include platform engineers for infrastructure changes and application owners for application changes. Reviewers verify that changes are correct, follow conventions, and do not introduce conflicts.

Continuous integration validates pull request changes before merge. Validation includes linting of YAML files, kustomize build verification, and optional deployment to test clusters. Pull requests cannot be merged until all validation passes.

Merge to the main branch triggers synchronization in development environments. ArgoCD detects the change and begins synchronization automatically. Monitor the synchronization to verify that changes apply correctly.

### Environment Promotion

Changes progress through environments in a controlled promotion process. This process validates changes in lower environments before production deployment.

Development environment receives all changes automatically after merge to main. This environment provides rapid feedback on changes and enables immediate validation. Development environment issues should be fixed by modifying the change and pushing updates.

Staging environment receives changes that have passed development validation. Promotion to staging typically requires explicit action or automatic scheduling. Staging validation includes integration testing and performance assessment.

Production environment receives changes after staging validation is complete. Production deployment requires explicit approval from the change approver. The approval process ensures that business stakeholders are aware of and approve production changes.

Rollback procedures enable rapid recovery when production issues are discovered. Rollback reverts the Git manifests to the previous known-good state. ArgoCD synchronization applies the reverted manifests, removing the problematic changes.

### Branching Strategy

The branching strategy defines how branches are used to manage changes across environments. The strategy balances flexibility for development with stability for production.

The main branch contains the current state of all environments. Changes are merged to main after review and validation. The main branch is protected and cannot be force-pushed or deleted.

Environment branches contain environment-specific configuration that is not appropriate for the main branch. These branches might contain secrets or environment-specific values that should not be in the main branch. Environment branches are managed carefully to prevent conflicts.

Feature branches are created for new applications or significant changes. These branches are short-lived and should be merged and deleted after the feature is complete. Long-lived feature branches accumulate merge conflicts and should be avoided.

Release branches can be created to stabilize a specific version for production. These branches receive only bug fixes while new development continues on the main branch. Release branches are deleted after the release is complete and support ends.

### Access Control

Repository access control restricts who can modify manifests and approve changes. Access control follows the principle of least privilege, granting only the permissions required for each role.

Write access to the manifest repository enables pushing branches and creating pull requests. Developers with application ownership have write access to their application directories. Platform engineers have write access to infrastructure directories.

Merge approval requires additional permissions beyond write access. Pull request approvers are specified per directory or globally. Approval requirements ensure that changes are reviewed before they can be merged.

Branch protection rules enforce access control at the repository level. The main branch requires pull requests and approval for all merges. Protected branches cannot be force-pushed or deleted.

## Configuration Validation

### YAML Linting

YAML linting validates the syntax and structure of manifest files before they are applied. Linting catches common errors that cause synchronization failures or unexpected behavior.

YAML syntax validation checks for correct indentation, proper data types, and valid YAML constructs. Syntax errors cause parsing failures during synchronization, resulting in failed resources. Catch syntax errors early through automated linting.

Schema validation compares YAML structures against Kubernetes resource schemas. Schema validation catches incorrect field names, missing required fields, and invalid field values. Kubernetes provides partial schema validation through kubectl, and additional tools provide more comprehensive validation.

Custom lint rules enforce organizational conventions beyond basic YAML syntax. Custom rules might check for required labels, naming conventions, or resource limits. Custom rules are implemented through configuration in the CI pipeline.

### Kustomize Build Verification

Kustomize build verification confirms that kustomization files are valid and produce the expected output. This verification catches configuration errors that cause sync failures.

Build validation executes kustomize build and verifies that the command succeeds. Build failures indicate syntax errors in kustomization files or missing resources. The build output can be stored as an artifact for debugging.

Output validation compares the build output against expected structures. Validation ensures that overlays produce the intended modifications and that resources are correctly composed. Unexpected output indicates configuration errors that require correction.

Diff validation compares the current build output against the previous version. This comparison highlights the changes that will be applied during synchronization. Review diffs to verify that changes are intentional and correct.

### Pre-Submit Validation

Pre-submit validation runs before changes are committed or merged, catching issues before they enter the repository. Pre-submit validation is integrated into development workflows to provide immediate feedback.

Local validation tools enable developers to validate changes before pushing. Pre-commit hooks run validation locally before git commits, catching issues early. Configure pre-commit hooks to match CI validation for consistency.

IDE integration provides real-time feedback as developers edit manifests. YAML language servers provide syntax highlighting and validation. Configure IDE plugins to use the same schema validation as CI pipelines.

Validation caching improves performance by avoiding redundant validation. Cache validation results based on file contents and skip validation when files have not changed. Implement caching in CI pipelines to speed up validation of unchanged files.