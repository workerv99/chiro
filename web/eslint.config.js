import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';

export default [
  js.configs.recommended,
  ...svelte.configs['flat/recommended'],
  {
    ignores: ['build/', '.svelte-kit/', 'node_modules/']
  },
  {
    files: ['**/*.{js,svelte}'],
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node
      }
    }
  },
  {
    rules: {
      'svelte/no-navigation-without-resolve': 'off'
    }
  }
];
