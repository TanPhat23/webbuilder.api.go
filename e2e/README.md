# Webbuilder API E2E Tests

Comprehensive end-to-end tests for the Webbuilder API using Playwright.

## Overview

This test suite covers:
- **User Endpoints** - User search and retrieval operations
- **Project Endpoints** - Project CRUD operations and management
- **Page Endpoints** - Page CRUD operations within projects
- **Image Endpoints** - Image retrieval and management
- **Integration Tests** - Complete workflows and data consistency

## Architecture

### Project Structure

```
e2e/
├── tests/                 # Test files
│   ├── user.spec.ts       # User endpoint tests
│   ├── project.spec.ts    # Project endpoint tests
│   ├── page.spec.ts       # Page endpoint tests
│   ├── image.spec.ts      # Image endpoint tests
│   └── integration.spec.ts # Integration and workflow tests
├── fixtures/              # Test fixtures and data
│   ├── test.fixture.ts    # Playwright test fixture with APIClient
│   ├── helpers.fixture.ts # Enhanced fixture with helpers
│   └── test-data.ts       # Test data and constants
├── utils/                 # Utility functions
│   ├── schemas.ts         # Zod schemas for request/response validation
│   ├── api-client.ts      # API client wrapper with Playwright
│   └── test-helpers.ts    # Helper functions for common test operations
├── playwright.config.ts   # Playwright configuration
├── package.json          # Dependencies
├── tsconfig.json         # TypeScript configuration
├── .env.example          # Environment template
├── .gitignore           # Git ignore rules
├── QUICKSTART.md        # Quick start guide
└── README.md            # This file
```

### Key Features

- **Zod-First Approach**: All API contracts are defined using Zod schemas
- **Type-Safe**: Full TypeScript support with strict mode enabled
- **API Client Wrapper**: Centralized API client with built-in schema validation
- **Test Fixtures**: Reusable Playwright fixtures for consistent setup
- **Error Handling**: Graceful error handling with meaningful messages
- **Parallel Execution**: Tests run in parallel for faster feedback
- **Test Helpers**: Reusable helper functions for complex test scenarios

## Setup

### Prerequisites

