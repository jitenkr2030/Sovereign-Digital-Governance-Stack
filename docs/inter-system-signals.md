# Inter-System Signals

## Signal Architecture

The inter-system signal architecture establishes the mechanisms through which CSIC and NEAM platforms exchange information while maintaining their operational independence. This architecture recognizes that effective sovereign governance requires correlation between financial and economic domains, but that this correlation must be achieved without creating the coupling and fragility that would result from shared databases or tight integration. The signal architecture provides the connective tissue that enables unified governance while preserving platform autonomy.

Signals represent a unidirectional flow of information from source to destination, with well-defined content, timing, and processing expectations. Unlike APIs that enable interactive transactions, signals are more akin to event notifications with attached payloads. This design choice simplifies both source and destination systems, reduces coupling, and creates clear boundaries between platform responsibilities. The signal model also provides natural audit points, as all inter-system communication occurs through defined signal channels.

The signal architecture implements a publish-subscribe pattern where sources publish signals to topics and destinations subscribe to topics of interest. This pattern enables flexible routing without requiring source systems to know about destination systems. Multiple destinations can receive the same signal, supporting broadcast of information that is relevant to multiple consumers. The publish-subscribe model also supports filtering and transformation at the topic level.

Quality of service guarantees ensure that signals are delivered reliably even in the face of transient failures. Signals are persisted until delivery confirmation prevents loss during destination unavailability. Ordering guarantees within signal categories ensure that related signals arrive in expected sequence. Delivery confirmation enables source systems to verify that signals have been received and processed.

## Signal Categories

### Financial Risk Signals

Financial risk signals originate in CSIC and provide NEAM with intelligence about financial sector conditions, risks, and developments. These signals enable NEAM to incorporate financial context into economic analysis and to understand how financial sector developments may affect real economic activity.

Transaction monitoring signals report patterns of concern detected in financial transaction monitoring. These signals include indicators of potential money laundering, fraud, or other financial crimes. Signal content includes pattern descriptions, affected entities, and severity assessments. These signals inform NEAM's understanding of shadow economy activities and their economic impacts.

Compliance enforcement signals report enforcement actions and their targets. These signals include information about violations detected, investigations conducted, and penalties imposed. Signal content includes entity identification, violation categories, and enforcement outcomes. These signals inform NEAM's understanding of regulatory pressure on economic sectors.

Financial stability signals report conditions that may indicate systemic risk. These signals include indicators of liquidity stress, counterparty concerns, or market dysfunction. Signal content includes affected institutions, risk indicators, and assessment of severity. These signals enable NEAM to incorporate financial stability considerations into economic monitoring.

### Macro-Economic Signals

Macro-economic signals originate in NEAM and provide CSIC with intelligence about real economic conditions, sector performance, and economic policy effects. These signals enable CSIC to incorporate economic context into financial surveillance and to understand how economic developments may affect financial sector risks.

Activity index signals report measures of overall economic activity. These signals include production indices, consumption indicators, and service sector measures. Signal content includes index values, sector breakdowns, and trend assessments. These signals inform CSIC's understanding of the economic backdrop for financial activity.

Sector performance signals report conditions in specific economic sectors. These signals include measures of sector output, employment, and investment. Signal content includes sector identification, performance metrics, and outlook assessments. These signals enable CSIC to target surveillance on sectors of concern.

Black economy signals report estimates of informal and underground economic activity. These signals include indicators of unreported economic activity and shadow market dimensions. Signal content includes activity estimates, methodology indicators, and confidence assessments. These signals inform CSIC's understanding of economic activity that may escape financial surveillance.

### Policy Signals

Policy signals represent bidirectional communication of regulatory intent and policy direction between platforms. These signals ensure that both platforms operate in alignment with broader governance objectives and that policy changes are reflected consistently across both systems.

Regulatory update signals communicate new or changed regulatory requirements. These signals originate in regulatory bodies and are routed to affected platforms. Signal content includes regulatory citations, effective dates, and implementation requirements. These signals trigger platform configuration changes and operational adjustments.

Intervention request signals communicate the need for targeted action based on policy priorities. These signals originate in policy bodies and request specific platform actions. Signal content includes target specifications, action requirements, and priority indications. These signals enable rapid response to emerging situations.

Policy coordination signals communicate alignment requirements between platforms. These signals originate in governance bodies and address coordination between CSIC and NEAM. Signal content includes coordination topics, alignment requirements, and implementation timelines. These signals ensure consistent platform behavior on cross-cutting issues.

## Signal Specifications

### Message Structure

Signal messages follow a structured format that enables consistent processing across the signal architecture. The structure includes metadata for routing and processing, a header for context, and a payload for the actual signal content. This consistent structure simplifies integration and ensures predictable behavior.

The metadata section includes fields used by the transport layer for routing and delivery confirmation. Metadata includes topic identification, message identification, timestamp, and priority. Transport systems use metadata for queue management and delivery optimization.

The header section includes contextual information about the signal. Header fields include source system identification, signal category, and classification level. Processing systems use header information for filtering, transformation, and access control.

