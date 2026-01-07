import { useMemo } from 'react'
import { useParams } from 'react-router'
import { getClusterClient } from '../utils/clusterApiClients'
import { getClusterConfig, getClusterIdFromFullName } from '../config/clusters'
import { useCluster } from './useCluster'
import type { Client } from '../client/client/types.gen'

/**
 * Hook to get the API client for a specific cluster
 * Can be used with an explicit clusterId or will use the cluster from URL params
 * 
 * @param explicitClusterId - Optional cluster ID to use instead of URL params
 * @returns The API client instance for the cluster
 */
export const useClusterClient = (explicitClusterId?: string): Client | null => {
  const { clusterName } = useParams<{ clusterName: string }>()
  const { currentCluster } = useCluster()

  const client = useMemo(() => {
    // Determine which cluster ID to use
    let clusterId: string | undefined

    if (explicitClusterId) {
      clusterId = explicitClusterId
    } else if (clusterName) {
      // URL cluster name is the same as cluster ID (e.g., 'mlx.hpc.uio.no')
      clusterId = getClusterIdFromFullName(clusterName)
    } else if (currentCluster) {
      clusterId = currentCluster
    }

    if (!clusterId) {
      console.warn('No cluster ID available for useClusterClient')
      return null
    }

    // Get cluster config
    const config = getClusterConfig(clusterId)
    if (!config) {
      console.error(`Cluster config not found for: ${clusterId}`)
      return null
    }

    // Get or create the client
    return getClusterClient(clusterId, config.apiBaseUrl)
  }, [explicitClusterId, clusterName, currentCluster])

  return client
}
