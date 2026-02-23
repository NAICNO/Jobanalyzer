import { SimpleGrid, VStack, Text, HStack, Badge, Tag, Progress, Stat, Tooltip } from '@chakra-ui/react'
import { useNavigate } from 'react-router'

import type { NodeInfoResponse, NodeStateResponse, PartitionResponse, NodeSampleProcessGpuAccResponse } from '../../client'
import { useClusterOverviewContext } from '../../contexts/ClusterOverviewContext'

export const ClusterOverviewCards = () => {
  const navigate = useNavigate()
  const {
    cluster,
    nodesQuery: nodesQ,
    nodesInfoQuery: infoQ,
    nodesStatesQuery: statesQ,
    gpuUtilQuery: gpuUtilQ,
    partitionsQuery: partitionsQ,
  } = useClusterOverviewContext()

  // Process nodes data
  const nodes = (nodesQ.data ?? []) as string[]
  const infoMap = (infoQ.data ?? {}) as Record<string, NodeInfoResponse>
  const statesArr = (statesQ.data ?? []) as NodeStateResponse[]

  const totalNodes = nodes.length
  const nodeSet = new Set(nodes)
  const reportingNodes = Object.keys(infoMap).filter(n => nodeSet.has(n)).length
  const idleNodes = (() => {
    const latestByNode = new Map<string, NodeStateResponse>()
    for (const s of statesArr) {
      const existing = latestByNode.get(s.node)
      if (!existing || new Date(s.time) > new Date(existing.time)) {
        latestByNode.set(s.node, s)
      }
    }
    let count = 0
    for (const s of latestByNode.values()) {
      if (Array.isArray(s.states) && s.states.includes('IDLE')) count += 1
    }
    return count
  })()

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

  const Card = ({ label, value, loading, errors, onClick }: { label: string; value: string | number; loading?: boolean; errors?: string[]; onClick?: () => void }) => {
    const hasErrors = errors && errors.length > 0
    const errorMsg = hasErrors ? errors.join('; ') : ''
    return (
      <Stat.Root
        borderWidth="1px"
        borderColor="gray.200"
        rounded="md"
        p={2}
        bg="white"
        cursor={onClick ? 'pointer' : undefined}
        _hover={onClick ? { borderColor: 'blue.300', bg: 'gray.50' } : undefined}
        onClick={onClick}
      >
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
          onClick={() => navigate(`/v2/${cluster}/nodes`)}
        />
        <Card
          label="Reporting Nodes"
          value={reportingNodes}
          loading={infoQ.isLoading}
          errors={infoQ.isError ? [getErrMsg(infoQ.error)] : []}
          onClick={() => navigate(`/v2/${cluster}/nodes`)}
        />
        <Card
          label="Idle Nodes"
          value={idleNodes}
          loading={statesQ.isLoading}
          errors={statesQ.isError ? [getErrMsg(statesQ.error)] : []}
          onClick={() => navigate(`/v2/${cluster}/nodes`)}
        />

        {/* Partitions metrics */}
        <Card
          label="Partitions"
          value={totalPartitions}
          loading={partitionsQ.isLoading}
          errors={partitionsQ.isError ? [getErrMsg(partitionsQ.error)] : []}
          onClick={() => navigate(`/v2/${cluster}/partitions`)}
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
        <Stat.Root
          borderWidth="1px"
          borderColor="gray.200"
          rounded="md"
          p={2}
          bg="white"
          cursor="pointer"
          _hover={{ borderColor: 'blue.300', bg: 'gray.50' }}
          onClick={() => navigate(`/v2/${cluster}/nodes`)}
        >
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
          <Progress.Root value={gpuUtilPct} max={100} colorPalette={gpuColor} w="100%" size="xs" aria-label={`GPU utilization: ${gpuUtilPct}% (${gpusInUse} of ${totalGpus} in use)`}>
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
          onClick={() => navigate(`/v2/${cluster}/jobs`)}
        />
      </SimpleGrid>
    </VStack>
  )
}
