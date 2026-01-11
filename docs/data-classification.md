# Data Classification

## Classification Framework

The data classification framework establishes the structure through which data assets are categorized based on sensitivity, regulatory requirements, and business criticality. This framework serves as the foundation for all data protection decisions, from access controls through retention policies to disposal procedures. Effective classification ensures that data receives protection appropriate to its sensitivity while enabling necessary access for authorized purposes.

The classification framework recognizes that data exists on a spectrum of sensitivity, from information that may be publicly disclosed without concern through information whose disclosure could cause severe harm to individuals, institutions, or national interests. This spectrum requires multiple classification levels rather than a simple binary public/private distinction. Each level has defined characteristics that determine the appropriate handling requirements.

Classification decisions consider multiple factors including the potential impact of unauthorized disclosure, legal and regulatory requirements, contractual obligations, and the sensitivity of the data subject. The framework provides guidance for consistent classification while recognizing that some classification decisions require judgment based on specific circumstances. Documentation requirements ensure that classification decisions are recorded and can be reviewed.

Classification is a continuous process rather than a one-time activity. Data may change classification over time as circumstances evolve, and the classification framework includes procedures for reclassification when appropriate. Regular review processes verify that classification remains appropriate as data ages and as the operating environment changes.

## Classification Levels

### Public Data

Public data includes information that is authorized for release to the general public without restriction. This classification encompasses information explicitly intended for public distribution as well as information that has been approved for publication through formal review processes. Public data may be freely shared through any channel and requires no special protection beyond standard handling procedures.

Characteristics of public data include the absence of harm from unauthorized disclosure, the absence of privacy concerns regarding data subjects, and the absence of security implications from public knowledge. Examples include published economic statistics, general policy documents, and public education materials. The classification acknowledges that public data may nevertheless require integrity protection to prevent malicious modification.

Public data handling requirements focus on availability and integrity rather than confidentiality. Access controls are minimal, allowing broad access to support the data's public purpose. Integrity controls prevent unauthorized modification that could mislead public users. Availability controls ensure that public data remains accessible through designated channels.

### Internal Data

Internal data includes information intended for use within the organization but not suitable for unrestricted public disclosure. This classification encompasses operational information, internal procedures, and aggregated data that does not individually identify sensitive subjects. Internal data may be shared within the organization without restriction but requires protection from external disclosure.

Characteristics of internal data include potential embarrassment or operational disruption from public disclosure, absence of severe harm from external knowledge, and the existence of a need-to-know basis for access. Examples include internal policy drafts, operational statistics, and non-sensitive administrative information. The classification recognizes that while external disclosure is not appropriate, the data does not require the most stringent protections.

Internal data handling requirements implement basic confidentiality protections. Access is limited to authenticated users with legitimate operational needs. External transmission requires appropriate encryption. Logging tracks access to support audit requirements. These controls provide reasonable protection while minimizing friction for legitimate internal use.

### Sensitive Data

Sensitive data includes information whose unauthorized disclosure could cause significant harm to individuals, organizations, or state interests. This classification encompasses personal data, proprietary business information, and operational intelligence that could be exploited by adversaries. Sensitive data requires substantial protection against both external and internal unauthorized access.

Characteristics of sensitive data include the potential for significant harm from disclosure, regulatory requirements for protection, and the existence of specific authorized recipients. Examples include individually identifiable financial records, internal analytical products, and system security configurations. The classification requires careful control of all access and comprehensive audit logging.

Sensitive data handling requirements implement strong confidentiality protections. Access is limited to specifically authorized users with documented need-to-know. All access is logged in tamper-evident fashion. Encryption is required for all storage and transmission. Physical security controls supplement logical protections for the most sensitive data.

### Restricted Data

Restricted data includes information whose unauthorized disclosure could cause severe harm to national security, economic stability, or individual welfare. This classification represents the highest sensitivity level and applies to the most critical information assets. Restricted data requires the most stringent protection measures and access controls.

Characteristics of restricted data include the potential for catastrophic harm from disclosure, specific legal protections, and extremely limited authorized access. Examples include classified intelligence products, highly sensitive financial surveillance data, and critical infrastructure vulnerability information. The classification requires individual authorization for each access with comprehensive audit trails.

Restricted data handling requirements implement maximum confidentiality protections. Access requires multiple independent authorizations and may require physical access controls. All access is monitored and recorded. Handling occurs only in controlled areas with appropriate security measures. Disposition requires specific authorization and follows documented procedures.

## Data Type Specifications

### Financial Data

Financial data encompasses all information related to monetary transactions, financial positions, and financial instruments. This data type includes individual transaction records, account information, regulatory filings, and aggregated financial statistics. Financial data requires careful classification given the sensitivity of financial information and the regulatory requirements governing its handling.

Individual transaction data records specific financial activities including parties, amounts, dates, and purposes. This data typically requires sensitive classification given its value for identity theft, fraud, and other criminal activities. Regulatory requirements may mandate specific handling procedures for transaction data. Retention requirements balance operational needs against privacy considerations.

Aggregated financial data combines individual transactions into statistical summaries. Classification depends on the granularity and sensitivity of the underlying data. Highly granular aggregates may retain individual sensitivities requiring sensitive classification. Broad statistical summaries may qualify for internal or even public classification depending on their characteristics.

Financial intelligence products combine raw financial data with analytical processing to identify patterns, risks, or violations. These products are typically classified as sensitive or restricted given their operational value and sensitivity. Intelligence products may reveal investigative interests and methodologies that require protection.

