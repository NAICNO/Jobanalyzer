import { createContext, useContext, useState, useMemo, useCallback, type ReactNode } from 'react'
import type { UseQueryResult } from '@tanstack/react-query'

import { useClusterClient } from '../hooks/useClusterClient'
import { useClusterNodes, useClusterNodesInfo, useClusterNodesStates, useClusterNodesProcessGpuUtil } from '../hooks/v2/useNodeQueries'
import { useClusterPartitions } from '../hooks/v2/useClusterQueries'
import type { TimeRange } from '../components/TimeRangePicker'
import { timeRangeToTimestamps } from '../util/timeRangeUtils'

const DEFAULT_TIME_RANGE: TimeRange = {
  type: 'relative',
  label: 'Last 1 day',
  value: '1d',
  endAt: 'now',
  refreshIntervalSec: 0,
}

export interface ClusterOverviewContextValue {
  cluster: string
  // Shared query results
  partitionsQuery: UseQueryResult
  nodesQuery: UseQueryResult
  nodesInfoQuery: UseQueryResult
  nodesStatesQuery: UseQueryResult
  gpuUtilQuery: UseQueryResult
  // Time range
  timeRange: TimeRange
  setTimeRange: (range: TimeRange) => void
  startTimeInS: number
  endTimeInS: number
  // Refresh
  refetchAll: () => void
  isFetching: boolean
  oldestDataUpdatedAt: number
}

const ClusterOverviewContext = createContext<ClusterOverviewContextValue | undefined>(undefined)

interface ClusterOverviewProviderProps {
  cluster: string
  children: ReactNode
}

export const ClusterOverviewProvider = ({ cluster, children }: ClusterOverviewProviderProps) => {
  const client = useClusterClient(cluster)
  const [timeRange, setTimeRange] = useState<TimeRange>(DEFAULT_TIME_RANGE)

  // Compute timestamps from time range
  const { startTimeInS, endTimeInS } = useMemo(() => {
    const result = timeRangeToTimestamps(timeRange)
    const now = Math.floor(Date.now() / 1000)
    return {
      startTimeInS: result.startTimeInS ?? now - 24 * 60 * 60,
      endTimeInS: result.endTimeInS ?? now,
    }
  }, [timeRange])

  // Shared queries — called once, consumed by all child components
  const partitionsQuery = useClusterPartitions({ cluster, client })
  const nodesQuery = useClusterNodes({ cluster, client })
  const nodesInfoQuery = useClusterNodesInfo({ cluster, client })
  const nodesStatesQuery = useClusterNodesStates({ cluster, client })
  const gpuUtilQuery = useClusterNodesProcessGpuUtil({ cluster, client })

  const refetchAll = useCallback(() => {
    // Only refetch dynamic data — skip nodesQuery (node list) and
    // nodesInfoQuery (hardware specs) as those rarely change.
    partitionsQuery.refetch()  // job counts, GPU in-use
    nodesStatesQuery.refetch() // node IDLE/ALLOC/DOWN states
    gpuUtilQuery.refetch()     // GPU utilization
  }, [partitionsQuery, nodesStatesQuery, gpuUtilQuery])

  const isFetching = partitionsQuery.isFetching
    || nodesStatesQuery.isFetching
    || gpuUtilQuery.isFetching

  // Find the oldest dataUpdatedAt among all queries for staleness indicator
  const oldestDataUpdatedAt = useMemo(() => {
    const timestamps = [
      partitionsQuery.dataUpdatedAt,
      nodesQuery.dataUpdatedAt,
      nodesInfoQuery.dataUpdatedAt,
      nodesStatesQuery.dataUpdatedAt,
      gpuUtilQuery.dataUpdatedAt,
    ].filter(t => t > 0)
    return timestamps.length > 0 ? Math.min(...timestamps) : 0
  }, [
    partitionsQuery.dataUpdatedAt,
    nodesQuery.dataUpdatedAt,
    nodesInfoQuery.dataUpdatedAt,
    nodesStatesQuery.dataUpdatedAt,
    gpuUtilQuery.dataUpdatedAt,
  ])

  const value: ClusterOverviewContextValue = {
    cluster,
    partitionsQuery,
    nodesQuery,
    nodesInfoQuery,
    nodesStatesQuery,
    gpuUtilQuery,
    timeRange,
    setTimeRange,
    startTimeInS,
    endTimeInS,
    refetchAll,
    isFetching,
    oldestDataUpdatedAt,
  }

  return (
    <ClusterOverviewContext.Provider value={value}>
      {children}
    </ClusterOverviewContext.Provider>
  )
}

export const useClusterOverviewContext = () => {
  const context = useContext(ClusterOverviewContext)
  if (context === undefined) {
    throw new Error('useClusterOverviewContext must be used within a ClusterOverviewProvider')
  }
  return context
}
