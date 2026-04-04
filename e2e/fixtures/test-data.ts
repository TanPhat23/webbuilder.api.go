export const testData = {
  validUserEmail: 'test@example.com',
  validUsername: 'testuser',
  invalidEmail: 'invalid-email',
  searchQuery: 'test',
  
  page: {
    valid: {
      name: 'Test Page',
      type: 'landing',
      styles: { color: 'blue' },
    },
    invalid: {
      name: '',
      type: '',
    },
  },

  project: {
    valid: {
      name: 'Test Project',
      description: 'A test project',
      published: false,
    },
    update: {
      name: 'Updated Project',
      description: 'An updated test project',
      published: true,
    },
  },

  image: {
    validName: 'test-image',
    base64Data: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
  },
};
