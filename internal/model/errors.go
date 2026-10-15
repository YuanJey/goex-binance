package model

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrCredentialsRequired = errors.New("binance: API credentials are required")
	ErrOutcomeUnknown      = errors.New("binance: request outcome is unknown")
	ErrStreamClosed        = errors.New("binance: stream closed")
	ErrSlowConsumer        = errors.New("binance: stream consumer is too slow; events may be missing")
)

type APIError struct {
	HTTPStatus int
	Code       int    `json:"code"`
	Message    string `json:"msg"`
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	return fmt.Sprintf("binance API error: status=%d code=%d msg=%q", e.HTTPStatus, e.Code, e.Message)
}

func IsAPIError(err error, code int) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Code == code
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return "binance validation: " + e.Message
	}
	return fmt.Sprintf("binance validation: %s: %s", e.Field, e.Message)
}

// RateLimitSnapshot contains the latest counters returned by Binance response
// headers. A zero field means the header was not present.
type RateLimitSnapshot struct {
	UsedWeight1m  int
	OrderCount10s int
	OrderCount1m  int
	ObservedAt    time.Time
}
