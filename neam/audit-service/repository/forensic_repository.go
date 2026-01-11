package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"neam-platform/audit-service/models"
)

// ForensicRepository provides forensic investigation data access
type ForensicRepository interface {
	// Investigation operations
	Create(ctx context.Context, investigation *models.ForensicRequest) error
	GetByID(ctx context.Context, id string) (*models.ForensicRequest, error)
	Update(ctx context.Context, investigation *models.ForensicRequest) error
	List(ctx context.Context, limit, offset int) ([]models.ForensicRequest, error)

	// Event operations for forensic queries
	GetEventByID(ctx context.Context, eventID string) (*models.AuditEvent, error)
	QueryEvents(ctx context.Context, query models.ForensicQuery) ([]models.AuditEvent, error)
}

// forensicRepository implements ForensicRepository
type forensicRepository struct {
	pool *pgxpool.Pool
}

// NewForensicRepository creates a new ForensicRepository
func NewForensicRepository(pool *pgxpool.Pool) ForensicRepository {
	return &forensicRepository{pool: pool}
}

// NewPool creates a new database connection pool
func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	return pgxpool.Connect(ctx, connString)
}

// Create creates a new forensic investigation
func (r *forensicRepository) Create(ctx context.Context, investigation *models.ForensicRequest) error {
	query := `
		INSERT INTO forensic_investigations (
			id, requester_id, requester_role, query, justification,
			status, approved_by, approved_at, created_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	queryJSON, _ := json.Marshal(investigation.Query)

	_, err := r.pool.Exec(ctx, query,
		investigation.ID, investigation.RequesterID, investigation.RequesterRole, queryJSON, investigation.Justification,
		investigation.Status, investigation.ApprovedBy, investigation.ApprovedAt, investigation.CreatedAt, investigation.CompletedAt,
	)

	return err
}

// GetByID retrieves a forensic investigation by ID
func (r *forensicRepository) GetByID(ctx context.Context, id string) (*models.ForensicRequest, error) {
	query := `
		SELECT id, requester_id, requester_role, query, justification,
			status, approved_by, approved_at, created_at, completed_at
		FROM forensic_investigations
		WHERE id = $1
	`

	var investigation models.ForensicRequest
	var queryJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&investigation.ID, &investigation.RequesterID, &investigation.RequesterRole, &queryJSON, &investigation.Justification,
		&investigation.Status, &investigation.ApprovedBy, &investigation.ApprovedAt, &investigation.CreatedAt, &investigation.CompletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("investigation not found")
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(queryJSON, &investigation.Query)
	return &investigation, nil
}

// Update updates a forensic investigation
func (r *forensicRepository) Update(ctx context.Context, investigation *models.ForensicRequest) error {
	query := `
		UPDATE forensic_investigations SET
			requester_id = $2, requester_role = $3, query = $4, justification = $5,
			status = $6, approved_by = $7, approved_at = $8, completed_at = $9
		WHERE id = $1
	`

	queryJSON, _ := json.Marshal(investigation.Query)

	_, err := r.pool.Exec(ctx, query,
		investigation.ID, investigation.RequesterID, investigation.RequesterRole, queryJSON, investigation.Justification,
		investigation.Status, investigation.ApprovedBy, investigation.ApprovedAt, investigation.CompletedAt,
	)

	return err
}

// List lists forensic investigations with pagination
func (r *forensicRepository) List(ctx context.Context, limit, offset int) ([]models.ForensicRequest, error) {
	query := `
		SELECT id, requester_id, requester_role, query, justification,
			status, approved_by, approved_at, created_at, completed_at
		FROM forensic_investigations
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var investigations []models.ForensicRequest
	for rows.Next() {
		var investigation models.ForensicRequest
		var queryJSON []byte

		err := rows.Scan(
			&investigation.ID, &investigation.RequesterID, &investigation.RequesterRole, &queryJSON, &investigation.Justification,
			&investigation.Status, &investigation.ApprovedBy, &investigation.ApprovedAt, &investigation.CreatedAt, &investigation.CompletedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(queryJSON, &investigation.Query)
		investigations = append(investigations, investigation)
	}

	return investigations, rows.Err()
}

// GetEventByID retrieves an audit event by ID
func (r *forensicRepository) GetEventByID(ctx context.Context, eventID string) (*models.AuditEvent, error) {
	query := `
		SELECT id, timestamp, event_type, actor_id, actor_type, actor_name,
			action, resource, resource_id, resource_type, outcome,
			details, ip_address, user_agent, location, session_id,
			request_id, trace_id, integrity_hash, previous_hash,
			chain_index, sealed, sealed_at, merkle_proof, compliance_tags,
			created_at
		FROM audit_events
		WHERE id = $1
	`

	var event models.AuditEvent
	var details []byte

	err := r.pool.QueryRow(ctx, query, eventID).Scan(
		&event.ID, &event.Timestamp, &event.EventType, &event.ActorID, &event.ActorType, &event.ActorName,
		&event.Action, &event.Resource, &event.ResourceID, &event.ResourceType, &event.Outcome,
		&details, &event.IPAddress, &event.UserAgent, &event.Location, &event.SessionID,
		&event.RequestID, &event.TraceID, &event.IntegrityHash, &event.PreviousHash,
		&event.ChainIndex, &event.Sealed, &event.SealedAt, &event.MerkleProof, &event.ComplianceTags,
		&event.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("event not found")
	}
	if err != nil {
		return nil, err
	}

	if details != nil {
		json.Unmarshal(details, &event.Details)
	}

	return &event, nil
}

// QueryEvents queries audit events for forensic investigation
func (r *forensicRepository) QueryEvents(ctx context.Context, query models.ForensicQuery) ([]models.AuditEvent, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if len(query.EventTypes) > 0 {
		whereClause += fmt.Sprintf(" AND event_type = ANY($%d)", argPos)
		args = append(args, query.EventTypes)
		argPos++
	}

	if len(query.ActorIDs) > 0 {
		whereClause += fmt.Sprintf(" AND actor_id = ANY($%d)", argPos)
		args = append(args, query.ActorIDs)
		argPos++
	}

	if len(query.Resources) > 0 {
		whereClause += fmt.Sprintf(" AND resource = ANY($%d)", argPos)
		args = append(args, query.Resources)
		argPos++
	}

	if len(query.Outcomes) > 0 {
		whereClause += fmt.Sprintf(" AND outcome = ANY($%d)", argPos)
		args = append(args, query.Outcomes)
		argPos++
	}

	whereClause += fmt.Sprintf(" AND timestamp >= $%d AND timestamp <= $%d", argPos, argPos+1)
	args = append(args, query.TimeRange.Start, query.TimeRange.End)
	argPos += 2

	// Set limit
	limit := 1000
	if query.Limit > 0 {
		limit = query.Limit
	}

	queryString := fmt.Sprintf(`
		SELECT id, timestamp, event_type, actor_id, actor_type, actor_name,
			action, resource, resource_id, resource_type, outcome,
			details, ip_address, user_agent, location, session_id,
			request_id, trace_id, integrity_hash, previous_hash,
			chain_index, sealed, sealed_at, merkle_proof, compliance_tags,
			created_at
		FROM audit_events
		%s
		ORDER BY timestamp ASC
		LIMIT $%d
	`, whereClause, argPos)

	args = append(args, limit)

	rows, err := r.pool.Query(ctx, queryString, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.AuditEvent
	for rows.Next() {
		var event models.AuditEvent
		var details []byte

		err := rows.Scan(
			&event.ID, &event.Timestamp, &event.EventType, &event.ActorID, &event.ActorType, &event.ActorName,
			&event.Action, &event.Resource, &event.ResourceID, &event.ResourceType, &event.Outcome,
			&details, &event.IPAddress, &event.UserAgent, &event.Location, &event.SessionID,
			&event.RequestID, &event.TraceID, &event.IntegrityHash, &event.PreviousHash,
			&event.ChainIndex, &event.Sealed, &event.SealedAt, &event.MerkleProof, &event.ComplianceTags,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if details != nil {
			json.Unmarshal(details, &event.Details)
		}

		events = append(events, event)
	}

	return events, rows.Err()
}
