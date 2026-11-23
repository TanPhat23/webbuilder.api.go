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
	key, err := validateMoveParams(kind, pageID, name, to)
	if err != nil {
		return none, none, err
	}
	tx, cleanup, err := r.beginTx(ctx)
	if err != nil {
		return none, none, err
	}
	defer cleanup()

	project, pages, page, err := loadMoveTargetDocs(ctx, tx, projectID, pageID, userID)
	if err != nil {
		return none, none, err
	}
	if err := applyAndValidateMove(project, pages, page, key, name, to); err != nil {
		return none, none, err
	}
	return commitMovedLogic(ctx, tx, projectID, pageID, project, page)
}

func validateMoveParams(kind, pageID, name, to string) (string, error) {
	key, err := logicSectionKey(kind)
	if err != nil {
		return "", err
	}
	if pageID == "" || name == "" {
		return "", fmt.Errorf("%w: invalid move", domainErrors.ErrValidation)
	}
	if to != "project" && to != "page" {
		return "", fmt.Errorf("%w: invalid move", domainErrors.ErrValidation)
	}
	return key, nil
}

func loadMoveTargetDocs(ctx context.Context, tx *sql.Tx, projectID, pageID, userID string) (logic.Document, map[string]logic.Document, logic.Document, error) {
	project, err := lockAndAuthorizeProject(ctx, tx, projectID, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	pages, err := loadProjectPages(ctx, tx, projectID)
	if err != nil {
		return nil, nil, nil, err
	}
	page, ok := pages[pageID]
	if !ok {
		return nil, nil, nil, domainErrors.ErrNotFound
	}
	return project, pages, page, nil
}

func applyAndValidateMove(project logic.Document, pages map[string]logic.Document, page logic.Document, key, name, to string) error {
	if err := transferLogicEntry(project, page, key, name, to); err != nil {
		return err
	}
	if err := logic.ValidateScopes(project, pages); err != nil {
		return fmt.Errorf("%w: %v", domainErrors.ErrValidation, err)
	}
	return nil
}

func commitMovedLogic(ctx context.Context, tx *sql.Tx, projectID, pageID string, project, page logic.Document) (logic.Revisioned, logic.Revisioned, error) {
	var none logic.Revisioned
	projectRevision, err := saveProjectLogic(ctx, tx, projectID, project)
	if err != nil {
		return none, none, err
	}
	pageRevision, err := savePageLogic(ctx, tx, projectID, pageID, page)
	if err != nil {
		return none, none, err
	}
	if err := tx.Commit(); err != nil {
		return none, none, err
	}
	return logic.Revisioned{Doc: project, Revision: projectRevision}, logic.Revisioned{Doc: page, Revision: pageRevision}, nil
}

func logicSectionKey(kind string) (string, error) {
	switch kind {
	case "variable":
		return "variables", nil
	case "flow":
		return "flows", nil
	default:
		return "", fmt.Errorf("%w: invalid logic kind", domainErrors.ErrValidation)
	}
}

func (r *LogicRepository) beginTx(ctx context.Context) (*sql.Tx, func(), error) {
	db, err := r.db.DB()
	if err != nil {
		return nil, nil, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, nil, err
	}
	return tx, func() { _ = tx.Rollback() }, nil
}

func lockAndAuthorizeProject(ctx context.Context, tx *sql.Tx, projectID, userID string) (logic.Document, error) {
	var raw []byte
	var owner string
	if err := tx.QueryRowContext(ctx, `SELECT "Logic", "OwnerId" FROM "Project" WHERE "Id" = $1 AND "DeletedAt" IS NULL FOR UPDATE`, projectID).Scan(&raw, &owner); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("project: %w", err)
	}
	if err := authorizeLogic(ctx, tx, projectID, userID, owner, true); err != nil {
		return nil, err
	}
	return decodeLogic(raw)
}

