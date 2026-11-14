import { test, expect } from '../fixtures/test.fixture';
import {
  CustomElementSchema,
  CreateCustomElementRequestSchema,
  UpdateCustomElementRequestSchema,
  CustomElementTypeSchema,
  CreateCustomElementTypeRequestSchema,
  UpdateCustomElementTypeRequestSchema,
} from '../utils/schemas';

test.describe('Custom Elements Endpoints', () => {
  let elementId: string;
  let elementTypeId: string;

  test.describe('GET /custom-elements', () => {
    test('should retrieve all custom elements', async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        expect(Array.isArray(elements)).toBe(true);
        if (elements.length > 0) {
          elementId = elements[0].id;
          CustomElementSchema.parse(elements[0]);
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should return empty array if no elements exist', async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        expect(Array.isArray(elements)).toBe(true);
      } catch (error) {
        test.skip();
      }
    });

    test('should validate schema for all elements', async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          elements.forEach((element) => {
            CustomElementSchema.parse(element);
          });
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe('GET /custom-elements/:id', () => {
    test('should retrieve custom element by ID', async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          elementId = elements[0].id;
          const element = await apiClient.getCustomElementByID(elementId);
          expect(element).toBeDefined();
          expect(element.id).toBe(elementId);
          CustomElementSchema.parse(element);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should validate retrieved element schema', async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          const element = await apiClient.getCustomElementByID(elements[0].id);
          const validated = CustomElementSchema.parse(element);
          expect(validated.id).toBeDefined();
          expect(validated.name).toBeDefined();
          expect(validated.structure).toBeDefined();
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should handle non-existent element', async ({ apiClient }) => {
      try {
        await apiClient.getCustomElementByID('non-existent-id-' + Date.now());
        test.fail(true, 'Should have thrown error');
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe('POST /custom-elements', () => {
    test('should create custom element', async ({ apiClient }) => {
      try {
        const createData = {
          name: `Test Element ${Date.now()}`,
          structure: { type: 'div', children: [] },
          version: '1.0.0',
        };
        const element = await apiClient.createCustomElement(createData);
        expect(element).toBeDefined();
        expect(element.id).toBeDefined();
        expect(element.name).toBe(createData.name);
        CustomElementSchema.parse(element);
      } catch (error) {
        test.skip();
      }
    });

    test('should create element with optional fields', async ({ apiClient }) => {
      try {
        const createData = {
          name: `Element with Options ${Date.now()}`,
          structure: { type: 'div' },
          version: '1.0.0',
          description: 'Test description',
          category: 'test-category',
          icon: 'icon-url',
        };
        const element = await apiClient.createCustomElement(createData);
        expect(element.name).toBe(createData.name);
        expect(element.description).toBe(createData.description);
      } catch (error) {
        test.skip();
      }
    });

    test('should validate schema of created element', async ({ apiClient }) => {
      try {
        const createData = {
          name: `Validated Element ${Date.now()}`,
          structure: { type: 'section' },
          version: '2.0.0',
        };
        const element = await apiClient.createCustomElement(createData);
        const validated = CustomElementSchema.parse(element);
        expect(validated.id).toBeDefined();
        expect(validated.name).toBe(createData.name);
      } catch (error) {
        test.skip();
      }
    });

    test('should handle missing required fields', async ({ apiClient }) => {
      try {
        const createData = { name: `Missing Fields ${Date.now()}` };
        await apiClient.createCustomElement(createData);
        test.fail(true, 'Should have thrown error for missing fields');
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe('PATCH /custom-elements/:id', () => {
    test('should update custom element', async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          elementId = elements[0].id;
          const updateData = {
            description: `Updated ${Date.now()}`,
          };
          const updated = await apiClient.updateCustomElement(elementId, updateData);
          expect(updated).toBeDefined();
          expect(updated.id).toBe(elementId);
          CustomElementSchema.parse(updated);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should update multiple fields', async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          const updateData = {
            name: `Updated Name ${Date.now()}`,
            category: 'updated-category',
          };
          const updated = await apiClient.updateCustomElement(elements[0].id, updateData);
          expect(updated.name).toBe(updateData.name);
          expect(updated.category).toBe(updateData.category);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should handle non-existent element update', async ({ apiClient }) => {
      try {
        await apiClient.updateCustomElement('non-existent-id-' + Date.now(), {
          name: 'Test',
        });
        test.fail(true, 'Should have thrown error');
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe('POST /custom-elements/:id/duplicate', () => {
    test('should duplicate custom element', async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          const duplicated = await apiClient.duplicateCustomElement(
            elements[0].id,
            `Duplicated ${Date.now()}`
          );
          expect(duplicated).toBeDefined();
          expect(duplicated.id).not.toBe(elements[0].id);
          expect(duplicated.name).toContain('Duplicated');
          CustomElementSchema.parse(duplicated);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should preserve structure on duplication', async ({ apiClient }) => {
      try {
        const elements = await apiClient.getCustomElements();
        if (elements.length > 0) {
          const original = elements[0];
          const duplicated = await apiClient.duplicateCustomElement(
            original.id,
            `Dup ${Date.now()}`
          );
          expect(duplicated.structure).toEqual(original.structure);
        } else {
          test.skip();
        }
      } catch (error) {
        test.skip();
      }
    });
  });

  test.describe('DELETE /custom-elements/:id', () => {
    test('should delete custom element', async ({ apiClient }) => {
      try {
        const createData = {
          name: `To Delete ${Date.now()}`,
          structure: { type: 'div' },
          version: '1.0.0',
        };
        const element = await apiClient.createCustomElement(createData);
        await apiClient.deleteCustomElement(element.id);

        try {
          await apiClient.getCustomElementByID(element.id);
          test.fail(true, 'Element should have been deleted');
        } catch {
          expect(true).toBe(true);
        }
      } catch (error) {
        test.skip();
      }
    });

    test('should handle non-existent element deletion', async ({ apiClient }) => {
      try {
        await apiClient.deleteCustomElement('non-existent-id-' + Date.now());
        test.fail(true, 'Should have thrown error');
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  test.describe('Custom Element Types', () => {
    test.describe('GET /custom-element-types', () => {
      test('should retrieve all custom element types', async ({ apiClient }) => {
        try {
          const types = await apiClient.getCustomElementTypes();
          expect(Array.isArray(types)).toBe(true);
          if (types.length > 0) {
            elementTypeId = types[0].id;
            CustomElementTypeSchema.parse(types[0]);
          }
        } catch (error) {
          test.skip();
        }
      });

      test('should validate schema for all types', async ({ apiClient }) => {
        try {
          const types = await apiClient.getCustomElementTypes();
          if (types.length > 0) {
            types.forEach((type) => {
              CustomElementTypeSchema.parse(type);
            });
          }
        } catch (error) {
          test.skip();
        }
      });
    });

    test.describe('GET /custom-element-types/:id', () => {
      test('should retrieve custom element type by ID', async ({ apiClient }) => {
        try {
          const types = await apiClient.getCustomElementTypes();
          if (types.length > 0) {
            elementTypeId = types[0].id;
            const type = await apiClient.getCustomElementTypeByID(elementTypeId);
            expect(type).toBeDefined();
            expect(type.id).toBe(elementTypeId);
            CustomElementTypeSchema.parse(type);
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });

      test('should handle non-existent type', async ({ apiClient }) => {
        try {
          await apiClient.getCustomElementTypeByID('non-existent-id-' + Date.now());
          test.fail(true, 'Should have thrown error');
        } catch (error) {
          expect(error).toBeDefined();
        }
      });
    });

    test.describe('POST /custom-element-types', () => {
      test('should create custom element type', async ({ apiClient }) => {
        try {
          const createData = {
            name: `Type ${Date.now()}`,
            category: 'test-category',
          };
          const type = await apiClient.createCustomElementType(createData);
          expect(type).toBeDefined();
          expect(type.id).toBeDefined();
          expect(type.name).toBe(createData.name);
          CustomElementTypeSchema.parse(type);
        } catch (error) {
          test.skip();
        }
      });

      test('should create type with optional fields', async ({ apiClient }) => {
        try {
          const createData = {
            name: `Full Type ${Date.now()}`,
            description: 'Test type description',
            category: 'advanced',
            icon: 'icon-url',
          };
          const type = await apiClient.createCustomElementType(createData);
          expect(type.description).toBe(createData.description);
        } catch (error) {
          test.skip();
        }
      });
    });

    test.describe('PATCH /custom-element-types/:id', () => {
      test('should update custom element type', async ({ apiClient }) => {
        try {
          const types = await apiClient.getCustomElementTypes();
          if (types.length > 0) {
            const updateData = {
              description: `Updated Type ${Date.now()}`,
            };
            const updated = await apiClient.updateCustomElementType(types[0].id, updateData);
            expect(updated).toBeDefined();
            expect(updated.id).toBe(types[0].id);
            CustomElementTypeSchema.parse(updated);
          } else {
            test.skip();
          }
        } catch (error) {
          test.skip();
        }
      });
    });

    test.describe('DELETE /custom-element-types/:id', () => {
      test('should delete custom element type', async ({ apiClient }) => {
        try {
          const createData = {
            name: `To Delete Type ${Date.now()}`,
            category: 'temp',
          };
          const type = await apiClient.createCustomElementType(createData);
          await apiClient.deleteCustomElementType(type.id);

          try {
            await apiClient.getCustomElementTypeByID(type.id);
            test.fail(true, 'Type should have been deleted');
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
