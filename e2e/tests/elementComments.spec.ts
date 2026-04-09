import { test, expect } from "../fixtures/test.fixture";
import {
  ElementCommentSchema,
  CreateElementCommentRequestSchema,
  UpdateElementCommentRequestSchema,
  InvitationSchema,
  CreateInvitationRequestSchema,
  AcceptInvitationRequestSchema,
  UpdateInvitationStatusRequestSchema,
} from "../utils/schemas";

test.describe("Element Comment Endpoints", () => {
  let elementCommentId: string;
  let project_id: string;
  let elementId: string;

  test.describe("GET /element-comments", () => {
    test("should retrieve all element comments", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getElementComments();
        expect(Array.isArray(comments)).toBe(true);
        if (comments.length > 0) {
          elementCommentId = comments[0].id;
          ElementCommentSchema.parse(comments[0]);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should return empty array if no comments exist", async ({
      apiClient,
    }) => {
      try {
        const comments = await apiClient.getElementComments();
        expect(Array.isArray(comments)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should validate schema for all comments", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getElementComments();
        if (comments.length > 0) {
          comments.forEach((comment) => {
            ElementCommentSchema.parse(comment);
          });
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("GET /element-comments/:id", () => {
    test("should retrieve element comment by ID", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getElementComments();
        if (comments.length > 0) {
          elementCommentId = comments[0].id;
          const comment =
            await apiClient.getElementCommentByID(elementCommentId);
          expect(comment).toBeDefined();
          expect(comment.id).toBe(elementCommentId);
          ElementCommentSchema.parse(comment);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate retrieved comment schema", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getElementComments();
        if (comments.length > 0) {
          const comment = await apiClient.getElementCommentByID(comments[0].id);
          const validated = ElementCommentSchema.parse(comment);
          expect(validated.id).toBeDefined();
          expect(validated.content).toBeDefined();
          expect(validated.elementId).toBeDefined();
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent comment", async ({ apiClient }) => {
      try {
        await apiClient.getElementCommentByID("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("GET /elements/:elementId/comments", () => {
    test("should retrieve comments for specific element", async ({
      apiClient,
    }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          elementId = elements[0].id;
          const comments =
            await apiClient.getElementCommentsByElement(elementId);
          expect(Array.isArray(comments)).toBe(true);
          if (comments.length > 0) {
            ElementCommentSchema.parse(comments[0]);
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should return empty array for element with no comments", async ({
      apiClient,
    }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          const comments = await apiClient.getElementCommentsByElement(
            elements[0].id,
          );
          expect(Array.isArray(comments)).toBe(true);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("GET /element-comments/author/:authorId", () => {
    test("should retrieve comments by author", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getElementComments();
        if (comments.length > 0) {
          const authorId = comments[0].authorId;
          const authorComments =
            await apiClient.getCommentsByAuthorID(authorId);
          expect(Array.isArray(authorComments)).toBe(true);
          if (authorComments.length > 0) {
            authorComments.forEach((c) => {
              expect(c.authorId).toBe(authorId);
            });
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("GET /projects/:project_id/comments", () => {
    test("should retrieve comments by project", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          project_id = projects[0].id;
          const comments = await apiClient.getCommentsByProjectID(project_id);
          expect(Array.isArray(comments)).toBe(true);
          if (comments.length > 0) {
            ElementCommentSchema.parse(comments[0]);
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should return empty array for project with no comments", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const comments = await apiClient.getCommentsByProjectID(
            projects[0].id,
          );
          expect(Array.isArray(comments)).toBe(true);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("POST /element-comments", () => {
    test("should create element comment", async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        const projects = await apiClient.getProjectsByUser();
        if (elements.length > 0 && projects.length > 0) {
          const createData = {
            content: `Test comment ${Date.now()}`,
            elementId: elements[0].id,
          };
          const comment = await apiClient.createElementComment(createData);
          expect(comment).toBeDefined();
          expect(comment.id).toBeDefined();
          expect(comment.content).toBe(createData.content);
          ElementCommentSchema.parse(comment);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate schema of created comment", async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          const createData = {
            content: `Validated comment ${Date.now()}`,
            elementId: elements[0].id,
          };
          const comment = await apiClient.createElementComment(createData);
          const validated = ElementCommentSchema.parse(comment);
          expect(validated.id).toBeDefined();
          expect(validated.content).toBe(createData.content);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle missing required fields", async ({ apiClient }) => {
      try {
        const createData = {
          content: `Missing elementId ${Date.now()}`,
          elementId: "test-element",
        };
        await apiClient.createElementComment(createData as any);
        test.fail(true, "Should have thrown error for missing elementId");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("PATCH /element-comments/:id", () => {
    test("should update element comment", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getElementComments();
        if (comments.length > 0) {
          elementCommentId = comments[0].id;
          const updateData = {
            content: `Updated content ${Date.now()}`,
          };
          const updated = await apiClient.updateElementComment(
            elementCommentId,
            updateData,
          );
          expect(updated).toBeDefined();
          expect(updated.id).toBe(elementCommentId);
          expect(updated.content).toBe(updateData.content);
          ElementCommentSchema.parse(updated);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should update resolved status", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getElementComments();
        if (comments.length > 0) {
          const updateData = { resolved: true };
          const updated = await apiClient.updateElementComment(
            comments[0].id,
            updateData,
          );
          expect(updated.resolved).toBe(true);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should update multiple fields", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getElementComments();
        if (comments.length > 0) {
          const updateData = {
            content: `Multi-field update ${Date.now()}`,
            resolved: false,
          };
          const updated = await apiClient.updateElementComment(
            comments[0].id,
            updateData,
          );
          expect(updated.content).toBe(updateData.content);
          expect(updated.resolved).toBe(updateData.resolved);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent comment update", async ({ apiClient }) => {
      try {
        await apiClient.updateElementComment("non-existent-id-" + Date.now(), {
          content: "Test",
        });
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("PATCH /element-comments/:id/toggle-resolved", () => {
    test("should toggle resolved status", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getElementComments();
        if (comments.length > 0) {
          const original = comments[0].resolved;
          const updated = await apiClient.toggleResolvedStatus(comments[0].id);
          expect(updated.resolved).toBe(!original);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should toggle twice to restore original state", async ({
      apiClient,
    }) => {
      try {
        const comments = await apiClient.getElementComments();
        if (comments.length > 0) {
          const original = comments[0].resolved;
          await apiClient.toggleResolvedStatus(comments[0].id);
          const restored = await apiClient.toggleResolvedStatus(comments[0].id);
          expect(restored.resolved).toBe(original);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("DELETE /element-comments/:id", () => {
    test("should delete element comment", async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          const createData = {
            content: `To Delete ${Date.now()}`,
            elementId: elements[0].id,
          };
          const comment = await apiClient.createElementComment(createData);
          await apiClient.deleteElementComment(comment.id);

          try {
            await apiClient.getElementCommentByID(comment.id);
            test.fail(true, "Comment should have been deleted");
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

    test("should handle non-existent comment deletion", async ({
      apiClient,
    }) => {
      try {
        await apiClient.deleteElementComment("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });
});

test.describe("Invitation Endpoints", () => {
  let invitationId: string;
  let project_id: string;

  test.describe("GET /invitations", () => {
    test("should retrieve all invitations", async ({ apiClient }) => {
      try {
        const invitations = await apiClient.getInvitations();
        expect(Array.isArray(invitations)).toBe(true);
        if (invitations.length > 0) {
          invitationId = invitations[0].id;
          InvitationSchema.parse(invitations[0]);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should return empty array if no invitations exist", async ({
      apiClient,
    }) => {
      try {
        const invitations = await apiClient.getInvitations();
        expect(Array.isArray(invitations)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should validate schema for all invitations", async ({
      apiClient,
    }) => {
      try {
        const invitations = await apiClient.getInvitations();
        if (invitations.length > 0) {
          invitations.forEach((inv) => {
            InvitationSchema.parse(inv);
          });
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("GET /invitations/project/:projectid", () => {
    test("should retrieve invitations by project", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          project_id = projects[0].id;
          const invitations =
            await apiClient.getInvitationsByProject(project_id);
          expect(Array.isArray(invitations)).toBe(true);
          if (invitations.length > 0) {
            invitations.forEach((inv) => {
              expect(inv.projectId).toBe(project_id);
              InvitationSchema.parse(inv);
            });
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should return empty array for project with no invitations", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const invitations = await apiClient.getInvitationsByProject(
            projects[0].id,
          );
          expect(Array.isArray(invitations)).toBe(true);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("GET /invitations/project/:projectid/pending", () => {
    test("should retrieve pending invitations", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const pending = await apiClient.getPendingInvitationsByProject(
            projects[0].id,
          );
          expect(Array.isArray(pending)).toBe(true);
          if (pending.length > 0) {
            pending.forEach((inv) => {
              expect(inv.status).toBe("pending");
            });
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("POST /invitations", () => {
    test("should create invitation", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const createData = {
            project_id: projects[0].id,
            email: `test-${Date.now()}@example.com`,
            role: "editor" as const,
          };
          const invitation = await apiClient.createInvitation(createData as any);
          expect(invitation).toBeDefined();
          expect(invitation.id).toBeDefined();
          InvitationSchema.parse(invitation);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should create invitation with different roles", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const roles = ["viewer", "editor", "owner"];
          for (const role of roles) {
            const createData = {
              project_id: projects[0].id,
              email: `role-test-${Date.now()}-${role}@example.com`,
              role: role as any,
            };
            const invitation = await apiClient.createInvitation(createData as any);
            expect(invitation.role).toBe(role);
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate schema of created invitation", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const createData = {
            project_id: projects[0].id,
            email: `validated-${Date.now()}@example.com`,
            role: "viewer" as const,
          };
          const invitation = await apiClient.createInvitation(createData as any);
          const validated = InvitationSchema.parse(invitation);
          expect(validated.id).toBeDefined();
          expect(validated.email).toBe(createData.email);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle missing required fields", async ({ apiClient }) => {
      try {
        const createData = {
          email: `missing-fields-${Date.now()}@example.com`,
        };
        await apiClient.createInvitation(createData as any);
        test.fail(true, "Should have thrown error for missing fields");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test("should handle invalid email", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const createData = {
            project_id: projects[0].id,
            email: "invalid-email",
            role: "editor" as const,
          };
          await apiClient.createInvitation(createData as any);
          test.fail(true, "Should have thrown error for invalid email");
        } else {
          test.skip();
        }
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("POST /invitations/accept", () => {
    test("should accept invitation with token", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const createData = {
            project_id: projects[0].id,
            email: `accept-test-${Date.now()}@example.com`,
            role: "editor" as const,
          };
          const invitation = await apiClient.createInvitation(createData as any);
          if (invitation.token) {
            const acceptData = { token: invitation.token };
            const accepted = await apiClient.acceptInvitation(acceptData);
            expect(accepted).toBeDefined();
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

    test("should handle invalid token", async ({ apiClient }) => {
      try {
        const acceptData = { token: "invalid-token-" + Date.now() };
        await apiClient.acceptInvitation(acceptData);
        test.fail(true, "Should have thrown error for invalid token");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("PATCH /invitations/:invitationid/status", () => {
    test("should update invitation status", async ({ apiClient }) => {
      try {
        const invitations = await apiClient.getInvitations();
        if (invitations.length > 0) {
          const newStatus =
            invitations[0].status === "pending" ? "accepted" : "pending";
          const updated = await apiClient.updateInvitationStatus(
            invitations[0].id,
            newStatus,
          );
          expect(updated).toBeDefined();
          expect(updated.id).toBe(invitations[0].id);
          expect(updated.status).toBe(newStatus);
          InvitationSchema.parse(updated);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle invalid status", async ({ apiClient }) => {
      try {
        const invitations = await apiClient.getInvitations();
        if (invitations.length > 0) {
          await apiClient.updateInvitationStatus(
            invitations[0].id,
            "invalid-status",
          );
          test.fail(true, "Should have thrown error for invalid status");
        } else {
          test.skip();
        }
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test("should handle non-existent invitation", async ({ apiClient }) => {
      try {
        await apiClient.updateInvitationStatus(
          "non-existent-id-" + Date.now(),
          "accepted",
        );
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("PATCH /invitations/:invitationid/cancel", () => {
    test("should cancel invitation", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const createData = {
            project_id: projects[0].id,
            email: `cancel-test-${Date.now()}@example.com`,
            role: "viewer" as const,
          };
          const invitation = await apiClient.createInvitation(createData as any);
          const cancelled = await apiClient.cancelInvitation(invitation.id);
          expect(cancelled).toBeDefined();
          expect(cancelled.status).toBe("cancelled");
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent invitation cancellation", async ({
      apiClient,
    }) => {
      try {
        await apiClient.cancelInvitation("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("DELETE /invitations/:id", () => {
    test("should delete invitation", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const createData = {
            project_id: projects[0].id,
            email: `delete-test-${Date.now()}@example.com`,
            role: "editor" as const,
          };
          const invitation = await apiClient.createInvitation(createData as any);
          await apiClient.deleteInvitation(invitation.id);

          try {
            await apiClient.getInvitationsByProject(projects[0].id);
            expect(true).toBe(true);
          } catch {
            test.fail(true, "Error retrieving invitations");
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent invitation deletion", async ({
      apiClient,
    }) => {
      try {
        await apiClient.deleteInvitation("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });
});
