import { z } from "zod";

export const UserSchema = z.object({
  id: z.string(),
  email: z.string().email(),
  firstName: z.string().optional(),
  lastName: z.string().optional(),
  imageUrl: z.string().optional(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
});

export type User = z.infer<typeof UserSchema>;

export const ProjectSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  styles: z.unknown().optional(),
  header: z.unknown().optional(),
  published: z.boolean(),
  subdomain: z.string().optional(),
  ownerId: z.string(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
  deletedAt: z.string().datetime().optional(),
});

export type Project = z.infer<typeof ProjectSchema>;

export const PageSchema = z.object({
  Id: z.string(),
  Name: z.string(),
  Type: z.string(),
  Styles: z.unknown().optional(),
  ProjectId: z.string(),
  CreatedAt: z.string().datetime().optional(),
  UpdatedAt: z.string().datetime().optional(),
});

export type Page = z.infer<typeof PageSchema>;

export const ImageSchema = z.object({
  imageId: z.string(),
  imageLink: z.string(),
  imageName: z.string().optional(),
  userId: z.string(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
});

export type Image = z.infer<typeof ImageSchema>;

export const CreatePageRequestSchema = z.object({
  name: z.string().min(1),
  type: z.string().min(1),
  styles: z.unknown().optional(),
});

export type CreatePageRequest = z.infer<typeof CreatePageRequestSchema>;

export const UpdatePageRequestSchema = z.object({
  name: z.string().min(1).optional(),
  type: z.string().min(1).optional(),
  styles: z.unknown().optional(),
});

export type UpdatePageRequest = z.infer<typeof UpdatePageRequestSchema>;

export const UpdateProjectRequestSchema = z.object({
  name: z.string().optional(),
  description: z.string().optional(),
  published: z.boolean().optional(),
  subdomain: z.string().optional(),
  styles: z.unknown().optional(),
  header: z.unknown().optional(),
});

export type UpdateProjectRequest = z.infer<typeof UpdateProjectRequestSchema>;

// ============================================================================
// ELEMENT SCHEMAS
// ============================================================================

export const ElementSchema = z.object({
  id: z.string(),
  type: z.string(),
  pageId: z.string().optional(),
  parentId: z.string().optional(),
  name: z.string().optional(),
  content: z.string().optional(),
  href: z.string().optional(),
  src: z.string().optional(),
  tailwindStyles: z.string().optional(),
  styles: z.unknown().optional(),
  settings: z.unknown().optional(),
  order: z.number().optional(),
});

export type Element = z.infer<typeof ElementSchema>;

// ============================================================================
// COLLABORATOR SCHEMAS
// ============================================================================

export const CollaboratorSchema = z.object({
  id: z.string(),
  projectId: z.string(),
  userId: z.string(),
  role: z.enum(["owner", "editor", "viewer"]),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
  user: UserSchema.optional(),
});

export type Collaborator = z.infer<typeof CollaboratorSchema>;

export const CreateCollaboratorRequestSchema = z.object({
  projectId: z.string().min(1),
  userId: z.string().min(1),
  role: z.enum(["owner", "editor", "viewer"]),
});

export type CreateCollaboratorRequest = z.infer<
  typeof CreateCollaboratorRequestSchema
>;

export const UpdateCollaboratorRoleRequestSchema = z.object({
  role: z.enum(["owner", "editor", "viewer"]),
});

export type UpdateCollaboratorRoleRequest = z.infer<
  typeof UpdateCollaboratorRoleRequestSchema
>;

// ============================================================================
// CUSTOM ELEMENT SCHEMAS
// ============================================================================

export const CustomElementTypeSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  category: z.string().optional(),
  icon: z.string().optional(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
});

export type CustomElementType = z.infer<typeof CustomElementTypeSchema>;

export const CreateCustomElementTypeRequestSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
  category: z.string().optional(),
  icon: z.string().optional(),
});

export type CreateCustomElementTypeRequest = z.infer<
  typeof CreateCustomElementTypeRequestSchema
>;

export const UpdateCustomElementTypeRequestSchema = z.object({
  name: z.string().min(1).optional(),
  description: z.string().optional(),
  category: z.string().optional(),
  icon: z.string().optional(),
});

export type UpdateCustomElementTypeRequest = z.infer<
  typeof UpdateCustomElementTypeRequestSchema
>;

