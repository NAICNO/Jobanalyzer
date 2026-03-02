import { useQuery, useQueries } from '@tanstack/react-query'
import {
  getClusterByClusterNodesOptions,
  getClusterByClusterNodesInfoOptions,
  getClusterByClusterNodesInfoPagesOptions,
  getClusterByClusterNodesStatesOptions,
  getClusterByClusterNodesProcessGpuUtilOptions,
  getClusterByClusterNodesMemoryTimeseriesOptions,
  getClusterByClusterNodesLastProbeTimestampOptions,
  getClusterByClusterNodesByNodenameTopologyOptions,
  getClusterByClusterNodesByNodenameStatesOptions,
  getClusterByClusterNodesByNodenameErrorMessagesOptions,
  getClusterByClusterNodesByNodenameInfoOptions,
  getClusterByClusterNodesByNodenameCpuTimeseriesOptions,
  getClusterByClusterNodesByNodenameDiskstatsTimeseriesOptions,
  getClusterByClusterNodesByNodenameMemoryTimeseriesOptions,
  getClusterByClusterNodesByNodenameGpuTimeseriesOptions,
  getClusterByClusterNodesByNodenameProcessGpuUtilOptions,
} from '../../client/@tanstack/react-query.gen'
import type { Client } from '../../client/client/types.gen'

// ---------------------------------------------------------------------------
// Shared interfaces
// ---------------------------------------------------------------------------

interface ClusterNodeOptions {
  cluster: string
  client: Client | null
  enabled?: boolean
}

interface SingleNodeOptions {
  cluster: string
  nodename: string
  client: Client | null
  enabled?: boolean
}

interface NodeTimeseriesOptions extends SingleNodeOptions {
  resolutionSec?: number
}

// ---------------------------------------------------------------------------
// Cluster-level node hooks
// ---------------------------------------------------------------------------

