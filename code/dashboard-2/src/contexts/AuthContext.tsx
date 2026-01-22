import { createContext, useEffect, useState, useCallback, type ReactNode } from 'react'
import type { User } from 'oidc-client-ts'
import {
  getUserManager,
  loginToCluster,
  logoutFromCluster,
  silentLogoutFromCluster,
  getClusterUser,
  clearUserManager,
} from '../utils/oidcManager'
import { setClusterAuth, removeClusterAuth } from '../utils/secureStorage'
import { toaster } from '../components/ui/toaster'

export interface AuthState {
  [clusterId: string]: {
    user: User | null
    isAuthenticated: boolean
    isLoading: boolean
    error: string | null
  }
}

export interface AuthContextValue {
  authState: AuthState
  login: (clusterId: string, returnPath?: string) => Promise<void>
  logout: (clusterId: string) => Promise<void>
  silentLogout: (clusterId: string) => Promise<void>
  getUser: (clusterId: string) => Promise<User | null>
  isAuthenticated: (clusterId: string) => boolean
  refreshAuthState: (clusterId: string) => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export interface AuthProviderProps {
  children: ReactNode
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [authState, setAuthState] = useState<AuthState>({})

  /**
   * Updates the auth state for a specific cluster
   */
  const updateClusterAuthState = useCallback(
    (clusterId: string, updates: Partial<AuthState[string]>) => {
      setAuthState((prev) => ({
        ...prev,
        [clusterId]: {
          ...(prev[clusterId] || {
            user: null,
            isAuthenticated: false,
            isLoading: false,
            error: null,
          }),
          ...updates,
        },
      }))
    },
    []
  )

  /**
   * Syncs UserManager state with secureStorage
   */
  const syncTokenStorage = useCallback(async (clusterId: string, user: User | null) => {
    if (user && !user.expired) {
      // Store tokens in secureStorage
      await setClusterAuth(clusterId, {
        accessToken: user.access_token,
        refreshToken: user.refresh_token,
        expiresAt: user.expires_at,
        userId: user.profile.sub,
      })
    } else {
      // Clear tokens from secureStorage
      removeClusterAuth(clusterId)
    }
  }, [])

  /**
   * Refreshes the authentication state for a cluster
   */
  const refreshAuthState = useCallback(
    async (clusterId: string) => {
      updateClusterAuthState(clusterId, { isLoading: true, error: null })

      try {
        const user = await getClusterUser(clusterId)
        const authenticated = user !== null && !user.expired

        updateClusterAuthState(clusterId, {
          user,
          isAuthenticated: authenticated,
          isLoading: false,
        })

        // Sync with secureStorage
        await syncTokenStorage(clusterId, user)
      } catch (error) {
        console.error('Failed to refresh auth state:', error)
        updateClusterAuthState(clusterId, {
          user: null,
          isAuthenticated: false,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Unknown error',
        })
      }
    },
    [updateClusterAuthState, syncTokenStorage]
  )

  /**
   * Sets up UserManager event listeners for a cluster
   */
  const setupEventListeners = useCallback(
    (clusterId: string) => {
      const userManager = getUserManager(clusterId)
      if (!userManager) return

      // Note: oidc-client-ts event removal requires the original handler function
      // Since we're using inline handlers, we can't remove them individually
      // Instead, we rely on the UserManager cache to reuse the same instance

      // User loaded event
      userManager.events.addUserLoaded(async (user) => {
        console.log(`User loaded for cluster ${clusterId}`)
        updateClusterAuthState(clusterId, {
          user,
          isAuthenticated: true,
          isLoading: false,
        })
        await syncTokenStorage(clusterId, user)
      })

      // User unloaded event
      userManager.events.addUserUnloaded(() => {
        console.log(`User unloaded for cluster ${clusterId}`)
        updateClusterAuthState(clusterId, {
          user: null,
          isAuthenticated: false,
        })
        removeClusterAuth(clusterId)
      })

      // Access token expiring (5 minutes before expiry)
      userManager.events.addAccessTokenExpiring(() => {
        console.log(`Access token expiring for cluster ${clusterId}`)
        toaster.create({
          title: 'Session expiring soon',
          description: 'Your session will be refreshed automatically',
          type: 'info',
          duration: 3000,
        })
      })

      // Access token expired
      userManager.events.addAccessTokenExpired(() => {
        console.log(`Access token expired for cluster ${clusterId}`)
        updateClusterAuthState(clusterId, {
          user: null,
          isAuthenticated: false,
        })
        removeClusterAuth(clusterId)
        
        toaster.create({
          title: 'Session expired',
          description: 'Please log in again',
          type: 'warning',
          duration: 5000,
        })
      })

      // Silent renew error - don't show toast for every error to avoid spam
      userManager.events.addSilentRenewError((error) => {
        console.warn(`Silent renew error for cluster ${clusterId}:`, error)
        // Only remove user if the error is critical (not a network timeout)
        // This prevents excessive re-authentication attempts
        if (error.message && !error.message.includes('timeout')) {
          updateClusterAuthState(clusterId, {
            user: null,
            isAuthenticated: false,
          })
          removeClusterAuth(clusterId)
        }
      })

      // User signed out
      userManager.events.addUserSignedOut(() => {
        console.log(`User signed out for cluster ${clusterId}`)
        updateClusterAuthState(clusterId, {
          user: null,
          isAuthenticated: false,
        })
        removeClusterAuth(clusterId)
      })
    },
    [updateClusterAuthState, syncTokenStorage]
  )

  /**
   * Login to a cluster
   */
  const login = useCallback(
    async (clusterId: string, returnPath?: string) => {
      updateClusterAuthState(clusterId, { isLoading: true, error: null })

      try {
        // Setup event listeners before login
        setupEventListeners(clusterId)
        
        // Initiate login flow
        await loginToCluster(clusterId, returnPath)
      } catch (error) {
        console.error('Login failed:', error)
        const errorMessage = error instanceof Error ? error.message : 'Unknown error'
        
        updateClusterAuthState(clusterId, {
          isLoading: false,
          error: errorMessage,
        })

        toaster.create({
          title: 'Authentication failed',
          description: errorMessage,
          type: 'error',
          duration: 5000,
        })
        
        throw error
      }
    },
    [updateClusterAuthState, setupEventListeners]
  )

  /**
   * Logout from a cluster
   */
  const logout = useCallback(
    async (clusterId: string) => {
      updateClusterAuthState(clusterId, { isLoading: true })

      try {
        await logoutFromCluster(clusterId)
        removeClusterAuth(clusterId)
        clearUserManager(clusterId)
        
        updateClusterAuthState(clusterId, {
          user: null,
          isAuthenticated: false,
          isLoading: false,
        })

        toaster.create({
          title: 'Logged out',
          description: 'Successfully logged out from cluster',
          type: 'success',
          duration: 3000,
        })
      } catch (error) {
        console.error('Logout failed:', error)
        
        // Even if logout fails, clear local state
        removeClusterAuth(clusterId)
        clearUserManager(clusterId)
        updateClusterAuthState(clusterId, {
          user: null,
          isAuthenticated: false,
          isLoading: false,
        })

        toaster.create({
          title: 'Logout error',
          description: 'Session cleared locally',
          type: 'warning',
          duration: 3000,
        })
      }
    },
    [updateClusterAuthState]
  )

  /**
   * Silent logout from a cluster (no redirect, no toast)
   * Used when removing a cluster from selection
   */
  const silentLogout = useCallback(
    async (clusterId: string) => {
      try {
        await silentLogoutFromCluster(clusterId)
        removeClusterAuth(clusterId)
        clearUserManager(clusterId)
        
        updateClusterAuthState(clusterId, {
          user: null,
          isAuthenticated: false,
          isLoading: false,
        })
      } catch (error) {
        console.error('Silent logout failed:', error)
        
        // Even if logout fails, clear local state
        removeClusterAuth(clusterId)
        clearUserManager(clusterId)
        updateClusterAuthState(clusterId, {
          user: null,
          isAuthenticated: false,
          isLoading: false,
        })
      }
    },
    [updateClusterAuthState]
  )

  /**
   * Get user for a cluster
   */
  const getUser = useCallback(
    async (clusterId: string): Promise<User | null> => {
      return await getClusterUser(clusterId)
    },
    []
  )

  /**
   * Check if user is authenticated for a cluster
   */
  const isAuthenticated = useCallback(
    (clusterId: string): boolean => {
      return authState[clusterId]?.isAuthenticated || false
    },
    [authState]
  )

  // Initialize auth state from existing sessions on mount
  useEffect(() => {
    const initializeAuthState = async () => {
      // Check for existing sessions in secureStorage
      const storedSessions = Object.keys(localStorage)
        .filter((key) => key.startsWith('cluster_auth_'))
        .map((key) => key.replace('cluster_auth_', ''))

      for (const clusterId of storedSessions) {
        setupEventListeners(clusterId)
        await refreshAuthState(clusterId)
      }
    }

    initializeAuthState()
  }, [setupEventListeners, refreshAuthState])

  const value: AuthContextValue = {
    authState,
    login,
    logout,
    silentLogout,
    getUser,
    isAuthenticated,
    refreshAuthState,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
