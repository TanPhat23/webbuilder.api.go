-- +goose Up
ALTER TABLE "Page" ADD COLUMN IF NOT EXISTS "ContentTypeId" TEXT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'Page_ContentTypeId_fkey'
    ) THEN
        ALTER TABLE "Page"
            ADD CONSTRAINT "Page_ContentTypeId_fkey"
            FOREIGN KEY ("ContentTypeId") REFERENCES "ContentType"("Id")
            ON DELETE SET NULL ON UPDATE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS "Page_ContentTypeId_idx"
    ON "Page" ("ContentTypeId")
    WHERE "ContentTypeId" IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS "Page_ContentTypeId_idx";
ALTER TABLE "Page" DROP CONSTRAINT IF EXISTS "Page_ContentTypeId_fkey";
ALTER TABLE "Page" DROP COLUMN IF EXISTS "ContentTypeId";
