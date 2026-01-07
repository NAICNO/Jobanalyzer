import { VStack, HStack, Text, Badge, Alert, Spinner, SimpleGrid, Box, Stat, Tag } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'

import {
  getClusterByClusterErrorMessagesOptions,
  getClusterByClusterNodesStatesOptions,
  getClusterByClusterNodesLastProbeTimestampOptions,
  getClusterByClusterNodesOptions
} from '../../client/@tanstack/react-query.gen'
import type { ErrorMessageResponse, NodeStateResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'

interface Props {
  cluster: string
}

export const ClusterHealthStatus = ({ cluster }: Props) => {
  const client = useClusterClient(cluster)
  
  if (!client) {
    return <Spinner />
  }
  
  const baseURL = client.getConfig().baseURL
  
  const errorsQuery = useQuery({
    ...getClusterByClusterErrorMessagesOptions({ path: { cluster }, client, baseURL }),
    enabled: !!cluster,
  })
  const statesQuery = useQuery({
    ...getClusterByClusterNodesStatesOptions({ path: { cluster }, client, baseURL }),
    enabled: !!cluster,
  })
  const timestampsQuery = useQuery({
    ...getClusterByClusterNodesLastProbeTimestampOptions({ path: { cluster }, client, baseURL }),
    enabled: !!cluster,
  })
  const nodesQ = useQuery({
    ...getClusterByClusterNodesOptions({ path: { cluster }, client, baseURL }),
    enabled: !!cluster,
  })

  const errorsMap = (errorsQuery.data ?? {}) as Record<string, ErrorMessageResponse>
  const errors = Object.values(errorsMap)
  const recentErrors = errors
    .sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime())
    .slice(0, 10)

  const statesArr = (statesQuery.data ?? []) as NodeStateResponse[]
  const statesCounts: Record<string, number> = {}
  for (const s of statesArr) {
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
  let staleNodes = 0
  let missingNodes = 0
  
  for (const node of nodes) {
    const ts = timestampsMap[node]
    if (!ts) {
      missingNodes += 1
    } else {
      const age = now.getTime() - new Date(ts).getTime()
      if (age > staleThresholdMs) {
        staleNodes += 1
      }
    }
  }

  const isLoading = errorsQuery.isLoading || statesQuery.isLoading || timestampsQuery.isLoading || nodesQ.isLoading
  const hasError = errorsQuery.isError || statesQuery.isError || timestampsQuery.isError || nodesQ.isError

  const getStateColor = (state: string): string => {
    const s = state.toUpperCase()
    if (s.includes('DOWN') || s.includes('DRAIN') || s.includes('FAIL')) return 'red'
    if (s.includes('IDLE')) return 'green'
    if (s.includes('ALLOC') || s.includes('MIX')) return 'blue'
    return 'gray'
  }

  return (
    <VStack w="100%" align="start" gap={4}>
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

      <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} gap={4} w="100%">
        {/* Recent Error Messages */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <HStack justify="space-between" w="100%">
              <Text fontSize="sm" fontWeight="semibold" color="gray.700">Recent Errors</Text>
              <Badge colorPalette={recentErrors.length > 0 ? 'red' : 'green'}>
                {recentErrors.length}
              </Badge>
            </HStack>
            <VStack align="start" gap={1} w="100%" maxH="200px" overflowY="auto">
              {recentErrors.length === 0 ? (
                <Text fontSize="xs" color="gray.500">No recent errors</Text>
              ) : (
                recentErrors.map((err, idx) => (
                  <Box key={idx} w="100%" borderBottomWidth="1px" borderColor="gray.100" pb={1}>
                    <HStack justify="space-between" align="start" w="100%">
                      <Text fontSize="xs" fontWeight="medium" color="gray.700" truncate>{err.node}</Text>
                      <Text fontSize="2xs" color="gray.500" whiteSpace="nowrap">
                        {new Date(err.time).toLocaleTimeString()}
                      </Text>
                    </HStack>
                    <Text fontSize="2xs" color="gray.600" lineClamp={2}>{err.details}</Text>
                  </Box>
                ))
              )}
            </VStack>
          </VStack>
        </Box>

        {/* Node States Breakdown */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Node States</Text>
            <VStack align="start" gap={1} w="100%" maxH="200px" overflowY="auto">
              {Object.entries(statesCounts).length === 0 ? (
                <Text fontSize="xs" color="gray.500">No state data available</Text>
              ) : (
                Object.entries(statesCounts)
                  .sort((a, b) => b[1] - a[1])
                  .map(([state, count]) => (
                    <HStack key={state} justify="space-between" w="100%">
                      <Tag.Root size="sm" colorPalette={getStateColor(state)}>
                        <Tag.Label>{state}</Tag.Label>
                      </Tag.Root>
                      <Text fontSize="sm" fontWeight="semibold">{count}</Text>
                    </HStack>
                  ))
              )}
            </VStack>
          </VStack>
        </Box>

        {/* Staleness Indicators */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={3} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Data Freshness</Text>
            <Stat.Root>
              <Stat.Label fontSize="xs" color="gray.600">Stale Nodes (&gt;5m)</Stat.Label>
              <HStack gap={2}>
                <Stat.ValueText fontSize="xl" fontWeight="semibold">{staleNodes}</Stat.ValueText>
                {staleNodes > 0 && <Badge colorPalette="orange">warning</Badge>}
              </HStack>
              <Stat.HelpText fontSize="2xs" color="gray.500">
                No updates in last 5 minutes
              </Stat.HelpText>
            </Stat.Root>
            <Stat.Root>
              <Stat.Label fontSize="xs" color="gray.600">Non-reporting Nodes</Stat.Label>
              <HStack gap={2}>
                <Stat.ValueText fontSize="xl" fontWeight="semibold">{missingNodes}</Stat.ValueText>
                {missingNodes > 0 && <Badge colorPalette="red">error</Badge>}
              </HStack>
              <Stat.HelpText fontSize="2xs" color="gray.500">
                No data ever received
              </Stat.HelpText>
            </Stat.Root>
          </VStack>
        </Box>
      </SimpleGrid>
    </VStack>
  )
}
