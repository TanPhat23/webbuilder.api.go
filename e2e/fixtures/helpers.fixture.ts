import { test as base } from '@playwright/test';
import { APIClient } from '../utils/api-client';
import { TestHelpers } from '../utils/test-helpers';

export interface TestFixturesWithHelpers {
  apiClient: APIClient;
  helpers: TestHelpers;
}

export const test = base.extend<TestFixturesWithHelpers>({
  apiClient: async ({ request }, use) => {
    const baseURL = process.env.API_BASE_URL || 'http://localhost:8080/api/v1';
    const authToken = process.env.AUTH_TOKEN;
    const client = new APIClient(request, baseURL, authToken);
    await use(client);
  },

  helpers: async ({ apiClient }, use) => {
    const helpers = new TestHelpers(apiClient);
    await use(helpers);
  },
});

export { expect } from '@playwright/test';
