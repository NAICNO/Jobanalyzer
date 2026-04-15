import { UserManager } from 'oidc-client-ts'

// This script runs in a hidden iframe to handle silent token renewal.
// It uses the same bundled oidc-client-ts as the main app.
new UserManager({ response_mode: 'query' })
  .signinSilentCallback()
  .catch((err) => {
    console.error('Silent renew error:', err)
  })
