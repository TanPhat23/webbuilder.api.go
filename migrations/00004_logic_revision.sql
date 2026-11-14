-- +goose Up
-- Incremented by realtime on every committed logic update; sent with each
-- logic document so clients can drop one older than what they already have.
ALTER TABLE "Project" ADD COLUMN "LogicRevision" BIGINT NOT NULL DEFAULT 0;
ALTER TABLE "Page" ADD COLUMN "LogicRevision" BIGINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE "Page" DROP COLUMN "LogicRevision";
ALTER TABLE "Project" DROP COLUMN "LogicRevision";
