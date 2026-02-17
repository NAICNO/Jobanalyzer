import { useQuery } from '@tanstack/react-query'
import { getClusterByClusterJobsByJobIdProcessTreeOptions } from '../client/@tanstack/react-query.gen'
import type { Client } from '../client/client/types.gen'

const STALE_TIME_MS = 5 * 60 * 1000 // 5 minutes
const GC_TIME_MS = 10 * 60 * 1000 // 10 minutes

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
    staleTime: STALE_TIME_MS,
    gcTime: GC_TIME_MS,
  })
}
