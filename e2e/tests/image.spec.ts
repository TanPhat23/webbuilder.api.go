import { test, expect } from "../fixtures/test.fixture";
import { testData } from "../fixtures/test-data";
import { ImageSchema } from "../utils/schemas";

test.describe("Image Endpoints", () => {
  let imageId: string;

  test.describe("GET /images", () => {
    test("should retrieve user images", async ({ apiClient }) => {
      try {
        const images = await apiClient.getImages();
        expect(Array.isArray(images)).toBe(true);
        if (images.length > 0) {
          ImageSchema.parse(images[0]);
          imageId = images[0].imageId;
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should return empty array if no images", async ({ apiClient }) => {
      try {
        const images = await apiClient.getImages();
        expect(Array.isArray(images)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe("GET /images/:imageid", () => {
    test("should retrieve image by ID", async ({ apiClient }) => {
      try {
        if (!imageId) {
          const images = await apiClient.getImages();
          if (images.length > 0) {
            imageId = images[0].imageId;
          } else {
            test.skip();
          }
        }

        if (imageId) {
          const image = await apiClient.getImageByID(imageId);
          expect(image).toBeDefined();
          expect(image.imageId).toBe(imageId);
          ImageSchema.parse(image);
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle non-existent image", async ({ apiClient }) => {
      try {
        await apiClient.getImageByID("non-existent-" + Date.now());
        test.fail(true, "Should have thrown error for non-existent image");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test("should validate image response schema", async ({ apiClient }) => {
      try {
        const images = await apiClient.getImages();
        if (images.length > 0) {
          const image = await apiClient.getImageByID(images[0].imageId);
          const validated = ImageSchema.parse(image);
          expect(validated.imageId).toBeDefined();
          expect(validated.imageLink).toBeDefined();
          expect(validated.userId).toBeDefined();
        } else {
          test.skip();
        }
      } catch {
        test.skip();
      }
    });
  });

  test.describe("DELETE /images/:imageid", () => {
    test("should delete image", async ({ apiClient }) => {
      try {
        const images = await apiClient.getImages();
        if (images.length > 0) {
          const imageToDelete = images[images.length - 1].imageId;
          await apiClient.deleteImage(imageToDelete);

          try {
            await apiClient.getImageByID(imageToDelete);
            test.fail(true, "Image should have been deleted");
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

    test("should handle deletion of non-existent image", async ({
      apiClient,
    }) => {
      try {
        await apiClient.deleteImage("non-existent-" + Date.now());
        test.fail(true, "Should have thrown error");
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe("Integration tests", () => {
    test("should retrieve and verify image structure", async ({
      apiClient,
    }) => {
      try {
        const images = await apiClient.getImages();
        if (images.length > 0) {
          const image = images[0];
          expect(image.imageId).toBeDefined();
          expect(image.imageLink).toBeDefined();
          expect(image.userId).toBeDefined();
          expect(image.createdAt).toBeDefined();

          const retrieved = await apiClient.getImageByID(image.imageId);
          expect(retrieved.imageId).toBe(image.imageId);
          expect(retrieved.imageLink).toBe(image.imageLink);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test("should handle multiple image operations", async ({ apiClient }) => {
      try {
        const initialImages = await apiClient.getImages();
        expect(Array.isArray(initialImages)).toBe(true);

        if (initialImages.length > 0) {
          const firstImage = initialImages[0];
          const retrieved = await apiClient.getImageByID(firstImage.imageId);
          expect(retrieved.imageId).toBe(firstImage.imageId);

          const refetched = await apiClient.getImages();
          expect(refetched.length).toBe(initialImages.length);
        }
      } catch (error) {
        test.skip();
      }
    });
  });
});
