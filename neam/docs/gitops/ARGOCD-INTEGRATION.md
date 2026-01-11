# ArgoCD Integration Guide

This comprehensive guide provides detailed procedures for deploying, configuring, and operating ArgoCD as the GitOps reconciliation engine for the NEAM Platform. The guide covers installation procedures, cluster connection configuration, application definition management, synchronization policies, health assessment, and authentication integration.

## Introduction to ArgoCD in NEAM Platform

ArgoCD serves as the central GitOps controller for the NEAM Platform, continuously monitoring Git repositories that contain declarative infrastructure and application definitions, and automatically applying those definitions to target Kubernetes clusters. This approach eliminates the need for imperative deployment tools and ensures that the actual state of the platform always matches the desired state recorded in Git.

The decision to adopt ArgoCD for the NEAM Platform was driven by several key requirements. First, the platform needed a declarative approach to configuration management where all changes are version-controlled and auditable. Second, the platform required automatic synchronization capabilities that could recover from cluster failures without manual intervention. Third, the platform needed robust multi-tenancy support to isolate teams and workloads while providing centralized management. ArgoCD satisfies all of these requirements while providing an intuitive web interface that simplifies operations and troubleshooting.

ArgoCD operates by continuously polling Git repositories at configurable intervals to detect changes to manifest files. When changes are detected, ArgoCD compares the desired state from Git against the actual state running in the cluster. Differences trigger synchronization operations that update, create, or delete Kubernetes resources to achieve the desired state. This reconciliation loop runs continuously, ensuring that cluster state remains consistent with Git even when external modifications occur.

The platform architecture separates concerns between development and production environments. Development environments operate with automatic synchronization, enabling rapid iteration as developers push changes. Production environments require manual approval for synchronization operations, providing an opportunity for change review before deployment. This separation balances development velocity with production stability.

## Architecture Overview

The ArgoCD architecture for the NEAM Platform consists of multiple components that work together to provide GitOps capabilities. Understanding these components is essential for effective installation, configuration, and troubleshooting.

The API server exposes the ArgoCD API and web interface, handling all incoming requests from users and external systems. The API server validates resource definitions, manages authentication and authorization, and coordinates synchronization operations. For high availability deployments, multiple API server replicas can be running behind a load balancer.

The application controller is the core reconciliation engine that monitors Application custom resources and performs synchronization operations. The controller continuously watches Git repositories for changes, compares desired and actual states, and executes synchronization actions to reconcile differences. The controller also manages resource pruning and garbage collection when manifests are removed from Git.

The repository server maintains a cache of Git repositories and Helm charts referenced by Application resources. When the application controller needs to reconcile an Application, it requests the current manifests from the repository server, which clones or updates the local cache and returns the rendered manifests.

The Dex server provides identity provider integration for Single Sign-On authentication. Dex acts as an intermediary between ArgoCD and external identity providers, supporting OIDC, SAML, and OAuth2 protocols. This integration enables centralized user management and eliminates the need to manage separate ArgoCD credentials.

The Redis server provides caching and session management for the API server. In high availability configurations, Redis can be deployed as a cluster for fault tolerance. The Redis server also supports webhook handling for Git repository notifications.

The notification server handles delivery of alerts and notifications to external systems. This component supports multiple delivery channels including Slack, email, and webhooks. Notifications are triggered by Application lifecycle events, synchronization results, and health assessment changes.

## Installation Procedures

### Prerequisites Verification

Before beginning ArgoCD installation, verify that all prerequisites are satisfied. Kubernetes cluster version must be 1.22 or later, as ArgoCD requires features available in recent Kubernetes versions. The cluster must have sufficient resources to run ArgoCD components, typically requiring at least 2 CPU cores and 4GB of memory for the control plane components.

Certificate management infrastructure should be available for production deployments. While self-signed certificates can be used for evaluation, production deployments require certificates from a trusted certificate authority. Cert-manager integration is recommended for automated certificate management.

