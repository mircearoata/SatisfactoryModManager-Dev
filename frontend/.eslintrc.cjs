module.exports = {
  root: true,
  parser: '@typescript-eslint/parser',
  extends: ['eslint:recommended', 'plugin:@typescript-eslint/recommended'],
  plugins: ['svelte3', '@typescript-eslint'],
  overrides: [{ files: ['*.svelte'], processor: 'svelte3/svelte3' }],
  settings: {
    'svelte3/typescript': () => require('typescript')
  },
  parserOptions: {
    sourceType: 'module',
    ecmaVersion: 2020
  },
  env: {
    browser: true,
    es2017: true,
    node: true
  },
  rules: {
    'no-multi-spaces': 'error',
    indent: ['error', 2],
    quotes: ['error', 'single'],
    curly: ['error', 'multi-line'],
    'no-extra-semi': 'error',
    'no-var': 'error',
    'quote-props': ['error', 'as-needed', { keywords: false, unnecessary: true, numbers: false }],
    semi: ['error', 'always'],
  }
};