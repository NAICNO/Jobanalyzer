import { useClusterContext } from '../contexts/ClusterContext'
import { getClusterConfig } from '../config/clusters'
import type { ClusterConfig } from '../config/clusters'

export interface UseClusterReturn {
  selectedClusters: string[]
  currentCluster: string | null
  currentClusterConfig: ClusterConfig | undefined
  addCluster: (clusterId: string) => void
  removeCluster: (clusterId: string) => Promise<void>
  switchCluster: (clusterId: string) => void
  reorderClusters: (newOrder: string[]) => void
  hasSelectedClusters: boolean
  getClusterConfigById: (clusterId: string) => ClusterConfig | undefined
}

/**
 * Hook to access cluster management functionality
 * Must be used within a ClusterProvider
 */
export const useCluster = (): UseClusterReturn => {
  const context = useClusterContext()
  
  const currentClusterConfig = context.currentCluster
    ? getClusterConfig(context.currentCluster)
    : undefined

  return {
    ...context,
    currentClusterConfig,
    getClusterConfigById: getClusterConfig,
  }
}
