// Import testing-library utilities
import '@testing-library/jest-dom';
import {render} from '@testing-library/react';

// Mock global objects that might be undefined in the test environment
global.MutationObserver = class {
    constructor() {}
    disconnect() {}
    observe() {}
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

const renderWithIntl = (ui, options = {}) => {
    return render(ui, options);
};

export {renderWithIntl};

export const setup = () => {
    // Setup code here
}; 