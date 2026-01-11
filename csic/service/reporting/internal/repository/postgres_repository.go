package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"csic-platform/service/reporting/internal/db"
	"csic-platform/service/reporting/internal/domain"
)

// PostgresReportTemplateRepository implements ReportTemplateRepository for PostgreSQL
type PostgresReportTemplateRepository struct {
	db *db.Database
}

// NewPostgresReportTemplateRepository creates a new PostgreSQL report template repository
func NewPostgresReportTemplateRepository(database *db.Database) ReportTemplateRepository {
	return &PostgresReportTemplateRepository{db: database}
}

// Create creates a new report template
func (r *PostgresReportTemplateRepository) Create(ctx context.Context, template *domain.ReportTemplate) error {
	templateSchemaJSON, err := json.Marshal(template.TemplateSchema)
	if err != nil {
		return fmt.Errorf("failed to marshal template schema: %w", err)
	}
	
	generationRulesJSON, err := json.Marshal(template.GenerationRules)
	if err != nil {
		return fmt.Errorf("failed to marshal generation rules: %w", err)
	}
	
	outputFormatsJSON, err := json.Marshal(template.OutputFormats)
	if err != nil {
		return fmt.Errorf("failed to marshal output formats: %w", err)
	}

	query := `
		INSERT INTO report_templates (
			id, name, description, report_type, regulator_id, version,
			template_schema, generation_rules, output_formats, frequency,
			cron_schedule, is_active, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
	`

	_, err = r.db.ExecContext(ctx, query,
		template.ID, template.Name, template.Description, template.ReportType,
		template.RegulatorID, template.Version, templateSchemaJSON, generationRulesJSON,
		outputFormatsJSON, template.Frequency, template.CronSchedule, template.IsActive,
		template.CreatedBy,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create report template: %w", err)
	}
	
	return nil
}

// Update updates an existing report template
func (r *PostgresReportTemplateRepository) Update(ctx context.Context, template *domain.ReportTemplate) error {
	templateSchemaJSON, err := json.Marshal(template.TemplateSchema)
	if err != nil {
		return fmt.Errorf("failed to marshal template schema: %w", err)
	}
	
	generationRulesJSON, err := json.Marshal(template.GenerationRules)
	if err != nil {
		return fmt.Errorf("failed to marshal generation rules: %w", err)
	}
	
	outputFormatsJSON, err := json.Marshal(template.OutputFormats)
	if err != nil {
		return fmt.Errorf("failed to marshal output formats: %w", err)
	}

	query := `
		UPDATE report_templates SET
			name = $2, description = $3, report_type = $4, regulator_id = $5,
			version = $6, template_schema = $7, generation_rules = $8,
			output_formats = $9, frequency = $10, cron_schedule = $11,
			is_active = $12, updated_at = NOW()
		WHERE id = $1
	`

	_, err = r.db.ExecContext(ctx, query,
		template.ID, template.Name, template.Description, template.ReportType,
		template.RegulatorID, template.Version, templateSchemaJSON, generationRulesJSON,
		outputFormatsJSON, template.Frequency, template.CronSchedule, template.IsActive,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update report template: %w", err)
	}
	
	return nil
}

// Delete soft-deletes a report template
func (r *PostgresReportTemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE report_templates SET is_active = false, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete report template: %w", err)
	}
	
	return nil
}

// GetByID retrieves a report template by ID
func (r *PostgresReportTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ReportTemplate, error) {
	query := `
		SELECT id, name, description, report_type, regulator_id, version,
			template_schema, generation_rules, output_formats, frequency,
			cron_schedule, is_active, created_by, created_at, updated_at
		FROM report_templates WHERE id = $1
	`

	var template domain.ReportTemplate
	var templateSchemaJSON, generationRulesJSON, outputFormatsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&template.ID, &template.Name, &template.Description, &template.ReportType,
		&template.RegulatorID, &template.Version, &templateSchemaJSON, &generationRulesJSON,
		&outputFormatsJSON, &template.Frequency, &template.CronSchedule, &template.IsActive,
		&template.CreatedBy, &template.CreatedAt, &template.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get report template: %w", err)
	}

	if err := json.Unmarshal(templateSchemaJSON, &template.TemplateSchema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template schema: %w", err)
	}
	
	if err := json.Unmarshal(generationRulesJSON, &template.GenerationRules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal generation rules: %w", err)
	}
	
	if err := json.Unmarshal(outputFormatsJSON, &template.OutputFormats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal output formats: %w", err)
	}

	return &template, nil
}

