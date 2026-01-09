import { VStack, Text, SimpleGrid, Box, Table, Stat, Spinner } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'

import { getClusterByClusterPartitionsOptions } from '../../client/@tanstack/react-query.gen'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import type { PartitionResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'

interface Props {
  cluster: string
}

export const ClusterWaitTimeAnalysis = ({ cluster }: Props) => {
  const client = useClusterClient(cluster)
  
  if (!client) {
    return <Spinner />
  }
  
  const partitionsQ = useQuery({
    ...getClusterByClusterPartitionsOptions({ path: { cluster }, client }),
    enabled: !!cluster,
  })

  const partitionsMap = (partitionsQ.data ?? {}) as Record<string, PartitionResponse>
  const partitions = Object.values(partitionsMap)

  // Collect wait time statistics
  const waitTimeStats: Array<{
    partition: string
    latestWaitTimeSec: number
    latestWaitTimeMin: number
    pendingCount: number
    oldestPendingMin: number
  }> = []

  const now = Date.now()

  for (const p of partitions) {
    const pendingCount = p.jobs_pending?.length ?? 0
    const latestWaitTimeSec = p.running_latest_wait_time ?? 0
    const latestWaitTimeMin = Math.round(latestWaitTimeSec / 60)

    // Calculate how long the oldest pending job has been waiting
    let oldestPendingMin = 0
    if (p.pending_max_submit_time) {
      const submitTime = new Date(p.pending_max_submit_time).getTime()
      oldestPendingMin = Math.round((now - submitTime) / (1000 * 60))
    }

    waitTimeStats.push({
      partition: p.name ?? 'Unknown',
      latestWaitTimeSec,
      latestWaitTimeMin,
      pendingCount,
      oldestPendingMin,
    })
  }

  // Sort by latest wait time
  const sortedByWaitTime = waitTimeStats
    .filter(s => s.latestWaitTimeMin > 0)
    .sort((a, b) => b.latestWaitTimeMin - a.latestWaitTimeMin)

  // Sort by oldest pending
  const sortedByOldestPending = waitTimeStats
    .filter(s => s.oldestPendingMin > 0)
    .sort((a, b) => b.oldestPendingMin - a.oldestPendingMin)
    .slice(0, 10)

  // Calculate overall statistics
  const avgWaitTime = sortedByWaitTime.length > 0
    ? Math.round(sortedByWaitTime.reduce((sum, s) => sum + s.latestWaitTimeMin, 0) / sortedByWaitTime.length)
    : 0

  const maxWaitTime = sortedByWaitTime.length > 0
    ? sortedByWaitTime[0].latestWaitTimeMin
    : 0

  const totalPending = waitTimeStats.reduce((sum, s) => sum + s.pendingCount, 0)

  const maxOldestPending = sortedByOldestPending.length > 0
    ? sortedByOldestPending[0].oldestPendingMin
    : 0

  const formatTime = (minutes: number) => {
    if (minutes < 60) return `${minutes}m`
    const hours = Math.floor(minutes / 60)
    const mins = minutes % 60
    return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`
  }

  // Prepare data for bar chart
  const chartData = sortedByWaitTime.slice(0, 10).map(s => ({
    partition: s.partition,
    waitTime: s.latestWaitTimeMin,
  }))

  return (
    <VStack w="100%" align="start" gap={4}>
      <Text fontSize="lg" fontWeight="semibold">Queue Wait Time Analysis</Text>

      {/* Summary Cards */}
      <SimpleGrid columns={{ base: 2, md: 4 }} gap={3} w="100%">
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Total Pending</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold" color="orange.600">
              {totalPending}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Avg Wait Time</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold">
              {avgWaitTime > 0 ? formatTime(avgWaitTime) : 'N/A'}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Max Wait Time</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold" color={maxWaitTime > 60 ? 'red.600' : 'inherit'}>
              {maxWaitTime > 0 ? formatTime(maxWaitTime) : 'N/A'}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Longest Pending</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold" color={maxOldestPending > 120 ? 'red.600' : 'inherit'}>
              {maxOldestPending > 0 ? formatTime(maxOldestPending) : 'N/A'}
            </Stat.ValueText>
          </Stat.Root>
        </Box>
      </SimpleGrid>

      <SimpleGrid columns={{ base: 1, lg: 2 }} gap={4} w="100%">
        {/* Wait Time Chart */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Latest Wait Times by Partition</Text>
            {partitionsQ.isLoading ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg" />
              </Box>
            ) : chartData.length === 0 ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Text fontSize="sm" color="gray.500">No wait time data available</Text>
              </Box>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis 
                    dataKey="partition" 
                    tick={{ fontSize: 10 }}
                    angle={-45}
                    textAnchor="end"
                    height={80}
                  />
                  <YAxis 
                    tick={{ fontSize: 10 }}
                    label={{ value: 'Wait Time (min)', angle: -90, position: 'insideLeft', style: { fontSize: 10 } }}
                  />
                  <Tooltip 
                    contentStyle={{ fontSize: 12 }}
                    formatter={(value: number) => `${formatTime(value)}`}
                  />
                  <Legend wrapperStyle={{ fontSize: 12 }} />
                  <Bar 
                    dataKey="waitTime" 
                    fill="#ed8936" 
                    name="Wait Time"
                  />
                </BarChart>
              </ResponsiveContainer>
            )}
          </VStack>
        </Box>

        {/* Oldest Pending Jobs Table */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Partitions with Longest Pending Jobs</Text>
            <Box w="100%" maxH="300px" overflowY="auto">
              <Table.Root size="sm" variant="outline">
                <Table.Header>
                  <Table.Row bg="gray.50">
                    <Table.ColumnHeader fontSize="xs">Partition</Table.ColumnHeader>
                    <Table.ColumnHeader fontSize="xs" textAlign="center">Pending</Table.ColumnHeader>
                    <Table.ColumnHeader fontSize="xs" textAlign="center">Waiting Since</Table.ColumnHeader>
                  </Table.Row>
                </Table.Header>
                <Table.Body>
                  {sortedByOldestPending.length === 0 ? (
                    <Table.Row>
                      <Table.Cell colSpan={3}>
                        <Text fontSize="xs" color="gray.500" textAlign="center">No pending jobs</Text>
                      </Table.Cell>
                    </Table.Row>
                  ) : (
                    sortedByOldestPending.map((s) => (
                      <Table.Row key={s.partition}>
                        <Table.Cell fontSize="xs" fontWeight="medium">{s.partition}</Table.Cell>
                        <Table.Cell fontSize="xs" textAlign="center">{s.pendingCount}</Table.Cell>
                        <Table.Cell 
                          fontSize="xs" 
                          textAlign="center"
                          color={s.oldestPendingMin > 120 ? 'red.600' : 'inherit'}
                          fontWeight={s.oldestPendingMin > 120 ? 'bold' : 'normal'}
                        >
                          {formatTime(s.oldestPendingMin)}
                        </Table.Cell>
                      </Table.Row>
                    ))
                  )}
                </Table.Body>
              </Table.Root>
            </Box>
          </VStack>
        </Box>
      </SimpleGrid>
    </VStack>
  )
}
