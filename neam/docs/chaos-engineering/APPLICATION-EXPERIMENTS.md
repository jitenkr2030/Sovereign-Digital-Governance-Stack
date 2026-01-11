# Application Experiments Guide

This document provides comprehensive procedures for configuring and executing application-level chaos experiments in the NEAM Platform. Application experiments target the service layer, simulating failures that occur at the application level including HTTP service errors, latency injection, and service dependency unavailability. These experiments validate that services handle errors gracefully, maintain acceptable behavior when dependencies fail, and implement proper resilience patterns throughout the system.

## Understanding Application Chaos

Application chaos engineering focuses on validating the behavior of application services under failure conditions. Unlike infrastructure chaos that targets the underlying compute and network layers, application chaos operates at the HTTP and service protocol level, allowing precise control over failure scenarios that simulate real-world application failures.

Application experiments are particularly valuable for validating resilience patterns such as circuit breakers, retry logic, timeouts, and fallback behaviors. These patterns are essential for building resilient distributed systems, but they can only be validated by simulating the failure conditions they are designed to handle. Application chaos experiments provide the controlled injection of these failure conditions.

The application chaos capabilities in the NEAM Platform are implemented through Chaos Mesh's HTTPChaos resource, which enables sophisticated manipulation of HTTP traffic. This includes injecting latency, returning errors, and modifying responses, all of which allow comprehensive testing of application resilience patterns.

Application experiments should be conducted with careful consideration of the blast radius, as they directly affect service behavior. While infrastructure chaos affects resource availability, application chaos affects the behavior of services that are available, making it essential to validate that services degrade gracefully rather than failing completely when problems occur.

## HTTP Latency Injection

HTTP latency experiments simulate increased response times from HTTP services, validating that the platform correctly handles slow dependencies. These experiments are fundamental for validating timeout configurations, circuit breaker behavior, and retry logic throughout the service mesh.

### Latency Injection Fundamentals

Latency injection works by intercepting HTTP requests and adding configurable delay before forwarding them to the target service. The delay can be configured as a fixed value or as a variable range with specified minimum and maximum values. This flexibility allows simulation of various network conditions, from consistent latency degradation to variable conditions that more closely resemble real-world network behavior.

The latency injection mechanism operates at the sidecar proxy level for services deployed with service mesh, allowing fine-grained control over which services are affected and which traffic paths experience latency. For services without sidecar proxies, latency injection can be configured at the service level, affecting all incoming traffic.

When configuring latency injection, it is essential to understand the cumulative effect of latency across service dependencies. A latency injection of 500 milliseconds on a frequently-called service can cause cascading effects throughout the system, particularly if calling services have improperly configured timeouts. Always start with conservative latency values and observe system behavior before increasing to more aggressive values.

### Latency Experiment Configuration

The HTTPChaos resource with latency action provides comprehensive configuration options for latency injection experiments. The configuration includes the target selector to identify affected services, the delay specification with latency and jitter values, and the scope definition to limit which traffic is affected.

The following configuration demonstrates latency injection targeting the payment service with a fixed 200-millisecond delay:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: payment-service-latency
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: payment-service
  mode: all
  action: delay
  delay: "200ms"
  target: "service"
```

For more realistic latency simulation, configure variable latency with jitter using the following format:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: inventory-service-latency-variable
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: inventory-service
  mode: all
  action: delay
  delay:
    latency: 100ms
    jitter: 50ms
```

The jitter parameter adds variability to the latency, creating more realistic network conditions where latency varies between requests. The correlation parameter can be used to control the relationship between consecutive request latencies, with higher values creating more burst-like patterns.

### Latency Experiment Procedures

Before executing latency experiments, complete the pre-experiment checklist including verification of monitoring systems, notification of relevant teams, and documentation of expected behavior. Pay particular attention to timeout configurations in services that call the target service, as latency injection may cause these services to experience timeouts.

Execute the experiment by applying the HTTPChaos manifest. Monitor the experiment using the Chaos Mesh dashboard or kubectl commands to verify that latency injection is active. Observe the following key metrics during the experiment: response latency distributions for calling services, timeout rates in dependent services, circuit breaker states, and overall system throughput.

When latency experiments reveal timeout or circuit breaker issues, document the specific configurations that need adjustment. Common findings include timeouts that are too short to accommodate expected latency variation, missing circuit breaker configurations, and retry policies that exacerbate latency problems by immediately retrying slow requests.

## HTTP Error Injection

HTTP error injection experiments simulate application-level errors by returning error responses to requests. These experiments validate that error handling paths throughout the system function correctly and that the system degrades gracefully when components encounter errors.

### Error Injection Fundamentals

