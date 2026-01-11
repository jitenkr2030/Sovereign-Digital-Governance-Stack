package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"csic-platform/service/security/internal/domain/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgreSQL SIEM event repository for managing security events to external systems.
type PostgresSIEMEventRepository struct {
	db          *sql.DB
	tablePrefix string
}

// NewPostgresSIEMEventRepository creates a new PostgreSQL SIEM event repository.
func NewPostgresSIEMEventRepository(db *sql.DB, tablePrefix string) *PostgresSIEMEventRepository {
	return &PostgresSIEMEventRepository{
		db:          db,
		tablePrefix: tablePrefix,
	}
}

// tableName returns the full table name with prefix.
func (r *PostgresSIEMEventRepository) tableName() string {
	return r.tablePrefix + "_siem_events"
}

// StoreEvent stores a SIEM event for forwarding to external systems.
func (r *PostgresSIEMEventRepository) StoreEvent(ctx context.Context, event *models.SIEMEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	event.CreatedAt = time.Now().UTC()
	event.Status = models.SIEMEventStatusPending

	// Serialize event data for storage.
	eventDataJSON, err := json.Marshal(event.EventData)
	if err != nil {
		return fmt.Errorf("failed to serialize event data: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, event_type, severity, source, source_ip, destination_ip,
			user_id, session_id, action, outcome, event_data,
			raw_event, correlation_id, timestamp, status,
			retry_count, last_retry_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, r.tableName())

	_, err = r.db.ExecContext(ctx, query,
		event.ID,
		event.EventType,
		event.Severity,
		event.Source,
		event.SourceIP,
		event.DestinationIP,
		event.UserID,
		event.SessionID,
		event.Action,
		event.Outcome,
		string(eventDataJSON),
		event.RawEvent,
		event.CorrelationID,
		event.Timestamp,
		event.Status,
		event.RetryCount,
		event.LastRetryAt,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store SIEM event: %w", err)
	}

	return nil
}

// GetPendingEvents retrieves events that haven't been forwarded yet.
func (r *PostgresSIEMEventRepository) GetPendingEvents(ctx context.Context, limit int) ([]*models.SIEMEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	query := fmt.Sprintf(`
		SELECT id, event_type, severity, source, source_ip, destination_ip,
			   user_id, session_id, action, outcome, event_data,
			   raw_event, correlation_id, timestamp, status,
			   retry_count, last_retry_at, created_at
		FROM %s
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`, r.tableName())

	rows, err := r.db.QueryContext(ctx, query, models.SIEMEventStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending events: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// MarkEventForwarded marks an event as successfully forwarded.
func (r *PostgresSIEMEventRepository) MarkEventForwarded(ctx context.Context, eventID string, remoteID string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, remote_event_id = $2, forwarded_at = $3
		WHERE id = $4
	`, r.tableName())

	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query,
		models.SIEMEventStatusForwarded,
		remoteID,
		now,
		eventID,
	)

	if err != nil {
		return fmt.Errorf("failed to mark event as forwarded: %w", err)
	}

	return nil
}

// MarkEventFailed marks an event as failed after maximum retries.
func (r *PostgresSIEMEventRepository) MarkEventFailed(ctx context.Context, eventID string, errorMsg string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, last_error = $2, last_retry_at = $3, retry_count = retry_count + 1
		WHERE id = $4
	`, r.tableName())

	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query,
		models.SIEMEventStatusFailed,
		errorMsg,
		now,
		eventID,
	)

	if err != nil {
		return fmt.Errorf("failed to mark event as failed: %w", err)
	}

	return nil
}

// GetEventStats retrieves statistics about SIEM event forwarding.
func (r *PostgresSIEMEventRepository) GetEventStats(ctx context.Context) (*models.SIEMEventStats, error) {
	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN status = $1 THEN 1 END) as pending,
			COUNT(CASE WHEN status = $2 THEN 1 END) as forwarded,
			COUNT(CASE WHEN status = $3 THEN 1 END) as failed,
			COUNT(CASE WHEN status = $4 THEN 1 END) as retrying,
			MIN(created_at) as oldest_event,
			MAX(created_at) as newest_event
		FROM %s
	`, r.tableName())

	stats := &models.SIEMEventStats{}

	err := r.db.QueryRowContext(ctx, query,
		models.SIEMEventStatusPending,
		models.SIEMEventStatusForwarded,
		models.SIEMEventStatusFailed,
		models.SIEMEventStatusRetrying,
	).Scan(
		&stats.TotalEvents,
		&stats.PendingEvents,
		&stats.ForwardedEvents,
		&stats.FailedEvents,
		&stats.RetryingEvents,
		&stats.OldestEvent,
		&stats.NewestEvent,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return stats, nil
		}
		return nil, fmt.Errorf("failed to get event stats: %w", err)
	}

	return stats, nil
}

// scanEvents scans database rows into SIEMEvent structs.
func (r *PostgresSIEMEventRepository) scanEvents(rows *sql.Rows) ([]*models.SIEMEvent, error) {
	var events []*models.SIEMEvent

	for rows.Next() {
		event := &models.SIEMEvent{}
		var eventDataJSON []byte
		var rawEvent sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.Severity,
			&event.Source,
			&event.SourceIP,
			&event.DestinationIP,
			&event.UserID,
			&event.SessionID,
			&event.Action,
			&event.Outcome,
			&eventDataJSON,
			&rawEvent,
			&event.CorrelationID,
			&event.Timestamp,
			&event.Status,
			&event.RetryCount,
			&event.LastRetryAt,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SIEM event: %w", err)
		}

		if rawEvent.Valid {
			event.RawEvent = rawEvent.String
		}

		if err := json.Unmarshal(eventDataJSON, &event.EventData); err != nil {
			event.EventData = make(map[string]interface{})
		}

		events = append(events, event)
	}

	return events, rows.Err()
}
