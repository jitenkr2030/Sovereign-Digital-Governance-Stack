# Chaos Engineering Guide

This comprehensive guide provides detailed documentation for implementing and operating Chaos Engineering practices within the NEAM Platform. The documentation covers infrastructure experiments, application-level chaos injection, resilience validation procedures, and performance degradation testing methodologies.

## Introduction to Chaos Engineering

Chaos Engineering is a disciplined approach to identifying failures before they become outages. By proactively injecting failures into systems, teams can build confidence in their capability to withstand realistic failure conditions. The NEAM Platform implements Chaos Engineering through Chaos Mesh, a cloud-native chaos engineering platform that orchestrates chaos experiments across Kubernetes clusters.

The philosophy behind chaos engineering in the NEAM Platform is to embrace failure as a natural part of distributed systems operation. Rather than attempting to prevent all failures, which is often impossible in complex distributed architectures, chaos engineering focuses on building systems that gracefully handle failure scenarios. This approach leads to more resilient applications, improved incident response procedures, and greater confidence in system reliability.

Chaos engineering experiments follow a scientific method: first, a steady state hypothesis is defined, then hypothetical failure scenarios are introduced, and finally, the system's behavior is observed and compared against the expected behavior. This systematic approach ensures that chaos engineering activities produce actionable insights rather than random disruptions.

The NEAM Platform's chaos engineering framework is organized into four primary categories: infrastructure experiments that target the underlying compute and network layers, application experiments that simulate failures at the service level, data loss validation procedures that verify data integrity during failure scenarios, and performance degradation testing that measures system behavior under stress conditions.

## Architecture Overview

The chaos engineering infrastructure consists of several interconnected components that work together to enable safe and controlled failure injection. At the center of this architecture is Chaos Mesh, which provides the orchestration capabilities for running chaos experiments. Chaos Mesh operates as a set of Kubernetes custom resource definitions and controllers, allowing chaos experiments to be defined and managed through standard Kubernetes manifests.

The experiment manifests are stored in the `k8s/chaos/` directory and are organized by experiment type. Infrastructure experiments are defined in `infrastructure-experiments.yaml`, which contains chaos experiments targeting pods, networks, and compute resources. Application experiments are defined in `application-experiments.yaml`, which contains experiments that simulate HTTP latency, error injection, and service dependency failures.

The validation layer consists of Python scripts located in the `scripts/chaos/` directory. The `data_loss_validator.py` script provides frameworks for measuring data consistency and detecting data loss during chaos events. The `performance_degradation_test.py` script combines load testing with chaos injection to measure performance bounds under various failure conditions.

The architecture is designed with safety as a primary concern. All chaos experiments include configurable scope definitions that specify which resources can be affected by the experiment. Experiments also include abort conditions that automatically stop the experiment if system behavior deviates significantly from expectations. This safety-first approach ensures that chaos engineering activities can be conducted in production environments without causing uncontrolled outages.

## Prerequisites and Requirements

Before implementing chaos engineering in the NEAM Platform, several prerequisites must be satisfied to ensure safe and effective experiment execution. This section outlines the requirements for both the control plane and the target systems where chaos experiments will be executed.

### Cluster Requirements

The Kubernetes cluster must be running version 1.18 or later, as Chaos Mesh requires certain Kubernetes features that are not available in older versions. The cluster should have sufficient resources to run both the Chaos Mesh control plane components and the workloads that will be subjected to chaos experiments. As a general guideline, allocate at least 2 CPU cores and 4GB of memory to the Chaos Mesh control plane, with additional resources allocated based on the scale of experiments being run.

Chaos Mesh components should be deployed in a dedicated namespace, typically called `chaos-mesh`, to provide isolation from other workloads. This namespace should have appropriate resource quotas and network policies configured to prevent chaos experiments from affecting the Chaos Mesh control plane itself.

### RBAC Configuration

