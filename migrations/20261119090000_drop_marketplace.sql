-- +goose Up
-- The template marketplace and the comments that hung off its items are gone.
-- Dropping the tables deletes their rows; the Down migration recreates the
-- empty schema only. Take a backup first if the data matters.
DROP TABLE IF EXISTS "CommentReaction";
DROP TABLE IF EXISTS "Comment";
DROP TABLE IF EXISTS "MarketplaceItemTag";
DROP TABLE IF EXISTS "MarketplaceItemCategory";
DROP TABLE IF EXISTS "MarketplaceItem";
DROP TABLE IF EXISTS "Tag";
DROP TABLE IF EXISTS "Category";

-- +goose Down
-- Schema only: rows dropped by Up are not restored.
CREATE TABLE "MarketplaceItem" (
    "Id" TEXT NOT NULL,
    "Title" VARCHAR(500) NOT NULL,
    "Description" TEXT NOT NULL,
    "Preview" VARCHAR(2048),
    "TemplateType" VARCHAR(50) NOT NULL DEFAULT 'block',
    "Featured" BOOLEAN NOT NULL DEFAULT false,
    "PageCount" INTEGER,
    "Downloads" INTEGER NOT NULL DEFAULT 0,
    "Likes" INTEGER NOT NULL DEFAULT 0,
    "AuthorId" TEXT NOT NULL,
    "AuthorName" VARCHAR(255) NOT NULL,
    "Verified" BOOLEAN NOT NULL DEFAULT false,
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt" TIMESTAMP(6) NOT NULL,
    "DeletedAt" TIMESTAMP(6),
    "ProjectId" TEXT,
    "Views" INTEGER NOT NULL DEFAULT 0,

    CONSTRAINT "MarketplaceItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE "Category" (
    "Id" TEXT NOT NULL,
    "Name" VARCHAR(100) NOT NULL,

    CONSTRAINT "Category_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE "Tag" (
    "Id" TEXT NOT NULL,
    "Name" VARCHAR(100) NOT NULL,

    CONSTRAINT "Tag_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE "MarketplaceItemTag" (
    "ItemId" TEXT NOT NULL,
    "TagId" TEXT NOT NULL,

    CONSTRAINT "MarketplaceItemTag_pkey" PRIMARY KEY ("ItemId","TagId")
);

CREATE TABLE "MarketplaceItemCategory" (
    "ItemId" TEXT NOT NULL,
    "CategoryId" TEXT NOT NULL,

    CONSTRAINT "MarketplaceItemCategory_pkey" PRIMARY KEY ("ItemId","CategoryId")
);

CREATE TABLE "Comment" (
    "Id" TEXT NOT NULL,
    "Content" TEXT NOT NULL,
    "AuthorId" TEXT NOT NULL,
    "ItemId" TEXT NOT NULL,
    "ParentId" TEXT,
    "Status" VARCHAR(50) NOT NULL DEFAULT 'published',
    "Edited" BOOLEAN NOT NULL DEFAULT false,
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt" TIMESTAMP(6) NOT NULL,
    "DeletedAt" TIMESTAMP(6),

    CONSTRAINT "Comment_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE "CommentReaction" (
    "Id" TEXT NOT NULL,
    "CommentId" TEXT NOT NULL,
    "UserId" TEXT NOT NULL,
    "Type" VARCHAR(50) NOT NULL,
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "CommentReaction_pkey" PRIMARY KEY ("Id")
);

