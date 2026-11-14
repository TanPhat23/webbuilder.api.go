package workflowmigration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"my-go-app/internal/realtime/domain/logic"
)

// Apply writes a complete clean plan in a single serializable transaction.
// Callers must first stop the editor and flush/invalidate its realtime cache;
// PostgreSQL cannot coordinate with Redis or already connected clients.
func Apply(ctx context.Context, db *sql.DB) (Plan, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return Plan{}, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `LOCK TABLE "Project", "Page", "Element", "EventWorkflow", "ElementEventWorkflow" IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return Plan{}, err
	}
	plan, err := planInTx(ctx, tx)
	if err != nil {
		return Plan{}, err
	}
	if len(plan.Issues) > 0 {
		return plan, nil
	}
	for _, w := range plan.Interactions {
		var raw []byte
		if err := tx.QueryRowContext(ctx, `SELECT "Settings" FROM "Element" WHERE "Id"=$1`, w.ElementID).Scan(&raw); err != nil {
			return Plan{}, err
		}
		settings, err := object(raw)
		if err != nil {
			return Plan{}, err
		}
		var interactions map[string]any
		if v, ok := settings["interactions"]; ok {
			var valid bool
			interactions, valid = v.(map[string]any)
			if !valid {
				return Plan{}, fmt.Errorf("invalid interactions for %s", w.ElementID)
			}
		} else {
			interactions = map[string]any{}
		}
		interactions[w.ID] = w.Interaction
		settings["interactions"] = interactions
		var marks map[string]any
		if v, ok := settings["migratedWorkflowLinks"]; ok {
			var valid bool
			marks, valid = v.(map[string]any)
			if !valid {
				return Plan{}, fmt.Errorf("invalid migration markers for %s", w.ElementID)
			}
		} else {
			marks = map[string]any{}
		}
		marks[w.LinkID] = true
		settings["migratedWorkflowLinks"] = marks
		encoded, err := json.Marshal(settings)
		if err != nil {
			return Plan{}, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE "Element" SET "Settings"=$1::jsonb WHERE "Id"=$2`, string(encoded), w.ElementID); err != nil {
			return Plan{}, err
		}
	}
	for _, w := range plan.Flows {
		table, id := "Project", w.ProjectID
		if w.PageID != "" {
			table, id = "Page", w.PageID
		}
		var raw []byte
		if err := tx.QueryRowContext(ctx, `SELECT "Logic" FROM "`+table+`" WHERE "Id"=$1`, id).Scan(&raw); err != nil {
			return Plan{}, err
		}
		doc := logic.Empty()
		if len(raw) > 0 && string(raw) != "null" {
			if err := json.Unmarshal(raw, &doc); err != nil {
				return Plan{}, err
			}
		}
		flows := doc["flows"].(map[string]any)
		flows[w.Name] = map[string]any{"do": w.Do}
		encoded, err := json.Marshal(doc)
		if err != nil {
			return Plan{}, err
		}
		var normalized logic.Document
		if err := json.Unmarshal(encoded, &normalized); err != nil {
			return Plan{}, err
		}
		if err := logic.Validate(normalized); err != nil {
			return Plan{}, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE "`+table+`" SET "Logic"=$1::jsonb WHERE "Id"=$2`, string(encoded), id); err != nil {
			return Plan{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Plan{}, err
	}
	return plan, nil
}
