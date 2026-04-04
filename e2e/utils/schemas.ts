import { z } from 'zod';

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
