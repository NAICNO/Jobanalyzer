import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { ChakraProvider } from '@chakra-ui/react'
import '@fontsource/ibm-plex-sans/300.css'
import '@fontsource/ibm-plex-sans/400.css'
import '@fontsource/ibm-plex-sans/500.css'
import '@fontsource/ibm-plex-sans/700.css'
import '../index.css'

import App from './App.tsx'
import { system } from './theme.ts'
import { ColorModeProvider } from './components/ui/color-mode.tsx'

const queryClient = new QueryClient()

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ChakraProvider value={system}>
      <ColorModeProvider>
        <QueryClientProvider client={queryClient}>
          <App/>
          <ReactQueryDevtools initialIsOpen={false}/>
        </QueryClientProvider>
      </ColorModeProvider>
    </ChakraProvider>
  </React.StrictMode>,
)
