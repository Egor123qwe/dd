package util

import (
	"fmt"
	"time"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
)

// ParseMySQLDateTime parses datetime string from database.
// Supports: RFC3339 (PostgreSQL, e.g. "2006-01-02T15:04:05Z"), MySQL datetime ("2006-01-02 15:04:05"), date-only ("2006-01-02").
func ParseMySQLDateTime(dateStr string) (time.Time, error) {
	formats := []string{time.RFC3339Nano, time.RFC3339, time.DateTime, time.DateOnly}
	for _, layout := range formats {
		if parsed, err := time.Parse(layout, dateStr); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("%w: failed to parse datetime: %s", errs.ErrInternalError, dateStr)
}

// NormalizeToUTC normalizes time to UTC timezone.
// If time is already in UTC, returns as is. Otherwise converts to UTC.
func NormalizeToUTC(t time.Time) time.Time {
	if t.Location() == time.UTC {
		return t
	}
	return t.UTC()
}

// ParseAndNormalizeUTC parses datetime string in RFC3339 format and normalizes to UTC.
// This is used for incoming API requests that should be in UTC.
func ParseAndNormalizeUTC(dateStr string) (time.Time, error) {
	parsedDate, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: invalid datetime format, expected RFC3339 (e.g., 2006-01-02T15:04:05Z): %w", errs.ErrInvalidRequest, err)
	}
	return NormalizeToUTC(parsedDate), nil
}
