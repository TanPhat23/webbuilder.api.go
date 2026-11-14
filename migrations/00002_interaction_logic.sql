-- +goose Up
ALTER TABLE "Project" ADD COLUMN "Logic" jsonb;
ALTER TABLE "Page" ADD COLUMN "Logic" jsonb;

-- +goose Down
ALTER TABLE "Page" DROP COLUMN "Logic";
ALTER TABLE "Project" DROP COLUMN "Logic";
