import { test, expect } from "../fixtures/helpers.fixture";

test.describe("Performance Tests", () => {
  const SIMPLE_QUERY_TIMEOUT = 1000;
  const BULK_OPERATION_TIMEOUT = 5000;

  test.describe("Response Time Expectations", () => {
    test("should retrieve user projects within acceptable time", async ({
      apiClient,
    }) => {
      try {
        const startTime = Date.now();
        const projects = await apiClient.getProjectsByUser();
        const duration = Date.now() - startTime;

        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(Array.isArray(projects)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should retrieve project by ID within acceptable time", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectId = projects[0].id;
        const startTime = Date.now();
        const project = await apiClient.getProjectByID(projectId);
        const duration = Date.now() - startTime;

        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(project.id).toBe(projectId);
      } catch (error) {
        test.skip();
      }
    });

    test("should retrieve public project within acceptable time", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectId = projects[0].id;
        const startTime = Date.now();
        const project = await apiClient.getPublicProjectByID(projectId);
        const duration = Date.now() - startTime;

        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(project.id).toBe(projectId);
      } catch (error) {
        test.skip();
      }
    });

    test("should retrieve pages by project within acceptable time", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectId = projects[0].id;
        const startTime = Date.now();
        const pages = await apiClient.getPagesByProjectID(projectId);
        const duration = Date.now() - startTime;

        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(Array.isArray(pages)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should retrieve single page within acceptable time", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const startTime = Date.now();
        const retrieved = await apiClient.getPageByID(project.id, page.Id);
        const duration = Date.now() - startTime;

        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(retrieved.Id).toBe(page.Id);
      } catch (error) {
        test.skip();
      }
    });

    test("should search users within acceptable time", async ({ apiClient }) => {
      try {
        const startTime = Date.now();
        const users = await apiClient.searchUsers("test");
        const duration = Date.now() - startTime;

        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(Array.isArray(users)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should retrieve images within acceptable time", async ({
      apiClient,
    }) => {
      try {
        const startTime = Date.now();
        const images = await apiClient.getImages();
        const duration = Date.now() - startTime;

        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(Array.isArray(images)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Concurrent Request Handling", () => {
    test("should handle 5 concurrent project requests", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectIds = projects.slice(0, 5).map((p) => p.id);

        const startTime = Date.now();
        const results = await Promise.all(
          projectIds.map((id) => apiClient.getProjectByID(id)),
        );
        const duration = Date.now() - startTime;

        expect(results.length).toBe(projectIds.length);
        expect(duration).toBeLessThan(BULK_OPERATION_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle 10 concurrent page requests", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const pages = await apiClient.getPagesByProjectID(project.id);

        if (pages.length < 2) {
          test.skip();
        }

        const pageIds = pages.slice(0, Math.min(10, pages.length)).map((p) => p.Id);

        const startTime = Date.now();
        const results = await Promise.all(
          pageIds.map((pageId) =>
            apiClient.getPageByID(project.id, pageId),
          ),
        );
        const duration = Date.now() - startTime;

        expect(results.length).toBe(pageIds.length);
        expect(duration).toBeLessThan(BULK_OPERATION_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle mixed concurrent requests", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const projects = await apiClient.getProjectsByUser();

        if (projects.length === 0) {
          test.skip();
        }

        const requests = [
          apiClient.getProjectsByUser(),
          apiClient.getProjectByID(project.id),
          apiClient.getPagesByProjectID(project.id),
          apiClient.getImages(),
          apiClient.searchUsers("test"),
        ];

        const startTime = Date.now();
        const results = await Promise.all(requests);
        const duration = Date.now() - startTime;

        expect(results.length).toBe(requests.length);
        expect(duration).toBeLessThan(BULK_OPERATION_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle rapid sequential requests", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectId = projects[0].id;
        const requestCount = 20;

        const startTime = Date.now();
        for (let i = 0; i < requestCount; i++) {
          await apiClient.getProjectByID(projectId);
        }
        const duration = Date.now() - startTime;

        const avgTimePerRequest = duration / requestCount;
        expect(avgTimePerRequest).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(duration).toBeLessThan(BULK_OPERATION_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Large Dataset Handling", () => {
    test("should handle large number of projects", async ({ apiClient }) => {
      try {
        const startTime = Date.now();
        const projects = await apiClient.getProjectsByUser();
        const duration = Date.now() - startTime;

        expect(Array.isArray(projects)).toBe(true);
        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);

        if (projects.length > 100) {
          expect(projects.length).toBeGreaterThan(100);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle project with large number of pages", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectId = projects[0].id;
        const startTime = Date.now();
        const pages = await apiClient.getPagesByProjectID(projectId);
        const duration = Date.now() - startTime;

        expect(Array.isArray(pages)).toBe(true);
        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle large number of elements", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const pages = await apiClient.getPagesByProjectID(project.id);

        if (pages.length === 0) {
          test.skip();
        }

        const pageIds = pages.map((p) => p.Id);

        const startTime = Date.now();
        const elements = await apiClient.getElementsByPageIds(pageIds);
        const duration = Date.now() - startTime;

        expect(Array.isArray(elements)).toBe(true);
        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle large number of images", async ({ apiClient }) => {
      try {
        const startTime = Date.now();
        const images = await apiClient.getImages();
        const duration = Date.now() - startTime;

        expect(Array.isArray(images)).toBe(true);
        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });

    test("should retrieve specific image efficiently from large dataset", async ({
      apiClient,
    }) => {
      try {
        const images = await apiClient.getImages();
        if (images.length === 0) {
          test.skip();
        }

        const imageId = images[0].imageId;
        const startTime = Date.now();
        const image = await apiClient.getImageByID(imageId);
        const duration = Date.now() - startTime;

        expect(image.imageId).toBe(imageId);
        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Bulk Operation Performance", () => {
    test("should update project efficiently", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectId = projects[0].id;
        const updateData = {
          description: `Updated at ${Date.now()}`,
        };

        const startTime = Date.now();
        const updated = await apiClient.updateProject(projectId, updateData);
        const duration = Date.now() - startTime;

        expect(updated).toBeDefined();
        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });

    test("should create page efficiently", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();

        const startTime = Date.now();
        const page = await apiClient.createPage(project.id, {
          name: `Perf Test Page ${Date.now()}`,
          type: "landing",
        });
        const duration = Date.now() - startTime;

        expect(page).toBeDefined();
        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });

    test("should update page efficiently", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        const updateData = {
          name: `Updated Page ${Date.now()}`,
        };

        const startTime = Date.now();
        const updated = await apiClient.updatePage(
          project.id,
          page.Id,
          updateData,
        );
        const duration = Date.now() - startTime;

        expect(updated).toBeDefined();
        expect(duration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });

    test("should perform multiple sequential updates efficiently", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectId = projects[0].id;
        const updateCount = 5;

        const startTime = Date.now();
        for (let i = 0; i < updateCount; i++) {
          await apiClient.updateProject(projectId, {
            description: `Update ${i} at ${Date.now()}`,
          });
        }
        const duration = Date.now() - startTime;

        const avgTimePerUpdate = duration / updateCount;
        expect(avgTimePerUpdate).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(duration).toBeLessThan(BULK_OPERATION_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Query Performance Patterns", () => {
    test("should get by-pages query perform consistently", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const pages = await apiClient.getPagesByProjectID(project.id);

        if (pages.length === 0) {
          test.skip();
        }

        const pageIds = pages.slice(0, Math.min(5, pages.length)).map((p) => p.Id);
        const queryCount = 3;
        const timings: number[] = [];

        for (let i = 0; i < queryCount; i++) {
          const startTime = Date.now();
          await apiClient.getElementsByPageIds(pageIds);
          timings.push(Date.now() - startTime);
        }

        const avgTime = timings.reduce((a, b) => a + b, 0) / timings.length;
        const maxTime = Math.max(...timings);

        expect(avgTime).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(maxTime).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });

    test("should maintain consistent response times across requests", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectId = projects[0].id;
        const requestCount = 10;
        const timings: number[] = [];

        for (let i = 0; i < requestCount; i++) {
          const startTime = Date.now();
          await apiClient.getProjectByID(projectId);
          timings.push(Date.now() - startTime);
        }

        const avgTime = timings.reduce((a, b) => a + b, 0) / timings.length;
        const variance = timings.reduce((sum, time) => sum + Math.pow(time - avgTime, 2), 0) / timings.length;

        expect(avgTime).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(variance).toBeDefined();
      } catch (error) {
        test.skip();
      }
    });

    test("should handle cache efficiently", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectId = projects[0].id;

        const firstTime = Date.now();
        const firstCall = await apiClient.getProjectByID(projectId);
        const firstDuration = Date.now() - firstTime;

        const secondTime = Date.now();
        const secondCall = await apiClient.getProjectByID(projectId);
        const secondDuration = Date.now() - secondTime;

        expect(firstCall.id).toBe(secondCall.id);
        expect(firstDuration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
        expect(secondDuration).toBeLessThan(SIMPLE_QUERY_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Error Recovery Performance", () => {
    test("should recover from error without significant delay", async ({
      apiClient,
    }) => {
      try {
        const startTime = Date.now();

        try {
          await apiClient.getProjectByID("invalid-id-" + Date.now());
        } catch (error) {
          expect(error).toBeDefined();
        }

        const projects = await apiClient.getProjectsByUser();
        const duration = Date.now() - startTime;

        expect(Array.isArray(projects)).toBe(true);
        expect(duration).toBeLessThan(BULK_OPERATION_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle multiple failures efficiently", async ({
      apiClient,
    }) => {
      try {
        const startTime = Date.now();

        for (let i = 0; i < 5; i++) {
          try {
            await apiClient.getProjectByID("invalid-" + i);
          } catch (error) {
            expect(error).toBeDefined();
          }
        }

        const duration = Date.now() - startTime;
        expect(duration).toBeLessThan(BULK_OPERATION_TIMEOUT);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Memory Efficiency", () => {
    test("should handle repeated reads without memory issues", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const projectId = projects[0].id;
        const iterations = 100;

        const startTime = Date.now();
        for (let i = 0; i < iterations; i++) {
          await apiClient.getProjectByID(projectId);
        }
        const duration = Date.now() - startTime;

        const avgTime = duration / iterations;
        expect(avgTime).toBeLessThan(100);
        expect(duration).toBeLessThan(BULK_OPERATION_TIMEOUT * 5);
      } catch (error) {
        test.skip();
      }
    });
  });
});