Storage provisioning must be configured for persistent volumes used by ArgoCD components. The repository server requires persistent storage for its Git cache, and the Redis server may require persistent storage depending on the deployment configuration. Ensure that the cluster has a default storage class configured or that specific storage claims can be satisfied.

Network access requirements include outbound access from the cluster to Git repositories over HTTPS, and inbound access to the ArgoCD API server from user browsers and CI/CD systems. Configure network policies and firewall rules to permit this traffic.

### Helm-Based Installation

The recommended installation method for the NEAM Platform uses Helm charts, which provide parameterized configuration and simplified upgrades. Begin by adding the ArgoCD Helm repository and updating the local cache.

First, add the ArgoCD Helm repository if it is not already present in your Helm configuration. The ArgoCD team maintains the official Helm chart repository, and this should be the source for production deployments. Use the Helm repository add command to include the ArgoCD charts in your available repositories.

After adding the repository, update the local cache to retrieve the latest chart information. This ensures that you have access to the most recent version of the ArgoCD chart and its dependencies. The Helm update command refreshes the repository cache and prepares Helm for installation operations.

Create a values file that customizes the ArgoCD installation for the NEAM Platform requirements. This file should specify the namespace for ArgoCD components, resource limits and requests for each component, ingress configuration for external access, and any customization of default behaviors. The values file serves as the authoritative configuration for the Helm release and should be stored in version control alongside other platform configuration.

Install ArgoCD using the Helm install command with the values file. This command creates the ArgoCD namespace, deploys all necessary resources, and performs initial configuration. The installation process typically takes two to five minutes depending on cluster performance and network conditions.

### Manifest-Based Installation

For environments where Helm is not available or where greater installation control is required, ArgoCD can be installed using raw Kubernetes manifests. This method provides complete visibility into all resources created during installation and enables fine-grained customization.

The ArgoCD project provides installation manifests in their GitHub repository. These manifests include all necessary resources organized by component. Begin by downloading the manifests for the desired ArgoCD version, ensuring compatibility between the manifest version and your Kubernetes cluster version.

Apply the namespace manifest to create the ArgoCD namespace. This namespace isolates ArgoCD components and provides a logical boundary for resource management. Configure namespace-level policies including resource quotas, network policies, and limit ranges according to your organization's requirements.

Apply the custom resource definition manifests to register the Application and AppProject CRDs with the Kubernetes API server. These CRDs extend the Kubernetes API to include ArgoCD-specific resources. The CRDs are cluster-scoped resources that must be installed before any Application resources can be created.

Apply the component manifests for each ArgoCD component. The recommended approach applies all manifests in the default installation folder, which includes the API server, application controller, repository server, Dex server, and Redis server. Review each manifest before application to ensure compatibility with your environment.

Configure resource limits and requests for each component based on your cluster's capacity and expected workload. The default manifests provide reasonable starting values for evaluation deployments, but production deployments typically require increased resource allocation. Monitor resource utilization after installation and adjust allocations as needed.

### Post-Installation Configuration

After ArgoCD is installed, several post-installation tasks complete the configuration for the NEAM Platform. These tasks configure external access, establish initial projects and applications, and integrate with platform monitoring systems.

Configure Ingress for external access to the ArgoCD API server and web interface. The Ingress configuration should specify the hostname for ArgoCD access, TLS certificate configuration, and any annotations required for the ingress controller. For production deployments, configure SSO authentication before exposing ArgoCD to external access.

Retrieve the initial admin password for ArgoCD access. The password is stored in a Kubernetes secret created during installation. Access the secret using kubectl and decode the base64-encoded password value. Store this password securely and plan to change it after initial access.

Create initial AppProject resources that define the scope of applications managed by ArgoCD. Each project specifies source repository patterns, destination cluster and namespace constraints, and cluster resource limits. Projects provide isolation between different teams or applications and enable role-based access control.

Configure cluster credentials for any additional clusters that ArgoCD will manage. ArgoCD requires credentials for each target cluster, stored as Kubernetes secrets in the ArgoCD namespace. The credentials must have sufficient permissions to create, update, and delete resources in the target namespaces.

