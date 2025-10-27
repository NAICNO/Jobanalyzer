import { Accordion, VStack, Text, Box, HStack, Tag, Table } from '@chakra-ui/react'

import type { PartitionResponseOutput, JobResponse } from '../../client'

interface Props {
  partition: PartitionResponseOutput
}

export const PartitionQueueOverview = ({ partition }: Props) => {
  const runningJobs = partition.jobs_running ?? []
  const pendingJobs = partition.jobs_pending ?? []
  
  // Calculate metrics
  const oldestPendingMs = partition.pending_max_submit_time 
    ? (() => {
      const submitTime = new Date(partition.pending_max_submit_time).getTime()
      // Ignore epoch time (1970-01-01)
      if (submitTime < 365 * 24 * 60 * 60 * 1000) return null
      return Date.now() - submitTime
    })()
    : null
  const oldestPendingHrs = oldestPendingMs ? Math.round(oldestPendingMs / (1000 * 60 * 60)) : null
  
  const latestWaitMin = partition.running_latest_wait_time 
    ? Math.round(partition.running_latest_wait_time / 60)
    : null

  const formatDuration = (ms: number) => {
    const hrs = Math.floor(ms / (1000 * 60 * 60))
    const mins = Math.floor((ms % (1000 * 60 * 60)) / (1000 * 60))
    if (hrs > 0) return `${hrs}h ${mins}m`
    return `${mins}m`
  }

  const formatWaitTime = (submitTime: string | Date | null | undefined) => {
    if (!submitTime) return 'N/A'
    const wait = Date.now() - new Date(submitTime).getTime()
    return formatDuration(wait)
  }

  const formatJobDuration = (startTime: string | Date | null | undefined) => {
    if (!startTime) return 'N/A'
    const duration = Date.now() - new Date(startTime).getTime()
    return formatDuration(duration)
  }

  return (
    <Accordion.Root variant="outline" multiple>
      <Accordion.Item value="queue">
        <Accordion.ItemTrigger px={3} py={2} _hover={{ bg: 'gray.50' }}>
          <Text fontWeight="semibold">Queue Overview</Text>
          <Accordion.ItemIndicator />
        </Accordion.ItemTrigger>
        <Accordion.ItemContent>
          <Accordion.ItemBody>
            <VStack align="start" gap={4} w="100%">
              {/* Metrics Chips */}
              <HStack gap={3} flexWrap="wrap">
                {oldestPendingHrs !== null && (
                  <Tag.Root colorPalette={oldestPendingHrs > 24 ? 'red' : oldestPendingHrs > 4 ? 'yellow' : 'gray'}>
                    <Tag.Label>Oldest Pending: {oldestPendingHrs}h</Tag.Label>
                  </Tag.Root>
                )}
                {latestWaitMin !== null && (
                  <Tag.Root colorPalette={latestWaitMin > 60 ? 'yellow' : 'gray'}>
                    <Tag.Label>Latest Wait: {latestWaitMin}m</Tag.Label>
                  </Tag.Root>
                )}
                <Tag.Root colorPalette="blue">
                  <Tag.Label>Running: {runningJobs.length}</Tag.Label>
                </Tag.Root>
                <Tag.Root colorPalette={pendingJobs.length > 10 ? 'orange' : 'gray'}>
                  <Tag.Label>Pending: {pendingJobs.length}</Tag.Label>
                </Tag.Root>
              </HStack>

              {/* Running Jobs Table */}
              {runningJobs.length > 0 && (
                <Box w="100%">
                  <Text fontWeight="semibold" mb={2}>Running Jobs (Top 5)</Text>
                  <Table.Root size="sm" variant="outline">
                    <Table.Header>
                      <Table.Row>
                        <Table.ColumnHeader>Job ID</Table.ColumnHeader>
                        <Table.ColumnHeader>User</Table.ColumnHeader>
                        <Table.ColumnHeader>Duration</Table.ColumnHeader>
                        <Table.ColumnHeader>CPUs</Table.ColumnHeader>
                      </Table.Row>
                    </Table.Header>
                    <Table.Body>
                      {runningJobs.slice(0, 5).map((job: JobResponse) => (
                        <Table.Row key={job.job_id ?? Math.random()}>
                          <Table.Cell fontFamily="mono" fontSize="sm">{job.job_id}</Table.Cell>
                          <Table.Cell>{job.user_name}</Table.Cell>
                          <Table.Cell>{formatJobDuration(job.start_time)}</Table.Cell>
                          <Table.Cell>{job.requested_cpus ?? 'N/A'}</Table.Cell>
                        </Table.Row>
                      ))}
                    </Table.Body>
                  </Table.Root>
                  {runningJobs.length > 5 && (
                    <Text fontSize="xs" color="gray.500" mt={1}>
                      … and {runningJobs.length - 5} more
                    </Text>
                  )}
                </Box>
              )}

              {/* Pending Jobs Table */}
              {pendingJobs.length > 0 && (
                <Box w="100%">
                  <Text fontWeight="semibold" mb={2}>Pending Jobs (Top 5 by Wait Time)</Text>
                  <Table.Root size="sm" variant="outline">
                    <Table.Header>
                      <Table.Row>
                        <Table.ColumnHeader>Job ID</Table.ColumnHeader>
                        <Table.ColumnHeader>User</Table.ColumnHeader>
                        <Table.ColumnHeader>Waiting Since</Table.ColumnHeader>
                        <Table.ColumnHeader>CPUs</Table.ColumnHeader>
                      </Table.Row>
                    </Table.Header>
                    <Table.Body>
                      {pendingJobs
                        .sort((a, b) => {
                          const aTime = a.submit_time ? new Date(a.submit_time).getTime() : 0
                          const bTime = b.submit_time ? new Date(b.submit_time).getTime() : 0
                          return aTime - bTime
                        })
                        .slice(0, 5)
                        .map((job: JobResponse) => (
                          <Table.Row key={job.job_id ?? Math.random()}>
                            <Table.Cell fontFamily="mono" fontSize="sm">{job.job_id}</Table.Cell>
                            <Table.Cell>{job.user_name}</Table.Cell>
                            <Table.Cell>{formatWaitTime(job.submit_time)}</Table.Cell>
                            <Table.Cell>{job.requested_cpus ?? 'N/A'}</Table.Cell>
                          </Table.Row>
                        ))}
                    </Table.Body>
                  </Table.Root>
                  {pendingJobs.length > 5 && (
                    <Text fontSize="xs" color="gray.500" mt={1}>
                      … and {pendingJobs.length - 5} more
                    </Text>
                  )}
                </Box>
              )}

              {runningJobs.length === 0 && pendingJobs.length === 0 && (
                <Text color="gray.500">No jobs in queue</Text>
              )}
            </VStack>
          </Accordion.ItemBody>
        </Accordion.ItemContent>
      </Accordion.Item>
    </Accordion.Root>
  )
}