The payload section contains the actual signal content. Payload structure varies by signal category but follows defined schemas for each signal type. Payloads are encoded in standard formats that support validation and transformation. Payload content is protected according to classification requirements.

### Topic Definitions

Topics define the categories of signals and their routing characteristics. Each topic has defined publishers, subscribers, and quality of service requirements. Topic definitions are maintained in a central registry that enables discovery and management.

Financial surveillance topics carry signals from CSIC financial monitoring functions. Topics are organized by monitoring domain including banking, payments, and crypto assets. Each topic has defined schema and semantics that subscribers can rely upon.

Economic monitoring topics carry signals from NEAM economic sensing functions. Topics are organized by economic sector and indicator type. Each topic has defined update frequency and latency expectations.

Policy coordination topics carry signals related to governance and coordination. Topics address regulatory updates, intervention requests, and alignment requirements. Access to policy topics is restricted to authorized platform functions.

### Schema Specifications

Schemas define the structure and semantics of signal payloads. Schemas are specified using standard definition languages and are versioned to support evolution. Schema validation ensures that signals conform to expectations before processing.

Financial signal schemas define the structure of financial risk and compliance signals. Schemas specify field names, types, and validation rules. Schema documentation explains the meaning and usage of each field.

Economic signal schemas define the structure of macro-economic and sector signals. Schemas balance generality with specificity to enable meaningful correlation while avoiding excessive detail.

Policy signal schemas define the structure of regulatory and coordination signals. Schemas emphasize clarity and actionability to ensure that policy signals are implemented correctly.

## Integration Patterns

### Real-Time Streaming

Real-time streaming delivers signals with minimal latency for time-sensitive communications. Streaming patterns use persistent connections and push delivery to enable immediate signal propagation. This pattern is appropriate for alerts, urgent notifications, and time-critical coordination.

Connection management maintains persistent connections between signal infrastructure and platform systems. Connection health monitoring detects and recovers from connection failures. Reconnection procedures restore streaming quickly after interruptions.

Flow control prevents overwhelming downstream systems with signal volume. Backpressure signals enable sources to adjust transmission rates. Buffer management prevents memory exhaustion during processing delays.

### Batch Processing

Batch processing delivers signals in groups for efficiency when real-time delivery is not required. Batching reduces processing overhead for high-volume but non-urgent signals. This pattern is appropriate for regular reports, statistical updates, and routine coordination.

Batch scheduling defines when batches are assembled and transmitted. Scheduling balances latency requirements against processing efficiency. Scheduling configurations can be adjusted based on signal characteristics and priorities.

Batch composition groups related signals for efficient transmission. Composition rules ensure that related signals are not split across batches inappropriately. Composition logic supports both automatic and manual batch formation.

### Request-Response

Request-response patterns enable subscribers to request specific information from signal sources. This pattern supports queries that cannot be addressed through pre-defined signals. Request-response is used sparingly to maintain the loose coupling that signals provide.

Query specifications define the information that can be queried and the response format. Query schemas balance flexibility with predictability to ensure meaningful responses.

Timeout management ensures that requests do not wait indefinitely for responses. Timeout values are calibrated based on expected processing time. Timeout handling provides appropriate feedback when responses are delayed.

## Governance of Signals

### Signal Approval

Signal creation and modification require formal approval processes. New signals must be justified, designed, and reviewed before deployment. Existing signals must be reviewed periodically to ensure continued relevance. Approval processes ensure that signals serve governance purposes and do not proliferate unnecessarily.

Business justification documents the need for a new signal and its expected use. Justification addresses what information the signal will provide, who will use it, and what decisions it will support. Justification review ensures that signals address genuine needs.

Technical review examines signal design for consistency with architecture and standards. Review addresses schema design, topic assignment, and integration requirements. Technical review ensures that signals can be implemented and operated reliably.

### Performance Monitoring

Signal performance is monitored to ensure that signals are delivered effectively. Monitoring tracks delivery latency, volume, and error rates. Performance data supports capacity planning and identifies issues requiring attention.

Latency tracking measures the time from signal generation to delivery. Latency metrics are tracked by signal category and destination. Latency degradation triggers investigation and remediation.

Volume monitoring tracks signal volumes across topics and time periods. Volume data supports capacity planning and anomaly detection. Unusual volume patterns trigger investigation.

Error tracking monitors signal transmission and processing failures. Error rates and types are tracked to identify systemic issues. Error investigation leads to remediation and process improvement.

### Version Management

Signal schemas evolve over time to reflect changing requirements. Version management ensures that evolution is controlled and that consumers can adapt to changes. Version practices balance the need for improvement against the cost of consumer disruption.

Version numbering identifies schema versions and indicates their relationship. Semantic versioning communicates the significance of changes. Version numbering enables consumers to identify and adapt to changes.

Change notification informs consumers of upcoming schema changes. Notification periods provide time for consumer adaptation. Change communication explains what is changing and how consumers should respond.

Compatibility management maintains backward compatibility where possible. Where breaking changes are necessary, migration support helps consumers transition. Compatibility practices minimize disruption from necessary evolution.
