# Workflow migration

The command reads a consistent PostgreSQL snapshot of legacy workflows, links,
elements, and Logic documents. It prints proposed interactions, flows, and
issues as JSON. It does not load `.env`; choose the database explicitly:

```sh
WORKFLOW_MIGRATION_DATABASE_URL='postgres://…' go run ./cmd/workflow-migration --dry-run > workflow-report.json
```

The report can include workflow content; treat it as project data. An exit code
of 0 means the report has no issues, 3 means one or more workflows require review,
and 1 or 2 means the command could not complete or its arguments were invalid.
Unsupported behavior, invalid links, conflicting IDs or flow names, and invalid
existing documents are reported with their workflow IDs. The command excludes
entire affected workflows from proposed writes.

## Local apply

The local-only apply mode replans and checks for issues inside a single
serializable transaction. If any issue exists it writes nothing. Migration
markers keep converted legacy links from also executing in the updated editor;
source workflow and link records are retained. Reapplying is idempotent.

Before applying, stop all editor and realtime processes connected to the target
database. Flush pending Redis page writes and clear the relevant cached page
documents using the normal service lifecycle before stopping it; otherwise a
stale cache can overwrite database changes. Ensure the updated editor code with
migration-marker handling will be used on restart. PostgreSQL alone cannot enforce
these cross-service requirements. Then run:

```sh
WORKFLOW_MIGRATION_DATABASE_URL='postgres://…@127.0.0.1/…' go run ./cmd/workflow-migration --apply --offline
```

`--offline` confirms these conditions. Apply currently accepts only local
PostgreSQL URLs. The dry run is a snapshot; apply performs its own checks because
the database may have changed since the report was produced.
