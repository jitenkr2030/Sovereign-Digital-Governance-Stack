# Resilience Validation Guide

This document provides comprehensive procedures for validating system resilience during and after chaos engineering experiments in the NEAM Platform. Resilience validation encompasses data loss detection, consistency verification, and recovery validation to ensure that the platform maintains data integrity and operational continuity under failure conditions.

## Understanding Resilience Validation

Resilience validation is the process of verifying that systems maintain acceptable behavior when subjected to failures and stress conditions. While chaos engineering experiments inject failures into the system, resilience validation confirms that the system responds appropriately to those failures and recovers correctly when the failure conditions are removed.

The NEAM Platform implements resilience validation through a combination of automated validation frameworks and manual verification procedures. The automated validation is provided by the data_loss_validator.py script, which establishes baseline data states, monitors consistency during chaos events, and verifies recovery behavior. Manual procedures complement the automated validation by providing human oversight and validation of behaviors that are difficult to automate.

Resilience validation serves multiple purposes within the chaos engineering workflow. First, it confirms that chaos experiments do not cause unintended data corruption or loss. Second, it validates that recovery mechanisms function correctly after failure events. Third, it provides quantitative metrics that can be used to assess system resilience over time and compare different system configurations.

## Data Loss Validation Framework

The data loss validation framework is implemented through the data_loss_validator.py script located in the scripts/chaos directory. This framework provides automated checking of data consistency during chaos events, enabling detection of data integrity issues that might not be apparent through monitoring alone.

### Framework Architecture

The data loss validation framework is designed with a plugin-based architecture that enables support for multiple data stores and consistency checking approaches. The core framework provides common functionality including baseline establishment, consistency checking orchestration, and result reporting. Data store specific plugins implement the actual consistency checking logic for particular storage technologies.

The framework operates on the principle of establishing a known good state before chaos injection, then verifying that this state is preserved during and after the chaos event. This approach enables detection of any modifications to data that occur during the chaos experiment, whether those modifications are intended consequences of the experiment or unintended side effects.

Key components of the framework include the baseline manager, which captures the initial data state and generates checksums or hashes for verification; the consistency checker, which performs ongoing validation during chaos events; the recovery verifier, which confirms that the system returns to a consistent state after chaos events; and the report generator, which produces detailed documentation of validation activities and findings.

### Supported Data Stores

The data loss validation framework supports multiple data store technologies commonly used in the NEAM Platform. Each supported data store has a dedicated plugin that implements the appropriate consistency checking logic for that technology.

Relational databases are supported through plugins that verify table contents, checksums, and referential integrity. The plugins can be configured to verify entire tables or specific rows identified by primary key. For databases with replication, the plugins can also verify replication lag and consistency between replicas.

Key-value stores and document databases are supported through plugins that verify document contents and checksums. The plugins can verify individual documents or collections of documents based on query criteria. For distributed key-value stores, the plugins can verify consistency across nodes.

Message queues and streaming platforms are supported through plugins that verify message ordering and delivery. The plugins can verify that messages are not lost during chaos events and that message ordering is preserved. For persistent queues, the plugins can verify that messages are correctly persisted and can be recovered after failures.

### Baseline Establishment

Baseline establishment is the process of capturing the known good state of the data before chaos injection. This captured state serves as the reference point for all subsequent consistency verification during and after the chaos experiment.

The baseline establishment process begins with a quiescence period during which the system operates normally and baseline metrics are collected. During this period, the validation framework captures checksums for data items, sequence numbers for ordered data, and other identifiers that can be used for later verification. The baseline period should be long enough to capture normal system behavior, including any periodic data operations that occur.

For each data store being validated, the baseline process captures the following information: data content checksums using cryptographic hash functions, sequence numbers or timestamps for ordered data, replication status and lag measurements for replicated stores, and schema version information to ensure compatibility. This information is stored in a baseline artifact that can be referenced during the validation process.

The baseline should be established immediately before the chaos experiment to minimize the window during which external data modifications could invalidate the baseline. For experiments that run during normal operation, the baseline process should capture enough information to distinguish between changes caused by the chaos experiment and changes caused by normal business operations.

### Consistency Checking

Consistency checking during chaos events verifies that data integrity is maintained despite the failure conditions. The checking process runs continuously throughout the chaos experiment, comparing current data state against the established baseline to detect any inconsistencies.

The consistency checking process operates at multiple levels depending on the data store and validation requirements. At the storage level, the framework verifies that internal data structures remain intact and that data is correctly persisted. At the application level, the framework verifies that application data invariants hold true regardless of the chaos conditions. At the transaction level, the framework verifies that transactional integrity is maintained for systems that support transactions.

