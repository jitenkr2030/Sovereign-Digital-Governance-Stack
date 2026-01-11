# Infrastructure Experiments Guide

This document provides detailed procedures for configuring and executing infrastructure-level chaos experiments in the NEAM Platform. Infrastructure experiments target the foundational compute, network, and storage layers of the platform, validating that the system can withstand failures at the infrastructure level while maintaining service continuity.

## Overview of Infrastructure Chaos

Infrastructure chaos engineering addresses failures that occur at the underlying layers supporting application workloads. These failures include pod crashes, network partitions, packet loss, CPU and memory pressure, and storage issues. By systematically testing the platform's response to these failures, teams can validate self-healing capabilities, identify single points of failure, and ensure that infrastructure-level resilience mechanisms function correctly.

The infrastructure chaos experiments are defined using Chaos Mesh custom resources and are organized into three primary categories: pod chaos experiments that simulate application container failures, network chaos experiments that simulate various network failure conditions, and stress chaos experiments that simulate resource contention scenarios. Each category addresses different aspects of infrastructure resilience and provides unique insights into system behavior.

Infrastructure experiments are generally safe to execute in production environments when proper safeguards are in place, as they target the infrastructure layer rather than application-specific functionality. However, the impact of infrastructure failures can cascade to application workloads, so careful blast radius management and monitoring are essential.

## Pod Chaos Experiments

Pod chaos experiments simulate failures that affect application containers running within Kubernetes pods. These experiments validate that the platform correctly detects and recovers from container failures, that applications handle unexpected termination gracefully, and that Kubernetes scheduling logic functions correctly under failure conditions.

### Pod Kill Experiment

The pod kill experiment immediately terminates target pods, simulating scenarios where containers crash due to application errors, resource exhaustion, or external termination signals. This experiment is fundamental for validating that the platform's self-healing mechanisms function correctly and that applications can recover from sudden termination.

To configure the pod kill experiment, create or modify the infrastructure-experiments.yaml file with a PodChaos resource specifying the pod-kill action. The experiment configuration must include a selector that identifies target pods based on namespace and labels. The selector should be precise enough to target only the intended pods, preventing unintended disruption to other workloads.

The experiment specification should include the mode of execution, which determines how many pods are affected simultaneously. The one mode affects one pod at a time, allowing observation of recovery behavior without overwhelming the system. The all mode affects all matching pods simultaneously, which is useful for testing how the system handles mass pod failures but requires careful consideration of blast radius.

Consider the following example configuration for a pod kill experiment targeting frontend services:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: frontend-pod-kill
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: frontend
  mode: one
  action: pod-kill
  gracePeriod: 0
```

The gracePeriod parameter specifies the termination grace period in seconds. A value of 0 causes immediate termination without allowing graceful shutdown procedures. This simulates a hard crash scenario where the container does not have opportunity to execute cleanup handlers or finish processing in-flight requests.

When executing pod kill experiments, monitor the following aspects of system behavior. First, verify that Kubernetes correctly detects the pod failure through its health checking mechanisms. Second, observe the pod rescheduling behavior, ensuring that replacement pods are created and scheduled appropriately. Third, validate that the application initializes correctly in the new pod, including database connections, cache warming, and any startup procedures.

### Container Kill Experiment

The container kill experiment terminates specific containers within multi-container pods, simulating scenarios where individual containers within a pod fail while others continue operating. This experiment is particularly relevant for validating sidecar patterns, ambassador containers, and other multi-container pod configurations.

The container kill action differs from pod kill in that it targets individual containers rather than the entire pod. This allows testing of scenarios where a container failure might affect other containers in the same pod, either through shared resources or inter-container communication dependencies.

The configuration for container kill experiments includes an additional container selector field that specifies which container within the target pod should be terminated. The container selector can match container names or use other identifying characteristics. All other configuration options remain consistent with pod kill experiments.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: sidecar-container-kill
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: api-gateway
  mode: one
  action: container-kill
  containerNames:
    - logging-sidecar
```