### Economic Data

Economic data encompasses all information related to economic activity, market conditions, and economic policy. This data type includes production statistics, price indices, employment data, and sector-specific indicators. Economic data requires classification balancing transparency objectives against concerns about market impact and national security.

Raw economic data includes observations of economic activity as collected through various sensing mechanisms. Classification depends on the data source and granularity. Data collected directly from economic actors may reflect commercial sensitivities. High-frequency indicators may have market impact implications requiring protection.

Processed economic data includes statistical products derived from raw data through analytical processes. These products are typically classified at internal or sensitive levels depending on their market sensitivity. Advance knowledge of economic releases could provide unfair trading advantages, requiring strict controls on pre-release access.

Economic intelligence combines economic data with analytical assessments to provide policy-relevant insights. This classification typically requires sensitive handling given its potential influence on economic decisions. Intelligence products may reveal analytical methodologies and policy deliberations requiring protection.

### Personal Data

Personal data encompasses all information relating to identified or identifiable individuals. This data type includes identification information, financial records, economic activities, and behavioral observations. Personal data requires careful classification balancing operational needs against privacy protections mandated by law.

Identifiable personal data includes information that directly identifies specific individuals through names, identification numbers, or other direct identifiers. This data typically requires sensitive classification given the potential for misuse and privacy violations. Legal requirements may mandate specific handling procedures including purpose limitation and retention constraints.

Indirectly identifiable personal data can be linked to specific individuals through reasonable effort even without direct identifiers. This data also requires sensitive classification given its re-identification potential. Aggregated data may present lower sensitivity when re-identification is impractical.

Sensitive personal data includes information requiring enhanced protection due to its sensitivity. This category includes financial hardship indicators, behavioral patterns, and other information whose disclosure could cause embarrassment or harm. Sensitive personal data typically requires restricted classification with enhanced access controls.

## Classification Procedures

### Initial Classification

Initial classification occurs when data is created or acquired. Creators and acquirers are responsible for applying appropriate initial classification based on the framework guidelines. Initial classification decisions are documented and can be reviewed for appropriateness. Training ensures that personnel understand classification requirements and apply them consistently.

Classification decisions consider the data type, source, content, and intended use. Creators evaluate sensitivity factors including the potential impact of unauthorized disclosure, the existence of regulatory requirements, and the expectations of data subjects. Classification levels are selected from the defined levels with clear justification.

Classification labels are applied consistently across data representations. The same logical data must receive consistent classification regardless of its physical form or location. Labels are preserved through processing and transformation to ensure that classification follows data throughout its lifecycle.

Derivative data inherits or may require new classification based on processing. Data derived from classified sources may retain the source classification or require higher classification depending on the processing. Aggregation, transformation, and analysis may change classification requirements.

### Review and Reclassification

Classification review ensures that classifications remain appropriate over time. Regular review cycles examine classified data to verify that initial classifications remain valid and that changes in circumstances do not require reclassification. Review frequency depends on data sensitivity and volatility.

Reclassification occurs when classification becomes inappropriate. Data may be downclassified when sensitivity decreases through the passage of time or changes in circumstances. Data may be upclassified when new information reveals higher sensitivity than initially recognized. Reclassification decisions are documented with justification.

Automatic expiration applies time-based declassification to certain data categories. Policies define which data types are subject to automatic expiration and the applicable timeframes. Expiration procedures verify that data no longer requires protection before releasing it from classification controls.

Disposal procedures ensure that classified data is securely destroyed when no longer needed. Disposal follows defined procedures appropriate to the classification level. Destruction verification ensures that data cannot be recovered through forensic techniques. Disposal is documented for compliance and audit purposes.

## Handling Requirements

### Access Controls

Access controls enforce classification-appropriate restrictions on data access. Control mechanisms are implemented at multiple levels including system, application, and data levels. The principle of least privilege ensures that users receive only the access necessary for their functions.

Role-based access control assigns permissions through role definitions rather than individual grants. Roles are designed to align with job functions and classification requirements. Role assignments are reviewed periodically to verify continued appropriateness. Excessive permissions trigger review and remediation.

Exception procedures address situations where standard access controls are inadequate. Exceptions require documented justification and approval from appropriate authorities. Exception duration is limited and subject to renewal. Exception usage is monitored and reviewed.

### Logging and Monitoring

Comprehensive logging supports audit requirements and security monitoring. All access to classified data is logged with sufficient detail to support forensic analysis and compliance verification. Log integrity is protected to prevent tampering or deletion.

Access logs record who accessed what data, when, and from where. Log entries include user identification, data identifiers, access type, and contextual information. Logs are retained for periods appropriate to the classification level and regulatory requirements.

Anomaly detection identifies unusual access patterns that may indicate misuse or compromise. Behavioral baselines establish normal access patterns for comparison. Alerts trigger investigation when anomalies are detected. Response procedures ensure timely handling of potential security events.

### Encryption Requirements

Encryption requirements ensure that classified data remains protected even when stored or transmitted outside controlled environments. Encryption standards specify algorithms, key lengths, and implementation requirements appropriate to each classification level.

At-rest encryption protects stored data from unauthorized access through media theft or unauthorized system access. Encryption is applied at the storage layer with keys managed through approved key management systems. Key recovery procedures ensure data availability while maintaining security.

In-transit encryption protects data during transmission between systems or to external parties. Transport layer security (TLS) or equivalent protocols provide encryption for network communications. Certificate management ensures that communication partners are authenticated. Forward secrecy protects past communications if keys are later compromised.