## Cluster Connection Configuration

### Cluster Registration

ArgoCD manages multiple Kubernetes clusters from a single control plane. Each target cluster must be registered with ArgoCD and provided with appropriate credentials. The cluster registration process establishes the trust relationship and permissions required for synchronization operations.

The preferred method for cluster registration uses the ArgoCD CLI, which handles credential storage and cluster validation. Use the argocd cluster add command to register a new cluster. This command generates a service account in the target cluster, retrieves its token, and stores the credentials in a Kubernetes secret in the ArgoCD namespace.

When adding a cluster, specify the context name from your kubeconfig file or provide the Kubernetes API server URL directly. The command creates a secret containing the service account token and updates the ArgoCD cluster configuration to include the new target. Verify successful registration by listing registered clusters through the CLI or web interface.

For clusters that cannot be accessed directly from the machine running the ArgoCD CLI, manual registration is required. This involves creating a service account in the target cluster, retrieving its token, and creating a Kubernetes secret in the ArgoCD namespace with the cluster configuration. The secret must include the cluster URL, CA data, and token encoded in base64 format.

### Credential Management

Credentials for target cluster access require careful management to maintain security. Service accounts should have minimum necessary permissions, limiting the scope of actions ArgoCD can perform in each cluster. Create dedicated service accounts for ArgoCD rather than using existing accounts.

Permission scoping follows the principle of least privilege, granting only the permissions required for the managed workloads. For development clusters, broader permissions may be acceptable to enable rapid iteration. For production clusters, restrict permissions to the namespaces and resource types that ArgoCD manages.

Credential rotation should be performed periodically and when personnel changes require access revocation. The rotation process creates a new service account token, updates the ArgoCD cluster secret, and revokes the old token after verifying that synchronization continues to function correctly. Automate credential rotation where possible to reduce the operational burden.

### Multi-Cluster Architecture

The NEAM Platform architecture supports multiple clusters organized by environment and function. Development clusters enable rapid iteration and testing, staging clusters validate deployments before production release, and production clusters serve end users. ArgoCD manages all clusters from a central control plane.

Cluster organization in ArgoCD reflects the platform environment structure. Each cluster is registered with a descriptive name that indicates its purpose and environment. The ArgoCD UI displays clusters with their connection status and managed application count, providing visibility into the entire platform from a single view.

Cross-cluster dependencies require special consideration in the platform architecture. Applications that span multiple clusters must be designed to handle network partitioning and partial availability. Service mesh configurations enable secure communication between clusters, and ArgoCD synchronization order ensures that dependencies are available before dependent resources are created.

## Application Definition Management

### Application Resource Structure

ArgoCD Application resources define the desired state for deployed workloads. Each Application specifies a source (Git repository and path), a destination (cluster and namespace), and synchronization behavior. Understanding the Application resource structure is essential for effective GitOps management.

The Application spec.source section defines where to find the manifests for the application. Source configuration includes the Git repository URL, the revision to track (branch, tag, or commit), the path within the repository containing manifests, and any configuration management tools used to render the manifests. ArgoCD supports plain YAML manifests, Helm charts, Kustomize overlays, and Jsonnet configurations.

The Application spec.destination section specifies where to deploy the application. Configuration includes the Kubernetes API server URL (or cluster name for registered clusters) and the target namespace. If the target namespace does not exist, ArgoCD creates it during synchronization. Namespaces should be pre-created when strict namespace isolation is required.

The Application spec.syncPolicy section controls how synchronization occurs. Sync policies specify whether synchronization is automatic or manual, whether resources are pruned when removed from Git, whether self-healing is enabled to correct drift, and the sync options for specific resource types. Production applications typically use manual sync with pruning enabled to maintain Git as the source of truth.

### Project Configuration

AppProject resources provide organizational boundaries for applications within ArgoCD. Projects define which source repositories can be used, which destination clusters and namespaces are accessible, and what resource types can be deployed. Projects enable multi-tenancy by isolating teams and applications.