// GetByRegulatorID retrieves all report templates for a regulator
func (r *PostgresReportTemplateRepository) GetByRegulatorID(ctx context.Context, regulatorID uuid.UUID) ([]*domain.ReportTemplate, error) {
	query := `
		SELECT id, name, description, report_type, regulator_id, version,
			template_schema, generation_rules, output_formats, frequency,
			cron_schedule, is_active, created_by, created_at, updated_at
		FROM report_templates WHERE regulator_id = $1 ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, regulatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to query report templates: %w", err)
	}
	defer rows.Close()

	return r.scanTemplates(rows)
}

// GetActiveTemplates retrieves all active report templates
func (r *PostgresReportTemplateRepository) GetActiveTemplates(ctx context.Context) ([]*domain.ReportTemplate, error) {
	query := `
		SELECT id, name, description, report_type, regulator_id, version,
			template_schema, generation_rules, output_formats, frequency,
			cron_schedule, is_active, created_by, created_at, updated_at
		FROM report_templates WHERE is_active = true ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active templates: %w", err)
	}
	defer rows.Close()

	return r.scanTemplates(rows)
}

// GetByReportType retrieves templates by report type
func (r *PostgresReportTemplateRepository) GetByReportType(ctx context.Context, reportType domain.ReportType) ([]*domain.ReportTemplate, error) {
	query := `
		SELECT id, name, description, report_type, regulator_id, version,
			template_schema, generation_rules, output_formats, frequency,
			cron_schedule, is_active, created_by, created_at, updated_at
		FROM report_templates WHERE report_type = $1 AND is_active = true ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, reportType)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates by type: %w", err)
	}
	defer rows.Close()

	return r.scanTemplates(rows)
}

// List retrieves paginated report templates
func (r *PostgresReportTemplateRepository) List(ctx context.Context, page, pageSize int) ([]*domain.ReportTemplate, int, error) {
	offset := (page - 1) * pageSize

	// Get total count
	countQuery := `SELECT COUNT(*) FROM report_templates WHERE is_active = true`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count templates: %w", err)
	}

	query := `
		SELECT id, name, description, report_type, regulator_id, version,
			template_schema, generation_rules, output_formats, frequency,
			cron_schedule, is_active, created_by, created_at, updated_at
		FROM report_templates WHERE is_active = true
		ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	templates, err := r.scanTemplates(rows)
	if err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// Search searches for templates by name or description
func (r *PostgresReportTemplateRepository) Search(ctx context.Context, query string, page, pageSize int) ([]*domain.ReportTemplate, int, error) {
	offset := (page - 1) * pageSize
	searchPattern := "%" + strings.ToLower(query) + "%"

	// Get total count
	countQuery := `SELECT COUNT(*) FROM report_templates WHERE is_active = true AND (LOWER(name) LIKE $1 OR LOWER(description) LIKE $1)`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, searchPattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count templates: %w", err)
	}

	sqlQuery := `
		SELECT id, name, description, report_type, regulator_id, version,
			template_schema, generation_rules, output_formats, frequency,
			cron_schedule, is_active, created_by, created_at, updated_at
		FROM report_templates WHERE is_active = true AND (LOWER(name) LIKE $1 OR LOWER(description) LIKE $1)
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, searchPattern, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search templates: %w", err)
	}
	defer rows.Close()

	templates, err := r.scanTemplates(rows)
	if err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// Activate activates a report template
func (r *PostgresReportTemplateRepository) Activate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE report_templates SET is_active = true, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to activate template: %w", err)
	}
	
	return nil
}

// Deactivate deactivates a report template
func (r *PostgresReportTemplateRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE report_templates SET is_active = false, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate template: %w", err)
	}
	
	return nil
}

// IncrementVersion increments the version of a template
func (r *PostgresReportTemplateRepository) IncrementVersion(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE report_templates SET version = version || '.1', updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment version: %w", err)
	}
	
	return nil
}

// scanTemplates helper function to scan templates from rows
func (r *PostgresReportTemplateRepository) scanTemplates(rows *sql.Rows) ([]*domain.ReportTemplate, error) {
	var templates []*domain.ReportTemplate

	for rows.Next() {
		var template domain.ReportTemplate
		var templateSchemaJSON, generationRulesJSON, outputFormatsJSON []byte

		err := rows.Scan(
			&template.ID, &template.Name, &template.Description, &template.ReportType,
			&template.RegulatorID, &template.Version, &templateSchemaJSON, &generationRulesJSON,
			&outputFormatsJSON, &template.Frequency, &template.CronSchedule, &template.IsActive,
			&template.CreatedBy, &template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		if err := json.Unmarshal(templateSchemaJSON, &template.TemplateSchema); err != nil {
			return nil, fmt.Errorf("failed to unmarshal template schema: %w", err)
		}
		
		if err := json.Unmarshal(generationRulesJSON, &template.GenerationRules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal generation rules: %w", err)
		}
		
		if err := json.Unmarshal(outputFormatsJSON, &template.OutputFormats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal output formats: %w", err)
		}

		templates = append(templates, &template)
	}

	return templates, nil
}

// PostgresGeneratedReportRepository implements GeneratedReportRepository for PostgreSQL
type PostgresGeneratedReportRepository struct {
	db *db.Database
}

// NewPostgresGeneratedReportRepository creates a new PostgreSQL generated report repository
func NewPostgresGeneratedReportRepository(database *db.Database) GeneratedReportRepository {
	return &PostgresGeneratedReportRepository{db: database}
}

// Create creates a new generated report
func (r *PostgresGeneratedReportRepository) Create(ctx context.Context, report *domain.GeneratedReport) error {
	outputFormatsJSON, err := json.Marshal(report.OutputFormats)
	if err != nil {
		return fmt.Errorf("failed to marshal output formats: %w", err)
	}

	customFieldsJSON, err := json.Marshal(report.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	query := `
		INSERT INTO generated_reports (
			id, template_id, template_name, report_type, regulator_id, regulator_name,
			status, version, title, description, period_start, period_end, report_date,
			generated_by, generation_method, triggered_by, output_formats,
			total_size, checksum, processing_time, retry_count,
			is_validated, validated_by, is_approved, approved_by,
			submission_status, retention_period, custom_fields,
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, NOW(), NOW())
	`

	_, err = r.db.ExecContext(ctx, query,
		report.ID, report.TemplateID, report.TemplateName, report.ReportType,
		report.RegulatorID, report.RegulatorName, report.Status, report.Version,
		report.Title, report.Description, report.PeriodStart, report.PeriodEnd, report.ReportDate,
		report.GeneratedBy, report.GenerationMethod, report.TriggeredBy, outputFormatsJSON,
		report.TotalSize, report.Checksum, report.ProcessingTime, report.RetryCount,
		report.IsValidated, report.ValidatedBy, report.IsApproved, report.ApprovedBy,
		report.SubmissionStatus, report.RetentionPeriod, customFieldsJSON,
		report.CreatedBy,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create generated report: %w", err)
	}
	
	return nil
}

// Update updates an existing generated report
func (r *PostgresGeneratedReportRepository) Update(ctx context.Context, report *domain.GeneratedReport) error {
	outputFormatsJSON, err := json.Marshal(report.OutputFormats)
	if err != nil {
		return fmt.Errorf("failed to marshal output formats: %w", err)
	}

	customFieldsJSON, err := json.Marshal(report.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	query := `
		UPDATE generated_reports SET
			template_id = $2, template_name = $3, report_type = $4, regulator_id = $5,
			regulator_name = $6, status = $7, version = $8, title = $9, description = $10,
			period_start = $11, period_end = $12, report_date = $13,
			generated_at = $14, generated_by = $15, generation_method = $16, triggered_by = $17,
			output_formats = $18, total_size = $19, checksum = $20,
			processing_time = $21, retry_count = $22, last_retry_at = $23, failure_reason = $24,
			is_validated = $25, validated_at = $26, validated_by = $27,
			is_approved = $28, approved_at = $29, approved_by = $30,
			submission_status = $31, submitted_at = $32, submitted_by = $33,
			submission_ref = $34, acknowledgment_ref = $35,
			is_archived = $36, archived_at = $37,
			custom_fields = $38, updated_by = $39, updated_at = NOW()
		WHERE id = $1
	`

	_, err = r.db.ExecContext(ctx, query,
		report.ID, report.TemplateID, report.TemplateName, report.ReportType,
		report.RegulatorID, report.RegulatorName, report.Status, report.Version,
		report.Title, report.Description, report.PeriodStart, report.PeriodEnd, report.ReportDate,
		report.GeneratedAt, report.GeneratedBy, report.GenerationMethod, report.TriggeredBy,
		outputFormatsJSON, report.TotalSize, report.Checksum,
		report.ProcessingTime, report.RetryCount, report.LastRetryAt, report.FailureReason,
		report.IsValidated, report.ValidatedAt, report.ValidatedBy,
		report.IsApproved, report.ApprovedAt, report.ApprovedBy,
		report.SubmissionStatus, report.SubmittedAt, report.SubmittedBy,
		report.SubmissionRef, report.AcknowledgmentRef,
		report.IsArchived, report.ArchivedAt,
		customFieldsJSON, report.UpdatedBy,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update generated report: %w", err)
	}
	
	return nil
}

// Delete soft-deletes a generated report
func (r *PostgresGeneratedReportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE generated_reports SET is_archived = true, archived_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete generated report: %w", err)
	}
	
	return nil
}

// GetByID retrieves a generated report by ID
func (r *PostgresGeneratedReportRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.GeneratedReport, error) {
	query := `
		SELECT id, template_id, template_name, report_type, regulator_id, regulator_name,
			status, version, title, description, period_start, period_end, report_date,
			generated_at, generated_by, generation_method, triggered_by, output_formats,
			total_size, checksum, processing_time, retry_count, last_retry_at, failure_reason,
			is_validated, validated_at, validated_by, is_approved, approved_at, approved_by,
			submission_status, submitted_at, submitted_by, submission_ref, acknowledgment_ref,
			is_archived, archived_at, retention_period, custom_fields,
			created_by, created_at, updated_by, updated_at
		FROM generated_reports WHERE id = $1
	`

	var report domain.GeneratedReport
	var outputFormatsJSON, customFieldsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID, &report.TemplateID, &report.TemplateName, &report.ReportType,
		&report.RegulatorID, &report.RegulatorName, &report.Status, &report.Version,
		&report.Title, &report.Description, &report.PeriodStart, &report.PeriodEnd, &report.ReportDate,
		&report.GeneratedAt, &report.GeneratedBy, &report.GenerationMethod, &report.TriggeredBy,
		&outputFormatsJSON, &report.TotalSize, &report.Checksum,
		&report.ProcessingTime, &report.RetryCount, &report.LastRetryAt, &report.FailureReason,
		&report.IsValidated, &report.ValidatedAt, &report.ValidatedBy,
		&report.IsApproved, &report.ApprovedAt, &report.ApprovedBy,
		&report.SubmissionStatus, &report.SubmittedAt, &report.SubmittedBy,
		&report.SubmissionRef, &report.AcknowledgmentRef,
		&report.IsArchived, &report.ArchivedAt, &report.RetentionPeriod, &customFieldsJSON,
		&report.CreatedAt, &report.UpdatedAt, &report.UpdatedBy,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get generated report: %w", err)
	}

	if err := json.Unmarshal(outputFormatsJSON, &report.OutputFormats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal output formats: %w", err)
	}
	
	if err := json.Unmarshal(customFieldsJSON, &report.CustomFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
	}

	return &report, nil
}

// List retrieves paginated generated reports with filters
func (r *PostgresGeneratedReportRepository) List(ctx context.Context, filter domain.ReportFilter) ([]*domain.GeneratedReport, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Build WHERE conditions
	if len(filter.RegulatorIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("regulator_id = ANY($%d)", argIndex))
		args = append(args, pq.Array(filter.RegulatorIDs))
		argIndex++
	}
	
	if len(filter.ReportTypes) > 0 {
		conditions = append(conditions, fmt.Sprintf("report_type = ANY($%d)", argIndex))
		args = append(args, pq.Array(filter.ReportTypes))
		argIndex++
	}
	
	if len(filter.Statuses) > 0 {
		conditions = append(conditions, fmt.Sprintf("status = ANY($%d)", argIndex))
		args = append(args, pq.Array(filter.Statuses))
		argIndex++
	}
	
	if filter.PeriodStart != nil {
		conditions = append(conditions, fmt.Sprintf("period_start >= $%d", argIndex))
		args = append(args, filter.PeriodStart)
		argIndex++
	}
	
	if filter.PeriodEnd != nil {
		conditions = append(conditions, fmt.Sprintf("period_end <= $%d", argIndex))
		args = append(args, filter.PeriodEnd)
		argIndex++
	}
	
	if filter.GeneratedAfter != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, filter.GeneratedAfter)
		argIndex++
	}
	
	if filter.GeneratedBefore != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, filter.GeneratedBefore)
		argIndex++
	}
	
	if filter.TemplateID != nil {
		conditions = append(conditions, fmt.Sprintf("template_id = $%d", argIndex))
		args = append(args, filter.TemplateID)
		argIndex++
	}
	
	if filter.SearchQuery != "" {
		searchPattern := "%" + strings.ToLower(filter.SearchQuery) + "%"
		conditions = append(conditions, fmt.Sprintf("(LOWER(title) LIKE $%d OR LOWER(description) LIKE $%d)", argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	// Add is_archived = false condition
	conditions = append(conditions, "is_archived = false")

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM generated_reports %s", whereClause)
	var total int
	countArgs := append([]interface{}{}, args...)
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	// Determine sort order
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	offset := (filter.Page - 1) * filter.PageSize
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	// Query with pagination
	query := fmt.Sprintf(`
		SELECT id, template_id, template_name, report_type, regulator_id, regulator_name,
			status, version, title, description, period_start, period_end, report_date,
			generated_at, generated_by, generation_method, triggered_by, output_formats,
			total_size, checksum, processing_time, retry_count, last_retry_at, failure_reason,
			is_validated, validated_at, validated_by, is_approved, approved_at, approved_by,
			submission_status, submitted_at, submitted_by, submission_ref, acknowledgment_ref,
			is_archived, archived_at, retention_period, custom_fields,
			created_by, created_at, updated_by, updated_at
		FROM generated_reports %s
		ORDER BY %s %s LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query reports: %w", err)
	}
	defer rows.Close()

	reports, err := r.scanReports(rows)
	if err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetByRegulatorID retrieves reports for a specific regulator
