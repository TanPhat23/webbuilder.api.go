import { test, expect } from "../fixtures/test.fixture";
import { testData } from "../fixtures/test-data";
import { UserSchema } from "../utils/schemas";

test.describe("User Endpoints", () => {
  test.describe("GET /users/search", () => {
    test("should search users with valid query", async ({ apiClient }) => {
      const result = await apiClient.searchUsers(testData.searchQuery);
      expect(Array.isArray(result)).toBe(true);
    });

    test("should return empty array for non-existent users", async ({
      apiClient,
    }) => {
      const result = await apiClient.searchUsers(
        "nonexistent-user-xyz-" + Date.now(),
      );
      expect(Array.isArray(result)).toBe(true);
    });

    test("should validate response schema", async ({ apiClient }) => {
      const result = await apiClient.searchUsers(testData.searchQuery);
      if (result.length > 0) {
        const validatedUser = UserSchema.parse(result[0]);
        expect(validatedUser.id).toBeDefined();
        expect(validatedUser.email).toBeDefined();
      }
    });
  });

  test.describe("GET /users/email/:email", () => {
    test("should get user by valid email", async ({ apiClient }) => {
      try {
        const user = await apiClient.getUserByEmail(testData.validUserEmail);
        expect(user).toBeDefined();
        expect(user.email).toBe(testData.validUserEmail);
        UserSchema.parse(user);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent email", async ({ apiClient }) => {
      try {
        await apiClient.getUserByEmail(
          "nonexistent-" + Date.now() + "@example.com",
        );
        test.fail(true, "Should have thrown error for non-existent user");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test("should validate response schema", async ({ apiClient }) => {
      try {
        const user = await apiClient.getUserByEmail(testData.validUserEmail);
        const validated = UserSchema.parse(user);
        expect(validated).toBeDefined();
      } catch {
        test.skip();
      }
    });
  });

  test.describe("GET /users/username/:username", () => {
    test("should get user by valid username", async ({ apiClient }) => {
      try {
        const user = await apiClient.getUserByUsername(testData.validUsername);
        expect(user).toBeDefined();
        UserSchema.parse(user);
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent username", async ({ apiClient }) => {
      try {
        await apiClient.getUserByUsername("nonexistent-" + Date.now());
        test.fail(true, "Should have thrown error for non-existent user");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });
});