Proper Role-Based Access Control (RBAC) configuration is essential for chaos engineering operations. The service account used by Chaos Mesh must have permissions to create, update, and delete the various chaos experiment resources. Additionally, the service account needs permissions to access and modify the target resources that will be affected by chaos experiments.

The RBAC configuration should follow the principle of least privilege, granting only the permissions necessary for the specific experiments being run. For example, if experiments only target pods in a specific namespace, the RBAC configuration should be scoped accordingly rather than granting cluster-wide permissions.

### Monitoring and Observability

Before running chaos experiments, ensure that comprehensive monitoring and observability tools are in place. This includes metrics collection through Prometheus, log aggregation, and distributed tracing capabilities. The observability stack should be configured to capture baseline metrics during normal operations, which will be used as a comparison point when evaluating experiment results.

Alerting rules should be configured to notify appropriate teams when system behavior deviates beyond acceptable thresholds during chaos experiments. However, it is important to distinguish between expected behavior during a chaos experiment and actual incidents that require immediate attention.

## Infrastructure Experiments

Infrastructure experiments target the foundational layers of the NEAM Platform, including compute resources, storage systems, and network components. These experiments simulate failures that can occur at the infrastructure level and help validate that the platform can continue operating despite such failures.

### Pod Failure Experiments

Pod failure experiments simulate scenarios where application pods become unavailable due to crashes, resource exhaustion, or container failures. These experiments are essential for validating that the platform's self-healing capabilities function correctly and that applications gracefully handle pod restarts.

The pod failure experiment is defined using the PodChaos custom resource in Chaos Mesh. The experiment configuration specifies the target namespace and label selector to identify which pods should be affected. The `action` field determines the type of pod failure to simulate, with options including pod-kill (immediately terminating the pod), container-kill (terminating a specific container within the pod), and pod-failure (making the pod unresponsive without terminating it).

When configuring pod failure experiments, it is important to consider the impact on dependent services. If a pod is terminated, any services that depend on that pod will experience failures until the pod is restarted. The experiment configuration should include appropriate grace periods and retry settings to ensure that the experiment reveals meaningful information about system behavior without causing cascading failures.

The experiment should be configured to target pods that have appropriate readiness and liveness probes configured. This ensures that the Kubernetes control plane can accurately detect pod failures and initiate recovery procedures. The experiment duration should be sufficient to observe the complete failure and recovery cycle, typically 5 to 15 minutes depending on the application's startup time.

### Network Chaos Experiments

Network chaos experiments simulate various network failure conditions including latency, packet loss, partition, and corruption. These experiments are crucial for validating that the platform behaves correctly when network conditions degrade, which is a common occurrence in distributed systems.

The NetworkChaos resource supports multiple actions that simulate different network failure scenarios. The `delay` action introduces latency into network communications by adding configurable delay to packets. The `loss` action simulates packet loss by randomly dropping a percentage of packets. The `partition` action creates network partitions by blocking communication between specific network segments. The `corruption` action simulates packet corruption by introducing bit errors into packets.

Network chaos experiments should be configured with realistic parameters based on the network conditions that the system is expected to handle. For example, an experiment might introduce 100ms of latency with 5% packet loss to simulate a degraded network connection. The experiment should be run for sufficient duration to observe the system's adaptive behavior under sustained network stress.

When configuring network partition experiments, carefully consider the partition topology. A complete network partition between all nodes would affect the Chaos Mesh control plane itself, so partitions should be scoped to affect only the target components. The partition configuration should simulate realistic failure scenarios such as the loss of connectivity between availability zones or between the application and its dependencies.

### Stress Experiments

Stress experiments subject compute resources to CPU and memory stress, validating that the system behaves correctly when resources become constrained. These experiments help identify resource leaks, capacity issues, and improper resource handling that might cause failures under high load.

