import { defineConfig } from '@hey-api/openapi-ts'

export default defineConfig({
  input: 'api.json',
  output: {
    lint: 'eslint',
    path: 'src/client',
  },
  plugins: [
    {
      dates: true,
      name: '@hey-api/transformers',
    },
    {
      name: '@hey-api/sdk',
      // Use operationId to drive function names
      operationId: false,
    },
    '@hey-api/client-axios',
    {
      name: '@tanstack/react-query',
      queryOptions: true,
    }
  ]
})
