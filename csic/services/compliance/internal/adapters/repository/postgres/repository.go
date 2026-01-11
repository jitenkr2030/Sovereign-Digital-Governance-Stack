package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/services/services/compliance/internal/core/domain"
	"github.com/csic-platform/services/services/compliance/internal/core/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Repository implements all compliance repository interfaces
type Repository struct {
	conn Conn
	log  *zap.Logger
}

// NewRepository creates a new Repository instance
func NewRepository(conn Conn, log *zap.Logger) *Repository {
	return &Repository{
		conn: conn,
		log:  log,
	}
}

// Helper functions for scanning

func scanLicense(row RowScanner) (*domain.License, error) {
	l := &domain.License{}
	err := row.Scan(
		&l.ID, &l.EntityID, &l.Type, &l.Status, &l.LicenseNumber,
		&l.IssuedDate, &l.ExpiryDate, &l.Conditions, &l.Restrictions,
		&l.Jurisdiction, &l.IssuedBy, &l.CreatedAt, &l.UpdatedAt,
		&l.RevokedAt, &l.RevocationReason,
	)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func scanApplication(row RowScanner) (*domain.LicenseApplication, error) {
	a := &domain.LicenseApplication{}
	err := row.Scan(
		&a.ID, &a.EntityID, &a.Type, &a.LicenseType, &a.Status,
		&a.SubmittedAt, &a.ReviewedAt, &a.ReviewerID, &a.ReviewerNotes,
		&a.RequestedTerms, &a.GrantedTerms, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func scanEntity(row RowScanner) (*domain.Entity, error) {
	e := &domain.Entity{}
	err := row.Scan(
		&e.ID, &e.Name, &e.LegalName, &e.RegistrationNum,
		&e.Jurisdiction, &e.EntityType, &e.Address, &e.ContactEmail,
		&e.Status, &e.RiskLevel, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func scanRegulation(row RowScanner) (*domain.Regulation, error) {
	r := &domain.Regulation{}
	err := row.Scan(
		&r.ID, &r.Title, &r.Description, &r.Category,
		&r.Jurisdiction, &r.EffectiveDate, &r.ParentID,
		&r.Requirements, &r.PenaltyDetails, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func scanComplianceScore(row RowScanner) (*domain.ComplianceScore, error) {
	s := &domain.ComplianceScore{}
	err := row.Scan(
		&s.ID, &s.EntityID, &s.TotalScore, &s.Tier, &s.Breakdown,
		&s.PeriodStart, &s.PeriodEnd, &s.CalculatedAt, &s.CalculationDetails,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func scanObligation(row RowScanner) (*domain.Obligation, error) {
	o := &domain.Obligation{}
	err := row.Scan(
		&o.ID, &o.EntityID, &o.RegulationID, &o.Description,
		&o.DueDate, &o.Status, &o.Priority, &o.EvidenceRefs,
		&o.FulfilledAt, &o.FulfilledEvidence, &o.ReminderSentAt,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func scanAuditRecord(row RowScanner) (*domain.AuditRecord, error) {
	a := &domain.AuditRecord{}
	err := row.Scan(
		&a.ID, &a.EntityID, &a.ActionType, &a.ActorID, &a.ActorType,
		&a.ResourceID, &a.ResourceType, &a.Timestamp,
		&a.OldValue, &a.NewValue, &a.Changes, &a.Metadata,
		&a.IPAddress, &a.UserAgent,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// License Repository Methods

func (r *Repository) CreateLicense(ctx context.Context, license *domain.License) error {
	query := `
		INSERT INTO compliance_licenses (
			id, entity_id, type, status, license_number, issued_date, expiry_date,
			conditions, restrictions, jurisdiction, issued_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.conn.Exec(ctx, query,
		license.ID, license.EntityID, license.Type, license.Status, license.LicenseNumber,
		license.IssuedDate, license.ExpiryDate, license.Conditions, license.Restrictions,
		license.Jurisdiction, license.IssuedBy, license.CreatedAt, license.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create license: %w", err)
	}
	return nil
}

func (r *Repository) GetLicense(ctx context.Context, id uuid.UUID) (*domain.License, error) {
	query := `SELECT * FROM compliance_licenses WHERE id = $1`
	row := r.conn.QueryRow(ctx, query, id)
	return scanLicense(row)
}

func (r *Repository) GetLicenseByNumber(ctx context.Context, number string) (*domain.License, error) {
	query := `SELECT * FROM compliance_licenses WHERE license_number = $1`
	row := r.conn.QueryRow(ctx, query, number)
	return scanLicense(row)
}

func (r *Repository) GetLicensesByEntity(ctx context.Context, entityID uuid.UUID) ([]domain.License, error) {
	query := `SELECT * FROM compliance_licenses WHERE entity_id = $1 ORDER BY created_at DESC`
	rows, err := r.conn.Query(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query licenses: %w", err)
	}
	defer rows.Close()

	var licenses []domain.License
	for rows.Next() {
		l, err := scanLicense(rows)
		if err != nil {
			return nil, err
		}
		licenses = append(licenses, *l)
	}
	return licenses, nil
}

func (r *Repository) UpdateLicense(ctx context.Context, license *domain.License) error {
	query := `
		UPDATE compliance_licenses SET
			status = $1, conditions = $2, restrictions = $3, updated_at = $4,
			revoked_at = $5, revocation_reason = $6
		WHERE id = $7
	`
	_, err := r.conn.Exec(ctx, query,
		license.Status, license.Conditions, license.Restrictions,
		time.Now().UTC(), license.RevokedAt, license.RevocationReason, license.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update license: %w", err)
	}
	return nil
}

func (r *Repository) UpdateLicenseStatus(ctx context.Context, id uuid.UUID, status domain.LicenseStatus) error {
	query := `UPDATE compliance_licenses SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.conn.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update license status: %w", err)
	}
	return nil
}

func (r *Repository) GetLicensesExpiringSoon(ctx context.Context, days int) ([]domain.License, error) {
	query := `
		SELECT * FROM compliance_licenses
		WHERE status = 'ACTIVE' AND expiry_date <= NOW() + INTERVAL '%d days'
		ORDER BY expiry_date ASC
	`
	rows, err := r.conn.Query(ctx, fmt.Sprintf(query, days))
	if err != nil {
		return nil, fmt.Errorf("failed to query expiring licenses: %w", err)
	}
	defer rows.Close()

	var licenses []domain.License
	for rows.Next() {
		l, err := scanLicense(rows)
		if err != nil {
			return nil, err
		}
		licenses = append(licenses, *l)
	}
	return licenses, nil
}

func (r *Repository) GetActiveLicenses(ctx context.Context) ([]domain.License, error) {
	query := `SELECT * FROM compliance_licenses WHERE status = 'ACTIVE' ORDER BY created_at DESC`
	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active licenses: %w", err)
	}
	defer rows.Close()

	var licenses []domain.License
	for rows.Next() {
		l, err := scanLicense(rows)
		if err != nil {
			return nil, err
		}
		licenses = append(licenses, *l)
	}
	return licenses, nil
}

// Application Repository Methods

func (r *Repository) CreateApplication(ctx context.Context, app *domain.LicenseApplication) error {
	query := `
		INSERT INTO compliance_applications (
			id, entity_id, type, license_type, status, submitted_at, reviewed_at,
			reviewer_id, reviewer_notes, requested_terms, granted_terms, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.conn.Exec(ctx, query,
		app.ID, app.EntityID, app.Type, app.LicenseType, app.Status, app.SubmittedAt,
		app.ReviewedAt, app.ReviewerID, app.ReviewerNotes, app.RequestedTerms,
		app.GrantedTerms, app.CreatedAt, app.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}
	return nil
}

func (r *Repository) GetApplication(ctx context.Context, id uuid.UUID) (*domain.LicenseApplication, error) {
	query := `SELECT * FROM compliance_applications WHERE id = $1`
	row := r.conn.QueryRow(ctx, query, id)
	return scanApplication(row)
}

func (r *Repository) GetApplicationsByEntity(ctx context.Context, entityID uuid.UUID) ([]domain.LicenseApplication, error) {
	query := `SELECT * FROM compliance_applications WHERE entity_id = $1 ORDER BY created_at DESC`
	rows, err := r.conn.Query(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	}
	defer rows.Close()

	var apps []domain.LicenseApplication
	for rows.Next() {
		a, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, *a)
	}
	return apps, nil
}

func (r *Repository) GetApplicationsByStatus(ctx context.Context, status domain.ApplicationStatus) ([]domain.LicenseApplication, error) {
	query := `SELECT * FROM compliance_applications WHERE status = $1 ORDER BY submitted_at DESC`
	rows, err := r.conn.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	}
	defer rows.Close()

	var apps []domain.LicenseApplication
	for rows.Next() {
		a, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, *a)
	}
	return apps, nil
}

func (r *Repository) UpdateApplication(ctx context.Context, app *domain.LicenseApplication) error {
	query := `
		UPDATE compliance_applications SET
			status = $1, reviewed_at = $2, reviewer_id = $3, reviewer_notes = $4,
			granted_terms = $5, updated_at = $6
		WHERE id = $7
	`
	_, err := r.conn.Exec(ctx, query,
		app.Status, app.ReviewedAt, app.ReviewerID, app.ReviewerNotes,
		app.GrantedTerms, time.Now().UTC(), app.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}
	return nil
}

func (r *Repository) UpdateApplicationStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, reviewerID *uuid.UUID, notes string) error {
	now := time.Now().UTC()
	var query string
	var args []interface{}

	if reviewerID != nil {
		query = `
			UPDATE compliance_applications SET
				status = $1, reviewed_at = $2, reviewer_id = $3, reviewer_notes = $4, updated_at = $5
			WHERE id = $6
		`
		args = []interface{}{status, now, *reviewerID, notes, now, id}
	} else {
		query = `
			UPDATE compliance_applications SET status = $1, updated_at = $2 WHERE id = $3
		`
		args = []interface{}{status, now, id}
	}

	_, err := r.conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update application status: %w", err)
	}
	return nil
}

// Entity Repository Methods

func (r *Repository) CreateEntity(ctx context.Context, entity *domain.Entity) error {
	query := `
		INSERT INTO compliance_entities (
			id, name, legal_name, registration_num, jurisdiction, entity_type,
			address, contact_email, status, risk_level, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.conn.Exec(ctx, query,
		entity.ID, entity.Name, entity.LegalName, entity.RegistrationNum,
		entity.Jurisdiction, entity.EntityType, entity.Address, entity.ContactEmail,
		entity.Status, entity.RiskLevel, entity.CreatedAt, entity.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}
	return nil
}

func (r *Repository) GetEntity(ctx context.Context, id uuid.UUID) (*domain.Entity, error) {
	query := `SELECT * FROM compliance_entities WHERE id = $1`
	row := r.conn.QueryRow(ctx, query, id)
	return scanEntity(row)
}

func (r *Repository) GetEntityByRegistration(ctx context.Context, regNum string) (*domain.Entity, error) {
	query := `SELECT * FROM compliance_entities WHERE registration_num = $1`
	row := r.conn.QueryRow(ctx, query, regNum)
	return scanEntity(row)
}

func (r *Repository) UpdateEntity(ctx context.Context, entity *domain.Entity) error {
	query := `
		UPDATE compliance_entities SET
			name = $1, legal_name = $2, jurisdiction = $3, entity_type = $4,
			address = $5, contact_email = $6, status = $7, risk_level = $8, updated_at = $9
		WHERE id = $10
	`
	_, err := r.conn.Exec(ctx, query,
		entity.Name, entity.LegalName, entity.Jurisdiction, entity.EntityType,
		entity.Address, entity.ContactEmail, entity.Status, entity.RiskLevel,
		time.Now().UTC(), entity.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}
	return nil
}

// Regulation Repository Methods

func (r *Repository) CreateRegulation(ctx context.Context, reg *domain.Regulation) error {
	query := `
		INSERT INTO compliance_regulations (
			id, title, description, category, jurisdiction, effective_date,
			parent_id, requirements, penalty_details, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.conn.Exec(ctx, query,
		reg.ID, reg.Title, reg.Description, reg.Category, reg.Jurisdiction,
		reg.EffectiveDate, reg.ParentID, reg.Requirements, reg.PenaltyDetails,
		reg.CreatedAt, reg.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create regulation: %w", err)
	}
	return nil
}

func (r *Repository) GetRegulation(ctx context.Context, id uuid.UUID) (*domain.Regulation, error) {
	query := `SELECT * FROM compliance_regulations WHERE id = $1`
	row := r.conn.QueryRow(ctx, query, id)
	return scanRegulation(row)
}

func (r *Repository) GetRegulationsByCategory(ctx context.Context, category string) ([]domain.Regulation, error) {
	query := `SELECT * FROM compliance_regulations WHERE category = $1 ORDER BY effective_date DESC`
	rows, err := r.conn.Query(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query regulations: %w", err)
	}
	defer rows.Close()

	var regs []domain.Regulation
	for rows.Next() {
		r, err := scanRegulation(rows)
		if err != nil {
			return nil, err
		}
		regs = append(regs, *r)
	}
	return regs, nil
}

// Compliance Repository Methods

func (r *Repository) CreateScore(ctx context.Context, score *domain.ComplianceScore) error {
	query := `
		INSERT INTO compliance_scores (
			id, entity_id, total_score, tier, breakdown, period_start, period_end,
			calculated_at, calculation_details
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.conn.Exec(ctx, query,
		score.ID, score.EntityID, score.TotalScore, score.Tier, score.Breakdown,
		score.PeriodStart, score.PeriodEnd, score.CalculatedAt, score.CalculationDetails,
	)
	if err != nil {
		return fmt.Errorf("failed to create score: %w", err)
	}
	return nil
}

func (r *Repository) GetCurrentScore(ctx context.Context, entityID uuid.UUID) (*domain.ComplianceScore, error) {
	query := `
		SELECT * FROM compliance_scores
		WHERE entity_id = $1
		ORDER BY calculated_at DESC
		LIMIT 1
	`
	row := r.conn.QueryRow(ctx, query, entityID)
	return scanComplianceScore(row)
}

func (r *Repository) GetScoreHistory(ctx context.Context, entityID uuid.UUID, limit int) ([]domain.ComplianceScore, error) {
	query := `
		SELECT * FROM compliance_scores
		WHERE entity_id = $1
		ORDER BY calculated_at DESC
		LIMIT $2
	`
	rows, err := r.conn.Query(ctx, query, entityID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query score history: %w", err)
	}
	defer rows.Close()

	var scores []domain.ComplianceScore
	for rows.Next() {
		s, err := scanComplianceScore(rows)
		if err != nil {
			return nil, err
		}
		scores = append(scores, *s)
	}
	return scores, nil
}

func (r *Repository) GetScoresByTier(ctx context.Context, tier domain.ComplianceTier) ([]domain.ComplianceScore, error) {
	query := `
		SELECT DISTINCT ON (entity_id) * FROM compliance_scores
		WHERE tier = $1
		ORDER BY entity_id, calculated_at DESC
	`
	rows, err := r.conn.Query(ctx, query, tier)
	if err != nil {
		return nil, fmt.Errorf("failed to query scores by tier: %w", err)
	}
	defer rows.Close()

	var scores []domain.ComplianceScore
	for rows.Next() {
		s, err := scanComplianceScore(rows)
		if err != nil {
			return nil, err
		}
		scores = append(scores, *s)
	}
	return scores, nil
}

func (r *Repository) CalculateAverageScore(ctx context.Context) (float64, error) {
	query := `SELECT AVG(total_score) FROM compliance_scores`
	var avg float64
	err := r.conn.QueryRow(ctx, query).Scan(&avg)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate average score: %w", err)
	}
	return avg, nil
}

// Obligation Repository Methods

func (r *Repository) CreateObligation(ctx context.Context, obligation *domain.Obligation) error {
	query := `
		INSERT INTO compliance_obligations (
			id, entity_id, regulation_id, description, due_date, status, priority,
			evidence_refs, fulfilled_at, fulfilled_evidence, reminder_sent_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.conn.Exec(ctx, query,
		obligation.ID, obligation.EntityID, obligation.RegulationID, obligation.Description,
		obligation.DueDate, obligation.Status, obligation.Priority, obligation.EvidenceRefs,
		obligation.FulfilledAt, obligation.FulfilledEvidence, obligation.ReminderSentAt,
		obligation.CreatedAt, obligation.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create obligation: %w", err)
	}
	return nil
}

func (r *Repository) GetObligation(ctx context.Context, id uuid.UUID) (*domain.Obligation, error) {
	query := `SELECT * FROM compliance_obligations WHERE id = $1`
	row := r.conn.QueryRow(ctx, query, id)
	return scanObligation(row)
}

func (r *Repository) GetObligationsByEntity(ctx context.Context, entityID uuid.UUID) ([]domain.Obligation, error) {
	query := `SELECT * FROM compliance_obligations WHERE entity_id = $1 ORDER BY due_date ASC`
	rows, err := r.conn.Query(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query obligations: %w", err)
	}
	defer rows.Close()

	var obligations []domain.Obligation
	for rows.Next() {
		o, err := scanObligation(rows)
		if err != nil {
			return nil, err
		}
		obligations = append(obligations, *o)
	}
	return obligations, nil
}

func (r *Repository) GetObligationsByStatus(ctx context.Context, status domain.ObligationStatus) ([]domain.Obligation, error) {
	query := `SELECT * FROM compliance_obligations WHERE status = $1 ORDER BY due_date ASC`
	rows, err := r.conn.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query obligations: %w", err)
	}
	defer rows.Close()

	var obligations []domain.Obligation
	for rows.Next() {
		o, err := scanObligation(rows)
		if err != nil {
			return nil, err
		}
		obligations = append(obligations, *o)
	}
	return obligations, nil
}

func (r *Repository) GetOverdueObligations(ctx context.Context) ([]domain.Obligation, error) {
	query := `
		SELECT * FROM compliance_obligations
		WHERE status = 'PENDING' AND due_date < NOW()
		ORDER BY due_date ASC
	`
	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query overdue obligations: %w", err)
	}
	defer rows.Close()

	var obligations []domain.Obligation
	for rows.Next() {
		o, err := scanObligation(rows)
		if err != nil {
			return nil, err
		}
		obligations = append(obligations, *o)
	}
	return obligations, nil
}

func (r *Repository) UpdateObligation(ctx context.Context, obligation *domain.Obligation) error {
	query := `
		UPDATE compliance_obligations SET
			status = $1, evidence_refs = $2, fulfilled_at = $3, fulfilled_evidence = $4,
			reminder_sent_at = $5, updated_at = $6
		WHERE id = $7
	`
	_, err := r.conn.Exec(ctx, query,
		obligation.Status, obligation.EvidenceRefs, obligation.FulfilledAt,
		obligation.FulfilledEvidence, obligation.ReminderSentAt,
		time.Now().UTC(), obligation.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update obligation: %w", err)
	}
	return nil
}

func (r *Repository) UpdateObligationStatus(ctx context.Context, id uuid.UUID, status domain.ObligationStatus) error {
	query := `UPDATE compliance_obligations SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.conn.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update obligation status: %w", err)
	}
	return nil
}

func (r *Repository) MarkObligationFulfilled(ctx context.Context, id uuid.UUID, evidence string) error {
	query := `
		UPDATE compliance_obligations SET
			status = 'FULFILLED', fulfilled_at = NOW(), fulfilled_evidence = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.conn.Exec(ctx, query, evidence, id)
	if err != nil {
		return fmt.Errorf("failed to mark obligation fulfilled: %w", err)
	}
	return nil
}

func (r *Repository) GetUpcomingObligations(ctx context.Context, days int) ([]domain.Obligation, error) {
	query := `
		SELECT * FROM compliance_obligations
		WHERE status = 'PENDING' AND due_date <= NOW() + INTERVAL '%d days'
		ORDER BY due_date ASC
	`
	rows, err := r.conn.Query(ctx, fmt.Sprintf(query, days))
	if err != nil {
		return nil, fmt.Errorf("failed to query upcoming obligations: %w", err)
	}
	defer rows.Close()

	var obligations []domain.Obligation
	for rows.Next() {
		o, err := scanObligation(rows)
		if err != nil {
			return nil, err
		}
		obligations = append(obligations, *o)
	}
	return obligations, nil
}

// Audit Repository Methods

func (r *Repository) CreateAuditRecord(ctx context.Context, record *domain.AuditRecord) error {
	query := `
		INSERT INTO compliance_audit_records (
			id, entity_id, action_type, actor_id, actor_type, resource_id, resource_type,
			timestamp, old_value, new_value, changes, metadata, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.conn.Exec(ctx, query,
		record.ID, record.EntityID, record.ActionType, record.ActorID, record.ActorType,
		record.ResourceID, record.ResourceType, record.Timestamp, record.OldValue,
		record.NewValue, record.Changes, record.Metadata, record.IPAddress, record.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("failed to create audit record: %w", err)
	}
	return nil
}

func (r *Repository) GetAuditRecord(ctx context.Context, id uuid.UUID) (*domain.AuditRecord, error) {
	query := `SELECT * FROM compliance_audit_records WHERE id = $1`
	row := r.conn.QueryRow(ctx, query, id)
	return scanAuditRecord(row)
}

func (r *Repository) GetAuditRecordsByEntity(ctx context.Context, entityID uuid.UUID, limit, offset int) ([]domain.AuditRecord, error) {
	query := `
		SELECT * FROM compliance_audit_records
		WHERE entity_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.conn.Query(ctx, query, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit records: %w", err)
	}
	defer rows.Close()

	var records []domain.AuditRecord
	for rows.Next() {
		a, err := scanAuditRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, *a)
	}
	return records, nil
}

func (r *Repository) GetAuditRecordsByResource(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]domain.AuditRecord, error) {
	query := `
		SELECT * FROM compliance_audit_records
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY timestamp DESC
	`
	rows, err := r.conn.Query(ctx, query, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit records: %w", err)
	}
	defer rows.Close()

	var records []domain.AuditRecord
	for rows.Next() {
		a, err := scanAuditRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, *a)
	}
	return records, nil
}

func (r *Repository) GetAuditRecordsByActor(ctx context.Context, actorID uuid.UUID, limit int) ([]domain.AuditRecord, error) {
	query := `
		SELECT * FROM compliance_audit_records
		WHERE actor_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`
	rows, err := r.conn.Query(ctx, query, actorID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit records: %w", err)
	}
	defer rows.Close()

	var records []domain.AuditRecord
	for rows.Next() {
		a, err := scanAuditRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, *a)
	}
	return records, nil
}

func (r *Repository) GetAuditRecordsByDateRange(ctx context.Context, from, to time.Time, limit, offset int) ([]domain.AuditRecord, error) {
	query := `
		SELECT * FROM compliance_audit_records
		WHERE timestamp >= $1 AND timestamp <= $2
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.conn.Query(ctx, query, from, to, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit records: %w", err)
	}
	defer rows.Close()

	var records []domain.AuditRecord
	for rows.Next() {
		a, err := scanAuditRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, *a)
	}
	return records, nil
}

func (r *Repository) GetAuditRecords(ctx context.Context, filter ports.AuditFilter) ([]domain.AuditRecord, error) {
	// Build dynamic query based on filter
	baseQuery := `SELECT * FROM compliance_audit_records WHERE 1=1`
	var args []interface{}
	argNum := 1

	if filter.EntityID != nil {
		baseQuery += fmt.Sprintf(" AND entity_id = $%d", argNum)
		args = append(args, *filter.EntityID)
		argNum++
	}
	if filter.ActorID != nil {
		baseQuery += fmt.Sprintf(" AND actor_id = $%d", argNum)
		args = append(args, *filter.ActorID)
		argNum++
	}
	if filter.ResourceType != "" {
		baseQuery += fmt.Sprintf(" AND resource_type = $%d", argNum)
		args = append(args, filter.ResourceType)
		argNum++
	}
	if filter.ResourceID != nil {
		baseQuery += fmt.Sprintf(" AND resource_id = $%d", argNum)
		args = append(args, *filter.ResourceID)
		argNum++
	}
	if filter.ActionType != "" {
		baseQuery += fmt.Sprintf(" AND action_type = $%d", argNum)
		args = append(args, filter.ActionType)
		argNum++
	}
	if filter.From != nil {
		baseQuery += fmt.Sprintf(" AND timestamp >= $%d", argNum)
		args = append(args, *filter.From)
		argNum++
	}
	if filter.To != nil {
		baseQuery += fmt.Sprintf(" AND timestamp <= $%d", argNum)
		args = append(args, *filter.To)
		argNum++
	}

	baseQuery += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.conn.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit records: %w", err)
	}
	defer rows.Close()

	var records []domain.AuditRecord
	for rows.Next() {
		a, err := scanAuditRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, *a)
	}
	return records, nil
}

func (r *Repository) CountAuditRecords(ctx context.Context, filter ports.AuditFilter) (int64, error) {
	baseQuery := `SELECT COUNT(*) FROM compliance_audit_records WHERE 1=1`
	var args []interface{}
	argNum := 1

	if filter.EntityID != nil {
		baseQuery += fmt.Sprintf(" AND entity_id = $%d", argNum)
		args = append(args, *filter.EntityID)
		argNum++
	}
	if filter.ActorID != nil {
		baseQuery += fmt.Sprintf(" AND actor_id = $%d", argNum)
		args = append(args, *filter.ActorID)
		argNum++
	}
	if filter.ResourceType != "" {
		baseQuery += fmt.Sprintf(" AND resource_type = $%d", argNum)
		args = append(args, filter.ResourceType)
		argNum++
	}

	var count int64
	err := r.conn.QueryRow(ctx, baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit records: %w", err)
	}
	return count, nil
}
