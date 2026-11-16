-- +goose Up
CREATE TABLE "AgentToken" (
    "Id"         VARCHAR(255) PRIMARY KEY,
    "UserId"     VARCHAR(255) NOT NULL,
    "Name"       VARCHAR(100) NOT NULL,
    "Prefix"     VARCHAR(16)  NOT NULL,
    "TokenHash"  CHAR(64)     NOT NULL,
    "CreatedAt"  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    "LastUsedAt" TIMESTAMPTZ,
    "RevokedAt"  TIMESTAMPTZ
);
CREATE UNIQUE INDEX "AgentToken_TokenHash_key" ON "AgentToken" ("TokenHash");
CREATE INDEX "AgentToken_UserId_idx" ON "AgentToken" ("UserId");

-- +goose Down
DROP TABLE "AgentToken";
