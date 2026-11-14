# GORM generics migration

GORM v1.31.2's `gorm.G[T]` is used for typed repository operations where it preserves the existing SQL and behavior. Filters are built on the generic chain: passing a prefiltered traditional query to `gorm.G` can discard conditions. Context is passed to terminal operations; transactions use their existing transaction handles. Project and page hard deletes retain unscoped behavior, and transactional project creation continues to omit associations.

The migration covers `collaborator`, `comment`, `contentField`, `contentFieldValue`, `contentItem`, `contentType`, `customElement`, `customElementType`, `element`, `elementComment`, `elementEventWorkflow`, `eventWorkflow`, `image`, `invitation`, `marketplace`, `notification`, `page`, `project`, `snapshot`, and `user` repositories. Some operations intentionally retain traditional GORM:

- Map-based PATCH operations retain `Updates(map)` so explicit false, zero and null values, field-name mapping, callbacks and timestamps remain intact. Generic `Updates` takes the model type and omits zero fields; generic `Set` uses explicit raw column assignments and is not an established equivalent.
- Some model-instance updates and reloads retain their populated destination semantics. Aggregate map scans and association/raw joins retain the existing SQL and result shapes. Generic and traditional operations can coexist inside the same transaction.
- Transaction boundaries and connection setup remain on `*gorm.DB`; seed and migration commands use `database/sql`, not GORM CRUD.

Verification: `go test ./...`, `go vet ./...`, and the opt-in PostgreSQL test `UUID_GORM_TEST_DSN='<disposable migrated PostgreSQL DSN>' go test ./internal/repositories -run TestPostgresProjectRepositoryParity -count=1`. Use a dedicated disposable database with the baseline migration; the test inserts and deletes its own records. It covers project and page creation/rollback, scoped access, cancellation, zero/null PATCH updates, content-field create/read/duplicate behavior, notification ownership, element batch creation and preload, hard-delete ownership, and snapshot reads/counts/deletes. See `benchmarks/README.md` for the separate UUID benchmark.

Known pre-existing limitation: baseline migration `00001_baseline.sql` does not define `Snapshot.UpdatedAt`, while the model embeds `BaseAuditFields`, so snapshot creation against that baseline fails when GORM includes the column. The migration does not change this schema or model mismatch.
