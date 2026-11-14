package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	"my-go-app/internal/realtime/domain/logic"
)

// LogicRepository serializes writes across every page of a project, including
// across server instances, so namespace validation and persistence are atomic.
type LogicRepository struct{ db *gorm.DB }

func NewLogicRepository(db *gorm.DB) *LogicRepository { return &LogicRepository{db: db} }

// Move transfers one definition between the project and a page in a single
// transaction, preserving its value and advancing both document revisions.
func (r *LogicRepository) Move(ctx context.Context, projectID, pageID, userID, kind, name, to string) (logic.Revisioned, logic.Revisioned, error) {
	var none logic.Revisioned
	key := ""
	switch kind {
	case "variable":
		key = "variables"
	case "flow":
		key = "flows"
	default:
		return none, none, fmt.Errorf("%w: invalid logic kind", domainErrors.ErrValidation)
	}
	if pageID == "" || name == "" || (to != "project" && to != "page") {
		return none, none, fmt.Errorf("%w: invalid move", domainErrors.ErrValidation)
	}
	db, err := r.db.DB()
	if err != nil {
		return none, none, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return none, none, err
	}
	defer tx.Rollback()
	var raw []byte
	var owner string
	if err := tx.QueryRowContext(ctx, `SELECT "Logic", "OwnerId" FROM "Project" WHERE "Id" = $1 AND "DeletedAt" IS NULL FOR UPDATE`, projectID).Scan(&raw, &owner); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return none, none, domainErrors.ErrNotFound
		}
		return none, none, err
	}
	if err := authorizeLogic(ctx, tx, projectID, userID, owner, true); err != nil {
		return none, none, err
	}
	project, err := decodeLogic(raw)
	if err != nil {
		return none, none, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT "Id", "Logic" FROM "Page" WHERE "ProjectId" = $1 AND "DeletedAt" IS NULL`, projectID)
	if err != nil {
		return none, none, err
	}
	pages := make(map[string]logic.Document)
	for rows.Next() {
		var id string
		var data []byte
		if err := rows.Scan(&id, &data); err != nil {
			rows.Close()
			return none, none, err
		}
		doc, err := decodeLogic(data)
		if err != nil {
			rows.Close()
			return none, none, err
		}
		pages[id] = doc
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return none, none, err
	}
	rows.Close()
	page, ok := pages[pageID]
	if !ok {
		return none, none, domainErrors.ErrNotFound
	}
	from, dest := project, page
	if to == "project" {
		from, dest = page, project
	}
	source, sourceOK := from[key].(map[string]any)
	target, targetOK := dest[key].(map[string]any)
	if !sourceOK || !targetOK {
		return none, none, fmt.Errorf("%w: invalid logic document", domainErrors.ErrValidation)
	}
	value, exists := source[name]
	if !exists {
		return none, none, fmt.Errorf("%w: definition not found", domainErrors.ErrNotFound)
	}
	if _, exists := target[name]; exists {
		return none, none, fmt.Errorf("%w: destination already contains definition", domainErrors.ErrValidation)
	}
	delete(source, name)
	target[name] = value
	if err := logic.ValidateScopes(project, pages); err != nil {
		return none, none, fmt.Errorf("%w: %v", domainErrors.ErrValidation, err)
	}
	projectJSON, err := json.Marshal(project)
	if err != nil {
		return none, none, err
	}
	pageJSON, err := json.Marshal(page)
	if err != nil {
		return none, none, err
	}
	var projectRevision, pageRevision int64
	if err := tx.QueryRowContext(ctx, `UPDATE "Project" SET "Logic" = $1::jsonb, "LogicRevision" = "LogicRevision" + 1 WHERE "Id" = $2 RETURNING "LogicRevision"`, string(projectJSON), projectID).Scan(&projectRevision); err != nil {
		return none, none, err
	}
	if err := tx.QueryRowContext(ctx, `UPDATE "Page" SET "Logic" = $1::jsonb, "LogicRevision" = "LogicRevision" + 1 WHERE "Id" = $2 AND "ProjectId" = $3 AND "DeletedAt" IS NULL RETURNING "LogicRevision"`, string(pageJSON), pageID, projectID).Scan(&pageRevision); err != nil {
		return none, none, err
	}
	if err := tx.Commit(); err != nil {
		return none, none, err
	}
	return logic.Revisioned{Doc: project, Revision: projectRevision}, logic.Revisioned{Doc: page, Revision: pageRevision}, nil
}

func decodeLogic(raw []byte) (logic.Document, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return logic.Empty(), nil
	}
	var doc logic.Document
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	if err := logic.Validate(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// Apply merges a patch into the project (pageID empty) or the specified page.
// Permission is checked inside the write transaction.
// The document's revision is incremented in the same statement that stores it.
func (r *LogicRepository) Apply(ctx context.Context, projectID, pageID, userID string, patch map[string]any, nullPaths ...[][]string) (logic.Revisioned, error) {
	var none logic.Revisioned
	db, err := r.db.DB()
	if err != nil {
		return none, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return none, err
	}
	defer tx.Rollback()
	var raw []byte
	var owner string
	if err := tx.QueryRowContext(ctx, `SELECT "Logic", "OwnerId" FROM "Project" WHERE "Id" = $1 AND "DeletedAt" IS NULL FOR UPDATE`, projectID).Scan(&raw, &owner); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return none, domainErrors.ErrNotFound
		}
		return none, fmt.Errorf("project: %w", err)
	}
	if err := authorizeLogic(ctx, tx, projectID, userID, owner, true); err != nil {
		return none, err
	}
	project, err := decodeLogic(raw)
	if err != nil {
		return none, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT "Id", "Logic" FROM "Page" WHERE "ProjectId" = $1 AND "DeletedAt" IS NULL`, projectID)
	if err != nil {
		return none, err
	}
	pages := make(map[string]logic.Document)
	for rows.Next() {
		var id string
		var data []byte
		if err := rows.Scan(&id, &data); err != nil {
			rows.Close()
			return none, err
		}
		doc, err := decodeLogic(data)
		if err != nil {
			rows.Close()
			return none, err
		}
		pages[id] = doc
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return none, err
	}
	rows.Close()
	current := project
	if pageID != "" {
		var ok bool
		current, ok = pages[pageID]
		if !ok {
			return none, fmt.Errorf("%w: page not found in project", domainErrors.ErrNotFound)
		}
	}
	var paths [][]string
	if len(nullPaths) > 0 {
		paths = nullPaths[0]
	}
	updated, err := logic.MergeWithNulls(current, patch, paths)
	if err != nil {
		return none, fmt.Errorf("%w: %v", domainErrors.ErrValidation, err)
	}
	if pageID == "" {
		project = updated
	} else {
		pages[pageID] = updated
	}
	if err := logic.ValidateScopes(project, pages); err != nil {
		return none, fmt.Errorf("%w: %v", domainErrors.ErrValidation, err)
	}
	encoded, err := json.Marshal(updated)
	if err != nil {
		return none, err
	}
	var revision int64
	if pageID == "" {
		err = tx.QueryRowContext(ctx, `UPDATE "Project" SET "Logic" = $1::jsonb, "LogicRevision" = "LogicRevision" + 1 WHERE "Id" = $2 RETURNING "LogicRevision"`, string(encoded), projectID).Scan(&revision)
	} else {
		err = tx.QueryRowContext(ctx, `UPDATE "Page" SET "Logic" = $1::jsonb, "LogicRevision" = "LogicRevision" + 1 WHERE "Id" = $2 AND "ProjectId" = $3 AND "DeletedAt" IS NULL RETURNING "LogicRevision"`, string(encoded), pageID, projectID).Scan(&revision)
	}
	if err != nil {
		return none, err
	}
	if err := tx.Commit(); err != nil {
		return none, err
	}
	return logic.Revisioned{Doc: updated, Revision: revision}, nil
}

