-- Baseline: the schema as it stood when Prisma (webbuilderv2/prisma/schema.prisma,
-- last Prisma migration 20261111120000_add_element_order_key) handed ownership to Go.
-- Generated with: prisma migrate diff --from-empty --to-schema prisma/schema.prisma --script
-- Databases that Prisma already migrated must NOT run this: use 'go run ./cmd/migrate baseline'.

-- +goose Up
-- CreateSchema
CREATE SCHEMA IF NOT EXISTS "public";

-- CreateEnum
CREATE TYPE "SnapshotType" AS ENUM ('working', 'version');

-- CreateEnum
CREATE TYPE "CollaboratorRole" AS ENUM ('owner', 'editor', 'viewer');

-- CreateTable
CREATE TABLE "Element" (
    "Content" VARCHAR(2048),
    "Href" VARCHAR(2048),
    "Id" TEXT NOT NULL,
    "Name" VARCHAR(255),
    "Order" INTEGER NOT NULL DEFAULT 0,
    "OrderKey" TEXT,
    "ParentId" TEXT,
    "Src" VARCHAR(2048),
    "Styles" JSONB,
    "TailwindStyles" TEXT,
    "Type" VARCHAR(32) NOT NULL,
    "PageId" TEXT,
    "IsLocked" BOOLEAN NOT NULL DEFAULT false,
    "Settings" JSONB,

    CONSTRAINT "Element_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "Image" (
    "ImageId" TEXT NOT NULL,
    "ImageLink" VARCHAR(2048) NOT NULL DEFAULT '',
    "ImageName" VARCHAR(255),
    "UserId" TEXT NOT NULL,
    "CreatedAt" TIMESTAMP(6) NOT NULL,
    "DeletedAt" TIMESTAMP(6),
    "UpdatedAt" TIMESTAMP(6) NOT NULL,

    CONSTRAINT "PK_Images" PRIMARY KEY ("ImageId")
);

