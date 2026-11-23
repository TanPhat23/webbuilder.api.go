package repositories

import "strings"

// normalizeUpdateColumns maps JSON-style field names to model field names.
func normalizeUpdateColumns(updates map[string]any) {
	for key, value := range updates {
		if key == "" {
			continue
		}
		column := strings.ToUpper(key[:1]) + key[1:]
		if column != key {
			if _, exists := updates[column]; !exists {
				updates[column] = value
			}
			delete(updates, key)
		}
	}
}
