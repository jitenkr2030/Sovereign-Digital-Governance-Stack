# Chaos Engineering Documentation

This directory contains comprehensive documentation for the Chaos Engineering implementation in the NEAM Platform. The documentation provides detailed procedures for implementing, configuring, and operating chaos engineering practices to validate system resilience under failure conditions.

## Documentation Overview

The chaos engineering documentation is organized into five primary guides, each addressing specific aspects of chaos engineering within the NEAM Platform. Together, these guides provide complete coverage of chaos engineering concepts, implementation procedures, and operational guidance.

The documentation suite begins with a comprehensive overview that introduces chaos engineering principles and their application within the NEAM Platform. This overview establishes the foundational knowledge necessary to understand and effectively utilize the more detailed procedure guides. The subsequent guides dive into specific technical areas, providing hands-on procedures for implementing and operating chaos engineering capabilities.

All documentation follows consistent conventions and terminology, enabling readers to move between guides while maintaining a coherent understanding of the overall chaos engineering framework. Cross-references between guides help readers navigate to related information and understand how different aspects of chaos engineering connect.

## Available Guides

### Chaos Engineering Guide (CHAOS-ENGINEERING-GUIDE.md)

The Chaos Engineering Guide serves as the primary entry point for understanding chaos engineering in the NEAM Platform. This comprehensive guide covers the fundamental concepts of chaos engineering, the architecture of the chaos engineering infrastructure, and the overall approach to implementing chaos experiments.

The guide provides an introduction to the philosophy behind chaos engineering and explains how this approach differs from traditional testing methodologies. It describes the disciplined scientific method that underlies effective chaos engineering, including hypothesis formation, experiment design, and result analysis.

The guide details the architecture of the chaos engineering infrastructure, explaining how Chaos Mesh provides the orchestration capabilities for running chaos experiments across Kubernetes clusters. It describes the organization of experiment manifests and the validation frameworks that accompany experiment execution.

Key topics covered include the four primary categories of chaos experiments in the NEAM Platform: infrastructure experiments targeting compute and network layers, application experiments targeting service-level failures, data loss validation procedures ensuring data integrity, and performance degradation testing measuring system behavior under stress.

### Infrastructure Experiments Guide (INFRASTRUCTURE-EXPERIMENTS.md)

The Infrastructure Experiments Guide provides detailed procedures for configuring and executing infrastructure-level chaos experiments. These experiments target the foundational layers of the platform, including compute resources, network components, and storage systems.

The guide covers three primary categories of infrastructure chaos. Pod chaos experiments simulate application container failures including pod termination, container crashes, and pod unavailability. These experiments validate that the platform's self-healing mechanisms function correctly and that applications handle sudden termination gracefully.

Network chaos experiments simulate various network failure conditions including latency, packet loss, network partitions, and packet corruption. These experiments validate that the platform handles degraded network conditions correctly and that circuit breakers, timeouts, and retry logic function appropriately.

Stress chaos experiments subject compute resources to pressure, simulating CPU exhaustion, memory pressure, and storage I/O bottlenecks. These experiments help identify capacity issues, resource leaks, and improper resource handling that might cause failures under high load conditions.

Each experiment category includes complete configuration examples, execution procedures, safety considerations, and troubleshooting guidance. The guide emphasizes proper safeguards and blast radius management to ensure safe experiment execution.

### Application Experiments Guide (APPLICATION-EXPERIMENTS.md)

The Application Experiments Guide covers chaos experiments that target the service layer of the NEAM Platform. These experiments simulate failures at the application level, including HTTP service errors, latency injection, and service dependency unavailability.

The guide details HTTP latency injection experiments that simulate increased response times from HTTP services. These experiments are fundamental for validating timeout configurations, circuit breaker behavior, and retry logic throughout the service mesh. The guide explains how to configure latency injection with fixed and variable delays, and how to interpret the results to identify configuration issues.

HTTP error injection experiments simulate application-level errors by returning error responses to requests. These experiments validate that error handling paths throughout the system function correctly and that the system degrades gracefully when components encounter errors. The guide covers multiple error types including transient errors, authentication errors, authorization errors, and resource errors.

Service dependency failure experiments simulate the complete unavailability of external dependencies. These experiments validate that the platform maintains functionality when critical dependencies are unavailable, either through graceful degradation or by providing alternative functionality.

The guide also covers advanced configuration options including path-specific chaos, header-based chaos, and method-specific chaos, enabling targeted testing of particular endpoints or request types.

### Resilience Validation Guide (RESILIENCE-VALIDATION.md)

The Resilience Validation Guide documents procedures for validating system resilience during and after chaos engineering experiments. This guide focuses on data loss detection, consistency verification, and recovery validation to ensure that the platform maintains data integrity and operational continuity under failure conditions.

The guide introduces the data loss validation framework implemented through the data_loss_validator.py script. It explains the framework architecture, including the baseline manager, consistency checker, recovery verifier, and report generator components. The framework supports multiple data store technologies with appropriate consistency checking logic for each.

The guide provides detailed procedures for conducting data loss validation including pre-validation setup, baseline capture, during-chaos validation, and post-experiment verification. These procedures ensure comprehensive validation that data integrity is maintained throughout chaos experiments.

