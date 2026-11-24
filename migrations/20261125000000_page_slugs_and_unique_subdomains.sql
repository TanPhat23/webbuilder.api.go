-- +goose Up
ALTER TABLE "Page" ADD COLUMN IF NOT EXISTS "Slug" VARCHAR(255);

-- Compute candidate slug per active route page; Home/Index/blank becomes ''.
WITH base AS (
    SELECT
        "Id",
        "ProjectId",
        "CreatedAt",
        CASE
            WHEN "Type" = 'cp' THEN ''
            WHEN LOWER(BTRIM("Name")) IN ('', 'home', 'index') THEN ''
            ELSE COALESCE(
                NULLIF(
                    BTRIM(REGEXP_REPLACE(LOWER(BTRIM("Name")), '[^a-z0-9]+', '-', 'g'), '-'),
                    ''
                ),
                LOWER(LEFT("Id", 8))
            )
        END AS candidate
    FROM "Page"
),
ranked AS (
    SELECT
        "Id",
        "Type",
        "DeletedAt",
        candidate,
        ROW_NUMBER() OVER (
            PARTITION BY "ProjectId", candidate
            ORDER BY "CreatedAt" ASC, "Id" ASC
        ) AS rn
    FROM (
        SELECT b.*, p."Type", p."DeletedAt"
        FROM base b
        JOIN "Page" p ON p."Id" = b."Id"
        WHERE p."DeletedAt" IS NULL AND p."Type" != 'cp'
    ) active_routes
)
UPDATE "Page" p
SET "Slug" = COALESCE(
    (
        SELECT CASE
            WHEN r.rn = 1 THEN r.candidate
            WHEN r.candidate = '' THEN 'page-' || r.rn::text
            ELSE r.candidate || '-' || r.rn::text
        END
        FROM ranked r
        WHERE r."Id" = p."Id"
    ),
    ''
)
WHERE p."Slug" IS NULL;

ALTER TABLE "Page" ALTER COLUMN "Slug" SET DEFAULT '';
ALTER TABLE "Page" ALTER COLUMN "Slug" SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS "Page_ProjectId_Slug_key"
    ON "Page" ("ProjectId", "Slug")
    WHERE "DeletedAt" IS NULL AND "Type" != 'cp';

-- Normalize subdomains and clear any accidental duplicates before enforcing uniqueness.
UPDATE "Project"
SET "Subdomain" = NULLIF(LOWER(BTRIM("Subdomain")), '')
WHERE "Subdomain" IS NOT NULL;

WITH dup_subdomains AS (
    SELECT
        "Id",
        ROW_NUMBER() OVER (
            PARTITION BY LOWER("Subdomain")
            ORDER BY "Published" DESC, "CreatedAt" ASC, "Id" ASC
        ) AS rn
    FROM "Project"
    WHERE "DeletedAt" IS NULL AND "Subdomain" IS NOT NULL
)
UPDATE "Project" p
SET "Subdomain" = NULL
FROM dup_subdomains d
WHERE p."Id" = d."Id" AND d.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS "Project_Subdomain_key"
    ON "Project" (LOWER("Subdomain"))
    WHERE "DeletedAt" IS NULL AND "Subdomain" IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS "Project_Subdomain_key";
DROP INDEX IF EXISTS "Page_ProjectId_Slug_key";
ALTER TABLE "Page" DROP COLUMN IF EXISTS "Slug";