The StressChaos resource supports both CPU stress and memory stress experiments. CPU stress experiments spawn worker processes that consume CPU cycles, simulating high computational load. Memory stress experiments allocate and hold memory to simulate memory pressure situations. The experiments can be configured to target specific containers within pods, allowing precise control over which components are stressed.

When running stress experiments, it is important to monitor both the stressed components and their dependencies. A component experiencing memory pressure might begin failing health checks, which could trigger cascading failures in dependent services. The experiment configuration should include appropriate thresholds for aborting the experiment if resource consumption becomes excessive.

## Application Experiments

Application experiments target the service layer of the NEAM Platform, simulating failures that occur at the application level. These experiments help validate that services handle errors gracefully and maintain acceptable behavior when dependencies fail or when the application itself encounters errors.

### HTTP Latency Injection

HTTP latency experiments simulate increased response times from HTTP services, helping to validate that the platform correctly handles slow dependencies. These experiments are particularly important for validating timeout configurations, circuit breaker behavior, and retry logic throughout the system.

The HTTPChaos resource in Chaos Mesh enables latency injection by intercepting HTTP requests and adding configurable delay before forwarding them to the target service. The latency can be configured as a fixed delay or as a variable delay with specified minimum and maximum values. The experiment can be scoped to specific HTTP methods, paths, or headers, allowing precise targeting of particular endpoints.

When configuring HTTP latency experiments, consider the impact on the overall request chain. Adding latency to a frequently-called service can cause thread pool exhaustion in calling services if timeout values are not properly configured. The experiment should start with conservative latency values and gradually increase to identify the system's breaking point.

The latency experiment should be combined with metrics collection to observe the impact on dependent services. Key metrics to monitor include thread pool utilization, queue depths, timeout rates, and error rates. These metrics help identify whether the system correctly propagates latency effects and whether timeout and circuit breaker configurations are appropriate.

### Error Injection

Error injection experiments simulate application-level errors by returning error responses to requests. These experiments validate that error handling paths throughout the system function correctly and that the system degrades gracefully when components fail.

The HTTPChaos resource supports several error injection modes. The `abort` mode immediately closes connections or returns error codes without forwarding requests to the target service. The `replace` mode intercepts successful responses and replaces them with error responses. The `mixed` mode combines success and error responses based on a configurable percentage.

Error injection experiments should cover various error scenarios including transient errors (5xx status codes), authentication errors (401), authorization errors (403), and resource not found errors (404). Each error type validates different error handling paths in the calling services. The experiment configuration should specify the appropriate error responses for the target service's API.

The experiment should validate that error handling is consistent across all services that depend on the target service. Check that retry logic correctly handles transient versus permanent errors, that circuit breakers open appropriately when error rates exceed thresholds, and that fallback paths are invoked when available.

### Service Dependency Failures

Service dependency failure experiments simulate the complete unavailability of external dependencies. These experiments validate that the platform maintains functionality when critical dependencies are unavailable, either through graceful degradation or by providing alternative functionality.

The experiments can target external services, databases, message queues, or any other dependencies that the application relies upon. The failure mode should match the realistic failure scenarios for the specific dependency type. For example, database dependency failures might simulate connection timeouts, while API dependency failures might simulate the service being completely unavailable.

The experiment configuration should include validation of fallback behavior. If the application implements circuit breaker patterns, verify that the circuit opens after the expected number of failures. If the application implements fallback logic, verify that the fallback provides acceptable functionality during the dependency outage. If the application degrades functionality, verify that the degradation is communicated appropriately to users.

## Data Loss Validation

Data loss validation procedures ensure that chaos engineering activities do not compromise data integrity and that the platform correctly protects against data loss during failure scenarios. These procedures combine automated validation with manual verification to provide comprehensive data integrity assurance.

### Data Consistency Framework

The data loss validation framework implemented in `scripts/chaos/data_loss_validator.py` provides automated checking of data consistency during chaos events. The framework operates by establishing a baseline data state, injecting chaos, and then verifying that the data remains consistent with the established baseline.

