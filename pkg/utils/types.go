package utils

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
)

// NullString creates sql.NullString ("" → invalid)
func ToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func ToNullUUID(u uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{UUID: u, Valid: u.String() != ""}
}

// FormatTime formats sql.NullTime safely
func FormatTime(nt sql.NullTime, fallback string) string {
	if nt.Valid {
		return nt.Time.Format("Jan 02 15:04")
	}
	return fallback
}

// FormatTheme formats themes name
func FormatTheme(theme string) string {
	return strings.ToLower(strings.ReplaceAll(theme, " ", "-"))
}
