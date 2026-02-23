import { VStack, HStack, Text, Badge, Alert, Box, Tag, Spinner, Wrap, Collapsible } from '@chakra-ui/react'
import { useNavigate } from 'react-router'
import { LuChevronRight } from 'react-icons/lu'

import type { ErrorMessageResponse, NodeStateResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'
import { useClusterNodesLastProbeTimestamp } from '../../hooks/v2/useNodeQueries'
import { useClusterErrorMessages } from '../../hooks/v2/useClusterQueries'
import { useClusterOverviewContext } from '../../contexts/ClusterOverviewContext'

interface Props {
  cluster: string
}

export const ClusterHealthStatus = ({ cluster }: Props) => {
  const navigate = useNavigate()
  const client = useClusterClient(cluster)
  const { nodesStatesQuery: statesQuery, nodesQuery: nodesQ } = useClusterOverviewContext()

  const errorsQuery = useClusterErrorMessages({ cluster, client })
  const timestampsQuery = useClusterNodesLastProbeTimestamp({ cluster, client })

  const errorsMap = (errorsQuery.data ?? {}) as Record<string, ErrorMessageResponse>
  const errors = Object.values(errorsMap)
  const recentErrors = errors
    .sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime())
    .slice(0, 10)

  // Deduplicate to latest state per node, then count states
  const statesArr = (statesQuery.data ?? []) as NodeStateResponse[]
  const latestStateByNode = new Map<string, NodeStateResponse>()
  for (const s of statesArr) {
    const existing = latestStateByNode.get(s.node)
    if (!existing || new Date(s.time) > new Date(existing.time)) {
      latestStateByNode.set(s.node, s)
    }
  }

  const statesCounts: Record<string, number> = {}
  for (const s of latestStateByNode.values()) {
    if (Array.isArray(s.states)) {
      for (const state of s.states) {
        statesCounts[state] = (statesCounts[state] ?? 0) + 1
      }
    }
  }

  const timestampsMap = (timestampsQuery.data ?? {}) as Record<string, Date | null>
  const nodes = (nodesQ.data ?? []) as string[]

  // Calculate staleness (nodes with no recent data)
  const now = new Date()
  const staleThresholdMs = 5 * 60 * 1000 // 5 minutes
  const staleNodesList: Array<{ node: string; lastProbe: Date | null; ageMinutes: number }> = []
  const missingNodesList: string[] = []

  for (const node of nodes) {
    const ts = timestampsMap[node]
    if (!ts) {
      missingNodesList.push(node)
    } else {
      const age = now.getTime() - new Date(ts).getTime()
      if (age > staleThresholdMs) {
        staleNodesList.push({
          node,
          lastProbe: new Date(ts),
          ageMinutes: Math.round(age / (60 * 1000))
        })
      }
    }
  }

  const staleNodes = staleNodesList.length
  const missingNodes = missingNodesList.length

  const isLoading = errorsQuery.isLoading || statesQuery.isLoading || timestampsQuery.isLoading || nodesQ.isLoading
  const hasError = errorsQuery.isError || statesQuery.isError || timestampsQuery.isError || nodesQ.isError

  const getStateColor = (state: string): string => {
    const s = state.toUpperCase()
    if (s.includes('DOWN') || s.includes('DRAIN') || s.includes('FAIL') || s.includes('INVALID')) return 'red'
    if (s.includes('IDLE')) return 'green'
    if (s.includes('ALLOC') || s.includes('MIX')) return 'blue'
    if (s.includes('COMPLETING') || s.includes('PLANNED') || s.includes('RESERVED') || s.includes('MAINTENANCE')) return 'yellow'
    return 'gray'
  }

  const sortedStates = Object.entries(statesCounts).sort((a, b) => b[1] - a[1])

  return (
    <VStack w="100%" align="start" gap={3}>
      <HStack justify="space-between" w="100%">
        <Text fontSize="lg" fontWeight="semibold">Cluster Health</Text>
        {isLoading && <Spinner size="sm" />}
      </HStack>

      {hasError && (
        <Alert.Root status="warning">
          <Alert.Indicator />
          <Alert.Description>Some health data could not be loaded.</Alert.Description>
        </Alert.Root>
      )}

      {/* Node States — horizontal wrapped tags */}
      <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white" w="100%">
        <VStack align="start" gap={2} w="100%">
          <Text fontSize="sm" fontWeight="semibold" color="gray.700">Node States</Text>
          {sortedStates.length === 0 ? (
            <Text fontSize="xs" color="gray.500">No state data available</Text>
          ) : (
            <Wrap gap={2}>
              {sortedStates.map(([state, count]) => (
                <Tag.Root key={state} size="sm" colorPalette={getStateColor(state)}>
                  <Tag.Label>{state}: {count}</Tag.Label>
                </Tag.Root>
              ))}
            </Wrap>
          )}

          {/* Freshness inline */}
          <HStack gap={4} w="100%" flexWrap="wrap">
            {staleNodes > 0 ? (
              <Collapsible.Root>
                <Collapsible.Trigger asChild>
                  <HStack gap={1} cursor="pointer" _hover={{ color: 'orange.700' }}>
                    <Text fontSize="xs" color="gray.500">Stale (&gt;5m):</Text>
                    <Badge colorPalette="orange" size="xs">{staleNodes}</Badge>
                    <Text fontSize="2xs" color="gray.400">(click)</Text>
                  </HStack>
                </Collapsible.Trigger>
                <Collapsible.Content>
                  <VStack align="start" mt={2} gap={1} maxH="150px" overflowY="auto" pl={2} borderLeftWidth="2px" borderColor="orange.200">
                    {staleNodesList.map(({ node, ageMinutes }) => (
                      <HStack key={node} gap={2} fontSize="xs">
                        <Text
                          fontWeight="medium"
                          color="orange.700"
                          cursor="pointer"
                          _hover={{ textDecoration: 'underline' }}
                          onClick={() => navigate(`/v2/${cluster}/nodes/${node}`)}
                        >
                          {node}
                        </Text>
                        <Text color="gray.500">
                          {ageMinutes}m ago
                        </Text>
                      </HStack>
                    ))}
                  </VStack>
                </Collapsible.Content>
              </Collapsible.Root>
            ) : (
              <Text fontSize="xs" color="gray.500">No stale nodes</Text>
            )}
            {missingNodes > 0 && (
              <Collapsible.Root>
                <Collapsible.Trigger asChild>
                  <HStack gap={1} cursor="pointer" _hover={{ color: 'red.700' }}>
                    <Text fontSize="xs" color="gray.500">Non-reporting:</Text>
                    <Badge colorPalette="red" size="xs">{missingNodes}</Badge>
                    <Text fontSize="2xs" color="gray.400">(click)</Text>
                  </HStack>
                </Collapsible.Trigger>
                <Collapsible.Content>
                  <VStack align="start" mt={2} gap={1} maxH="150px" overflowY="auto" pl={2} borderLeftWidth="2px" borderColor="red.200">
                    {missingNodesList.map(node => (
                      <Text
                        key={node}
                        fontSize="xs"
                        fontWeight="medium"
                        color="red.700"
                        cursor="pointer"
                        _hover={{ textDecoration: 'underline' }}
                        onClick={() => navigate(`/v2/${cluster}/nodes/${node}`)}
                      >
                        {node}
                      </Text>
                    ))}
                  </VStack>
                </Collapsible.Content>
              </Collapsible.Root>
            )}
          </HStack>
        </VStack>
      </Box>

      {/* Recent Errors — collapsible, closed by default when empty */}
      <Collapsible.Root defaultOpen={recentErrors.length > 0} w="100%">
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" bg="white" w="100%">
          <Collapsible.Trigger asChild>
            <HStack
              justify="space-between"
              w="100%"
              p={3}
              cursor="pointer"
              _hover={{ bg: 'gray.50' }}
              rounded="md"
            >
              <HStack gap={2}>
                <Collapsible.Indicator
                  transition="transform 0.2s"
                  _open={{ transform: 'rotate(90deg)' }}
                >
                  <LuChevronRight />
                </Collapsible.Indicator>
                <Text fontSize="sm" fontWeight="semibold" color="gray.700">Recent Errors</Text>
              </HStack>
              <Badge colorPalette={recentErrors.length > 0 ? 'red' : 'green'} size="sm">
                {recentErrors.length}
              </Badge>
            </HStack>
          </Collapsible.Trigger>
          <Collapsible.Content>
            <VStack align="start" gap={1} px={3} pb={3} maxH="300px" overflowY="auto">
              {recentErrors.length === 0 ? (
                <Text fontSize="xs" color="gray.500">No recent errors</Text>
              ) : (
                recentErrors.map((err, idx) => (
                  <Box key={idx} w="100%" borderBottomWidth="1px" borderColor="gray.100" pb={1}>
                    <HStack justify="space-between" align="start" w="100%">
                      <Text
                        fontSize="xs"
                        fontWeight="medium"
                        color="gray.700"
                        truncate
                        cursor="pointer"
                        _hover={{ color: 'blue.600', textDecoration: 'underline' }}
                        onClick={() => navigate(`/v2/${cluster}/nodes/${err.node}`)}
                      >
                        {err.node}
                      </Text>
                      <Text fontSize="2xs" color="gray.500" whiteSpace="nowrap">
                        {new Date(err.time).toLocaleTimeString()}
                      </Text>
                    </HStack>
                    <Text fontSize="2xs" color="gray.600" lineClamp={2}>{err.details}</Text>
                  </Box>
                ))
              )}
            </VStack>
          </Collapsible.Content>
        </Box>
      </Collapsible.Root>
    </VStack>
  )
}