-- CreateTable
CREATE TABLE "Project" (
    "Description" TEXT,
    "Id" TEXT NOT NULL,
    "Name" VARCHAR(255) NOT NULL,
    "OwnerId" TEXT NOT NULL,
    "Published" BOOLEAN NOT NULL DEFAULT false,
    "Subdomain" VARCHAR(63),
    "CreatedAt" TIMESTAMP(6) NOT NULL,
    "DeletedAt" TIMESTAMP(6),
    "UpdatedAt" TIMESTAMP(6) NOT NULL,
    "Styles" JSONB,
    "Header" JSONB,

    CONSTRAINT "PK_Projects" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "Page" (
    "Id" TEXT NOT NULL,
    "Name" VARCHAR(255) NOT NULL,
    "Type" VARCHAR(50) NOT NULL,
    "Styles" JSONB NOT NULL,
    "ProjectId" TEXT NOT NULL,
    "CreatedAt" TIMESTAMP(6) NOT NULL,
    "UpdatedAt" TIMESTAMP(6) NOT NULL,
    "DeletedAt" TIMESTAMP(6),

    CONSTRAINT "Page_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "EventWorkflow" (
    "Id" TEXT NOT NULL,
    "ProjectId" TEXT NOT NULL,
    "Name" VARCHAR(255) NOT NULL,
    "Description" TEXT,
    "Handlers" JSONB NOT NULL DEFAULT '[]',
    "Enabled" BOOLEAN NOT NULL DEFAULT true,
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt" TIMESTAMP(6) NOT NULL,
    "CanvasData" JSONB,

    CONSTRAINT "EventWorkflow_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "Snapshot" (
    "Id" TEXT NOT NULL,
    "ProjectId" TEXT NOT NULL,
    "Elements" JSONB NOT NULL,
    "Timestamp" BIGINT NOT NULL,
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "Name" VARCHAR(255) NOT NULL,
    "Type" "SnapshotType" NOT NULL,

    CONSTRAINT "Snapshot_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "User" (
    "CreatedAt" TIMESTAMPTZ(6) NOT NULL,
    "Email" VARCHAR(255) NOT NULL,
    "FirstName" VARCHAR(100),
    "Id" TEXT NOT NULL,
    "ImageUrl" VARCHAR(2048),
    "LastName" VARCHAR(100),
    "UpdatedAt" TIMESTAMPTZ(6) NOT NULL,

    CONSTRAINT "PK_Users" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "Subscription" (
    "Id" TEXT NOT NULL,
    "UserId" TEXT NOT NULL,
    "PlanId" VARCHAR(100) NOT NULL,
    "BillingPeriod" VARCHAR(50) NOT NULL,
    "Status" VARCHAR(50) NOT NULL DEFAULT 'pending',
    "StartDate" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "EndDate" TIMESTAMP(3),
    "Amount" DOUBLE PRECISION NOT NULL,
    "Currency" VARCHAR(10) NOT NULL DEFAULT 'VND',
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt" TIMESTAMP(6) NOT NULL,
    "BankCode" VARCHAR(50),
    "CardType" VARCHAR(50),
    "Email" VARCHAR(255),
    "PayDate" TIMESTAMP(3),
    "TransactionNo" VARCHAR(100),

    CONSTRAINT "Subscription_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "ContentType" (
    "CreatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "Description" TEXT,
    "Id" TEXT NOT NULL,
    "Name" VARCHAR(255) NOT NULL,
    "UpdatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "ContentType_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "ContentField" (
    "ContentTypeId" TEXT NOT NULL,
    "Id" TEXT NOT NULL,
    "Name" VARCHAR(255) NOT NULL,
    "Required" BOOLEAN NOT NULL DEFAULT false,
    "Type" VARCHAR(100) NOT NULL,

    CONSTRAINT "ContentField_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "ContentFieldValue" (
    "ContentItemId" TEXT NOT NULL,
    "FieldId" TEXT NOT NULL,
    "Id" TEXT NOT NULL,
    "Value" TEXT,

    CONSTRAINT "ContentFieldValue_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "ContentItem" (
    "ContentTypeId" TEXT NOT NULL,
    "CreatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "Id" TEXT NOT NULL,
    "Published" BOOLEAN NOT NULL DEFAULT false,
    "Slug" VARCHAR(255) NOT NULL,
    "Title" VARCHAR(500) NOT NULL,
    "UpdatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "ContentItem_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
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

-- CreateTable
CREATE TABLE "Category" (
    "Id" TEXT NOT NULL,
    "Name" VARCHAR(100) NOT NULL,

    CONSTRAINT "Category_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "Tag" (
    "Id" TEXT NOT NULL,
    "Name" VARCHAR(100) NOT NULL,

    CONSTRAINT "Tag_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "MarketplaceItemTag" (
    "ItemId" TEXT NOT NULL,
    "TagId" TEXT NOT NULL,

    CONSTRAINT "MarketplaceItemTag_pkey" PRIMARY KEY ("ItemId","TagId")
);

-- CreateTable
CREATE TABLE "MarketplaceItemCategory" (
    "ItemId" TEXT NOT NULL,
    "CategoryId" TEXT NOT NULL,

    CONSTRAINT "MarketplaceItemCategory_pkey" PRIMARY KEY ("ItemId","CategoryId")
);

-- CreateTable
CREATE TABLE "CustomElementType" (
    "Id" TEXT NOT NULL,
    "Name" VARCHAR(255) NOT NULL,
    "Description" TEXT,
    "Category" VARCHAR(100),
    "Icon" VARCHAR(255),
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt" TIMESTAMP(6) NOT NULL,

    CONSTRAINT "CustomElementType_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "CustomElement" (
    "Id" TEXT NOT NULL,
    "Name" VARCHAR(255) NOT NULL,
    "TypeId" TEXT,
    "Description" TEXT,
    "Category" VARCHAR(100),
    "Icon" VARCHAR(255),
    "Thumbnail" VARCHAR(255),
    "Structure" JSONB NOT NULL,
    "DefaultProps" JSONB,
    "Tags" VARCHAR(500),
    "UserId" TEXT NOT NULL,
    "IsPublic" BOOLEAN NOT NULL DEFAULT false,
    "Version" VARCHAR(20) NOT NULL DEFAULT '1.0.0',
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "CustomElement_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "Invitation" (
    "Id" TEXT NOT NULL,
    "Email" VARCHAR(255) NOT NULL,
    "ProjectId" TEXT NOT NULL,
    "Role" "CollaboratorRole" NOT NULL DEFAULT 'editor',
    "Token" VARCHAR(255) NOT NULL,
    "ExpiresAt" TIMESTAMP(3) NOT NULL,
    "CreatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "AcceptedAt" TIMESTAMP(3),
    "Status" VARCHAR(50) NOT NULL DEFAULT 'pending',

    CONSTRAINT "Invitation_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "Collaborator" (
    "Id" TEXT NOT NULL,
    "UserId" TEXT NOT NULL,
    "ProjectId" TEXT NOT NULL,
    "Role" "CollaboratorRole" NOT NULL DEFAULT 'editor',
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt" TIMESTAMP(6) NOT NULL,

    CONSTRAINT "Collaborator_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
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

-- CreateTable
CREATE TABLE "CommentReaction" (
    "Id" TEXT NOT NULL,
    "CommentId" TEXT NOT NULL,
    "UserId" TEXT NOT NULL,
    "Type" VARCHAR(50) NOT NULL,
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "CommentReaction_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "ElementComment" (
    "Id" TEXT NOT NULL,
    "Content" TEXT NOT NULL,
    "AuthorId" TEXT NOT NULL,
    "ElementId" TEXT NOT NULL,
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt" TIMESTAMP(6) NOT NULL,
    "DeletedAt" TIMESTAMP(6),
    "Resolved" BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT "ElementComment_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "ElementEventWorkflow" (
    "Id" TEXT NOT NULL,
    "ElementId" TEXT NOT NULL,
    "WorkflowId" TEXT NOT NULL,
    "EventName" VARCHAR(100) NOT NULL,
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "ElementEventWorkflow_pkey" PRIMARY KEY ("Id")
);

-- CreateTable
CREATE TABLE "Notification" (
    "Id" TEXT NOT NULL,
    "UserId" TEXT NOT NULL,
    "Type" VARCHAR(100) NOT NULL,
    "Title" VARCHAR(500) NOT NULL,
    "Description" TEXT NOT NULL,
    "Read" BOOLEAN NOT NULL DEFAULT false,
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt" TIMESTAMP(6) NOT NULL,

    CONSTRAINT "Notification_pkey" PRIMARY KEY ("Id")
);

-- CreateIndex
CREATE INDEX "Element_ParentId_idx" ON "Element"("ParentId");

-- CreateIndex
CREATE INDEX "Element_PageId_idx" ON "Element"("PageId");

-- CreateIndex
CREATE INDEX "Element_Type_idx" ON "Element"("Type");

-- CreateIndex
CREATE INDEX "Element_Order_idx" ON "Element"("Order");

-- CreateIndex
CREATE INDEX "Element_IsLocked_idx" ON "Element"("IsLocked");

-- CreateIndex
CREATE INDEX "Element_PageId_Order_idx" ON "Element"("PageId", "Order");

-- CreateIndex
CREATE INDEX "Element_ParentId_Order_idx" ON "Element"("ParentId", "Order");

-- CreateIndex
CREATE INDEX "Element_PageId_OrderKey_idx" ON "Element"("PageId", "OrderKey");

-- CreateIndex
CREATE INDEX "Element_ParentId_OrderKey_idx" ON "Element"("ParentId", "OrderKey");

-- CreateIndex
CREATE INDEX "Element_PageId_Type_idx" ON "Element"("PageId", "Type");

-- CreateIndex
CREATE INDEX "IX_Images_UserId" ON "Image"("UserId");

-- CreateIndex
CREATE INDEX "Image_DeletedAt_idx" ON "Image"("DeletedAt");

-- CreateIndex
CREATE INDEX "Image_UserId_DeletedAt_idx" ON "Image"("UserId", "DeletedAt");

-- CreateIndex
CREATE INDEX "IX_Projects_OwnerId" ON "Project"("OwnerId");

-- CreateIndex
CREATE INDEX "Project_Published_idx" ON "Project"("Published");

-- CreateIndex
CREATE INDEX "Project_Subdomain_idx" ON "Project"("Subdomain");

-- CreateIndex
CREATE INDEX "Project_CreatedAt_idx" ON "Project"("CreatedAt");

-- CreateIndex
CREATE INDEX "Project_DeletedAt_idx" ON "Project"("DeletedAt");

-- CreateIndex
CREATE INDEX "Project_OwnerId_Published_idx" ON "Project"("OwnerId", "Published");

-- CreateIndex
CREATE INDEX "Project_OwnerId_DeletedAt_Published_idx" ON "Project"("OwnerId", "DeletedAt", "Published");

-- CreateIndex
CREATE INDEX "Page_ProjectId_idx" ON "Page"("ProjectId");

-- CreateIndex
CREATE INDEX "Page_Type_idx" ON "Page"("Type");

-- CreateIndex
CREATE INDEX "Page_ProjectId_Type_idx" ON "Page"("ProjectId", "Type");

-- CreateIndex
CREATE UNIQUE INDEX "Page_ProjectId_Name_key" ON "Page"("ProjectId", "Name");

-- CreateIndex
CREATE INDEX "EventWorkflow_ProjectId_idx" ON "EventWorkflow"("ProjectId");

-- CreateIndex
CREATE INDEX "EventWorkflow_Enabled_idx" ON "EventWorkflow"("Enabled");

-- CreateIndex
CREATE INDEX "EventWorkflow_ProjectId_Enabled_idx" ON "EventWorkflow"("ProjectId", "Enabled");

-- CreateIndex
CREATE UNIQUE INDEX "EventWorkflow_ProjectId_Name_key" ON "EventWorkflow"("ProjectId", "Name");

-- CreateIndex
CREATE INDEX "Snapshot_ProjectId_Timestamp_idx" ON "Snapshot"("ProjectId", "Timestamp");

-- CreateIndex
CREATE INDEX "Snapshot_ProjectId_Type_idx" ON "Snapshot"("ProjectId", "Type");

-- CreateIndex
CREATE INDEX "Snapshot_CreatedAt_idx" ON "Snapshot"("CreatedAt");

-- CreateIndex
CREATE UNIQUE INDEX "User_Email_key" ON "User"("Email");

-- CreateIndex
CREATE INDEX "Subscription_UserId_idx" ON "Subscription"("UserId");

-- CreateIndex
CREATE INDEX "Subscription_Status_idx" ON "Subscription"("Status");

-- CreateIndex
CREATE UNIQUE INDEX "ContentType_Name_key" ON "ContentType"("Name");

-- CreateIndex
CREATE UNIQUE INDEX "ContentField_ContentTypeId_Name_key" ON "ContentField"("ContentTypeId", "Name");

-- CreateIndex
CREATE UNIQUE INDEX "ContentFieldValue_ContentItemId_FieldId_key" ON "ContentFieldValue"("ContentItemId", "FieldId");

-- CreateIndex
CREATE UNIQUE INDEX "ContentItem_Slug_key" ON "ContentItem"("Slug");

-- CreateIndex
CREATE UNIQUE INDEX "MarketplaceItem_ProjectId_key" ON "MarketplaceItem"("ProjectId");

-- CreateIndex
CREATE INDEX "MarketplaceItem_AuthorId_idx" ON "MarketplaceItem"("AuthorId");

-- CreateIndex
CREATE INDEX "MarketplaceItem_TemplateType_idx" ON "MarketplaceItem"("TemplateType");

-- CreateIndex
CREATE INDEX "MarketplaceItem_Featured_idx" ON "MarketplaceItem"("Featured");

-- CreateIndex
CREATE INDEX "MarketplaceItem_ProjectId_idx" ON "MarketplaceItem"("ProjectId");

-- CreateIndex
CREATE INDEX "MarketplaceItem_CreatedAt_idx" ON "MarketplaceItem"("CreatedAt");

-- CreateIndex
CREATE INDEX "MarketplaceItem_Downloads_idx" ON "MarketplaceItem"("Downloads");

-- CreateIndex
CREATE INDEX "MarketplaceItem_Likes_idx" ON "MarketplaceItem"("Likes");

-- CreateIndex
CREATE INDEX "MarketplaceItem_Views_idx" ON "MarketplaceItem"("Views");

-- CreateIndex
CREATE INDEX "MarketplaceItem_Featured_TemplateType_idx" ON "MarketplaceItem"("Featured", "TemplateType");

-- CreateIndex
CREATE INDEX "MarketplaceItem_TemplateType_Downloads_idx" ON "MarketplaceItem"("TemplateType", "Downloads");

-- CreateIndex
CREATE UNIQUE INDEX "Category_Name_key" ON "Category"("Name");

-- CreateIndex
CREATE UNIQUE INDEX "Tag_Name_key" ON "Tag"("Name");

-- CreateIndex
CREATE INDEX "MarketplaceItemTag_TagId_idx" ON "MarketplaceItemTag"("TagId");

-- CreateIndex
CREATE INDEX "MarketplaceItemCategory_CategoryId_idx" ON "MarketplaceItemCategory"("CategoryId");

-- CreateIndex
CREATE UNIQUE INDEX "CustomElementType_Name_key" ON "CustomElementType"("Name");

-- CreateIndex
CREATE INDEX "CustomElement_UserId_idx" ON "CustomElement"("UserId");

-- CreateIndex
CREATE INDEX "CustomElement_TypeId_idx" ON "CustomElement"("TypeId");

-- CreateIndex
CREATE UNIQUE INDEX "Invitation_Token_key" ON "Invitation"("Token");

-- CreateIndex
CREATE INDEX "Invitation_ProjectId_idx" ON "Invitation"("ProjectId");

-- CreateIndex
CREATE INDEX "Invitation_Token_idx" ON "Invitation"("Token");

-- CreateIndex
CREATE INDEX "Invitation_Status_idx" ON "Invitation"("Status");

-- CreateIndex
CREATE INDEX "Invitation_Email_idx" ON "Invitation"("Email");

-- CreateIndex
CREATE INDEX "Invitation_ExpiresAt_idx" ON "Invitation"("ExpiresAt");

-- CreateIndex
CREATE INDEX "Invitation_ProjectId_Status_idx" ON "Invitation"("ProjectId", "Status");

-- CreateIndex
CREATE INDEX "Collaborator_UserId_idx" ON "Collaborator"("UserId");

-- CreateIndex
CREATE INDEX "Collaborator_ProjectId_idx" ON "Collaborator"("ProjectId");

-- CreateIndex
CREATE UNIQUE INDEX "Collaborator_UserId_ProjectId_key" ON "Collaborator"("UserId", "ProjectId");

-- CreateIndex
CREATE INDEX "Comment_AuthorId_idx" ON "Comment"("AuthorId");

-- CreateIndex
CREATE INDEX "Comment_ItemId_idx" ON "Comment"("ItemId");

-- CreateIndex
CREATE INDEX "Comment_ParentId_idx" ON "Comment"("ParentId");

-- CreateIndex
CREATE INDEX "Comment_Status_idx" ON "Comment"("Status");

-- CreateIndex
CREATE INDEX "Comment_ItemId_Status_idx" ON "Comment"("ItemId", "Status");

-- CreateIndex
CREATE INDEX "Comment_ItemId_ParentId_idx" ON "Comment"("ItemId", "ParentId");

-- CreateIndex
CREATE INDEX "Comment_CreatedAt_idx" ON "Comment"("CreatedAt");

-- CreateIndex
CREATE INDEX "CommentReaction_CommentId_idx" ON "CommentReaction"("CommentId");

-- CreateIndex
CREATE INDEX "CommentReaction_UserId_idx" ON "CommentReaction"("UserId");

-- CreateIndex
CREATE INDEX "ElementComment_AuthorId_idx" ON "ElementComment"("AuthorId");

-- CreateIndex
CREATE INDEX "ElementComment_ElementId_idx" ON "ElementComment"("ElementId");

-- CreateIndex
CREATE INDEX "ElementComment_Resolved_idx" ON "ElementComment"("Resolved");

-- CreateIndex
CREATE INDEX "ElementComment_ElementId_Resolved_idx" ON "ElementComment"("ElementId", "Resolved");

-- CreateIndex
CREATE INDEX "ElementEventWorkflow_ElementId_idx" ON "ElementEventWorkflow"("ElementId");

-- CreateIndex
CREATE INDEX "ElementEventWorkflow_WorkflowId_idx" ON "ElementEventWorkflow"("WorkflowId");

-- CreateIndex
CREATE INDEX "ElementEventWorkflow_EventName_idx" ON "ElementEventWorkflow"("EventName");

-- CreateIndex
CREATE UNIQUE INDEX "ElementEventWorkflow_ElementId_WorkflowId_EventName_key" ON "ElementEventWorkflow"("ElementId", "WorkflowId", "EventName");

-- CreateIndex
CREATE INDEX "Notification_UserId_idx" ON "Notification"("UserId");

-- CreateIndex
CREATE INDEX "Notification_Read_idx" ON "Notification"("Read");

-- CreateIndex
CREATE INDEX "Notification_CreatedAt_idx" ON "Notification"("CreatedAt");

-- AddForeignKey
ALTER TABLE "Element" ADD CONSTRAINT "Element_PageId_fkey" FOREIGN KEY ("PageId") REFERENCES "Page"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Element" ADD CONSTRAINT "Element_ParentId_fkey" FOREIGN KEY ("ParentId") REFERENCES "Element"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Image" ADD CONSTRAINT "FK_Images_Users_UserId" FOREIGN KEY ("UserId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "Project" ADD CONSTRAINT "FK_Projects_Users_OwnerId" FOREIGN KEY ("OwnerId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "Page" ADD CONSTRAINT "Page_ProjectId_fkey" FOREIGN KEY ("ProjectId") REFERENCES "Project"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "EventWorkflow" ADD CONSTRAINT "EventWorkflow_ProjectId_fkey" FOREIGN KEY ("ProjectId") REFERENCES "Project"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Snapshot" ADD CONSTRAINT "Snapshot_ProjectId_fkey" FOREIGN KEY ("ProjectId") REFERENCES "Project"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Subscription" ADD CONSTRAINT "Subscription_UserId_fkey" FOREIGN KEY ("UserId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ContentField" ADD CONSTRAINT "ContentField_ContentTypeId_fkey" FOREIGN KEY ("ContentTypeId") REFERENCES "ContentType"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ContentFieldValue" ADD CONSTRAINT "ContentFieldValue_ContentItemId_fkey" FOREIGN KEY ("ContentItemId") REFERENCES "ContentItem"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ContentFieldValue" ADD CONSTRAINT "ContentFieldValue_FieldId_fkey" FOREIGN KEY ("FieldId") REFERENCES "ContentField"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ContentItem" ADD CONSTRAINT "ContentItem_ContentTypeId_fkey" FOREIGN KEY ("ContentTypeId") REFERENCES "ContentType"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "MarketplaceItem" ADD CONSTRAINT "MarketplaceItem_AuthorId_fkey" FOREIGN KEY ("AuthorId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "MarketplaceItem" ADD CONSTRAINT "MarketplaceItem_ProjectId_fkey" FOREIGN KEY ("ProjectId") REFERENCES "Project"("Id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "MarketplaceItemTag" ADD CONSTRAINT "MarketplaceItemTag_ItemId_fkey" FOREIGN KEY ("ItemId") REFERENCES "MarketplaceItem"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "MarketplaceItemTag" ADD CONSTRAINT "MarketplaceItemTag_TagId_fkey" FOREIGN KEY ("TagId") REFERENCES "Tag"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "MarketplaceItemCategory" ADD CONSTRAINT "MarketplaceItemCategory_CategoryId_fkey" FOREIGN KEY ("CategoryId") REFERENCES "Category"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "MarketplaceItemCategory" ADD CONSTRAINT "MarketplaceItemCategory_ItemId_fkey" FOREIGN KEY ("ItemId") REFERENCES "MarketplaceItem"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "CustomElement" ADD CONSTRAINT "CustomElement_TypeId_fkey" FOREIGN KEY ("TypeId") REFERENCES "CustomElementType"("Id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "CustomElement" ADD CONSTRAINT "CustomElement_UserId_fkey" FOREIGN KEY ("UserId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Invitation" ADD CONSTRAINT "Invitation_ProjectId_fkey" FOREIGN KEY ("ProjectId") REFERENCES "Project"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Collaborator" ADD CONSTRAINT "Collaborator_ProjectId_fkey" FOREIGN KEY ("ProjectId") REFERENCES "Project"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Collaborator" ADD CONSTRAINT "Collaborator_UserId_fkey" FOREIGN KEY ("UserId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Comment" ADD CONSTRAINT "Comment_AuthorId_fkey" FOREIGN KEY ("AuthorId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Comment" ADD CONSTRAINT "Comment_ItemId_fkey" FOREIGN KEY ("ItemId") REFERENCES "MarketplaceItem"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Comment" ADD CONSTRAINT "Comment_ParentId_fkey" FOREIGN KEY ("ParentId") REFERENCES "Comment"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "CommentReaction" ADD CONSTRAINT "CommentReaction_CommentId_fkey" FOREIGN KEY ("CommentId") REFERENCES "Comment"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "CommentReaction" ADD CONSTRAINT "CommentReaction_UserId_fkey" FOREIGN KEY ("UserId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ElementComment" ADD CONSTRAINT "ElementComment_AuthorId_fkey" FOREIGN KEY ("AuthorId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ElementComment" ADD CONSTRAINT "ElementComment_ElementId_fkey" FOREIGN KEY ("ElementId") REFERENCES "Element"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ElementEventWorkflow" ADD CONSTRAINT "ElementEventWorkflow_ElementId_fkey" FOREIGN KEY ("ElementId") REFERENCES "Element"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ElementEventWorkflow" ADD CONSTRAINT "ElementEventWorkflow_WorkflowId_fkey" FOREIGN KEY ("WorkflowId") REFERENCES "EventWorkflow"("Id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Notification" ADD CONSTRAINT "Notification_UserId_fkey" FOREIGN KEY ("UserId") REFERENCES "User"("Id") ON DELETE CASCADE ON UPDATE CASCADE;


-- +goose Down
-- Irreversible: dropping the baseline would drop every table.
SELECT 1;
