package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"intervention-service/models"
	"intervention-service/repository"
)

// EmergencyService handles emergency control business logic
type EmergencyService struct {
	repo  *repository.PostgresRepository
	redis *redis.Client
}

// NewEmergencyService creates a new emergency service
func NewEmergencyService(repo *repository.PostgresRepository, redisClient *redis.Client) *EmergencyService {
	return &EmergencyService{
		repo:  repo,
		redis: redisClient,
	}
}

// CreateEmergencyLock creates a new emergency lock
func (s *EmergencyService) CreateEmergencyLock(ctx context.Context, lock *models.EmergencyLock) error {
	lock.ID = uuid.New()
	lock.Status = models.LockStatusActive
	lock.CreatedAt = time.Now()
	lock.UpdatedAt = time.Now()

	// Validate lock
	if err := s.validateEmergencyLock(lock); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create Redis-based lock for immediate effect
	redisLockKey := fmt.Sprintf("emergency:lock:%s:%s", lock.ResourceType, lock.ResourceID)
	lockData := map[string]interface{}{
		"id":         lock.ID,
		"type":       lock.LockType,
		"reason":     lock.Reason,
		"priority":   lock.Priority,
		"created_at": lock.CreatedAt,
	}
	
	jsonData, _ := json.Marshal(lockData)
	
	var ttl time.Duration
	if lock.ExpiresAt != nil {
		ttl = time.Until(*lock.ExpiresAt)
	} else {
		ttl = 24 * time.Hour // Default 24 hour lock
	}

	s.redis.Set(ctx, redisLockKey, jsonData, ttl)

	// Create in database
	if err := s.repo.CreateEmergencyLock(ctx, lock); err != nil {
		return fmt.Errorf("failed to create emergency lock: %w", err)
	}

	log.Printf("Created emergency lock %s of type %s for resource %s", 
		lock.ID, lock.LockType, lock.ResourceID)

	return nil
}

// ListEmergencyLocks retrieves all active emergency locks
func (s *EmergencyService) ListEmergencyLocks(ctx context.Context, filters map[string]interface{}) ([]*models.EmergencyLock, error) {
	return s.repo.ListEmergencyLocks(ctx, filters)
}

// ReleaseEmergencyLock releases an emergency lock
func (s *EmergencyService) ReleaseEmergencyLock(ctx context.Context, id uuid.UUID, releasedBy uuid.UUID) error {
	lock, err := s.repo.GetEmergencyLock(ctx, id)
	if err != nil {
		return err
	}

	if lock.Status != models.LockStatusActive {
		return fmt.Errorf("lock is not active")
	}

	now := time.Now()
	lock.Status = models.LockStatusReleased
	lock.ReleasedAt = &now
	lock.ReleasedBy = &releasedBy
	lock.UpdatedAt = now

	// Remove Redis lock
	redisLockKey := fmt.Sprintf("emergency:lock:%s:%s", lock.ResourceType, lock.ResourceID)
	s.redis.Del(ctx, redisLockKey)

	if err := s.repo.UpdateEmergencyLock(ctx, lock); err != nil {
		return err
	}

	log.Printf("Released emergency lock %s by user %s", id, releasedBy)

	return nil
}

// CreateOverride creates a new override authorization
func (s *EmergencyService) CreateOverride(ctx context.Context, override *models.Override) error {
	override.ID = uuid.New()
	override.Status = "active"
	override.CreatedAt = time.Now()

	if err := s.validateOverride(override); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return s.repo.CreateOverride(ctx, override)
}

// GetEmergencyStatus returns the current emergency status
func (s *EmergencyService) GetEmergencyStatus(ctx context.Context) (map[string]interface{}, error) {
	activeLocks, err := s.repo.GetActiveLocksCount(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"emergency_mode":    activeLocks > 0,
		"active_locks":      activeLocks,
		"subsidy_overrides": 0,
		"last_updated":      time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// CheckLockExists checks if a lock exists for a resource
func (s *EmergencyService) CheckLockExists(ctx context.Context, resourceType, resourceID string) (bool, *models.EmergencyLock, error) {
	redisLockKey := fmt.Sprintf("emergency:lock:%s:%s", resourceType, resourceID)
	data, err := s.redis.Get(ctx, redisLockKey).Result()
	
	if err == redis.Nil {
		// Check database
		return s.repo.CheckLockExists(ctx, resourceType, resourceID)
	}
	
	if err != nil {
		return false, nil, err
	}

	// Parse Redis data
	var lockData map[string]interface{}
	json.Unmarshal([]byte(data), &lockData)
	
	lock := &models.EmergencyLock{
		ID:         lockData["id"].(uuid.UUID),
		LockType:   models.EmergencyLockType(lockData["type"].(string)),
		ResourceID: resourceID,
		ResourceType: resourceType,
		Status:     models.LockStatusActive,
	}
	
	return true, lock, nil
}

func (s *EmergencyService) validateEmergencyLock(lock *models.EmergencyLock) error {
	if lock.LockType == "" {
		return fmt.Errorf("lock type is required")
	}
	if lock.ResourceID == "" {
		return fmt.Errorf("resource ID is required")
	}
	if lock.ResourceType == "" {
		return fmt.Errorf("resource type is required")
	}
	if lock.Reason == "" {
		return fmt.Errorf("lock reason is required")
	}
	if lock.AuthorizedBy == uuid.Nil {
		return fmt.Errorf("authorized by is required")
	}
	return nil
}

func (s *EmergencyService) validateOverride(override *models.Override) error {
	if override.Type == "" {
		return fmt.Errorf("override type is required")
	}
	if override.TargetID == "" {
		return fmt.Errorf("target ID is required")
	}
	if override.AuthorizedBy == uuid.Nil {
		return fmt.Errorf("authorized by is required")
	}
	return nil
}
