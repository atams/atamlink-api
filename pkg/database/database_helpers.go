package database

import (
	"strings"
	"time"
	
	"database/sql"
	"github.com/lib/pq"
)

// PostgreSQL error codes
const (
	// Class 23 - Integrity Constraint Violation
	UniqueViolation     = "23505"
	ForeignKeyViolation = "23503"
	NotNullViolation    = "23502"
	CheckViolation      = "23514"
)

// IsUniqueViolation check if error is unique constraint violation
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	// Check PostgreSQL error
	if pgErr, ok := err.(*pq.Error); ok {
		return pgErr.Code == UniqueViolation
	}

	// Fallback untuk error string
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") || 
	       strings.Contains(errStr, "unique constraint") ||
	       strings.Contains(errStr, "already exists")
}

// IsForeignKeyViolation check if error is foreign key constraint violation
func IsForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}

	// Check PostgreSQL error
	if pgErr, ok := err.(*pq.Error); ok {
		return pgErr.Code == ForeignKeyViolation
	}

	// Fallback untuk error string
	errStr := err.Error()
	return strings.Contains(errStr, "foreign key constraint") ||
	       strings.Contains(errStr, "violates foreign key")
}

// IsNotNullViolation check if error is not null constraint violation
func IsNotNullViolation(err error) bool {
	if err == nil {
		return false
	}

	// Check PostgreSQL error
	if pgErr, ok := err.(*pq.Error); ok {
		return pgErr.Code == NotNullViolation
	}

	// Fallback untuk error string
	errStr := err.Error()
	return strings.Contains(errStr, "not-null constraint") ||
	       strings.Contains(errStr, "violates not-null")
}

// IsCheckViolation check if error is check constraint violation
func IsCheckViolation(err error) bool {
	if err == nil {
		return false
	}

	// Check PostgreSQL error
	if pgErr, ok := err.(*pq.Error); ok {
		return pgErr.Code == CheckViolation
	}

	// Fallback untuk error string
	errStr := err.Error()
	return strings.Contains(errStr, "check constraint") ||
	       strings.Contains(errStr, "violates check")
}

// GetConstraintName extract constraint name from PostgreSQL error
func GetConstraintName(err error) string {
	if err == nil {
		return ""
	}

	// Get from PostgreSQL error
	if pgErr, ok := err.(*pq.Error); ok {
		return pgErr.Constraint
	}

	return ""
}

// GetTableName extract table name from PostgreSQL error
func GetTableName(err error) string {
	if err == nil {
		return ""
	}

	// Get from PostgreSQL error
	if pgErr, ok := err.(*pq.Error); ok {
		return pgErr.Table
	}

	return ""
}

// GetColumnName extract column name from PostgreSQL error
func GetColumnName(err error) string {
	if err == nil {
		return ""
	}

	// Get from PostgreSQL error
	if pgErr, ok := err.(*pq.Error); ok {
		return pgErr.Column
	}

	return ""
}

// NullString convert string to sql.NullString
func NullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

// NullInt64 convert int64 to sql.NullInt64
func NullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: i,
		Valid: i > 0,
	}
}

// NullInt64FromPtr convert *int64 to sql.NullInt64
func NullInt64FromPtr(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{
			Valid: false,
		}
	}
	return sql.NullInt64{
		Int64: *i,
		Valid: true,
	}
}

// NullBool convert bool to sql.NullBool
func NullBool(b bool) sql.NullBool {
	return sql.NullBool{
		Bool:  b,
		Valid: true,
	}
}

// NullFloat64 convert float64 to sql.NullFloat64
func NullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: f,
		Valid:   f != 0,
	}
}

// NullTime convert *time.Time to sql.NullTime
func NullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{
			Valid: false,
		}
	}
	return sql.NullTime{
		Time:  *t,
		Valid: true,
	}
}

// NullTimeNow convert time.Now() to sql.NullTime
func NullTimeNow() sql.NullTime {
	return sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
}

// StringValue get string value from sql.NullString
func StringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// Int64Value get int64 value from sql.NullInt64
func Int64Value(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

// BoolValue get bool value from sql.NullBool
func BoolValue(nb sql.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false
}

// Float64Value get float64 value from sql.NullFloat64
func Float64Value(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

// TimeValue get *time.Time value from sql.NullTime
func TimeValue(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}