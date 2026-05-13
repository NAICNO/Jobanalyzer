/**
 * MSW browser worker setup for demo mode.
 *
 * Auto-selects the demo cluster in localStorage so the app opens
 * directly to the cluster overview instead of the selection page.
 */
import { setupWorker } from 'msw/browser'
import { handlers } from './handlers'

// Pre-select the demo cluster so ClusterProvider picks it up on init
localStorage.setItem(
  'user_selected_clusters',
  JSON.stringify(['demo.hpc.example.org']),
)

export const worker = setupWorker(...handlers)