For eventually consistent systems, the consistency checking process accounts for temporary inconsistencies that may occur during normal operation. The framework defines acceptable inconsistency windows based on the consistency guarantees provided by the data store, and only flags inconsistencies that exceed these windows or that indicate data loss rather than temporary divergence.

The consistency checker produces continuous output during the chaos experiment, logging each consistency check result and flagging any inconsistencies detected. This log provides detailed information about when inconsistencies occurred, what data was affected, and what the nature of the inconsistency was. This information is valuable for post-experiment analysis and for understanding how the system behaves under failure conditions.

### Recovery Verification

Recovery verification confirms that the system returns to a fully consistent state after the chaos event concludes. This phase begins when the chaos experiment is stopped and continues until the system has fully recovered and all consistency checks pass.

During the recovery verification phase, the validation framework continues monitoring data consistency while allowing the system's normal recovery mechanisms to operate. The framework observes how quickly the system recovers, whether recovery is complete or some inconsistencies persist, and whether any manual intervention is required to complete recovery.

Recovery verification includes specific checks for common recovery scenarios. For replicated systems, the framework verifies that all replicas synchronize correctly after network partitions heal. For cached systems, the framework verifies that caches are correctly invalidated and repopulated. For partitioned systems, the framework verifies that data is correctly redistributed after node failures.

The recovery verification phase produces a recovery assessment that characterizes the recovery behavior. This assessment includes metrics such as time to consistency, data loss magnitude if any, and any manual intervention required. The assessment provides input for understanding system resilience and identifying areas for improvement.

## Validation Procedures

This section provides detailed procedures for conducting data loss validation during chaos engineering experiments. These procedures should be followed for all experiments where data integrity verification is required.

### Pre-Validation Setup

Before conducting data loss validation, complete the following setup procedures to ensure the validation framework is properly configured and ready for operation.

First, configure the validation framework by specifying the data stores to be validated and the validation parameters for each store. The configuration includes connection information for each data store, the specific data items or collections to validate, the validation frequency during the experiment, and the thresholds for flagging inconsistencies.

Second, verify that the validation framework has appropriate access to all data stores being validated. This includes network connectivity, authentication credentials, and any necessary permissions for reading data and performing consistency checks. Test the connectivity and access before proceeding to baseline establishment.

Third, coordinate with the chaos engineering team to align the validation timeline with the experiment timeline. Confirm when the chaos experiment will start and stop, and ensure that the validation framework will be capturing baseline data, running consistency checks, and performing recovery verification at the appropriate times.

### Baseline Capture Procedure

The baseline capture procedure establishes the reference state against which all subsequent validation will be compared. This procedure should be executed immediately before the chaos experiment to minimize the window for external changes.

Begin by pausing or reducing normal data operations if possible to minimize changes during baseline capture. If operations cannot be paused, document the expected changes so they can be distinguished from changes caused by the chaos experiment.

Execute the baseline capture command, which will iterate through all configured data stores and capture the necessary reference information. Monitor the baseline capture process to ensure it completes successfully and captures data from all configured stores.

After baseline capture completes, verify that the baseline artifact was created successfully and contains valid data for all configured stores. If any stores failed to capture baseline data, investigate and resolve the issue before proceeding with the chaos experiment.

### During Chaos Validation

During the chaos event, the validation framework runs continuous consistency checks against the configured data stores. Monitor the validation output throughout the experiment to detect any inconsistencies as they occur.

The validation output includes information about each consistency check performed, including the timestamp, the data store, the specific data items checked, and the result. Review this output regularly during the experiment to identify any patterns or concerning behaviors.

If significant inconsistencies are detected, evaluate whether the experiment should be continued or aborted. The decision should consider the nature and magnitude of the inconsistencies, whether the inconsistencies indicate data loss or just temporary divergence, and the experiment objectives and acceptable outcomes.

Document any inconsistencies detected during the experiment, including the timing, affected data, and apparent cause. This documentation supports post-experiment analysis and helps identify areas where the system or the validation process needs improvement.

### Post-Experiment Verification

After the chaos experiment concludes, the validation framework enters the recovery verification phase. This phase continues until the system has fully recovered and all consistency checks pass consistently.

Allow adequate time for system recovery before evaluating the results. The recovery time varies depending on the chaos scenario and system characteristics. Monitor the validation output during recovery to observe the healing process and identify any issues that prevent complete recovery.

After recovery is complete, perform a final comprehensive consistency check across all configured data stores. This check verifies that the system has returned to a fully consistent state and that no data was lost or corrupted during the experiment.

Compile the validation results into a comprehensive report that documents the baseline state, consistency checks performed during the experiment, any inconsistencies detected, recovery behavior observed, and the final validation status. This report serves as the authoritative record of the validation activity.

## Consistency Models and Validation

Different data stores provide different consistency guarantees, and the validation approach must be appropriate for the consistency model of each data store. This section explains how validation is adapted for different consistency models.