Container kill experiments should focus on validating that the primary application container continues functioning correctly when a sidecar container fails. Monitor application logs to verify that the primary container does not crash due to missing sidecar functionality, and validate that any inter-container communication mechanisms handle the failure gracefully.

### Pod Failure Experiment

The pod failure experiment makes a pod unavailable without terminating it, simulating scenarios where pods become unresponsive due to network issues, resource starvation, or application hang conditions. Unlike pod kill, this experiment does not trigger Kubernetes rescheduling, allowing observation of recovery behavior when pods enter failed states without being replaced.

The pod failure action intercepts network traffic to the target pod, making it unreachable without actually terminating the pod processes. This simulates conditions such as network partitions, firewall rule changes, or routing issues that prevent access to a running pod.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: api-pod-failure
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      tier: api
  mode: all
  action: pod-failure
```

Pod failure experiments are valuable for validating that load balancers and service meshes correctly detect pod unavailability and route traffic away from failed pods. Monitor service endpoint health checks to verify that Kubernetes service controllers detect the failure and remove the pod from service endpoints.

## Network Chaos Experiments

Network chaos experiments simulate various network failure conditions that commonly occur in distributed systems. These experiments validate that the platform handles degraded network conditions correctly, that circuit breakers and retry logic function appropriately, and that the system maintains acceptable behavior when network connectivity is impaired.

### Network Delay Experiment

The network delay experiment introduces configurable latency into network communications, simulating conditions such as network congestion, long-distance communication paths, or overloaded network intermediaries. This experiment is essential for validating that timeout configurations are appropriate and that the system remains responsive under elevated latency conditions.

The delay action adds specified latency to network packets using Linux traffic control mechanisms. The latency can be configured as a fixed value or as a range with minimum and maximum values to simulate variable network conditions. Jitter can also be added to introduce variability in latency, making the simulation more realistic.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: database-latency-injection
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      tier: database
  mode: all
  action: delay
  delay:
    latency: 100ms
    jitter: 20ms
    correlation: "50"
  direction: both
```

The correlation parameter specifies the correlation coefficient between successive packet delays. A correlation of 50 means that 50% of the latency variation is correlated with the previous packet, creating more realistic burst patterns rather than purely random latency variation.

When configuring network delay experiments, consider the cumulative effect of latency across service dependencies. Adding latency to a frequently-called database service can cause thread pool exhaustion in application services if timeout values are not properly configured. Start with conservative latency values and observe system behavior before increasing to more aggressive values.

### Network Loss Experiment

The network loss experiment simulates packet loss by randomly dropping a percentage of network packets. This experiment validates that the system correctly handles unreliable network conditions and that retry mechanisms function effectively when packets are lost.

The loss action can be configured with a percentage value that determines the probability of packet loss. The correlation parameter controls whether packet loss is random or bursty, with higher correlation values creating more realistic burst loss patterns that mimic real-world network conditions.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: service-mesh-packet-loss
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
  mode: all
  action: loss
  loss:
    loss: "10"
    correlation: "50"
  direction: both
```

Network loss experiments should focus on validating that applications implement appropriate retry logic for transient failures. Monitor error rates during the experiment to verify that lost packets are retried successfully and that the application eventually completes requests despite packet loss.

### Network Partition Experiment

The network partition experiment creates isolation between network segments, simulating scenarios where connectivity between components is completely lost. This experiment validates that the system handles partition scenarios correctly, including database cluster partitions, multi-region connectivity loss, and service dependency unavailability.

The partition action blocks traffic in the specified direction, preventing communication between source and target components. Unlike latency or loss experiments that degrade communication, partition experiments completely block communication, testing the system's ability to function when dependencies are completely unavailable.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: availability-zone-partition
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      zone: primary
  mode: all
  action: partition
  direction: both
  target:
    selector:
      namespaces:
        - production
      labelSelectors:
        zone: secondary
```

