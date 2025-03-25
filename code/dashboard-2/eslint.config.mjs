import { defineConfig } from 'eslint/config'
import typescriptEslint from '@typescript-eslint/eslint-plugin'
import react from 'eslint-plugin-react'
import globals from 'globals'
import tsParser from '@typescript-eslint/parser'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import js from '@eslint/js'
import { FlatCompat } from '@eslint/eslintrc'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const compat = new FlatCompat({
  baseDirectory: __dirname,
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
})

export default defineConfig({
  ignores: ['dist'],

  extends: compat.extends(
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:react/recommended',
    'plugin:@tanstack/eslint-plugin-query/recommended',
  ),

  plugins: {
    '@typescript-eslint': typescriptEslint,
    react,
  },

  languageOptions: {
    globals: {
      ...globals.browser,
    },
    parser: tsParser,
    ecmaVersion: 'latest',
    sourceType: 'module',
  },

  settings: {
    react: {
      version: 'detect',
    },
  },

  rules: {
    indent: ['error', 2],
    'linebreak-style': ['error', 'unix'],
    quotes: ['error', 'single'],
    semi: ['error', 'never'],
    'react/react-in-jsx-scope': 'off',
    'react/jsx-uses-react': 'off',
  },
})
