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
import { ClusterProvider } from './contexts/ClusterContext.tsx'
import { AuthProvider } from './contexts/AuthContext.tsx'
import { Toaster } from './components/ui/toaster.tsx'
import { loadClusterConfig } from './config/clusters.ts'

ModuleRegistry.registerModules([AllCommunityModule])

const queryClient = new QueryClient()

loadClusterConfig()
  .then(() => {
    ReactDOM.createRoot(document.getElementById('root')!).render(
      <React.StrictMode>
        <ChakraProvider value={system}>
          <ColorModeProvider>
            <QueryClientProvider client={queryClient}>
              <AuthProvider>
                <ClusterProvider>
                  <App/>
                  <ReactQueryDevtools initialIsOpen={false}/>
                  <Toaster />
                </ClusterProvider>
              </AuthProvider>
            </QueryClientProvider>
          </ColorModeProvider>
        </ChakraProvider>
      </React.StrictMode>,
    )
  })
  .catch((error) => {
    console.error('Fatal: Failed to load cluster configuration', error)
    const root = document.getElementById('root')!
    root.innerHTML = `
      <div style="display:flex;flex-direction:column;align-items:center;justify-content:center;min-height:100vh;font-family:system-ui,sans-serif;color:#333">
        <h1 style="margin-bottom:8px">Configuration Error</h1>
        <p>Failed to load cluster configuration.</p>
        <p style="color:#666;font-size:14px">${error instanceof Error ? error.message : String(error)}</p>
        <button onclick="window.location.reload()" style="margin-top:16px;padding:8px 24px;cursor:pointer;border:1px solid #ccc;border-radius:4px;background:#fff">
          Retry
        </button>
      </div>
    `
  })
