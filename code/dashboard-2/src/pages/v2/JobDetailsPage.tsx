import { useMemo } from 'react'
import { Box, Text, VStack, HStack, Spinner, Alert, Badge, SimpleGrid, Card, Separator } from '@chakra-ui/react'
import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'

import { getClusterByClusterJobsByJobIdOptions } from '../../client/@tanstack/react-query.gen'
import { useClusterClient } from '../../hooks/useClusterClient'
import { formatDuration, formatMemory, getJobStateColor } from '../../util/formatters'

export const JobDetailsPage = () => {
  const { clusterName, jobId } = useParams<{ clusterName: string; jobId: string }>()
  const client = useClusterClient(clusterName)

  const jobIdNum = useMemo(() => {
    const parsed = parseInt(jobId ?? '', 10)
    return isNaN(parsed) ? undefined : parsed
  }, [jobId])

  const { data: job, isLoading, isError, error } = useQuery({
    ...getClusterByClusterJobsByJobIdOptions({
      path: { 
        cluster: clusterName ?? '', 
        job_id: jobIdNum ?? 0 
      },
      client: client ?? undefined,
    }),
    enabled: !!clusterName && !!jobIdNum && !!client,
  })

  const elapsed = useMemo(() => {
    if (!job?.start_time) return 0
    const start = new Date(job.start_time).getTime()
    const end = job.end_time ? new Date(job.end_time).getTime() : Date.now()
    return Math.floor((end - start) / 1000)
  }, [job?.start_time, job?.end_time])

  // Extract GPU count from resource strings
  const gpuInfo = useMemo(() => {
    const extractGpuCount = (resources: string | null | undefined): number => {
      if (!resources) return 0
      const match = resources.match(/gres\/gpu[=:](\d+)/)
      return match ? parseInt(match[1], 10) : 0
    }

    return {
      requested: extractGpuCount(job?.requested_resources),
      allocated: extractGpuCount(job?.allocated_resources),
      uuids: job?.used_gpu_uuids || [],
      gresDetail: job?.gres_detail
    }
  }, [job?.requested_resources, job?.allocated_resources, job?.used_gpu_uuids, job?.gres_detail])

  // Render loading/error states after all hooks
  if (!client) {
    return (
      <VStack p={4} align="start">
        <Spinner />
      </VStack>
    )
  }

  if (!clusterName || !jobIdNum) {
    return (
      <VStack p={4} align="start">
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>Invalid job ID or cluster name</Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  if (isError) {
    return (
      <VStack p={4} align="start" gap={4}>
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>
            Failed to load job details: {error?.message || 'Unknown error'}
          </Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  if (isLoading) {
    return (
      <Box w="100%" h="400px" display="flex" alignItems="center" justifyContent="center">
        <Spinner size="xl" />
      </Box>
    )
  }

  if (!job) {
    return (
      <VStack p={4} align="start">
        <Alert.Root status="warning">
          <Alert.Indicator />
          <Alert.Description>Job not found</Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  return (
    <VStack w="100%" align="start" gap={6} p={6}>
      {/* Header */}
      <VStack align="start" gap={2} w="100%">
        <HStack gap={3}>
          <Text fontSize="3xl" fontWeight="bold">
            Job {job.job_id}
          </Text>
          <Badge colorPalette={getJobStateColor(job.job_state)} size="lg">
            {job.job_state}
          </Badge>
        </HStack>
        <Text fontSize="md" color="fg.muted">
          Cluster: {clusterName}
        </Text>
      </VStack>

      {/* Job Information */}
      <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} gap={4} w="100%">
        <Card.Root>
          <Card.Body gap={2}>
            <Text fontSize="sm" color="fg.muted">Job Name</Text>
            <Text fontSize="lg" fontWeight="medium">{job.job_name || 'N/A'}</Text>
          </Card.Body>
        </Card.Root>

        <Card.Root>
          <Card.Body gap={2}>
            <Text fontSize="sm" color="fg.muted">Job Step</Text>
            <Text fontSize="lg" fontWeight="medium">{job.job_step || '(main)'}</Text>
          </Card.Body>
        </Card.Root>

        <Card.Root>
          <Card.Body gap={2}>
            <Text fontSize="sm" color="fg.muted">User</Text>
            <Text fontSize="lg" fontWeight="medium">{job.user_name || 'N/A'}</Text>
          </Card.Body>
        </Card.Root>

        <Card.Root>
          <Card.Body gap={2}>
            <Text fontSize="sm" color="fg.muted">Account</Text>
            <Text fontSize="lg" fontWeight="medium">{job.account || 'N/A'}</Text>
          </Card.Body>
        </Card.Root>

        <Card.Root>
          <Card.Body gap={2}>
            <Text fontSize="sm" color="fg.muted">Partition</Text>
            <Text fontSize="lg" fontWeight="medium">{job.partition || 'N/A'}</Text>
          </Card.Body>
        </Card.Root>

        <Card.Root>
          <Card.Body gap={2}>
            <Text fontSize="sm" color="fg.muted">Reservation</Text>
            <Text fontSize="lg" fontWeight="medium">{job.reservation || 'None'}</Text>
          </Card.Body>
        </Card.Root>

        <Card.Root>
          <Card.Body gap={2}>
            <Text fontSize="sm" color="fg.muted">Priority</Text>
            <Text fontSize="lg" fontWeight="medium">{job.priority ?? 'N/A'}</Text>
          </Card.Body>
        </Card.Root>

        <Card.Root>
          <Card.Body gap={2}>
            <Text fontSize="sm" color="fg.muted">Exit Code</Text>
            <Text fontSize="lg" fontWeight="medium">{job.exit_code ?? 'N/A'}</Text>
          </Card.Body>
        </Card.Root>
      </SimpleGrid>

      <Separator />

      {/* Time Information */}
      <VStack align="start" gap={4} w="100%">
        <Text fontSize="xl" fontWeight="semibold">Time Information</Text>
        <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} gap={4} w="100%">
          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">Timestamp</Text>
              <Text fontSize="md" fontWeight="medium">
                {job.time ? new Date(job.time).toLocaleString() : 'N/A'}
              </Text>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">Submit Time</Text>
              <Text fontSize="md" fontWeight="medium">
                {job.submit_time ? new Date(job.submit_time).toLocaleString() : 'N/A'}
              </Text>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">Start Time</Text>
              <Text fontSize="md" fontWeight="medium">
                {job.start_time ? new Date(job.start_time).toLocaleString() : 'N/A'}
              </Text>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">End Time</Text>
              <Text fontSize="md" fontWeight="medium">
                {job.end_time ? new Date(job.end_time).toLocaleString() : 'Running'}
              </Text>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">Elapsed Time</Text>
              <Text fontSize="md" fontWeight="medium">
                {formatDuration(elapsed)}
              </Text>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">Suspend Time</Text>
              <Text fontSize="md" fontWeight="medium">
                {formatDuration(job.suspend_time)}
              </Text>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">Time Limit</Text>
              <Text fontSize="md" fontWeight="medium">
                {formatDuration(job.time_limit)}
              </Text>
            </Card.Body>
          </Card.Root>
        </SimpleGrid>
      </VStack>

      <Separator />

      {/* Resource Information */}
      <VStack align="start" gap={4} w="100%">
        <Text fontSize="xl" fontWeight="semibold">Resource Information</Text>
        <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} gap={4} w="100%">
          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">Nodes</Text>
              <Text fontSize="md" fontWeight="medium">{job.requested_node_count || 'N/A'}</Text>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">CPUs</Text>
              <Text fontSize="md" fontWeight="medium">{job.requested_cpus || 'N/A'}</Text>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">Memory Per Node</Text>
              <Text fontSize="md" fontWeight="medium">
                {formatMemory(job.requested_memory_per_node)}
              </Text>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">Distribution</Text>
              <Text fontSize="md" fontWeight="medium">
                {job.distribution || 'N/A'}
              </Text>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">Requested Resources</Text>
              <Text fontSize="md" fontWeight="medium" wordBreak="break-word">
                {job.requested_resources || 'N/A'}
              </Text>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Body gap={2}>
              <Text fontSize="sm" color="fg.muted">Allocated Resources</Text>
              <Text fontSize="md" fontWeight="medium" wordBreak="break-word">
                {job.allocated_resources || 'N/A'}
              </Text>
            </Card.Body>
          </Card.Root>
        </SimpleGrid>
      </VStack>

      <Separator />

      {/* Node Information */}
      <VStack align="start" gap={4} w="100%">
        <Text fontSize="xl" fontWeight="semibold">Node Information</Text>
        <Card.Root w="100%">
          <Card.Body gap={2}>
            <Text fontSize="sm" color="fg.muted">Nodes</Text>
            <Text fontSize="md" fontWeight="medium">
              {Array.isArray(job.nodes) ? job.nodes.join(', ') : job.nodes || 'N/A'}
            </Text>
          </Card.Body>
        </Card.Root>
      </VStack>

      {/* GPU Information */}
      {(gpuInfo.requested > 0 || gpuInfo.allocated > 0 || gpuInfo.uuids.length > 0) && (
        <>
          <Separator />
          <VStack align="start" gap={4} w="100%">
            <Text fontSize="xl" fontWeight="semibold">GPU Information</Text>
            <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} gap={4} w="100%">
              {gpuInfo.requested > 0 && (
                <Card.Root>
                  <Card.Body gap={2}>
                    <Text fontSize="sm" color="fg.muted">Requested GPUs</Text>
                    <Text fontSize="lg" fontWeight="medium">{gpuInfo.requested}</Text>
                  </Card.Body>
                </Card.Root>
              )}

              {gpuInfo.allocated > 0 && (
                <Card.Root>
                  <Card.Body gap={2}>
                    <Text fontSize="sm" color="fg.muted">Allocated GPUs</Text>
                    <Text fontSize="lg" fontWeight="medium">{gpuInfo.allocated}</Text>
                  </Card.Body>
                </Card.Root>
              )}

              {gpuInfo.uuids.length > 0 && (
                <Card.Root>
                  <Card.Body gap={2}>
                    <Text fontSize="sm" color="fg.muted">GPU Count</Text>
                    <Text fontSize="lg" fontWeight="medium">{gpuInfo.uuids.length}</Text>
                  </Card.Body>
                </Card.Root>
              )}
            </SimpleGrid>

            {gpuInfo.gresDetail && (
              <Card.Root w="100%">
                <Card.Body gap={2}>
                  <Text fontSize="sm" color="fg.muted">GRES Details</Text>
                  <Text fontSize="md" fontWeight="medium" wordBreak="break-word">
                    {gpuInfo.gresDetail}
                  </Text>
                </Card.Body>
              </Card.Root>
            )}

            {gpuInfo.uuids.length > 0 && (
              <Card.Root w="100%">
                <Card.Body gap={2}>
                  <Text fontSize="sm" color="fg.muted">GPU UUIDs</Text>
                  <VStack align="start" gap={1}>
                    {gpuInfo.uuids.map((uuid, idx) => (
                      <Text key={idx} fontSize="sm" fontFamily="mono">{uuid}</Text>
                    ))}
                  </VStack>
                </Card.Body>
              </Card.Root>
            )}
          </VStack>
        </>
      )}

      {/* Heterogeneous Job Information */}
      {job.het_job_id > 0 && (
        <>
          <Separator />
          <VStack align="start" gap={4} w="100%">
            <Text fontSize="xl" fontWeight="semibold">Heterogeneous Job Information</Text>
            <SimpleGrid columns={{ base: 1, md: 2 }} gap={4} w="100%">
              <Card.Root>
                <Card.Body gap={2}>
                  <Text fontSize="sm" color="fg.muted">Het Job ID</Text>
                  <Text fontSize="md" fontWeight="medium">{job.het_job_id}</Text>
                </Card.Body>
              </Card.Root>

              <Card.Root>
                <Card.Body gap={2}>
                  <Text fontSize="sm" color="fg.muted">Het Job Offset</Text>
                  <Text fontSize="md" fontWeight="medium">{job.het_job_offset}</Text>
                </Card.Body>
              </Card.Root>
            </SimpleGrid>
          </VStack>
        </>
      )}

      {/* Array Job Information */}
      {job.array_job_id != null && job.array_job_id > 0 && (
        <>
          <Separator />
          <VStack align="start" gap={4} w="100%">
            <Text fontSize="xl" fontWeight="semibold">Array Job Information</Text>
            <SimpleGrid columns={{ base: 1, md: 2 }} gap={4} w="100%">
              <Card.Root>
                <Card.Body gap={2}>
                  <Text fontSize="sm" color="fg.muted">Array Job ID</Text>
                  <Text fontSize="md" fontWeight="medium">{job.array_job_id}</Text>
                </Card.Body>
              </Card.Root>

              {job.array_task_id !== null && job.array_task_id !== undefined && (
                <Card.Root>
                  <Card.Body gap={2}>
                    <Text fontSize="sm" color="fg.muted">Array Task ID</Text>
                    <Text fontSize="md" fontWeight="medium">{job.array_task_id}</Text>
                  </Card.Body>
                </Card.Root>
              )}
            </SimpleGrid>
          </VStack>
        </>
      )}
    </VStack>
  )
}
