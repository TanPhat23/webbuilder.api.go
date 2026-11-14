-- +goose Up
-- Existing project-wide snapshots have no page; realtime autosaves record one.
ALTER TABLE "Snapshot" ADD COLUMN "PageId" TEXT;
CREATE INDEX "Snapshot_ProjectId_PageId_Type_Timestamp_idx" ON "Snapshot" ("ProjectId", "PageId", "Type", "Timestamp" DESC);

-- +goose Down
DROP INDEX "Snapshot_ProjectId_PageId_Type_Timestamp_idx";
ALTER TABLE "Snapshot" DROP COLUMN "PageId";