Source repository restrictions in projects specify allowed Git repository URLs using glob patterns. This prevents Applications from referencing untrusted repositories and ensures that manifest sources are within approved organizational boundaries. Repository patterns should be specific enough to prevent repository spoofing attacks.

Destination restrictions limit where Applications can deploy resources. Projects can whitelist specific clusters and namespaces, preventing Applications from deploying to unexpected locations. For production projects, destination restrictions should be as narrow as possible, allowing only approved target environments.

Cluster resource limits prevent runaway resource consumption by Applications within a project. These limits specify maximum CPU, memory, and pod counts that the project's Applications can request. Resource limits protect cluster stability by preventing individual Applications from consuming excessive resources.

### Source Configuration

Source configuration determines how ArgoCD retrieves and renders application manifests. Different source types have different configuration requirements and provide different capabilities for manifest management.

Git sources are the most common source type for GitOps workflows. Configure the repository URL using HTTPS or SSH protocol, depending on your network topology and authentication preferences. The revision field specifies which Git reference to track, typically a branch name for development or a tag for production releases. The path field specifies the directory within the repository containing manifests.

Helm sources treat Helm charts as the application source. Configure the chart repository URL, chart name, and version range. Helm sources support value files from the Git repository, enabling environment-specific configuration through different value files in different branches or directories.

Kustomize sources use Kustomize for manifest rendering. Configure the Git repository and path, and Kustomize overlays can provide environment-specific modifications to base manifests. Kustomize sources support multiple overlay directories that are applied in sequence, enabling layered configuration management.

### Destination Configuration

Destination configuration specifies where ArgoCD deploys application resources. Proper destination configuration ensures that applications are deployed to the correct clusters and namespaces with appropriate permissions.

The cluster destination can be specified as the Kubernetes API server URL or the cluster name for registered clusters. For registered clusters, use the cluster name as it appears in the ArgoCD cluster list. This provides consistent reference across management operations.

Namespace destination should specify the target namespace for application resources. If the namespace does not exist, ArgoCD creates it during synchronization. For namespaces with strict security requirements, pre-create the namespace and configure appropriate labels and annotations before creating the Application resource.

## Synchronization Policies

### Automatic vs Manual Sync

Synchronization policies determine how ArgoCD responds to changes in Git repositories. Automatic synchronization enables rapid deployment cycles by applying changes immediately when detected. Manual synchronization requires explicit approval before changes are applied, providing a review opportunity.

Automatic synchronization is appropriate for development environments where rapid iteration is valued and the impact of incorrect deployments is limited. When enabled, ArgoCD automatically synchronizes Applications when the Git repository changes. Configure an appropriate poll interval to balance responsiveness against Git repository load.

Manual synchronization is required for production environments where change review is mandatory. When enabled, changes are detected but not applied until an operator approves the synchronization through the CLI or web interface. Manual sync provides an opportunity to review changes, run additional validation, and coordinate deployments with operational procedures.

The sync policy is configured per Application and can be changed without modifying the Application resource. Consider using project-level defaults for consistency across applications while allowing exceptions where operational requirements differ.

### Sync Waves

Sync waves provide ordered synchronization of resources within an Application. Resources can be annotated with wave numbers that control the order of creation, update, and deletion operations. This ensures that dependencies are satisfied before dependent resources are created.

Sync waves are specified using the argocd.argoproj.io/sync-wave annotation on Kubernetes resources. Resources without the annotation are assigned to wave 0. Wave numbers can be negative to enable resources that must be created first. The recommended practice uses non-negative wave numbers with clear separation between resource types.

Typical wave organization places namespace and CRD definitions in wave 0, ensuring that custom resources are available before resources that reference them. Service accounts and RBAC resources follow in wave 1, enabling permissions for subsequent resources. Storage and database resources deploy in wave 2, making storage available for application pods. Application deployments occupy wave 3 or higher, depending on complexity.

