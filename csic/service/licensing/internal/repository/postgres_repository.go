package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-licensing/internal/domain/models"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// PostgresRepository implements all repository interfaces for PostgreSQL
type PostgresRepository struct {
	db    *sql.DB
	logger *zap.Logger
}

// NewPostgresRepository creates a new PostgreSQL repository instance
func NewPostgresRepository(db *sql.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:    db,
		logger: logger,
	}
}

// License Repository Implementation

// Create creates a new license record
func (r *PostgresRepository) Create(ctx context.Context, license *models.License) error {
	query := `
		INSERT INTO licenses (
			id, exchange_id, license_number, jurisdiction, type, status,
			issued_date, expiry_date, conditions, notes, approved_by,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	conditions, _ := json.Marshal(license.Conditions)

	_, err := r.db.ExecContext(ctx, query,
		license.ID, license.ExchangeID, license.LicenseNumber,
		license.Jurisdiction, license.Type, license.Status,
		license.IssuedDate, license.ExpiryDate, conditions,
		license.Notes, license.ApprovedBy,
		license.CreatedAt, license.UpdatedAt,
	)

	return err
}

// GetByID retrieves a license by its ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*models.License, error) {
	query := `
		SELECT id, exchange_id, license_number, jurisdiction, type, status,
			   issued_date, expiry_date, conditions, notes, approved_by,
			   created_at, updated_at
		FROM licenses WHERE id = $1
	`

	var license models.License
	var conditions []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&license.ID, &license.ExchangeID, &license.LicenseNumber,
		&license.Jurisdiction, &license.Type, &license.Status,
		&license.IssuedDate, &license.ExpiryDate, &conditions,
		&license.Notes, &license.ApprovedBy,
		&license.CreatedAt, &license.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(conditions, &license.Conditions)
	return &license, nil
}

// GetByNumber retrieves a license by its license number
func (r *PostgresRepository) GetByNumber(ctx context.Context, number string) (*models.License, error) {
	query := `
		SELECT id, exchange_id, license_number, jurisdiction, type, status,
			   issued_date, expiry_date, conditions, notes, approved_by,
			   created_at, updated_at
		FROM licenses WHERE license_number = $1
	`

	var license models.License
	var conditions []byte

	err := r.db.QueryRowContext(ctx, query, number).Scan(
		&license.ID, &license.ExchangeID, &license.LicenseNumber,
		&license.Jurisdiction, &license.Type, &license.Status,
		&license.IssuedDate, &license.ExpiryDate, &conditions,
		&license.Notes, &license.ApprovedBy,
		&license.CreatedAt, &license.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(conditions, &license.Conditions)
	return &license, nil
}

// GetByExchange retrieves all licenses for an exchange
func (r *PostgresRepository) GetByExchange(ctx context.Context, exchangeID string) ([]*models.License, error) {
	query := `
		SELECT id, exchange_id, license_number, jurisdiction, type, status,
			   issued_date, expiry_date, conditions, notes, approved_by,
			   created_at, updated_at
		FROM licenses WHERE exchange_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, exchangeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var licenses []*models.License
	for rows.Next() {
		var license models.License
		var conditions []byte

		if err := rows.Scan(
			&license.ID, &license.ExchangeID, &license.LicenseNumber,
			&license.Jurisdiction, &license.Type, &license.Status,
			&license.IssuedDate, &license.ExpiryDate, &conditions,
			&license.Notes, &license.ApprovedBy,
			&license.CreatedAt, &license.UpdatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal(conditions, &license.Conditions)
		licenses = append(licenses, &license)
	}

	return licenses, rows.Err()
}

// GetByStatus retrieves all licenses with a specific status
func (r *PostgresRepository) GetByStatus(ctx context.Context, status models.LicenseStatus) ([]*models.License, error) {
	query := `
		SELECT id, exchange_id, license_number, jurisdiction, type, status,
			   issued_date, expiry_date, conditions, notes, approved_by,
			   created_at, updated_at
		FROM licenses WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var licenses []*models.License
	for rows.Next() {
		var license models.License
		var conditions []byte

		if err := rows.Scan(
			&license.ID, &license.ExchangeID, &license.LicenseNumber,
			&license.Jurisdiction, &license.Type, &license.Status,
			&license.IssuedDate, &license.ExpiryDate, &conditions,
			&license.Notes, &license.ApprovedBy,
			&license.CreatedAt, &license.UpdatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal(conditions, &license.Conditions)
		licenses = append(licenses, &license)
	}

	return licenses, rows.Err()
}

// GetExpiringSoon retrieves licenses expiring within the specified days
func (r *PostgresRepository) GetExpiringSoon(ctx context.Context, days int) ([]*models.License, error) {
	query := `
		SELECT id, exchange_id, license_number, jurisdiction, type, status,
			   issued_date, expiry_date, conditions, notes, approved_by,
			   created_at, updated_at
		FROM licenses
		WHERE status = 'active'
		  AND expiry_date BETWEEN NOW() AND NOW() + INTERVAL '%d days'
		ORDER BY expiry_date ASC
	`

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(query, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var licenses []*models.License
	for rows.Next() {
		var license models.License
		var conditions []byte

		if err := rows.Scan(
			&license.ID, &license.ExchangeID, &license.LicenseNumber,
			&license.Jurisdiction, &license.Type, &license.Status,
			&license.IssuedDate, &license.ExpiryDate, &conditions,
			&license.Notes, &license.ApprovedBy,
			&license.CreatedAt, &license.UpdatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal(conditions, &license.Conditions)
		licenses = append(licenses, &license)
	}

	return licenses, rows.Err()
}

// Update updates an existing license
func (r *PostgresRepository) Update(ctx context.Context, license *models.License) error {
	query := `
		UPDATE licenses SET
			status = $1, conditions = $2, notes = $3, updated_at = $4
		WHERE id = $5
	`

	conditions, _ := json.Marshal(license.Conditions)

	_, err := r.db.ExecContext(ctx, query,
		license.Status, conditions, license.Notes,
		time.Now(), license.ID,
	)

	return err
}

// Delete deletes a license by ID
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM licenses WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List retrieves licenses with optional filters
func (r *PostgresRepository) List(ctx context.Context, filter LicenseFilter) ([]*models.License, error) {
	baseQuery := `
		SELECT id, exchange_id, license_number, jurisdiction, type, status,
			   issued_date, expiry_date, conditions, notes, approved_by,
			   created_at, updated_at
		FROM licenses WHERE 1=1
	`

	args := []interface{}{}
	argNum := 1

	if filter.ExchangeID != "" {
		baseQuery += fmt.Sprintf(" AND exchange_id = $%d", argNum)
		args = append(args, filter.ExchangeID)
		argNum++
	}

	if filter.Jurisdiction != "" {
		baseQuery += fmt.Sprintf(" AND jurisdiction = $%d", argNum)
		args = append(args, filter.Jurisdiction)
		argNum++
	}

	if filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, filter.Status)
		argNum++
	}

	if filter.Type != "" {
		baseQuery += fmt.Sprintf(" AND type = $%d", argNum)
		args = append(args, filter.Type)
		argNum++
	}

	baseQuery += " ORDER BY created_at DESC"

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var licenses []*models.License
	for rows.Next() {
		var license models.License
		var conditions []byte

		if err := rows.Scan(
			&license.ID, &license.ExchangeID, &license.LicenseNumber,
			&license.Jurisdiction, &license.Type, &license.Status,
			&license.IssuedDate, &license.ExpiryDate, &conditions,
			&license.Notes, &license.ApprovedBy,
			&license.CreatedAt, &license.UpdatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal(conditions, &license.Conditions)
		licenses = append(licenses, &license)
	}

	return licenses, rows.Err()
}

// Count returns the count of licenses matching the filter
func (r *PostgresRepository) Count(ctx context.Context, filter LicenseFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM licenses WHERE 1=1`

	if filter.ExchangeID != "" {
		query += " AND exchange_id = $1"
	}
	if filter.Jurisdiction != "" {
		query += " AND jurisdiction = $2"
	}
	if filter.Status != "" {
		query += " AND status = $3"
	}

	return 0, nil // Implementation would complete this
}

// Violation Repository Implementation

// Create creates a new violation record
func (r *PostgresRepository) CreateViolation(ctx context.Context, violation *models.ComplianceViolation) error {
	query := `
		INSERT INTO violations (
			id, license_id, exchange_id, severity, type, title, description,
			details, status, detected_at, assigned_to, assigned_team,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	details, _ := json.Marshal(violation.Details)

	_, err := r.db.ExecContext(ctx, query,
		violation.ID, violation.LicenseID, violation.ExchangeID,
		violation.Severity, violation.Type, violation.Title,
		violation.Description, details, violation.Status,
		violation.DetectedAt, violation.AssignedTo, violation.AssignedTeam,
		violation.CreatedAt, violation.UpdatedAt,
	)

	return err
}

// GetByID retrieves a violation by ID
func (r *PostgresRepository) GetViolationByID(ctx context.Context, id string) (*models.ComplianceViolation, error) {
	query := `
		SELECT id, license_id, exchange_id, severity, type, title, description,
			   details, status, detected_at, resolved_at, resolved_by, resolution,
			   assigned_to, assigned_team, escalated_to, created_at, updated_at
		FROM violations WHERE id = $1
	`

	var violation models.ComplianceViolation
	var details []byte
	var resolvedAt, resolvedBy, resolution, assignedTo, assignedTeam, escalatedTo sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&violation.ID, &violation.LicenseID, &violation.ExchangeID,
		&violation.Severity, &violation.Type, &violation.Title,
		&violation.Description, &details, &violation.Status,
		&violation.DetectedAt, &resolvedAt, &resolvedBy, &resolution,
		&assignedTo, &assignedTeam, &violation.EscalatedTo,
		&violation.CreatedAt, &violation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(details, &violation.Details)

	if resolvedAt.Valid {
		violation.ResolvedAt = &resolvedAt.Time
	}
	violation.ResolvedBy = resolvedBy.String
	violation.Resolution = resolution.String
	violation.AssignedTo = assignedTo.String
	violation.AssignedTeam = assignedTeam.String

	return &violation, nil
}

// GetOpenViolations retrieves all open violations
func (r *PostgresRepository) GetOpenViolations(ctx context.Context) ([]*models.ComplianceViolation, error) {
	query := `
		SELECT id, license_id, exchange_id, severity, type, title, description,
			   details, status, detected_at, assigned_to, assigned_team,
			   created_at, updated_at
		FROM violations
		WHERE status IN ('open', 'investigating', 'pending_review')
		ORDER BY severity DESC, detected_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var violations []*models.ComplianceViolation
	for rows.Next() {
		var v models.ComplianceViolation
		var details []byte

		if err := rows.Scan(
			&v.ID, &v.LicenseID, &v.ExchangeID, &v.Severity, &v.Type,
			&v.Title, &v.Description, &details, &v.Status,
			&v.DetectedAt, &v.AssignedTo, &v.AssignedTeam,
			&v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal(details, &v.Details)
		violations = append(violations, &v)
	}

	return violations, rows.Err()
}

// Update updates an existing violation
func (r *PostgresRepository) UpdateViolation(ctx context.Context, violation *models.ComplianceViolation) error {
	query := `
		UPDATE violations SET
			status = $1, resolved_at = $2, resolved_by = $3, resolution = $4,
			assigned_to = $5, assigned_team = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := r.db.ExecContext(ctx, query,
		violation.Status, violation.ResolvedAt, violation.ResolvedBy,
		violation.Resolution, violation.AssignedTo, violation.AssignedTeam,
		time.Now(), violation.ID,
	)

	return err
}

// GetStats retrieves violation statistics
func (r *PostgresRepository) GetViolationStats(ctx context.Context) (*models.ViolationStats, error) {
	stats := &models.ViolationStats{
		BySeverity: make(map[models.ViolationSeverity]int64),
		ByType:     make(map[models.ViolationType]int64),
	}

	query := `SELECT COUNT(*) FROM violations`
	var count int64
	r.db.QueryRowContext(ctx, query).Scan(&count)
	stats.TotalViolations = count

	query = `SELECT COUNT(*) FROM violations WHERE status IN ('open', 'investigating', 'pending_review')`
	r.db.QueryRowContext(ctx, query).Scan(&count)
	stats.OpenViolations = count

	query = `SELECT COUNT(*) FROM violations WHERE status = 'resolved'`
	r.db.QueryRowContext(ctx, query).Scan(&count)
	stats.ResolvedViolations = count

	query = `SELECT COUNT(*) FROM violations WHERE severity = 'critical'`
	r.db.QueryRowContext(ctx, query).Scan(&count)
	stats.CriticalCount = count

	query = `SELECT COUNT(*) FROM violations WHERE severity = 'high'`
	r.db.QueryRowContext(ctx, query).Scan(&count)
	stats.HighCount = count

	return stats, nil
}

// GetTrends retrieves violation trends
func (r *PostgresRepository) GetViolationTrends(ctx context.Context, startDate, endDate time.Time) ([]models.ViolationTrend, error) {
	query := `
		SELECT
			DATE(detected_at) as date,
			COUNT(*) as new_violations,
			0 as resolved_violations,
			0 as open_violations,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_count
		FROM violations
		WHERE detected_at BETWEEN $1 AND $2
		GROUP BY DATE(detected_at)
		ORDER BY date
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []models.ViolationTrend
	for rows.Next() {
		var t models.ViolationTrend
		if err := rows.Scan(&t.Date, &t.NewViolations, &t.ResolvedViolations, &t.OpenViolations, &t.CriticalCount); err != nil {
			return nil, err
		}
		trends = append(trends, t)
	}

	return trends, rows.Err()
}

// Report Repository Implementation

// Create creates a new report record
func (r *PostgresRepository) CreateReport(ctx context.Context, report *models.Report) error {
	query := `
		INSERT INTO reports (
			id, report_type, title, description, period_start, period_end,
			status, license_id, exchange_id, generated_by, template_id,
			parameters, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	parameters, _ := json.Marshal(report.Parameters)
	metadata, _ := json.Marshal(report.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		report.ID, report.ReportType, report.Title, report.Description,
		report.PeriodStart, report.PeriodEnd, report.Status,
		report.LicenseID, report.ExchangeID, report.GeneratedBy,
		report.TemplateID, parameters, metadata,
		report.CreatedAt, report.UpdatedAt,
	)

	return err
}

// GetByID retrieves a report by ID
func (r *PostgresRepository) GetReportByID(ctx context.Context, id string) (*models.Report, error) {
	query := `
		SELECT id, report_type, title, description, period_start, period_end,
			   status, license_id, exchange_id, generated_by, reviewed_by,
			   submitted_at, submitted_to, file_path, file_size, mime_type,
			   checksum, template_id, parameters, metadata, created_at, updated_at
		FROM reports WHERE id = $1
	`

	var report models.Report
	var parameters, metadata []byte
	var reviewedBy, submittedAt, submittedTo, filePath, mimeType, checksum, templateID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID, &report.ReportType, &report.Title, &report.Description,
		&report.PeriodStart, &report.PeriodEnd, &report.Status,
		&report.LicenseID, &report.ExchangeID, &report.GeneratedBy,
		&reviewedBy, &submittedAt, &submittedTo, &filePath,
		&report.FileSize, &mimeType, &checksum, &templateID,
		&parameters, &metadata, &report.CreatedAt, &report.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(parameters, &report.Parameters)
	json.Unmarshal(metadata, &report.Metadata)

	if reviewedBy.Valid {
		report.ReviewedBy = reviewedBy.String
	}
	if submittedAt.Valid {
		report.SubmittedAt = &submittedAt.Time
	}
	if submittedTo.Valid {
		report.SubmittedTo = submittedTo.String
	}
	if filePath.Valid {
		report.FilePath = filePath.String
	}
	if mimeType.Valid {
		report.MimeType = mimeType.String
	}
	if checksum.Valid {
		report.Checksum = checksum.String
	}
	if templateID.Valid {
		report.TemplateID = templateID.String
	}

	return &report, nil
}

// Update updates an existing report
func (r *PostgresRepository) UpdateReport(ctx context.Context, report *models.Report) error {
	query := `
		UPDATE reports SET
			status = $1, reviewed_by = $2, submitted_at = $3, submitted_to = $4,
			file_path = $5, file_size = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := r.db.ExecContext(ctx, query,
		report.Status, report.ReviewedBy, report.SubmittedAt,
		report.SubmittedTo, report.FilePath, report.FileSize,
		time.Now(), report.ID,
	)

	return err
}

// List retrieves reports with optional filters
func (r *PostgresRepository) ListReports(ctx context.Context, filter ReportFilter) ([]*models.Report, error) {
	baseQuery := `
		SELECT id, report_type, title, period_start, period_end,
			   status, license_id, exchange_id, generated_by
		FROM reports WHERE 1=1
	`

	if filter.LicenseID != "" {
		baseQuery += " AND license_id = $1"
	}

	baseQuery += " ORDER BY created_at DESC"

	if filter.Page > 0 && filter.PageSize > 0 {
		baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.PageSize, (filter.Page-1)*filter.PageSize)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		var report models.Report
		if err := rows.Scan(
			&report.ID, &report.ReportType, &report.Title,
			&report.PeriodStart, &report.PeriodEnd, &report.Status,
			&report.LicenseID, &report.ExchangeID, &report.GeneratedBy,
		); err != nil {
			return nil, err
		}
		reports = append(reports, &report)
	}

	return reports, rows.Err()
}

// GetStats retrieves report statistics
func (r *PostgresRepository) GetReportStats(ctx context.Context) (*models.ReportStats, error) {
	stats := &models.ReportStats{
		ByType:    make(map[models.ReportType]int64),
		ByLicense: make(map[string]int64),
	}

	query := `SELECT COUNT(*) FROM reports`
	var count int64
	r.db.QueryRowContext(ctx, query).Scan(&count)
	stats.TotalReports = count

	return stats, nil
}

// GetOverdueReports retrieves reports past their due date
func (r *PostgresRepository) GetOverdueReports(ctx context.Context) ([]*models.Report, error) {
	query := `
		SELECT id, report_type, title, period_start, period_end,
			   status, license_id, exchange_id, generated_by
		FROM reports
		WHERE status IN ('pending', 'generating')
		  AND period_end < NOW()
		ORDER BY period_end ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		var report models.Report
		if err := rows.Scan(
			&report.ID, &report.ReportType, &report.Title,
			&report.PeriodStart, &report.PeriodEnd, &report.Status,
			&report.LicenseID, &report.ExchangeID, &report.GeneratedBy,
		); err != nil {
			return nil, err
		}
		reports = append(reports, &report)
	}

	return reports, rows.Err()
}

// Helper: Convert pq.StringArray to []string
func stringArrayToSlice(arr pq.StringArray) []string {
	return []string(arr)
}

// Helper: Convert []string to pq.StringArray
func sliceToStringArray(s []string) pq.StringArray {
	return pq.StringArray(s)
}
