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

// AuditRepository provides audit data access
type AuditRepository interface {
	// CRUD operations
	Create(ctx context.Context, event *models.AuditEvent) error
	GetByID(ctx context.Context, id string) (*models.AuditEvent, error)
	GetByIDs(ctx context.Context, ids []string) ([]models.AuditEvent, error)
	Update(ctx context.Context, event *models.AuditEvent) error
	Delete(ctx context.Context, id string) error

	// Query operations
	Query(ctx context.Context, query models.AuditQuery) ([]models.AuditEvent, error)
	GetStatistics(ctx context.Context) (*models.AuditStatistics, error)

	// Chain operations
	GetLastHash(ctx context.Context) (string, error)
	GetNextChainIndex(ctx context.Context) (int64, error)
	GetUnsealedEvents(ctx context.Context) ([]models.AuditEvent, error)

	// Seal operations
	CreateSeal(ctx context.Context, seal *models.SealRecord) error
	GetLatestSeal(ctx context.Context) (*models.SealRecord, error)
	GetSealByID(ctx context.Context, id string) (*models.SealRecord, error)
	UpdateSeal(ctx context.Context, seal *models.SealRecord) error
	GetRecentSeals(ctx context.Context, limit int) ([]models.SealRecord, error)
	GetEventChain(ctx context.Context, eventID string) ([]models.SealRecord, error)
	MarkSealed(ctx context.Context, eventIDs []string, sealedAt *time.Time) error
}

// auditRepository implements AuditRepository
type auditRepository struct {
	pool *pgxpool.Pool
}

// NewAuditRepository creates a new AuditRepository
func NewAuditRepository(pool *pgxpool.Pool) AuditRepository {
	return &auditRepository{pool: pool}
}

// Create creates a new audit event
func (r *auditRepository) Create(ctx context.Context, event *models.AuditEvent) error {
	details, _ := json.Marshal(event.Details)

	query := `
		INSERT INTO audit_events (
			id, timestamp, event_type, actor_id, actor_type, actor_name,
			action, resource, resource_id, resource_type, outcome,
			details, ip_address, user_agent, location, session_id,
			request_id, trace_id, integrity_hash, previous_hash,
			chain_index, sealed, sealed_at, merkle_proof, compliance_tags,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
	`

	_, err := r.pool.Exec(ctx, query,
		event.ID, event.Timestamp, event.EventType, event.ActorID, event.ActorType, event.ActorName,
		event.Action, event.Resource, event.ResourceID, event.ResourceType, event.Outcome,
		details, event.IPAddress, event.UserAgent, event.Location, event.SessionID,
		event.RequestID, event.TraceID, event.IntegrityHash, event.PreviousHash,
		event.ChainIndex, event.Sealed, event.SealedAt, event.MerkleProof, event.ComplianceTags,
		event.CreatedAt,
	)

	return err
}