The validation framework supports multiple data consistency models depending on the characteristics of the data being validated. For strongly consistent data, the framework verifies that all reads return the most recently written value. For eventually consistent data, the framework allows for temporary inconsistencies during the chaos event but verifies that consistency is restored after the event concludes.

The framework is designed to be extensible, with plugins for different data stores and consistency checking approaches. The base framework provides common functionality including baseline establishment, consistency checking, and result reporting. Data store specific plugins implement the actual consistency checking logic for particular storage technologies.

### Validation Procedures

The data loss validation procedure begins with a quiescence period during which the system operates normally and baseline metrics are collected. During this period, the validation framework captures the expected data state, including checksums and sequence numbers where applicable. The baseline period should be long enough to capture normal system behavior and any periodic data operations.

After the baseline period, the chaos experiment is executed while the validation framework continuously monitors data consistency. The monitoring includes checking data store internals such as replication lag and consensus status, as well as checking application-level data invariants that should hold regardless of the underlying storage state.

Following the chaos event, the validation framework enters a recovery verification phase. During this phase, the system is allowed to recover from the chaos event while the framework continues monitoring. The recovery verification ensures that any temporary inconsistencies are resolved and that the data returns to a consistent state. The validation report includes details of any inconsistencies detected, their severity, and the system's recovery behavior.

### Handling Data Loss

If data loss is detected during validation, the procedure includes steps for investigation and remediation. The investigation begins by determining the scope and magnitude of the data loss. Small amounts of data loss might be acceptable for experiments targeting eventually consistent data stores, while any data loss in strongly consistent stores requires immediate attention.

The remediation approach depends on the nature and cause of the data loss. If the data loss is caused by an unexpected interaction between the chaos event and the application, the application logic may need to be modified to handle the failure scenario more gracefully. If the data loss is caused by improper configuration of the data store, the configuration should be corrected and the experiment repeated to verify the fix.

All detected data loss events should be documented with details about the chaos experiment, the data store configuration, the application version, and the investigation findings. This documentation feeds back into the chaos engineering process, helping to identify which failure scenarios require application changes and which can be handled through configuration adjustments.

## Performance Degradation Testing

Performance degradation testing combines load generation with chaos injection to measure how system performance changes under failure conditions. This testing validates that the system maintains acceptable performance even when components are experiencing failures, and it helps identify performance boundaries that might cause user-visible issues.

### Load Testing Framework

The performance degradation test framework implemented in `scripts/chaos/performance_degradation_test.py` combines a load testing component with chaos injection capabilities. The load testing component generates synthetic traffic that exercises the system's critical paths, while the chaos component injects failures during the load test.

The load testing framework supports various traffic patterns including constant load, step load, and burst load patterns. The traffic pattern should be chosen based on the expected production traffic characteristics and the specific performance aspects being tested. For example, a step load pattern might be used to identify how performance degrades as load increases, while a burst pattern might be used to validate recovery behavior after traffic spikes.

Key performance metrics captured during the test include response time percentiles (p50, p95, p99), throughput, error rates, and resource utilization. These metrics are collected continuously during the test and aggregated into time series for analysis. The framework also captures system metrics such as CPU, memory, and network utilization to correlate performance changes with resource constraints.

### Degradation Analysis

Performance degradation analysis compares the system's performance during chaos events against the baseline performance established during normal operation. The analysis identifies how much performance degrades under various failure scenarios and validates that degradation stays within acceptable bounds.

The analysis considers multiple dimensions of performance degradation. Response time degradation measures how much slower requests are served during chaos events compared to normal operation. Throughput degradation measures how much capacity is lost during chaos events. Error rate degradation measures how many additional errors are generated during chaos events.

