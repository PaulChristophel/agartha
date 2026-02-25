import js from '@eslint/js';
import globals from 'globals';
import tseslint from 'typescript-eslint';
import jsxA11y from 'eslint-plugin-jsx-a11y';
import reactPlugin from 'eslint-plugin-react';
import importPlugin from 'eslint-plugin-import';
import reactHooks from 'eslint-plugin-react-hooks';
import prettierPlugin from 'eslint-plugin-prettier';
import prettierConfig from 'eslint-config-prettier';
import reactRefresh from 'eslint-plugin-react-refresh';
import perfectionist from 'eslint-plugin-perfectionist';
import unusedImports from 'eslint-plugin-unused-imports';

export default [
  { ignores: ['dist/**', 'build/**', 'coverage/**', 'node_modules/**'] },

  js.configs.recommended,
  ...tseslint.configs.recommended,

  {
    files: ['**/*.{js,cjs,mjs,jsx,ts,tsx}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.es2021,
      },
      parserOptions: {
        ecmaFeatures: { jsx: true },
      },
    },
    settings: {
      react: { version: 'detect' },
      'import/resolver': {
        alias: {
          map: [['src', './src']],
          extensions: ['.js', '.jsx', '.ts', '.tsx'],
        },
      },
    },
    plugins: {
      react: reactPlugin,
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
      import: importPlugin,
      'jsx-a11y': jsxA11y,
      'unused-imports': unusedImports,
      perfectionist,
      prettier: prettierPlugin,
    },
    rules: {
      // your original disables
      'no-alert': 'off',
      camelcase: 'off',
      'no-console': 'off',
      'no-param-reassign': 'off',
      'default-param-last': 'off',
      'no-underscore-dangle': 'off',
      'no-use-before-define': 'off',
      'no-restricted-exports': 'off',
      'no-promise-executor-return': 'off',

      'react/no-children-prop': 'off',
      'react/forbid-prop-types': 'off',
      'react/react-in-jsx-scope': 'off',
      'react/no-array-index-key': 'off',
      'react/require-default-props': 'off',
      'react/jsx-filename-extension': 'off',
      'react/jsx-props-no-spreading': 'off',
      'react/function-component-definition': 'off',

      'jsx-a11y/anchor-is-valid': 'off',
      'jsx-a11y/control-has-associated-label': 'off',

      'import/prefer-default-export': 'off',

      // your warns
      'react/jsx-no-useless-fragment': ['warn', { allowExpressions: true }],
      'prefer-destructuring': ['warn', { object: true, array: false }],
      'react/no-unstable-nested-components': ['warn', { allowAsProps: true }],
      'no-unused-vars': ['warn', { args: 'none' }],
      'react/jsx-no-duplicate-props': ['warn', { ignoreCase: false }],

      // unused-imports
      'unused-imports/no-unused-imports': 'warn',
      'unused-imports/no-unused-vars': [
        'off',
        {
          vars: 'all',
          varsIgnorePattern: '^_',
          args: 'after-used',
          argsIgnorePattern: '^_',
        },
      ],

      // perfectionist
      'perfectionist/sort-named-imports': ['warn', { order: 'asc', type: 'line-length' }],
      'perfectionist/sort-named-exports': ['warn', { order: 'asc', type: 'line-length' }],
      'perfectionist/sort-exports': ['warn', { order: 'asc', type: 'line-length' }],
      'perfectionist/sort-imports': [
        'warn',
        {
          order: 'asc',
          type: 'line-length',
          newlinesBetween: 1,
          groups: [
            ['builtin', 'external'],
            'custom-mui',
            'custom-routes',
            'custom-hooks',
            'custom-utils',
            'internal',
            'custom-components',
            'custom-sections',
            'custom-types',
            ['parent', 'sibling', 'index'],
            'unknown',
          ],
          customGroups: [
            { groupName: 'custom-mui', elementNamePattern: '^@mui/' },
            { groupName: 'custom-routes', elementNamePattern: '^src/routes/' },
            { groupName: 'custom-hooks', elementNamePattern: '^src/hooks/' },
            { groupName: 'custom-utils', elementNamePattern: '^src/utils/' },
            { groupName: 'custom-components', elementNamePattern: '^src/components/' },
            { groupName: 'custom-sections', elementNamePattern: '^src/sections/' },
            { groupName: 'custom-types', elementNamePattern: '^src/types/' },
          ],
          internalPattern: ['^src/'],
        },
      ],

      // prettier plugin present in your old config, keep behavior explicit
      'prettier/prettier': 'error',

      // airbnb/hooks equivalent core behavior: enforce hooks rules
      ...reactHooks.configs.recommended.rules,

      // new rule surfaced by newer react-hooks versions; your codebase uses this pattern heavily
      'react-hooks/set-state-in-effect': 'off',
      'react-hooks/preserve-manual-memoization': 'off',
      // common vite/react-refresh rule (keep it mild)
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
      'no-constant-binary-expression': 'off',
    },
  },

  // your TS override block
  {
    files: ['**/*.{ts,tsx}'],
    rules: {
      'no-unused-vars': 'off',
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          args: 'none',
          vars: 'all',
          varsIgnorePattern: '^_',
          caughtErrors: 'all',
          caughtErrorsIgnorePattern: '^_',
        },
      ],
      '@typescript-eslint/no-explicit-any': 'error',

      // these empty marker interfaces exist in your codebase; keep them allowed for now
      '@typescript-eslint/no-empty-object-type': 'off',
    },
  },

  // JS/JSX: do not apply TS unused-vars to plain JS files
  {
    files: ['**/*.{js,jsx,cjs,mjs}'],
    rules: {
      '@typescript-eslint/no-unused-vars': 'off',
    },
  },
  // last
  prettierConfig,
];