// GetByID retrieves an audit event by ID
func (r *auditRepository) GetByID(ctx context.Context, id string) (*models.AuditEvent, error) {
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

	err := r.pool.QueryRow(ctx, query, id).Scan(
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

// GetByIDs retrieves multiple audit events by IDs
func (r *auditRepository) GetByIDs(ctx context.Context, ids []string) ([]models.AuditEvent, error) {
	if len(ids) == 0 {
		return []models.AuditEvent{}, nil
	}

	query := `
		SELECT id, timestamp, event_type, actor_id, actor_type, actor_name,
			action, resource, resource_id, resource_type, outcome,
			details, ip_address, user_agent, location, session_id,
			request_id, trace_id, integrity_hash, previous_hash,
			chain_index, sealed, sealed_at, merkle_proof, compliance_tags,
			created_at
		FROM audit_events
		WHERE id = ANY($1)
		ORDER BY timestamp ASC
	`

	rows, err := r.pool.Query(ctx, query, ids)
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

// Update updates an audit event
func (r *auditRepository) Update(ctx context.Context, event *models.AuditEvent) error {
	details, _ := json.Marshal(event.Details)

	query := `
		UPDATE audit_events SET
			timestamp = $2, event_type = $3, actor_id = $4, actor_type = $5, actor_name = $6,
			action = $7, resource = $8, resource_id = $9, resource_type = $10, outcome = $11,
			details = $12, ip_address = $13, user_agent = $14, location = $15, session_id = $16,
			request_id = $17, trace_id = $18, integrity_hash = $19, previous_hash = $20,
			chain_index = $21, sealed = $22, sealed_at = $23, merkle_proof = $24, compliance_tags = $25
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		event.ID, event.Timestamp, event.EventType, event.ActorID, event.ActorType, event.ActorName,
		event.Action, event.Resource, event.ResourceID, event.ResourceType, event.Outcome,
		details, event.IPAddress, event.UserAgent, event.Location, event.SessionID,
		event.RequestID, event.TraceID, event.IntegrityHash, event.PreviousHash,
		event.ChainIndex, event.Sealed, event.SealedAt, event.MerkleProof, event.ComplianceTags,
	)

	return err
}

// Delete deletes an audit event
func (r *auditRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM audit_events WHERE id = $1"
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// Query retrieves audit events based on query
func (r *auditRepository) Query(ctx context.Context, query models.AuditQuery) ([]models.AuditEvent, error) {
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

	if query.TimeRange != nil {
		whereClause += fmt.Sprintf(" AND timestamp >= $%d AND timestamp <= $%d", argPos, argPos+1)
		args = append(args, query.TimeRange.Start, query.TimeRange.End)
		argPos += 2
	}

	if !query.IncludeSealed {
		whereClause += " AND sealed = false"
	}

	// Determine sort order
	sortOrder := "DESC"
	if query.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	sortBy := "timestamp"
	switch query.SortBy {
	case "actor_id", "resource", "event_type", "outcome":
		sortBy = query.SortBy
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
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argPos, argPos+1)

	args = append(args, query.Limit, query.Offset)

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

// GetStatistics retrieves audit statistics
func (r *auditRepository) GetStatistics(ctx context.Context) (*models.AuditStatistics, error) {
	stats := &models.AuditStatistics{
		EventsByType: make(map[string]int64),
		EventsByOutcome: make(map[string]int64),
		EventsByActor: make(map[string]int64),
		EventsByResource: make(map[string]int64),
	}

	// Total events
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_events").Scan(&stats.TotalEvents)
	if err != nil {
		return nil, err
	}

	// Events by type
	rows, err := r.pool.Query(ctx, "SELECT event_type, COUNT(*) FROM audit_events GROUP BY event_type")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var eventType string
		var count int64
		if err := rows.Scan(&eventType, &count); err != nil {
			return nil, err
		}
		stats.EventsByType[eventType] = count
	}

	// Events by outcome
	rows, err = r.pool.Query(ctx, "SELECT outcome, COUNT(*) FROM audit_events GROUP BY outcome")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var outcome string
		var count int64
		if err := rows.Scan(&outcome, &count); err != nil {
			return nil, err
		}
		stats.EventsByOutcome[outcome] = count
	}

	// Failed events
	err = r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_events WHERE outcome IN ('FAILURE', 'DENIED')").Scan(&stats.FailedEvents)
	if err != nil {
		return nil, err
	}

	// Sealed events
	err = r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_events WHERE sealed = true").Scan(&stats.SealedEvents)
	if err != nil {
		return nil, err
	}

	// Unsealed events
	err = r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_events WHERE sealed = false").Scan(&stats.UnsealedEvents)
	if err != nil {
		return nil, err
	}

	// Events in last 24 hours
	err = r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_events WHERE timestamp >= NOW() - INTERVAL '24 hours'").Scan(&stats.EventsLast24h)
	if err != nil {
		return nil, err
	}

	// Events in last hour
	err = r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_events WHERE timestamp >= NOW() - INTERVAL '1 hour'").Scan(&stats.EventsLastHour)
	if err != nil {
		return nil, err
	}

	// Calculate average events per hour
	if stats.EventsLast24h > 0 {
		stats.AverageEventsPerHour = float64(stats.EventsLast24h) / 24.0
	}

	stats.GeneratedAt = time.Now()
	return stats, nil
}

// GetLastHash retrieves the hash of the last audit event
func (r *auditRepository) GetLastHash(ctx context.Context) (string, error) {
	var hash string
	err := r.pool.QueryRow(ctx, "SELECT integrity_hash FROM audit_events ORDER BY chain_index DESC LIMIT 1").Scan(&hash)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return hash, err
}

// GetNextChainIndex retrieves the next chain index
func (r *auditRepository) GetNextChainIndex(ctx context.Context) (int64, error) {
	var index int64
	err := r.pool.QueryRow(ctx, "SELECT COALESCE(MAX(chain_index), -1) + 1 FROM audit_events").Scan(&index)
	return index, err
}

// GetUnsealedEvents retrieves all unsealed events
func (r *auditRepository) GetUnsealedEvents(ctx context.Context) ([]models.AuditEvent, error) {
	query := `
		SELECT id, timestamp, event_type, actor_id, actor_type, actor_name,
			action, resource, resource_id, resource_type, outcome,
			details, ip_address, user_agent, location, session_id,
			request_id, trace_id, integrity_hash, previous_hash,
			chain_index, sealed, sealed_at, merkle_proof, compliance_tags,
			created_at
		FROM audit_events
		WHERE sealed = false
		ORDER BY chain_index ASC
	`

	rows, err := r.pool.Query(ctx, query)
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

// CreateSeal creates a new seal record
func (r *auditRepository) CreateSeal(ctx context.Context, seal *models.SealRecord) error {
	metadata, _ := json.Marshal(seal.Metadata)

	query := `
		INSERT INTO audit_seals (
			id, timestamp, root_hash, first_event_id, last_event_id,
			event_count, signature, signer_id, signer_type, algorithm,
			witness_type, witness_id, witness_signature,
			previous_seal_id, previous_root_hash, metadata, verified, verified_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	_, err := r.pool.Exec(ctx, query,
		seal.ID, seal.Timestamp, seal.RootHash, seal.FirstEventID, seal.LastEventID,
		seal.EventCount, seal.Signature, seal.SignerID, seal.SignerType, seal.Algorithm,
		seal.WitnessType, seal.WitnessID, seal.WitnessSignature,
		seal.PreviousSealID, seal.PreviousRootHash, metadata, seal.Verified, seal.VerifiedAt, seal.CreatedAt,
	)

	return err
}

// GetLatestSeal retrieves the latest seal record
func (r *auditRepository) GetLatestSeal(ctx context.Context) (*models.SealRecord, error) {
	var seal models.SealRecord
	var metadata []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, timestamp, root_hash, first_event_id, last_event_id,
			event_count, signature, signer_id, signer_type, algorithm,
			witness_type, witness_id, witness_signature,
			previous_seal_id, previous_root_hash, metadata, verified, verified_at, created_at
		FROM audit_seals
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(
		&seal.ID, &seal.Timestamp, &seal.RootHash, &seal.FirstEventID, &seal.LastEventID,
		&seal.EventCount, &seal.Signature, &seal.SignerID, &seal.SignerType, &seal.Algorithm,
		&seal.WitnessType, &seal.WitnessID, &seal.WitnessSignature,
		&seal.PreviousSealID, &seal.PreviousRootHash, &metadata, &seal.Verified, &seal.VerifiedAt, &seal.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if metadata != nil {
		json.Unmarshal(metadata, &seal.Metadata)
	}

	return &seal, nil
}

// GetSealByID retrieves a seal record by ID
func (r *auditRepository) GetSealByID(ctx context.Context, id string) (*models.SealRecord, error) {
	var seal models.SealRecord
	var metadata []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, timestamp, root_hash, first_event_id, last_event_id,
			event_count, signature, signer_id, signer_type, algorithm,
			witness_type, witness_id, witness_signature,
			previous_seal_id, previous_root_hash, metadata, verified, verified_at, created_at
		FROM audit_seals
		WHERE id = $1
	`).Scan(
		&seal.ID, &seal.Timestamp, &seal.RootHash, &seal.FirstEventID, &seal.LastEventID,
		&seal.EventCount, &seal.Signature, &seal.SignerID, &seal.SignerType, &seal.Algorithm,
		&seal.WitnessType, &seal.WitnessID, &seal.WitnessSignature,
		&seal.PreviousSealID, &seal.PreviousRootHash, &metadata, &seal.Verified, &seal.VerifiedAt, &seal.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("seal not found")
	}
	if err != nil {
		return nil, err
	}

	if metadata != nil {
		json.Unmarshal(metadata, &seal.Metadata)
	}

	return &seal, nil
}

// UpdateSeal updates a seal record
func (r *auditRepository) UpdateSeal(ctx context.Context, seal *models.SealRecord) error {
	metadata, _ := json.Marshal(seal.Metadata)

	query := `
		UPDATE audit_seals SET
			verified = $2, verified_at = $3, metadata = $4
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, seal.ID, seal.Verified, seal.VerifiedAt, metadata)
	return err
}

// GetRecentSeals retrieves recent seal records
func (r *auditRepository) GetRecentSeals(ctx context.Context, limit int) ([]models.SealRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, timestamp, root_hash, first_event_id, last_event_id,
			event_count, signature, signer_id, signer_type, algorithm,
			witness_type, witness_id, witness_signature,
			previous_seal_id, previous_root_hash, metadata, verified, verified_at, created_at
		FROM audit_seals
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seals []models.SealRecord
	for rows.Next() {
		var seal models.SealRecord
		var metadata []byte

		err := rows.Scan(
			&seal.ID, &seal.Timestamp, &seal.RootHash, &seal.FirstEventID, &seal.LastEventID,
			&seal.EventCount, &seal.Signature, &seal.SignerID, &seal.SignerType, &seal.Algorithm,
			&seal.WitnessType, &seal.WitnessID, &seal.WitnessSignature,
			&seal.PreviousSealID, &seal.PreviousRootHash, &metadata, &seal.Verified, &seal.VerifiedAt, &seal.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if metadata != nil {
			json.Unmarshal(metadata, &seal.Metadata)
		}

		seals = append(seals, seal)
	}

	return seals, rows.Err()
}

// GetEventChain retrieves the seal chain for an event
func (r *auditRepository) GetEventChain(ctx context.Context, eventID string) ([]models.SealRecord, error) {
	// Get the event first
	event, err := r.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Get all seals after this event's chain index
	rows, err := r.pool.Query(ctx, `
		SELECT id, timestamp, root_hash, first_event_id, last_event_id,
			event_count, signature, signer_id, signer_type, algorithm,
			witness_type, witness_id, witness_signature,
			previous_seal_id, previous_root_hash, metadata, verified, verified_at, created_at
		FROM audit_seals
		WHERE first_event_id <= $1
		ORDER BY created_at ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seals []models.SealRecord
	for rows.Next() {
		var seal models.SealRecord
		var metadata []byte

		err := rows.Scan(
			&seal.ID, &seal.Timestamp, &seal.RootHash, &seal.FirstEventID, &seal.LastEventID,
			&seal.EventCount, &seal.Signature, &seal.SignerID, &seal.SignerType, &seal.Algorithm,
			&seal.WitnessType, &seal.WitnessID, &seal.WitnessSignature,
			&seal.PreviousSealID, &seal.PreviousRootHash, &metadata, &seal.Verified, &seal.VerifiedAt, &seal.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if metadata != nil {
			json.Unmarshal(metadata, &seal.Metadata)
		}

		seals = append(seals, seal)
	}

	return seals, rows.Err()
}

// MarkSealed marks events as sealed
func (r *auditRepository) MarkSealed(ctx context.Context, eventIDs []string, sealedAt *time.Time) error {
	query := `
		UPDATE audit_events SET
			sealed = true, sealed_at = $2
		WHERE id = ANY($1) AND sealed = false
	`

	_, err := r.pool.Exec(ctx, query, eventIDs, sealedAt)
	return err
}
