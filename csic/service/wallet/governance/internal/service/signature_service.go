package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/csic/wallet-governance/internal/domain/models"
	"github.com/csic/wallet-governance/internal/repository"
	"github.com/google/uuid"
)

// SignatureService handles digital signature operations
type SignatureService struct {
	signatureRepo repository.SignatureRepository
	walletRepo    repository.WalletRepository
	hsmService    *HSMService
	auditRepo     repository.AuditRepository
}

// NewSignatureService creates a new signature service
func NewSignatureService(
	signatureRepo repository.SignatureRepository,
	walletRepo repository.WalletRepository,
	hsmService *HSMService,
	auditRepo repository.AuditRepository,
) *SignatureService {
	return &SignatureService{
		signatureRepo: signatureRepo,
		walletRepo:    walletRepo,
		hsmService:    hsmService,
		auditRepo:     auditRepo,
	}
}

// RequestSignature requests a new signature
func (s *SignatureService) RequestSignature(ctx context.Context, walletID, signerID uuid.UUID, message, signatureType string) (*models.SignatureRequest, error) {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found")
	}

	// Hash the message
	messageHash := sha256.Sum256([]byte(message))
	hexHash := hex.EncodeToString(messageHash[:])

	// Create signature request
	req := &models.SignatureRequest{
		RequestID:     fmt.Sprintf("SIG-%s", uuid.New().String()[:8]),
		WalletID:      walletID,
		SignerID:      signerID,
		MessageHash:   hexHash,
		Message:       message,
		SignatureType: signatureType,
		Status:        models.SignatureStatusPending,
		ExpiresAt:     time.Now().Add(5 * time.Minute),
		RetryCount:    0,
	}

	if err := s.signatureRepo.Create(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to create signature request: %w", err)
	}

	return req, nil
}

// ProcessSignature processes a signature request
func (s *SignatureService) ProcessSignature(ctx context.Context, requestID uuid.UUID) (*models.SignatureRequest, error) {
	req, err := s.signatureRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get signature request: %w", err)
	}
	if req == nil {
		return nil, fmt.Errorf("signature request not found")
	}

	// Check if expired
	if time.Now().After(req.ExpiresAt) {
		req.Status = models.SignatureStatusFailed
		req.FailureReason = "Signature request expired"
		s.signatureRepo.Update(ctx, req)
		return nil, fmt.Errorf("signature request expired")
	}

	// Generate signature using HSM
	result, err := s.hsmService.Sign(ctx, req.MessageHash)
	if err != nil {
		req.Status = models.SignatureStatusFailed
		req.FailureReason = err.Error()
		req.RetryCount++
		s.signatureRepo.Update(ctx, req)
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}

	// Update request with signature
	req.Status = models.SignatureStatusCompleted
	req.Signature = result.Signature
	req.PublicKey = result.PublicKey
	now := time.Now()
	req.CompletedAt = &now
	s.signatureRepo.Update(ctx, req)

	return req, nil
}

// GetSignatureStatus retrieves the status of a signature request
func (s *SignatureService) GetSignatureStatus(ctx context.Context, id uuid.UUID) (*models.SignatureRequest, error) {
	return s.signatureRepo.GetByID(ctx, id)
}

// VerifySignature verifies a signature
func (s *SignatureService) VerifySignature(ctx context.Context, messageHash, signature, publicKey string) (bool, error) {
	return s.hsmService.Verify(ctx, messageHash, signature, publicKey)
}

// CancelSignature cancels a pending signature request
func (s *SignatureService) CancelSignature(ctx context.Context, id uuid.UUID) error {
	req, err := s.signatureRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get signature request: %w", err)
	}
	if req == nil {
		return fmt.Errorf("signature request not found")
	}

	if req.Status != models.SignatureStatusPending {
		return fmt.Errorf("can only cancel pending signature requests")
	}

	req.Status = models.SignatureStatusCancelled
	return s.signatureRepo.Update(ctx, req)
}

// GetPendingSignatures retrieves pending signature requests for a wallet
func (s *SignatureService) GetPendingSignatures(ctx context.Context, walletID uuid.UUID) ([]*models.SignatureRequest, error) {
	return s.signatureRepo.GetPendingByWallet(ctx, walletID)
}

// CleanupExpiredSignatures removes expired signature requests
func (s *SignatureService) CleanupExpiredSignatures(ctx context.Context) error {
	expired, err := s.signatureRepo.GetExpired(ctx)
	if err != nil {
		return fmt.Errorf("failed to get expired signatures: %w", err)
	}

	for _, req := range expired {
		req.Status = models.SignatureStatusFailed
		req.FailureReason = "Request expired"
		s.signatureRepo.Update(ctx, req)
	}

	return nil
}