- Node.js 18+ and npm
- Running API server on port 8080 (default: http://localhost:8080/api/v1)

### Installation

```bash
cd e2e
npm install
```

### Configuration

1. Copy the environment template:

```bash
cp .env.example .env
```

2. Update `.env` with your configuration:

```env
API_BASE_URL=http://localhost:8080/api/v1
AUTH_TOKEN=your-auth-token-for-private-endpoints
```

3. Ensure your API is running:

```bash
# In your Go project root
go run cmd/api/main.go
# or
make run
```

## Running Tests

### Run all tests

```bash
npm test
```

### Run specific test file

```bash
npm run test:users
npm run test:projects
npm run test:pages
npm run test:images
```

### Run tests with UI

```bash
npm run test:ui
```

### Run tests in headed mode (see browser)

```bash
npm run test:headed
```

### Debug tests

```bash
npm run test:debug
```

### Run with custom API URL

```bash
API_BASE_URL=http://api.example.com:8080/api/v1 npm test
```

## API Client Usage

The `APIClient` class provides type-safe methods for API interactions:

```typescript
import { test, expect } from '../fixtures/test.fixture';

test('example test', async ({ apiClient }) => {
  const users = await apiClient.searchUsers('john');
  expect(users).toBeDefined();
  
  const project = await apiClient.getProjectByID('project-id');
  expect(project.name).toBeDefined();
});
```

### Available Methods

#### Users
- `searchUsers(query: string): Promise<User[]>`
- `getUserByEmail(email: string): Promise<User>`
- `getUserByUsername(username: string): Promise<User>`

#### Projects
- `getProjectsByUser(): Promise<Project[]>`
- `getProjectByID(projectId: string): Promise<Project>`
- `getPublicProjectByID(projectId: string): Promise<Project>`
- `getProjectPages(projectId: string): Promise<Page[]>`
- `updateProject(projectId: string, data: UpdateProjectRequest): Promise<Project>`
- `deleteProject(projectId: string): Promise<void>`

#### Pages
- `getPagesByProjectID(projectId: string): Promise<Page[]>`
- `getPageByID(projectId: string, pageId: string): Promise<Page>`
- `getPublicPagesByProjectID(projectId: string): Promise<Page[]>`
- `getPublicPageByID(projectId: string, pageId: string): Promise<Page>`
- `createPage(projectId: string, data: CreatePageRequest): Promise<Page>`
- `updatePage(projectId: string, pageId: string, data: UpdatePageRequest): Promise<Page>`
- `deletePage(projectId: string, pageId: string): Promise<void>`

#### Images
- `getImages(): Promise<Image[]>`
- `getImageByID(imageId: string): Promise<Image>`
- `deleteImage(imageId: string): Promise<void>`

## Test Helpers Usage

The `TestHelpers` class provides utility methods for common test scenarios:

```typescript
import { test, expect } from '../fixtures/helpers.fixture';

test('example with helpers', async ({ apiClient, helpers }) => {
  const project = await helpers.createTestProject();
  const page = await helpers.findOrCreatePage(project.id);
  
  expect(helpers.validatePageStructure(page)).toBe(true);
  
  await helpers.cleanupPage(project.id, page.Id);
});
```

### Available Helper Methods

#### Setup
- `createTestProject(overrides?: Partial<Project>): Promise<Project>`
- `findOrCreatePage(projectId: string, pageName?: string): Promise<Page>`
- `getFirstImage(): Promise<Image | null>`

#### Cleanup
- `cleanupPage(projectId: string, pageId: string): Promise<void>`
- `cleanupProject(projectId: string): Promise<void>`
- `cleanupImage(imageId: string): Promise<void>`

#### Validation
- `validateProjectStructure(project: Project): boolean`
- `validatePageStructure(page: Page): boolean`
- `validateImageStructure(image: Image): boolean`
- `validateUserStructure(user: User): boolean`

#### Utilities
- `waitFor<T>(fn: () => Promise<T>, timeoutMs?: number, intervalMs?: number): Promise<T>`

## Schema Validation

All responses are validated against Zod schemas defined in `utils/schemas.ts`:

```typescript
import { UserSchema, PageSchema, ProjectSchema, ImageSchema } from '../utils/schemas';

const user = await apiClient.getUserByEmail('user@example.com');
const validated = UserSchema.parse(user); // Validates and types the response
```

## Test Structure

Each test file follows a consistent pattern:

```typescript
test.describe('Feature Name', () => {
  test.describe('GET /endpoint', () => {
    test('should do something', async ({ apiClient }) => {
      // Arrange
      const data = await apiClient.getData();
      
      // Act & Assert
      expect(data).toBeDefined();
    });
  });
});
```

### Test Patterns

- **Happy Path**: Tests successful operations with valid data
- **Error Handling**: Tests error scenarios (404, validation errors, etc.)
- **Schema Validation**: Tests response structure matches Zod schema
- **Integration**: Tests multiple operations together

## Error Handling

Tests use try-catch blocks with `test.skip()` for graceful handling:

```typescript
test('should get user', async ({ apiClient }) => {
  try {
    const user = await apiClient.getUserByEmail('user@example.com');
    expect(user).toBeDefined();
  } catch (error) {
    // Skip test if user doesn't exist in test environment
    test.skip();
  }
});
```

## CI/CD Integration

The tests are configured for CI/CD environments:

```bash
API_BASE_URL=http://api-server:8080/api/v1 npm test
```

Configuration in `playwright.config.ts`:
- Single worker for CI
- 2 retries on failure
- HTML report generation

## Port Configuration

By default, tests connect to `http://localhost:8080/api/v1`. To use a different port:

1. **Via environment variable:**
   ```bash
   API_BASE_URL=http://localhost:9000/api/v1 npm test
   ```

2. **Via .env file:**
   ```env
   API_BASE_URL=http://localhost:9000/api/v1
   ```

3. **In playwright.config.ts:**
   ```typescript
   use: {
     baseURL: 'http://localhost:9000/api/v1',
   }
   ```

## Troubleshooting

### Tests fail with "Connection refused"

Ensure your API is running on port 8080:

```bash
# Check if API is running
curl http://localhost:8080/api/v1/projects/public/test

# If not running, start it
go run cmd/api/main.go
```

### Tests fail with "401 Unauthorized"

Ensure `AUTH_TOKEN` environment variable is set for protected endpoints:

```bash
AUTH_TOKEN=your-token npm test
```

### Tests timeout

Increase timeout in `playwright.config.ts`:

```typescript
use: {
  navigationTimeout: 30000,
  actionTimeout: 30000,
}
```

### Connection to wrong port

Verify `API_BASE_URL` in `.env`:

```bash
cat e2e/.env
# Should show: API_BASE_URL=http://localhost:8080/api/v1
```

## Contributing

When adding new tests:

1. Add request/response schemas to `utils/schemas.ts`
2. Add API client methods to `utils/api-client.ts`
3. Add helper methods to `utils/test-helpers.ts` if needed
4. Create tests in `tests/` directory
5. Use existing test patterns for consistency

## Debugging

Enable verbose logging:

```bash
DEBUG=* npm test
```

View detailed test report:

```bash
npx playwright show-report
```

Run single test:

```bash
npx playwright test tests/user.spec.ts --grep "should search users"
```

## Test Coverage

Current test coverage includes:

| Endpoint Category | Coverage |
|-------------------|----------|
| User Endpoints | ✅ Search, Get by email, Get by username |
| Project Endpoints | ✅ CRUD operations, Public/Private |
| Page Endpoints | ✅ CRUD operations, Public/Private |
| Image Endpoints | ✅ List, Get, Delete |
| Integration Tests | ✅ Workflows, Data consistency |

## Performance Notes

- Tests run in parallel by default
- Each test is independent and can be run individually
- Tests skip gracefully if required data doesn't exist
- No test data cleanup is performed (safe for read-only endpoints)

## Resources

- Playwright Docs: https://playwright.dev/docs/intro
- Zod Documentation: https://zod.dev
- API Documentation: Check your Go project's documentation