Validation of sync wave configuration should verify that dependencies between resources are correctly represented. Use the ArgoCD CLI to view the sync wave plan before applying changes. This displays the resources that will be created, updated, or deleted in each wave, enabling verification of ordering before synchronization occurs.

### Pruning and Self-Healing

Pruning controls whether resources that are removed from Git are deleted from the cluster. When pruning is enabled, ArgoCD deletes resources that exist in the cluster but are not present in the Git manifests. This maintains consistency between Git and cluster state by removing orphaned resources.

Enable pruning for production applications to prevent resource drift. When manifests are removed from Git, the corresponding resources should be removed from the cluster. Configure pruning to run automatically or require manual approval depending on your change management procedures.

Self-healing enables ArgoCD to correct cluster drift automatically. When self-healing is enabled, ArgoCD detects any differences between the Git state and cluster state, including manual changes made directly to the cluster. Self-healing immediately synchronizes to restore the Git-defined state.

Self-healing is valuable for maintaining consistency but can interfere with debugging. When investigating issues, consider disabling self-healing temporarily to prevent automatic reversion of manual changes. Re-enable self-healing after investigation is complete.

### Sync Options

Sync options provide fine-grained control over specific synchronization behaviors. These options enable or disable particular aspects of the synchronization process, customizing it for application requirements.

The Validate option controls whether ArgoCD validates manifests before application. Validation uses kubectl apply --validate to check resource definitions for common errors. Disable validation for resources that trigger false-positive validation failures.

The CreateNamespace option controls whether ArgoCD creates the target namespace if it does not exist. When enabled, the namespace is automatically created during synchronization. Disable this option when namespace creation must be controlled through separate processes.

The RespectIgnoreDifferences option controls whether ArgoCD respects the ignoreDifferences configuration in the Application spec. When enabled, differences in fields listed in ignoreDifferences are not synchronized, allowing fields to be modified without being overwritten by Git state.

## Health Assessment Configuration

### Built-in Health Checks

ArgoCD provides built-in health checks for many Kubernetes resources, evaluating the health status based on resource-specific conditions. Understanding these built-in checks helps interpret Application health status and identify issues requiring attention.

Deployment health is assessed by checking the deployment's availability condition. A deployment is healthy when the availability condition is true and the observed generation matches the specified generation. Unhealthy deployments typically have availability conditions set to false, indicating that pods are not running as expected.

StatefulSet health follows a similar pattern, checking availability conditions and observed generation. StatefulSets have additional considerations for rolling updates, where the health check waits for all replicas to be ready before reporting healthy status.

Service health is typically immediate since services do not have health conditions. A service is considered healthy if it exists in the cluster, regardless of whether endpoints are ready. This means that services may appear healthy even when their backing pods are not running.

Ingress health checks the presence of the ingress resource and validates that an external IP or hostname is assigned. Some ingress controllers do not populate the status immediately, causing ArgoCD to report the ingress as progressing until the address is assigned.

### Custom Health Checks

Custom health checks extend ArgoCD's ability to evaluate the health of resources that lack built-in health assessment or require custom evaluation logic. Custom health checks are defined in the Application spec.ignoreDifferences or through health check scripts.

Custom resource health checks are defined using Lua scripts that evaluate resource-specific conditions. The Lua script receives the resource object and returns a health status of healthy, progressing, or degraded, along with a message explaining the status. This enables health assessment for custom resources defined by platform operators.

Health check scripts are suitable for complex health evaluation logic that exceeds Lua's capabilities. Scripts are stored in the ArgoCD repository and referenced by Application configuration. The script receives the resource definition as input and returns health status as output. This approach supports arbitrary health evaluation logic but requires additional configuration and maintenance.

Aggregate health status combines the health of multiple resources to produce an overall Application health status. The aggregation algorithm typically uses a worst-case approach where any degraded resource causes the overall status to be degraded. This approach ensures that issues are visible at the Application level.

### Health Assessment Best Practices

Effective health assessment requires configuration that accurately reflects application operational status. Misconfigured health checks can obscure issues by reporting healthy status when resources are not functioning correctly.

