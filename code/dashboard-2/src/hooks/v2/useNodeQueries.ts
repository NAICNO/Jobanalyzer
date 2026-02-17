import { useQuery } from '@tanstack/react-query'
import {
  getClusterByClusterNodesOptions,
  getClusterByClusterNodesInfoOptions,
  getClusterByClusterNodesStatesOptions,
  getClusterByClusterNodesProcessGpuUtilOptions,
  getClusterByClusterNodesLastProbeTimestampOptions,
  getClusterByClusterNodesByNodenameTopologyOptions,
  getClusterByClusterNodesByNodenameStatesOptions,
  getClusterByClusterNodesByNodenameErrorMessagesOptions,
  getClusterByClusterNodesByNodenameInfoOptions,
  getClusterByClusterNodesByNodenameCpuTimeseriesOptions,
  getClusterByClusterNodesByNodenameDiskstatsTimeseriesOptions,
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
  })
}

export const useClusterNodesInfo = ({ cluster, client, enabled = true }: ClusterNodeOptions) => {
  return useQuery({
    ...getClusterByClusterNodesInfoOptions({
      path: { cluster },
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

const NODE_TIMESERIES_STALE_TIME_MS = 60_000 // 60 seconds

export const useNodeCpuTimeseries = ({ cluster, nodename, client, resolutionSec, enabled = true }: NodeTimeseriesOptions) => {
  return useQuery({
    ...getClusterByClusterNodesByNodenameCpuTimeseriesOptions({
      path: { cluster, nodename },
      query: resolutionSec ? { resolution_in_s: resolutionSec } : undefined,
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!nodename,
    staleTime: NODE_TIMESERIES_STALE_TIME_MS,
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
  })
}
