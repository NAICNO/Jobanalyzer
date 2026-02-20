import { memo, useMemo } from 'react'
import { SimpleGrid, VStack, Text, HStack, Badge, Tag, Progress, Stat, Tooltip, Card } from '@chakra-ui/react'
import { TbBroadcast, TbBroadcastOff } from 'react-icons/tb'

import type { NodeInfoResponse, NodeStateResponse, NodeSampleProcessGpuAccResponse } from '../../client'
import { getLivenessStats, LIVENESS_COLORS, LIVENESS_THRESHOLDS } from '../../utils/nodeLiveness'
import { getUtilizationColorPalette, calculateUtilization } from '../../utils/utilizationColors'
import { useClusterClient } from '../../hooks/useClusterClient'
import { useClusterNodes, useClusterNodesInfo, useClusterNodesStates, useClusterNodesProcessGpuUtil, useClusterNodesLastProbeTimestamp } from '../../hooks/v2/useNodeQueries'

interface Props {
  cluster: string
}

const GRID_COLUMNS = { base: 1, sm: 2, lg: 3 } as const

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

interface NodesCardProps {
  totalNodes: number
  idleNodes: number
  loading?: boolean
  errors?: string[]
}

const NodesCard = memo(({ totalNodes, idleNodes, loading, errors }: NodesCardProps) => {
  const hasErrors = errors && errors.length > 0
  const errorMsg = hasErrors ? errors.join('; ') : ''
  return (
    <Card.Root size="sm">
      <Card.Body>
        <Stat.Root>
          <HStack gap={1} justify="space-between">
            <Stat.Label fontSize="sm" color="gray.600">Nodes</Stat.Label>
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
            <Stat.ValueText fontSize="2xl" fontWeight="semibold">{totalNodes}</Stat.ValueText>
            <Text fontSize="sm" color="gray.500">total</Text>
            <Text fontSize="sm" color="gray.400">/</Text>
            <Stat.ValueText fontSize="2xl" fontWeight="semibold">{idleNodes}</Stat.ValueText>
            <Text fontSize="sm" color="gray.500">idle</Text>
          </HStack>
        </Stat.Root>
      </Card.Body>
    </Card.Root>
  )
})

NodesCard.displayName = 'NodesCard'

interface NodeHealthCardProps {
  liveNodes: number
  staleNodes: number
  offlineNodes: number
  loading?: boolean
  error?: unknown
}

const NodeHealthCard = memo(({ liveNodes, staleNodes, offlineNodes, loading, error }: NodeHealthCardProps) => (
  <Card.Root size="sm">
    <Card.Body>
      <Stat.Root>
        <HStack gap={1} justify="space-between">
          <Stat.Label fontSize="sm" color="gray.600">Node Health</Stat.Label>
          {!!error && (
            <Tooltip.Root>
              <Tooltip.Trigger asChild>
                <Badge colorPalette="red" variant="solid" size="xs" cursor="pointer">!</Badge>
              </Tooltip.Trigger>
              <Tooltip.Positioner>
                <Tooltip.Content>{getErrMsg(error)}</Tooltip.Content>
              </Tooltip.Positioner>
            </Tooltip.Root>
          )}
        </HStack>
        {loading && (
          <Badge colorPalette="blue" size="xs" mb={1}>updating</Badge>
        )}
        <HStack gap={4} mt={1} wrap="wrap">
          <Tooltip.Root>
            <Tooltip.Trigger asChild>
              <HStack gap={1.5} cursor="default">
                <TbBroadcast style={{ color: LIVENESS_COLORS.LIVE, fontSize: '1.1em' }} />
                <Text fontSize="md" color="green.600" fontWeight="semibold">{liveNodes}</Text>
                <Text fontSize="xs" color="gray.500">Live</Text>
              </HStack>
            </Tooltip.Trigger>
            <Tooltip.Positioner>
              <Tooltip.Content>Last seen &lt; {LIVENESS_THRESHOLDS.LIVE} min ago</Tooltip.Content>
            </Tooltip.Positioner>
          </Tooltip.Root>
          <Tooltip.Root>
            <Tooltip.Trigger asChild>
              <HStack gap={1.5} cursor="default">
                <TbBroadcast style={{ color: LIVENESS_COLORS.STALE, fontSize: '1.1em' }} />
                <Text fontSize="md" color="yellow.600" fontWeight="semibold">{staleNodes}</Text>
                <Text fontSize="xs" color="gray.500">Stale</Text>
              </HStack>
            </Tooltip.Trigger>
            <Tooltip.Positioner>
              <Tooltip.Content>{LIVENESS_THRESHOLDS.LIVE}–{LIVENESS_THRESHOLDS.STALE} min since last probe</Tooltip.Content>
            </Tooltip.Positioner>
          </Tooltip.Root>
          <Tooltip.Root>
            <Tooltip.Trigger asChild>
              <HStack gap={1.5} cursor="default">
                <TbBroadcastOff style={{ color: LIVENESS_COLORS.OFFLINE, fontSize: '1.1em' }} />
                <Text fontSize="md" color="red.600" fontWeight="semibold">{offlineNodes}</Text>
                <Text fontSize="xs" color="gray.500">Offline</Text>
              </HStack>
            </Tooltip.Trigger>
            <Tooltip.Positioner>
              <Tooltip.Content>&gt; {LIVENESS_THRESHOLDS.STALE} min since last probe</Tooltip.Content>
            </Tooltip.Positioner>
          </Tooltip.Root>
        </HStack>
      </Stat.Root>
    </Card.Body>
  </Card.Root>
))

NodeHealthCard.displayName = 'NodeHealthCard'

