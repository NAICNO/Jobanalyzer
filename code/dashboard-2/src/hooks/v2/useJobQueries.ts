import { useQuery } from '@tanstack/react-query'
import {
  getClusterByClusterJobsByJobIdOptions,
  getClusterByClusterJobsByJobIdReportOptions,
  getClusterByClusterJobsByJobIdProcessTreeOptions,
  getClusterByClusterQueryJobsPagesOptions,
} from '../../client/@tanstack/react-query.gen'
import type { Client } from '../../client/client/types.gen'
import type { GetClusterByClusterQueryJobsPagesData } from '../../client'

// ---------------------------------------------------------------------------
// useJobDetails
// ---------------------------------------------------------------------------

const JOB_STALE_TIME_MS = 5 * 60 * 1000 // 5 minutes
const JOB_GC_TIME_MS = 10 * 60 * 1000 // 10 minutes

interface UseJobDetailsOptions {
  cluster: string
  jobId: number
  client: Client | null
  enabled?: boolean
}

export const useJobDetails = ({ cluster, jobId, client, enabled = true }: UseJobDetailsOptions) => {
  return useQuery({
    ...getClusterByClusterJobsByJobIdOptions({
      path: { cluster, job_id: jobId },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!jobId,
    staleTime: JOB_STALE_TIME_MS,
    gcTime: JOB_GC_TIME_MS,
  })
}

// ---------------------------------------------------------------------------
// useJobReport
// ---------------------------------------------------------------------------

export const useJobReport = ({ cluster, jobId, client, enabled = true }: UseJobDetailsOptions) => {
  return useQuery({
    ...getClusterByClusterJobsByJobIdReportOptions({
      path: { cluster, job_id: jobId },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!jobId,
    staleTime: JOB_STALE_TIME_MS,
    gcTime: JOB_GC_TIME_MS,
  })
}

// ---------------------------------------------------------------------------
// useJobQueryPages
// ---------------------------------------------------------------------------

interface UseJobQueryPagesOptions {
  cluster: string
  client: Client | null
  queryParams: Omit<NonNullable<GetClusterByClusterQueryJobsPagesData['query']>, 'page' | 'page_size'>
  page: number
  pageSize: number
  enabled?: boolean
}

export const useJobQueryPages = ({ cluster, client, queryParams, page, pageSize, enabled = true }: UseJobQueryPagesOptions) => {
  return useQuery({
    ...getClusterByClusterQueryJobsPagesOptions({
      path: { cluster },
      query: {
        ...queryParams,
        page,
        page_size: pageSize,
      },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
  })
}

// ---------------------------------------------------------------------------
// useJobProcessTree
// ---------------------------------------------------------------------------

interface UseJobProcessTreeOptions {
  cluster: string
  jobId: number
  client: Client | null
  nodename?: string | null
  enabled?: boolean
}

export const useJobProcessTree = ({
  cluster,
  jobId,
  client,
  nodename,
  enabled = true,
}: UseJobProcessTreeOptions) => {
  return useQuery({
    ...getClusterByClusterJobsByJobIdProcessTreeOptions({
      path: { cluster, job_id: jobId },
      query: {
        nodename: nodename,
      },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster && !!jobId,
    staleTime: JOB_STALE_TIME_MS,
    gcTime: JOB_GC_TIME_MS,
  })
}
