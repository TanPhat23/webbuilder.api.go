import { APIClient } from './api-client';
import * as schemas from './schemas';

export class TestHelpers {
  constructor(private apiClient: APIClient) {}

  async createTestProject(overrides?: Partial<schemas.Project>): Promise<schemas.Project> {
    const projects = await this.apiClient.getProjectsByUser();
    if (projects.length > 0) {
      return projects[0];
    }
    throw new Error('No projects available for testing');
  }

  async findOrCreatePage(
    projectId: string,
    pageName?: string,
  ): Promise<schemas.Page> {
    const pages = await this.apiClient.getPagesByProjectID(projectId);
    
    if (pages.length > 0) {
      return pages[0];
    }

    const newPage = await this.apiClient.createPage(projectId, {
      name: pageName || `Test Page ${Date.now()}`,
      type: 'landing',
    });

    return newPage;
  }

  async getFirstImage(): Promise<schemas.Image | null> {
    const images = await this.apiClient.getImages();
    return images.length > 0 ? images[0] : null;
  }

  async cleanupPage(projectId: string, pageId: string): Promise<void> {
    try {
      await this.apiClient.deletePage(projectId, pageId);
    } catch (error) {
      console.log('Cleanup: Failed to delete page', error);
    }
  }

  async cleanupProject(projectId: string): Promise<void> {
    try {
      await this.apiClient.deleteProject(projectId);
    } catch (error) {
      console.log('Cleanup: Failed to delete project', error);
    }
  }

  async cleanupImage(imageId: string): Promise<void> {
    try {
      await this.apiClient.deleteImage(imageId);
    } catch (error) {
      console.log('Cleanup: Failed to delete image', error);
    }
  }

  async waitFor<T>(
    fn: () => Promise<T>,
    timeoutMs: number = 5000,
    intervalMs: number = 500,
  ): Promise<T> {
    const startTime = Date.now();

    while (Date.now() - startTime < timeoutMs) {
      try {
        return await fn();
      } catch (error) {
        await new Promise(resolve => setTimeout(resolve, intervalMs));
      }
    }

    throw new Error(`Timeout waiting for condition after ${timeoutMs}ms`);
  }

  validateProjectStructure(project: schemas.Project): boolean {
    return (
      !!project.id &&
      !!project.name &&
      typeof project.published === 'boolean' &&
      !!project.ownerId
    );
  }

  validatePageStructure(page: schemas.Page): boolean {
    return (
      !!page.Id &&
      !!page.Name &&
      !!page.Type &&
      !!page.ProjectId
    );
  }

  validateImageStructure(image: schemas.Image): boolean {
    return (
      !!image.imageId &&
      !!image.imageLink &&
      !!image.userId
    );
  }

  validateUserStructure(user: schemas.User): boolean {
    return (
      !!user.id &&
      !!user.email
    );
  }
}