Configure health checks to match application operational requirements. For applications with database dependencies, include health checks that verify database connectivity. For applications with external dependencies, configure health checks that detect dependency unavailability.

Review health check configuration when deploying new resource types. Resources without built-in or custom health checks appear as progressing indefinitely, obscuring the status of other resources in the Application. Implement custom health checks for new resource types during initial deployment.

Monitor health assessment behavior over time to identify false positives or false negatives. Adjust health check configuration based on operational experience to improve accuracy. Document health check behavior for application operators.

## Authentication Configuration

### Initial Admin Authentication

ArgoCD creates an initial admin user during installation with a randomly generated password. This password must be retrieved and changed for secure operation. The admin account provides full access to ArgoCD functionality and should be protected accordingly.

The admin password is stored in the argocd-initial-admin-secret secret in the ArgoCD namespace. Retrieve the password using kubectl, decode the base64-encoded value, and use it to log in to the ArgoCD web interface or CLI. Change the password immediately after first access through the user settings page.

For production deployments, disable the admin account after establishing alternative authentication methods. The admin account's broad permissions make it a high-value target, and its static credentials create security risks. Use SSO authentication for regular access and reserve admin access for emergency recovery scenarios.

Password policies should be configured to enforce complexity requirements and expiration. ArgoCD supports password policies that require minimum length, character variety, and periodic rotation. Configure these policies through the ArgoCD configuration CM.

### Single Sign-On Configuration

SSO integration replaces local authentication with enterprise identity provider authentication. This provides centralized user management, enables multi-factor authentication, and eliminates the need to manage separate ArgoCD credentials.

ArgoCD uses Dex as an identity provider broker, supporting multiple identity provider types. Configure Dex connectors for your organization's identity provider type, which may include OIDC, SAML, LDAP, or OAuth2 providers. The Dex configuration is stored in the argocd-secret secret and must be updated when connector configuration changes.

OIDC configuration is the most common integration pattern. Configure the OIDC provider with the ArgoCD callback URL, client ID, and client secret. The client secret should be stored securely and rotated periodically. Configure claim mappings to extract user information from OIDC tokens for ArgoCD authorization decisions.

SAML configuration requires the identity provider's SAML metadata document and certificate. Configure attribute mappings to translate SAML assertions into ArgoCD user information. SAML integration typically requires additional configuration for session management and logout handling.

### Role-Based Access Control

ArgoCD implements role-based access control through AppProject resources and ArgoCD RBAC configuration. Combined, these mechanisms provide granular control over who can access and modify Applications.

AppProject-level permissions restrict which projects a user can access. Assign users to projects based on their role and responsibility. Project membership is managed through the ArgoCD CLI or web interface, with project admins able to add and remove members.

ArgoCD RBAC configuration extends permissions within projects using standard RBAC patterns. Roles define sets of permissions, and users or groups are assigned to roles. Default roles include admin (full access), api (API access), and readonly (view access). Custom roles can be defined for specific access patterns.

Permission scoping follows the principle of least privilege, granting only the permissions required for each user's responsibilities. Regular developers typically have create and update permissions for specific projects without delete permissions. Operations teams have broader permissions for troubleshooting and recovery. Change approval workflows require specialized permissions that enable review without modification.

## Operational Procedures

### Daily Operations

Daily GitOps operations include monitoring Application status, reviewing synchronization results, and addressing any issues that arise. Establish operational rhythms that include regular status reviews and proactive issue identification.

Morning status review should check all Applications for healthy status. Investigate any Applications in degraded or progressing status to determine the cause. Review synchronization history to identify any failed or partially completed synchronizations from the previous day.

Change coordination should document expected changes for the day and coordinate with affected teams. When deployments are planned, verify that the target environment is ready and that rollback procedures are documented. Communicate deployment windows to stakeholders.

End-of-day summary should document any outstanding issues, planned actions for the next day, and any deployments in progress. Update operational documentation to reflect any process improvements or lessons learned.

