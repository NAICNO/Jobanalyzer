import { useQuery } from '@tanstack/react-query'
import { getClusterByClusterBenchmarksByBenchmarkNameOptions } from '../../client/@tanstack/react-query.gen'
import type { Client } from '../../client/client/types.gen'
import type { BenchmarkRecord } from '../../types/benchmark'

interface UseBenchmarksOptions {
  cluster: string
  client: Client | null
  benchmarkName?: string
  enabled?: boolean
}

export const useBenchmarks = ({
  cluster,
  client,
  benchmarkName = 'lambdal',
  enabled = true,
}: UseBenchmarksOptions) => {
  return useQuery({
    ...getClusterByClusterBenchmarksByBenchmarkNameOptions({
      path: { cluster, benchmark_name: benchmarkName },
      client: client || undefined,
    }),
    enabled: enabled && !!client && !!cluster,
    staleTime: 10 * 60 * 1000,
    refetchOnWindowFocus: false,
    select: (data) => data as BenchmarkRecord[],
  })
}
