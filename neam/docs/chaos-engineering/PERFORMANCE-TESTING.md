# Performance Testing Guide

This document provides comprehensive procedures for conducting performance degradation testing in the NEAM Platform. Performance testing combines load generation with chaos injection to measure how system performance changes under failure conditions, validating that the platform maintains acceptable performance even when components experience failures or when resources become constrained.

## Understanding Performance Degradation Testing

Performance degradation testing is a specialized form of chaos engineering that focuses on measuring and validating system performance characteristics under adverse conditions. While traditional chaos engineering validates behavioral correctness under failure, performance degradation testing validates that performance remains within acceptable bounds when the system is stressed.

The NEAM Platform implements performance degradation testing through the performance_degradation_test.py script, which combines load testing capabilities with chaos injection mechanisms. This combination enables measurement of how performance degrades under various failure scenarios and validation that degradation stays within acceptable thresholds.

Performance degradation testing serves multiple objectives within the overall reliability engineering strategy. First, it establishes quantitative bounds on system behavior under stress, enabling confident capacity planning and service level commitments. Second, it identifies performance bottlenecks and weaknesses that might not be apparent under normal load conditions. Third, it validates that degradation behavior is graceful, with the system reducing capability smoothly rather than failing abruptly.

The testing approach measures performance across multiple dimensions including response latency, throughput capacity, error rates, and resource efficiency. By measuring these dimensions under various failure conditions, the testing provides a comprehensive picture of how the system performs when things go wrong.

## Performance Testing Framework

The performance degradation test framework is implemented through the performance_degradation_test.py script located in the scripts/chaos directory. This framework integrates load generation, chaos injection, and performance measurement into a unified testing system.

### Framework Architecture

The performance degradation test framework consists of four primary components that work together to enable comprehensive performance testing under failure conditions. Understanding these components helps in designing effective tests and interpreting results correctly.

The load generator component produces synthetic traffic that exercises the system's critical paths. The load generator supports multiple traffic patterns including constant load for baseline measurement, step load for capacity identification, and burst load for stress testing. The generator is configurable in terms of request rate, concurrency, and request characteristics, enabling testing across a range of load conditions.

The chaos orchestrator component manages the injection of failure conditions during the test. The orchestrator coordinates with Chaos Mesh to activate chaos experiments at specified times and for specified durations. This coordination ensures that chaos injection occurs at predictable points in the test timeline, enabling clean comparison of performance before, during, and after chaos events.

The metrics collector component captures performance metrics throughout the test execution. The collector gathers both application-level metrics such as response times and error rates, and system-level metrics such as CPU utilization and memory consumption. The collector stores metrics with high temporal resolution to enable detailed analysis of performance changes.

The analyzer component processes collected metrics to identify performance degradation and assess whether degradation stays within acceptable bounds. The analyzer compares metrics during chaos periods against baseline metrics to quantify degradation, and evaluates degradation against configured thresholds to determine pass or fail status.

### Load Configuration

The load configuration determines what traffic pattern is generated during the test. Effective load configuration is essential for producing meaningful results that accurately reflect system behavior under realistic conditions.

The request rate configuration specifies how many requests per second the load generator should produce. The rate can be configured as a constant value or as a variable pattern. For baseline testing, a constant rate at the expected production load level is appropriate. For capacity testing, increasing the rate stepwise can identify the system's maximum capacity.

The concurrency configuration specifies how many simultaneous connections or requests the load generator should maintain. Higher concurrency tests the system's ability to handle parallel requests and reveals contention issues that might not appear at lower concurrency levels. The concurrency should be set based on expected peak traffic characteristics.

The request characteristics configuration specifies the nature of the requests generated. This includes request paths, payload sizes, and any parameters or headers that should be included. The characteristics should match the expected production traffic profile to ensure that test results are representative of real-world behavior.

### Traffic Patterns

Different traffic patterns reveal different aspects of system behavior. Selecting the appropriate pattern depends on the testing objectives and the system characteristics being investigated.

