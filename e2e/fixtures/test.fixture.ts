import { test as base, APIRequestContext } from '@playwright/test';
import { APIClient } from '../utils/api-client';

export interface TestFixtures {
  apiClient: APIClient;
}

export const test = base.extend<TestFixtures>({
  apiClient: async ({ request }, use) => {
    const baseURL = process.env.API_BASE_URL || 'http://localhost:8080/api/v1';
    const authToken = process.env.AUTH_TOKEN;
    const client = new APIClient(request, baseURL, authToken);
    await use(client);
  },
});

export { expect } from '@playwright/test';
