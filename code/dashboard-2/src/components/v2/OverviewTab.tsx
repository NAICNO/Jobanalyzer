import { memo } from 'react'
import { VStack, SimpleGrid, Card, Text, DataList, Separator } from '@chakra-ui/react'
import type { JobResponse } from '../../client/types.gen'
import { formatDuration, formatMemory } from '../../util/formatters'

type OverviewTabProps = {
  job: JobResponse
  elapsed: number
  gpuInfo: {
    requested: number
    allocated: number
    uuids: string[]
    gresDetail: string[] | null | undefined
  }
}

export const OverviewTab = memo(({ job, elapsed, gpuInfo }: OverviewTabProps) => {
  return (
    <VStack align="start" gap={6} w="100%">
      {/* Job Details Grid */}
      <SimpleGrid columns={{ base: 1, lg: 3 }} gap={6} w="100%">
        <Card.Root>
          <Card.Body gap={4}>
            <Text fontSize="xl" fontWeight="semibold">Job Information</Text>
            <DataList.Root orientation="horizontal" size="md">
              <DataList.Item>
                <DataList.ItemLabel>Job Name</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">{job.job_name || 'N/A'}</DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Job Step</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">{job.job_step || '(main)'}</DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>User</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">{job.user_name || 'N/A'}</DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Account</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">{job.account || 'N/A'}</DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Partition</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">{job.partition || 'N/A'}</DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Reservation</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">{job.reservation || 'None'}</DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Priority</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">{job.priority ?? 'N/A'}</DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Exit Code</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">{job.exit_code ?? 'N/A'}</DataList.ItemValue>
              </DataList.Item>
            </DataList.Root>
          </Card.Body>
        </Card.Root>

        <Card.Root>
          <Card.Body gap={4}>
            <Text fontSize="xl" fontWeight="semibold">Time Information</Text>
            <DataList.Root orientation="horizontal" size="md">
              <DataList.Item>
                <DataList.ItemLabel>Timestamp</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">
                  {job.time ? new Date(job.time).toLocaleString() : 'N/A'}
                </DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Submit Time</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">
                  {job.submit_time ? new Date(job.submit_time).toLocaleString() : 'N/A'}
                </DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Start Time</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">
                  {job.start_time ? new Date(job.start_time).toLocaleString() : 'N/A'}
                </DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>End Time</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">
                  {job.end_time ? new Date(job.end_time).toLocaleString() : 'Running'}
                </DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Elapsed Time</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">
                  {formatDuration(elapsed)}
                </DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Suspend Time</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">
                  {formatDuration(job.suspend_time)}
                </DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Time Limit</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">
                  {formatDuration(job.time_limit)}
                </DataList.ItemValue>
              </DataList.Item>
            </DataList.Root>
          </Card.Body>
        </Card.Root>

        <Card.Root>
          <Card.Body gap={4}>
            <Text fontSize="xl" fontWeight="semibold">Resource Information</Text>
            <DataList.Root orientation="horizontal" size="md">
              <DataList.Item>
                <DataList.ItemLabel>Nodes</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">{job.requested_node_count || 'N/A'}</DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>CPUs</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">{job.requested_cpus || 'N/A'}</DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Memory Per Node</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">
                  {formatMemory(job.requested_memory_per_node / 1024)}
                </DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Distribution</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">
                  {job.distribution || 'N/A'}
                </DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Requested Resources</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium" wordBreak="break-word">
                  {job.requested_resources || 'N/A'}
                </DataList.ItemValue>
              </DataList.Item>
              <DataList.Item>
                <DataList.ItemLabel>Allocated Resources</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium" wordBreak="break-word">
                  {job.allocated_resources || 'N/A'}
                </DataList.ItemValue>
              </DataList.Item>
            </DataList.Root>
          </Card.Body>
        </Card.Root>

        <Card.Root>
          <Card.Body gap={4}>
            <Text fontSize="xl" fontWeight="semibold">Node Information</Text>
            <DataList.Root orientation="horizontal" size="md">
              <DataList.Item>
                <DataList.ItemLabel>Nodes</DataList.ItemLabel>
                <DataList.ItemValue fontWeight="medium">
                  {Array.isArray(job.nodes) ? job.nodes.join(', ') : job.nodes || 'N/A'}
                </DataList.ItemValue>
              </DataList.Item>
            </DataList.Root>
          </Card.Body>
        </Card.Root>
      </SimpleGrid>

      {/* SLURM Accounting Data */}
      {job.sacct && (
        <>
          <Separator />
          <Card.Root>
            <Card.Body gap={4}>
              <Text fontSize="xl" fontWeight="semibold">SLURM Accounting Data</Text>
              <DataList.Root orientation="horizontal" size="md">
                <DataList.Item>
                  <DataList.ItemLabel>Allocated TRES</DataList.ItemLabel>
                  <DataList.ItemValue fontWeight="medium" wordBreak="break-word">
                    {job.sacct.AllocTRES}
                  </DataList.ItemValue>
                </DataList.Item>
                <DataList.Item>
                  <DataList.ItemLabel>System CPU Time</DataList.ItemLabel>
                  <DataList.ItemValue fontWeight="medium">
                    {formatDuration(job.sacct.SystemCPU)}
                  </DataList.ItemValue>
                </DataList.Item>
                <DataList.Item>
                  <DataList.ItemLabel>User CPU Time</DataList.ItemLabel>
                  <DataList.ItemValue fontWeight="medium">
                    {formatDuration(job.sacct.UserCPU)}
                  </DataList.ItemValue>
                </DataList.Item>
                <DataList.Item>
                  <DataList.ItemLabel>Average CPU</DataList.ItemLabel>
                  <DataList.ItemValue fontWeight="medium">
                    {formatDuration(job.sacct.AveCPU)}
                  </DataList.ItemValue>
                </DataList.Item>
                <DataList.Item>
                  <DataList.ItemLabel>Min CPU</DataList.ItemLabel>
                  <DataList.ItemValue fontWeight="medium">
                    {formatDuration(job.sacct.MinCPU)}
                  </DataList.ItemValue>
                </DataList.Item>
                <DataList.Item>
                  <DataList.ItemLabel>Max RSS</DataList.ItemLabel>
                  <DataList.ItemValue fontWeight="medium">
                    {formatMemory(job.sacct.MaxRSS / 1024)}
                  </DataList.ItemValue>
                </DataList.Item>
                <DataList.Item>
                  <DataList.ItemLabel>Average RSS</DataList.ItemLabel>
                  <DataList.ItemValue fontWeight="medium">
                    {formatMemory(job.sacct.AveRSS / 1024)}
                  </DataList.ItemValue>
                </DataList.Item>
                <DataList.Item>
                  <DataList.ItemLabel>Max VM Size</DataList.ItemLabel>
                  <DataList.ItemValue fontWeight="medium">
                    {formatMemory(job.sacct.MaxVMSize / 1024)}
                  </DataList.ItemValue>
                </DataList.Item>
                <DataList.Item>
                  <DataList.ItemLabel>Average VM Size</DataList.ItemLabel>
                  <DataList.ItemValue fontWeight="medium">
                    {formatMemory(job.sacct.AveVMSize / 1024)}
                  </DataList.ItemValue>
                </DataList.Item>
                <DataList.Item>
                  <DataList.ItemLabel>Average Disk Read</DataList.ItemLabel>
                  <DataList.ItemValue fontWeight="medium">
                    {formatMemory(job.sacct.AveDiskRead / 1024)}
                  </DataList.ItemValue>
                </DataList.Item>
                <DataList.Item>
                  <DataList.ItemLabel>Average Disk Write</DataList.ItemLabel>
                  <DataList.ItemValue fontWeight="medium">
                    {formatMemory(job.sacct.AveDiskWrite / 1024)}
                  </DataList.ItemValue>
                </DataList.Item>
              </DataList.Root>
            </Card.Body>
          </Card.Root>
        </>
      )}

      {/* GPU Information */}
      {(gpuInfo.requested > 0 || gpuInfo.allocated > 0 || gpuInfo.uuids.length > 0) && (
        <>
          <Separator />
          <Card.Root>
            <Card.Body gap={4}>
              <Text fontSize="xl" fontWeight="semibold">GPU Information</Text>
              <DataList.Root orientation="horizontal" size="md">
                {gpuInfo.requested > 0 && (
                  <DataList.Item>
                    <DataList.ItemLabel>Requested GPUs</DataList.ItemLabel>
                    <DataList.ItemValue fontSize="lg" fontWeight="medium">{gpuInfo.requested}</DataList.ItemValue>
                  </DataList.Item>
                )}
                {gpuInfo.allocated > 0 && (
                  <DataList.Item>
                    <DataList.ItemLabel>Allocated GPUs</DataList.ItemLabel>
                    <DataList.ItemValue fontSize="lg" fontWeight="medium">{gpuInfo.allocated}</DataList.ItemValue>
                  </DataList.Item>
                )}
                {gpuInfo.uuids.length > 0 && (
                  <DataList.Item>
                    <DataList.ItemLabel>GPU Count</DataList.ItemLabel>
                    <DataList.ItemValue fontSize="lg" fontWeight="medium">{gpuInfo.uuids.length}</DataList.ItemValue>
                  </DataList.Item>
                )}
                {gpuInfo.gresDetail && (
                  <DataList.Item>
                    <DataList.ItemLabel>GRES Details</DataList.ItemLabel>
                    <DataList.ItemValue fontWeight="medium" wordBreak="break-word">
                      {gpuInfo.gresDetail}
                    </DataList.ItemValue>
                  </DataList.Item>
                )}
                {gpuInfo.uuids.length > 0 && (
                  <DataList.Item>
                    <DataList.ItemLabel>GPU UUIDs</DataList.ItemLabel>
                    <DataList.ItemValue>
                      <VStack align="start" gap={1}>
                        {gpuInfo.uuids.map((uuid, idx) => (
                          <Text key={idx} fontSize="sm" fontFamily="mono">{uuid}</Text>
                        ))}
                      </VStack>
                    </DataList.ItemValue>
                  </DataList.Item>
                )}
              </DataList.Root>
            </Card.Body>
          </Card.Root>
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
})

OverviewTab.displayName = 'OverviewTab'