Network partition experiments should validate that the system correctly detects partitions and implements appropriate fallback behavior. This might include read-only operation modes, cached data usage, or graceful degradation of functionality. Monitor application logs to verify that partition detection mechanisms are functioning and that fallback behaviors are triggered appropriately.

### Network Corruption Experiment

The network corruption experiment introduces bit errors into network packets, simulating conditions such as electromagnetic interference, faulty network hardware, or software bugs in packet processing. This experiment validates that the system's error detection mechanisms correctly identify corrupted packets and trigger appropriate retry behavior.

The corruption action modifies a percentage of packets by introducing random bit flips. The error detection capability of the transport protocol (typically TCP checksum validation) will cause corrupted packets to be dropped, triggering retransmission. This validates that the system handles corruption-induced packet loss correctly.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: storage-network-corruption
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      component: storage
  mode: all
  action: corruption
  corruption:
    corruption: "5"
    correlation: "50"
```

Network corruption experiments are less common than latency or loss experiments but provide valuable validation of error handling paths. These experiments are particularly relevant for storage traffic and other scenarios where data integrity is critical.

## Stress Chaos Experiments

Stress chaos experiments subject compute resources to pressure, simulating scenarios where CPU, memory, or other system resources become constrained. These experiments help identify capacity issues, resource leaks, and improper resource handling that might cause failures under high load conditions.

### CPU Stress Experiment

The CPU stress experiment spawns worker processes that consume CPU cycles, simulating high computational load on the target system. This experiment validates that the system correctly handles CPU resource exhaustion and maintains acceptable behavior when CPU resources are constrained.

The CPU stress action creates processes that perform computational work, consuming CPU cycles according to the configured workers and load percentage. The configuration allows specifying the number of worker processes and the percentage of CPU capacity each worker should consume.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: api-cpu-stress
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: api-service
  mode: one
  action: cpu
  stressors:
    cpu:
      workers: 4
      load: 80
```

CPU stress experiments should validate that the application correctly handles resource constraints. Monitor request latency and error rates during the experiment to verify that the application degrades gracefully under CPU pressure. Also verify that Kubernetes resource limits and requests function correctly, causing appropriate scheduling decisions.

### Memory Stress Experiment

The memory stress experiment allocates and holds memory in the target containers, simulating memory pressure conditions that can occur under high load or due to memory leaks. This experiment validates that the system handles memory exhaustion correctly and that containers are terminated appropriately when memory limits are reached.

The memory stress action allocates memory in configurable block sizes, holding the memory to simulate sustained memory pressure. The configuration allows specifying the memory size to allocate and the allocation worker count.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: processor-memory-pressure
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: data-processor
  mode: one
  action: memory
  stressors:
    memory:
      workers: 2
      size: "1GB"
```

Memory stress experiments should verify that Kubernetes memory limits function correctly and that containers are terminated when memory limits are exceeded. Monitor container restart behavior and verify that the out-of-memory killer behaves as expected. Also validate that any in-flight requests are handled gracefully during container termination.

### IO Stress Experiment

The IO stress experiment creates load on the storage subsystem by performing read and write operations. This experiment validates that the system handles storage I/O bottlenecks correctly and that applications maintain acceptable behavior when storage performance degrades.

The IO stress action performs file system operations including reads, writes, and syncs to generate storage I/O load. The configuration allows specifying the number of workers and the operations per second for each worker.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: database-io-stress
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      tier: database
  mode: one
  action: io
  stressors:
    io:
      workers: 4
```

IO stress experiments should focus on validating storage-dependent operations. Monitor I/O wait times, disk queue depths, and application latency to verify that the system handles storage bottlenecks correctly. Also validate that database connections and other storage-dependent resources are managed correctly under I/O pressure.

