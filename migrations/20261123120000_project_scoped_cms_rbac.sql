-- +goose Up
-- Scope CMS ContentTypes to a Project so CMS permissions (cms:view, cms:edit,
-- cms:publish) can be enforced per project role instead of globally.
ALTER TABLE "ContentType"
    ADD COLUMN IF NOT EXISTS "ProjectId" TEXT;

DROP INDEX IF EXISTS "ContentType_Name_key";

CREATE INDEX IF NOT EXISTS "ContentType_ProjectId_idx"
    ON "ContentType"("ProjectId");

CREATE UNIQUE INDEX IF NOT EXISTS "ContentType_ProjectId_Name_key"
    ON "ContentType"("ProjectId", "Name")
    WHERE "ProjectId" IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS "ContentType_Unscoped_Name_key"
    ON "ContentType"("Name")
    WHERE "ProjectId" IS NULL;

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'ContentType_ProjectId_fkey'
    ) THEN
        ALTER TABLE "ContentType"
            ADD CONSTRAINT "ContentType_ProjectId_fkey"
            FOREIGN KEY ("ProjectId") REFERENCES "Project"("Id")
            ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE "ContentType" DROP CONSTRAINT IF EXISTS "ContentType_ProjectId_fkey";
DROP INDEX IF EXISTS "ContentType_ProjectId_Name_key";
DROP INDEX IF EXISTS "ContentType_Unscoped_Name_key";
DROP INDEX IF EXISTS "ContentType_ProjectId_idx";
CREATE UNIQUE INDEX IF NOT EXISTS "ContentType_Name_key" ON "ContentType"("Name");
ALTER TABLE "ContentType" DROP COLUMN IF EXISTS "ProjectId";
