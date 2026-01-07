import { createContext, useContext, useEffect, useState, useCallback, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { getClusterConfig } from '../config/clusters'
import { clearClusterClient } from '../utils/clusterApiClients'
import { removeClusterAuth } from '../utils/secureStorage'

const STORAGE_KEY = 'user_selected_clusters'

export interface ClusterContextValue {
  selectedClusters: string[]
  currentCluster: string | null
  addCluster: (clusterId: string) => void
  removeCluster: (clusterId: string) => void
  switchCluster: (clusterId: string) => void
  reorderClusters: (newOrder: string[]) => void
  hasSelectedClusters: boolean
}

const ClusterContext = createContext<ClusterContextValue | undefined>(undefined)

export interface ClusterProviderProps {
  children: ReactNode
}

export const ClusterProvider = ({ children }: ClusterProviderProps) => {
  const queryClient = useQueryClient()
  const [selectedClusters, setSelectedClusters] = useState<string[]>(() => {
    // Initialize from localStorage immediately
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        return JSON.parse(stored) as string[]
      }
    } catch (error) {
      console.error('Failed to load selected clusters:', error)
    }
    return []
  })
  const [currentCluster, setCurrentCluster] = useState<string | null>(() => {
    // Set first cluster as current on initialization
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        const clusters = JSON.parse(stored) as string[]
        return clusters.length > 0 ? clusters[0] : null
      }
    } catch (error) {
      console.error('Failed to load current cluster:', error)
    }
    return null
  })

  // Save selected clusters to localStorage whenever they change
  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(selectedClusters))
    } catch (error) {
      console.error('Failed to save selected clusters:', error)
    }
  }, [selectedClusters])

  const addCluster = useCallback((clusterId: string) => {
    const config = getClusterConfig(clusterId)
    if (!config) {
      console.error(`Cluster config not found for: ${clusterId}`)
      return
    }

    setSelectedClusters((prev) => {
      if (prev.includes(clusterId)) {
        return prev
      }
      return [...prev, clusterId]
    })

    // Set as current cluster if it's the first one
    setCurrentCluster((prev) => prev || clusterId)
  }, [])

  const removeCluster = useCallback(
    (clusterId: string) => {
      setSelectedClusters((prev) => prev.filter((id) => id !== clusterId))

      // Clear React Query cache for this cluster
      const config = getClusterConfig(clusterId)
      if (config) {
        queryClient.removeQueries({
          predicate: (query) => {
            const queryKey = query.queryKey
            // Check if queryKey contains the cluster's baseURL
            return JSON.stringify(queryKey).includes(config.apiBaseUrl)
          },
        })
      }

      // Clear API client cache
      clearClusterClient(clusterId)

      // Clear auth data
      removeClusterAuth(clusterId)

      // Switch to another cluster if we removed the current one
      if (currentCluster === clusterId) {
        setCurrentCluster(() => {
          const remaining = selectedClusters.filter((id) => id !== clusterId)
          return remaining.length > 0 ? remaining[0] : null
        })
      }
    },
    [currentCluster, queryClient, selectedClusters]
  )

  const switchCluster = useCallback((clusterId: string) => {
    if (!selectedClusters.includes(clusterId)) {
      console.error(`Cannot switch to non-selected cluster: ${clusterId}`)
      return
    }
    setCurrentCluster(clusterId)
  }, [selectedClusters])

  const reorderClusters = useCallback((newOrder: string[]) => {
    // Validate that newOrder contains the same clusters
    if (newOrder.length !== selectedClusters.length) {
      console.error('Invalid cluster order: length mismatch')
      return
    }
    
    const isValid = newOrder.every(id => selectedClusters.includes(id))
    if (!isValid) {
      console.error('Invalid cluster order: contains unknown cluster IDs')
      return
    }
    
    setSelectedClusters(newOrder)
  }, [selectedClusters])

  const value: ClusterContextValue = {
    selectedClusters,
    currentCluster,
    addCluster,
    removeCluster,
    switchCluster,
    reorderClusters,
    hasSelectedClusters: selectedClusters.length > 0,
  }

  return <ClusterContext.Provider value={value}>{children}</ClusterContext.Provider>
}

export const useClusterContext = () => {
  const context = useContext(ClusterContext)
  if (context === undefined) {
    throw new Error('useClusterContext must be used within a ClusterProvider')
  }
  return context
}