CREATE UNIQUE INDEX "MarketplaceItem_ProjectId_key" ON "MarketplaceItem"("ProjectId");
CREATE INDEX "MarketplaceItem_AuthorId_idx" ON "MarketplaceItem"("AuthorId");
CREATE INDEX "MarketplaceItem_TemplateType_idx" ON "MarketplaceItem"("TemplateType");
CREATE INDEX "MarketplaceItem_Featured_idx" ON "MarketplaceItem"("Featured");
CREATE INDEX "MarketplaceItem_ProjectId_idx" ON "MarketplaceItem"("ProjectId");
CREATE INDEX "MarketplaceItem_CreatedAt_idx" ON "MarketplaceItem"("CreatedAt");
CREATE INDEX "MarketplaceItem_Downloads_idx" ON "MarketplaceItem"("Downloads");
CREATE INDEX "MarketplaceItem_Likes_idx" ON "MarketplaceItem"("Likes");
CREATE INDEX "MarketplaceItem_Views_idx" ON "MarketplaceItem"("Views");
CREATE INDEX "MarketplaceItem_Featured_TemplateType_idx" ON "MarketplaceItem"("Featured", "TemplateType");
CREATE INDEX "MarketplaceItem_TemplateType_Downloads_idx" ON "MarketplaceItem"("TemplateType", "Downloads");
CREATE UNIQUE INDEX "Category_Name_key" ON "Category"("Name");
CREATE UNIQUE INDEX "Tag_Name_key" ON "Tag"("Name");
CREATE INDEX "MarketplaceItemTag_TagId_idx" ON "MarketplaceItemTag"("TagId");
CREATE INDEX "MarketplaceItemCategory_CategoryId_idx" ON "MarketplaceItemCategory"("CategoryId");
CREATE INDEX "Comment_AuthorId_idx" ON "Comment"("AuthorId");
CREATE INDEX "Comment_ItemId_idx" ON "Comment"("ItemId");
CREATE INDEX "Comment_ParentId_idx" ON "Comment"("ParentId");
CREATE INDEX "Comment_Status_idx" ON "Comment"("Status");
CREATE INDEX "Comment_ItemId_Status_idx" ON "Comment"("ItemId", "Status");
CREATE INDEX "Comment_ItemId_ParentId_idx" ON "Comment"("ItemId", "ParentId");
CREATE INDEX "Comment_CreatedAt_idx" ON "Comment"("CreatedAt");
CREATE INDEX "CommentReaction_CommentId_idx" ON "CommentReaction"("CommentId");
CREATE INDEX "CommentReaction_UserId_idx" ON "CommentReaction"("UserId");

ALTER TABLE "MarketplaceItem" ADD CONSTRAINT "MarketplaceItem_AuthorId_fkey" FOREIGN KEY ("AuthorId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "MarketplaceItem" ADD CONSTRAINT "MarketplaceItem_ProjectId_fkey" FOREIGN KEY ("ProjectId") REFERENCES "Project"("Id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "MarketplaceItemTag" ADD CONSTRAINT "MarketplaceItemTag_ItemId_fkey" FOREIGN KEY ("ItemId") REFERENCES "MarketplaceItem"("Id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "MarketplaceItemTag" ADD CONSTRAINT "MarketplaceItemTag_TagId_fkey" FOREIGN KEY ("TagId") REFERENCES "Tag"("Id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "MarketplaceItemCategory" ADD CONSTRAINT "MarketplaceItemCategory_CategoryId_fkey" FOREIGN KEY ("CategoryId") REFERENCES "Category"("Id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "MarketplaceItemCategory" ADD CONSTRAINT "MarketplaceItemCategory_ItemId_fkey" FOREIGN KEY ("ItemId") REFERENCES "MarketplaceItem"("Id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "Comment" ADD CONSTRAINT "Comment_AuthorId_fkey" FOREIGN KEY ("AuthorId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "Comment" ADD CONSTRAINT "Comment_ItemId_fkey" FOREIGN KEY ("ItemId") REFERENCES "MarketplaceItem"("Id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "Comment" ADD CONSTRAINT "Comment_ParentId_fkey" FOREIGN KEY ("ParentId") REFERENCES "Comment"("Id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "CommentReaction" ADD CONSTRAINT "CommentReaction_CommentId_fkey" FOREIGN KEY ("CommentId") REFERENCES "Comment"("Id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "CommentReaction" ADD CONSTRAINT "CommentReaction_UserId_fkey" FOREIGN KEY ("UserId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE CASCADE;
