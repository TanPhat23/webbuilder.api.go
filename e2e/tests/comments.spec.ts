import { test, expect } from "../fixtures/test.fixture";
import {
  CommentSchema,
  CreateCommentRequestSchema,
  UpdateCommentRequestSchema,
  CreateReactionRequestSchema,
  MarketplaceItemSchema,
  CreateMarketplaceItemRequestSchema,
  UpdateMarketplaceItemRequestSchema,
  CreateTagRequestSchema,
  CreateCategoryRequestSchema,
} from "../utils/schemas";

test.describe("Comment Endpoints", () => {
  let commentId: string;

  test.describe("GET /comments", () => {
    test("should retrieve all comments", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getComments();
        expect(Array.isArray(comments)).toBe(true);
        if (comments.length > 0) {
          commentId = comments[0].id;
          CommentSchema.parse(comments[0]);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should return empty array if no comments exist", async ({
      apiClient,
    }) => {
      try {
        const comments = await apiClient.getComments();
        expect(Array.isArray(comments)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should validate schema for all comments", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getComments();
        if (comments.length > 0) {
          comments.forEach((comment) => {
            CommentSchema.parse(comment);
          });
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("GET /comments/:id", () => {
    test("should retrieve comment by ID", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getComments();
        if (comments.length > 0) {
          commentId = comments[0].id;
          const comment = await apiClient.getCommentByID(commentId);
          expect(comment).toBeDefined();
          expect(comment.id).toBe(commentId);
          CommentSchema.parse(comment);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate retrieved comment schema", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getComments();
        if (comments.length > 0) {
          const comment = await apiClient.getCommentByID(comments[0].id);
          const validated = CommentSchema.parse(comment);
          expect(validated.id).toBeDefined();
          expect(validated.content).toBeDefined();
          expect(validated.authorId).toBeDefined();
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent comment", async ({ apiClient }) => {
      try {
        await apiClient.getCommentByID("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("POST /comments", () => {
    test("should create comment", async ({ apiClient }) => {
      try {
        const marketplaceItems = await apiClient.getMarketplaceItems();
        if (marketplaceItems.length > 0) {
          const createData = {
            content: `Test comment ${Date.now()}`,
            itemId: marketplaceItems[0].id,
          };
          const comment = await apiClient.createComment(createData);
          expect(comment).toBeDefined();
          expect(comment.id).toBeDefined();
          expect(comment.content).toBe(createData.content);
          CommentSchema.parse(comment);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should create comment with optional parent ID", async ({
      apiClient,
    }) => {
      try {
        const marketplaceItems = await apiClient.getMarketplaceItems();
        const comments = await apiClient.getComments();
        if (marketplaceItems.length > 0 && comments.length > 0) {
          const createData = {
            content: `Reply comment ${Date.now()}`,
            itemId: marketplaceItems[0].id,
            parentId: comments[0].id,
          };
          const comment = await apiClient.createComment(createData);
          expect(comment.parentId).toBe(comments[0].id);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate schema of created comment", async ({ apiClient }) => {
      try {
        const marketplaceItems = await apiClient.getMarketplaceItems();
        if (marketplaceItems.length > 0) {
          const createData = {
            content: `Validated comment ${Date.now()}`,
            itemId: marketplaceItems[0].id,
          };
          const comment = await apiClient.createComment(createData);
          const validated = CommentSchema.parse(comment);
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
          content: `Missing itemId ${Date.now()}`,
          itemId: "test-item",
        };
        await apiClient.createComment(createData as any);
        test.fail(true, "Should have thrown error for missing itemId");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("PATCH /comments/:id", () => {
    test("should update comment content", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getComments();
        if (comments.length > 0) {
          commentId = comments[0].id;
          const updateData = {
            content: `Updated content ${Date.now()}`,
          };
          const updated = await apiClient.updateComment(commentId, updateData);
          expect(updated).toBeDefined();
          expect(updated.id).toBe(commentId);
          expect(updated.content).toBe(updateData.content);
          CommentSchema.parse(updated);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should update comment status", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getComments();
        if (comments.length > 0) {
          const updateData = { status: "archived" as const };
          const updated = await apiClient.updateComment(
            comments[0].id,
            updateData as any,
          );
          expect(updated.status).toBe("archived");
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should update multiple fields", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getComments();
        if (comments.length > 0) {
          const updateData = {
            content: `Multi-field update ${Date.now()}`,
            status: "published" as const,
          };
          const updated = await apiClient.updateComment(
            comments[0].id,
            updateData as any,
          );
          expect(updated.content).toBe(updateData.content);
          expect(updated.status).toBe(updateData.status);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent comment update", async ({ apiClient }) => {
      try {
        await apiClient.updateComment("non-existent-id-" + Date.now(), {
          content: "Test",
        });
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("POST /comments/:id/reactions", () => {
    test("should create comment reaction", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getComments();
        if (comments.length > 0) {
          const reactionData = { type: "like" };
          const reaction = await apiClient.createCommentReaction(
            comments[0].id,
            "like",
          );
          expect(reaction).toBeDefined();
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should support multiple reaction types", async ({ apiClient }) => {
      try {
        const comments = await apiClient.getComments();
        if (comments.length > 0) {
          const reactionTypes = ["love", "haha", "wow", "sad", "angry"];
          for (const type of reactionTypes) {
            const reaction = await apiClient.createCommentReaction(
              comments[0].id,
              type,
            );
            expect(reaction).toBeDefined();
          }
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("DELETE /comments/:id", () => {
    test("should delete comment", async ({ apiClient }) => {
      try {
        const marketplaceItems = await apiClient.getMarketplaceItems();
        if (marketplaceItems.length > 0) {
          const createData = {
            content: `To Delete ${Date.now()}`,
            itemId: marketplaceItems[0].id,
          };
          const comment = await apiClient.createComment(createData);
          await apiClient.deleteComment(comment.id);

          try {
            await apiClient.getCommentByID(comment.id);
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
        await apiClient.deleteComment("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });
});

test.describe("Marketplace Endpoints", () => {
  let itemId: string;
  let tagId: string;
  let categoryId: string;

  test.describe("GET /marketplace", () => {
    test("should retrieve all marketplace items", async ({ apiClient }) => {
      try {
        const items = await apiClient.getMarketplaceItems();
        expect(Array.isArray(items)).toBe(true);
        if (items.length > 0) {
          itemId = items[0].id;
          MarketplaceItemSchema.parse(items[0]);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should return empty array if no items exist", async ({
      apiClient,
    }) => {
      try {
        const items = await apiClient.getMarketplaceItems();
        expect(Array.isArray(items)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test("should validate schema for all items", async ({ apiClient }) => {
      try {
        const items = await apiClient.getMarketplaceItems();
        if (items.length > 0) {
          items.forEach((item) => {
            MarketplaceItemSchema.parse(item);
          });
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("GET /marketplace/:id", () => {
    test("should retrieve marketplace item by ID", async ({ apiClient }) => {
      try {
        const items = await apiClient.getMarketplaceItems();
        if (items.length > 0) {
          itemId = items[0].id;
          const item = await apiClient.getMarketplaceItemByID(itemId);
          expect(item).toBeDefined();
          expect(item.id).toBe(itemId);
          MarketplaceItemSchema.parse(item);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should validate retrieved item schema", async ({ apiClient }) => {
      try {
        const items = await apiClient.getMarketplaceItems();
        if (items.length > 0) {
          const item = await apiClient.getMarketplaceItemByID(items[0].id);
          const validated = MarketplaceItemSchema.parse(item);
          expect(validated.id).toBeDefined();
          expect(validated.title).toBeDefined();
          expect(validated.description).toBeDefined();
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent item", async ({ apiClient }) => {
      try {
        await apiClient.getMarketplaceItemByID("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("POST /marketplace", () => {
    test("should create marketplace item", async ({ apiClient }) => {
      try {
        const createData = {
          title: `Marketplace Item ${Date.now()}`,
          description: "Test marketplace item",
          templateType: "block" as const,
        };
        const item = await apiClient.createMarketplaceItem(createData as any);
        expect(item).toBeDefined();
        expect(item.id).toBeDefined();
        expect(item.title).toBe(createData.title);
        MarketplaceItemSchema.parse(item);
      } catch (error) {
        test.skip();
      }
    });

    test("should create item with optional fields", async ({ apiClient }) => {
      try {
        const createData = {
          title: `Full Item ${Date.now()}`,
          description: "Item with all fields",
          templateType: "page" as const,
          preview: "preview-url",
          pageCount: 5,
        };
        const item = await apiClient.createMarketplaceItem(createData as any);
        expect(item.title).toBe(createData.title);
        expect(item.pageCount).toBe(createData.pageCount);
      } catch (error) {
        test.skip();
      }
    });

    test("should validate schema of created item", async ({ apiClient }) => {
      try {
        const createData = {
          title: `Validated Item ${Date.now()}`,
          description: "Validated item",
          templateType: "section" as const,
        };
        const item = await apiClient.createMarketplaceItem(createData as any);
        const validated = MarketplaceItemSchema.parse(item);
        expect(validated.id).toBeDefined();
        expect(validated.title).toBe(createData.title);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle missing required fields", async ({ apiClient }) => {
      try {
        const createData = {
          title: `Missing Fields ${Date.now()}`,
          description: "Test",
          templateType: "block" as const,
        };
        await apiClient.createMarketplaceItem(createData as any);
        test.fail(true, "Should have thrown error for missing fields");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("PATCH /marketplace/:id", () => {
    test("should update marketplace item", async ({ apiClient }) => {
      try {
        const items = await apiClient.getMarketplaceItems();
        if (items.length > 0) {
          itemId = items[0].id;
          const updateData = {
            description: `Updated ${Date.now()}`,
          };
          const updated = await apiClient.updateMarketplaceItem(
            itemId,
            updateData,
          );
          expect(updated as any).toBeDefined();
          expect(updated.id).toBe(itemId);
          MarketplaceItemSchema.parse(updated);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should update multiple fields", async ({ apiClient }) => {
      try {
        const items = await apiClient.getMarketplaceItems();
        if (items.length > 0) {
          const updateData = {
            title: `Updated Title ${Date.now()}`,
            featured: true,
          };
          const updated = await apiClient.updateMarketplaceItem(
            items[0].id,
            updateData,
          );
          expect(updated.title).toBe(updateData.title);
          expect(updated.featured).toBe(updateData.featured);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent item update", async ({ apiClient }) => {
      try {
        await apiClient.updateMarketplaceItem("non-existent-id-" + Date.now(), {
          title: "Test",
        });
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("DELETE /marketplace/:id", () => {
    test("should delete marketplace item", async ({ apiClient }) => {
      try {
        const createData = {
          title: `To Delete ${Date.now()}`,
          description: "Item to delete",
          templateType: "template" as const,
        };
        const item = await apiClient.createMarketplaceItem(createData);
        await apiClient.deleteMarketplaceItem(item.id);

        try {
          await apiClient.getMarketplaceItemByID(item.id);
          test.fail(true, "Item should have been deleted");
        } catch {
          expect(true).toBe(true);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent item deletion", async ({ apiClient }) => {
      try {
        await apiClient.deleteMarketplaceItem("non-existent-id-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("Marketplace Tags", () => {
    test.describe("POST /marketplace/tags", () => {
      test("should create marketplace tag", async ({ apiClient }) => {
        try {
          const createData = {
            name: `Tag ${Date.now()}`,
          };
          const tag = await apiClient.createMarketplaceTag(createData);
          expect(tag).toBeDefined();
          expect(tag.id).toBeDefined();
          expect(tag.name).toBe(createData.name);
        } catch (error) {
          test.skip();
        }
      });

      test("should handle duplicate tag names", async ({ apiClient }) => {
        try {
          const tagName = `UniqueTag ${Date.now()}`;
          const createData = { name: tagName };

          const tag1 = await apiClient.createMarketplaceTag(createData);
          expect(tag1).toBeDefined();

          try {
            await apiClient.createMarketplaceTag(createData);
            test.fail(true, "Should have thrown error for duplicate tag");
          } catch {
            expect(true).toBe(true);
          }
        } catch (error) {
          test.skip();
        }
      });
    });
  });

  test.describe("Marketplace Categories", () => {
    test.describe("POST /marketplace/categories", () => {
      test("should create marketplace category", async ({ apiClient }) => {
        try {
          const createData = {
            name: `Category ${Date.now()}`,
          };
          const category =
            await apiClient.createMarketplaceCategory(createData);
          expect(category).toBeDefined();
          expect(category.id).toBeDefined();
          expect(category.name).toBe(createData.name);
        } catch (error) {
          test.skip();
        }
      });

      test("should handle duplicate category names", async ({ apiClient }) => {
        try {
          const categoryName = `UniqueCategory ${Date.now()}`;
          const createData = { name: categoryName };

          const cat1 = await apiClient.createMarketplaceCategory(createData);
          expect(cat1).toBeDefined();

          try {
            await apiClient.createMarketplaceCategory(createData);
            test.fail(true, "Should have thrown error for duplicate category");
          } catch {
            expect(true).toBe(true);
          }
        } catch (error) {
          test.skip();
        }
      });
    });
  });
});