Error injection allows precise control over the error conditions returned to requestors. Unlike network-level failures that prevent requests from reaching the service, HTTP error injection returns valid HTTP responses with error status codes, simulating application-level failures that the service handles internally.

The HTTPChaos resource supports multiple error injection modes. The abort mode immediately terminates the request without forwarding it to the target service, returning an error response directly. The replace mode forwards the request to the target service but replaces successful responses with error responses. The mixed mode combines successful and error responses based on a configurable percentage.

Error injection is particularly valuable for testing retry logic, circuit breaker behavior, and fallback mechanisms. By controlling exactly which errors are returned and how frequently, these experiments provide consistent and repeatable validation of error handling code paths.

### Error Configuration

Error injection is configured through the HTTPChaos resource with appropriate action and error specification. The following configuration demonstrates returning 503 Service Unavailable errors from the checkout service:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: checkout-service-errors
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: checkout-service
  mode: all
  action: abort
  abort:
    code: 503
```

For testing error handling with replacement mode, where some successful responses are replaced with errors:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: user-service-mixed-errors
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: user-service
  mode: all
  action: replace
  replace:
    statusCode: 500
  percent: "10"
```

The percent field specifies the percentage of requests that should receive error responses. A value of 10 causes 10% of requests to receive error responses while the remaining 90% are processed normally. This allows testing error handling without completely disabling the service.

### Error Type Coverage

Comprehensive error injection testing should cover multiple error types to validate different error handling paths. The following error types should be included in testing:

Transient errors with 5xx status codes simulate temporary failures that might be resolved by retrying. These errors validate retry logic and circuit breaker configuration for transient failure scenarios. Common 5xx codes include 500 Internal Server Error, 502 Bad Gateway, 503 Service Unavailable, and 504 Gateway Timeout.

Authentication errors with 401 Unauthorized simulate authentication failures, validating that the system correctly handles authentication errors and that fallback authentication paths are invoked when available. These errors are particularly important for services that support multiple authentication methods.

Authorization errors with 403 Forbidden simulate authorization failures, validating that the system correctly restricts access when users attempt operations they are not permitted to perform. These errors ensure that authorization checks are enforced correctly even when other system components are experiencing failures.

Resource errors with 404 Not Found simulate missing resource scenarios, validating that the system handles missing resource errors correctly and provides appropriate responses to clients attempting to access non-existent resources.

### Error Handling Validation

When executing error injection experiments, validate that error handling behavior is consistent across dependent services. Check that retry logic correctly handles transient versus permanent errors, that circuit breakers open appropriately when error rates exceed thresholds, and that fallback paths are invoked when available.

Monitor the following during error injection experiments: error rates in the target service and dependent services, circuit breaker state transitions, retry attempt counts and success rates, and fallback path execution frequency. These metrics help identify whether error handling is working as expected and where improvements might be needed.

## Service Dependency Failures

Service dependency failure experiments simulate the complete unavailability of external dependencies. These experiments validate that the platform maintains functionality when critical dependencies are unavailable, either through graceful degradation or by providing alternative functionality.

### Dependency Failure Scenarios

Service dependency failures can occur for various reasons including infrastructure issues, deployment problems, capacity constraints, and upstream service outages. The impact of these failures depends on how critical the dependency is to the application's functionality and how well the application handles dependency unavailability.

Database dependency failures simulate scenarios where the primary database becomes unavailable. These experiments validate that the application correctly handles connection failures, that read replicas are used appropriately when available, and that write operations are handled correctly when the primary database is unavailable.

External API dependency failures simulate scenarios where third-party services become unavailable. These experiments validate that the application implements appropriate timeout and circuit breaker patterns, that cached data is used when available, and that functionality is degraded gracefully when external services are unavailable.

Message queue dependency failures simulate scenarios where message queuing systems become unavailable. These experiments validate that the application correctly handles queue unavailability, that in-flight messages are preserved appropriately, and that asynchronous processing can resume correctly when the queue becomes available again.

### Dependency Failure Configuration

Dependency failure experiments are configured using HTTPChaos with abort action targeting the specific dependency service. The following configuration demonstrates simulating unavailability of an external payment gateway:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: payment-gateway-failure
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: payment-gateway
  mode: all
  action: abort
  abort:
    code: 503
```

For more realistic failure simulation that mimics actual service failures rather than explicit aborts, use the delay action with very long delays that will cause timeouts in calling services:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: analytics-service-slow
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: analytics-service
  mode: all
  action: delay
  delay:
    latency: 60000ms
    jitter: 10000ms
```

### Fallback Behavior Validation

When conducting dependency failure experiments, the primary objective is to validate fallback and degradation behavior. The following validation checks should be performed during dependency failure experiments.