func (r *PostgresGeneratedReportRepository) GetByRegulatorID(ctx context.Context, regulatorID uuid.UUID, filter domain.ReportFilter) ([]*domain.GeneratedReport, int, error) {
	filter.RegulatorIDs = []uuid.UUID{regulatorID}
	return r.List(ctx, filter)
}

// GetByStatus retrieves reports by status
func (r *PostgresGeneratedReportRepository) GetByStatus(ctx context.Context, status domain.ReportStatus, limit int) ([]*domain.GeneratedReport, error) {
	query := fmt.Sprintf(`
		SELECT id, template_id, template_name, report_type, regulator_id, regulator_name,
			status, version, title, description, period_start, period_end, report_date,
			generated_at, generated_by, generation_method, triggered_by, output_formats,
			total_size, checksum, processing_time, retry_count, last_retry_at, failure_reason,
			is_validated, validated_at, validated_by, is_approved, approved_at, approved_by,
			submission_status, submitted_at, submitted_by, submission_ref, acknowledgment_ref,
			is_archived, archived_at, retention_period, custom_fields,
			created_by, created_at, updated_by, updated_at
		FROM generated_reports WHERE status = $1 AND is_archived = false
		ORDER BY created_at ASC LIMIT $2
	`, limit)

	rows, err := r.db.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query reports by status: %w", err)
	}
	defer rows.Close()

	return r.scanReports(rows)
}

