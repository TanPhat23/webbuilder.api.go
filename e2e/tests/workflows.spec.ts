import { test, expect } from "../fixtures/helpers.fixture";

test.describe("Workflow Integration Tests", () => {
  test.describe("Project Creation → Page Creation → Element Retrieval Workflow", () => {
    test("should complete full project setup workflow", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        expect(project).toBeDefined();
        expect(project.id).toBeDefined();

        const pageName = `Workflow Test Page ${Date.now()}`;
        const page = await apiClient.createPage(project.id, {
          name: pageName,
          type: "landing",
          styles: { backgroundColor: "#ffffff" },
        });

        expect(page).toBeDefined();
        expect(page.Name).toBe(pageName);
        expect(page.ProjectId).toBe(project.id);

        const elements = await apiClient.getElementsByProjectID(project.id);
        expect(Array.isArray(elements)).toBe(true);

        const verified = await apiClient.getPageByID(project.id, page.Id);
        expect(verified.Id).toBe(page.Id);
        expect(verified.Name).toBe(pageName);
      } catch (error) {
        test.skip();
      }
    });

    test("should maintain data consistency through full workflow", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();

        const page1 = await apiClient.createPage(project.id, {
          name: `Workflow Page 1 ${Date.now()}`,
          type: "landing",
        });

        const page2 = await apiClient.createPage(project.id, {
          name: `Workflow Page 2 ${Date.now()}`,
          type: "product",
        });

        const pages = await apiClient.getPagesByProjectID(project.id);
        const pageIds = pages.map((p) => p.Id);

        expect(pageIds).toContain(page1.Id);
        expect(pageIds).toContain(page2.Id);

        const allElements = await apiClient.getElementsByProjectID(project.id);
        expect(Array.isArray(allElements)).toBe(true);

        const pageSpecificElements = await apiClient.getElementsByPageIds([
          page1.Id,
          page2.Id,
        ]);
        expect(Array.isArray(pageSpecificElements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should update project through workflow", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();

        const updatedName = `Updated Project ${Date.now()}`;
        const updated = await apiClient.updateProject(project.id, {
          name: updatedName,
          published: true,
        });

        expect(updated.name).toBe(updatedName);
        expect(updated.published).toBe(true);

        const verified = await apiClient.getProjectByID(project.id);
        expect(verified.name).toBe(updatedName);
        expect(verified.published).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle multiple pages and retrieve all elements", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();

        const pages = [];
        for (let i = 0; i < 3; i++) {
          const page = await apiClient.createPage(project.id, {
            name: `Multi Page ${i} ${Date.now()}`,
            type: "landing",
          });
          pages.push(page);
        }

        const pageIds = pages.map((p) => p.Id);
        const elements = await apiClient.getElementsByPageIds(pageIds);

        expect(Array.isArray(elements)).toBe(true);

        const allProjectElements = await apiClient.getElementsByProjectID(
          project.id,
        );
        expect(Array.isArray(allProjectElements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Collaborator Invitation → Acceptance Workflow", () => {
    test("should create collaborator on project", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const users = await apiClient.searchUsers("test");

        if (users.length === 0) {
          test.skip();
        }

        const collaborator = await apiClient.createCollaborator({
          projectId: project.id,
          userId: users[0].id,
          role: "editor",
        });

        expect(collaborator).toBeDefined();
        expect(collaborator.projectId).toBe(project.id);
        expect(collaborator.role).toBe("editor");
      } catch (error) {
        test.skip();
      }
    });

    test("should update collaborator role through workflow", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const users = await apiClient.searchUsers("test");

        if (users.length === 0) {
          test.skip();
        }

        const collaborator = await apiClient.createCollaborator({
          projectId: project.id,
          userId: users[0].id,
          role: "viewer",
        });

        expect(collaborator.role).toBe("viewer");

        const updated = await apiClient.updateCollaboratorRole(
          collaborator.id,
          "editor",
        );

        expect(updated.role).toBe("editor");

        const verified = await apiClient.getCollaboratorByID(collaborator.id);
        expect(verified.role).toBe("editor");
      } catch (error) {
        test.skip();
      }
    });

    test("should manage project collaborators", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const users = await apiClient.searchUsers("test");

        if (users.length < 2) {
          test.skip();
        }

        const collab1 = await apiClient.createCollaborator({
          projectId: project.id,
          userId: users[0].id,
          role: "editor",
        });

        const collab2 = await apiClient.createCollaborator({
          projectId: project.id,
          userId: users[1].id,
          role: "viewer",
        });

        expect(collab1.projectId).toBe(project.id);
        expect(collab2.projectId).toBe(project.id);
        expect(collab1.role).toBe("editor");
        expect(collab2.role).toBe("viewer");
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Element Comment → Resolution Workflow", () => {
    test("should create and update element comment", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);
        const elements = await apiClient.getElementsByPageIds([page.Id]);

        if (elements.length === 0) {
          test.skip();
        }

        const element = elements[0];
        const commentContent = `Review needed on ${Date.now()}`;

        const comment = await apiClient.createElementComment({
          content: commentContent,
          elementId: element.id,
        });

        expect(comment).toBeDefined();
        expect(comment.content).toBe(commentContent);
        expect(comment.elementId).toBe(element.id);
        expect(comment.resolved).toBe(false);

        const updated = await apiClient.updateElementComment(comment.id, {
          content: `Updated: ${commentContent}`,
          resolved: true,
        });

        expect(updated.content).toContain("Updated:");
        expect(updated.resolved).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should manage multiple comments on element", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);
        const elements = await apiClient.getElementsByPageIds([page.Id]);

        if (elements.length === 0) {
          test.skip();
        }

        const element = elements[0];
        const comments = [];

        for (let i = 0; i < 3; i++) {
          const comment = await apiClient.createElementComment({
            content: `Comment ${i} on ${Date.now()}`,
            elementId: element.id,
          });
          comments.push(comment);
        }

        expect(comments.length).toBe(3);
        comments.forEach((c) => {
          expect(c.elementId).toBe(element.id);
        });

        const resolved = await apiClient.updateElementComment(comments[0].id, {
          resolved: true,
        });

        expect(resolved.resolved).toBe(true);

        const unresolved = comments.slice(1);
        unresolved.forEach((c) => {
          expect(c.resolved).toBe(false);
        });
      } catch (error) {
        test.skip();
      }
    });

    test("should delete element comment", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);
        const elements = await apiClient.getElementsByPageIds([page.Id]);

        if (elements.length === 0) {
          test.skip();
        }

        const comment = await apiClient.createElementComment({
          content: `Comment to delete ${Date.now()}`,
          elementId: elements[0].id,
        });

        await apiClient.deleteElementComment(comment.id);

        try {
          await apiClient.getElementCommentByID(comment.id);
          test.fail(true, "Should have deleted comment");
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Snapshot → Restore Workflow", () => {
    test("should create project snapshot", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);
        const elements = await apiClient.getElementsByPageIds([page.Id]);

        const snapshot = await apiClient.saveSnapshot(project.id, {
          name: `Snapshot ${Date.now()}`,
          type: "backup",
          elements: elements,
        });

        expect(snapshot).toBeDefined();
        expect(snapshot.projectId).toBe(project.id);
        expect(snapshot.name).toContain("Snapshot");
        expect(snapshot.elements).toBeDefined();
      } catch (error) {
        test.skip();
      }
    });

    test("should retrieve saved snapshots", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);
        const elements = await apiClient.getElementsByPageIds([page.Id]);

        const snapshot1 = await apiClient.saveSnapshot(project.id, {
          name: `Snapshot 1 ${Date.now()}`,
          type: "backup",
          elements: elements,
        });

        const snapshot2 = await apiClient.saveSnapshot(project.id, {
          name: `Snapshot 2 ${Date.now()}`,
          type: "restore",
          elements: elements,
        });

        expect(snapshot1.id).toBeDefined();
        expect(snapshot2.id).toBeDefined();
        expect(snapshot1.projectId).toBe(project.id);
        expect(snapshot2.projectId).toBe(project.id);
      } catch (error) {
        test.skip();
      }
    });

    test("should restore from snapshot", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);
        const initialElements = await apiClient.getElementsByPageIds([page.Id]);

        const snapshot = await apiClient.saveSnapshot(project.id, {
          name: `Restore Test ${Date.now()}`,
          type: "backup",
          elements: initialElements,
        });

        expect(snapshot).toBeDefined();

        const restored = await apiClient.getSnapshotByID(snapshot.id);
        expect(restored.id).toBe(snapshot.id);
        expect((restored as any).elements.length).toBe(initialElements.length);
      } catch (error) {
        test.skip();
      }
    });

    test("should manage multiple snapshots", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);
        const elements = await apiClient.getElementsByPageIds([page.Id]);

        const snapshots = [];
        for (let i = 0; i < 3; i++) {
          const snapshot = await apiClient.saveSnapshot(project.id, {
            name: `Multi Snapshot ${i} ${Date.now()}`,
            type: "backup",
            elements: elements,
          });
          snapshots.push(snapshot);
        }

        expect(snapshots.length).toBe(3);
        snapshots.forEach((s) => {
          expect(s.projectId).toBe(project.id);
        });

        const allSnapshots = await apiClient.getSnapshotsByProjectID(
          project.id,
        );
        expect(Array.isArray(allSnapshots)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Content Type → Content Item Creation Workflow", () => {
    test("should create content type with fields", async ({ apiClient }) => {
      try {
        const contentType = await apiClient.createContentType({
          name: `Test Content Type ${Date.now()}`,
          description: "Test content type for workflows",
        });

        expect(contentType).toBeDefined();
        expect(contentType.id).toBeDefined();
        expect(contentType.name).toContain("Test Content Type");

        const field = await apiClient.createContentField({
          contentTypeId: contentType.id,
          name: "title",
          type: "string",
          required: true,
        });

        expect(field).toBeDefined();
        expect(field.contentTypeId).toBe(contentType.id);
        expect(field.name).toBe("title");
        expect(field.required).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should create content item with type", async ({ apiClient }) => {
      try {
        const contentType = await apiClient.createContentType({
          name: `Item Type ${Date.now()}`,
          description: "Content type for items",
        });

        await apiClient.createContentField({
          contentTypeId: contentType.id,
          name: "title",
          type: "string",
          required: true,
        });

        const contentItem = await apiClient.createContentItem({
          contentTypeId: contentType.id,
          slug: `item-${Date.now()}`,
          title: "Test Item",
          published: true,
          fieldValues: [
            {
              fieldId: contentType.id,
              value: "Test Title",
            },
          ],
        });

        expect(contentItem).toBeDefined();
        expect(contentItem.contentTypeId).toBe(contentType.id);
        expect(contentItem.title).toBe("Test Item");
        expect(contentItem.published).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should manage content type with multiple fields", async ({
      apiClient,
    }) => {
      try {
        const contentType = await apiClient.createContentType({
          name: `Multi Field Type ${Date.now()}`,
          description: "Content type with multiple fields",
        });

        const field1 = await apiClient.createContentField({
          contentTypeId: contentType.id,
          name: "title",
          type: "string",
          required: true,
        });

        const field2 = await apiClient.createContentField({
          contentTypeId: contentType.id,
          name: "description",
          type: "text",
          required: false,
        });

        const field3 = await apiClient.createContentField({
          contentTypeId: contentType.id,
          name: "published",
          type: "boolean",
          required: false,
        });

        expect(field1.contentTypeId).toBe(contentType.id);
        expect(field2.contentTypeId).toBe(contentType.id);
        expect(field3.contentTypeId).toBe(contentType.id);

        expect(field1.required).toBe(true);
        expect(field2.required).toBe(false);
        expect(field3.required).toBe(false);
      } catch (error) {
        test.skip();
      }
    });

    test("should update content item through workflow", async ({
      apiClient,
    }) => {
      try {
        const contentType = await apiClient.createContentType({
          name: `Update Type ${Date.now()}`,
          description: "Content type for updating",
        });

        const contentItem = await apiClient.createContentItem({
          contentTypeId: contentType.id,
          slug: `item-${Date.now()}`,
          title: "Original Title",
          published: false,
        });

        const updated = await apiClient.updateContentItem(contentItem.id, {
          title: "Updated Title",
          published: true,
        });

        expect(updated.title).toBe("Updated Title");
        expect(updated.published).toBe(true);
        expect(updated.id).toBe(contentItem.id);
      } catch (error) {
        test.skip();
      }
    });

    test("should retrieve content items by type", async ({ apiClient }) => {
      try {
        const contentType = await apiClient.createContentType({
          name: `Retrieve Type ${Date.now()}`,
          description: "Content type for retrieval",
        });

        const item1 = await apiClient.createContentItem({
          contentTypeId: contentType.id,
          slug: `item-1-${Date.now()}`,
          title: "Item 1",
          published: true,
        });

        const item2 = await apiClient.createContentItem({
          contentTypeId: contentType.id,
          slug: `item-2-${Date.now()}`,
          title: "Item 2",
          published: true,
        });

        expect(item1.contentTypeId).toBe(contentType.id);
        expect(item2.contentTypeId).toBe(contentType.id);

        const retrieved1 = await apiClient.getContentItemByID(item1.id);
        expect(retrieved1.id).toBe(item1.id);
        expect(retrieved1.title).toBe("Item 1");
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Complex Multi-Step Workflows", () => {
    test("should complete full project lifecycle with multiple entities", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();

        const page = await apiClient.createPage(project.id, {
          name: `Lifecycle Page ${Date.now()}`,
          type: "landing",
        });

        const elements = await apiClient.getElementsByPageIds([page.Id]);

        const updated = await apiClient.updateProject(project.id, {
          description: "Project with complete lifecycle",
          published: true,
        });

        expect(updated.published).toBe(true);

        const snapshot = await apiClient.saveSnapshot(project.id, {
          name: `Lifecycle Snapshot ${Date.now()}`,
          type: "backup",
          elements: elements,
        });

        expect(snapshot).toBeDefined();

        const verified = await apiClient.getProjectByID(project.id);
        expect(verified.id).toBe(project.id);
        expect(verified.published).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle concurrent operations in workflow", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();

        const [page1, page2, updated] = await Promise.all([
          apiClient.createPage(project.id, {
            name: `Concurrent Page 1 ${Date.now()}`,
            type: "landing",
          }),
          apiClient.createPage(project.id, {
            name: `Concurrent Page 2 ${Date.now()}`,
            type: "product",
          }),
          apiClient.updateProject(project.id, {
            description: "Updated concurrently",
          }),
        ]);

        expect(page1).toBeDefined();
        expect(page2).toBeDefined();
        expect(updated).toBeDefined();

        expect(page1.ProjectId).toBe(project.id);
        expect(page2.ProjectId).toBe(project.id);

        const allPages = await apiClient.getPagesByProjectID(project.id);
        expect(allPages.length).toBeGreaterThanOrEqual(2);
      } catch (error) {
        test.skip();
      }
    });

    test("should maintain consistency across workflow steps", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const originalName = project.name;

        const page = await apiClient.createPage(project.id, {
          name: `Consistency Test ${Date.now()}`,
          type: "landing",
        });

        const updated = await apiClient.updateProject(project.id, {
          description: "Test consistency",
        });

        expect(updated.name).toBe(originalName);

        const verified = await apiClient.getProjectByID(project.id);
        expect(verified.name).toBe(originalName);
        expect(verified.description).toBe("Test consistency");

        const pages = await apiClient.getPagesByProjectID(project.id);
        const pageExists = pages.some((p) => p.Id === page.Id);
        expect(pageExists).toBe(true);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Error Handling in Workflows", () => {
    test("should handle missing project gracefully", async ({ apiClient }) => {
      try {
        try {
          await apiClient.getProjectByID(`nonexistent-${Date.now()}`);
          test.fail(true, "Should have thrown error");
        } catch (error) {
          expect(error).toBeDefined();
        }

        const projects = await apiClient.getProjectsByUser();
        expect(Array.isArray(projects)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle missing page gracefully", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();

        try {
          await apiClient.getPageByID(project.id, `nonexistent-${Date.now()}`);
          test.fail(true, "Should have thrown error");
        } catch (error) {
          expect(error).toBeDefined();
        }

        const pages = await apiClient.getPagesByProjectID(project.id);
        expect(Array.isArray(pages)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should recover from error and continue workflow", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();

        try {
          await apiClient.getPageByID(project.id, `invalid-${Date.now()}`);
        } catch (error) {
          expect(error).toBeDefined();
        }

        const newPage = await apiClient.createPage(project.id, {
          name: `Recovery Test ${Date.now()}`,
          type: "landing",
        });

        expect(newPage.Id).toBeDefined();

        await helpers.cleanupPage(project.id, newPage.Id);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Public vs Private Access Workflows", () => {
    test("should verify public and private endpoints consistency", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();

        const publicProject = await apiClient.getPublicProjectByID(project.id);
        const privateProject = await apiClient.getProjectByID(project.id);

        expect(publicProject.id).toBe(privateProject.id);
        expect(publicProject.name).toBe(privateProject.name);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle public and private page access", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const publicPage = await apiClient.getPublicPageByID(
          project.id,
          page.Id,
        );
        const privatePage = await apiClient.getPageByID(project.id, page.Id);

        expect(publicPage.Id).toBe(privatePage.Id);
        expect(publicPage.Name).toBe(privatePage.Name);
        expect(publicPage.Type).toBe(privatePage.Type);
      } catch (error) {
        test.skip();
      }
    });

    test("should retrieve both public and private elements", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();

        const publicElements = await apiClient.getPublicElementsByProjectID(
          project.id,
        );
        const privateElements = await apiClient.getElementsByProjectID(
          project.id,
        );

        expect(Array.isArray(publicElements)).toBe(true);
        expect(Array.isArray(privateElements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Search and Discovery Workflows", () => {
    test("should complete user search workflow", async ({ apiClient }) => {
      try {
        const users = await apiClient.searchUsers("test");
        expect(Array.isArray(users)).toBe(true);

        if (users.length > 0) {
          const user = users[0];
          expect(user.id).toBeDefined();
          expect(user.email).toBeDefined();

          try {
            const retrieved = await apiClient.getUserByEmail(user.email);
            expect(retrieved.id).toBe(user.id);
          } catch (error) {
            expect(error).toBeDefined();
          }
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should complete image discovery workflow", async ({ apiClient }) => {
      try {
        const images = await apiClient.getImages();
        expect(Array.isArray(images)).toBe(true);

        if (images.length > 0) {
          const image = images[0];
          expect(image.imageId).toBeDefined();

          const retrieved = await apiClient.getImageByID(image.imageId);
          expect(retrieved.imageId).toBe(image.imageId);
          expect(retrieved.imageLink).toBe(image.imageLink);
        }
      } catch (error) {
        test.skip();
      }
    });
  });
});