### Troubleshooting Synchronization Issues

Synchronization failures require systematic diagnosis to identify the root cause and implement corrective action. Common failure categories include manifest validation errors, resource conflicts, and permission issues.

Manifest validation errors indicate that Kubernetes cannot apply the manifests as defined. Review the error message to identify the specific resource and validation failure. Common causes include syntax errors, missing required fields, and incompatible API versions. Fix the manifests in Git and allow the synchronization to retry.

Resource conflicts occur when existing resources cannot be updated to match the desired state. This typically happens when resources have immutable fields that differ between the current and desired state. Review the conflict details and determine whether the resource must be deleted and recreated or whether the manifest must be modified.

Permission issues indicate that the ArgoCD service account lacks required permissions in the target cluster. Review the required permissions for the resources being synchronized and compare against the service account's permissions. Request additional permissions through your cluster management process.

### Upgrade Procedures

ArgoCD upgrades should be performed carefully to maintain operational continuity. Review release notes for breaking changes and test upgrades in non-production environments before applying to production.

Before upgrading, backup the ArgoCD configuration including Application resources, project definitions, and cluster configurations. This backup enables recovery if the upgrade encounters issues. Store the backup in a location separate from the ArgoCD namespace.

Perform the upgrade using Helm upgrade with the updated values file. Review the values file for any new configuration options that should be set. The upgrade process preserves existing Application resources while updating the ArgoCD control plane components.

After upgrade, verify that all Applications are still healthy and that synchronization continues to function correctly. Test a manual synchronization to verify that the control plane can reconcile with target clusters. Monitor the ArgoCD components for any errors or performance issues in the hours following the upgrade.

## Security Considerations

### Secret Management

ArgoCD requires access to multiple types of secrets including Git credentials, registry credentials, and cluster credentials. Proper secret management is essential for maintaining security throughout the GitOps workflow.

Git credentials for private repositories should be stored as Kubernetes secrets referenced by Application configuration. Use separate credentials for each repository to limit the impact of credential compromise. Rotate credentials periodically and immediately when personnel changes require access revocation.

Registry credentials for pulling container images should be stored as Kubernetes image pull secrets in the target namespaces. ArgoCD can create image pull secrets during synchronization if configured in the Application spec. Ensure that registry credentials have only pull permissions, not push permissions.

Cluster credentials for target clusters are stored as Kubernetes secrets in the ArgoCD namespace. These secrets contain service account tokens that ArgoCD uses for API access. Restrict access to the ArgoCD namespace to prevent unauthorized retrieval of cluster credentials.

### Network Security

Network security for ArgoCD involves controlling access to the API server, securing communication with Git repositories, and protecting inter-component communication within the cluster.

External access to the ArgoCD API server should be limited to authorized networks and users. Configure network policies that restrict access to the ArgoCD namespace from untrusted networks. Use VPN or zero-trust network access for remote administrative access.

Git repository access should use encrypted protocols (HTTPS or SSH with certificate validation). Configure ArgoCD to verify TLS certificates for Git servers. For internal Git servers, configure certificates from your organization's certificate authority.

Inter-component communication within the cluster should use mTLS where possible. ArgoCD components communicate over the cluster network, which is isolated from external networks by default. For high-security environments, consider service mesh integration to provide mTLS between components.

### Audit Logging

Audit logging captures all significant operations performed through ArgoCD, providing an audit trail for security investigations and compliance requirements. Configure audit logging to capture authentication events, synchronization operations, and configuration changes.

Authentication events include login attempts (successful and failed), session creation and termination, and credential use. These events should be forwarded to a centralized logging system for analysis and retention.

Synchronization events include the start and completion of each synchronization, resources created, updated, or deleted during synchronization, and any errors or warnings encountered. These events provide visibility into deployment activity and enable reconstruction of the deployment history.

Configuration change events capture modifications to Application resources, project configurations, and ArgoCD settings. These events enable tracking of who made changes and when, supporting change management and troubleshooting processes.