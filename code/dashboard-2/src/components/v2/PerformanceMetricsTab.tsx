import { useMemo, memo } from 'react'
import { Box, Text, VStack, HStack, SimpleGrid, Card, Alert, Progress, Separator, Tooltip, Icon } from '@chakra-ui/react'
import { HiInformationCircle } from 'react-icons/hi2'
import type { JobReport, JobResponse } from '../../client/types.gen'
import { formatMemory, formatEfficiency, formatIORate } from '../../util/formatters'
import {
  calculateEfficiencyMetrics,
  getEfficiencyColor,
  getEfficiencyLabel,
  calculateWastedCpuHours,
  calculateWastedMemory,
} from '../../util/efficiency'
import { JobStatCard } from './JobStatCard'

type PerformanceMetricsTabProps = {
  job: JobResponse
  report?: JobReport
  elapsed: number
}

export const PerformanceMetricsTab = memo(({ job, report, elapsed }: PerformanceMetricsTabProps) => {
  const efficiency = useMemo(
    () => calculateEfficiencyMetrics(job, report, elapsed),
    [job, report, elapsed]
  )

  const wastedResources = useMemo(() => {
    return {
      cpuHours: calculateWastedCpuHours(job, elapsed),
      memory: calculateWastedMemory(job),
    }
  }, [job, elapsed])

  // Parse stats from report
  const parseStats = (stats: { [key: string]: number } | undefined) => {
    if (!stats) return { mean: 0, stddev: 0 }
    const mean = typeof stats.mean === 'number' ? stats.mean : 0
    const stddev = typeof stats.stddev === 'number' ? stats.stddev : 0
    return { mean, stddev }
  }

  const cpuUtil = parseStats(report?.cpu_util)
  const residentMemory = parseStats(report?.resident_memory)
  const numThreads = parseStats(report?.num_threads)
  const dataRead = parseStats(report?.data_read)
  const dataWritten = parseStats(report?.data_written)

  if (!report) {
    return (
      <VStack align="start" gap={4} w="100%">
        <Alert.Root status="info">
          <Alert.Indicator />
          <Alert.Description>
            Performance metrics are not available for this job. This may be because the job is still running or performance data has not been collected yet.
          </Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  return (
    <VStack align="start" gap={4} w="100%">
      {/* Performance Warnings - High Priority */}
      {report.warnings && report.warnings.length > 0 && (
        <Box w="100%">
          <Text fontSize="xl" fontWeight="semibold" mb={3}>
            Performance Warnings
          </Text>
          <VStack align="start" gap={2}>
            {report.warnings.map((warning, idx) => (
              <Alert.Root key={idx} status="warning">
                <Alert.Indicator />
                <Alert.Description>{warning}</Alert.Description>
              </Alert.Root>
            ))}
          </VStack>
        </Box>
      )}

      {report.warnings && report.warnings.length > 0 && <Separator />}

      {/* Summary Statistics */}
      <Box w="100%">
        <Text fontSize="xl" fontWeight="semibold" mb={3}>
          Performance Summary
        </Text>
        <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} gap={3}>
          <JobStatCard
            label="Average CPU Utilization"
            value={`${cpuUtil.mean.toFixed(1)}%`}
            helpText={`± ${cpuUtil.stddev.toFixed(1)}%`}
            tooltip="Average CPU usage across the job's lifetime. Shows mean ± standard deviation to indicate consistency of CPU utilization."
          />

          <JobStatCard
            label="Peak Resident Memory"
            value={formatMemory(residentMemory.mean / 1024)}
            helpText={`± ${formatMemory(residentMemory.stddev / 1024)}`}
            tooltip="Average peak resident set size (RSS) - physical memory actually used by the job processes, excluding swapped or virtual memory."
          />

          <JobStatCard
            label="Average Threads"
            value={numThreads.mean.toFixed(0)}
            helpText={`± ${numThreads.stddev.toFixed(0)}`}
            tooltip="Average number of threads/processes running during the job execution. Higher thread count may indicate parallel processing."
          />

          <JobStatCard
            label="I/O Bandwidth"
            value={formatIORate((dataRead.mean + dataWritten.mean) / elapsed)}
            helpText="Read + Write"
            tooltip="Combined read and write I/O throughput. Calculated as total data transferred divided by job elapsed time."
          />
        </SimpleGrid>
      </Box>

      <Separator />

      {/* Efficiency Indicators */}
      <Box w="100%">
        <Text fontSize="xl" fontWeight="semibold" mb={3}>
          Resource Efficiency
        </Text>
        <VStack align="start" gap={3} w="100%">
          {efficiency.cpu !== null && (
            <Box w="100%">
              <HStack justify="space-between" mb={1}>
                <HStack gap={1}>
                  <Text fontSize="md" fontWeight="medium">
                    CPU Efficiency
                  </Text>
                  <Tooltip.Root openDelay={300}>
                    <Tooltip.Trigger asChild>
                      <Icon size="md" color="gray.400" cursor="help">
                        <HiInformationCircle />
                      </Icon>
                    </Tooltip.Trigger>
                    <Tooltip.Positioner>
                      <Tooltip.Content maxW="300px">
                        <Text fontSize="xs">
                          Measures how much of the allocated CPU capacity was actually used. 
                          Calculated as: (Actual CPU Time Used) / (Requested CPUs × Elapsed Time) × 100%
                        </Text>
                      </Tooltip.Content>
                    </Tooltip.Positioner>
                  </Tooltip.Root>
                </HStack>
                <HStack gap={2}>
                  <Text fontSize="md" color="fg.muted">
                    {formatEfficiency(efficiency.cpu)}
                  </Text>
                  <Text fontSize="sm" color="fg.muted">
                    ({getEfficiencyLabel(efficiency.cpu)})
                  </Text>
                </HStack>
              </HStack>
              <Progress.Root
                value={efficiency.cpu}
                max={100}
                colorPalette={getEfficiencyColor(efficiency.cpu)}
              >
                <Progress.Track>
                  <Progress.Range />
                </Progress.Track>
              </Progress.Root>
              <Text fontSize="xs" color="fg.muted" mt={1}>
                Ratio of actual CPU time used to allocated CPU time
              </Text>
            </Box>
          )}

          {efficiency.memory !== null && (
            <Box w="100%">
              <HStack justify="space-between" mb={1}>
                <HStack gap={1}>
                  <Text fontSize="md" fontWeight="medium">
                    Memory Efficiency
                  </Text>
                  <Tooltip.Root openDelay={300}>
                    <Tooltip.Trigger asChild>
                      <Icon size="md" color="gray.400" cursor="help">
                        <HiInformationCircle />
                      </Icon>
                    </Tooltip.Trigger>
                    <Tooltip.Positioner>
                      <Tooltip.Content maxW="300px">
                        <Text fontSize="xs">
                          Shows how much of the requested memory was actually used at peak.
                          Calculated as: (Peak Memory Used) / (Requested Memory) × 100%
                        </Text>
                      </Tooltip.Content>
                    </Tooltip.Positioner>
                  </Tooltip.Root>
                </HStack>
                <HStack gap={2}>
                  <Text fontSize="md" color="fg.muted">
                    {formatEfficiency(efficiency.memory)}
                  </Text>
                  <Text fontSize="sm" color="fg.muted">
                    ({getEfficiencyLabel(efficiency.memory)})
                  </Text>
                </HStack>
              </HStack>
              <Progress.Root
                value={efficiency.memory}
                max={100}
                colorPalette={getEfficiencyColor(efficiency.memory)}
              >
                <Progress.Track>
                  <Progress.Range />
                </Progress.Track>
              </Progress.Root>
              <Text fontSize="xs" color="fg.muted" mt={1}>
                Ratio of peak memory used to allocated memory
              </Text>
            </Box>
          )}

          {efficiency.time !== null && (
            <Box w="100%">
              <HStack justify="space-between" mb={1}>
                <Text fontSize="md" fontWeight="medium">
                  Time Efficiency
                </Text>
                <HStack gap={2}>
                  <Text fontSize="md" color="fg.muted">
                    {formatEfficiency(efficiency.time)}
                  </Text>
                  <Text fontSize="sm" color="fg.muted">
                    ({getEfficiencyLabel(efficiency.time)})
                  </Text>
                </HStack>
              </HStack>
              <Progress.Root
                value={efficiency.time}
                max={100}
                colorPalette={getEfficiencyColor(efficiency.time)}
              >
                <Progress.Track>
                  <Progress.Range />
                </Progress.Track>
              </Progress.Root>
              <Text fontSize="xs" color="fg.muted" mt={1}>
                Ratio of elapsed time to time limit
              </Text>
            </Box>
          )}

          {efficiency.overall !== null && (
            <>
              <Separator />
              <Box w="100%">
                <HStack justify="space-between" mb={1}>
                  <Text fontSize="lg" fontWeight="semibold">
                    Overall Efficiency Score
                  </Text>
                  <HStack gap={2}>
                    <Text fontSize="lg" fontWeight="bold" color={`${getEfficiencyColor(efficiency.overall)}.600`}>
                      {formatEfficiency(efficiency.overall)}
                    </Text>
                    <Text fontSize="md" color="fg.muted">
                      ({getEfficiencyLabel(efficiency.overall)})
                    </Text>
                  </HStack>
                </HStack>
                <Progress.Root
                  value={efficiency.overall}
                  max={100}
                  colorPalette={getEfficiencyColor(efficiency.overall)}
                  size="lg"
                >
                  <Progress.Track>
                    <Progress.Range />
                  </Progress.Track>
                </Progress.Root>
                <Text fontSize="xs" color="fg.muted" mt={1}>
                  Weighted average of CPU (40%), Memory (30%), and Time (10%) efficiency
                </Text>
              </Box>
            </>
          )}
        </VStack>
      </Box>

      <Separator />

      {/* Resource Comparison */}
      <Box w="100%">
        <Text fontSize="xl" fontWeight="semibold" mb={3}>
          Requested vs. Used Resources
        </Text>
        <SimpleGrid columns={{ base: 1, md: 2 }} gap={3}>
          <Card.Root size={'sm'}>
            <Card.Body gap={2}>
              <Text fontSize="md" fontWeight="semibold">
                CPU Resources
              </Text>
              <HStack justify="space-between">
                <Text fontSize="sm" color="fg.muted">
                  Requested CPUs:
                </Text>
                <Text fontSize="sm" fontWeight="medium">
                  {job.requested_cpus ?? 'N/A'}
                </Text>
              </HStack>
              <HStack justify="space-between">
                <Text fontSize="sm" color="fg.muted">
                  Average Utilization:
                </Text>
                <Text fontSize="sm" fontWeight="medium">
                  {cpuUtil.mean.toFixed(1)}%
                </Text>
              </HStack>
              {wastedResources.cpuHours > 0 && (
                <HStack justify="space-between">
                  <Text fontSize="sm" color="orange.600">
                    Wasted CPU-hours:
                  </Text>
                  <Text fontSize="sm" fontWeight="medium" color="orange.600">
                    {wastedResources.cpuHours.toFixed(2)}
                  </Text>
                </HStack>
              )}
            </Card.Body>
          </Card.Root>

          <Card.Root size={'sm'}>
            <Card.Body gap={2}>
              <Text fontSize="md" fontWeight="semibold">
                Memory Resources
              </Text>
              <HStack justify="space-between">
                <Text fontSize="sm" color="fg.muted">
                  Requested Memory:
                </Text>
                <Text fontSize="sm" fontWeight="medium">
                  {formatMemory((job.requested_memory_per_node ?? 0) * (job.requested_node_count ?? 1) / 1024)}
                </Text>
              </HStack>
              <HStack justify="space-between">
                <Text fontSize="sm" color="fg.muted">
                  Peak Used Memory:
                </Text>
                <Text fontSize="sm" fontWeight="medium">
                  {formatMemory(residentMemory.mean / 1024)}
                </Text>
              </HStack>
              {wastedResources.memory > 0 && (
                <HStack justify="space-between">
                  <Text fontSize="sm" color="orange.600">
                    Wasted Memory:
                  </Text>
                  <Text fontSize="sm" fontWeight="medium" color="orange.600">
                    {formatMemory(wastedResources.memory / 1024)}
                  </Text>
                </HStack>
              )}
            </Card.Body>
          </Card.Root>
        </SimpleGrid>
      </Box>

      {/* Recommendations Section */}
      {(efficiency.cpu !== null && efficiency.cpu < 50) || (efficiency.memory !== null && efficiency.memory < 50) && (
        <>
          <Separator />
          <Box w="100%">
            <Text fontSize="xl" fontWeight="semibold" mb={3}>
              Recommendations
            </Text>
            <Card.Root>
              <Card.Body gap={2}>
                <Text fontSize="md" fontWeight="medium">
                  For future jobs like this:
                </Text>
                <VStack align="start" gap={2}>
                  {efficiency.cpu !== null && efficiency.cpu < 50 && (
                    <Text fontSize="sm">
                      • Consider requesting fewer CPUs. Your job used only {cpuUtil.mean.toFixed(0)}% of allocated CPU resources on average.
                    </Text>
                  )}
                  {efficiency.memory !== null && efficiency.memory < 50 && (
                    <Text fontSize="sm">
                      • Consider requesting less memory. Your job used only {efficiency.memory.toFixed(0)}% of allocated memory.
                    </Text>
                  )}
                  <Text fontSize="sm" color="fg.muted">
                    Rightsizing your resource requests helps improve cluster utilization and may reduce your job&apos;s queue time.
                  </Text>
                </VStack>
              </Card.Body>
            </Card.Root>
          </Box>
        </>
      )}
    </VStack>
  )
})

PerformanceMetricsTab.displayName = 'PerformanceMetricsTab'
