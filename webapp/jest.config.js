module.exports = {
  testEnvironment: 'jsdom',
  testMatch: ['**/*.test.js', '**/*.test.jsx'],
  moduleNameMapper: {
    '^.+\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$': '<rootDir>/src/tests/fileMock.js',
    '^.+\\.(css|less|scss)$': 'identity-obj-proxy'
  },
  setupFilesAfterEnv: ['<rootDir>/src/tests/setup.js'],
  transformIgnorePatterns: [
    'node_modules/(?!(mattermost-redux)/)'
  ],
  collectCoverageFrom: [
    'src/**/*.{js,jsx}',
    '!src/tests/**',
    '!**/node_modules/**'
  ],
  coverageReporters: ['text', 'lcov'],
  verbose: true,
}; 