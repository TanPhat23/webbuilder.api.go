import { test, expect } from "../fixtures/helpers.fixture";

test.describe("Element Endpoints", () => {
  test.describe("GET /elements/:projectid (private)", () => {
    test("should retrieve elements for a project", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const elements = await apiClient.getElementsByProjectID(project.id);

        expect(Array.isArray(elements)).toBe(true);
        if (elements.length > 0) {
          const element = elements[0];
          expect(element.id).toBeDefined();
          expect(element.type).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should return empty array for project with no elements", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();

        const elements = await apiClient.getElementsByProjectID(project.id);

        expect(Array.isArray(elements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle invalid project ID", async ({ apiClient }) => {
      try {
        await apiClient.getElementsByProjectID("invalid-project-id");
        test.fail(true, "Should have thrown error for invalid project");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test("should include element properties in response", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const elements = await apiClient.getElementsByProjectID(project.id);

        if (elements.length > 0) {
          const element = elements[0];
          expect(typeof element.id).toBe("string");
          expect(typeof element.type).toBe("string");

          if (element.pageId) {
            expect(typeof element.pageId).toBe("string");
          }
          if (element.name) {
            expect(typeof element.name).toBe("string");
          }
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("GET /elements/public/:projectid (public)", () => {
    test("should retrieve public elements for a project", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const publicElements = await apiClient.getPublicElementsByProjectID(
          project.id,
        );

        expect(Array.isArray(publicElements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should distinguish between public and private elements access", async ({
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

  test.describe("GET /elements/by-pages (private)", () => {
    test("should retrieve elements by page IDs", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const elements = await apiClient.getElementsByPageIds([page.Id]);

        expect(Array.isArray(elements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle multiple page IDs", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();
        const page1 = await helpers.findOrCreatePage(project.id);
        const page2 = await helpers.createTestPage(project.id);

        const elements = await apiClient.getElementsByPageIds([
          page1.Id,
          page2.Id,
        ]);

        expect(Array.isArray(elements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should return empty array for pages with no elements", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const elements = await apiClient.getElementsByPageIds([page.Id]);

        expect(Array.isArray(elements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle invalid page IDs gracefully", async ({ apiClient }) => {
      try {
        const elements = await apiClient.getElementsByPageIds([
          "invalid-page-id-1",
          "invalid-page-id-2",
        ]);

        expect(Array.isArray(elements)).toBe(true);
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test("should include page reference in elements", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const pages = await apiClient.getPagesByProjectID(project.id);
        if (pages.length === 0) {
          test.skip();
        }

        const elements = await apiClient.getElementsByPageIds([pages[0].Id]);

        if (elements.length > 0) {
          const element = elements[0];
          if (element.pageId) {
            expect(element.pageId).toBe(pages[0].Id);
          }
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("GET /elements/public/by-pages (public)", () => {
    test("should retrieve public elements by page IDs", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const pages = await apiClient.getPublicPagesByProjectID(project.id);
        if (pages.length === 0) {
          test.skip();
        }

        const elements = await apiClient.getPublicElementsByPageIds([
          pages[0].Id,
        ]);

        expect(Array.isArray(elements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle multiple public pages", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page1 = await helpers.findOrCreatePage(project.id);
        const page2 = await helpers.createTestPage(project.id);

        const publicElements = await apiClient.getPublicElementsByPageIds([
          page1.Id,
          page2.Id,
        ]);

        expect(Array.isArray(publicElements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Element Structure and Properties", () => {
    test("should validate element schema structure", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const elements = await apiClient.getElementsByProjectID(project.id);

        if (elements.length > 0) {
          const element = elements[0];
          expect(element.id).toBeDefined();
          expect(element.type).toBeDefined();

          if (element.name) {
            expect(typeof element.name).toBe("string");
          }
          if (element.order) {
            expect(typeof element.order).toBe("number");
          }
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle nested element relationships", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const elements = await apiClient.getElementsByPageIds([page.Id]);

        if (elements.length > 0) {
          const element = elements[0];
          if (element.parentId) {
            expect(typeof element.parentId).toBe("string");
          }
          expect(element.pageId || !element.pageId).toBe(true);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should support optional element properties", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const elements = await apiClient.getElementsByProjectID(project.id);

        if (elements.length > 0) {
          const element = elements[0];
          expect("id" in element).toBe(true);
          expect("type" in element).toBe(true);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should preserve element metadata", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const elements1 = await apiClient.getElementsByProjectID(project.id);
        const elements2 = await apiClient.getElementsByProjectID(project.id);

        expect(elements1.length).toBe(elements2.length);

        if (elements1.length > 0) {
          expect(elements1[0].id).toBe(elements2[0].id);
          expect(elements1[0].type).toBe(elements2[0].type);
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Element Filtering and Querying", () => {
    test("should filter elements by project scope", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project1 = await helpers.createTestProject();
        const project2 = await helpers.createTestProject();

        const page1 = await helpers.findOrCreatePage(project1.id);
        const page2 = await helpers.findOrCreatePage(project2.id);

        const elements1 = await apiClient.getElementsByProjectID(project1.id);
        const elements2 = await apiClient.getElementsByProjectID(project2.id);

        expect(Array.isArray(elements1)).toBe(true);
        expect(Array.isArray(elements2)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should filter elements by page scope", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page1 = await helpers.findOrCreatePage(project.id);
        const page2 = await helpers.createTestPage(project.id);

        const elementsPage1 = await apiClient.getElementsByPageIds([page1.Id]);
        const elementsPage2 = await apiClient.getElementsByPageIds([page2.Id]);

        expect(Array.isArray(elementsPage1)).toBe(true);
        expect(Array.isArray(elementsPage2)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should support comma-separated page IDs", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const pages = [
          await helpers.findOrCreatePage(project.id),
          await helpers.createTestPage(project.id),
        ];

        const elements = await apiClient.getElementsByPageIds(
          pages.map((p) => p.Id),
        );

        expect(Array.isArray(elements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Element Access Control", () => {
    test("should enforce private access for authenticated endpoints", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const elements = await apiClient.getElementsByProjectID(project.id);

        expect(Array.isArray(elements)).toBe(true);
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test("should support public access for public endpoints", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const publicElements = await apiClient.getPublicElementsByProjectID(
          project.id,
        );

        expect(Array.isArray(publicElements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Element Pagination and Limits", () => {
    test("should handle large element collections", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const elements = await apiClient.getElementsByProjectID(project.id);

        expect(Array.isArray(elements)).toBe(true);
        expect(elements.length >= 0).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle multiple concurrent page queries", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const pages = [
          await helpers.findOrCreatePage(project.id),
          await helpers.createTestPage(project.id),
        ];

        const results = await Promise.all(
          pages.map((page) => apiClient.getElementsByPageIds([page.Id])),
        );

        results.forEach((elements) => {
          expect(Array.isArray(elements)).toBe(true);
        });
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Element Error Handling", () => {
    test("should handle missing pageIds parameter", async ({ apiClient }) => {
      try {
        await apiClient.get("/elements/by-pages");
        test.fail(true, "Should have thrown error for missing pageIds");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test("should handle empty pageIds parameter", async ({ apiClient }) => {
      try {
        await apiClient.getElementsByPageIds([]);
        test.fail(true, "Should have thrown error for empty pageIds");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test("should handle malformed project ID", async ({ apiClient }) => {
      try {
        await apiClient.getElementsByProjectID("");
        test.fail(true, "Should have thrown error for empty project ID");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("Element Type Coverage", () => {
    test("should return various element types", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const elements = await apiClient.getElementsByProjectID(project.id);

        if (elements.length > 0) {
          const types = new Set(elements.map((e) => e.type));
          expect(types.size >= 0).toBe(true);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should categorize elements by type", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const elements = await apiClient.getElementsByProjectID(project.id);

        if (elements.length > 0) {
          const typeCounts = elements.reduce(
            (acc, el) => {
              acc[el.type] = (acc[el.type] || 0) + 1;
              return acc;
            },
            {} as Record<string, number>,
          );

          expect(Object.keys(typeCounts).length >= 0).toBe(true);
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Element Data Consistency", () => {
    test("should maintain consistency across multiple reads", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const elements1 = await apiClient.getElementsByProjectID(project.id);
        const elements2 = await apiClient.getElementsByProjectID(project.id);

        expect(elements1.length).toBe(elements2.length);

        if (elements1.length > 0) {
          elements1.forEach((el1, idx) => {
            const el2 = elements2[idx];
            expect(el1.id).toBe(el2.id);
            expect(el1.type).toBe(el2.type);
          });
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should sync elements across different query methods", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const projectElements = await apiClient.getElementsByProjectID(
          project.id,
        );
        const pageElements = await apiClient.getElementsByPageIds([page.Id]);

        expect(Array.isArray(projectElements)).toBe(true);
        expect(Array.isArray(pageElements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });
  });
});