// GetPendingReports retrieves pending reports
func (r *PostgresGeneratedReportRepository) GetPendingReports(ctx context.Context, limit int) ([]*domain.GeneratedReport, error) {
	return r.GetByStatus(ctx, domain.ReportStatusPending, limit)
}

// GetOverdueReports retrieves reports past their deadline
func (r *PostgresGeneratedReportRepository) GetOverdueReports(ctx context.Context, deadline time.Time) ([]*domain.GeneratedReport, error) {
	query := `
		SELECT id, template_id, template_name, report_type, regulator_id, regulator_name,
			status, version, title, description, period_start, period_end, report_date,
			generated_at, generated_by, generation_method, triggered_by, output_formats,
			total_size, checksum, processing_time, retry_count, last_retry_at, failure_reason,
			is_validated, validated_at, validated_by, is_approved, approved_at, approved_by,
			submission_status, submitted_at, submitted_by, submission_ref, acknowledgment_ref,
			is_archived, archived_at, retention_period, custom_fields,
			created_by, created_at, updated_by, updated_at
		FROM generated_reports WHERE report_date < $1 AND status = 'pending' AND is_archived = false
		ORDER BY report_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, deadline)
	if err != nil {
		return nil, fmt.Errorf("failed to query overdue reports: %w", err)
	}
	defer rows.Close()

	return r.scanReports(rows)
}

// UpdateStatus updates the status of a report
func (r *PostgresGeneratedReportRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReportStatus, reason string) error {
	query := `UPDATE generated_reports SET status = $2, failure_reason = $3, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id, status, reason)
	if err != nil {
		return fmt.Errorf("failed to update report status: %w", err)
	}
	
	return nil
}