interface GpuUtilizationCardProps {
  gpusInUse: number
  totalGpus: number
  gpuUtilPct: number
  gpuColor: string
  loading?: boolean
  errors?: string[]
}

const GpuUtilizationCard = memo(({ gpusInUse, totalGpus, gpuUtilPct, gpuColor, loading, errors }: GpuUtilizationCardProps) => {
  const hasErrors = errors && errors.length > 0
  const errorMsg = hasErrors ? errors.join('; ') : ''

  return (
    <Card.Root size="sm">
      <Card.Body>
        <Stat.Root>
          <HStack gap={1} justify="space-between">
            <Stat.Label fontSize="sm" color="gray.600">GPU Utilization</Stat.Label>
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
          <HStack justify="space-between" w="100%" mb={1}>
            <Text fontSize="xs" color="gray.600">{gpusInUse} / {totalGpus}</Text>
            <Tag.Root size="sm" colorPalette={gpuColor}><Tag.Label>{gpuUtilPct}%</Tag.Label></Tag.Root>
          </HStack>
          <Progress.Root value={gpuUtilPct} max={100} colorPalette={gpuColor} w="100%" size="xs">
            <Progress.Track>
              <Progress.Range />
            </Progress.Track>
          </Progress.Root>
          {loading && (
            <Badge colorPalette="blue" size="xs" mt={1}>updating</Badge>
          )}
        </Stat.Root>
      </Card.Body>
    </Card.Root>
  )
})

GpuUtilizationCard.displayName = 'GpuUtilizationCard'

export const NodeOverviewCards = ({ cluster }: Props) => {
  const client = useClusterClient(cluster)

  const nodesQ = useClusterNodes({ cluster, client })
  const infoQ = useClusterNodesInfo({ cluster, client })
  const statesQ = useClusterNodesStates({ cluster, client })
  const gpuUtilQ = useClusterNodesProcessGpuUtil({ cluster, client })
  const lastProbeQ = useClusterNodesLastProbeTimestamp({ cluster, client })

  const nodes = (nodesQ.data ?? []) as string[]
  const infoMap = (infoQ.data ?? {}) as Record<string, NodeInfoResponse>
  const statesArr = (statesQ.data ?? []) as NodeStateResponse[]
  const lastProbeMap = (lastProbeQ.data ?? {}) as Record<string, Date | null>

  const totalNodes = nodes.length
  const totalGpus = useMemo(
    () => Object.values(infoMap).reduce((sum, n) => sum + (Array.isArray(n.cards) ? n.cards.length : 0), 0),
    [infoMap]
  )
  const idleNodes = useMemo(() => {
    // Deduplicate by node, keeping the latest entry per node
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
  }, [statesArr])

  // Calculate liveness stats using utility
  const { liveNodes, staleNodes, offlineNodes } = useMemo(
    () => getLivenessStats(lastProbeMap),
    [lastProbeMap]
  )

  // GPUs in use derived from latest process GPU util samples (> 0%)
  const gpusInUse = useMemo(() => {
    const utilData = (gpuUtilQ.data ?? undefined) as NodeSampleProcessGpuAccResponse | undefined
    let count = 0
    if (utilData && typeof utilData === 'object') {
      const utilDataMap = utilData as Record<string, Record<string, { gpu_util?: number }>>
      for (const nodeMap of Object.values(utilDataMap)) {
        for (const sample of Object.values(nodeMap)) {
          const util = sample.gpu_util ?? 0
          if (util > 0) count += 1
        }
      }
    }
    return count
  }, [gpuUtilQ.data])

  const gpuUtilPct = useMemo(
    () => calculateUtilization(gpusInUse, totalGpus),
    [gpusInUse, totalGpus]
  )
  const gpuColor = useMemo(
    () => totalGpus === 0 ? 'gray' : getUtilizationColorPalette(gpuUtilPct),
    [totalGpus, gpuUtilPct]
  )

  const nodesErrors = useMemo(() => {
    const errors = [
      ...(nodesQ.isError ? [getErrMsg(nodesQ.error)] : []),
      ...(statesQ.isError ? [getErrMsg(statesQ.error)] : [])
    ]
    return errors.length > 0 ? errors : undefined
  }, [nodesQ.isError, nodesQ.error, statesQ.isError, statesQ.error])

  const gpuErrors = useMemo(() => {
    const errors = [
      ...(gpuUtilQ.isError ? [getErrMsg(gpuUtilQ.error)] : []),
      ...(infoQ.isError ? [getErrMsg(infoQ.error)] : [])
    ]
    return errors.length > 0 ? errors : undefined
  }, [gpuUtilQ.isError, gpuUtilQ.error, infoQ.isError, infoQ.error])

  return (
    <VStack w="100%" align="start" gap={2}>
      <Text fontSize="lg" fontWeight="semibold">Nodes overview</Text>
      <SimpleGrid columns={GRID_COLUMNS} gap={3} w="100%" alignItems="stretch">
        <NodesCard
          totalNodes={totalNodes}
          idleNodes={idleNodes}
          loading={nodesQ.isLoading || statesQ.isLoading}
          errors={nodesErrors}
        />
        <NodeHealthCard
          liveNodes={liveNodes}
          staleNodes={staleNodes}
          offlineNodes={offlineNodes}
          loading={lastProbeQ.isLoading}
          error={lastProbeQ.isError ? lastProbeQ.error : undefined}
        />
        <GpuUtilizationCard
          gpusInUse={gpusInUse}
          totalGpus={totalGpus}
          gpuUtilPct={gpuUtilPct}
          gpuColor={gpuColor}
          loading={gpuUtilQ.isLoading || infoQ.isLoading}
          errors={gpuErrors}
        />
      </SimpleGrid>
    </VStack>
  )
}