The degradation analysis should establish acceptable thresholds for each performance dimension. These thresholds define the maximum acceptable degradation for each failure scenario. If degradation exceeds the threshold, the experiment is considered to have revealed a resilience issue that requires remediation. The remediation might involve application changes, infrastructure scaling, or configuration adjustments.

### Performance Baselines

Performance baselines provide reference points for comparing performance during chaos experiments. The baselines are established by running load tests under normal operating conditions, capturing the expected performance characteristics of the system.

Baseline establishment should account for variability in system performance. Multiple test runs should be conducted to capture the normal range of performance variation. The baseline should include not only average performance but also the expected variance, as this is important for distinguishing normal variation from degradation caused by chaos events.

Baselines should be updated periodically to reflect changes in the system. Major application releases, infrastructure changes, or significant traffic pattern changes should trigger baseline re-establishment. The baseline management process should include documentation of when baselines were established and what system changes have occurred since baseline establishment.

## Experiment Configuration Procedures

This section provides detailed procedures for configuring and executing chaos experiments in the NEAM Platform. The procedures cover the complete experiment lifecycle from initial planning through execution and analysis.

### Planning Chaos Experiments

Effective chaos engineering begins with thorough planning. Each experiment should have a clear objective that specifies what is being validated and what success looks like. The objective should be tied to a specific system property such as availability, durability, or performance.

The planning process should identify the minimum blast radius for the experiment. The blast radius should be small enough to limit potential impact while still being large enough to generate meaningful results. For production environments, start with small blast radius experiments and gradually expand as confidence in the system's resilience grows.

Experiment conditions define the specific failure scenarios to be tested. These conditions should be based on realistic failure modes that the system might encounter in production. Consider both common failures such as network latency and rare but severe failures such as complete service unavailability. The conditions should be documented with their expected system behavior and any metrics that indicate success or failure.

### Configuring Experiment Manifests

Experiment manifests are defined using Kubernetes custom resources and should be stored in the appropriate location within the repository. Infrastructure experiments are stored in `k8s/chaos/infrastructure-experiments.yaml`, while application experiments are stored in `k8s/chaos/application-experiments.yaml`.

When creating experiment manifests, start by specifying the experiment metadata including name, namespace, and labels. The name should be descriptive and follow naming conventions that identify the experiment type and target. Labels should enable filtering and organization of experiments.

The experiment specification defines the chaos action, target selector, and experiment mode. The target selector uses standard Kubernetes label selectors to identify the resources affected by the experiment. The experiment mode specifies whether the experiment runs once, continuously, or in a scheduled manner.

Include appropriate duration and scheduler settings in the experiment manifest. The duration should be sufficient to observe the expected behavior but not so long as to cause unnecessary disruption. For continuous experiments, the scheduler should define when the experiment runs and how long it runs each time.

### Executing Experiments

Before executing an experiment, verify that all prerequisites are satisfied. Confirm that monitoring and observability tools are functioning and that alerts are configured appropriately. Ensure that on-call personnel are aware of the planned experiment and have been briefed on what to expect.

Execute the experiment using kubectl to apply the experiment manifest. Monitor the experiment status using the Chaos Mesh dashboard or kubectl commands. During the experiment, observe the system metrics and verify that the experiment is affecting the target components as expected.

If unexpected behavior occurs during the experiment, be prepared to abort the experiment immediately. The abort mechanism should be documented and tested before experiment execution. Have the commands ready to stop the experiment and verify that the system returns to normal operation.

### Documenting Results

Experiment results should be documented thoroughly for future reference and organizational learning. The documentation should include the experiment objective, configuration, execution details, and observations. Quantitative metrics should be captured and compared against expected values.

The documentation should note any anomalies or unexpected behaviors observed during the experiment, even if they do not directly relate to the experiment objective. These observations might reveal other resilience issues that should be investigated separately. Also document any procedural issues encountered during experiment execution, as these might indicate opportunities to improve the chaos engineering process.

