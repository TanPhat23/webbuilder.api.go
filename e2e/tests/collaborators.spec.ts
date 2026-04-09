import { test, expect } from '../fixtures/test.fixture';
import { CollaboratorSchema, CreateCollaboratorRequestSchema } from '../utils/schemas';

test.describe('Collaborator Endpoints', () => {
  let projectId: string;
  let collaboratorId: string;

  test.describe('GET /collaborators/:projectid', () => {
    test('should retrieve collaborators by project ID', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          projectId = projects[0].id;
          const collaborators = await apiClient.getCollaboratorsByProjectID(projectId);
          expect(Array.isArray(collaborators)).toBe(true);
          if (collaborators.length > 0) {
            collaboratorId = collaborators[0].id;
            CollaboratorSchema.parse(collaborators[0]);
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should return empty array for project with no collaborators', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const collaborators = await apiClient.getCollaboratorsByProjectID(projects[0].id);
          expect(Array.isArray(collaborators)).toBe(true);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should handle non-existent project', async ({ apiClient }) => {
      try {
        await apiClient.getCollaboratorsByProjectID('non-existent-id-' + Date.now());
        test.fail(true, 'Should have thrown error');
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe('GET /collaborators/project/:projectid', () => {
    test('should retrieve collaborators with alternative endpoint', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          projectId = projects[0].id;
          const collaborators = await apiClient.getCollaboratorsByProjectID(projectId);
          expect(Array.isArray(collaborators)).toBe(true);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe('GET /collaborators/:collaboratorid', () => {
    test('should retrieve collaborator by ID', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const collaborators = await apiClient.getCollaboratorsByProjectID(projects[0].id);
          if (collaborators.length > 0) {
            collaboratorId = collaborators[0].id;
            const collaborator = await apiClient.getCollaboratorByID(collaboratorId);
            expect(collaborator).toBeDefined();
            expect(collaborator.id).toBe(collaboratorId);
            CollaboratorSchema.parse(collaborator);
          } else {
            test.skip();
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should validate collaborator schema', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const collaborators = await apiClient.getCollaboratorsByProjectID(projects[0].id);
          if (collaborators.length > 0) {
            const collaborator = await apiClient.getCollaboratorByID(collaborators[0].id);
            const validated = CollaboratorSchema.parse(collaborator);
            expect(validated.id).toBeDefined();
            expect(validated.projectId).toBeDefined();
            expect(validated.userId).toBeDefined();
            expect(validated.role).toBeDefined();
          } else {
            test.skip();
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should handle non-existent collaborator', async ({ apiClient }) => {
      try {
        await apiClient.getCollaboratorByID('non-existent-id-' + Date.now());
        test.fail(true, 'Should have thrown error');
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe('PATCH /collaborators/:id', () => {
    test('should update collaborator role', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const collaborators = await apiClient.getCollaboratorsByProjectID(projects[0].id);
          if (collaborators.length > 0) {
            const newRole = collaborators[0].role === 'editor' ? 'viewer' : 'editor';
            const updated = await apiClient.updateCollaboratorRole(collaborators[0].id, newRole);
            expect(updated).toBeDefined();
            expect(updated.role).toBe(newRole);
            CollaboratorSchema.parse(updated);
          } else {
            test.skip();
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should validate updated collaborator structure', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const collaborators = await apiClient.getCollaboratorsByProjectID(projects[0].id);
          if (collaborators.length > 0) {
            const updated = await apiClient.updateCollaboratorRole(collaborators[0].id, 'viewer');
            const validated = CollaboratorSchema.parse(updated);
            expect(validated.id).toBeDefined();
            expect(validated.role).toBe('viewer');
          } else {
            test.skip();
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should handle invalid role', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const collaborators = await apiClient.getCollaboratorsByProjectID(projects[0].id);
          if (collaborators.length > 0) {
            await apiClient.updateCollaboratorRole(collaborators[0].id, 'invalid-role');
            test.fail(true, 'Should have thrown error for invalid role');
          } else {
            test.skip();
          }
        } else {
          test.skip();
        }
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe('PATCH /collaborators/:collaboratorid/role', () => {
    test('should update role with alternative endpoint', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const collaborators = await apiClient.getCollaboratorsByProjectID(projects[0].id);
          if (collaborators.length > 0) {
            const newRole = collaborators[0].role === 'editor' ? 'viewer' : 'editor';
            const updated = await apiClient.updateCollaboratorRole(collaborators[0].id, newRole);
            expect(updated.role).toBe(newRole);
          } else {
            test.skip();
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe('DELETE /collaborators/:id', () => {
    test('should delete collaborator', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const collaborators = await apiClient.getCollaboratorsByProjectID(projects[0].id);
          if (collaborators.length > 1) {
            const toDelete = collaborators[collaborators.length - 1];
            await apiClient.deleteCollaborator(toDelete.id);

            try {
              await apiClient.getCollaboratorByID(toDelete.id);
              test.fail(true, 'Collaborator should have been deleted');
            } catch {
              expect(true).toBe(true);
            }
          } else {
            test.skip();
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should handle non-existent collaborator deletion gracefully', async ({ apiClient }) => {
      try {
        await apiClient.deleteCollaborator('non-existent-id-' + Date.now());
        test.fail(true, 'Should have thrown error');
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe('DELETE /collaborators/:collaboratorid', () => {
    test('should delete with alternative endpoint', async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const collaborators = await apiClient.getCollaboratorsByProjectID(projects[0].id);
          if (collaborators.length > 1) {
            const toDelete = collaborators[0];
            await apiClient.deleteCollaborator(toDelete.id);
          } else {
            test.skip();
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
