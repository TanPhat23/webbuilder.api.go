package utils

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

// BuildColumnUpdates converts a JSON-key map into a DB-column map using an
// explicit allowlist. Only keys present in the allowlist are included; the
// map value is the DB column name to write to.
//
// Values that are themselves objects or arrays are automatically marshalled
// into json.RawMessage so GORM writes them as JSONB.
//
//	allowed := map[string]string{"name": "Name", "styles": "Styles"}
//	cols, err := utils.BuildColumnUpdates(raw, allowed)
func BuildColumnUpdates(updates map[string]any, allowed map[string]string) (map[string]any, error) {
	cols := make(map[string]any, len(updates))
	for jsonKey, colName := range allowed {
		val, ok := updates[jsonKey]
		if !ok {
			continue
		}
		switch v := val.(type) {
		case map[string]any, []any:
			b, err := json.Marshal(v)
			if err != nil {
				return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid value for field "+jsonKey+": "+err.Error())
			}
			cols[colName] = json.RawMessage(b)
		default:
			cols[colName] = v
		}
	}
	return cols, nil
}