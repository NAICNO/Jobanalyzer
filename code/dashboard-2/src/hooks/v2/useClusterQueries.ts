import { useQuery } from '@tanstack/react-query'
import {
  getClusterByClusterPartitionsOptions,
  getClusterByClusterJobsOptions,
  getClusterByClusterErrorMessagesOptions,
  getClusterByClusterNodesGpuTimeseriesOptions,
  getClusterByClusterNodesCpuTimeseriesOptions,
  getClusterByClusterNodesMemoryTimeseriesOptions,
  getClusterByClusterNodesDiskstatsTimeseriesOptions,
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
    refetchOnWindowFocus: false, // Don't refetch on window focus
    staleTime: 2 * 60 * 1000,    // 2 min - errors don't need immediate updates
  })
}

// ---------------------------------------------------------------------------
// Timeseries hooks
// ---------------------------------------------------------------------------

const TIMESERIES_STALE_TIME_MS = 5 * 60 * 1000 // 5 minutes
const TIMESERIES_GC_TIME_MS = 30 * 60 * 1000 // 30 minutes

interface UseClusterTimeseriesOptions extends ClusterQueryOptions {
  startTimeInS: number
  endTimeInS: number
  resolutionInS: number
}

// ---------------------------------------------------------------------------
// useClusterGpuTimeseries
// ---------------------------------------------------------------------------

export const useClusterGpuTimeseries = ({
  cluster, client, startTimeInS, endTimeInS, resolutionInS, enabled = true,
}: UseClusterTimeseriesOptions) => {
  return useQuery({
    ...getClusterByClusterNodesGpuTimeseriesOptions({
      path: { cluster },
      query: { start_time_in_s: startTimeInS, end_time_in_s: endTimeInS, resolution_in_s: resolutionInS },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
    staleTime: TIMESERIES_STALE_TIME_MS,
    gcTime: TIMESERIES_GC_TIME_MS,
  })
}

// ---------------------------------------------------------------------------
// useClusterCpuTimeseries
// ---------------------------------------------------------------------------

export const useClusterCpuTimeseries = ({
  cluster, client, startTimeInS, endTimeInS, resolutionInS, enabled = true,
}: UseClusterTimeseriesOptions) => {
  return useQuery({
    ...getClusterByClusterNodesCpuTimeseriesOptions({
      path: { cluster },
      query: { start_time_in_s: startTimeInS, end_time_in_s: endTimeInS, resolution_in_s: resolutionInS },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
    staleTime: TIMESERIES_STALE_TIME_MS,
    gcTime: TIMESERIES_GC_TIME_MS,
  })
}

// ---------------------------------------------------------------------------
// useClusterMemoryTimeseries
// ---------------------------------------------------------------------------

export const useClusterMemoryTimeseries = ({
  cluster, client, startTimeInS, endTimeInS, resolutionInS, enabled = true,
}: UseClusterTimeseriesOptions) => {
  return useQuery({
    ...getClusterByClusterNodesMemoryTimeseriesOptions({
      path: { cluster },
      query: { start_time_in_s: startTimeInS, end_time_in_s: endTimeInS, resolution_in_s: resolutionInS },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
    staleTime: TIMESERIES_STALE_TIME_MS,
    gcTime: TIMESERIES_GC_TIME_MS,
  })
}

// ---------------------------------------------------------------------------
// useClusterDiskTimeseries
// ---------------------------------------------------------------------------

export const useClusterDiskTimeseries = ({
  cluster, client, startTimeInS, endTimeInS, resolutionInS, enabled = true,
}: UseClusterTimeseriesOptions) => {
  return useQuery({
    ...getClusterByClusterNodesDiskstatsTimeseriesOptions({
      path: { cluster },
      query: { start_time_in_s: startTimeInS, end_time_in_s: endTimeInS, resolution_in_s: resolutionInS },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
    staleTime: TIMESERIES_STALE_TIME_MS,
    gcTime: TIMESERIES_GC_TIME_MS,
  })
}
