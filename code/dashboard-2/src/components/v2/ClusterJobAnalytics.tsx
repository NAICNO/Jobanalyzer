import { VStack, Text, SimpleGrid, Box, Table, Tag, Stat, Spinner } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'

import { getClusterByClusterJobsOptions } from '../../client/@tanstack/react-query.gen'
import type { JobsResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'
import { getJobStateColor } from '../../util/formatters'

interface Props {
  cluster: string
}

export const ClusterJobAnalytics = ({ cluster }: Props) => {
  const client = useClusterClient(cluster)
  const now = Math.floor(Date.now() / 1000)
  const last24h = now - (24 * 60 * 60)

  if (!client) {
    return <Spinner />
  }

  // Get recent jobs (last 24 hours)
  const jobsQ = useQuery({
    ...getClusterByClusterJobsOptions({
      path: { cluster },
      query: {
        start_time_in_s: last24h,
        end_time_in_s: now,
      },
      client,
    }),
    enabled: !!cluster,
  })

  const jobsData = (jobsQ.data as JobsResponse) ?? { jobs: [] }
  const jobs = jobsData.jobs ?? []

  // Calculate statistics
  const completedJobs = jobs.filter(j => j.job_state === 'COMPLETED')
  const failedJobs = jobs.filter(j => j.job_state === 'FAILED')
  const cancelledJobs = jobs.filter(j => j.job_state === 'CANCELLED')
  const timeoutJobs = jobs.filter(j => j.job_state === 'TIMEOUT')
  
  const totalFinished = completedJobs.length + failedJobs.length + cancelledJobs.length + timeoutJobs.length
  const successRate = totalFinished > 0 ? Math.round((completedJobs.length / totalFinished) * 100) : 0

  // Calculate completion times for completed jobs
  const completionTimes: number[] = []
  for (const job of completedJobs) {
    if (job.start_time && job.end_time) {
      const startMs = new Date(job.start_time).getTime()
      const endMs = new Date(job.end_time).getTime()
      const durationMin = (endMs - startMs) / (1000 * 60)
      if (durationMin > 0) {
        completionTimes.push(durationMin)
      }
    }
  }

  const avgCompletionTime = completionTimes.length > 0
    ? Math.round(completionTimes.reduce((a, b) => a + b, 0) / completionTimes.length)
    : 0

  const medianCompletionTime = completionTimes.length > 0
    ? Math.round(completionTimes.sort((a, b) => a - b)[Math.floor(completionTimes.length / 2)])
    : 0

  // Group by user
  const userStats: Record<string, { completed: number; failed: number; total: number }> = {}
  for (const job of jobs) {
    const user = job.user_name ?? 'Unknown'
    if (!userStats[user]) {
      userStats[user] = { completed: 0, failed: 0, total: 0 }
    }
    userStats[user].total++
    if (job.job_state === 'COMPLETED') userStats[user].completed++
    if (job.job_state === 'FAILED') userStats[user].failed++
  }

  const topUsers = Object.entries(userStats)
    .map(([user, stats]) => ({
      user,
      ...stats,
      successRate: stats.total > 0 ? Math.round((stats.completed / stats.total) * 100) : 0
    }))
    .sort((a, b) => b.total - a.total)
    .slice(0, 10)

  // Recent jobs for table (last 20)
  const recentJobs = jobs
    .filter(j => j.end_time)
    .sort((a, b) => {
      const aTime = a.end_time ? new Date(a.end_time).getTime() : 0
      const bTime = b.end_time ? new Date(b.end_time).getTime() : 0
      return bTime - aTime
    })
    .slice(0, 20)

  const formatDuration = (minutes: number) => {
    if (minutes < 60) return `${Math.round(minutes)}m`
    const hours = Math.floor(minutes / 60)
    const mins = Math.round(minutes % 60)
    return `${hours}h ${mins}m`
  }

  return (
    <VStack w="100%" align="start" gap={4}>
      <Text fontSize="lg" fontWeight="semibold">Job Analytics (Last 24 Hours)</Text>

      {/* Summary Cards */}
      <SimpleGrid columns={{ base: 2, md: 4, lg: 6 }} gap={3} w="100%">
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Total Jobs</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold">
              {jobs.length}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Completed</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold" color="green.600">
              {completedJobs.length}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Failed</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold" color="red.600">
              {failedJobs.length}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Success Rate</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold" color={successRate >= 80 ? 'green.600' : 'orange.600'}>
              {successRate}%
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Avg Duration</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold">
              {avgCompletionTime > 0 ? formatDuration(avgCompletionTime) : 'N/A'}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Median Duration</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold">
              {medianCompletionTime > 0 ? formatDuration(medianCompletionTime) : 'N/A'}
            </Stat.ValueText>
          </Stat.Root>
        </Box>
      </SimpleGrid>

      <SimpleGrid columns={{ base: 1, lg: 2 }} gap={4} w="100%">
        {/* Recent Jobs Table */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Recent Completed Jobs</Text>
            {jobsQ.isLoading ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg" />
              </Box>
            ) : (
              <Box w="100%" maxH="300px" overflowY="auto">
                <Table.Root size="sm" variant="outline">
                  <Table.Header>
                    <Table.Row bg="gray.50">
                      <Table.ColumnHeader fontSize="xs">Job ID</Table.ColumnHeader>
                      <Table.ColumnHeader fontSize="xs">User</Table.ColumnHeader>
                      <Table.ColumnHeader fontSize="xs">State</Table.ColumnHeader>
                      <Table.ColumnHeader fontSize="xs">Duration</Table.ColumnHeader>
                    </Table.Row>
                  </Table.Header>
                  <Table.Body>
                    {recentJobs.length === 0 ? (
                      <Table.Row>
                        <Table.Cell colSpan={4}>
                          <Text fontSize="xs" color="gray.500" textAlign="center">No completed jobs in last 24h</Text>
                        </Table.Cell>
                      </Table.Row>
                    ) : (
                      recentJobs.map((job) => {
                        const duration = job.start_time && job.end_time
                          ? (new Date(job.end_time).getTime() - new Date(job.start_time).getTime()) / (1000 * 60)
                          : 0

                        return (
                          <Table.Row key={`${job.job_id}-${job.job_step}`}>
                            <Table.Cell fontSize="xs" fontWeight="medium">{job.job_id}</Table.Cell>
                            <Table.Cell fontSize="xs" truncate>{job.user_name}</Table.Cell>
                            <Table.Cell fontSize="xs">
                              <Tag.Root size="sm" colorPalette={getJobStateColor(job.job_state ?? '')}>
                                <Tag.Label>{job.job_state}</Tag.Label>
                              </Tag.Root>
                            </Table.Cell>
                            <Table.Cell fontSize="xs">{duration > 0 ? formatDuration(duration) : 'N/A'}</Table.Cell>
                          </Table.Row>
                        )
                      })
                    )}
                  </Table.Body>
                </Table.Root>
              </Box>
            )}
          </VStack>
        </Box>

        {/* User Statistics */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Top Users by Job Count</Text>
            <Box w="100%" maxH="300px" overflowY="auto">
              <Table.Root size="sm" variant="outline">
                <Table.Header>
                  <Table.Row bg="gray.50">
                    <Table.ColumnHeader fontSize="xs">User</Table.ColumnHeader>
                    <Table.ColumnHeader fontSize="xs" textAlign="center">Total</Table.ColumnHeader>
                    <Table.ColumnHeader fontSize="xs" textAlign="center">Success %</Table.ColumnHeader>
                  </Table.Row>
                </Table.Header>
                <Table.Body>
                  {topUsers.length === 0 ? (
                    <Table.Row>
                      <Table.Cell colSpan={3}>
                        <Text fontSize="xs" color="gray.500" textAlign="center">No user data</Text>
                      </Table.Cell>
                    </Table.Row>
                  ) : (
                    topUsers.map((u) => (
                      <Table.Row key={u.user}>
                        <Table.Cell fontSize="xs" fontWeight="medium" truncate>{u.user}</Table.Cell>
                        <Table.Cell fontSize="xs" textAlign="center">{u.total}</Table.Cell>
                        <Table.Cell fontSize="xs" textAlign="center">
                          <Tag.Root size="sm" colorPalette={u.successRate >= 80 ? 'green' : 'orange'}>
                            <Tag.Label>{u.successRate}%</Tag.Label>
                          </Tag.Root>
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