func loadProjectPages(ctx context.Context, tx *sql.Tx, projectID string) (map[string]logic.Document, error) {
	rows, err := tx.QueryContext(ctx, `SELECT "Id", "Logic" FROM "Page" WHERE "ProjectId" = $1 AND "DeletedAt" IS NULL`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pages := make(map[string]logic.Document)
	for rows.Next() {
		var id string
		var data []byte
		if err := rows.Scan(&id, &data); err != nil {
			return nil, err
		}
		doc, err := decodeLogic(data)
		if err != nil {
			return nil, err
		}
		pages[id] = doc
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pages, nil
}

func transferLogicEntry(project, page logic.Document, key, name, to string) error {
	from, dest := project, page
	if to == "project" {
		from, dest = page, project
	}
	source, sourceOK := from[key].(map[string]any)
	target, targetOK := dest[key].(map[string]any)
	if !sourceOK || !targetOK {
		return fmt.Errorf("%w: invalid logic document", domainErrors.ErrValidation)
	}
	value, exists := source[name]
	if !exists {
		return fmt.Errorf("%w: definition not found", domainErrors.ErrNotFound)
	}
	if _, exists := target[name]; exists {
		return fmt.Errorf("%w: destination already contains definition", domainErrors.ErrValidation)
	}
	delete(source, name)
	target[name] = value
	return nil
}

func saveProjectLogic(ctx context.Context, tx *sql.Tx, projectID string, doc logic.Document) (int64, error) {
	encoded, err := json.Marshal(doc)
	if err != nil {
		return 0, err
	}
	var revision int64
	err = tx.QueryRowContext(ctx, `UPDATE "Project" SET "Logic" = $1::jsonb, "LogicRevision" = "LogicRevision" + 1 WHERE "Id" = $2 RETURNING "LogicRevision"`, string(encoded), projectID).Scan(&revision)
	return revision, err
}

func savePageLogic(ctx context.Context, tx *sql.Tx, projectID, pageID string, doc logic.Document) (int64, error) {
	encoded, err := json.Marshal(doc)
	if err != nil {
		return 0, err
	}
	var revision int64
	err = tx.QueryRowContext(ctx, `UPDATE "Page" SET "Logic" = $1::jsonb, "LogicRevision" = "LogicRevision" + 1 WHERE "Id" = $2 AND "ProjectId" = $3 AND "DeletedAt" IS NULL RETURNING "LogicRevision"`, string(encoded), pageID, projectID).Scan(&revision)
	return revision, err
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
	tx, cleanup, err := r.beginTx(ctx)
	if err != nil {
		return none, err
	}
	defer cleanup()

	project, err := lockAndAuthorizeProject(ctx, tx, projectID, userID)
	if err != nil {
		return none, err
	}
	pages, err := loadProjectPages(ctx, tx, projectID)
	if err != nil {
		return none, err
	}

	updated, err := mergeScopedLogicDoc(project, pages, pageID, patch, firstNullPaths(nullPaths))
	if err != nil {
		return none, err
	}
	revision, err := saveScopedLogicDoc(ctx, tx, projectID, pageID, updated)
	if err != nil {
		return none, err
	}
	if err := tx.Commit(); err != nil {
		return none, err
	}
	return logic.Revisioned{Doc: updated, Revision: revision}, nil
}

func firstNullPaths(nullPaths [][][]string) [][]string {
	if len(nullPaths) > 0 {
		return nullPaths[0]
	}
	return nil
}

func mergeScopedLogicDoc(project logic.Document, pages map[string]logic.Document, pageID string, patch map[string]any, paths [][]string) (logic.Document, error) {
	current, err := selectLogicScope(project, pages, pageID)
	if err != nil {
		return nil, err
	}
	updated, err := logic.MergeWithNulls(current, patch, paths)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainErrors.ErrValidation, err)
	}
	if pageID == "" {
		project = updated
	} else {
		pages[pageID] = updated
	}
	if err := logic.ValidateScopes(project, pages); err != nil {
		return nil, fmt.Errorf("%w: %v", domainErrors.ErrValidation, err)
	}
	return updated, nil
}

func selectLogicScope(project logic.Document, pages map[string]logic.Document, pageID string) (logic.Document, error) {
	if pageID == "" {
		return project, nil
	}
	current, ok := pages[pageID]
	if !ok {
		return nil, fmt.Errorf("%w: page not found in project", domainErrors.ErrNotFound)
	}
	return current, nil
}

func saveScopedLogicDoc(ctx context.Context, tx *sql.Tx, projectID, pageID string, updated logic.Document) (int64, error) {
	if pageID == "" {
		return saveProjectLogic(ctx, tx, projectID, updated)
	}
	return savePageLogic(ctx, tx, projectID, pageID, updated)
}

// Get reads both scopes while holding a project share lock, so no concurrent
// logic writer can interleave between the two reads.
func (r *LogicRepository) Get(ctx context.Context, projectID, pageID, userID string) (logic.Revisioned, logic.Revisioned, error) {
	var none logic.Revisioned
	tx, cleanup, err := r.beginTx(ctx)
	if err != nil {
		return none, none, err
	}
	defer cleanup()

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
