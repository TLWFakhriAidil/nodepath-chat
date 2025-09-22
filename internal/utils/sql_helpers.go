package utils

import (
	"database/sql"
	"strings"
)

// SplitAndTrim splits a string by delimiter and trims each part
func SplitAndTrim(s string, delimiter string) []string {
	parts := strings.Split(s, delimiter)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GeneratePlaceholders creates SQL placeholders like ?, ?, ?
func GeneratePlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ", ")
}

// GetStringValue safely gets string value from sql.NullString
func GetStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
