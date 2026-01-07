import { useEffect, useMemo, useRef } from 'react'
import { HStack, SimpleGrid, Text, VStack, Tag, Progress, Stat, Badge } from '@chakra-ui/react'

import type { PartitionResponse } from '../../client'

interface Props {
  partitions: PartitionResponse[]
  isFetching?: boolean
}

export const PartitionOverviewCards = ({ partitions, isFetching }: Props) => {
  const totals = useMemo(() => {
    const totalPartitions = partitions.length
    let totalCpus = 0
    let totalGpus = 0
    let gpusInUse = 0
    let pendingJobs = 0
    for (const p of partitions) {
      totalCpus += p.total_cpus ?? 0
      totalGpus += p.total_gpus ?? 0
      gpusInUse += p.gpus_in_use?.length ?? 0
      pendingJobs += p.jobs_pending?.length ?? 0
    }
    const gpuUtilPct = totalGpus > 0 ? Math.round((gpusInUse / totalGpus) * 100) : 0
    return { totalPartitions, totalCpus, totalGpus, gpusInUse, gpuUtilPct, pendingJobs }
  }, [partitions])

  // Simple trend indicator comparing to last snapshot
  const lastPendingRef = useRef<number>(totals.pendingJobs)
  const trend: 'up' | 'down' | 'flat' =
    totals.pendingJobs > lastPendingRef.current ? 'up' : totals.pendingJobs < lastPendingRef.current ? 'down' : 'flat'
  useEffect(() => {
    lastPendingRef.current = totals.pendingJobs
  }, [totals.pendingJobs])

  const gpuColor = totals.gpuUtilPct > 90 ? 'red' : totals.gpuUtilPct > 50 ? 'yellow' : 'green'

  return (
    <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} gap={3} w="100%">
      {/* Card 1: Total Partitions */}
      <Stat.Root borderWidth="1px" borderColor="gray.200" rounded="md" p={2} bg="white">
        <Stat.Label fontSize="sm" color="gray.600">Total Partitions</Stat.Label>
        <HStack gap={1} align="baseline">
          <Stat.ValueText fontSize="2xl" fontWeight="semibold">{totals.totalPartitions}</Stat.ValueText>
          {isFetching && <Tag.Root size="sm"><Tag.Label>updating</Tag.Label></Tag.Root>}
        </HStack>
        <Stat.HelpText fontSize="xs" color="gray.500">Aggregated for current cluster</Stat.HelpText>
      </Stat.Root>

      {/* Card 2: Total Resources */}
      <Stat.Root borderWidth="1px" borderColor="gray.200" rounded="md" p={2} bg="white">
        <Stat.Label fontSize="sm" color="gray.600">Total Resources</Stat.Label>
        <HStack gap={4} wrap="wrap">
          <VStack align="start" gap={0}>
            <Text fontSize="xs" color="gray.500">CPUs</Text>
            <Stat.ValueText as="span" fontSize="lg" fontWeight="semibold">{totals.totalCpus}</Stat.ValueText>
          </VStack>
          <VStack align="start" gap={0}>
            <Text fontSize="xs" color="gray.500">GPUs</Text>
            <Stat.ValueText as="span" fontSize="lg" fontWeight="semibold">{totals.totalGpus}</Stat.ValueText>
          </VStack>
        </HStack>
      </Stat.Root>

      {/* Card 3: GPU Utilization */}
      <Stat.Root borderWidth="1px" borderColor="gray.200" rounded="md" p={2} bg="white">
        <Stat.Label fontSize="sm" color="gray.600">GPU Utilization</Stat.Label>
        <HStack justify="space-between" w="100%" mb={1}>
          <Text fontSize="xs" color="gray.600">{totals.gpusInUse} / {totals.totalGpus}</Text>
          <Tag.Root size="sm" colorPalette={gpuColor}><Tag.Label>{totals.gpuUtilPct}%</Tag.Label></Tag.Root>
        </HStack>
        <Progress.Root value={totals.gpuUtilPct} max={100} colorPalette={gpuColor} size="xs">
          <Progress.Track>
            <Progress.Range />
          </Progress.Track>
        </Progress.Root>
      </Stat.Root>

      {/* Card 4: Queue Pressure */}
      <Stat.Root borderWidth="1px" borderColor="gray.200" rounded="md" p={2} bg="white">
        <Stat.Label fontSize="sm" color="gray.600">Queue Pressure</Stat.Label>
        <HStack gap={1} align="center">
          <Stat.ValueText fontSize="2xl" fontWeight="semibold">{totals.pendingJobs}</Stat.ValueText>
          {trend !== 'flat' && (
            <Badge colorPalette={trend === 'up' ? 'green' : 'red'} variant="subtle" px="1" gap="0">
              {trend === 'up' ? <Stat.UpIndicator /> : <Stat.DownIndicator />}
            </Badge>
          )}
        </HStack>
        <Stat.HelpText fontSize="xs" color="gray.500">Pending jobs across partitions</Stat.HelpText>
      </Stat.Root>
    </SimpleGrid>
  )
}
