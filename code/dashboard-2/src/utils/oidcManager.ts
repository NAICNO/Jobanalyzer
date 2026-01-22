import { UserManager, WebStorageStateStore, type UserManagerSettings } from 'oidc-client-ts'
import { getClusterConfig } from '../config/clusters'

// Cache UserManager instances per cluster
const userManagers = new Map<string, UserManager>()

/**
 * Creates a UserManager instance for a specific cluster with PKCE enabled
 * @param clusterId - The cluster identifier
 * @returns Configured UserManager instance
 */
export const createUserManager = (clusterId: string): UserManager | null => {
  const config = getClusterConfig(clusterId)
  
  if (!config) {
    console.error(`Cluster config not found for: ${clusterId}`)
    return null
  }

  const settings: UserManagerSettings = {
    authority: config.authEndpoint.authorization.replace(/\/protocol\/openid-connect\/auth$/, ''),
    client_id: config.authEndpoint.clientId,
    redirect_uri: config.authEndpoint.redirectUri,
    silent_redirect_uri: `${window.location.origin}/silent-renew.html`,
    response_type: 'code',
    scope: config.authEndpoint.scope || 'openid profile email',
    
    // PKCE settings (Proof Key for Code Exchange) - essential for browser-based apps
    // PKCE is automatically enabled when response_type is 'code'
    response_mode: 'query',
    
    // Token endpoints
    metadata: {
      authorization_endpoint: config.authEndpoint.authorization,
      token_endpoint: config.authEndpoint.token,
      userinfo_endpoint: config.authEndpoint.userInfo,
      end_session_endpoint: config.authEndpoint.authorization.replace('/auth', '/logout'),
      issuer: config.authEndpoint.authorization.replace(/\/protocol\/openid-connect\/auth$/, ''),
    },
    
    // Storage settings - use WebStorageStateStore wrapper
    userStore: new WebStorageStateStore({ store: window.localStorage }),
    stateStore: new WebStorageStateStore({ store: window.sessionStorage }),
    
    // Enable automatic silent renew with proper configuration
    automaticSilentRenew: true,
    silentRequestTimeoutInSeconds: 10,
    accessTokenExpiringNotificationTimeInSeconds: 60, // Start renewal 60 seconds before expiry
    
    // Additional settings
    loadUserInfo: false, // Don't load user info on silent renewal to reduce requests
    filterProtocolClaims: true,
    includeIdTokenInSilentRenew: true,
    monitorSession: false, // Disable session monitoring to prevent unnecessary checks
  }

  const userManager = new UserManager(settings)
  
  // Cache the instance
  userManagers.set(clusterId, userManager)
  
  return userManager
}

/**
 * Gets a cached UserManager instance or creates a new one
 * @param clusterId - The cluster identifier
 * @returns UserManager instance for the cluster
 */
export const getUserManager = (clusterId: string): UserManager | null => {
  const cached = userManagers.get(clusterId)
  if (cached) {
    return cached
  }
  
  return createUserManager(clusterId)
}

/**
 * Clears the cached UserManager for a cluster
 * @param clusterId - The cluster identifier
 */
export const clearUserManager = (clusterId: string): void => {
  userManagers.delete(clusterId)
}

/**
 * Initiates the OIDC login flow for a cluster
 * @param clusterId - The cluster identifier
 * @param returnPath - Optional path to return to after authentication
 */
export const loginToCluster = async (clusterId: string, returnPath?: string): Promise<void> => {
  const userManager = getUserManager(clusterId)
  
  if (!userManager) {
    throw new Error(`Cannot create UserManager for cluster: ${clusterId}`)
  }

  // Store cluster ID and return path for use after callback
  sessionStorage.setItem('auth_redirect_cluster', clusterId)
  if (returnPath) {
    sessionStorage.setItem('auth_redirect_path', returnPath)
  }

  try {
    // Initiate the authorization request with PKCE
    await userManager.signinRedirect({
      state: JSON.stringify({ clusterId, returnPath }),
    })
  } catch (error) {
    console.error('Failed to initiate OIDC login:', error)
    throw error
  }
}

