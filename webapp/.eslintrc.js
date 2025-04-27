module.exports = {
  root: true,
  parser: '@babel/eslint-parser',
  parserOptions: {
    ecmaVersion: 2022,
    sourceType: 'module',
    ecmaFeatures: {
      jsx: true,
      impliedStrict: true,
    },
    requireConfigFile: false,
    babelOptions: {
      presets: ['@babel/preset-react'],
    },
  },
  env: {
    browser: true,
    node: true,
    es6: true,
    jest: true,
  },
  extends: [
    'eslint:recommended',
    'plugin:react/recommended',
    'plugin:import/errors',
    'plugin:import/warnings',
  ],
  plugins: [
    'react',
    'import',
  ],
  settings: {
    react: {
      version: 'detect',
    },
    'import/resolver': {
      node: {
        extensions: ['.js', '.jsx', '.ts', '.tsx'],
      },
    },
  },
  rules: {
    'react/prop-types': 'warn',
    'no-console': ['warn', { allow: ['warn', 'error'] }],
    'no-unused-vars': ['error', { ignoreRestSiblings: true }],
    'import/no-unresolved': 'error',
  },
}; 