-- Migration: Create Comment and CommentReaction tables
-- Description: Adds comment functionality to marketplace items with reactions and threaded comments

-- Create CommentStatus enum if it doesn't exist
DO $$ BEGIN
    CREATE TYPE "CommentStatus" AS ENUM ('published', 'pending', 'flagged', 'deleted');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create Comment table
CREATE TABLE IF NOT EXISTS public."Comment" (
    "Id" VARCHAR(255) NOT NULL PRIMARY KEY,
    "Content" TEXT NOT NULL,
    "AuthorId" VARCHAR(255) NOT NULL,
    "ItemId" VARCHAR(255) NOT NULL,
    "ParentId" VARCHAR(255),
    "Status" VARCHAR(50) NOT NULL DEFAULT 'published',
    "Edited" BOOLEAN NOT NULL DEFAULT false,
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "DeletedAt" TIMESTAMP(6),

    -- Foreign key constraints
    CONSTRAINT "FK_Comment_User" FOREIGN KEY ("AuthorId")
        REFERENCES public."User"("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_Comment_MarketplaceItem" FOREIGN KEY ("ItemId")
        REFERENCES public."MarketplaceItem"("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_Comment_Parent" FOREIGN KEY ("ParentId")
        REFERENCES public."Comment"("Id") ON DELETE CASCADE
);

-- Create indexes for Comment table
CREATE INDEX IF NOT EXISTS "IX_Comment_AuthorId" ON public."Comment"("AuthorId");
CREATE INDEX IF NOT EXISTS "IX_Comment_ItemId" ON public."Comment"("ItemId");
CREATE INDEX IF NOT EXISTS "IX_Comment_ParentId" ON public."Comment"("ParentId");
CREATE INDEX IF NOT EXISTS "IX_Comment_Status" ON public."Comment"("Status");
CREATE INDEX IF NOT EXISTS "IX_Comment_CreatedAt" ON public."Comment"("CreatedAt" DESC);
CREATE INDEX IF NOT EXISTS "IX_Comment_DeletedAt" ON public."Comment"("DeletedAt");

-- Create CommentReaction table
CREATE TABLE IF NOT EXISTS public."CommentReaction" (
    "Id" VARCHAR(255) NOT NULL PRIMARY KEY,
    "CommentId" VARCHAR(255) NOT NULL,
    "UserId" VARCHAR(255) NOT NULL,
    "Type" VARCHAR(50) NOT NULL,
    "CreatedAt" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraints
    CONSTRAINT "FK_CommentReaction_Comment" FOREIGN KEY ("CommentId")
        REFERENCES public."Comment"("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_CommentReaction_User" FOREIGN KEY ("UserId")
        REFERENCES public."User"("Id") ON DELETE CASCADE,

    -- Ensure unique reaction per user per comment per type
    CONSTRAINT "UQ_CommentReaction_CommentId_UserId_Type"
        UNIQUE ("CommentId", "UserId", "Type")
);

-- Create indexes for CommentReaction table
CREATE INDEX IF NOT EXISTS "IX_CommentReaction_CommentId" ON public."CommentReaction"("CommentId");
CREATE INDEX IF NOT EXISTS "IX_CommentReaction_UserId" ON public."CommentReaction"("UserId");
CREATE INDEX IF NOT EXISTS "IX_CommentReaction_Type" ON public."CommentReaction"("Type");

-- Add comments to tables for documentation
COMMENT ON TABLE public."Comment" IS 'Stores comments on marketplace items with support for threaded replies';
COMMENT ON TABLE public."CommentReaction" IS 'Stores user reactions (likes, hearts, etc.) to comments';

COMMENT ON COLUMN public."Comment"."Id" IS 'Unique identifier for the comment';
COMMENT ON COLUMN public."Comment"."Content" IS 'The comment text content';
COMMENT ON COLUMN public."Comment"."AuthorId" IS 'User who created the comment';
COMMENT ON COLUMN public."Comment"."ItemId" IS 'Marketplace item the comment belongs to';
COMMENT ON COLUMN public."Comment"."ParentId" IS 'Parent comment ID for threaded replies (NULL for top-level comments)';
COMMENT ON COLUMN public."Comment"."Status" IS 'Moderation status: published, pending, flagged, or deleted';
COMMENT ON COLUMN public."Comment"."Edited" IS 'Whether the comment has been edited after creation';
COMMENT ON COLUMN public."Comment"."CreatedAt" IS 'When the comment was created';
COMMENT ON COLUMN public."Comment"."UpdatedAt" IS 'When the comment was last updated';
COMMENT ON COLUMN public."Comment"."DeletedAt" IS 'Soft delete timestamp';

COMMENT ON COLUMN public."CommentReaction"."Id" IS 'Unique identifier for the reaction';
COMMENT ON COLUMN public."CommentReaction"."CommentId" IS 'Comment the reaction is on';
COMMENT ON COLUMN public."CommentReaction"."UserId" IS 'User who created the reaction';
COMMENT ON COLUMN public."CommentReaction"."Type" IS 'Type of reaction (like, heart, helpful, etc.)';
COMMENT ON COLUMN public."CommentReaction"."CreatedAt" IS 'When the reaction was created';