// UpdateSubmissionStatus updates the submission status of a report
func (r *PostgresGeneratedReportRepository) UpdateSubmissionStatus(ctx context.Context, id uuid.UUID, status string, ref string) error {
	query := `UPDATE generated_reports SET submission_status = $2, submission_ref = $3, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id, status, ref)
	if err != nil {
		return fmt.Errorf("failed to update submission status: %w", err)
	}
	
	return nil
}

// Archive archives a report
func (r *PostgresGeneratedReportRepository) Archive(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE generated_reports SET is_archived = true, archived_at = NOW(), updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to archive report: %w", err)
	}
	
	return nil
}

// Validate marks a report as validated
func (r *PostgresGeneratedReportRepository) Validate(ctx context.Context, id uuid.UUID, validatedBy string) error {
	query := `UPDATE generated_reports SET is_validated = true, validated_at = NOW(), validated_by = $2, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id, validatedBy)
	if err != nil {
		return fmt.Errorf("failed to validate report: %w", err)
	}
	
	return nil
}

// Approve marks a report as approved
func (r *PostgresGeneratedReportRepository) Approve(ctx context.Context, id uuid.UUID, approvedBy string) error {
	query := `UPDATE generated_reports SET is_approved = true, approved_at = NOW(), approved_by = $2, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id, approvedBy)
	if err != nil {
		return fmt.Errorf("failed to approve report: %w", err)
	}
	
	return nil
}

// RecordGeneration records the generation time of a report
func (r *PostgresGeneratedReportRepository) RecordGeneration(ctx context.Context, id uuid.UUID, processingTime int64) error {
	query := `UPDATE generated_reports SET status = 'completed', generated_at = NOW(), processing_time = $2, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id, processingTime)
	if err != nil {
		return fmt.Errorf("failed to record generation: %w", err)
	}
	
	return nil
}

