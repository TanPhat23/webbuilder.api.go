-- +goose Up
ALTER TABLE public."ContentType"
    ADD COLUMN IF NOT EXISTS "Slug" varchar(255),
    ADD COLUMN IF NOT EXISTS "PublicRead" boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS "PublicCreate" boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS "DefaultPublished" boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS "RateLimitPerMin" integer NOT NULL DEFAULT 30;

ALTER TABLE public."ContentItem"
    ADD COLUMN IF NOT EXISTS "DeletedAt" timestamptz;

UPDATE public."ContentType"
SET "Slug" = COALESCE(
    NULLIF(TRIM(BOTH '-' FROM REGEXP_REPLACE(LOWER("Name"), '[^a-z0-9]+', '-', 'g')), ''),
    'collection'
) || '-' || RIGHT(REPLACE("Id", '-', ''), 6)
WHERE "Slug" IS NULL OR "Slug" = '';

ALTER TABLE public."ContentType"
    ALTER COLUMN "Slug" SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_contenttype_project_slug
    ON public."ContentType" ("ProjectId", "Slug")
    WHERE "ProjectId" IS NOT NULL;

CREATE TABLE IF NOT EXISTS public."ProjectWebhook" (
    "Id" varchar(255) PRIMARY KEY,
    "ProjectId" varchar(255) NOT NULL REFERENCES public."Project"("Id") ON DELETE CASCADE,
    "ContentTypeId" varchar(255) NULL REFERENCES public."ContentType"("Id") ON DELETE CASCADE,
    "Name" varchar(255) NOT NULL,
    "Url" text NOT NULL,
    "Secret" varchar(255) NOT NULL,
    "Events" jsonb NOT NULL DEFAULT '[]'::jsonb,
    "Enabled" boolean NOT NULL DEFAULT true,
    "CreatedAt" timestamptz NOT NULL DEFAULT now(),
    "UpdatedAt" timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_project_webhook_project_enabled
    ON public."ProjectWebhook" ("ProjectId", "Enabled");

CREATE TABLE IF NOT EXISTS public."WebhookDelivery" (
    "Id" varchar(255) PRIMARY KEY,
    "WebhookId" varchar(255) NOT NULL REFERENCES public."ProjectWebhook"("Id") ON DELETE CASCADE,
    "ProjectId" varchar(255) NOT NULL,
    "Event" varchar(64) NOT NULL,
    "Payload" jsonb NOT NULL,
    "Status" varchar(32) NOT NULL DEFAULT 'pending',
    "HttpStatus" integer NULL,
    "ResponseBody" text NULL,
    "ErrorMessage" text NULL,
    "Attempts" integer NOT NULL DEFAULT 0,
    "NextRetryAt" timestamptz NULL,
    "CreatedAt" timestamptz NOT NULL DEFAULT now(),
    "CompletedAt" timestamptz NULL
);

CREATE INDEX IF NOT EXISTS idx_webhook_delivery_webhook_created
    ON public."WebhookDelivery" ("WebhookId", "CreatedAt" DESC);

CREATE INDEX IF NOT EXISTS idx_webhook_delivery_pending_retry
    ON public."WebhookDelivery" ("Status", "NextRetryAt")
    WHERE "Status" = 'pending';

-- +goose Down
DROP TABLE IF EXISTS public."WebhookDelivery";
DROP TABLE IF EXISTS public."ProjectWebhook";
DROP INDEX IF EXISTS idx_contenttype_project_slug;
ALTER TABLE public."ContentType"
    DROP COLUMN IF EXISTS "RateLimitPerMin",
    DROP COLUMN IF EXISTS "DefaultPublished",
    DROP COLUMN IF EXISTS "PublicCreate",
    DROP COLUMN IF EXISTS "PublicRead",
    DROP COLUMN IF EXISTS "Slug";