### Strong Consistency Validation

Strongly consistent data stores guarantee that all reads return the most recently written value. Validation of strongly consistent data focuses on detecting any deviation from this guarantee, which would indicate a data integrity issue.

For strongly consistent stores, the validation framework verifies that all reads return the expected values based on the established baseline. Any variation from the baseline indicates a potential data integrity issue that requires investigation. The validation is performed at a frequency that ensures timely detection of inconsistencies.

Strong consistency validation also includes verification of replication and distribution mechanisms. For replicated stores, the framework verifies that all replicas contain identical data. For partitioned stores, the framework verifies that data is correctly distributed across partitions according to the configured partitioning scheme.

### Eventual Consistency Validation

Eventually consistent data stores allow temporary inconsistencies during normal operation, with the guarantee that all replicas will eventually converge to the same state. Validation of eventually consistent data must account for this expected temporary divergence.

For eventually consistent stores, the validation framework defines acceptable inconsistency windows based on the store's consistency guarantees and typical convergence times. The framework monitors the inconsistency state during the chaos experiment and flags only those inconsistencies that exceed the expected windows or that indicate data loss rather than temporary divergence.

After chaos events, the validation framework monitors the convergence process to verify that the system returns to full consistency. The framework measures the time to convergence and verifies that all inconsistencies are resolved. Persistent inconsistencies that do not converge indicate a problem that requires investigation.

### Transactional Consistency Validation

Data stores that support transactions provide ACID guarantees that validation must verify. Transactional validation focuses on ensuring that all transaction properties are maintained during chaos events.

Atomicity validation verifies that all operations within a transaction complete together or are all rolled back. The validation framework tests this by observing transaction outcomes during chaos events and verifying that partial transactions do not leave data in inconsistent states.

Isolation validation verifies that concurrent transactions do not interfere with each other in ways that violate isolation level guarantees. The validation framework exercises concurrent operations during chaos events and verifies that the outcomes match the expected isolation level behavior.

Durability validation verifies that committed transactions survive system failures. The validation framework confirms that transactions committed before chaos events are present in the baseline and remain present during and after the experiment.

## Handling Data Loss

If data loss is detected during validation, a structured process should be followed to investigate the loss, assess its impact, and implement remediation measures.

### Loss Detection and Assessment

When data loss is detected, the first step is to assess the scope and magnitude of the loss. Determine how many data items are affected, what the data represents, and whether the loss is partial or complete. This assessment helps prioritize the investigation and determine the appropriate response.

For small amounts of data loss, particularly in eventually consistent stores where temporary divergence is expected, the loss may be acceptable if convergence occurs quickly. Document the loss event and continue monitoring to verify convergence.

For significant data loss, particularly in strongly consistent stores where any loss is unexpected, immediate investigation is required. The investigation should determine the root cause of the loss, whether the loss is ongoing or was a single event, and what data was affected.

### Root Cause Analysis

Root cause analysis of data loss events should follow established incident investigation procedures. The analysis seeks to understand not only the immediate cause of the data loss but also the underlying factors that allowed the loss to occur.

Common root causes of data loss during chaos events include improper error handling that fails to roll back incomplete transactions, race conditions that cause data to be written in unexpected orders, resource exhaustion that causes writes to fail silently, and configuration issues that prevent proper replication or backup.

The root cause analysis should produce a clear understanding of what went wrong and why. This understanding informs the remediation approach and helps prevent similar losses in the future.

### Remediation Approaches

Remediation of data loss vulnerabilities can take multiple forms depending on the root cause and the affected components. The appropriate remediation approach depends on the nature of the vulnerability and the effort required to address it.

Application-level remediation addresses vulnerabilities in application code that handles data operations. This might include fixing error handling logic, adding missing transaction boundaries, or correcting race conditions through proper synchronization.

Configuration-level remediation addresses vulnerabilities in data store configuration. This might include adjusting replication settings, modifying consistency levels, or enabling additional backup or recovery mechanisms.

Architecture-level remediation addresses fundamental design issues that cause data loss. This might include redesigning data models, changing data flow patterns, or implementing additional redundancy for critical data.

All remediation actions should be documented and validated through follow-up chaos experiments to confirm that the vulnerability has been addressed.

## Recovery Validation

Recovery validation confirms that the system can recover from chaos events and return to normal operation. This validation is essential for building confidence in the system's ability to heal itself after failures.

### Recovery Time Measurement

Recovery time measurement quantifies how long the system takes to return to normal operation after a chaos event. This metric is important for understanding system resilience and for setting appropriate service level objectives.

The recovery time is measured from the moment the chaos event ends until the system returns to a fully operational state. The fully operational state is defined based on specific criteria including response time within normal bounds, error rate at baseline levels, and all consistency checks passing.

