import { DataList, VStack, Text, Box, HStack, Progress } from '@chakra-ui/react'

import type { PartitionResponse } from '../../client'

interface Props {
  partition: PartitionResponse
}

export const PartitionSummary = ({ partition }: Props) => {
  const timeStr = partition.time.toLocaleString()
  
  const totalGpus = partition.total_gpus ?? 0
  const gpusReserved = partition.gpus_reserved ?? 0
  const gpusInUse = partition.gpus_in_use?.length ?? 0
  
  const totalCpus = partition.total_cpus ?? 0
  const cpusInUse = (partition.jobs_running ?? []).reduce((sum, job) => sum + (job.requested_cpus ?? 0), 0)
  
  const gpuUtilPercent = totalGpus > 0 ? Math.round((gpusInUse / totalGpus) * 100) : 0
  const cpuUtilPercent = totalCpus > 0 ? Math.round((cpusInUse / totalCpus) * 100) : 0
  
  const nodesCount = partition.nodes?.length ?? 0
  const nodesCompact = Array.isArray(partition.nodes_compact) && partition.nodes_compact.length > 0
    ? partition.nodes_compact.join(', ')
    : 'N/A'

  return (
    <VStack align="start" w="100%" gap={4}>
      <DataList.Root maxW="xl" width="100%" variant="subtle" size="md" orientation="horizontal">
        <DataList.Item>
          <DataList.ItemLabel>Partition</DataList.ItemLabel>
          <DataList.ItemValue fontWeight="semibold">{partition.name}</DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>Cluster</DataList.ItemLabel>
          <DataList.ItemValue>{partition.cluster}</DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>Last Updated</DataList.ItemLabel>
          <DataList.ItemValue>{timeStr}</DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>Nodes</DataList.ItemLabel>
          <DataList.ItemValue>{nodesCount} ({nodesCompact})</DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>Total CPUs</DataList.ItemLabel>
          <DataList.ItemValue>{totalCpus}</DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>Total GPUs</DataList.ItemLabel>
          <DataList.ItemValue>{totalGpus} (Reserved: {gpusReserved})</DataList.ItemValue>
        </DataList.Item>
      </DataList.Root>

      {/* Capacity & Utilization */}
      <VStack align="start" w="100%" gap={3}>
        <Text fontWeight="semibold" fontSize="md">Capacity & Utilization</Text>

        <Box w="100%">
          <HStack justify="space-between" mb={1}>
            <Text fontSize="sm" fontWeight="medium">GPU Usage</Text>
            <Text fontSize="sm" color="gray.600">{gpusInUse} / {totalGpus} ({gpuUtilPercent}%)</Text>
          </HStack>
          <Progress.Root value={gpuUtilPercent} max={100} colorPalette={gpuUtilPercent > 90 ? 'red' : gpuUtilPercent > 50 ? 'yellow' : 'green'}>
            <Progress.Track>
              <Progress.Range />
            </Progress.Track>
          </Progress.Root>
        </Box>

        <Box w="100%">
          <HStack justify="space-between" mb={1}>
            <Text fontSize="sm" fontWeight="medium">CPU Usage (Estimated)</Text>
            <Text fontSize="sm" color="gray.600">{cpusInUse} / {totalCpus} ({cpuUtilPercent}%)</Text>
          </HStack>
          <Progress.Root value={cpuUtilPercent} max={100} colorPalette={cpuUtilPercent > 90 ? 'red' : cpuUtilPercent > 50 ? 'yellow' : 'green'}>
            <Progress.Track>
              <Progress.Range />
            </Progress.Track>
          </Progress.Root>
        </Box>

        <HStack gap={4} flexWrap="wrap">
          <Box>
            <Text fontSize="xs" color="gray.500">Running Jobs</Text>
            <Text fontSize="lg" fontWeight="semibold">{partition.jobs_running?.length ?? 0}</Text>
          </Box>
          <Box>
            <Text fontSize="xs" color="gray.500">Pending Jobs</Text>
            <Text fontSize="lg" fontWeight="semibold">{partition.jobs_pending?.length ?? 0}</Text>
          </Box>
        </HStack>
      </VStack>
    </VStack>
  )
}