## Execution Procedures

This section provides step-by-step procedures for executing infrastructure chaos experiments safely and effectively.

### Pre-Experiment Checklist

Before executing any infrastructure chaos experiment, complete the following checklist to ensure safe experiment execution. First, verify that monitoring and alerting systems are operational and configured to capture relevant metrics. Second, ensure that on-call personnel are aware of the planned experiment and have confirmed availability. Third, confirm that rollback procedures are documented and that personnel are trained in their execution.

The experiment configuration should be reviewed to verify that the blast radius is appropriate for the environment. Production experiments should have minimal blast radius, targeting specific pods or namespaces rather than cluster-wide effects. Verify that the experiment duration is set appropriately, long enough to observe meaningful behavior but short enough to limit disruption.

Document the expected system behavior during the experiment, including acceptable and unacceptable outcomes. This documentation serves as a reference for evaluating experiment results and determining whether the experiment succeeded in its objectives.

### Experiment Execution

Execute the experiment using kubectl apply to create the chaos resource. Immediately verify that the experiment is creating by checking the experiment status with kubectl describe. Monitor the experiment progress through the Chaos Mesh dashboard or by watching related resources.

During the experiment execution, continuously monitor system behavior. Key metrics to observe include error rates, latency distributions, throughput, and resource utilization. Also monitor application logs for unexpected messages or error patterns. If any metric deviates significantly from expectations or if unacceptable behavior is observed, immediately execute the abort procedure.

Document observations throughout the experiment execution. Note the time of significant events, metric changes, and any anomalies observed. This documentation is valuable for post-experiment analysis and for building organizational knowledge about system behavior under failure conditions.

### Post-Experiment Verification

After the experiment completes or is aborted, verify that all chaos resources have been cleaned up. Check that no lingering chaos resources are affecting the system and that all pods, network configurations, and stress processes have returned to normal operation.

Conduct a comprehensive health check of affected systems to verify normal operation. This includes checking pod health, service connectivity, database connections, and any other critical paths. If any issues are discovered during the health check, investigate and remediate before concluding the experiment.

Compile experiment results including the experiment configuration, observations during execution, and comparison against expected behavior. If the experiment revealed unexpected system behavior, document the behavior and create follow-up tasks to investigate and address the underlying causes.

## Safety Guidelines

Infrastructure chaos experiments can have significant impact on system behavior and must be conducted with appropriate safety measures in place.

Always implement experiment safeguards including scope limitations, duration limits, and abort conditions. Scope limitations should restrict experiments to specific namespaces or label selectors, preventing broad impact. Duration limits should ensure experiments do not run longer than necessary. Abort conditions should trigger automatic experiment termination if system behavior deviates significantly from expectations.

For production environments, consider time-of-day scheduling to minimize user impact. Experiments affecting customer-facing services should generally be conducted during low-traffic periods. Coordinate with support teams to ensure that experiment-related issues are not misinterpreted as customer incidents.

Maintain clear communication throughout the experiment lifecycle. Notify relevant teams before experiment execution begins and provide regular status updates during execution. Establish clear channels for reporting issues discovered during experiments and for communicating experiment status.

## Troubleshooting

Common issues encountered during infrastructure chaos experiments and their resolutions are documented here to support experiment execution.

If the experiment does not appear to be affecting target resources, verify that the selector is correctly matching the intended resources. Use kubectl get pods with label selectors to verify that the expected pods are being matched. Also verify that the Chaos Mesh controllers have permission to access and modify the target resources.

If the experiment causes unexpected system behavior, immediately abort the experiment and verify system recovery. If the system does not recover automatically, manual intervention may be required. Document the unexpected behavior and investigate the root cause before attempting similar experiments.

If chaos resources cannot be deleted after an experiment, use kubectl delete with the force flag to remove stuck resources. If the resources persist, restart the Chaos Mesh controllers to clear any reconciliation issues.