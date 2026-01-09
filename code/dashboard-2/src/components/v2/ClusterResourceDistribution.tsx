import { VStack, HStack, Text, SimpleGrid, Box, Progress, Stat, Tag, Spinner } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'

import {
  getClusterByClusterPartitionsOptions,
  getClusterByClusterNodesInfoOptions,
  getClusterByClusterNodesStatesOptions
} from '../../client/@tanstack/react-query.gen'
import type { PartitionResponse, NodeInfoResponse, NodeStateResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'

interface Props {
  cluster: string
}

export const ClusterResourceDistribution = ({ cluster }: Props) => {
  const client = useClusterClient(cluster)
  
  if (!client) {
    return <Spinner />
  }
  
  const partitionsQ = useQuery({
    ...getClusterByClusterPartitionsOptions({ path: { cluster }, client }),
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

  const partitionsMap = (partitionsQ.data ?? {}) as Record<string, PartitionResponse>
  const partitions = Object.values(partitionsMap).sort((a, b) => (b.total_gpus ?? 0) - (a.total_gpus ?? 0))

  const infoMap = (infoQ.data ?? {}) as Record<string, NodeInfoResponse>
  const statesArr = (statesQ.data ?? []) as NodeStateResponse[]

  // Calculate GPU types distribution
  const gpuTypes: Record<string, number> = {}
  for (const node of Object.values(infoMap)) {
    if (Array.isArray(node.cards)) {
      for (const card of node.cards) {
        const model = card.model ?? 'Unknown'
        gpuTypes[model] = (gpuTypes[model] ?? 0) + 1
      }
    }
  }

  // Get top nodes by GPU count
  const nodesByGpuCount = Object.entries(infoMap)
    .map(([name, info]) => ({
      name,
      gpuCount: Array.isArray(info.cards) ? info.cards.length : 0,
      cpuCount: (info.sockets ?? 0) * (info.cores_per_socket ?? 0) * (info.threads_per_core ?? 0),
      memoryGb: Math.round((info.memory ?? 0) / (1024 * 1024))
    }))
    .sort((a, b) => b.gpuCount - a.gpuCount)
    .slice(0, 10)

  // Node state summary
  const nodeStateMap = new Map<string, string[]>()
  for (const s of statesArr) {
    nodeStateMap.set(s.node ?? '', s.states ?? [])
  }

  return (
    <VStack w="100%" align="start" gap={4}>
      <Text fontSize="lg" fontWeight="semibold">Resource Distribution</Text>

      <SimpleGrid columns={{ base: 1, lg: 2 }} gap={4} w="100%">
        {/* Partition Resources */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Resources by Partition</Text>
            <VStack align="start" gap={2} w="100%" maxH="300px" overflowY="auto">
              {partitions.length === 0 ? (
                <Text fontSize="xs" color="gray.500">No partition data available</Text>
              ) : (
                partitions.map((partition) => {
                  const cpuTotal = partition.total_cpus ?? 0
                  const gpuTotal = partition.total_gpus ?? 0
                  const gpuInUse = partition.gpus_in_use?.length ?? 0
                  const gpuUtilPct = gpuTotal > 0 ? Math.round((gpuInUse / gpuTotal) * 100) : 0
                  const gpuColor = gpuUtilPct > 90 ? 'red' : gpuUtilPct > 50 ? 'yellow' : 'green'

                  return (
                    <Box key={partition.name} w="100%" borderBottomWidth="1px" borderColor="gray.100" pb={2}>
                      <HStack justify="space-between" w="100%" mb={1}>
                        <Text fontSize="xs" fontWeight="medium" color="gray.700">{partition.name}</Text>
                        <HStack gap={1}>
                          <Tag.Root size="sm" variant="subtle">
                            <Tag.Label>{cpuTotal} CPUs</Tag.Label>
                          </Tag.Root>
                          <Tag.Root size="sm" variant="subtle">
                            <Tag.Label>{gpuTotal} GPUs</Tag.Label>
                          </Tag.Root>
                        </HStack>
                      </HStack>
                      {gpuTotal > 0 && (
                        <HStack gap={2} w="100%">
                          <Text fontSize="2xs" color="gray.600" minW="60px">GPU: {gpuInUse}/{gpuTotal}</Text>
                          <Progress.Root value={gpuUtilPct} max={100} colorPalette={gpuColor} flex={1} size="xs">
                            <Progress.Track>
                              <Progress.Range />
                            </Progress.Track>
                          </Progress.Root>
                          <Text fontSize="2xs" color="gray.600" minW="35px" textAlign="right">{gpuUtilPct}%</Text>
                        </HStack>
                      )}
                    </Box>
                  )
                })
              )}
            </VStack>
          </VStack>
        </Box>

        {/* GPU Types Distribution */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">GPU Types</Text>
            <VStack align="start" gap={1} w="100%" maxH="300px" overflowY="auto">
              {Object.entries(gpuTypes).length === 0 ? (
                <Text fontSize="xs" color="gray.500">No GPU data available</Text>
              ) : (
                Object.entries(gpuTypes)
                  .sort((a, b) => b[1] - a[1])
                  .map(([model, count]) => (
                    <HStack key={model} justify="space-between" w="100%">
                      <Text fontSize="xs" color="gray.700" truncate flex={1}>{model}</Text>
                      <Stat.Root>
                        <Stat.ValueText fontSize="md" fontWeight="semibold">{count}</Stat.ValueText>
                      </Stat.Root>
                    </HStack>
                  ))
              )}
            </VStack>
          </VStack>
        </Box>

        {/* Top Nodes by GPU Count */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Top Nodes by GPU Count</Text>
            <VStack align="start" gap={1} w="100%" maxH="300px" overflowY="auto">
              {nodesByGpuCount.length === 0 ? (
                <Text fontSize="xs" color="gray.500">No node data available</Text>
              ) : (
                nodesByGpuCount.map((node) => {
                  const states = nodeStateMap.get(node.name) ?? []
                  const stateColor = states.some(s => s.includes('IDLE')) ? 'green' : states.some(s => s.includes('DOWN') || s.includes('DRAIN')) ? 'red' : 'blue'
                  
                  return (
                    <HStack key={node.name} justify="space-between" w="100%" borderBottomWidth="1px" borderColor="gray.100" pb={1}>
                      <VStack align="start" gap={0} flex={1}>
                        <HStack gap={1}>
                          <Text fontSize="xs" fontWeight="medium" color="gray.700">{node.name}</Text>
                          {states.length > 0 && (
                            <Tag.Root size="sm" colorPalette={stateColor}>
                              <Tag.Label>{states[0]}</Tag.Label>
                            </Tag.Root>
                          )}
                        </HStack>
                        <Text fontSize="2xs" color="gray.500">
                          {node.cpuCount} CPUs · {node.memoryGb}GB RAM
                        </Text>
                      </VStack>
                      <Tag.Root size="sm" colorPalette="blue">
                        <Tag.Label>{node.gpuCount} GPUs</Tag.Label>
                      </Tag.Root>
                    </HStack>
                  )
                })
              )}
            </VStack>
          </VStack>
        </Box>

        {/* Node Memory Distribution (Top 10) */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Top Nodes by Memory</Text>
            <VStack align="start" gap={1} w="100%" maxH="300px" overflowY="auto">
              {Object.entries(infoMap).length === 0 ? (
                <Text fontSize="xs" color="gray.500">No memory data available</Text>
              ) : (
                Object.entries(infoMap)
                  .map(([name, info]) => ({
                    name,
                    memoryGb: Math.round((info.memory ?? 0) / (1024 * 1024)),
                    cpuCount: (info.sockets ?? 0) * (info.cores_per_socket ?? 0) * (info.threads_per_core ?? 0)
                  }))
                  .sort((a, b) => b.memoryGb - a.memoryGb)
                  .slice(0, 10)
                  .map((node) => (
                    <HStack key={node.name} justify="space-between" w="100%">
                      <VStack align="start" gap={0} flex={1}>
                        <Text fontSize="xs" fontWeight="medium" color="gray.700">{node.name}</Text>
                        <Text fontSize="2xs" color="gray.500">{node.cpuCount} CPUs</Text>
                      </VStack>
                      <Tag.Root size="sm" colorPalette="purple">
                        <Tag.Label>{node.memoryGb} GB</Tag.Label>
                      </Tag.Root>
                    </HStack>
                  ))
              )}
            </VStack>
          </VStack>
        </Box>
      </SimpleGrid>
    </VStack>
  )
}