The constant load pattern maintains a steady request rate throughout the test. This pattern is useful for establishing baseline performance metrics and for observing the steady-state impact of chaos events. The constant load should be set to a level that exercises the system meaningfully without immediately overwhelming it.

The step load pattern increases the request rate in discrete steps, holding each rate level for a period before increasing. This pattern is useful for identifying system capacity limits and for observing how performance degrades as load increases. Each step should be held long enough for the system to reach steady state at that load level.

The burst load pattern produces short periods of high request rate followed by lower rate periods. This pattern is useful for testing the system's ability to handle traffic spikes and for measuring recovery behavior after burst periods. The burst characteristics should match the expected traffic spike patterns in production.

## Performance Metrics

Comprehensive performance measurement requires capturing metrics across multiple dimensions. This section describes the key metrics that should be collected and analyzed during performance degradation testing.

### Latency Metrics

Latency metrics measure the time required to process requests. These metrics are fundamental indicators of system performance and user experience. Latency should be measured at multiple percentiles to capture both typical and worst-case behavior.

The mean latency represents the average response time across all requests. While this metric provides a general indication of performance, it can be misleading if the distribution is skewed, as a small number of very slow requests can pull the mean up without affecting the median.

The median latency (p50) represents the response time at the 50th percentile, meaning half of requests complete faster than this time and half take longer. The median is less affected by outliers than the mean and provides a good indication of typical user experience.

The p95 latency represents the response time at the 95th percentile, meaning 95% of requests complete faster than this time. This metric captures the experience of the vast majority of users and is often used in service level objectives.

The p99 latency represents the response time at the 99th percentile, capturing the experience of the slowest 1% of requests. This metric is important for identifying tail latency issues that affect a small but significant number of users.

### Throughput Metrics

Throughput metrics measure the system's capacity to process requests. These metrics indicate how much traffic the system can handle and how capacity changes under failure conditions.

The request throughput measures the number of requests processed per second. This metric should remain stable under normal conditions and may decrease under failure conditions as the system struggles to process requests.

The byte throughput measures the volume of data processed per second, including both request and response bytes. This metric is important for understanding network and I/O capacity requirements.

The throughput capacity represents the maximum throughput the system can sustain. This capacity typically decreases under failure conditions, and measuring the capacity reduction provides a quantitative indication of resilience.

### Error Metrics

Error metrics measure the frequency and types of errors occurring during the test. These metrics indicate how well the system handles failure conditions and whether errors are handled gracefully.

The error rate measures the percentage of requests that result in errors. This rate should remain low under normal conditions and may increase under failure conditions. The acceptable error rate depends on the application and service level objectives.

The error types distribution breaks down errors by their nature, such as timeout errors, connection errors, and application errors. This breakdown helps identify the specific failure modes occurring during the test.

The error burst patterns identify whether errors occur in bursts or are spread evenly over time. Burst patterns may indicate specific failure triggers or threshold effects in the system's error handling.

### Resource Metrics

Resource metrics measure the utilization of system resources including CPU, memory, disk, and network. These metrics help explain performance behavior and identify resource constraints affecting performance.

CPU utilization measures the percentage of available CPU capacity being used. High CPU utilization often correlates with increased latency and reduced throughput. Understanding CPU behavior under failure conditions helps identify capacity constraints.

Memory utilization measures the percentage of available memory being used. Memory pressure can cause increased garbage collection activity, swapping, or out-of-memory conditions that affect performance.

Disk I/O metrics measure read and write operations and queue depths. Disk bottlenecks can cause significant latency increases and throughput limitations.

Network metrics measure bandwidth utilization and error rates. Network constraints can limit communication between services and affect distributed system performance.

## Chaos Injection During Testing

Performance degradation testing requires coordinated injection of failure conditions during load testing. This section describes how chaos injection is integrated with load testing.

### Injection Timing

The timing of chaos injection significantly affects the information that can be derived from the test. Different timing approaches serve different testing objectives.

