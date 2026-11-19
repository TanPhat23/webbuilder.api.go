-- +goose Up
-- Existing duplicate names must be resolved before creating this index.
CREATE UNIQUE INDEX "CustomElement_UserId_Name_key"
    ON "CustomElement" ("UserId", lower("Name"));

-- +goose Down
DROP INDEX "CustomElement_UserId_Name_key";