Results should be shared with relevant stakeholders including development teams, operations teams, and management. The sharing format should be appropriate to the audience, with technical details for engineering teams and summary findings for management. Consider presenting results at team meetings or reliability guilds to promote organizational learning.

## Safety Considerations

Safety is paramount in chaos engineering. While the goal is to discover system weaknesses, experiments must be conducted in a manner that does not cause uncontrolled harm to the system or its users.

### Experiment Safeguards

All chaos experiments should include safeguards that limit their impact and enable quick termination. These safeguards include scope limitations that restrict which resources can be affected, duration limits that automatically stop the experiment, and abort conditions that trigger experiment termination based on system metrics.

Scope limitations should be implemented through Kubernetes namespaces and label selectors. Experiments should target specific namespaces or pods rather than applying chaos broadly. Consider using chaos mesh namespaces to separate experimental workloads from production workloads.

Duration limits ensure that experiments do not run longer than necessary. Even if the experiment appears to be having no effect, prolonged experiments consume resources and may mask gradual degradation. Set durations based on the time needed to observe meaningful behavior, typically 5 to 30 minutes.

Abort conditions define thresholds for automatic experiment termination. These conditions might include high error rates, significant latency degradation, or resource exhaustion. The abort thresholds should be set conservatively to catch problems early while avoiding false positives during normal variation.

### Production Deployment Guidelines

When conducting chaos engineering in production environments, additional precautions are necessary. Coordinate experiments with relevant teams and ensure that incident response procedures are in place. Consider scheduling experiments during lower-traffic periods to minimize user impact.

Communication before and during experiments is essential. Notify relevant teams before the experiment begins and provide regular updates during execution. Establish clear channels for reporting issues discovered during experiments and for communicating experiment status.

Post-experiment verification should confirm that the system has fully recovered from the experiment. Check that all chaos resources have been cleaned up and that the system is operating normally. If recovery issues are discovered, investigate and remediate before the next experiment.

## Best Practices

The following best practices have been developed through experience with chaos engineering in the NEAM Platform and should guide all chaos engineering activities.

Start experiments in lower environments before running them in production. Staging and pre-production environments provide opportunities to refine experiments and identify issues without affecting users. Experiments that succeed in lower environments provide confidence for production execution.

Automate experiment execution and analysis to enable regular and consistent chaos engineering. Manual experiments are valuable for exploration but are not sustainable for ongoing resilience validation. Automated experiments can be run on schedules, ensuring regular testing of system resilience.

Integrate chaos engineering with the software development lifecycle. Consider chaos experiments as part of the testing process for new releases. Run experiments that target new features before they are deployed to production to validate that new code handles failure scenarios correctly.

Build a culture of psychological safety around chaos engineering. Experiments are intended to find weaknesses, and finding weaknesses is valuable. Celebrate experiments that reveal issues, as these represent opportunities to improve the system. Avoid blaming individuals for issues discovered through chaos engineering.

Maintain a backlog of experiment ideas and prioritize based on system risk. Not all failure scenarios are equally likely or impactful. Prioritize experiments that target high-risk components and realistic failure scenarios. The backlog should be regularly reviewed and updated based on system changes.

## Conclusion

Chaos engineering is an essential practice for building and maintaining resilient systems. The NEAM Platform's chaos engineering framework provides comprehensive capabilities for validating system behavior under failure conditions. By systematically injecting failures and observing system behavior, teams can identify and remediate weaknesses before they cause outages.

The documentation provided in this guide enables teams to configure and execute chaos experiments effectively. Following the procedures and best practices outlined here will help ensure that chaos engineering activities are safe, productive, and valuable for improving system reliability.

As the NEAM Platform evolves, so too should its chaos engineering practices. Regularly review and update experiment configurations to reflect system changes. Incorporate lessons learned from experiments into system design and operational procedures. By making chaos engineering a continuous practice, the platform can maintain and improve its resilience over time.