Pre-load injection applies chaos before load testing begins, allowing observation of how the system recovers from failure under load. This approach is useful for validating recovery procedures and measuring recovery time.

Mid-load injection applies chaos during steady-state load testing, allowing direct comparison of performance before, during, and after the chaos event. This approach is most common for performance degradation testing as it provides clear before-and-after comparison.

Step-load with chaos combines increasing load patterns with chaos injection, allowing observation of how chaos effects vary at different load levels. This approach is useful for understanding capacity versus resilience trade-offs.

### Injection Duration

The duration of chaos injection affects the observations that can be made. Shorter durations provide cleaner comparisons but may not reveal sustained effects. Longer durations reveal steady-state behavior under chaos but require more test time.

Short-duration injection (seconds to a minute) reveals immediate system response to failure conditions and validates that the system correctly detects and responds to failures. This duration is appropriate for testing error handling and initial response behavior.

Medium-duration injection (minutes) reveals how the system behaves under sustained failure and whether steady-state degradation is different from initial response. This duration is appropriate for most performance degradation testing.

Long-duration injection (tens of minutes or longer) reveals long-term effects such as resource accumulation, connection pool exhaustion, and gradual degradation. This duration is appropriate for specialized testing of sustained failure scenarios.

### Chaos Selection

The type of chaos injected should be selected based on the failure scenarios being investigated. Different chaos types affect different aspects of system performance.

Network chaos such as latency, loss, and partition affects inter-service communication and can reveal issues with timeout configuration, circuit breaker behavior, and retry logic.

Resource chaos such as CPU and memory stress affects processing capacity and can reveal capacity constraints and resource management issues.

Application chaos such as service errors affects request processing and can reveal error handling behavior and fallback effectiveness.

## Test Procedures

This section provides detailed procedures for conducting performance degradation tests.

### Test Planning

Effective performance testing requires thorough planning to ensure that tests produce meaningful results and that the testing process is safe for the system under test.

Define clear test objectives before beginning test design. The objectives might include validating performance under specific failure scenarios, measuring degradation thresholds, or identifying performance bottlenecks. Each objective should have specific success criteria that can be evaluated after the test.

Identify the specific chaos scenarios to be tested based on the objectives. Each scenario should be documented with the expected impact on the system and the metrics that will indicate success or failure. Consider testing multiple scenarios to build a comprehensive picture of performance characteristics.

Determine appropriate load levels for the test. The load should be sufficient to exercise the system meaningfully but not so high that the system is already near capacity. Baseline load at expected production levels is a good starting point, with higher loads used for capacity-focused testing.

### Test Configuration

Configure the test parameters based on the test plan. The configuration includes load parameters, chaos parameters, and metrics collection parameters.

Load configuration specifies the traffic pattern, request rate, concurrency, and request characteristics. These parameters should match the testing objectives and the expected production traffic profile.

Chaos configuration specifies the chaos experiments to be injected, the timing and duration of injection, and any specific parameters for each chaos type. Each chaos experiment should be configured to simulate realistic failure conditions.

Metrics configuration specifies which metrics to collect, the collection frequency, and the storage location for results. The configuration should balance the desire for comprehensive metrics against the overhead of metrics collection.

### Pre-Test Verification

Before beginning the test, verify that all components are properly configured and ready for testing. This verification prevents wasted test time and ensures valid results.

Verify that the system under test is in a known good state with no residual effects from previous tests or operations. Consider running a brief smoke test to confirm basic functionality.

Verify that the load generation infrastructure is properly configured and can produce the expected traffic pattern. Run a brief load test without chaos to confirm that the load generator is functioning correctly.

Verify that the metrics collection infrastructure is properly configured and is receiving data. Confirm that key metrics are being collected and stored correctly.

Verify that chaos experiment configurations are valid and can be applied. Test that the chaos manifests can be applied without errors.

### Test Execution

Execute the test according to the planned procedure. During execution, monitor both the test progress and system behavior to ensure the test is proceeding as expected and to detect any issues that require intervention.

