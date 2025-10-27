import { VStack, HStack, Text, SimpleGrid, Box, Table, Tag, Stat, Badge } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'

import { getClusterByClusterPartitionsOptions } from '../../client/@tanstack/react-query.gen'
import type { PartitionResponseOutput } from '../../client'

interface Props {
  cluster: string
}

export const ClusterQueueActivity = ({ cluster }: Props) => {
  const partitionsQ = useQuery({
    ...getClusterByClusterPartitionsOptions({ path: { cluster } }),
    enabled: !!cluster,
  })

  const partitionsMap = (partitionsQ.data ?? {}) as Record<string, PartitionResponseOutput>
  const partitions = Object.values(partitionsMap)

  // Aggregate job counts across all partitions
  let totalRunning = 0
  let totalPending = 0
  const jobsByUser: Record<string, { running: number; pending: number }> = {}
  const jobsByPartition: Array<{ partition: string; running: number; pending: number }> = []

  for (const partition of partitions) {
    const runningJobs = partition.jobs_running ?? []
    const pendingJobs = partition.jobs_pending ?? []
    
    totalRunning += runningJobs.length
    totalPending += pendingJobs.length

    jobsByPartition.push({
      partition: partition.name ?? 'Unknown',
      running: runningJobs.length,
      pending: pendingJobs.length
    })

    // Track by user
    for (const job of runningJobs) {
      const user = job.user_name ?? 'Unknown'
      if (!jobsByUser[user]) {
        jobsByUser[user] = { running: 0, pending: 0 }
      }
      jobsByUser[user].running++
    }

    for (const job of pendingJobs) {
      const user = job.user_name ?? 'Unknown'
      if (!jobsByUser[user]) {
        jobsByUser[user] = { running: 0, pending: 0 }
      }
      jobsByUser[user].pending++
    }
  }

  // Sort users by total jobs
  const topUsers = Object.entries(jobsByUser)
    .map(([user, counts]) => ({
      user,
      running: counts.running,
      pending: counts.pending,
      total: counts.running + counts.pending
    }))
    .sort((a, b) => b.total - a.total)
    .slice(0, 10)

  // Sort partitions by total jobs
  const sortedPartitions = jobsByPartition
    .map(p => ({ ...p, total: p.running + p.pending }))
    .sort((a, b) => b.total - a.total)
    .slice(0, 10)

  // Calculate trends (mock for now - would need historical data)
  const runningTrend = totalRunning > 0 ? 'up' : 'neutral'
  const pendingTrend = totalPending > 10 ? 'up' : totalPending > 0 ? 'neutral' : 'down'

  return (
    <VStack w="100%" align="start" gap={4}>
      <Text fontSize="lg" fontWeight="semibold">Queue Activity</Text>

      {/* Summary Cards */}
      <SimpleGrid columns={{ base: 2, md: 4 }} gap={3} w="100%">
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Running Jobs</Stat.Label>
            <HStack gap={2}>
              <Stat.ValueText fontSize="2xl" fontWeight="bold" color="green.600">
                {totalRunning}
              </Stat.ValueText>
              {runningTrend === 'up' && (
                <Stat.UpIndicator>
                  <Badge colorPalette="green" size="sm">Active</Badge>
                </Stat.UpIndicator>
              )}
            </HStack>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Pending Jobs</Stat.Label>
            <HStack gap={2}>
              <Stat.ValueText fontSize="2xl" fontWeight="bold" color="orange.600">
                {totalPending}
              </Stat.ValueText>
              {pendingTrend === 'up' && (
                <Stat.UpIndicator>
                  <Badge colorPalette="orange" size="sm">High</Badge>
                </Stat.UpIndicator>
              )}
            </HStack>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Total Jobs</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold">
              {totalRunning + totalPending}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Active Users</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold">
              {Object.keys(jobsByUser).length}
            </Stat.ValueText>
          </Stat.Root>
        </Box>
      </SimpleGrid>

      {/* Jobs by Partition and User Tables */}
      <SimpleGrid columns={{ base: 1, lg: 2 }} gap={4} w="100%">
        {/* Jobs by Partition */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Jobs by Partition</Text>
            <Box w="100%" maxH="300px" overflowY="auto">
              <Table.Root size="sm" variant="outline">
                <Table.Header>
                  <Table.Row bg="gray.50">
                    <Table.ColumnHeader fontSize="xs">Partition</Table.ColumnHeader>
                    <Table.ColumnHeader fontSize="xs" textAlign="center">Running</Table.ColumnHeader>
                    <Table.ColumnHeader fontSize="xs" textAlign="center">Pending</Table.ColumnHeader>
                    <Table.ColumnHeader fontSize="xs" textAlign="center">Total</Table.ColumnHeader>
                  </Table.Row>
                </Table.Header>
                <Table.Body>
                  {sortedPartitions.length === 0 ? (
                    <Table.Row>
                      <Table.Cell colSpan={4}>
                        <Text fontSize="xs" color="gray.500" textAlign="center">No partition data</Text>
                      </Table.Cell>
                    </Table.Row>
                  ) : (
                    sortedPartitions.map((p) => (
                      <Table.Row key={p.partition}>
                        <Table.Cell fontSize="xs" fontWeight="medium">{p.partition}</Table.Cell>
                        <Table.Cell fontSize="xs" textAlign="center">
                          <Tag.Root size="sm" colorPalette="green">
                            <Tag.Label>{p.running}</Tag.Label>
                          </Tag.Root>
                        </Table.Cell>
                        <Table.Cell fontSize="xs" textAlign="center">
                          <Tag.Root size="sm" colorPalette="orange">
                            <Tag.Label>{p.pending}</Tag.Label>
                          </Tag.Root>
                        </Table.Cell>
                        <Table.Cell fontSize="xs" textAlign="center" fontWeight="semibold">
                          {p.total}
                        </Table.Cell>
                      </Table.Row>
                    ))
                  )}
                </Table.Body>
              </Table.Root>
            </Box>
          </VStack>
        </Box>

        {/* Jobs by User */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Top Users (by total jobs)</Text>
            <Box w="100%" maxH="300px" overflowY="auto">
              <Table.Root size="sm" variant="outline">
                <Table.Header>
                  <Table.Row bg="gray.50">
                    <Table.ColumnHeader fontSize="xs">User</Table.ColumnHeader>
                    <Table.ColumnHeader fontSize="xs" textAlign="center">Running</Table.ColumnHeader>
                    <Table.ColumnHeader fontSize="xs" textAlign="center">Pending</Table.ColumnHeader>
                    <Table.ColumnHeader fontSize="xs" textAlign="center">Total</Table.ColumnHeader>
                  </Table.Row>
                </Table.Header>
                <Table.Body>
                  {topUsers.length === 0 ? (
                    <Table.Row>
                      <Table.Cell colSpan={4}>
                        <Text fontSize="xs" color="gray.500" textAlign="center">No user data</Text>
                      </Table.Cell>
                    </Table.Row>
                  ) : (
                    topUsers.map((u) => (
                      <Table.Row key={u.user}>
                        <Table.Cell fontSize="xs" fontWeight="medium">{u.user}</Table.Cell>
                        <Table.Cell fontSize="xs" textAlign="center">
                          <Tag.Root size="sm" colorPalette="green">
                            <Tag.Label>{u.running}</Tag.Label>
                          </Tag.Root>
                        </Table.Cell>
                        <Table.Cell fontSize="xs" textAlign="center">
                          <Tag.Root size="sm" colorPalette="orange">
                            <Tag.Label>{u.pending}</Tag.Label>
                          </Tag.Root>
                        </Table.Cell>
                        <Table.Cell fontSize="xs" textAlign="center" fontWeight="semibold">
                          {u.total}
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
