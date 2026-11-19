-- +goose Up
-- Indexes that duplicate, or are a leading prefix of, another index on the same
-- table. Postgres answers the same lookups from the wider index, so these only
-- add cost to every insert, update and delete (the editor rewrites whole pages).
DROP INDEX IF EXISTS "Element_PageId_idx"; -- prefix of Element_PageId_Order_idx
DROP INDEX IF EXISTS "Element_ParentId_idx"; -- prefix of Element_ParentId_Order_idx
DROP INDEX IF EXISTS "Page_ProjectId_idx"; -- prefix of Page_ProjectId_Type_idx and Page_ProjectId_Name_key
DROP INDEX IF EXISTS "Invitation_Token_idx"; -- same column as the unique Invitation_Token_key
DROP INDEX IF EXISTS "Invitation_ProjectId_idx"; -- prefix of Invitation_ProjectId_Status_idx
DROP INDEX IF EXISTS "Collaborator_UserId_idx"; -- prefix of Collaborator_UserId_ProjectId_key
DROP INDEX IF EXISTS "ElementComment_ElementId_idx"; -- prefix of ElementComment_ElementId_Resolved_idx
DROP INDEX IF EXISTS "ElementEventWorkflow_ElementId_idx"; -- prefix of ElementEventWorkflow_ElementId_WorkflowId_EventName_key
DROP INDEX IF EXISTS "EventWorkflow_ProjectId_idx"; -- prefix of EventWorkflow_ProjectId_Enabled_idx and EventWorkflow_ProjectId_Name_key

-- +goose Down
CREATE INDEX "Element_PageId_idx" ON "Element"("PageId");
CREATE INDEX "Element_ParentId_idx" ON "Element"("ParentId");
CREATE INDEX "Page_ProjectId_idx" ON "Page"("ProjectId");
CREATE INDEX "Invitation_Token_idx" ON "Invitation"("Token");
CREATE INDEX "Invitation_ProjectId_idx" ON "Invitation"("ProjectId");
CREATE INDEX "Collaborator_UserId_idx" ON "Collaborator"("UserId");
CREATE INDEX "ElementComment_ElementId_idx" ON "ElementComment"("ElementId");
CREATE INDEX "ElementEventWorkflow_ElementId_idx" ON "ElementEventWorkflow"("ElementId");
CREATE INDEX "EventWorkflow_ProjectId_idx" ON "EventWorkflow"("ProjectId");
