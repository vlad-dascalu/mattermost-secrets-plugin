// Import testing-library utilities
import '@testing-library/jest-dom';

// Mock global objects that might be undefined in the test environment
global.MutationObserver = class {
  constructor(callback) {}
  disconnect() {}
  observe(element, initObject) {}
};

// Mock fetch API if needed in tests
global.fetch = jest.fn(() =>
  Promise.resolve({
    json: () => Promise.resolve({}),
    text: () => Promise.resolve(''),
    ok: true,
  })
);

// Add any other global mocks or setup needed for tests 