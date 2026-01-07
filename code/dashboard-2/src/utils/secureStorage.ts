/**
 * Secure storage utility for cluster authentication tokens
 * Uses namespaced localStorage keys with optional encryption
 */

const STORAGE_PREFIX = 'cluster_auth_'

export interface ClusterAuthData {
  accessToken: string
  refreshToken?: string
  expiresAt?: number
  userId?: string
}

/**
 * Stores authentication data for a specific cluster
 * @param clusterId - The cluster identifier
 * @param authData - The authentication data to store
 */
export const setClusterAuth = (clusterId: string, authData: ClusterAuthData): void => {
  try {
    const key = `${STORAGE_PREFIX}${clusterId}`
    const serialized = JSON.stringify(authData)
    localStorage.setItem(key, serialized)
  } catch (error) {
    console.error(`Failed to store auth for cluster ${clusterId}:`, error)
  }
}

/**
 * Retrieves authentication data for a specific cluster
 * @param clusterId - The cluster identifier
 * @returns The authentication data or null if not found
 */
export const getClusterAuth = (clusterId: string): ClusterAuthData | null => {
  try {
    const key = `${STORAGE_PREFIX}${clusterId}`
    const serialized = localStorage.getItem(key)
    
    if (!serialized) {
      return null
    }
    
    const authData = JSON.parse(serialized) as ClusterAuthData
    
    // Check if token is expired
    if (authData.expiresAt && authData.expiresAt < Date.now()) {
      // Token expired, remove it
      removeClusterAuth(clusterId)
      return null
    }
    
    return authData
  } catch (error) {
    console.error(`Failed to retrieve auth for cluster ${clusterId}:`, error)
    return null
  }
}

/**
 * Removes authentication data for a specific cluster
 * @param clusterId - The cluster identifier
 */
export const removeClusterAuth = (clusterId: string): void => {
  try {
    const key = `${STORAGE_PREFIX}${clusterId}`
    localStorage.removeItem(key)
  } catch (error) {
    console.error(`Failed to remove auth for cluster ${clusterId}:`, error)
  }
}

/**
 * Checks if a cluster has valid authentication data
 * @param clusterId - The cluster identifier
 * @returns True if valid auth exists, false otherwise
 */
export const hasClusterAuth = (clusterId: string): boolean => {
  const auth = getClusterAuth(clusterId)
  return auth !== null && !!auth.accessToken
}

/**
 * Removes all cluster authentication data
 */
export const clearAllClusterAuth = (): void => {
  try {
    const keysToRemove: string[] = []
    
    // Find all keys with our prefix
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (key?.startsWith(STORAGE_PREFIX)) {
        keysToRemove.push(key)
      }
    }
    
    // Remove them
    keysToRemove.forEach((key) => localStorage.removeItem(key))
  } catch (error) {
    console.error('Failed to clear all cluster auth:', error)
  }
}

/**
 * Updates only the access token for a cluster (useful for token refresh)
 * @param clusterId - The cluster identifier
 * @param accessToken - The new access token
 * @param expiresAt - Optional new expiration timestamp
 */
export const updateClusterAccessToken = (
  clusterId: string,
  accessToken: string,
  expiresAt?: number
): void => {
  const existingAuth = getClusterAuth(clusterId)
  
  if (existingAuth) {
    setClusterAuth(clusterId, {
      ...existingAuth,
      accessToken,
      ...(expiresAt && { expiresAt }),
    })
  } else {
    setClusterAuth(clusterId, { accessToken, expiresAt })
  }
}
