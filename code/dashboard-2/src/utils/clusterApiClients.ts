import { createClient, createConfig, type ClientOptions } from '../client/client/index'
import type { Client } from '../client/client/types.gen'

// Cache for cluster-specific API clients
const clusterClients = new Map<string, Client>()

/**
 * Creates a new API client instance for a specific cluster
 * @param clusterId - The cluster identifier
 * @param apiBaseUrl - The base URL for the cluster's API
 * @returns A configured API client instance
 */
export const createClusterClient = (
  clusterId: string,
  apiBaseUrl: string
): Client => {
  const client = createClient(createConfig<ClientOptions>({ baseURL: apiBaseUrl }))
  
  // Store in cache
  clusterClients.set(clusterId, client)
  
  return client
}

/**
 * Gets the cached API client for a cluster, or creates a new one if it doesn't exist
 * @param clusterId - The cluster identifier
 * @param apiBaseUrl - The base URL for the cluster's API
 * @returns The API client instance for the cluster
 */
export const getClusterClient = (
  clusterId: string,
  apiBaseUrl: string
): Client => {
  const cachedClient = clusterClients.get(clusterId)
  
  if (cachedClient) {
    // Update baseURL in case it changed
    cachedClient.setConfig({ baseURL: apiBaseUrl })
    return cachedClient
  }
  
  return createClusterClient(clusterId, apiBaseUrl)
}

/**
 * Clears the cached client for a specific cluster
 * Useful when removing a cluster or when auth tokens change
 * @param clusterId - The cluster identifier
 */
export const clearClusterClient = (clusterId: string): void => {
  clusterClients.delete(clusterId)
}

/**
 * Clears all cached cluster clients
 */
export const clearAllClusterClients = (): void => {
  clusterClients.clear()
}

/**
 * Updates the auth token for a cluster's client
 * @param clusterId - The cluster identifier
 * @param token - The auth token to set
 */
export const setClusterClientAuth = (clusterId: string, token: string): void => {
  const client = clusterClients.get(clusterId)
  if (client) {
    client.setConfig({
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
  }
}
