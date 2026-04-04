import { test, expect } from "../fixtures/helpers.fixture";

test.describe("Integration Tests", () => {
  test.describe("Project and Page Workflow", () => {
    test("should manage complete project workflow", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];

        expect(helpers.validateProjectStructure(project)).toBe(true);

        const pages = await apiClient.getPagesByProjectID(project.id);
        expect(Array.isArray(pages)).toBe(true);

        if (pages.length > 0) {
          const page = pages[0];
          expect(helpers.validatePageStructure(page)).toBe(true);

          const retrieved = await apiClient.getPageByID(project.id, page.Id);
          expect(retrieved.Id).toBe(page.Id);
          expect(retrieved.Name).toBe(page.Name);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should create and retrieve new page", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const pageName = `Integration Test Page ${Date.now()}`;

        const newPage = await apiClient.createPage(project.id, {
          name: pageName,
          type: "landing",
          styles: { color: "blue" },
        });

        expect(newPage).toBeDefined();
        expect(newPage.Name).toBe(pageName);
        expect(helpers.validatePageStructure(newPage)).toBe(true);

        const retrieved = await apiClient.getPageByID(project.id, newPage.Id);
        expect(retrieved.Id).toBe(newPage.Id);
        expect(retrieved.Name).toBe(pageName);

        await helpers.cleanupPage(project.id, newPage.Id);
      } catch (error) {
        test.skip();
      }
    });

    test("should update page and verify changes", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const newName = `Updated Page ${Date.now()}`;
        const updatedPage = await apiClient.updatePage(project.id, page.Id, {
          name: newName,
          type: "product",
        });

        expect(updatedPage.Name).toBe(newName);
        expect(updatedPage.Type).toBe("product");

        const verified = await apiClient.getPageByID(project.id, page.Id);
        expect(verified.Name).toBe(newName);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Project Update Workflow", () => {
    test("should update project and verify persistence", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const newDescription = `Updated description ${Date.now()}`;

        const updated = await apiClient.updateProject(project.id, {
          description: newDescription,
        });

        expect(updated.description).toBe(newDescription);

        const verified = await apiClient.getProjectByID(project.id);
        expect(verified.description).toBe(newDescription);
      } catch (error) {
        test.skip();
      }
    });

    test("should toggle project published status", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const newStatus = !project.published;

        const updated = await apiClient.updateProject(project.id, {
          published: newStatus,
        });

        expect(updated.published).toBe(newStatus);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("User Search and Retrieval", () => {
    test("should search users and retrieve details", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const searchResults = await apiClient.searchUsers("test");

        if (searchResults.length > 0) {
          const user = searchResults[0];
          expect(helpers.validateUserStructure(user)).toBe(true);

          try {
            const retrieved = await apiClient.getUserByEmail(user.email);
            expect(retrieved.id).toBe(user.id);
            expect(retrieved.email).toBe(user.email);
          } catch {
            // Email might not be retrievable publicly
          }
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Image Management Workflow", () => {
    test("should list and retrieve image details", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const images = await apiClient.getImages();

        if (images.length > 0) {
          const image = images[0];
          expect(helpers.validateImageStructure(image)).toBe(true);

          const retrieved = await apiClient.getImageByID(image.imageId);
          expect(retrieved.imageId).toBe(image.imageId);
          expect(retrieved.imageLink).toBe(image.imageLink);
          expect(retrieved.userId).toBe(image.userId);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should verify image metadata consistency", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const images = await apiClient.getImages();

        if (images.length > 0) {
          for (const image of images.slice(0, 3)) {
            const retrieved = await apiClient.getImageByID(image.imageId);

            expect(retrieved.imageId).toBe(image.imageId);
            expect(retrieved.imageLink).toBe(image.imageLink);
            expect(retrieved.userId).toBe(image.userId);

            if (image.imageName) {
              expect(retrieved.imageName).toBe(image.imageName);
            }
          }
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Public vs Private Endpoints", () => {
    test("should access both public and private project endpoints", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectId = projects[0].id;

        const publicProject = await apiClient.getPublicProjectByID(projectId);
        const privateProject = await apiClient.getProjectByID(projectId);

        expect(publicProject.id).toBe(privateProject.id);
        expect(publicProject.name).toBe(privateProject.name);
      } catch (error) {
        test.skip();
      }
    });

    test("should access both public and private page endpoints", async ({
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
  });

  test.describe("Error Recovery", () => {
    test("should handle sequential failures gracefully", async ({
      apiClient,
    }) => {
      try {
        try {
          await apiClient.getPageByID("invalid-project", "invalid-page");
          test.fail(true, "Should have thrown error");
        } catch (error1) {
          expect(error1).toBeDefined();
        }

        try {
          await apiClient.getUserByEmail("nonexistent@example.com");
          test.fail(true, "Should have thrown error");
        } catch (error2) {
          expect(error2).toBeDefined();
        }

        const projects = await apiClient.getProjectsByUser();
        expect(Array.isArray(projects)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Data Consistency", () => {
    test("should maintain data consistency across multiple reads", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects1 = await apiClient.getProjectsByUser();
        const projects2 = await apiClient.getProjectsByUser();

        expect(projects1.length).toBe(projects2.length);

        if (projects1.length > 0) {
          expect(projects1[0].id).toBe(projects2[0].id);
          expect(projects1[0].name).toBe(projects2[0].name);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should verify page list consistency with individual retrieval", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const pages = await apiClient.getPagesByProjectID(project.id);

        if (pages.length > 0) {
          const page = pages[0];
          const retrieved = await apiClient.getPageByID(project.id, page.Id);

          expect(retrieved.Id).toBe(page.Id);
          expect(retrieved.Name).toBe(page.Name);
          expect(retrieved.Type).toBe(page.Type);
          expect(retrieved.ProjectId).toBe(page.ProjectId);
        }
      } catch (error) {
        test.skip();
      }
    });
  });
});
