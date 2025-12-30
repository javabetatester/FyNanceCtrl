package shared

import (
	"strings"
)		

func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "23505") ||
		strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "violates unique constraint") ||
		strings.Contains(errStr, "idx_categories_user_name")
}

func NormalizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}
	words := strings.Fields(name)
	normalized := make([]string, 0, len(words))
	for _, word := range words {
		if len(word) > 0 {
			if len(word) == 1 {
				normalized = append(normalized, strings.ToUpper(word))
			} else {
				normalized = append(normalized, strings.ToUpper(string(word[0]))+strings.ToLower(word[1:]))
			}
		}
	}
	return strings.Join(normalized, " ")
}
