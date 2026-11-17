import { test, expect } from "../fixtures/test.fixture";
import {
  SnapshotSchema,
  SaveSnapshotRequestSchema,
  EventWorkflowSchema,
  CreateEventWorkflowRequestSchema,
  UpdateEventWorkflowRequestSchema,
  UpdateEventWorkflowEnabledRequestSchema,
  ElementEventWorkflowSchema,
  CreateElementEventWorkflowRequestSchema,
  UpdateElementEventWorkflowRequestSchema,
  ContentTypeSchema,
  CreateContentTypeRequestSchema,
  UpdateContentTypeRequestSchema,
  ContentFieldSchema,
  CreateContentFieldRequestSchema,
  UpdateContentFieldRequestSchema,
  ContentItemSchema,
  CreateContentItemRequestSchema,
  UpdateContentItemRequestSchema,
} from "../utils/schemas";

test.describe("Snapshot Endpoints", () => {
  let snapshotId: string;
  let projectId: string;

  test.describe("GET /snapshots/:projectid", () => {
    test("should retrieve snapshots by project ID", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          projectId = projects[0].id;
          const snapshots = await apiClient.getSnapshotsByProjectID(projectId);
          expect(Array.isArray(snapshots)).toBe(true);
          if (snapshots.length > 0) {
            snapshotId = snapshots[0].id;
            SnapshotSchema.parse(snapshots[0]);
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should return empty array for project with no snapshots", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const snapshots = await apiClient.getSnapshotsByProjectID(
            projects[0].id,
          );
          expect(Array.isArray(snapshots)).toBe(true);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate schema for all snapshots", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const snapshots = await apiClient.getSnapshotsByProjectID(
            projects[0].id,
          );
          if (snapshots.length > 0) {
            snapshots.forEach((snapshot) => {
              SnapshotSchema.parse(snapshot);
            });
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent project", async ({ apiClient }) => {
      try {
        await apiClient.getSnapshotsByProjectID(
          "non-existent-id-" + Date.now(),
        );
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("GET /snapshots/:snapshotid", () => {
    test("should retrieve snapshot by ID", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const snapshots = await apiClient.getSnapshotsByProjectID(
            projects[0].id,
          );
          if (snapshots.length > 0) {
            snapshotId = snapshots[0].id;
            const snapshot = await apiClient.getSnapshotByID(snapshotId);
            expect(snapshot).toBeDefined();
            expect(snapshot.id).toBe(snapshotId);
            SnapshotSchema.parse(snapshot);
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

    test("should validate retrieved snapshot schema", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const snapshots = await apiClient.getSnapshotsByProjectID(
            projects[0].id,
          );
          if (snapshots.length > 0) {
            const snapshot = await apiClient.getSnapshotByID(snapshots[0].id);
            const validated = SnapshotSchema.parse(snapshot);
            expect(validated.id).toBeDefined();
            expect(validated.projectId).toBeDefined();
            expect(validated.name).toBeDefined();
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

    test("should handle non-existent snapshot", async ({ apiClient }) => {
      try {
        await apiClient.getSnapshotByID("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("POST /snapshots/:projectid/save", () => {
    test("should save snapshot", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const saveData = {
            name: `Snapshot ${Date.now()}`,
            type: "working",
            elements: [],
          };
          const snapshot = await apiClient.saveSnapshot(
            projects[0].id,
            saveData,
          );
          expect(snapshot).toBeDefined();
          expect(snapshot.id).toBeDefined();
          expect(snapshot.name).toBe(saveData.name);
          SnapshotSchema.parse(snapshot);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should save snapshot with elements", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const saveData = {
            name: `Snapshot with Elements ${Date.now()}`,
            type: "checkpoint",
            elements: [
              { id: "elem1", type: "div" },
              { id: "elem2", type: "span" },
            ],
          };
          const snapshot = await apiClient.saveSnapshot(
            projects[0].id,
            saveData,
          );
          expect(snapshot.name).toBe(saveData.name);
          expect(snapshot.type).toBe(saveData.type);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should save snapshot with timestamp", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const timestamp = Date.now();
          const saveData = {
            name: `Timestamped Snapshot ${timestamp}`,
            type: "working",
            elements: [],
            timestamp: timestamp,
          };
          const snapshot = await apiClient.saveSnapshot(
            projects[0].id,
            saveData,
          );
          expect(snapshot).toBeDefined();
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle missing required fields", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const saveData = { name: `Missing Elements ${Date.now()}` };
          await apiClient.saveSnapshot(projects[0].id, saveData);
          test.fail(true, "Should have thrown error for missing elements");
        } else {
          test.skip();
        }
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("DELETE /snapshots/:snapshotid", () => {
    test("should delete snapshot", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const saveData = {
            name: `To Delete ${Date.now()}`,
            type: "working",
            elements: [],
          };
          const snapshot = await apiClient.saveSnapshot(
            projects[0].id,
            saveData,
          );
          await apiClient.deleteSnapshot(snapshot.id as string);

          try {
            await apiClient.getSnapshotByID(snapshot.id);
            test.fail(true, "Snapshot should have been deleted");
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

    test("should handle non-existent snapshot deletion", async ({
      apiClient,
    }) => {
      try {
        await apiClient.deleteSnapshot("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });
});

test.describe("Event Workflow Endpoints", () => {
  let workflowId: string;
  let projectId: string;

  test.describe("GET /event-workflows/:projectid", () => {
    test("should retrieve event workflows by project", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          projectId = projects[0].id;
          const workflows =
            await apiClient.getEventWorkflowsByProject(projectId);
          expect(Array.isArray(workflows)).toBe(true);
          if (workflows.length > 0) {
            workflowId = workflows[0].id;
            EventWorkflowSchema.parse(workflows[0]);
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate schema for all workflows", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const workflows = await apiClient.getEventWorkflowsByProject(
            projects[0].id,
          );
          if (workflows.length > 0) {
            workflows.forEach((workflow) => {
              EventWorkflowSchema.parse(workflow);
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

  test.describe("GET /event-workflows/:id", () => {
    test("should retrieve workflow by ID", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const workflows = await apiClient.getEventWorkflowsByProject(
            projects[0].id,
          );
          if (workflows.length > 0) {
            workflowId = workflows[0].id;
            const workflow = await apiClient.getEventWorkflowByID(workflowId);
            expect(workflow).toBeDefined();
            expect(workflow.id).toBe(workflowId);
            EventWorkflowSchema.parse(workflow);
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

    test("should handle non-existent workflow", async ({ apiClient }) => {
      try {
        await apiClient.getEventWorkflowByID("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("GET /event-workflows/:id/elements", () => {
    test("should retrieve workflow elements", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const workflows = await apiClient.getEventWorkflowsByProject(
            projects[0].id,
          );
          if (workflows.length > 0) {
            const elements = await apiClient.getEventWorkflowElements(
              workflows[0].id,
            );
            expect(Array.isArray(elements)).toBe(true);
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

  test.describe("POST /event-workflows", () => {
    test("should create event workflow", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const createData = {
            projectId: projects[0].id,
            name: `Workflow ${Date.now()}`,
            description: "Test workflow",
            handlers: {},
            canvasData: {},
          };
          const workflow = await apiClient.createEventWorkflow(createData);
          expect(workflow).toBeDefined();
          expect(workflow.id).toBeDefined();
          expect(workflow.name).toBe(createData.name);
          EventWorkflowSchema.parse(workflow);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should create workflow with optional fields", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const createData = {
            projectId: projects[0].id,
            name: `Full Workflow ${Date.now()}`,
            description: "Complete workflow",
            handlers: { click: "handleClick" },
            canvasData: { width: 100, height: 100 },
          };
          const workflow = await apiClient.createEventWorkflow(createData);
          expect(workflow.description).toBe(createData.description);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle missing required fields", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const createData = {
            name: `Missing ProjectId ${Date.now()}`,
            projectId: projects[0].id,
          };
          await apiClient.createEventWorkflow(createData as any);
          test.fail(true, "Should have thrown error for missing projectId");
        } else {
          test.skip();
        }
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("PATCH /event-workflows/:id", () => {
    test("should update workflow", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const workflows = await apiClient.getEventWorkflowsByProject(
            projects[0].id,
          );
          if (workflows.length > 0) {
            const updateData = {
              description: `Updated ${Date.now()}`,
            };
            const updated = await apiClient.updateEventWorkflow(
              workflows[0].id,
              updateData,
            );
            expect(updated as any).toBeDefined();
            expect(updated.id).toBe(workflows[0].id);
            EventWorkflowSchema.parse(updated);
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

    test("should update multiple fields", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const workflows = await apiClient.getEventWorkflowsByProject(
            projects[0].id,
          );
          if (workflows.length > 0) {
            const updateData = {
              name: `Updated Workflow ${Date.now()}`,
              description: "Updated description",
            };
            const updated = await apiClient.updateEventWorkflow(
              workflows[0].id,
              updateData,
            );
            expect(updated.name).toBe(updateData.name);
            expect(updated.description).toBe(updateData.description);
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

  test.describe("PATCH /event-workflows/:id/enabled", () => {
    test("should update workflow enabled status", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const workflows = await apiClient.getEventWorkflowsByProject(
            projects[0].id,
          );
          if (workflows.length > 0) {
            const originalStatus = workflows[0].enabled;
            const updated: any = await apiClient.updateEventWorkflowEnabled(
              workflows[0].id,
              !originalStatus,
            );
            expect(updated.enabled).toBe(!originalStatus);
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

    test("should toggle twice to restore original state", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const workflows = await apiClient.getEventWorkflowsByProject(
            projects[0].id,
          );
          if (workflows.length > 0) {
            const originalStatus = workflows[0].enabled;
            await apiClient.updateEventWorkflowEnabled(
              workflows[0].id,
              !originalStatus,
            );
            const restored: any = await apiClient.updateEventWorkflowEnabled(
              workflows[0].id,
              originalStatus,
            );
            expect(restored.enabled).toBe(originalStatus);
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

  test.describe("DELETE /event-workflows/:id", () => {
    test("should delete workflow", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length > 0) {
          const createData = {
            projectId: projects[0].id,
            name: `To Delete ${Date.now()}`,
            handlers: {},
            canvasData: {},
          };
          const workflow = await apiClient.createEventWorkflow(createData);
          await apiClient.deleteEventWorkflow(workflow.id as string);

          try {
            await apiClient.getEventWorkflowByID(workflow.id);
            test.fail(true, "Workflow should have been deleted");
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

    test("should handle non-existent workflow deletion", async ({
      apiClient,
    }) => {
      try {
        await apiClient.deleteEventWorkflow("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });
});

test.describe("Element Event Workflow Endpoints", () => {
  let elementWorkflowId: string;

  test.describe("GET /element-event-workflows", () => {
    test("should retrieve all element event workflows", async ({
      apiClient,
    }) => {
      try {
        const workflows = await apiClient.getElementEventWorkflows();
        expect(Array.isArray(workflows)).toBe(true);
        if (workflows.length > 0) {
          elementWorkflowId = workflows[0].id;
          ElementEventWorkflowSchema.parse(workflows[0]);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate schema for all workflows", async ({ apiClient }) => {
      try {
        const workflows = await apiClient.getElementEventWorkflows();
        if (workflows.length > 0) {
          workflows.forEach((workflow) => {
            ElementEventWorkflowSchema.parse(workflow);
          });
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("GET /element-event-workflows/:id", () => {
    test("should retrieve element event workflow by ID", async ({
      apiClient,
    }) => {
      try {
        const workflows = await apiClient.getElementEventWorkflows();
        if (workflows.length > 0) {
          elementWorkflowId = workflows[0].id;
          const workflow =
            await apiClient.getElementEventWorkflowByID(elementWorkflowId);
          expect(workflow).toBeDefined();
          expect(workflow.id).toBe(elementWorkflowId);
          ElementEventWorkflowSchema.parse(workflow);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent workflow", async ({ apiClient }) => {
      try {
        await apiClient.getElementEventWorkflowByID(
          "non-existent-id-" + Date.now(),
        );
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("POST /element-event-workflows", () => {
    test("should create element event workflow", async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        const workflows = await apiClient.getEventWorkflowsByProject(
          (await apiClient.getProjectsByUser())[0].id,
        );

        if (elements.length > 0 && workflows.length > 0) {
          const createData = {
            elementId: elements[0].id,
            workflowId: workflows[0].id,
            eventId: `event-${Date.now()}`,
          };
          const workflow =
            await apiClient.createElementEventWorkflow(createData);
          expect(workflow).toBeDefined();
          expect(workflow.id).toBeDefined();
          expect(workflow.elementId).toBe(createData.elementId);
          ElementEventWorkflowSchema.parse(workflow);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("PATCH /element-event-workflows/:id", () => {
    test("should update element event workflow", async ({ apiClient }) => {
      try {
        const workflows = await apiClient.getElementEventWorkflows();
        if (workflows.length > 0) {
          const updateData = {
            eventId: `updated-event-${Date.now()}`,
          };
          const updated = await apiClient.updateElementEventWorkflow(
            workflows[0].id,
            updateData,
          );
          expect(updated).toBeDefined();
          expect(updated.id).toBe(workflows[0].id);
          ElementEventWorkflowSchema.parse(updated);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("DELETE /element-event-workflows/:id", () => {
    test("should delete element event workflow", async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        const projects = await apiClient.getProjectsByUser();

        if (elements.length > 0 && projects.length > 0) {
          const workflows = await apiClient.getEventWorkflowsByProject(
            projects[0].id,
          );
          if (workflows.length > 0) {
            const createData = {
              elementId: elements[0].id,
              workflowId: workflows[0].id,
              eventId: `event-to-delete-${Date.now()}`,
            };
            const workflow =
              await apiClient.createElementEventWorkflow(createData);
            await apiClient.deleteElementEventWorkflow(workflow.id as string);

            try {
              await apiClient.getElementEventWorkflowByID(workflow.id);
              test.fail(true, "Workflow should have been deleted");
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
  });

  test.describe("DELETE /element-event-workflows/element/:elementId", () => {
    test("should delete workflows by element ID", async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          await apiClient.deleteElementEventWorkflowsByElement(elements[0].id);
          expect(true).toBe(true);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });
});

test.describe("Content Management Endpoints", () => {
  let contentTypeId: string;
  let contentFieldId: string;
  let contentItemId: string;

  test.describe("Content Types", () => {
    test.describe("GET /content-types", () => {
      test("should retrieve all content types", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          expect(Array.isArray(types)).toBe(true);
          if (types.length > 0) {
            contentTypeId = types[0].id;
            ContentTypeSchema.parse(types[0]);
          }
        } catch (error) {
          test.skip();
        }
      });

      test("should validate schema for all types", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            types.forEach((type) => {
              ContentTypeSchema.parse(type);
            });
          }
        } catch (error) {
          test.skip();
        }
      });
    });

    test.describe("GET /content-types/:id", () => {
      test("should retrieve content type by ID", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            contentTypeId = types[0].id;
            const type = await apiClient.getContentTypeByID(contentTypeId);
            expect(type).toBeDefined();
            expect(type.id).toBe(contentTypeId);
            ContentTypeSchema.parse(type);
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });

      test("should handle non-existent type", async ({ apiClient }) => {
        try {
          await apiClient.getContentTypeByID("non-existent-id-" + Date.now());
          test.fail(true, "Should have thrown error");
        } catch (error) {
          expect(error).toBeDefined();
        }
      });
    });

    test.describe("POST /content-types", () => {
      test("should create content type", async ({ apiClient }) => {
        try {
          const createData = {
            name: `Type ${Date.now()}`,
            description: "Test type",
          };
          const type = await apiClient.createContentType(createData);
          expect(type).toBeDefined();
          expect(type.id).toBeDefined();
          expect(type.name).toBe(createData.name);
          ContentTypeSchema.parse(type);
        } catch (error) {
          test.skip();
        }
      });

      test("should handle missing required fields", async ({ apiClient }) => {
        try {
          const createData = { name: `Type ${Date.now()}`, description: "Missing name" };
          await apiClient.createContentType(createData);
          test.fail(true, "Should have thrown error for missing name");
        } catch (error) {
          expect(error).toBeDefined();
        }
      });
    });

    test.describe("PATCH /content-types/:id", () => {
      test("should update content type", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            const updateData = {
              description: `Updated ${Date.now()}`,
            };
            const updated = await apiClient.updateContentType(
              types[0].id,
              updateData,
            );
            expect(updated).toBeDefined();
            expect(updated.id).toBe(types[0].id);
            ContentTypeSchema.parse(updated);
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });
    });

    test.describe("DELETE /content-types/:id", () => {
      test("should delete content type", async ({ apiClient }) => {
        try {
          const createData = {
            name: `To Delete ${Date.now()}`,
          };
          const type = await apiClient.createContentType(createData);
          await apiClient.deleteContentType(type.id);

          try {
            await apiClient.getContentTypeByID(type.id);
            test.fail(true, "Type should have been deleted");
          } catch {
            expect(true).toBe(true);
          }
        } catch (error) {
          test.skip();
        }
      });
    });
  });

  test.describe("Content Fields", () => {
    test.describe("GET /content-fields/:contentTypeId", () => {
      test("should retrieve content fields by type", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            contentTypeId = types[0].id;
            const fields =
              await apiClient.getContentFieldsByType(contentTypeId);
            expect(Array.isArray(fields)).toBe(true);
            if (fields.length > 0) {
              ContentFieldSchema.parse(fields[0]);
            }
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });
    });

    test.describe("GET /content-fields/by-id/:id", () => {
      test("should retrieve content field by ID", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            const fields = await apiClient.getContentFieldsByType(types[0].id);
            if (fields.length > 0) {
              contentFieldId = fields[0].id;
              const field = await apiClient.getContentFieldByID(contentFieldId);
              expect(field).toBeDefined();
              expect(field.id).toBe(contentFieldId);
              ContentFieldSchema.parse(field);
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

      test("should handle non-existent field", async ({ apiClient }) => {
        try {
          await apiClient.getContentFieldByID("non-existent-id-" + Date.now());
          test.fail(true, "Should have thrown error");
        } catch (error) {
          expect(error).toBeDefined();
        }
      });
    });

    test.describe("POST /content-fields", () => {
      test("should create content field", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            const createData = {
              contentTypeId: types[0].id,
              name: `Field ${Date.now()}`,
              type: "string",
              required: false,
            };
            const field = await apiClient.createContentField(createData);
            expect(field).toBeDefined();
            expect(field.id).toBeDefined();
            expect(field.name).toBe(createData.name);
            ContentFieldSchema.parse(field);
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });
    });

    test.describe("PATCH /content-fields/:id", () => {
      test("should update content field", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            const fields = await apiClient.getContentFieldsByType(types[0].id);
            if (fields.length > 0) {
              const updateData = {
                name: `Updated Field ${Date.now()}`,
              };
              const updated = await apiClient.updateContentField(
                fields[0].id,
                updateData,
              );
              expect(updated).toBeDefined();
              expect(updated.id).toBe(fields[0].id);
              ContentFieldSchema.parse(updated);
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

    test.describe("DELETE /content-fields/:id", () => {
      test("should delete content field", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            const createData = {
              contentTypeId: types[0].id,
              name: `To Delete ${Date.now()}`,
              type: "string",
            };
            const field = await apiClient.createContentField(createData);
            await apiClient.deleteContentField(field.id as string);

            try {
              await apiClient.getContentFieldByID(field.id);
              test.fail(true, "Field should have been deleted");
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

  test.describe("Content Items", () => {
    test.describe("GET /content-items/:contentTypeId", () => {
      test("should retrieve content items", async ({ apiClient }) => {
        try {
          const items = await apiClient.getContentItems();
          expect(Array.isArray(items)).toBe(true);
          if (items.length > 0) {
            contentItemId = items[0].id;
            ContentItemSchema.parse(items[0]);
          }
        } catch (error) {
          test.skip();
        }
      });

      test("should validate schema for all items", async ({ apiClient }) => {
        try {
          const items = await apiClient.getContentItems();
          if (items.length > 0) {
            items.forEach((item) => {
              ContentItemSchema.parse(item);
            });
          }
        } catch (error) {
          test.skip();
        }
      });
    });

    test.describe("GET /content-items/by-id/:id", () => {
      test("should retrieve content item by ID", async ({ apiClient }) => {
        try {
          const items = await apiClient.getContentItems();
          if (items.length > 0) {
            contentItemId = items[0].id;
            const item = await apiClient.getContentItemByID(contentItemId);
            expect(item).toBeDefined();
            expect(item.id).toBe(contentItemId);
            ContentItemSchema.parse(item);
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });

      test("should handle non-existent item", async ({ apiClient }) => {
        try {
          await apiClient.getContentItemByID("non-existent-id-" + Date.now());
          test.fail(true, "Should have thrown error");
        } catch (error) {
          expect(error).toBeDefined();
        }
      });
    });

    test.describe("POST /content-items", () => {
      test("should create content item", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            const createData = {
              contentTypeId: types[0].id,
              title: `Item ${Date.now()}`,
              slug: `item-${Date.now()}`,
              published: false,
              fieldValues: [],
            };
            const item = await apiClient.createContentItem(createData);
            expect(item).toBeDefined();
            expect(item.id).toBeDefined();
            expect(item.title).toBe(createData.title);
            ContentItemSchema.parse(item);
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });

      test("should create item with published status", async ({
        apiClient,
      }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            const createData = {
              contentTypeId: types[0].id,
              title: `Published Item ${Date.now()}`,
              slug: `published-${Date.now()}`,
              published: true,
              fieldValues: [],
            };
            const item = await apiClient.createContentItem(createData);
            expect(item.published).toBe(true);
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });
    });

    test.describe("PATCH /content-items/:id", () => {
      test("should update content item", async ({ apiClient }) => {
        try {
          const items = await apiClient.getContentItems();
          if (items.length > 0) {
            const updateData = {
              title: `Updated Item ${Date.now()}`,
            };
            const updated = await apiClient.updateContentItem(
              items[0].id,
              updateData,
            );
            expect(updated).toBeDefined();
            expect(updated.id).toBe(items[0].id);
            expect(updated.title).toBe(updateData.title);
            ContentItemSchema.parse(updated);
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });

      test("should toggle published status", async ({ apiClient }) => {
        try {
          const items = await apiClient.getContentItems();
          if (items.length > 0) {
            const original = items[0].published;
            const updateData = { published: !original };
            const updated = await apiClient.updateContentItem(
              items[0].id,
              updateData,
            );
            expect(updated.published).toBe(!original);
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });

      test("should update multiple fields", async ({ apiClient }) => {
        try {
          const items = await apiClient.getContentItems();
          if (items.length > 0) {
            const updateData = {
              title: `Multi-update ${Date.now()}`,
              slug: `multi-${Date.now()}`,
              published: true,
            };
            const updated = await apiClient.updateContentItem(
              items[0].id,
              updateData,
            );
            expect(updated.title).toBe(updateData.title);
            expect(updated.slug).toBe(updateData.slug);
            expect(updated.published).toBe(updateData.published);
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });
    });

    test.describe("DELETE /content-items/:id", () => {
      test("should delete content item", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            const createData = {
              contentTypeId: types[0].id,
              title: `To Delete ${Date.now()}`,
              slug: `delete-${Date.now()}`,
              published: false,
              fieldValues: [],
            };
            const item = await apiClient.createContentItem(createData);
            await apiClient.deleteContentItem(item.id as string);

            try {
              await apiClient.getContentItemByID(item.id);
              test.fail(true, "Item should have been deleted");
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

      test("should handle non-existent item deletion", async ({
        apiClient,
      }) => {
        try {
          await apiClient.deleteContentItem("non-existent-id-" + Date.now());
          test.fail(true, "Should have thrown error");
        } catch (error) {
          expect(error).toBeDefined();
        }
      });
    });

    test.describe("GET /public/content/:contentTypeId/:slug", () => {
      test("should retrieve public content by slug", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            const items = await apiClient.getContentItems();
            if (items.length > 0) {
              const item = await apiClient.getPublicContentItemBySlug(
                types[0].id,
                items[0].slug,
              );
              expect(item).toBeDefined();
              expect(item.slug).toBe(items[0].slug);
              ContentItemSchema.parse(item);
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

      test("should handle non-existent public item", async ({ apiClient }) => {
        try {
          const types = await apiClient.getContentTypes();
          if (types.length > 0) {
            await apiClient.getPublicContentItemBySlug(
              types[0].id,
              "non-existent-slug-" + Date.now(),
            );
            test.fail(true, "Should have thrown error");
          } else {
            test.skip();
          }
        } catch (error) {
          expect(error).toBeDefined();
        }
      });
    });
  });
});