Begin with a baseline period during which load is applied without chaos. This baseline period establishes the performance reference against which degradation will be measured. The baseline period should be long enough to reach steady state.

Apply chaos according to the test plan. Monitor system behavior during chaos injection, watching for unexpected responses or rapid degradation. Have abort procedures ready in case the test needs to be stopped immediately.

Continue load testing through the chaos period and into a recovery period. The recovery period allows observation of how the system recovers after the chaos event concludes.

Throughout the test, document observations including any anomalies, unexpected behaviors, or issues that arise. This documentation supports post-test analysis.

### Post-Test Analysis

After the test completes, analyze the collected metrics to evaluate performance degradation and assess whether the test objectives were achieved.

Compare metrics during chaos periods against baseline metrics to quantify degradation. Calculate the percentage increase in latency, decrease in throughput, and change in error rates. These comparisons provide quantitative measures of resilience.

Evaluate degradation against success criteria defined in the test plan. Determine whether degradation was within acceptable bounds or exceeded thresholds that would indicate resilience issues.

Identify any anomalies or unexpected behaviors observed during the test. These findings may indicate issues that require further investigation or remediation.

Document the test results in a comprehensive report that includes test configuration, observations, analysis, and conclusions. This report serves as a reference for understanding system performance characteristics.

## Degradation Analysis

Degradation analysis interprets test results to understand how the system performs under failure conditions and to identify areas for improvement.

### Degradation Quantification

Quantifying degradation requires comparing metrics during chaos periods against baseline metrics. The comparison should be made for each relevant metric to build a comprehensive picture of performance impact.

Latency degradation is quantified as the percentage increase in response times at various percentiles. For example, if baseline p99 latency is 100ms and p99 latency during chaos is 300ms, the degradation is 200% (or a 3x increase).

Throughput degradation is quantified as the percentage decrease in request processing capacity. For example, if baseline throughput is 1000 requests per second and throughput during chaos is 700 requests per second, the degradation is 30%.

Error rate degradation is quantified as the increase in error rate from baseline. For example, if baseline error rate is 0.1% and error rate during chaos is 5%, the degradation is 4.9 percentage points.

### Threshold Evaluation

Evaluate whether degradation exceeds defined thresholds that would indicate unacceptable performance. Thresholds should be defined based on service level objectives and business requirements.

Define acceptable degradation thresholds for each metric. For example, latency degradation might be acceptable up to 200% increase, while error rate increase might be acceptable up to 1 percentage point.

Compare observed degradation against thresholds to determine pass or fail status. Metrics that exceed thresholds indicate areas where resilience improvements are needed.

Consider the combined effect of multiple metrics degrading simultaneously. A system might meet individual thresholds but still provide unacceptable user experience when multiple metrics are degraded.

### Root Cause Analysis

When significant degradation or unexpected behavior is observed, conduct root cause analysis to understand why the degradation occurred. This analysis informs remediation efforts.

Begin by examining the relationship between the chaos injection and the observed degradation. Determine whether the degradation is a direct expected consequence of the chaos or an unexpected secondary effect.

Examine resource metrics to identify resource constraints that might be causing or contributing to degradation. CPU exhaustion, memory pressure, and I/O bottlenecks are common causes of performance degradation.

Examine application metrics to identify application-level issues such as connection pool exhaustion, thread starvation, or inefficient processing that might cause or exacerbate degradation.

The root cause analysis should produce a clear understanding of why degradation occurred and what factors contributed to the degradation. This understanding guides remediation efforts.

## Performance Baselines

Performance baselines provide reference points for comparing performance during chaos experiments. Well-established baselines enable meaningful degradation measurement.

### Baseline Establishment

Performance baselines should be established under stable, well-understood conditions. The baseline establishment process captures the normal performance characteristics of the system.

Select a load level that represents typical production traffic. This load should be high enough to exercise the system meaningfully but not so high that the system is already stressed.

Allow the system to reach steady state at the selected load level before capturing baseline metrics. Steady state is indicated by stable latency, throughput, and resource utilization over time.