// RecordFailure records a generation failure
func (r *PostgresGeneratedReportRepository) RecordFailure(ctx context.Context, id uuid.UUID, reason string) error {
	query := `UPDATE generated_reports SET status = 'failed', last_retry_at = NOW(), retry_count = retry_count + 1, failure_reason = $2, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id, reason)
	if err != nil {
		return fmt.Errorf("failed to record failure: %w", err)
	}
	
	return nil
}

// Search searches for reports
func (r *PostgresGeneratedReportRepository) Search(ctx context.Context, query string, page, pageSize int) ([]*domain.GeneratedReport, int, error) {
	filter := domain.ReportFilter{
		SearchQuery: query,
		Page:        page,
		PageSize:    pageSize,
	}
	return r.List(ctx, filter)
}

// scanReports helper function to scan reports from rows
func (r *PostgresGeneratedReportRepository) scanReports(rows *sql.Rows) ([]*domain.GeneratedReport, error) {
	var reports []*domain.GeneratedReport

	for rows.Next() {
		var report domain.GeneratedReport
		var outputFormatsJSON, customFieldsJSON []byte

		err := rows.Scan(
			&report.ID, &report.TemplateID, &report.TemplateName, &report.ReportType,
			&report.RegulatorID, &report.RegulatorName, &report.Status, &report.Version,
			&report.Title, &report.Description, &report.PeriodStart, &report.PeriodEnd, &report.ReportDate,
			&report.GeneratedAt, &report.GeneratedBy, &report.GenerationMethod, &report.TriggeredBy,
			&outputFormatsJSON, &report.TotalSize, &report.Checksum,
			&report.ProcessingTime, &report.RetryCount, &report.LastRetryAt, &report.FailureReason,
			&report.IsValidated, &report.ValidatedAt, &report.ValidatedBy,
			&report.IsApproved, &report.ApprovedAt, &report.ApprovedBy,
			&report.SubmissionStatus, &report.SubmittedAt, &report.SubmittedBy,
			&report.SubmissionRef, &report.AcknowledgmentRef,
			&report.IsArchived, &report.ArchivedAt, &report.RetentionPeriod, &customFieldsJSON,
			&report.CreatedAt, &report.UpdatedAt, &report.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}

		if err := json.Unmarshal(outputFormatsJSON, &report.OutputFormats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal output formats: %w", err)
		}
		
		if err := json.Unmarshal(customFieldsJSON, &report.CustomFields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
		}

		reports = append(reports, &report)
	}

	return reports, nil
}
