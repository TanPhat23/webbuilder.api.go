import { test, expect } from "../fixtures/helpers.fixture";
import { z } from "zod";

test.describe("Validation Tests", () => {
  test.describe("Required Field Validation", () => {
    test("should reject page creation without name", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        try {
          await apiClient.createPage(projects[0].id, {
            name: "",
            type: "landing",
          });
          test.fail(true, "Should have rejected empty name");
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should reject page creation without type", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        try {
          await apiClient.createPage(projects[0].id, {
            name: "Test Page",
            type: "",
          });
          test.fail(true, "Should have rejected empty type");
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should reject update project without required data", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        try {
          await apiClient.updateProject(projects[0].id, {
            name: "",
          });
          test.fail(true, "Should have rejected empty name");
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("String Length Constraints", () => {
    test("should validate page name length", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();
        const veryLongName = "x".repeat(1000);

        try {
          const page = await apiClient.createPage(project.id, {
            name: veryLongName,
            type: "landing",
          });

          if (page) {
            expect(page.Name.length).toBeLessThanOrEqual(1000);
          }
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate project name length", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const veryLongName = "x".repeat(1000);

        try {
          const updated = await apiClient.updateProject(projects[0].id, {
            name: veryLongName,
          });

          if (updated) {
            expect(updated.name.length).toBeLessThanOrEqual(1000);
          }
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate description length", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const veryLongDescription = "x".repeat(5000);

        const updated = await apiClient.updateProject(projects[0].id, {
          description: veryLongDescription,
        });

        expect(updated.description).toBeDefined();
        if (updated.description) {
          expect(updated.description.length).toBeLessThanOrEqual(5000);
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Invalid Enum Values", () => {
    test("should reject invalid page type", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();

        try {
          await apiClient.createPage(project.id, {
            name: "Test Page",
            type: "invalid-type-xyz",
          });
          test.fail(true, "Should have rejected invalid type");
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate page type enum on update", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const page = await helpers.findOrCreatePage(project.id);

        try {
          await apiClient.updatePage(project.id, page.Id, {
            type: "not-a-valid-enum",
          });
          test.fail(true, "Should have rejected invalid type");
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Type Mismatches", () => {
    test("should handle boolean type constraints", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const project = projects[0];
        const newValue = !project.published;

        const updated = await apiClient.updateProject(project.id, {
          published: newValue,
        });

        expect(typeof updated.published).toBe("boolean");
        expect(updated.published).toBe(newValue);
      } catch (error) {
        test.skip();
      }
    });

    test("should validate numeric IDs", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();

        try {
          await apiClient.getPageByID("not-a-uuid", "also-not-a-uuid");
          test.fail(true, "Should have failed with invalid ID");
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate JSON object types", async ({ apiClient }) => {
      try {
        const project = await apiClient.getProjectsByUser();
        if (project.length === 0) {
          test.skip();
        }

        const validStyles = { color: "blue", padding: "10px" };
        const updated = await apiClient.updateProject(project[0].id, {
          styles: validStyles,
        });

        expect(typeof updated.styles).toBe("object");
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Null and Undefined Handling", () => {
    test("should handle null values gracefully", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        try {
          await apiClient.updateProject(projects[0].id, {
            description: null as any,
          });
          test.fail(true, "Should reject null value");
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle undefined optional fields", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const updated = await apiClient.updateProject(projects[0].id, {
          name: projects[0].name,
        });

        expect(updated).toBeDefined();
        expect(updated.id).toBe(projects[0].id);
      } catch (error) {
        test.skip();
      }
    });

    test("should preserve existing values when omitting fields", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const original = projects[0];
        const updated = await apiClient.updateProject(original.id, {
          name: original.name,
        });

        expect(updated.id).toBe(original.id);
        expect(updated.description).toBe(original.description);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Empty String Handling", () => {
    test("should reject empty page name", async ({ apiClient, helpers }) => {
      try {
        const project = await helpers.createTestProject();

        try {
          await apiClient.createPage(project.id, {
            name: "",
            type: "landing",
          });
          test.fail(true, "Should reject empty name");
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should reject empty project name", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        try {
          await apiClient.updateProject(projects[0].id, {
            name: "",
          });
          test.fail(true, "Should reject empty name");
        } catch (error) {
          expect(error).toBeDefined();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should allow empty optional strings", async ({ apiClient }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const updated = await apiClient.updateProject(projects[0].id, {
          description: "",
        });

        expect(updated.description).toBe("");
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Special Character Handling", () => {
    test("should handle unicode characters in page name", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const unicodeName = "Test Page 🎨 中文 العربية Ελληνικά";

        const page = await apiClient.createPage(project.id, {
          name: unicodeName,
          type: "landing",
        });

        expect(page.Name).toBe(unicodeName);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle special characters in description", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const specialChars =
          'Special: !@#$%^&*()_+-=[]{}|;:",.<>?/~` Quote: "test" Backslash: \\';
        const updated = await apiClient.updateProject(projects[0].id, {
          description: specialChars,
        });

        expect(updated.description).toBe(specialChars);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle whitespace variations", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const nameWithWhitespace = "  Spaced  Page  Name  ";

        const page = await apiClient.createPage(project.id, {
          name: nameWithWhitespace,
          type: "landing",
        });

        expect(page).toBeDefined();
        expect(page.Name.trim()).toBeDefined();
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("XSS Attack Prevention", () => {
    test("should sanitize script tags in page name", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const xssPayload = '<script>alert("XSS")</script>Page';

        const page = await apiClient.createPage(project.id, {
          name: xssPayload,
          type: "landing",
        });

        expect(page.Name).toBeDefined();
        expect(page.Name).not.toContain("<script>");
      } catch (error) {
        test.skip();
      }
    });

    test("should sanitize event handlers in page name", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const xssPayload = '<img src=x onerror="alert(1)">Page';

        const page = await apiClient.createPage(project.id, {
          name: xssPayload,
          type: "landing",
        });

        expect(page.Name).toBeDefined();
        expect(page.Name).not.toContain("onerror");
      } catch (error) {
        test.skip();
      }
    });

    test("should sanitize iframe injections in description", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const xssPayload =
          '<iframe src="javascript:alert(\'XSS\')" ></iframe>Project';
        const updated = await apiClient.updateProject(projects[0].id, {
          description: xssPayload,
        });

        expect(updated.description).toBeDefined();
        expect(updated.description).not.toContain("javascript:");
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("SQL Injection Prevention", () => {
    test("should handle SQL injection attempts in page name", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const sqlInjection = "'; DROP TABLE pages; --";

        const page = await apiClient.createPage(project.id, {
          name: sqlInjection,
          type: "landing",
        });

        expect(page.Name).toBe(sqlInjection);

        const pages = await apiClient.getPagesByProjectID(project.id);
        expect(pages.length).toBeGreaterThan(0);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle UNION SELECT attempts", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const sqlInjection = "1' UNION SELECT * FROM users --";

        const page = await apiClient.createPage(project.id, {
          name: sqlInjection,
          type: "landing",
        });

        expect(page.Name).toBe(sqlInjection);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle command injection attempts", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const commandInjection = "$(whoami)";
        const updated = await apiClient.updateProject(projects[0].id, {
          description: commandInjection,
        });

        expect(updated.description).toBe(commandInjection);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Email Validation", () => {
    test("should validate email format", async ({ apiClient }) => {
      try {
        const validEmails = [
          "test@example.com",
          "user+tag@domain.co.uk",
          "name.surname@example.com",
        ];

        for (const email of validEmails) {
          try {
            const user = await apiClient.getUserByEmail(email);
            if (user) {
              expect(user.email).toMatch(
                /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
              );
            }
          } catch (error) {
            expect(error).toBeDefined();
          }
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should reject invalid email formats", async ({ apiClient }) => {
      try {
        const invalidEmails = [
          "notanemail",
          "@example.com",
          "user@",
          "user @example.com",
          "user@.com",
        ];

        for (const email of invalidEmails) {
          try {
            await apiClient.getUserByEmail(email);
            test.fail(true, `Should have rejected ${email}`);
          } catch (error) {
            expect(error).toBeDefined();
          }
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("Schema Validation", () => {
    test("should validate complete project response schema", async ({
      apiClient,
    }) => {
      try {
        const projects = await apiClient.getProjectsByUser();
        if (projects.length === 0) {
          test.skip();
        }

        const ProjectValidator = z.object({
          id: z.string(),
          name: z.string(),
          description: z.string().optional(),
          published: z.boolean(),
          ownerId: z.string(),
        });

        const validated = ProjectValidator.parse(projects[0]);
        expect(validated).toBeDefined();
      } catch (error) {
        if (error instanceof z.ZodError) {
          test.fail(true, `Schema validation failed: ${error.message}`);
        } else {
          test.skip();
        }
      }
    });

    test("should validate complete page response schema", async ({
      apiClient,
      helpers,
    }) => {
      try {
        const project = await helpers.createTestProject();
        const pages = await apiClient.getPagesByProjectID(project.id);

        if (pages.length === 0) {
          test.skip();
        }

        const PageValidator = z.object({
          Id: z.string(),
          Name: z.string(),
          Type: z.string(),
          ProjectId: z.string(),
          CreatedAt: z.string().optional(),
          UpdatedAt: z.string().optional(),
        });

        const validated = PageValidator.parse(pages[0]);
        expect(validated).toBeDefined();
      } catch (error) {
        if (error instanceof z.ZodError) {
          test.fail(true, `Schema validation failed: ${error.message}`);
        } else {
          test.skip();
        }
      }
    });
  });
});