Verify that circuit breakers open correctly when the dependency becomes unavailable. Circuit breakers should open after a configurable number of failures, preventing additional requests from being sent to the failing dependency. Monitor circuit breaker state transitions during the experiment to validate proper configuration.

Validate that fallback paths are invoked when circuit breakers are open. If the application implements fallback logic, verify that the fallback provides acceptable functionality during the dependency outage. The fallback might use cached data, alternative services, or simplified functionality depending on the application design.

Check that the system degrades functionality appropriately when dependencies are unavailable. Users should receive clear indications when functionality is degraded, and the degraded functionality should be usable even if it provides fewer features than the full service.

Verify that the system recovers correctly when the dependency becomes available again. Circuit breakers should close after a configurable period, and the application should resume normal operation. Monitor recovery time and verify that no data or functionality was lost during the dependency outage.

## Comprehensive Experiment Procedures

This section provides detailed procedures for conducting comprehensive application chaos experiments that validate multiple aspects of application resilience.

### Multi-Service Experiment Design

Complex distributed systems often require multi-service experiments that simulate failures affecting multiple components simultaneously. These experiments validate that the system handles cascading failures and that resilience patterns work correctly across service boundaries.

When designing multi-service experiments, identify the critical paths that should be tested and the failure scenarios that are most likely to occur. Consider scenarios where multiple dependencies fail simultaneously, where partial recovery occurs with some dependencies still unavailable, and where failure conditions persist longer than expected.

Multi-service experiments should be executed with careful monitoring and documented abort procedures. The combined impact of multiple service failures can be difficult to predict, and the ability to abort the experiment quickly is essential for limiting potential impact.

### Experiment Execution Checklist

Before executing application chaos experiments, complete the following comprehensive checklist:

First, verify that all monitoring and observability tools are operational and configured to capture the relevant metrics. This includes metrics collection for response times, error rates, circuit breaker states, and fallback invocations. Ensure that log aggregation is capturing application logs with sufficient detail to analyze error handling behavior.

Second, confirm that all relevant teams are aware of the planned experiment. This includes development teams responsible for the affected services, operations teams responsible for system monitoring, and support teams who might receive customer inquiries related to experiment effects. Provide clear communication about the experiment scope, timing, and expected impact.

Third, review and document the abort conditions for the experiment. Define specific metrics or behaviors that would trigger immediate experiment termination. Ensure that all experiment participants know how to abort the experiment and verify that abort procedures have been tested.

Fourth, verify that the experiment configuration is appropriate for the target environment. Experiments in production environments should have minimal blast radius and short duration. Staging environment experiments can be more aggressive but should still follow safety guidelines.

### Real-Time Monitoring

During experiment execution, maintain continuous monitoring of key system metrics. The following metrics should be monitored in real-time and documented throughout the experiment:

Response time metrics including average, p95, and p99 latencies provide immediate feedback on how the system is performing under failure conditions. Significant increases in latency may indicate that the system is struggling to handle the failure scenario.

Error rate metrics for both the target service and dependent services indicate whether error handling is working correctly. Monitor error rates by error type to understand which error conditions are occurring most frequently.

Circuit breaker state transitions indicate whether resilience patterns are being triggered correctly. Log state transitions to understand the sequence of events during the experiment and validate that circuit breakers are configured appropriately.

Fallback invocation counts indicate whether fallback paths are being used and how frequently. High fallback invocation rates combined with acceptable user experience validate that fallback logic is working correctly.

### Post-Experiment Analysis

After experiment completion, conduct comprehensive analysis of the results. Compare observed behavior against the documented expected behavior to determine whether the experiment succeeded in its objectives.

Analyze any unexpected behaviors discovered during the experiment. These unexpected behaviors often reveal important insights about system behavior that were not anticipated during experiment design. Document unexpected behaviors thoroughly and create follow-up tasks to investigate and address any issues.

Review the metrics collected during the experiment to identify trends and patterns. Compare baseline metrics collected before the experiment against metrics collected during the experiment to quantify the impact of the failure scenario. This comparison helps prioritize improvements based on the magnitude of impact.

Document lessons learned from the experiment, including both positive findings that validate good resilience patterns and negative findings that indicate areas for improvement. Share these lessons with the broader team to promote organizational learning.

## Advanced Configuration

This section covers advanced configuration options for application chaos experiments that enable more sophisticated testing scenarios.

### Path-Specific Chaos

Application chaos experiments can be scoped to specific HTTP paths, allowing targeted testing of particular endpoints or API methods. This is valuable for testing critical paths without affecting the entire service.