export const useClusterNodes = ({ cluster, client, enabled = true }: ClusterNodeOptions) => {
  return useQuery({
    ...getClusterByClusterNodesOptions({
      path: { cluster },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
    staleTime: 5 * 60 * 1000,  // 5 min - node list rarely changes
    gcTime: 30 * 60 * 1000,    // 30 min - keep in cache longer
  })
}

export const useClusterNodesInfo = ({ cluster, client, enabled = true }: ClusterNodeOptions) => {
  return useQuery({
    ...getClusterByClusterNodesInfoOptions({
      path: { cluster },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
    staleTime: 15 * 60 * 1000, // 15 min - hardware specs are static
    gcTime: 60 * 60 * 1000,    // 60 min - very safe to cache long
  })
}

interface UseClusterNodesInfoPagesOptions {
  cluster: string
  client: Client | null
  page: number
  pageSize: number
  enabled?: boolean
}

export const useClusterNodesInfoPages = ({ cluster, client, page, pageSize, enabled = true }: UseClusterNodesInfoPagesOptions) => {
  return useQuery({
    ...getClusterByClusterNodesInfoPagesOptions({
      path: { cluster },
      query: {
        page,
        page_size: pageSize,
      },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
  })
}

export const useClusterNodesStates = ({ cluster, client, enabled = true }: ClusterNodeOptions) => {
  return useQuery({
    ...getClusterByClusterNodesStatesOptions({
      path: { cluster },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
  })
}

export const useClusterNodesProcessGpuUtil = ({ cluster, client, enabled = true }: ClusterNodeOptions) => {
  return useQuery({
    ...getClusterByClusterNodesProcessGpuUtilOptions({
      path: { cluster },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
  })
}

export const useClusterNodesLastProbeTimestamp = ({ cluster, client, enabled = true }: ClusterNodeOptions) => {
  return useQuery({
    ...getClusterByClusterNodesLastProbeTimestampOptions({
      path: { cluster },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
    refetchOnWindowFocus: false, // Don't refetch on window focus
    staleTime: 2 * 60 * 1000,    // 2 min - probe timestamps update slowly
  })
}

interface ClusterNodesMemoryTimeseriesOptions extends ClusterNodeOptions {
  startTimeInS?: number
  endTimeInS?: number
  resolutionInS?: number
}

export const useClusterNodesMemoryTimeseries = ({ cluster, client, enabled = true, startTimeInS, endTimeInS, resolutionInS }: ClusterNodesMemoryTimeseriesOptions) => {
  const query = startTimeInS != null || endTimeInS != null || resolutionInS != null
    ? { start_time_in_s: startTimeInS, end_time_in_s: endTimeInS, resolution_in_s: resolutionInS }
    : undefined

  return useQuery({
    ...getClusterByClusterNodesMemoryTimeseriesOptions({
      path: { cluster },
      query,
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
    staleTime: NODE_TIMESERIES_STALE_TIME_MS,
    gcTime: NODE_TIMESERIES_GC_TIME_MS,
  })
}

// ---------------------------------------------------------------------------
// Per-node hooks
// ---------------------------------------------------------------------------

export const useNodeTopology = ({ cluster, nodename, client, enabled = true }: SingleNodeOptions) => {
  return useQuery({
    ...getClusterByClusterNodesByNodenameTopologyOptions({
      path: { cluster, nodename },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!nodename,
  })
}

export const useNodeStates = ({ cluster, nodename, client, enabled = true }: SingleNodeOptions) => {
  return useQuery({
    ...getClusterByClusterNodesByNodenameStatesOptions({
      path: { cluster, nodename },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!nodename,
  })
}

export const useNodeErrorMessages = ({ cluster, nodename, client, enabled = true }: SingleNodeOptions) => {
  return useQuery({
    ...getClusterByClusterNodesByNodenameErrorMessagesOptions({
      path: { cluster, nodename },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!nodename,
  })
}

export const useNodeInfo = ({ cluster, nodename, client, enabled = true }: SingleNodeOptions) => {
  return useQuery({
    ...getClusterByClusterNodesByNodenameInfoOptions({
      path: { cluster, nodename },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!nodename,
  })
}

export const useMultiNodeInfo = ({
  cluster,
  nodenames,
  client,
  enabled = true,
}: {
  cluster: string
  nodenames: string[]
  client: Client | null
  enabled?: boolean
}) => {
  return useQueries({
    queries: nodenames.map((nodename) => ({
      ...getClusterByClusterNodesByNodenameInfoOptions({
        path: { cluster, nodename },
        client: client || undefined,
      }),
      enabled: enabled && !!client && !!cluster && !!nodename,
      staleTime: 5 * 60 * 1000,
    })),
  })
}

const NODE_TIMESERIES_STALE_TIME_MS = 5 * 60 * 1000 // 5 minutes
const NODE_TIMESERIES_GC_TIME_MS = 30 * 60 * 1000 // 30 minutes

export const useNodeCpuTimeseries = ({ cluster, nodename, client, resolutionSec, enabled = true }: NodeTimeseriesOptions) => {
  return useQuery({
    ...getClusterByClusterNodesByNodenameCpuTimeseriesOptions({
      path: { cluster, nodename },
      query: resolutionSec ? { resolution_in_s: resolutionSec } : undefined,
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!nodename,
    staleTime: NODE_TIMESERIES_STALE_TIME_MS,
    gcTime: NODE_TIMESERIES_GC_TIME_MS,
  })
}

export const useNodeDiskstatsTimeseries = ({ cluster, nodename, client, resolutionSec, enabled = true }: NodeTimeseriesOptions) => {
  return useQuery({
    ...getClusterByClusterNodesByNodenameDiskstatsTimeseriesOptions({
      path: { cluster, nodename },
      query: resolutionSec ? { resolution_in_s: resolutionSec } : undefined,
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!nodename,
    staleTime: NODE_TIMESERIES_STALE_TIME_MS,
    gcTime: NODE_TIMESERIES_GC_TIME_MS,
  })
}

export const useNodeMemoryTimeseries = ({ cluster, nodename, client, resolutionSec, enabled = true }: NodeTimeseriesOptions) => {
  return useQuery({
    ...getClusterByClusterNodesByNodenameMemoryTimeseriesOptions({
      path: { cluster, nodename },
      query: resolutionSec ? { resolution_in_s: resolutionSec } : undefined,
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!nodename,
    staleTime: NODE_TIMESERIES_STALE_TIME_MS,
    gcTime: NODE_TIMESERIES_GC_TIME_MS,
  })
}

export const useNodeGpuTimeseries = ({ cluster, nodename, client, resolutionSec, enabled = true }: NodeTimeseriesOptions) => {
  return useQuery({
    ...getClusterByClusterNodesByNodenameGpuTimeseriesOptions({
      path: { cluster, nodename },
      query: resolutionSec ? { resolution_in_s: resolutionSec } : undefined,
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!nodename,
    staleTime: NODE_TIMESERIES_STALE_TIME_MS,
    gcTime: NODE_TIMESERIES_GC_TIME_MS,
  })
}

export const useNodeProcessGpuUtil = ({ cluster, nodename, client, enabled = true }: SingleNodeOptions) => {
  return useQuery({
    ...getClusterByClusterNodesByNodenameProcessGpuUtilOptions({
      path: { cluster, nodename },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!nodename,
  })
}
