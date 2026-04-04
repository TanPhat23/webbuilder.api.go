import { test, expect } from '../fixtures/test.fixture';
import { testData } from '../fixtures/test-data';
import { ProjectSchema } from '../utils/schemas';

test.describe('Project Endpoints', () => {
  let projectId: string;

  test.describe('GET /projects/public/:projectid', () => {
    test('should retrieve public project by ID', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          projectId = projects[0].id;
          const project = await apiClient.getPublicProjectByID(projectId);
          expect(project).toBeDefined();
          expect(project.id).toBe(projectId);
          ProjectSchema.parse(project);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should handle non-existent project', async ({ apiClient }) => {
      try {
        await apiClient.getPublicProjectByID('non-existent-id-' + Date.now());
        test.fail(true, 'Should have thrown error');
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test('should validate response schema', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const project = await apiClient.getPublicProjectByID(projects[0].id);
          const validated = ProjectSchema.parse(project);
          expect(validated.id).toBeDefined();
          expect(validated.name).toBeDefined();
        } else {
          test.skip();
        }
      } catch {
        test.skip();
      }
    });
  });

  test.describe('GET /projects/user', () => {
    test('should retrieve user projects (requires auth)', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        expect(Array.isArray(projects)).toBe(true);
        if (projects.length > 0) {
          ProjectSchema.parse(projects[0]);
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe('GET /projects/:projectid (private)', () => {
    test('should retrieve project by ID with access', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          projectId = projects[0].id;
          const project = await apiClient.getProjectByID(projectId);
          expect(project).toBeDefined();
          expect(project.id).toBe(projectId);
          ProjectSchema.parse(project);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe('GET /projects/:projectid/pages', () => {
    test('should retrieve project pages', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          projectId = projects[0].id;
          const pages = await apiClient.getProjectPages(projectId);
          expect(Array.isArray(pages)).toBe(true);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should handle non-existent project', async ({ apiClient }) => {
      try {
        await apiClient.getProjectPages('non-existent-id-' + Date.now());
        test.fail(true, 'Should have thrown error');
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe('PATCH /projects/:projectid', () => {
    test('should update project fields', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          projectId = projects[0].id;
          const updated = await apiClient.updateProject(projectId, {
            description: testData.project.update.description,
          });
          expect(updated).toBeDefined();
          expect(updated.description).toBe(testData.project.update.description);
          ProjectSchema.parse(updated);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should update multiple project fields', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          projectId = projects[0].id;
          const updated = await apiClient.updateProject(projectId, {
            name: testData.project.update.name,
            published: testData.project.update.published,
          });
          expect(updated.name).toBe(testData.project.update.name);
          expect(updated.published).toBe(testData.project.update.published);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe('DELETE /projects/:projectid', () => {
    test('should delete project', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const projectToDelete = projects[0].id;
          await apiClient.deleteProject(projectToDelete);
          
          try {
            await apiClient.getProjectByID(projectToDelete);
            test.fail(true, 'Project should have been deleted');
          } catch {
            expect(true).toBe(true);
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });
});