The guide explains how validation approaches are adapted for different consistency models including strong consistency, eventual consistency, and transactional consistency. Understanding these adaptations is essential for correctly interpreting validation results.

Guidance on handling data loss events includes root cause analysis procedures and remediation approaches at the application, configuration, and architecture levels. Recovery validation procedures confirm that systems return to fully consistent states after chaos events.

### Performance Testing Guide (PERFORMANCE-TESTING.md)

The Performance Testing Guide documents procedures for conducting performance degradation testing in the NEAM Platform. This guide explains how to combine load generation with chaos injection to measure how system performance changes under failure conditions.

The guide introduces the performance degradation test framework implemented through the performance_degradation_test.py script. It explains the framework architecture including the load generator, chaos orchestrator, metrics collector, and analyzer components. Understanding these components helps in designing effective tests and interpreting results correctly.

The guide details the performance metrics that should be collected during testing including latency metrics at multiple percentiles, throughput metrics, error metrics, and resource metrics. Comprehensive metrics collection enables detailed analysis of performance behavior under stress.

The guide explains how chaos injection is integrated with load testing including injection timing strategies, injection duration considerations, and chaos type selection. These factors significantly affect the information that tests can reveal.

Procedures cover test planning, configuration, pre-test verification, test execution, and post-test analysis. The guide also covers degradation analysis techniques, performance baseline establishment and maintenance, and advanced testing techniques for specialized scenarios.

## Implementation Files

The chaos engineering implementation consists of Kubernetes manifests and validation scripts that enable experiment execution and result collection.

### Kubernetes Manifests

The Kubernetes manifests for chaos experiments are located in the k8s/chaos directory. These manifests define the chaos experiments that can be applied to the cluster using kubectl.

The infrastructure-experiments.yaml file contains definitions for infrastructure-level chaos experiments including PodChaos experiments for pod and container failures, NetworkChaos experiments for network failures, and StressChaos experiments for resource pressure simulation.

The application-experiments.yaml file contains definitions for application-level chaos experiments using HTTPChaos resources to inject latency, errors, and other failures into HTTP traffic.

### Validation Scripts

The validation scripts located in the scripts/chaos directory provide automated validation capabilities that complement manual experiment observation and analysis.

The data_loss_validator.py script implements the data loss validation framework, providing capabilities for baseline establishment, consistency checking during chaos events, and recovery verification. The script supports multiple data store technologies through a plugin-based architecture.

The performance_degradation_test.py script implements the performance testing framework, combining load generation with chaos injection to measure system performance under failure conditions. The script provides comprehensive metrics collection and analysis capabilities.

## Getting Started

To begin using chaos engineering in the NEAM Platform, follow these steps to understand the approach and prepare for experiment execution.

First, review the Chaos Engineering Guide to understand the fundamental concepts and architecture. This foundational knowledge will help you understand the purpose and value of chaos engineering and how it fits into the overall reliability strategy of the platform.

Second, familiarize yourself with the experiment manifests and validation scripts. Review the example configurations to understand how experiments are defined and how validation is performed. Consider running initial experiments in a staging environment before conducting experiments in production.

Third, establish performance baselines for the systems you will be testing. Baselines provide the reference points against which degradation is measured, and they are essential for meaningful experiment results.

Fourth, start with simple experiments that validate basic functionality before progressing to more complex scenarios. A simple latency injection experiment provides a good starting point to validate that the infrastructure is working correctly and that you understand how to interpret results.

Finally, integrate chaos engineering into your regular operational and development practices. Regular chaos experiments provide ongoing validation of system resilience and help identify issues before they affect production users.

## Safety Considerations

Chaos engineering involves intentionally injecting failures into systems, and this must be done with appropriate safety measures to prevent unintended impact. The following principles should guide all chaos engineering activities.

Always implement safeguards that limit experiment impact and enable quick termination. Define blast radius restrictions through namespace and label selectors. Set duration limits to prevent experiments from running longer than necessary. Configure abort conditions that trigger automatic experiment termination based on system metrics.

For production environments, coordinate with relevant teams and ensure incident response procedures are in place. Consider scheduling experiments during lower-traffic periods. Communicate clearly about experiment scope, timing, and expected impact.

Maintain the ability to abort experiments immediately if unexpected behavior occurs. Ensure that all experiment participants know the abort procedures and have practiced executing them. Verify that systems return to normal after experiment termination.

## Additional Resources

For more information about chaos engineering principles and practices, consider the following resources that provide additional context and guidance.

The Principles of Chaos Engineering document establishes the fundamental principles that guide chaos engineering practice. These principles provide a philosophical foundation for understanding the purpose and approach of chaos engineering.

The Chaos Mesh documentation provides detailed information about the Chaos Mesh platform, including comprehensive resource definitions and configuration options. This documentation complements the NEAM Platform-specific guides with platform-level details.

Case studies from organizations that have implemented chaos engineering provide practical examples of how chaos engineering has revealed issues and improved system reliability. These examples can inspire and inform your own chaos engineering efforts.

## Support and Feedback

For questions about implementing chaos engineering in the NEAM Platform or to provide feedback on these documentation materials, contact the platform reliability team. Documentation improvements are welcomed and should be submitted through the standard contribution process.