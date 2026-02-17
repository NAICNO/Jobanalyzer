import { useQuery } from '@tanstack/react-query'
import {
  getClusterByClusterPartitionsOptions,
  getClusterByClusterJobsOptions,
  getClusterByClusterErrorMessagesOptions,
  getClusterByClusterNodesGpuTimeseriesOptions,
  getClusterByClusterNodesCpuTimeseriesOptions,
  getClusterByClusterNodesMemoryTimeseriesOptions,
} from '../../client/@tanstack/react-query.gen'
import type { Client } from '../../client/client/types.gen'

// ---------------------------------------------------------------------------
// Shared interface
// ---------------------------------------------------------------------------

interface ClusterQueryOptions {
  cluster: string
  client: Client | null
  enabled?: boolean
}

// ---------------------------------------------------------------------------
// useClusterPartitions
// ---------------------------------------------------------------------------

export const useClusterPartitions = ({ cluster, client, enabled = true }: ClusterQueryOptions) => {
  return useQuery({
    ...getClusterByClusterPartitionsOptions({
      path: { cluster },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
  })
}

// ---------------------------------------------------------------------------
// useClusterJobs
// ---------------------------------------------------------------------------

interface UseClusterJobsOptions extends ClusterQueryOptions {
  startTimeInS?: number
  endTimeInS?: number
}

export const useClusterJobs = ({ cluster, client, startTimeInS, endTimeInS, enabled = true }: UseClusterJobsOptions) => {
  const query = startTimeInS != null || endTimeInS != null
    ? { start_time_in_s: startTimeInS, end_time_in_s: endTimeInS }
    : undefined

  return useQuery({
    ...getClusterByClusterJobsOptions({
      path: { cluster },
      query,
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
  })
}

// ---------------------------------------------------------------------------
// useClusterErrorMessages
// ---------------------------------------------------------------------------

export const useClusterErrorMessages = ({ cluster, client, enabled = true }: ClusterQueryOptions) => {
  return useQuery({
    ...getClusterByClusterErrorMessagesOptions({
      path: { cluster },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
  })
}

// ---------------------------------------------------------------------------
// useClusterTimeseries (composite: GPU + CPU + Memory)
// ---------------------------------------------------------------------------

const TIMESERIES_STALE_TIME_MS = 5 * 60 * 1000 // 5 minutes
const TIMESERIES_GC_TIME_MS = 10 * 60 * 1000 // 10 minutes

interface UseClusterTimeseriesOptions extends ClusterQueryOptions {
  startTimeInS: number
  endTimeInS: number
  resolutionInS: number
}

export const useClusterTimeseries = ({
  cluster,
  client,
  startTimeInS,
  endTimeInS,
  resolutionInS,
  enabled = true,
}: UseClusterTimeseriesOptions) => {
  const isEnabled = enabled && !!client && !!cluster

  const gpuQuery = useQuery({
    ...getClusterByClusterNodesGpuTimeseriesOptions({
      path: { cluster },
      query: {
        start_time_in_s: startTimeInS,
        end_time_in_s: endTimeInS,
        resolution_in_s: resolutionInS,
      },
      client: client || undefined,
    }),
    enabled: isEnabled,
    staleTime: TIMESERIES_STALE_TIME_MS,
    gcTime: TIMESERIES_GC_TIME_MS,
  })

  const cpuQuery = useQuery({
    ...getClusterByClusterNodesCpuTimeseriesOptions({
      path: { cluster },
      query: {
        start_time_in_s: startTimeInS,
        end_time_in_s: endTimeInS,
        resolution_in_s: resolutionInS,
      },
      client: client || undefined,
    }),
    enabled: isEnabled,
    staleTime: TIMESERIES_STALE_TIME_MS,
    gcTime: TIMESERIES_GC_TIME_MS,
  })

  const memoryQuery = useQuery({
    ...getClusterByClusterNodesMemoryTimeseriesOptions({
      path: { cluster },
      query: {
        start_time_in_s: startTimeInS,
        end_time_in_s: endTimeInS,
        resolution_in_s: resolutionInS,
      },
      client: client || undefined,
    }),
    enabled: isEnabled,
    staleTime: TIMESERIES_STALE_TIME_MS,
    gcTime: TIMESERIES_GC_TIME_MS,
  })

  return { gpuQuery, cpuQuery, memoryQuery }
}