/**
 * Handles the OIDC callback after authentication
 * @returns User information and original state
 */
export const handleAuthCallback = async () => {
  const clusterId = sessionStorage.getItem('auth_redirect_cluster')
  
  if (!clusterId) {
    throw new Error('No cluster ID found in session storage')
  }

  const userManager = getUserManager(clusterId)
  
  if (!userManager) {
    throw new Error(`Cannot create UserManager for cluster: ${clusterId}`)
  }

  try {
    // Process the callback and get the user
    const user = await userManager.signinRedirectCallback()
    
    // Return user and stored state
    return {
      user,
      clusterId,
      returnPath: sessionStorage.getItem('auth_redirect_path') || `/v2/${clusterId}/overview`,
    }
  } catch (error) {
    console.error('Failed to handle auth callback:', error)
    throw error
  } finally {
    // Clean up session storage
    sessionStorage.removeItem('auth_redirect_cluster')
    sessionStorage.removeItem('auth_redirect_path')
  }
}

/**
 * Logs out from a cluster
 * @param clusterId - The cluster identifier
 */
export const logoutFromCluster = async (clusterId: string): Promise<void> => {
  const userManager = getUserManager(clusterId)
  
  if (!userManager) {
    console.warn(`No UserManager found for cluster: ${clusterId}`)
    return
  }

  try {
    await userManager.signoutRedirect()
  } catch (error) {
    console.error('Failed to logout:', error)
    // Even if remote logout fails, remove local session
    await userManager.removeUser()
  }
}

/**
 * Silently logs out from a cluster without redirect
 * Terminates the Keycloak SSO session via iframe and clears local storage
 * @param clusterId - The cluster identifier
 */
export const silentLogoutFromCluster = async (clusterId: string): Promise<void> => {
  const userManager = getUserManager(clusterId)
  
  if (!userManager) {
    console.warn(`No UserManager found for cluster: ${clusterId}`)
    return
  }

  try {
    const user = await userManager.getUser()
    
    // If user exists, terminate SSO session on Keycloak
    if (user && user.id_token) {
      const settings = await userManager.settings
      const endSessionEndpoint = settings.metadata?.end_session_endpoint
      
      if (endSessionEndpoint) {
        // Create logout URL with id_token_hint
        const logoutUrl = new URL(endSessionEndpoint)
        logoutUrl.searchParams.append('id_token_hint', user.id_token)
        
        // Call logout endpoint via hidden iframe to avoid redirect
        await new Promise<void>((resolve, reject) => {
          const iframe = document.createElement('iframe')
          iframe.style.display = 'none'
          iframe.src = logoutUrl.toString()
          
          const timeout = setTimeout(() => {
            document.body.removeChild(iframe)
            reject(new Error('Logout timeout'))
          }, 5000)
          
          iframe.onload = () => {
            clearTimeout(timeout)
            document.body.removeChild(iframe)
            resolve()
          }
          
          iframe.onerror = () => {
            clearTimeout(timeout)
            document.body.removeChild(iframe)
            reject(new Error('Logout failed'))
          }
          
          document.body.appendChild(iframe)
        })
      }
    }
    
    // Remove user from local storage
    await userManager.removeUser()
  } catch (error) {
    console.error('Failed to perform silent logout:', error)
    // Even if server-side logout fails, remove local session
    await userManager.removeUser()
  }
}

/**
 * Gets the current user for a cluster
 * @param clusterId - The cluster identifier
 * @returns User object or null if not authenticated
 */
export const getClusterUser = async (clusterId: string) => {
  const userManager = getUserManager(clusterId)
  
  if (!userManager) {
    return null
  }

  try {
    return await userManager.getUser()
  } catch (error) {
    console.error('Failed to get user:', error)
    return null
  }
}

/**
 * Checks if a user is authenticated for a cluster
 * @param clusterId - The cluster identifier
 * @returns True if authenticated, false otherwise
 */
export const isClusterAuthenticated = async (clusterId: string): Promise<boolean> => {
  const user = await getClusterUser(clusterId)
  return user !== null && !user.expired
}
