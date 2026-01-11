package domain

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"time"
)

// generateUUID generates a new UUID v4 using crypto/rand.
func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return generateTimeBasedUUID()
	}

	// Set version to 4 (random) and variant to RFC 4122
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant RFC 4122

	return hex.EncodeToString(b)
}

// generateTimeBasedUUID generates a time-based UUID as a fallback.
func generateTimeBasedUUID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 10)
	_, _ = rand.Read(randomBytes)

	combined := make([]byte, 18)
	binary.LittleEndian.PutUint64(combined[:8], uint64(timestamp))
	copy(combined[8:], randomBytes)

	combined[6] = (combined[6] & 0x0f) | 0x10 // Version 1
	combined[8] = (combined[8] & 0x3f) | 0x80 // Variant RFC 4122

	return hex.EncodeToString(combined)
}