export const CustomElementSchema = z.object({
  id: z.string(),
  userId: z.string(),
  name: z.string(),
  version: z.string().optional(),
  typeId: z.string().optional(),
  description: z.string().optional(),
  category: z.string().optional(),
  icon: z.string().optional(),
  thumbnail: z.string().optional(),
  tags: z.string().optional(),
  structure: z.unknown(),
  defaultProps: z.unknown().optional(),
  isPublic: z.boolean(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
});

export type CustomElement = z.infer<typeof CustomElementSchema>;

export const CreateCustomElementRequestSchema = z.object({
  name: z.string().min(1),
  typeId: z.string().optional(),
  description: z.string().optional(),
  category: z.string().optional(),
  icon: z.string().optional(),
  thumbnail: z.string().optional(),
  tags: z.string().optional(),
  structure: z.unknown(),
  defaultProps: z.unknown().optional(),
  isPublic: z.boolean().optional(),
});

export type CreateCustomElementRequest = z.infer<
  typeof CreateCustomElementRequestSchema
>;

export const UpdateCustomElementRequestSchema = z.object({
  name: z.string().min(1).optional(),
  description: z.string().optional(),
  category: z.string().optional(),
  icon: z.string().optional(),
  thumbnail: z.string().optional(),
  tags: z.string().optional(),
  structure: z.unknown().optional(),
  defaultProps: z.unknown().optional(),
  isPublic: z.boolean().optional(),
});

export type UpdateCustomElementRequest = z.infer<
  typeof UpdateCustomElementRequestSchema
>;

// ============================================================================
// ELEMENT COMMENT SCHEMAS
// ============================================================================

export const ElementCommentSchema = z.object({
  id: z.string(),
  content: z.string(),
  authorId: z.string(),
  elementId: z.string(),
  resolved: z.boolean(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
  author: CommentAuthorSchema.optional(),
});

export type ElementComment = z.infer<typeof ElementCommentSchema>;

export const CreateElementCommentRequestSchema = z.object({
  content: z.string().min(1).max(5000),
  elementId: z.string().min(1),
});

export type CreateElementCommentRequest = z.infer<
  typeof CreateElementCommentRequestSchema
>;

export const UpdateElementCommentRequestSchema = z.object({
  content: z.string().min(1).max(5000).optional(),
  resolved: z.boolean().optional(),
});

export type UpdateElementCommentRequest = z.infer<
  typeof UpdateElementCommentRequestSchema
>;

// ============================================================================
// EVENT WORKFLOW SCHEMAS
// ============================================================================

export const EventWorkflowSchema = z.object({
  id: z.string(),
  projectId: z.string(),
  name: z.string(),
  description: z.string().optional(),
  handlers: z.unknown().optional(),
  canvasData: z.unknown().optional(),
  enabled: z.boolean(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
});

export type EventWorkflow = z.infer<typeof EventWorkflowSchema>;

export const CreateEventWorkflowRequestSchema = z.object({
  projectId: z.string().min(1),
  name: z.string().min(1),
  description: z.string().optional(),
  handlers: z.unknown().optional(),
  canvasData: z.unknown().optional(),
});

export type CreateEventWorkflowRequest = z.infer<
  typeof CreateEventWorkflowRequestSchema
>;

export const UpdateEventWorkflowRequestSchema = z.object({
  name: z.string().min(1).optional(),
  description: z.string().optional(),
  handlers: z.unknown().optional(),
  canvasData: z.unknown().optional(),
});

export type UpdateEventWorkflowRequest = z.infer<
  typeof UpdateEventWorkflowRequestSchema
>;

export const UpdateEventWorkflowEnabledRequestSchema = z.object({
  enabled: z.boolean(),
});

export type UpdateEventWorkflowEnabledRequest = z.infer<
  typeof UpdateEventWorkflowEnabledRequestSchema
>;

// ============================================================================
// ELEMENT EVENT WORKFLOW SCHEMAS
// ============================================================================

export const ElementEventWorkflowSchema = z.object({
  id: z.string(),
  elementId: z.string(),
  workflowId: z.string(),
  eventId: z.string().optional(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
});

export type ElementEventWorkflow = z.infer<typeof ElementEventWorkflowSchema>;

export const CreateElementEventWorkflowRequestSchema = z.object({
  elementId: z.string().min(1),
  workflowId: z.string().min(1),
  eventId: z.string().optional(),
});

export type CreateElementEventWorkflowRequest = z.infer<
  typeof CreateElementEventWorkflowRequestSchema
>;

export const UpdateElementEventWorkflowRequestSchema = z.object({
  eventId: z.string().optional(),
});

export type UpdateElementEventWorkflowRequest = z.infer<
  typeof UpdateElementEventWorkflowRequestSchema
>;

// ============================================================================
// INVITATION SCHEMAS
// ============================================================================

export const InvitationSchema = z.object({
  id: z.string(),
  projectId: z.string(),
  email: z.string().email(),
  token: z.string(),
  role: z.enum(["owner", "editor", "viewer"]),
  status: z.enum(["pending", "accepted", "expired", "cancelled"]),
  expiresAt: z.string().datetime(),
  createdAt: z.string().datetime().optional(),
  acceptedAt: z.string().datetime().optional(),
});

export type Invitation = z.infer<typeof InvitationSchema>;

export const CreateInvitationRequestSchema = z.object({
  project_id: z.string().min(1),
  email: z.string().email().max(255),
  role: z.enum(["owner", "editor", "viewer"]),
});

export type CreateInvitationRequest = z.infer<
  typeof CreateInvitationRequestSchema
>;

export const AcceptInvitationRequestSchema = z.object({
  token: z.string().min(1).max(255),
});

export type AcceptInvitationRequest = z.infer<
  typeof AcceptInvitationRequestSchema
>;

export const UpdateInvitationStatusRequestSchema = z.object({
  status: z.enum(["pending", "accepted", "expired", "cancelled"]),
});

export type UpdateInvitationStatusRequest = z.infer<
  typeof UpdateInvitationStatusRequestSchema
>;

// ============================================================================
// SNAPSHOT SCHEMAS
// ============================================================================

export const SnapshotSchema = z.object({
  id: z.string(),
  projectId: z.string(),
  name: z.string(),
  type: z.string(),
  elements: z.unknown(),
  timestamp: z.number(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
});

export type Snapshot = z.infer<typeof SnapshotSchema>;

export const SaveSnapshotRequestSchema = z.object({
  name: z.string().min(1),
  type: z.string().optional(),
  elements: z.unknown(),
});

export type SaveSnapshotRequest = z.infer<typeof SaveSnapshotRequestSchema>;

// ============================================================================
// CONTENT TYPE SCHEMAS
// ============================================================================

export const ContentTypeSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
});

export type ContentType = z.infer<typeof ContentTypeSchema>;

export const CreateContentTypeRequestSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
});

export type CreateContentTypeRequest = z.infer<
  typeof CreateContentTypeRequestSchema
>;

export const UpdateContentTypeRequestSchema = z.object({
  name: z.string().min(1).optional(),
  description: z.string().optional(),
});

export type UpdateContentTypeRequest = z.infer<
  typeof UpdateContentTypeRequestSchema
>;

// ============================================================================
// CONTENT FIELD SCHEMAS
// ============================================================================

export const ContentFieldSchema = z.object({
  id: z.string(),
  contentTypeId: z.string(),
  name: z.string(),
  type: z.string(),
  required: z.boolean(),
});

export type ContentField = z.infer<typeof ContentFieldSchema>;

export const CreateContentFieldRequestSchema = z.object({
  contentTypeId: z.string().min(1),
  name: z.string().min(1),
  type: z.string().min(1),
  required: z.boolean().optional(),
});

export type CreateContentFieldRequest = z.infer<
  typeof CreateContentFieldRequestSchema
>;

export const UpdateContentFieldRequestSchema = z.object({
  name: z.string().min(1).optional(),
  type: z.string().min(1).optional(),
  required: z.boolean().optional(),
});

export type UpdateContentFieldRequest = z.infer<
  typeof UpdateContentFieldRequestSchema
>;

// ============================================================================
// CONTENT ITEM SCHEMAS
// ============================================================================

export const ContentFieldValueSchema = z.object({
  fieldId: z.string(),
  value: z.unknown(),
});

export type ContentFieldValue = z.infer<typeof ContentFieldValueSchema>;

export const ContentItemSchema = z.object({
  id: z.string(),
  contentTypeId: z.string(),
  slug: z.string(),
  title: z.string(),
  published: z.boolean(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
  fieldValues: z.array(ContentFieldValueSchema).optional(),
});

export type ContentItem = z.infer<typeof ContentItemSchema>;

export const CreateContentItemRequestSchema = z.object({
  contentTypeId: z.string().min(1),
  slug: z.string().min(1),
  title: z.string().min(1),
  published: z.boolean().optional(),
  fieldValues: z.array(ContentFieldValueSchema).optional(),
});

export type CreateContentItemRequest = z.infer<
  typeof CreateContentItemRequestSchema
>;

export const UpdateContentItemRequestSchema = z.object({
  slug: z.string().min(1).optional(),
  title: z.string().min(1).optional(),
  published: z.boolean().optional(),
  fieldValues: z.array(ContentFieldValueSchema).optional(),
});

export type UpdateContentItemRequest = z.infer<
  typeof UpdateContentItemRequestSchema
>;

// ============================================================================
// IMAGE SCHEMAS
// ============================================================================

export const UploadBase64ImageRequestSchema = z.object({
  imageData: z.string(),
  fileName: z.string().optional(),
});

export type UploadBase64ImageRequest = z.infer<
  typeof UploadBase64ImageRequestSchema
>;
