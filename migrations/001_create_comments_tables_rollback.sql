-- Rollback Migration: Drop Comment and CommentReaction tables
-- Description: Removes comment functionality from marketplace items

-- Drop tables in reverse order (child tables first)
DROP TABLE IF EXISTS public."CommentReaction" CASCADE;
DROP TABLE IF EXISTS public."Comment" CASCADE;

-- Drop the enum type if it exists
DROP TYPE IF EXISTS "CommentStatus" CASCADE;

-- Note: This rollback script will remove all comment and reaction data
-- Make sure to backup data before running this migration if needed
