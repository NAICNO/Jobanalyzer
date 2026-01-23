import { useQuery } from '@tanstack/react-query'
import { 
  getClusterByClusterJobsByJobIdProcessTimeseriesOptions,
  getClusterByClusterJobsByJobIdProcessGpuTimeseriesOptions 
} from '../client/@tanstack/react-query.gen'
import type { Client } from '../client/client/types.gen'

const STALE_TIME_MS = 30 * 1000 // 30 seconds
const GC_TIME_MS = 5 * 60 * 1000 // 5 minutes

interface UseJobTimeseriesOptions {
  cluster: string
  jobId: number
  client: Client | null
  resolution?: number // Resolution in seconds
  startTimeInS?: number | null // Start time in seconds since epoch
  endTimeInS?: number | null // End time in seconds since epoch
  enabled?: boolean
}

/**
 * Hook to fetch job process timeseries data (CPU, Memory, Thread count)
 */
export const useJobProcessTimeseries = ({
  cluster,
  jobId,
  client,
  resolution = 30,
  startTimeInS,
  endTimeInS,
  enabled = true
}: UseJobTimeseriesOptions) => {
  return useQuery({
    ...getClusterByClusterJobsByJobIdProcessTimeseriesOptions({
      path: { cluster, job_id: jobId },
      query: { 
        resolution_in_s: resolution,
        start_time_in_s: startTimeInS,
        end_time_in_s: endTimeInS,
      },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!jobId,
    staleTime: STALE_TIME_MS,
    gcTime: GC_TIME_MS,
    placeholderData: (previousData) => previousData, // Keep previous data while fetching
  })
}

/**
 * Hook to fetch job GPU timeseries data
 */
export const useJobGpuTimeseries = ({
  cluster,
  jobId,
  client,
  resolution = 30,
  startTimeInS,
  endTimeInS,
  enabled = true
}: UseJobTimeseriesOptions) => {
  return useQuery({
    ...getClusterByClusterJobsByJobIdProcessGpuTimeseriesOptions({
      path: { cluster, job_id: jobId },
      query: { 
        resolution_in_s: resolution,
        start_time_in_s: startTimeInS,
        end_time_in_s: endTimeInS,
      },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!jobId,
    staleTime: STALE_TIME_MS,
    gcTime: GC_TIME_MS,
    placeholderData: (previousData) => previousData, // Keep previous data while fetching
  })
}
