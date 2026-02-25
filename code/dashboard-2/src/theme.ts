import { createSystem, defaultConfig, defineConfig } from '@chakra-ui/react'

const config = defineConfig({
  globalCss: {
    'html, body': {
      overflow: 'hidden',
      height: '100vh',
      margin: 0,
    },
  },
  theme: {
    tokens: {
      fonts: {
        heading: {
          value: '\'IBM Plex Sans\', sans-serif',
        },
        body: {
          value: '\'IBM Plex Sans\', sans-serif',
        },
      }
    },
  }
})

export const system = createSystem(defaultConfig, config)
