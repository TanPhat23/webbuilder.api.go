import { APIRequestContext } from "@playwright/test";
import * as schemas from "./schemas";

export class APIClient {
  constructor(
    private request: APIRequestContext,
    private baseURL: string = "http://localhost:8080/api/v1",
    private authToken?: string,
  ) {}

  private getHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };

    if (this.authToken) {
      headers["Authorization"] = `Bearer ${this.authToken}`;
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

  // ============================================================================
  // USER ENDPOINTS
  // ============================================================================

  async searchUsers(query: string): Promise<schemas.User[]> {
    return this.get(`/users/search?q=${encodeURIComponent(query)}`);
  }

  async getUserByEmail(email: string): Promise<schemas.User> {
    return this.get(`/users/email/${encodeURIComponent(email)}`);
  }

  async getUserByUsername(username: string): Promise<schemas.User> {
    return this.get(`/users/username/${encodeURIComponent(username)}`);
  }

  // ============================================================================
  // PROJECT ENDPOINTS
  // ============================================================================

  async getProjectsByUser(): Promise<schemas.Project[]> {
    return this.get("/projects/user");
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

  // ============================================================================
  // PAGE ENDPOINTS
  // ============================================================================

  async getPagesByProjectID(projectId: string): Promise<schemas.Page[]> {
    return this.get(`/pages/${projectId}`);
  }

  async getPageByID(projectId: string, pageId: string): Promise<schemas.Page> {
    return this.get(`/pages/${projectId}/${pageId}`);
  }

  async getPublicPagesByProjectID(projectId: string): Promise<schemas.Page[]> {
    return this.get(`/pages/public/${projectId}`);
  }

  async getPublicPageByID(
    projectId: string,
    pageId: string,
  ): Promise<schemas.Page> {
    return this.get(`/pages/public/${projectId}/${pageId}`);
  }

  async createPage(
    projectId: string,
    data: schemas.CreatePageRequest,
  ): Promise<schemas.Page> {
    return this.post(`/pages/${projectId}`, data);
  }

  async updatePage(
    projectId: string,
    pageId: string,
    data: schemas.UpdatePageRequest,
  ) {
    return this.patch<schemas.Page>(`/pages/${projectId}/${pageId}`, data);
  }

  async deletePage(projectId: string, pageId: string) {
    return this.delete(`/projects/${projectId}/pages/${pageId}`);
  }

  // ============================================================================
  // ELEMENT ENDPOINTS
  // ============================================================================

  async getElementsByProjectID(projectId: string): Promise<schemas.Element[]> {
    return this.get(`/elements/${projectId}`);
  }

  async getPublicElementsByProjectID(
    projectId: string,
  ): Promise<schemas.Element[]> {
    return this.get(`/elements/public/${projectId}`);
  }

  async getElementsByPageIds(pageIds: string[]): Promise<schemas.Element[]> {
    const pageIdsParam = pageIds.join(",");
    return this.get(
      `/elements/by-pages?pageIds=${encodeURIComponent(pageIdsParam)}`,
    );
  }

  async getPublicElementsByPageIds(
    pageIds: string[],
  ): Promise<schemas.Element[]> {
    const pageIdsParam = pageIds.join(",");
    return this.get(
      `/elements/public/by-pages?pageIds=${encodeURIComponent(pageIdsParam)}`,
    );
  }

  // ============================================================================
  // IMAGE ENDPOINTS
  // ============================================================================
  async getImages(): Promise<schemas.Image[]> {
    return this.get("/images");
  }

  async getImageByID(imageId: string): Promise<schemas.Image> {
    return this.get(`/images/${imageId}`);
  }

  async deleteImage(imageId: string) {
    return this.delete(`/images/${imageId}`);
  }

  async uploadBase64Image(
    data: schemas.UploadBase64ImageRequest,
  ): Promise<schemas.Image> {
    return this.post("/images", data);
  }

  // ============================================================================
  // COLLABORATOR ENDPOINTS
  // ============================================================================

  async getCollaboratorsByProjectID(
    projectId: string,
  ): Promise<schemas.Collaborator[]> {
    return this.get(`/collaborators/:projectid`).then(
      (res: any) => res.collaborators || res,
    );
  }

  async getCollaboratorByID(
    collaboratorId: string,
  ): Promise<schemas.Collaborator> {
    return this.get(`/collaborators/${collaboratorId}`);
  }

  async createCollaborator(
    data: schemas.CreateCollaboratorRequest,
  ): Promise<schemas.Collaborator> {
    return this.post("/collaborators", data);
  }

  async updateCollaboratorRole(
    collaboratorId: string,
    role: string,
  ): Promise<schemas.Collaborator> {
    return this.patch(`/collaborators/${collaboratorId}/role`, { role });
  }

  async deleteCollaborator(collaboratorId: string) {
    return this.delete(`/collaborators/${collaboratorId}`);
  }

  // ============================================================================
  // COMMENT ENDPOINTS (Marketplace)
  // ============================================================================

  async getComments(
    itemId?: string,
    limit?: number,
    offset?: number,
  ): Promise<schemas.Comment[]> {
    let path = "/comments";
    const params = new URLSearchParams();
    if (itemId) params.append("itemId", itemId);
    if (limit) params.append("limit", limit.toString());
    if (offset) params.append("offset", offset.toString());
    if (params.size > 0) path += `?${params.toString()}`;
    return this.get(path);
  }

  async getCommentByID(commentId: string): Promise<schemas.Comment> {
    return this.get(`/comments/${commentId}`);
  }

  async createComment(
    data: schemas.CreateCommentRequest,
  ): Promise<schemas.Comment> {
    return this.post("/comments", data);
  }

  async updateComment(
    commentId: string,
    data: schemas.UpdateCommentRequest,
  ): Promise<schemas.Comment> {
    return this.patch(`/comments/${commentId}`, data);
  }

  async deleteComment(commentId: string) {
    return this.delete(`/comments/${commentId}`);
  }

  async createCommentReaction(
    commentId: string,
    type: string,
  ): Promise<schemas.CommentReaction> {
    return this.post(`/comments/${commentId}/reactions`, { type });
  }

  // ============================================================================
  // CUSTOM ELEMENT ENDPOINTS
  // ============================================================================

  async getCustomElements(): Promise<schemas.CustomElement[]> {
    return this.get("/custom-elements");
  }

  async getCustomElementByID(
    elementId: string,
  ): Promise<schemas.CustomElement> {
    return this.get(`/custom-elements/${elementId}`);
  }

  async createCustomElement(
    data: schemas.CreateCustomElementRequest,
  ): Promise<schemas.CustomElement> {
    return this.post("/custom-elements", data);
  }

  async updateCustomElement(
    elementId: string,
    data: schemas.UpdateCustomElementRequest,
  ): Promise<schemas.CustomElement> {
    return this.patch(`/custom-elements/${elementId}`, data);
  }

  async deleteCustomElement(elementId: string) {
    return this.delete(`/custom-elements/${elementId}`);
  }

  async duplicateCustomElement(
    elementId: string,
    newName: string,
  ): Promise<schemas.CustomElement> {
    return this.post(`/custom-elements/${elementId}/duplicate`, {
      name: newName,
    });
  }

  // ============================================================================
  // CUSTOM ELEMENT TYPE ENDPOINTS
  // ============================================================================

  async getCustomElementTypes(): Promise<schemas.CustomElementType[]> {
    return this.get("/custom-element-types");
  }

  async getCustomElementTypeByID(
    typeId: string,
  ): Promise<schemas.CustomElementType> {
    return this.get(`/custom-element-types/${typeId}`);
  }

  async createCustomElementType(
    data: schemas.CreateCustomElementTypeRequest,
  ): Promise<schemas.CustomElementType> {
    return this.post("/custom-element-types", data);
  }

  async updateCustomElementType(
    typeId: string,
    data: schemas.UpdateCustomElementTypeRequest,
  ): Promise<schemas.CustomElementType> {
    return this.patch(`/custom-element-types/${typeId}`, data);
  }

  async deleteCustomElementType(typeId: string) {
    return this.delete(`/custom-element-types/${typeId}`);
  }

  // ============================================================================
  // ELEMENT COMMENT ENDPOINTS
  // ============================================================================

  async getElementComments(
    limit?: number,
    offset?: number,
  ): Promise<schemas.ElementComment[]> {
    let path = "/element-comments";
    const params = new URLSearchParams();
    if (limit) params.append("limit", limit.toString());
    if (offset) params.append("offset", offset.toString());
    if (params.size > 0) path += `?${params.toString()}`;
    return this.get(path);
  }

  async getElementCommentByID(
    commentId: string,
  ): Promise<schemas.ElementComment> {
    return this.get(`/element-comments/${commentId}`);
  }

  async getElementCommentsByElement(
    elementId: string,
  ): Promise<schemas.ElementComment[]> {
    return this.get(`/elements/${elementId}/comments`);
  }

  async getCommentsByAuthorID(
    authorId: string,
  ): Promise<schemas.ElementComment[]> {
    return this.get(`/element-comments/author/${authorId}`);
  }

  async getCommentsByProjectID(
    projectId: string,
  ): Promise<schemas.ElementComment[]> {
    return this.get(`/projects/${projectId}/comments`);
  }

  async createElementComment(
    data: schemas.CreateElementCommentRequest,
  ): Promise<schemas.ElementComment> {
    return this.post("/element-comments", data);
  }

  async updateElementComment(
    commentId: string,
    data: schemas.UpdateElementCommentRequest,
  ): Promise<schemas.ElementComment> {
    return this.patch(`/element-comments/${commentId}`, data);
  }

  async toggleResolvedStatus(
    commentId: string,
  ): Promise<schemas.ElementComment> {
    return this.patch(`/element-comments/${commentId}/toggle-resolved`, {});
  }

  async deleteElementComment(commentId: string) {
    return this.delete(`/element-comments/${commentId}`);
  }

  // ============================================================================
  // EVENT WORKFLOW ENDPOINTS
  // ============================================================================

  async getEventWorkflowsByProject(
    projectId: string,
  ): Promise<schemas.EventWorkflow[]> {
    return this.get(`/event-workflows/${projectId}`);
  }

  async getEventWorkflowByID(
    workflowId: string,
  ): Promise<schemas.EventWorkflow> {
    return this.get(`/event-workflows/${workflowId}`);
  }

  async getEventWorkflowElements(workflowId: string): Promise<unknown[]> {
    return this.get(`/event-workflows/${workflowId}/elements`);
  }

  async createEventWorkflow(
    data: schemas.CreateEventWorkflowRequest,
  ): Promise<schemas.EventWorkflow> {
    return this.post("/event-workflows", data);
  }

  async updateEventWorkflow(
    workflowId: string,
    data: schemas.UpdateEventWorkflowRequest,
  ): Promise<schemas.EventWorkflow> {
    return this.patch(`/event-workflows/${workflowId}`, data);
  }

  async updateEventWorkflowEnabled(workflowId: string, enabled: boolean) {
    return this.patch(`/event-workflows/${workflowId}/enabled`, { enabled });
  }

  async deleteEventWorkflow(workflowId: string) {
    return this.delete(`/event-workflows/${workflowId}`);
  }

  // ============================================================================
  // ELEMENT EVENT WORKFLOW ENDPOINTS
  // ============================================================================

  async getElementEventWorkflows(): Promise<schemas.ElementEventWorkflow[]> {
    return this.get("/element-event-workflows");
  }

  async getElementEventWorkflowByID(
    workflowId: string,
  ): Promise<schemas.ElementEventWorkflow> {
    return this.get(`/element-event-workflows/${workflowId}`);
  }

  async createElementEventWorkflow(
    data: schemas.CreateElementEventWorkflowRequest,
  ): Promise<schemas.ElementEventWorkflow> {
    return this.post("/element-event-workflows", data);
  }

  async updateElementEventWorkflow(
    workflowId: string,
    data: schemas.UpdateElementEventWorkflowRequest,
  ): Promise<schemas.ElementEventWorkflow> {
    return this.patch(`/element-event-workflows/${workflowId}`, data);
  }

  async deleteElementEventWorkflow(workflowId: string) {
    return this.delete(`/element-event-workflows/${workflowId}`);
  }

  async deleteElementEventWorkflowsByElement(elementId: string) {
    return this.delete(`/element-event-workflows/element/${elementId}`);
  }

  async deleteElementEventWorkflowsByWorkflow(workflowId: string) {
    return this.delete(`/element-event-workflows/workflow/${workflowId}`);
  }

  // ============================================================================
  // INVITATION ENDPOINTS
  // ============================================================================

  async getInvitations(): Promise<schemas.Invitation[]> {
    return this.get("/invitations");
  }

  async getInvitationsByProject(
    projectId: string,
  ): Promise<schemas.Invitation[]> {
    return this.get(`/invitations/project/${projectId}`);
  }

  async getPendingInvitationsByProject(
    projectId: string,
  ): Promise<schemas.Invitation[]> {
    return this.get(`/invitations/project/${projectId}/pending`);
  }

  async createInvitation(
    data: schemas.CreateInvitationRequest,
  ): Promise<schemas.Invitation> {
    return this.post("/invitations", data);
  }

  async acceptInvitation(data: schemas.AcceptInvitationRequest) {
    return this.post("/invitations/accept", data);
  }

  async acceptInvitationByToken(
    token: string,
    data: schemas.AcceptInvitationRequest,
  ) {
    return this.post(`/invitations/${token}/accept`, data);
  }

  async updateInvitationStatus(
    invitationId: string,
    status: string,
  ): Promise<schemas.Invitation> {
    return this.patch(`/invitations/${invitationId}/status`, { status });
  }

  async cancelInvitation(invitationId: string): Promise<schemas.Invitation> {
    return this.patch(`/invitations/${invitationId}/cancel`, {});
  }

  async deleteInvitation(invitationId: string) {
    return this.delete(`/invitations/${invitationId}`);
  }

  // ============================================================================
  // MARKETPLACE ENDPOINTS
  // ============================================================================

  async getMarketplaceItems(): Promise<schemas.MarketplaceItem[]> {
    return this.get("/marketplace");
  }

  async getMarketplaceItemByID(
    itemId: string,
  ): Promise<schemas.MarketplaceItem> {
    return this.get(`/marketplace/${itemId}`);
  }

  async createMarketplaceItem(
    data: schemas.CreateMarketplaceItemRequest,
  ): Promise<schemas.MarketplaceItem> {
    return this.post("/marketplace", data);
  }

  async updateMarketplaceItem(
    itemId: string,
    data: schemas.UpdateMarketplaceItemRequest,
  ): Promise<schemas.MarketplaceItem> {
    return this.patch(`/marketplace/${itemId}`, data);
  }

  async deleteMarketplaceItem(itemId: string) {
    return this.delete(`/marketplace/${itemId}`);
  }

  async createMarketplaceTag(
    data: schemas.CreateTagRequest,
  ): Promise<schemas.Tag> {
    return this.post("/marketplace/tags", data);
  }

  async createMarketplaceCategory(
    data: schemas.CreateCategoryRequest,
  ): Promise<schemas.Category> {
    return this.post("/marketplace/categories", data);
  }

  // ============================================================================
  // SNAPSHOT ENDPOINTS
  // ============================================================================

  async getSnapshotsByProjectID(
    projectId: string,
  ): Promise<schemas.Snapshot[]> {
    return this.get(`/snapshots/${projectId}`);
  }

  async getSnapshotByID(snapshotId: string): Promise<schemas.Snapshot> {
    return this.get(`/snapshots/${snapshotId}`);
  }

  async saveSnapshot(
    projectId: string,
    data: schemas.SaveSnapshotRequest,
  ): Promise<schemas.Snapshot> {
    return this.post(`/snapshots/${projectId}/save`, data);
  }

  async deleteSnapshot(snapshotId: string) {
    return this.delete(`/snapshots/${snapshotId}`);
  }

  // ============================================================================
  // CONTENT TYPE ENDPOINTS
  // ============================================================================

  async getContentTypes(): Promise<schemas.ContentType[]> {
    return this.get("/content-types");
  }

  async getContentTypeByID(typeId: string): Promise<schemas.ContentType> {
    return this.get(`/content-types/${typeId}`);
  }

  async createContentType(
    data: schemas.CreateContentTypeRequest,
  ): Promise<schemas.ContentType> {
    return this.post("/content-types", data);
  }

  async updateContentType(
    typeId: string,
    data: schemas.UpdateContentTypeRequest,
  ): Promise<schemas.ContentType> {
    return this.patch(`/content-types/${typeId}`, data);
  }

  async deleteContentType(typeId: string) {
    return this.delete(`/content-types/${typeId}`);
  }

  // ============================================================================
  // CONTENT FIELD ENDPOINTS
  // ============================================================================

  async getContentFieldsByType(
    contentTypeId: string,
  ): Promise<schemas.ContentField[]> {
    return this.get(`/content-fields/${contentTypeId}`);
  }

  async getContentFieldByID(fieldId: string): Promise<schemas.ContentField> {
    return this.get(`/content-fields/by-id/${fieldId}`);
  }

  async createContentField(
    data: schemas.CreateContentFieldRequest,
  ): Promise<schemas.ContentField> {
    return this.post("/content-fields", data);
  }

  async updateContentField(
    fieldId: string,
    data: schemas.UpdateContentFieldRequest,
  ): Promise<schemas.ContentField> {
    return this.patch(`/content-fields/${fieldId}`, data);
  }

  async deleteContentField(fieldId: string) {
    return this.delete(`/content-fields/${fieldId}`);
  }

  // ============================================================================
  // CONTENT ITEM ENDPOINTS
  // ============================================================================

  async getContentItems(): Promise<schemas.ContentItem[]> {
    return this.get(`/content-items`);
  }

  async getContentItemByID(itemId: string): Promise<schemas.ContentItem> {
    return this.get(`/content-items/by-id/${itemId}`);
  }

  async getPublicContentItems(): Promise<schemas.ContentItem[]> {
    return this.get("/public/content");
  }

  async getPublicContentItemBySlug(
    contentTypeId: string,
    slug: string,
  ): Promise<schemas.ContentItem> {
    return this.get(`/public/content/${contentTypeId}/${slug}`);
  }

  async createContentItem(
    data: schemas.CreateContentItemRequest,
  ): Promise<schemas.ContentItem> {
    return this.post("/content-items", data);
  }

  async updateContentItem(
    itemId: string,
    data: schemas.UpdateContentItemRequest,
  ): Promise<schemas.ContentItem> {
    return this.patch(`/content-items/${itemId}`, data);
  }

  async deleteContentItem(itemId: string) {
    return this.delete(`/content-items/${itemId}`);
  }
}
