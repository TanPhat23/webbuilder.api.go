-- +goose Up
ALTER TABLE public."Project"
    ADD COLUMN IF NOT EXISTS "AuthConfig" jsonb;

ALTER TABLE public."Page"
    ADD COLUMN IF NOT EXISTS "RequireAuth" boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE public."Page"
    DROP COLUMN IF EXISTS "RequireAuth";

ALTER TABLE public."Project"
    DROP COLUMN IF EXISTS "AuthConfig";