Multiple recovery time measurements under different chaos scenarios provide a picture of the system's recovery characteristics. This picture helps identify which failure types the system recovers from quickly and which require longer recovery periods.

### Recovery Completeness Verification

Recovery completeness verification confirms that the system returns to a fully consistent state after recovery, not just an operational state. This verification ensures that no data inconsistencies persist after the system appears to be operating normally.

The verification process includes comprehensive consistency checks across all data stores. These checks are more thorough than the continuous checks performed during the experiment and may take longer to complete. The checks verify not only that data is present but that data relationships and invariants are maintained.

If the verification reveals persistent inconsistencies, the system has not fully recovered. Further investigation is needed to understand why the inconsistencies persist and what additional actions are required to achieve full recovery.

### Manual Recovery Procedures

Some chaos scenarios may require manual intervention to complete recovery. The validation framework should document any manual steps required and verify that those steps are effective when performed.

Manual recovery procedures should be documented in detail, including the steps to perform, the conditions under which they should be performed, and how to verify their success. These procedures should be tested during chaos experiments to ensure they work correctly when needed.

After manual recovery is performed, the validation framework should verify that the system is fully operational and consistent. This verification provides confidence that manual recovery was successful and that no issues remain.

## Validation Reporting

Comprehensive documentation of validation activities provides a record of system resilience characteristics and supports ongoing improvement of both the system and the validation process.

### Validation Report Structure

Validation reports should follow a consistent structure that includes all relevant information about the validation activity. The report structure enables easy comparison across experiments and supports trend analysis over time.

The report header includes metadata about the validation activity: the experiment identifier, the date and time of validation, the data stores validated, and the chaos scenario tested. This header provides context for interpreting the report contents.

The baseline section documents the captured baseline state, including checksums, sequence numbers, and other reference information. This section provides the reference point against which all consistency checks were compared.

The experiment section documents the chaos event timeline, including when the event started, when it ended, and any significant observations during the event. This section provides context for understanding when inconsistencies occurred.

The findings section documents any inconsistencies detected during the experiment, including the timing, affected data, and nature of each inconsistency. This section is the primary record of what went wrong during the experiment.

The recovery section documents the recovery process, including recovery time, any manual intervention required, and the final recovery status. This section provides information about how well the system heals itself.

### Report Distribution

Validation reports should be distributed to relevant stakeholders to ensure that findings are communicated and addressed. The distribution list should include development teams responsible for affected services, operations teams responsible for system reliability, and management stakeholders with interest in system resilience.

Reports should be distributed promptly after validation completes, while the information is still fresh and relevant. For significant findings, consider immediate communication rather than waiting for formal report distribution.

### Trend Analysis

Over time, validation reports enable trend analysis of system resilience characteristics. By comparing reports across experiments, patterns can be identified in how the system responds to different failure scenarios and how resilience changes as the system evolves.

Trend analysis should look at metrics such as recovery time over time, which indicates whether the system is becoming more or less resilient; frequency and severity of inconsistencies, which indicates whether stability is improving or degrading; and types of issues discovered, which indicates where improvement efforts should be focused.

Trend analysis results should inform prioritization of resilience improvement efforts and investment in system hardening. The analysis helps ensure that resources are directed toward the areas where they will have the greatest impact on system resilience.

## Integration with Chaos Engineering

Data loss validation is integrated into the broader chaos engineering workflow, providing data integrity verification as a complement to the behavioral validation performed through other means.

### Validation Timing

Data loss validation should be coordinated with chaos experiment timing to ensure that baseline capture, continuous checking, and recovery verification all align with the experiment lifecycle. The validation framework should be started before the experiment begins and continued after the experiment ends.

For long-running chaos experiments, periodic baseline updates may be needed to account for expected data changes during normal operation. The validation framework supports baseline refresh to accommodate these scenarios.

### Experiment Selection

Not all chaos experiments require comprehensive data loss validation. The decision to perform validation should be based on the potential impact on data integrity and the value of the validation results.

Experiments that modify data or affect data stores should include data loss validation. This includes experiments targeting databases, message queues, and any service that writes data.

Experiments that only affect read paths or non-data components may not require data loss validation. The validation effort should be proportional to the potential impact on data integrity.

### Continuous Validation

For critical data stores, consider implementing continuous validation that runs outside of formal chaos experiments. This continuous validation provides ongoing assurance of data integrity and can detect issues that arise outside of experiment contexts.

Continuous validation operates with lower overhead than experiment-focused validation, performing checks at less frequent intervals. The checks are designed to detect significant data integrity issues rather than transient inconsistencies.

If continuous validation detects issues, these findings should be investigated immediately and may trigger formal chaos experiments to understand the failure conditions more thoroughly.