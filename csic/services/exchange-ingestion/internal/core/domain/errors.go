package domain

import "errors"

// Common error definitions for the exchange ingestion domain

var (
	// DataSource errors
	ErrDataSourceNotFound       = errors.New("data source not found")
	ErrDataSourceAlreadyExists  = errors.New("data source already exists")
	ErrDataSourceDisabled       = errors.New("data source is disabled")
	ErrInvalidDataSourceConfig  = errors.New("invalid data source configuration")

	// Connection errors
	ErrConnectionFailed         = errors.New("connection to exchange failed")
	ErrConnectionLost           = errors.New("connection to exchange lost")
	ErrConnectionTimeout        = errors.New("connection to exchange timed out")
	ErrNotConnected             = errors.New("not connected to exchange")

	// Authentication errors
	ErrAuthenticationFailed     = errors.New("authentication failed")
	ErrInvalidCredentials       = errors.New("invalid API credentials")
	ErrMissingCredentials       = errors.New("missing API credentials")

	// Data errors
	ErrInvalidMarketData        = errors.New("invalid market data received")
	ErrDataParseFailed          = errors.New("failed to parse exchange data")
	ErrDataOutOfOrder           = dataOutOfOrderError{}
	ErrDuplicateData            = duplicateDataError{}

	// Rate limiting errors
	ErrRateLimited              = errors.New("rate limit exceeded")
	ErrTooManyRequests          = errors.New("too many requests to exchange")

	// Repository errors
	ErrRepositoryNotFound       = errors.New("record not found in repository")
	ErrRepositoryConflict       = errors.New("record conflict in repository")
	ErrRepositoryUnavailable    = errors.New("repository unavailable")

	// Ingestion errors
	ErrIngestionNotRunning      = errors.New("ingestion is not running")
	ErrIngestionAlreadyRunning  = errors.New("ingestion is already running")
	ErrIngestionOverflow        = errors.New("ingestion buffer overflow")
	ErrProcessingTimeout        = errors.New("data processing timed out")
)

// dataOutOfOrderError is returned when data is received out of expected order
type dataOutOfOrderError struct{}

func (e dataOutOfOrderError) Error() string {
	return "market data received out of order"
}

// duplicateDataError is returned when duplicate data is detected
type duplicateDataError struct{}

func (e duplicateDataError) Error() string {
	return "duplicate market data detected"
}

// IsDataOutOfOrder checks if the error is a data out of order error
func IsDataOutOfOrder(err error) bool {
	_, ok := err.(dataOutOfOrderError)
	return ok
}

// IsDuplicateData checks if the error is a duplicate data error
func IsDuplicateData(err error) bool {
	_, ok := err.(duplicateDataError)
	return ok
}