Path-specific chaos is configured using the path field in the HTTPChaos resource. The following configuration demonstrates latency injection only on the checkout endpoint:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: checkout-endpoint-latency
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: order-service
  mode: all
  action: delay
  delay:
    latency: 500ms
  target: "request"
  path: "/api/v1/checkout"
```

Path-specific chaos enables testing of critical endpoints with more aggressive failure injection while keeping less critical endpoints operating normally. This approach limits blast radius while still providing meaningful validation of system behavior.

### Header-Based Chaos

Chaos experiments can be scoped based on HTTP headers, allowing targeted testing of requests with particular characteristics. This enables scenarios such as testing failure behavior only for requests from specific clients or with specific authentication tokens.

Header-based chaos is configured using the headers field in the HTTPChaos resource. The following configuration demonstrates error injection only for requests with a specific header:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: vip-user-error-injection
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: api-gateway
  mode: all
  action: abort
  abort:
    code: 503
  target: "request"
  headers:
    - x-user-tier: vip
```

Header-based chaos is particularly useful for testing scenarios that affect specific user segments or request types without impacting all users equally.

### Method-Specific Chaos

Chaos experiments can target specific HTTP methods, enabling validation of failure handling for particular types of operations. This is important because different HTTP methods often have different characteristics and failure handling requirements.

Method-specific chaos is configured using the method field in the HTTPChaos resource. The following configuration demonstrates latency injection only for POST requests:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: write-operation-latency
  namespace: chaos-mesh
spec:
  selector:
    namespaces:
      - production
    labelSelectors:
      app: data-service
  mode: all
  action: delay
  delay:
    latency: 1000ms
  target: "request"
  method: "POST"
```

Method-specific chaos enables testing of write operations, which often have different consistency requirements and failure handling than read operations.

## Integration with Resilience Patterns

Application chaos experiments should be designed to validate specific resilience patterns implemented in the application. This section describes how to configure experiments that validate common resilience patterns.

### Circuit Breaker Validation

Circuit breaker patterns prevent cascading failures by stopping requests to failing services. Application chaos experiments can validate that circuit breakers are correctly configured and functioning properly.

To validate circuit breaker configuration, inject errors or latency until the circuit breaker should open, then verify that the circuit breaker state changes correctly. Continue the experiment to verify that the circuit breaker eventually closes after the recovery period, and validate that normal operation resumes correctly after the circuit closes.

The experiment should verify that circuit breaker thresholds are appropriate for the expected failure rates and that recovery timeouts allow sufficient time for the underlying issue to be resolved. Document the circuit breaker configuration and observed behavior to identify any configuration changes needed.

### Retry Logic Validation

Retry logic enables applications to automatically recover from transient failures by re-attempting failed operations. Application chaos experiments can validate that retry logic is implemented correctly and configured appropriately.

To validate retry logic, inject transient errors that should be retried successfully and verify that operations complete after the expected number of retry attempts. Also inject permanent errors that should not be retried and verify that the application does not enter retry loops for these errors.

The experiment should verify that retry backoff strategies are implemented to prevent overwhelming failing services, that retry counts are appropriate for the expected failure characteristics, and that retry timeouts are configured correctly. Document any retry logic issues discovered during the experiment.

### Timeout Configuration Validation

Timeout configurations prevent applications from waiting indefinitely for responses from dependencies. Application chaos experiments can validate that timeout configurations are appropriate for the expected response time distributions.

To validate timeout configuration, inject latency that exceeds the configured timeout and verify that the application correctly times out the request. Also inject latency that is just below the timeout threshold and verify that requests complete successfully. This validation helps identify timeouts that are too aggressive or too lenient.

The experiment should verify that timeout values account for normal latency variation, that timeout handling provides appropriate error responses to clients, and that resources are correctly cleaned up after timeouts occur. Document timeout configuration issues discovered during the experiment.

## Best Practices

The following best practices have been developed through experience with application chaos engineering and should guide all application chaos experiments.

Start with conservative experiments that validate basic functionality before progressing to more aggressive scenarios. A simple latency injection on a non-critical service provides a good starting point to validate that the experiment infrastructure is working correctly before testing more complex scenarios.

Use application chaos experiments to validate specific hypotheses about system behavior rather than exploring randomly. Each experiment should have a clear objective and expected outcome. This approach produces more actionable results and makes it easier to determine whether the experiment succeeded.

Combine multiple failure modes in a single experiment to validate that the system handles realistic scenarios where multiple problems occur simultaneously. Real-world failures often involve multiple components being affected, and the interaction between failures can reveal issues that single-failure experiments do not detect.

Integrate application chaos experiments into the continuous integration and deployment pipeline. Run automated chaos experiments as part of the release process to validate that new deployments do not introduce resilience regressions. This proactive approach catches issues before they affect production users.