// Get reads both scopes while holding a project share lock, so no concurrent
// logic writer can interleave between the two reads.
func (r *LogicRepository) Get(ctx context.Context, projectID, pageID, userID string) (logic.Revisioned, logic.Revisioned, error) {
	var none logic.Revisioned
	db, err := r.db.DB()
	if err != nil {
		return none, none, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return none, none, err
	}
	defer tx.Rollback()
	var rawProject, rawPage []byte
	var projectRevision, pageRevision int64
	var owner string
	if err := tx.QueryRowContext(ctx, `SELECT "Logic", "LogicRevision", "OwnerId" FROM "Project" WHERE "Id" = $1 AND "DeletedAt" IS NULL FOR SHARE`, projectID).Scan(&rawProject, &projectRevision, &owner); err != nil {
		return none, none, err
	}
	if err := authorizeLogic(ctx, tx, projectID, userID, owner, false); err != nil {
		return none, none, err
	}
	if err := tx.QueryRowContext(ctx, `SELECT "Logic", "LogicRevision" FROM "Page" WHERE "Id" = $1 AND "ProjectId" = $2 AND "DeletedAt" IS NULL`, pageID, projectID).Scan(&rawPage, &pageRevision); err != nil {
		return none, none, err
	}
	project, err := decodeLogic(rawProject)
	if err != nil {
		return none, none, err
	}
	page, err := decodeLogic(rawPage)
	if err != nil {
		return none, none, err
	}
	if err := tx.Commit(); err != nil {
		return none, none, err
	}
	return logic.Revisioned{Doc: project, Revision: projectRevision}, logic.Revisioned{Doc: page, Revision: pageRevision}, nil
}

// Access returns project read and edit permissions for a user. A viewer may
// read but cannot change logic; an owner or editor can do both.
func (r *LogicRepository) Access(ctx context.Context, projectID, userID string) (bool, bool, error) {
	if projectID == "" || userID == "" {
		return false, false, nil
	}
	db, err := r.db.DB()
	if err != nil {
		return false, false, err
	}
	var owner bool
	var canRead, canWrite bool
	err = db.QueryRowContext(ctx, `SELECT ("OwnerId" = $2),
   EXISTS (SELECT 1 FROM "Collaborator" WHERE "ProjectId" = $1 AND "UserId" = $2 AND "Role" IN ('owner', 'editor', 'viewer')),
   EXISTS (SELECT 1 FROM "Collaborator" WHERE "ProjectId" = $1 AND "UserId" = $2 AND "Role" IN ('owner', 'editor'))
   FROM "Project" WHERE "Id" = $1 AND "DeletedAt" IS NULL`, projectID, userID).Scan(&owner, &canRead, &canWrite)
	if err == sql.ErrNoRows {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return owner || canRead, owner || canWrite, nil
}

func authorizeLogic(ctx context.Context, tx *sql.Tx, projectID, userID, owner string, write bool) error {
	if userID == "" {
		return domainErrors.ErrUnauthorized
	}
	if owner == userID {
		return nil
	}
	var allowed bool
	var err error
	if write {
		err = tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM "Collaborator" WHERE "ProjectId" = $1 AND "UserId" = $2 AND "Role" IN ('owner', 'editor'))`, projectID, userID).Scan(&allowed)
	} else {
		err = tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM "Collaborator" WHERE "ProjectId" = $1 AND "UserId" = $2 AND "Role" IN ('owner', 'editor', 'viewer'))`, projectID, userID).Scan(&allowed)
	}
	if err != nil {
		return err
	}
	if !allowed {
		return domainErrors.ErrUnauthorized
	}
	return nil
}
