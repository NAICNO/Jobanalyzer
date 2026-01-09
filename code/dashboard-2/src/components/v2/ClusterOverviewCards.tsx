import { SimpleGrid, VStack, Text, HStack, Badge, Tag, Progress, Stat, Tooltip, Spinner } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'

import {
  getClusterByClusterNodesOptions,
  getClusterByClusterNodesInfoOptions,
  getClusterByClusterNodesStatesOptions,
  getClusterByClusterPartitionsOptions,
  getClusterByClusterNodesProcessGpuUtilOptions
} from '../../client/@tanstack/react-query.gen'
import type { NodeInfoResponse, NodeStateResponse, PartitionResponse, NodeSampleProcessGpuAccResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'

interface Props {
  cluster: string
}

export const ClusterOverviewCards = ({ cluster }: Props) => {
  const client = useClusterClient(cluster)
  
  // Early return if client is not available
  if (!client) {
    return <Spinner />
  }
  
  // Fetch nodes data
  const nodesQ = useQuery({
    ...getClusterByClusterNodesOptions({ path: { cluster }, client }),
    enabled: !!cluster,
  })
  const infoQ = useQuery({
    ...getClusterByClusterNodesInfoOptions({ path: { cluster }, client }),
    enabled: !!cluster,
  })
  const statesQ = useQuery({
    ...getClusterByClusterNodesStatesOptions({ path: { cluster }, client }),
    enabled: !!cluster,
  })
  const gpuUtilQ = useQuery({
    ...getClusterByClusterNodesProcessGpuUtilOptions({ path: { cluster }, client }),
    enabled: !!cluster,
  })

  // Fetch partitions data
  const partitionsQ = useQuery({
    ...getClusterByClusterPartitionsOptions({ path: { cluster }, client }),
    enabled: !!cluster,
  })

  // Process nodes data
  const nodes = (nodesQ.data ?? []) as string[]
  const infoMap = (infoQ.data ?? {}) as Record<string, NodeInfoResponse>
  const statesArr = (statesQ.data ?? []) as NodeStateResponse[]

  const totalNodes = nodes.length
  const reportingNodes = Object.keys(infoMap).length
  const idleNodes = statesArr.reduce((acc, s) => acc + (Array.isArray(s.states) && s.states.includes('IDLE') ? 1 : 0), 0)

  // Process partitions data
  const partitionsMap = (partitionsQ.data ?? {}) as Record<string, PartitionResponse>
  const partitions = Object.values(partitionsMap)
  const totalPartitions = partitions.length

  // Calculate totals from partitions
  let totalCpus = 0
  let totalGpusFromPartitions = 0
  let gpusInUseFromPartitions = 0
  let totalPendingJobs = 0

  for (const p of partitions) {
    totalCpus += p.total_cpus ?? 0
    totalGpusFromPartitions += p.total_gpus ?? 0
    gpusInUseFromPartitions += p.gpus_in_use?.length ?? 0
    totalPendingJobs += p.jobs_pending?.length ?? 0
  }

  // Calculate GPU utilization from nodes (alternative source)
  const totalGpusFromNodes = Object.values(infoMap).reduce((sum, n) => sum + (Array.isArray(n.cards) ? n.cards.length : 0), 0)
  
  const utilData = (gpuUtilQ.data ?? undefined) as NodeSampleProcessGpuAccResponse | undefined
  let gpusInUseFromUtil = 0
  if (utilData && typeof utilData === 'object') {
    const utilDataMap = utilData as Record<string, Record<string, { gpu_util?: number }>>
    for (const nodeMap of Object.values(utilDataMap)) {
      for (const sample of Object.values(nodeMap)) {
        const util = sample.gpu_util ?? 0
        if (util > 0) gpusInUseFromUtil += 1
      }
    }
  }

  // Use the more comprehensive GPU data (prefer nodes info)
  const totalGpus = totalGpusFromNodes > 0 ? totalGpusFromNodes : totalGpusFromPartitions
  const gpusInUse = gpusInUseFromUtil > 0 ? gpusInUseFromUtil : gpusInUseFromPartitions
  const gpuUtilPct = totalGpus > 0 ? Math.round((gpusInUse / totalGpus) * 100) : 0
  const gpuColor = totalGpus === 0 ? 'gray' : gpuUtilPct > 90 ? 'red' : gpuUtilPct > 50 ? 'yellow' : 'green'

  const getErrMsg = (e: unknown): string => {
    if (!e) return 'Unknown error'
    if (e instanceof Error) return e.message
    if (typeof e === 'string') return e
    try {
      return JSON.stringify(e)
    } catch {
      return 'Unknown error'
    }
  }

  const Card = ({ label, value, loading, errors }: { label: string; value: string | number; loading?: boolean; errors?: string[] }) => {
    const hasErrors = errors && errors.length > 0
    const errorMsg = hasErrors ? errors.join('; ') : ''
    return (
      <Stat.Root borderWidth="1px" borderColor="gray.200" rounded="md" p={2} bg="white">
        <HStack gap={1} justify="space-between">
          <Stat.Label fontSize="sm" color="gray.600">{label}</Stat.Label>
          {hasErrors && (
            <Tooltip.Root>
              <Tooltip.Trigger asChild>
                <Badge colorPalette="red" variant="solid" size="xs" cursor="pointer">!</Badge>
              </Tooltip.Trigger>
              <Tooltip.Positioner>
                <Tooltip.Content>{errorMsg}</Tooltip.Content>
              </Tooltip.Positioner>
            </Tooltip.Root>
          )}
        </HStack>
        <HStack gap={1}>
          {loading ? <Badge colorPalette="blue" size="xs">updating</Badge> : null}
          <Stat.ValueText fontSize="2xl" fontWeight="semibold">{value}</Stat.ValueText>
        </HStack>
      </Stat.Root>
    )
  }

  return (
    <VStack w="100%" align="start" gap={2}>
      <Text fontSize="lg" fontWeight="semibold">Cluster overview</Text>
      <SimpleGrid columns={{ base: 1, sm: 2, md: 4, lg: 8 }} gap={3} w="100%">
        {/* Nodes metrics */}
        <Card 
          label="Total Nodes" 
          value={totalNodes} 
          loading={nodesQ.isLoading} 
          errors={nodesQ.isError ? [getErrMsg(nodesQ.error)] : []} 
        />
        <Card 
          label="Reporting Nodes" 
          value={reportingNodes} 
          loading={infoQ.isLoading} 
          errors={infoQ.isError ? [getErrMsg(infoQ.error)] : []} 
        />
        <Card 
          label="Idle Nodes" 
          value={idleNodes} 
          loading={statesQ.isLoading} 
          errors={statesQ.isError ? [getErrMsg(statesQ.error)] : []} 
        />

        {/* Partitions metrics */}
        <Card 
          label="Partitions" 
          value={totalPartitions} 
          loading={partitionsQ.isLoading} 
          errors={partitionsQ.isError ? [getErrMsg(partitionsQ.error)] : []} 
        />

        {/* Resource metrics */}
        <Card 
          label="Total CPUs" 
          value={totalCpus} 
          loading={partitionsQ.isLoading} 
          errors={partitionsQ.isError ? [getErrMsg(partitionsQ.error)] : []} 
        />
        <Card 
          label="Total GPUs" 
          value={totalGpus} 
          loading={infoQ.isLoading || partitionsQ.isLoading} 
          errors={[
            ...(infoQ.isError ? [getErrMsg(infoQ.error)] : []),
            ...(partitionsQ.isError ? [getErrMsg(partitionsQ.error)] : []),
          ]} 
        />

        {/* GPU Utilization Card */}
        <Stat.Root borderWidth="1px" borderColor="gray.200" rounded="md" p={2} bg="white">
          <HStack gap={1} justify="space-between">
            <Stat.Label fontSize="sm" color="gray.600">GPU Utilization</Stat.Label>
            {(gpuUtilQ.isError || infoQ.isError) && (
              <Tooltip.Root>
                <Tooltip.Trigger asChild>
                  <Badge colorPalette="red" variant="solid" size="xs" cursor="pointer">!</Badge>
                </Tooltip.Trigger>
                <Tooltip.Positioner>
                  <Tooltip.Content>
                    {[
                      ...(gpuUtilQ.isError ? [getErrMsg(gpuUtilQ.error)] : []),
                      ...(infoQ.isError ? [getErrMsg(infoQ.error)] : [])
                    ].join('; ')}
                  </Tooltip.Content>
                </Tooltip.Positioner>
              </Tooltip.Root>
            )}
          </HStack>
          <HStack justify="space-between" w="100%" mb={1}>
            <Text fontSize="xs" color="gray.600">{gpusInUse} / {totalGpus}</Text>
            <Tag.Root size="sm" colorPalette={gpuColor}><Tag.Label>{gpuUtilPct}%</Tag.Label></Tag.Root>
          </HStack>
          <Progress.Root value={gpuUtilPct} max={100} colorPalette={gpuColor} w="100%" size="xs">
            <Progress.Track>
              <Progress.Range />
            </Progress.Track>
          </Progress.Root>
          {(gpuUtilQ.isLoading || infoQ.isLoading) && (
            <Badge colorPalette="blue" size="xs" mt={1}>updating</Badge>
          )}
        </Stat.Root>

        {/* Queue metrics */}
        <Card 
          label="Pending Jobs" 
          value={totalPendingJobs} 
          loading={partitionsQ.isLoading} 
          errors={partitionsQ.isError ? [getErrMsg(partitionsQ.error)] : []} 
        />
      </SimpleGrid>
    </VStack>
  )
}
