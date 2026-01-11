package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neamplatform/sensing/payments/config"
	"github.com/neamplatform/sensing/payments/idempotency"
	"github.com/neamplatform/sensing/payments/iso20022"
	"go.uber.org/zap"
)

// BatchProcessor handles batch file processing
type BatchProcessor struct {
	cfg            *config.Config
	logger         *zap.Logger
	idempotencyMgr *idempotency.Manager
	processors     map[string]MessageProcessor
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(
	cfg *config.Config,
	logger *zap.Logger,
	idempotencyMgr *idempotency.Manager,
	processors map[string]MessageProcessor,
) *BatchProcessor {
	return &BatchProcessor{
		cfg:            cfg,
		logger:         logger,
		idempotencyMgr: idempotencyMgr,
		processors:     processors,
	}
}

// Process starts batch processing
func (b *BatchProcessor) Process(ctx context.Context) error {
	b.logger.Info("Starting batch processing",
		zap.String("directory", b.cfg.Processing.BatchDirectory))

	// Ensure batch directory exists
	if err := os.MkdirAll(b.cfg.Processing.BatchDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create batch directory: %w", err)
	}

	// Ensure archive directory exists
	if b.cfg.Batch.ArchiveProcessed {
		archivePath := b.cfg.Batch.ArchivePath
		if err := os.MkdirAll(archivePath, 0755); err != nil {
			return fmt.Errorf("failed to create archive directory: %w", err)
		}
	}

	// Start file scanner
	ticker := time.NewTicker(time.Duration(b.cfg.Batch.ScanInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			if err := b.scanAndProcess(ctx); err != nil {
				b.logger.Error("Batch scan failed", zap.Error(err))
			}
		}
	}
}

// scanAndProcess scans the batch directory and processes files
func (b *BatchProcessor) scanAndProcess(ctx context.Context) error {
	entries, err := ioutil.ReadDir(b.cfg.Processing.BatchDirectory)
	if err != nil {
		return fmt.Errorf("failed to read batch directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check file extension
		if !b.isSupportedFile(entry.Name()) {
			continue
		}

		// Skip files still being written
		if time.Since(entry.ModTime()) < 1*time.Minute {
			continue
		}

		// Process file
		filePath := filepath.Join(b.cfg.Processing.BatchDirectory, entry.Name())
		if err := b.processFile(ctx, filePath, entry.Name()); err != nil {
			b.logger.Error("Failed to process file",
				zap.String("file", filePath),
				zap.Error(err))

			// Move to error directory if configured
			b.moveToError(filePath, err)
			continue
		}

		// Archive processed file
		if b.cfg.Batch.ArchiveProcessed {
			b.archiveFile(filePath)
		}
	}

	return nil
}

// isSupportedFile checks if the file extension is supported
func (b *BatchProcessor) isSupportedFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, supported := range b.cfg.Batch.SupportedFormats {
		if ext == "."+supported {
			return true
		}
	}
	return false
}

// processFile processes a single batch file
func (b *BatchProcessor) processFile(ctx context.Context, filePath, filename string) error {
	b.logger.Info("Processing batch file",
		zap.String("file", filePath))

	startTime := time.Now()

	// Read file content
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Check file size
	if int64(len(data)) > b.cfg.Batch.MaxFileSize {
		return fmt.Errorf("file exceeds maximum size: %d > %d", len(data), b.cfg.Batch.MaxFileSize)
	}

	// Detect format
	format, _ := detectMessageFormat(data)
	processor, ok := b.processors[format]
	if !ok {
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Split file into individual messages
	messages := b.splitMessages(data, format)

	// Process messages
	processed := 0
	failed := 0

	for _, msgData := range messages {
		// Calculate message hash
		messageHash := idempotency.CalculateHash(msgData)

		// Check idempotency
		result, err := b.idempotencyMgr.Check(ctx, messageHash)
		if err != nil {
			b.logger.Error("Idempotency check failed", zap.Error(err))
			failed++
			continue
		}

		if result.Status == idempotency.StatusCompleted {
			processed++
			continue
		}

		// Parse message
		normalized, err := processor.Parse(ctx, msgData)
		if err != nil {
			b.logger.Error("Failed to parse message",
				zap.Error(err),
				zap.String("format", format))
			failed++
			continue
		}

		normalized.MessageHash = messageHash
		normalized.Timestamp = startTime

		// Output normalized message (would go to Kafka in real implementation)
		output, _ := json.Marshal(normalized)
		b.logger.Debug("Normalized message",
			zap.ByteString("message", output))

		// Mark as completed
		if err := b.idempotencyMgr.MarkCompleted(ctx, messageHash, normalized); err != nil {
			b.logger.Error("Failed to mark message as completed", zap.Error(err))
		}

		processed++
	}

	b.logger.Info("Batch file processed",
		zap.String("file", filename),
		zap.Int("total_messages", len(messages)),
		zap.Int("processed", processed),
		zap.Int("failed", failed),
		zap.Duration("duration", time.Since(startTime)))

	return nil
}

// splitMessages splits a file into individual messages
func (b *BatchProcessor) splitMessages(data []byte, format string) [][]byte {
	var messages [][]byte

	switch format {
	case "iso20022", "swift_mx":
		// XML-based messages
		messages = b.splitXMLMessages(data)
	case "swift_mt":
		// SWIFT MT messages (separated by newlines)
		messages = b.splitMTMessages(data)
	default:
		// Treat entire file as single message
		messages = [][]byte{data}
	}

	return messages
}

// splitXMLMessages splits XML content into individual messages
func (b *BatchProcessor) splitXMLMessages(data []byte) [][]byte {
	content := string(data)

	// Find document boundaries
	startTag := "<?xml"
	endTag := "</Document>"

	var messages [][]byte
	pos := 0

	for {
		// Find start of message
		startIdx := strings.Index(content[pos:], startTag)
		if startIdx == -1 {
			break
		}

		// Find end of message
		endIdx := strings.Index(content[pos+startIdx:], endTag)
		if endIdx == -1 {
			break
		}

		// Extract message (include closing tag)
		msgData := []byte(content[pos+startIdx : pos+startIdx+endIdx+len(endTag)])
		messages = append(messages, msgData)

		// Move position
		pos = pos + startIdx + endIdx + len(endTag)
	}

	return messages
}

// splitMTMessages splits SWIFT MT content into individual messages
func (b *BatchProcessor) splitMTMessages(data []byte) [][]byte {
	content := string(data)
	lines := strings.Split(content, "\n")

	var messages [][]byte
	var currentMsg []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for new message (starts with {1:)
		if strings.HasPrefix(line, "{1:") {
			// Save previous message if exists
			if len(currentMsg) > 0 {
				messages = append(messages, []byte(strings.Join(currentMsg, "\n")))
			}
			currentMsg = []string{line}
		} else if strings.HasPrefix(line, "{") && !strings.HasPrefix(line, "{1:") {
			// Continuation of current message
			currentMsg = append(currentMsg, line)
		} else if len(currentMsg) > 0 {
			// Regular line in message
			currentMsg = append(currentMsg, line)
		}
	}

	// Add last message
	if len(currentMsg) > 0 {
		messages = append(messages, []byte(strings.Join(currentMsg, "\n")))
	}

	return messages
}

// archiveFile moves a processed file to the archive directory
func (b *BatchProcessor) archiveFile(filePath string) {
	filename := filepath.Base(filePath)
	archivePath := filepath.Join(b.cfg.Batch.ArchivePath, filename)

	// Add timestamp to filename
	timestamp := time.Now().Format("20060102_150405")
	newPath := filepath.Join(b.cfg.Batch.ArchivePath, fmt.Sprintf("%s_%s", timestamp, filename))

	if err := os.Rename(filePath, newPath); err != nil {
		b.logger.Error("Failed to archive file",
			zap.String("from", filePath),
			zap.String("to", newPath),
			zap.Error(err))
	} else {
		b.logger.Debug("File archived",
			zap.String("from", filePath),
			zap.String("to", newPath))
	}
}

// moveToError moves a failed file to error directory
func (b *BatchProcessor) moveToError(filePath string, err error) {
	errorDir := filepath.Dir(filePath) + "_error"
	if err := os.MkdirAll(errorDir, 0755); err != nil {
		b.logger.Error("Failed to create error directory", zap.Error(err))
		return
	}

	filename := filepath.Base(filePath)
	errorPath := filepath.Join(errorDir, filename)

	if err := os.Rename(filePath, errorPath); err != nil {
		b.logger.Error("Failed to move file to error directory",
			zap.String("from", filePath),
			zap.String("to", errorPath),
			zap.Error(err))
	}

	// Log error
	errorLogPath := filepath.Join(errorDir, filename+".error")
	if err := ioutil.WriteFile(errorLogPath, []byte(err.Error()), 0644); err != nil {
		b.logger.Error("Failed to write error log", zap.Error(err))
	}
}
