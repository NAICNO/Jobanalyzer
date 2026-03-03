import { SimpleGrid, VStack, Text, HStack, Badge, Tag, Progress, Stat, Tooltip } from '@chakra-ui/react'
import { useNavigate } from 'react-router'

import type { NodeInfoResponse, NodeStateResponse, PartitionResponse, NodeSampleProcessGpuAccResponse, SampleProcessAccResponse } from '../../client'
import { useClusterOverviewContext } from '../../contexts/ClusterOverviewContext'
import { getUtilizationColor } from '../../util/colorStandards'

export const ClusterOverviewCards = () => {
  const navigate = useNavigate()
  const {
    cluster,
    nodesQuery: nodesQ,
    nodesInfoQuery: infoQ,
    nodesStatesQuery: statesQ,
    gpuUtilQuery: gpuUtilQ,
    memoryUtilQuery: memoryUtilQ,
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
  let totalRunningJobs = 0
  let allocatedCpus = 0

  for (const p of partitions) {
    totalCpus += p.total_cpus ?? 0
    totalGpusFromPartitions += p.total_gpus ?? 0
    gpusInUseFromPartitions += p.gpus_in_use?.length ?? 0
    totalPendingJobs += p.jobs_pending?.length ?? 0
    totalRunningJobs += p.jobs_running?.length ?? 0

    // Calculate allocated CPUs from running jobs
    for (const job of p.jobs_running ?? []) {
      allocatedCpus += job.requested_cpus ?? 0
    }
  }

  const cpuUtilPct = totalCpus > 0 ? Math.round((allocatedCpus / totalCpus) * 100) : 0
  const cpuColor = totalCpus === 0 ? 'gray' : getUtilizationColor(cpuUtilPct)

  // Calculate total memory from nodes
  let totalMemoryGB = 0
  for (const node of Object.values(infoMap)) {
    totalMemoryGB += (node.memory ?? 0) / (1024 * 1024) // Convert bytes to GB
  }
  totalMemoryGB = Math.round(totalMemoryGB)

  // Calculate memory utilization from timeseries (latest snapshot)
  const memoryData = (memoryUtilQ.data ?? {}) as Record<string, Array<SampleProcessAccResponse>>
  let totalMemoryUtilPct = 0
  let memorySampleCount = 0

  for (const samples of Object.values(memoryData)) {
    if (Array.isArray(samples) && samples.length > 0) {
      const latestSample = samples[samples.length - 1]
      totalMemoryUtilPct += latestSample.memory_util ?? 0
      memorySampleCount++
    }
  }

  const memoryUtilPct = memorySampleCount > 0 ? Math.round(totalMemoryUtilPct / memorySampleCount) : 0
  const usedMemoryGB = Math.round(totalMemoryGB * (memoryUtilPct / 100))
  const memoryColor = getUtilizationColor(memoryUtilPct)

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
  const gpuColor = totalGpus === 0 ? 'gray' : getUtilizationColor(gpuUtilPct)

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
      <SimpleGrid columns={{ base: 1, sm: 2, md: 3, lg: 4, xl: 9 }} gap={3} w="100%">
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
        {/* CPUs Card - Combined Total + Utilization */}
        <Stat.Root
          borderWidth="1px"
          borderColor="gray.200"
          rounded="md"
          p={2}
          bg="white"
        >
          <HStack gap={1} justify="space-between">
            <Tooltip.Root>
              <Tooltip.Trigger asChild>
                <Stat.Label fontSize="sm" color="gray.600" cursor="help">CPU Allocation</Stat.Label>
              </Tooltip.Trigger>
              <Tooltip.Positioner>
                <Tooltip.Content>Cores allocated to running jobs vs total cores across all partitions</Tooltip.Content>
              </Tooltip.Positioner>
            </Tooltip.Root>
            {partitionsQ.isError && (
              <Tooltip.Root>
                <Tooltip.Trigger asChild>
                  <Badge colorPalette="red" variant="solid" size="xs" cursor="pointer">!</Badge>
                </Tooltip.Trigger>
                <Tooltip.Positioner>
                  <Tooltip.Content>{getErrMsg(partitionsQ.error)}</Tooltip.Content>
                </Tooltip.Positioner>
              </Tooltip.Root>
            )}
          </HStack>
          <HStack justify="space-between" w="100%" mb={1}>
            <Text fontSize="xs" color="gray.600">{allocatedCpus} / {totalCpus}</Text>
            <Tag.Root size="sm" colorPalette={cpuColor}><Tag.Label>{cpuUtilPct}%</Tag.Label></Tag.Root>
          </HStack>
          <Progress.Root value={cpuUtilPct} max={100} colorPalette={cpuColor} w="100%" size="xs" aria-label={`CPU utilization: ${cpuUtilPct}% (${allocatedCpus} of ${totalCpus} allocated)`}>
            <Progress.Track>
              <Progress.Range />
            </Progress.Track>
          </Progress.Root>
          {partitionsQ.isLoading && (
            <Badge colorPalette="blue" size="xs" mt={1}>updating</Badge>
          )}
        </Stat.Root>

        {/* GPUs Card - Combined Total + Utilization */}
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
            <Tooltip.Root>
              <Tooltip.Trigger asChild>
                <Stat.Label fontSize="sm" color="gray.600" cursor="help">GPU Usage</Stat.Label>
              </Tooltip.Trigger>
              <Tooltip.Positioner>
                <Tooltip.Content>GPUs with utilization greater than 0% vs total GPUs available</Tooltip.Content>
              </Tooltip.Positioner>
            </Tooltip.Root>
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

        {/* Memory Card - Combined Total + Utilization */}
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
            <Tooltip.Root>
              <Tooltip.Trigger asChild>
                <Stat.Label fontSize="sm" color="gray.600" cursor="help">Memory Usage</Stat.Label>
              </Tooltip.Trigger>
              <Tooltip.Positioner>
                <Tooltip.Content>Actual memory utilization averaged across reporting nodes</Tooltip.Content>
              </Tooltip.Positioner>
            </Tooltip.Root>
            {(memoryUtilQ.isError || infoQ.isError) && (
              <Tooltip.Root>
                <Tooltip.Trigger asChild>
                  <Badge colorPalette="red" variant="solid" size="xs" cursor="pointer">!</Badge>
                </Tooltip.Trigger>
                <Tooltip.Positioner>
                  <Tooltip.Content>
                    {[
                      ...(memoryUtilQ.isError ? [getErrMsg(memoryUtilQ.error)] : []),
                      ...(infoQ.isError ? [getErrMsg(infoQ.error)] : [])
                    ].join('; ')}
                  </Tooltip.Content>
                </Tooltip.Positioner>
              </Tooltip.Root>
            )}
          </HStack>
          <HStack justify="space-between" w="100%" mb={1}>
            <Text fontSize="xs" color="gray.600">{usedMemoryGB} / {totalMemoryGB} GB</Text>
            <Tag.Root size="sm" colorPalette={memoryColor}><Tag.Label>{memoryUtilPct}%</Tag.Label></Tag.Root>
          </HStack>
          <Progress.Root value={memoryUtilPct} max={100} colorPalette={memoryColor} w="100%" size="xs" aria-label={`Memory utilization: ${memoryUtilPct}% (${usedMemoryGB} of ${totalMemoryGB} GB in use)`}>
            <Progress.Track>
              <Progress.Range />
            </Progress.Track>
          </Progress.Root>
          {(memoryUtilQ.isLoading || infoQ.isLoading) && (
            <Badge colorPalette="blue" size="xs" mt={1}>updating</Badge>
          )}
        </Stat.Root>

        {/* Queue metrics */}
        <Card
          label="Running Jobs"
          value={totalRunningJobs}
          loading={partitionsQ.isLoading}
          errors={partitionsQ.isError ? [getErrMsg(partitionsQ.error)] : []}
          onClick={() => navigate(`/v2/${cluster}/jobs/running`)}
        />
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
