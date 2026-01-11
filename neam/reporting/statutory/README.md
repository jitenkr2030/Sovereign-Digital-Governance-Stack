# Statutory Reporting Service

This package handles the generation and management of mandatory statutory reports for parliamentary committees and the Ministry of Finance.

## Overview

The Statutory Reporting Service ensures complete compliance with regulatory reporting requirements by automatically gathering data from across the NEAM platform and formatting it according to established government standards. The service manages the complete report lifecycle from data collection through generation, review, approval, and distribution.

## Report Types

### Parliamentary Reports
Parliamentary reports are generated for legislative oversight committees and include:
- Quarterly economic performance reviews
- Annual economic outlook statements
- Intervention effectiveness assessments
- Special investigation reports

### Finance Ministry Reports
Finance ministry reports support budgetary and fiscal planning:
- Monthly revenue and expenditure reports
- Quarterly fiscal updates
- Annual financial statements
- Intervention cost analyses

## Key Features

### Automated Report Generation
The service automatically collects required data from:
- Intervention management system
- Real-time monitoring databases
- Forensic evidence records
- Historical archives

Reports are generated using configurable templates that ensure consistency and compliance with regulatory formats.

### Multi-Format Export
Generated reports can be exported in multiple formats:
- PDF: Official document format for submission
- Excel: Spreadsheet format for detailed analysis
- CSV: Data format for further processing
- JSON: Machine-readable format for system integration

### Version Control
All reports maintain complete version history including:
- Draft versions during preparation
- Review versions with annotations
- Final approved versions
- Amendment records for corrections

### Approval Workflow
The service implements a review and approval workflow:
1. Report generated as draft
2. Submitted for review by authorized personnel
3. Reviewers can approve or request modifications
4. Approved reports are finalized and archived

## Architecture

### Data Collection Pipeline
1. Report template determines required data
2. Data collector queries relevant sources
3. Data is validated for completeness
4. Report generator assembles content
5. Report is formatted according to template

### Export Pipeline
1. Export request received with format specification
2. Report content converted to target format
3. Document properties applied (headers, footers, etc.)
4. File encrypted if required
5. Export metadata recorded for audit trail

## Usage

### Generate Parliamentary Report
```go
config := statutory.ParliamentReportConfig{
    SessionNumber:    15,
    CommitteeName:    "Economic Affairs Committee",
    RequiredSections: []string{"overview", "interventions", "recommendations"},
    DueDate:          time.Now().Add(7 * 24 * time.Hour),
}
report, err := service.GenerateParliamentReport(ctx, config)
```

### Generate Finance Ministry Report
```go
config := statutory.FinanceMinistryReportConfig{
    MinistryRef:    "MOF-2024-Q3",
    DepartmentCode: "ECO-001",
    ReportCode:     "Q3-2024",
    RequiredFields: []string{"revenue", "expenditure", "intervention_costs"},
    SubmissionDate: time.Now().Add(14 * 24 * time.Hour),
}
report, err := service.GenerateFinanceMinistryReport(ctx, config)
```

### Request Export
```go
export, err := service.RequestExport(ctx, reportID, "pdf", "analyst@government.gov")
```

## Dependencies
- ClickHouse: Primary data source for metrics
- PostgreSQL: Report metadata and workflow state
- Redis: Caching for report templates
- Kafka: Async report generation queue