Capture baseline metrics over a sufficient duration to establish stable average values and understand normal variation. The duration should include enough requests to establish stable percentile values.

Document the baseline configuration including the load level, system configuration, and any other relevant context. This documentation enables consistent baseline establishment for future tests.

### Baseline Maintenance

Performance baselines should be updated periodically to reflect changes in the system. Outdated baselines may not provide accurate comparisons for current system behavior.

Establish a schedule for baseline updates based on the rate of system change. Systems undergoing frequent changes may require more frequent baseline updates. Stable systems may require less frequent updates.

Trigger immediate baseline updates for significant changes such as major releases, infrastructure changes, or significant configuration modifications. These changes are likely to affect performance characteristics.

Maintain a history of baselines to enable tracking of performance trends over time. This history supports analysis of whether performance is improving or degrading and where improvement efforts are focused.

### Baseline Utilization

Baselines are utilized by comparing metrics during chaos experiments against the established baseline. This comparison quantifies degradation and enables objective evaluation of resilience.

When running chaos experiments, apply the same load configuration used during baseline establishment. This ensures that degradation is due to the chaos injection rather than load differences.

Compare metrics at matching load levels to isolate the effect of chaos from the effect of load variation. Load and chaos have independent effects on performance, and comparing at matching loads enables clean separation of these effects.

Utilize baseline variation data when evaluating degradation. Performance naturally varies within bounds even under stable conditions, and degradation should be evaluated against this natural variation rather than fixed thresholds alone.

## Best Practices

The following best practices have been developed through experience with performance degradation testing and should guide testing activities.

Isolate performance testing from production traffic to prevent interference and to enable safe injection of failure conditions. Use dedicated test environments or carefully scheduled testing windows in shared environments.

Automate performance testing to enable regular and consistent measurement of performance characteristics. Automated tests can be run on schedules or triggered by system changes, providing continuous visibility into performance behavior.

Correlate application metrics with system metrics to understand the relationship between resource utilization and application performance. This correlation helps identify the root causes of performance degradation and guides capacity planning.

Iterate on test scenarios based on findings to continuously improve test coverage. Each test may reveal new scenarios to explore or areas where more detailed testing is needed.

Integrate performance testing into the development process to catch performance regressions early. Include performance testing as part of the release process for significant changes.

## Advanced Techniques

This section describes advanced techniques for performance degradation testing that may be used for specialized testing needs.

### Capacity Testing Under Chaos

Capacity testing under chaos combines increasing load patterns with chaos injection to identify how system capacity changes under failure conditions. This testing reveals whether capacity reserves are sufficient to maintain acceptable performance during failures.

The test begins with load at typical production levels and increases load stepwise while chaos is active. At each load level, measure throughput and latency to determine whether the system can sustain that load under failure conditions.

The capacity under chaos is compared against the baseline capacity to quantify the capacity reduction caused by the failure scenario. This comparison provides input for capacity planning and resilience investment decisions.

### Graceful Degradation Testing

Graceful degradation testing validates that the system reduces capability smoothly rather than failing abruptly when stressed. This testing ensures that users receive acceptable service even when the system cannot provide full functionality.

The test measures multiple aspects of system behavior as stress increases: whether latency increases gradually or suddenly, whether errors increase gradually or in step fashion, and whether partial functionality is maintained.

Graceful degradation is evaluated based on the shape of the degradation curve and the behavior at various stress levels. Ideally, degradation is gradual and predictable, allowing users to continue getting value from the system even under stress.

### Recovery Pattern Testing

Recovery pattern testing focuses specifically on how the system recovers after chaos events conclude. This testing validates that the system can return to normal operation and identifies any persistent effects of chaos events.

The test includes extended monitoring after chaos injection stops to observe the recovery process. Metrics are tracked through the recovery period to measure recovery time and verify complete recovery.

Recovery pattern testing also validates that no data or state is lost or corrupted during the chaos event and subsequent recovery. Integration with data loss validation provides comprehensive verification of recovery behavior.