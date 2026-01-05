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
import { AllCommunityModule, ModuleRegistry } from 'ag-grid-community'

import App from './App.tsx'
import { system } from './theme.ts'
import { ColorModeProvider } from './components/ui/color-mode.tsx'
import { client } from './client/client.gen.ts'
import { EX3_API_ENDPOINT, UIO_API_ENDPOINT, } from './Constants.ts'

ModuleRegistry.registerModules([AllCommunityModule])

const queryClient = new QueryClient()

client.setConfig({
  baseURL: UIO_API_ENDPOINT,
})

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
