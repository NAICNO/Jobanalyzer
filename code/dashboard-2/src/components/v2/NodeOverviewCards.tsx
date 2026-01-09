import { SimpleGrid, VStack, Text, HStack, Badge, Tag, Progress, Stat, Tooltip, Spinner } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { TbBroadcast, TbBroadcastOff } from 'react-icons/tb'

import { getClusterByClusterNodesOptions, getClusterByClusterNodesInfoOptions, getClusterByClusterNodesStatesOptions, getClusterByClusterNodesProcessGpuUtilOptions, getClusterByClusterNodesLastProbeTimestampOptions } from '../../client/@tanstack/react-query.gen'
import type { NodeInfoResponse, NodeStateResponse, NodeSampleProcessGpuAccResponse } from '../../client'
import { getLivenessStats, LIVENESS_COLORS } from '../../utils/nodeLiveness'
import { getUtilizationColorPalette, calculateUtilization } from '../../utils/utilizationColors'
import { useClusterClient } from '../../hooks/useClusterClient'

interface Props {
  cluster: string
}

export const NodeOverviewCards = ({ cluster }: Props) => {
  const client = useClusterClient(cluster)
  
  if (!client) {
    return <Spinner />
  }
  
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
  const lastProbeQ = useQuery({
    ...getClusterByClusterNodesLastProbeTimestampOptions({ path: { cluster }, client }),
    enabled: !!cluster,
  })

  const nodes = (nodesQ.data ?? []) as string[]
  const infoMap = (infoQ.data ?? {}) as Record<string, NodeInfoResponse>
  const statesArr = (statesQ.data ?? []) as NodeStateResponse[]
  const lastProbeMap = (lastProbeQ.data ?? {}) as Record<string, Date | null>

  const totalNodes = nodes.length
  const totalGpus = Object.values(infoMap).reduce((sum, n) => sum + (Array.isArray(n.cards) ? n.cards.length : 0), 0)
  const idleNodes = statesArr.reduce((acc, s) => acc + (Array.isArray(s.states) && s.states.includes('IDLE') ? 1 : 0), 0)

  // Calculate liveness stats using utility
  const { liveNodes, staleNodes, offlineNodes } = getLivenessStats(lastProbeMap)

  // GPUs in use derived from latest process GPU util samples (> 0%)
  const utilData = (gpuUtilQ.data ?? undefined) as NodeSampleProcessGpuAccResponse | undefined
  let gpusInUse = 0
  if (utilData && typeof utilData === 'object') {
    const utilDataMap = utilData as Record<string, Record<string, { gpu_util?: number }>>
    for (const nodeMap of Object.values(utilDataMap)) {
      for (const sample of Object.values(nodeMap)) {
        const util = sample.gpu_util ?? 0
        if (util > 0) gpusInUse += 1
      }
    }
  }
  const gpuUtilPct = calculateUtilization(gpusInUse, totalGpus)
  const gpuColor = totalGpus === 0 ? 'gray' : getUtilizationColorPalette(gpuUtilPct)

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
      <SimpleGrid columns={{ base: 1, sm: 2, lg: 4 }} gap={3} w="100%">
        <Card label="Total Nodes" value={totalNodes} loading={nodesQ.isLoading} errors={nodesQ.isError ? [getErrMsg(nodesQ.error)] : []} />
        {/* Node Health Card with Live/Stale/Offline breakdown */}
        <Stat.Root borderWidth="1px" borderColor="gray.200" rounded="md" p={2} bg="white">
          <HStack gap={1} justify="space-between">
            <Stat.Label fontSize="sm" color="gray.600">Node Health</Stat.Label>
            {lastProbeQ.isError && (
              <Tooltip.Root>
                <Tooltip.Trigger asChild>
                  <Badge colorPalette="red" variant="solid" size="xs" cursor="pointer">!</Badge>
                </Tooltip.Trigger>
                <Tooltip.Positioner>
                  <Tooltip.Content>{getErrMsg(lastProbeQ.error)}</Tooltip.Content>
                </Tooltip.Positioner>
              </Tooltip.Root>
            )}
          </HStack>
          {lastProbeQ.isLoading && (
            <Badge colorPalette="blue" size="xs" mb={1}>updating</Badge>
          )}
          <VStack align="start" gap={0.5} mt={1}>
            <HStack gap={1.5}>
              <TbBroadcast style={{ color: LIVENESS_COLORS.LIVE, fontSize: '1.1em' }} />
              <Text fontSize="md" color="green.600" fontWeight="semibold">{liveNodes}</Text>
              <Text fontSize="sm" color="gray.600">Live (&lt;5m)</Text>
            </HStack>
            <HStack gap={1.5}>
              <TbBroadcast style={{ color: LIVENESS_COLORS.STALE, fontSize: '1.1em' }} />
              <Text fontSize="md" color="yellow.600" fontWeight="semibold">{staleNodes}</Text>
              <Text fontSize="sm" color="gray.600">Stale (5-15m)</Text>
            </HStack>
            <HStack gap={1.5}>
              <TbBroadcastOff style={{ color: LIVENESS_COLORS.OFFLINE, fontSize: '1.1em' }} />
              <Text fontSize="md" color="red.600" fontWeight="semibold">{offlineNodes}</Text>
              <Text fontSize="sm" color="gray.600">Offline (&gt;15m)</Text>
            </HStack>
          </VStack>
        </Stat.Root>
        <Card label="Idle Nodes" value={idleNodes} loading={statesQ.isLoading} errors={statesQ.isError ? [getErrMsg(statesQ.error)] : []} />
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
      </SimpleGrid>
    </VStack>
  )
}
