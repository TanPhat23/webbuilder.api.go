import { APIRequestContext } from '@playwright/test';
import * as schemas from './schemas';

export class APIClient {
  constructor(
    private request: APIRequestContext,
    private baseURL: string = 'http://localhost:8080/api/v1',
    private authToken?: string,
  ) {}

  private getHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (this.authToken) {
      headers['Authorization'] = `Bearer ${this.authToken}`;
    }

    return headers;
  }

  setAuthToken(token: string) {
    this.authToken = token;
  }

  async get<T>(path: string) {
    const response = await this.request.get(`${this.baseURL}${path}`, {
      headers: this.getHeaders(),
    });

    if (!response.ok()) {
      throw new Error(`GET ${path} failed: ${response.status()}`);
    }

    return response.json() as Promise<T>;
  }

  async post<T>(path: string, data?: unknown) {
    const response = await this.request.post(`${this.baseURL}${path}`, {
      headers: this.getHeaders(),
      data,
    });

    if (!response.ok()) {
      throw new Error(`POST ${path} failed: ${response.status()}`);
    }

    return response.json() as Promise<T>;
  }

  async patch<T>(path: string, data?: unknown) {
    const response = await this.request.patch(`${this.baseURL}${path}`, {
      headers: this.getHeaders(),
      data,
    });

    if (!response.ok()) {
      throw new Error(`PATCH ${path} failed: ${response.status()}`);
    }

    return response.json() as Promise<T>;
  }

  async delete(path: string) {
    const response = await this.request.delete(`${this.baseURL}${path}`, {
      headers: this.getHeaders(),
    });

    if (!response.ok() && response.status() !== 204) {
      throw new Error(`DELETE ${path} failed: ${response.status()}`);
    }

    return response.status() === 204 ? null : response.json();
  }

  async searchUsers(query: string): Promise<schemas.User[]> {
    return this.get(`/users/search?q=${encodeURIComponent(query)}`);
  }

  async getUserByEmail(email: string): Promise<schemas.User> {
    return this.get(`/users/email/${encodeURIComponent(email)}`);
  }

  async getUserByUsername(username: string): Promise<schemas.User> {
    return this.get(`/users/username/${encodeURIComponent(username)}`);
  }

  async getProjectsByUser(): Promise<schemas.Project[]> {
    return this.get('/projects/user');
  }

  async getProjectByID(projectId: string): Promise<schemas.Project> {
    return this.get(`/projects/${projectId}`);
  }

  async getPublicProjectByID(projectId: string): Promise<schemas.Project> {
    return this.get(`/projects/public/${projectId}`);
  }

  async getProjectPages(projectId: string): Promise<schemas.Page[]> {
    return this.get(`/projects/${projectId}/pages`);
  }

  async deleteProject(projectId: string) {
    return this.delete(`/projects/${projectId}`);
  }

  async updateProject(projectId: string, data: schemas.UpdateProjectRequest) {
    return this.patch<schemas.Project>(`/projects/${projectId}`, data);
  }

  async getPagesByProjectID(projectId: string): Promise<schemas.Page[]> {
    return this.get(`/pages/${projectId}`);
  }

  async getPageByID(projectId: string, pageId: string): Promise<schemas.Page> {
    return this.get(`/pages/${projectId}/${pageId}`);
  }

  async getPublicPagesByProjectID(projectId: string): Promise<schemas.Page[]> {
    return this.get(`/pages/public/${projectId}`);
  }

  async getPublicPageByID(projectId: string, pageId: string): Promise<schemas.Page> {
    return this.get(`/pages/public/${projectId}/${pageId}`);
  }

  async createPage(projectId: string, data: schemas.CreatePageRequest): Promise<schemas.Page> {
    return this.post(`/pages/${projectId}`, data);
  }

  async updatePage(projectId: string, pageId: string, data: schemas.UpdatePageRequest) {
    return this.patch<schemas.Page>(`/pages/${projectId}/${pageId}`, data);
  }

  async deletePage(projectId: string, pageId: string) {
    return this.delete(`/projects/${projectId}/pages/${pageId}`);
  }

  async getImages(): Promise<schemas.Image[]> {
    return this.get('/images');
  }

  async getImageByID(imageId: string): Promise<schemas.Image> {
    return this.get(`/images/${imageId}`);
  }

  async deleteImage(imageId: string) {
    return this.delete(`/images/${imageId}`);
  }
}
