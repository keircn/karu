package workflow

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func ValidateQuery(query string) error {
	if query == "" {
		return ValidationError{Field: "query", Message: "search query cannot be empty"}
	}

	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return ValidationError{Field: "query", Message: "search query must be at least 2 characters"}
	}

	return nil
}

func ValidateEpisodeRange(rangeStr string, maxEpisodes int) error {
	if rangeStr == "" {
		return ValidationError{Field: "range", Message: "episode range cannot be empty"}
	}

	if strings.Contains(rangeStr, "-") {
		parts := strings.Split(rangeStr, "-")
		if len(parts) != 2 {
			return ValidationError{Field: "range", Message: "invalid range format, use start-end (e.g., 1-5)"}
		}

		start := strings.TrimSpace(parts[0])
		end := strings.TrimSpace(parts[1])

		if start == "" || end == "" {
			return ValidationError{Field: "range", Message: "range values cannot be empty"}
		}
	} else if strings.Contains(rangeStr, ",") {
		episodes := strings.Split(rangeStr, ",")
		for _, ep := range episodes {
			if strings.TrimSpace(ep) == "" {
				return ValidationError{Field: "range", Message: "episode numbers cannot be empty"}
			}
		}
	}

	return nil
}
