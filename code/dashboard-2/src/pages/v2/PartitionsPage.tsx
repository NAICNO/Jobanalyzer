import { VStack, Text, Heading, Listbox, Input, useFilter, Spinner, Alert, createListCollection, Box, HStack, Badge } from '@chakra-ui/react'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'

import { getClusterByClusterPartitionsOptions } from '../../client/@tanstack/react-query.gen'
import { useClusterClient } from '../../hooks/useClusterClient'
import { ResizableColumns } from '../../components/v2/ResizableColumns'
import { PartitionSummary } from '../../components/v2/PartitionSummary'
import { PartitionQueueOverview } from '../../components/v2/PartitionQueueOverview'
import { PartitionNodes } from '../../components/v2/PartitionNodes'
import { PartitionGpus } from '../../components/v2/PartitionGpus'
import { PartitionOverviewCards } from '../../components/v2/PartitionOverviewCards'
import type { PartitionResponse } from '../../client'

export const PartitionsPage = () => {
  const { clusterName, partitionName } = useParams()
  const navigate = useNavigate()
  const minLeftWidth = 220
  const maxLeftWidth = 640
  const handleWidth = 6

  const client = useClusterClient(clusterName)
  if (!client) {
    return <Spinner />
  }

  // Fetch partitions for the cluster from /cluster/:cluster/partitions
  const baseQueryOptions = getClusterByClusterPartitionsOptions({
    path: { cluster: clusterName ?? '' },
    client,
  })
  const { data, isLoading, isError, error } = useQuery({
    ...baseQueryOptions,
    enabled: !!clusterName,
  })

  const partitionsMap = (data ?? {}) as Record<string, PartitionResponse>
  const partitions = Object.values(partitionsMap)

  type PartitionItem = { value: string; label: string; data: PartitionResponse }
  const items = useMemo<PartitionItem[]>(
    () => partitions.map(p => ({ value: p.name ?? '', label: p.name ?? '', data: p })),
    [partitions]
  )

  const [selectedPartitionValue, setSelectedPartitionValue] = useState<string[]>([])
  const { contains } = useFilter({ sensitivity: 'base' })
  const [filterValue, setFilterValue] = useState('')
  const filteredItems = useMemo(
    () => items.filter((it) => contains(it.label, filterValue)),
    [items, filterValue, contains]
  )
  const collection = useMemo(
    () => createListCollection<PartitionItem>({ items: filteredItems }),
    [filteredItems]
  )

  // Sync selection with URL param and initialize when data arrives
  useEffect(() => {
    if (items.length === 0) return

    // If URL contains a partitionName and it exists, sync selection to it
    if (partitionName) {
      const exists = items.some(it => it.value === partitionName)
      if (exists) {
        if (selectedPartitionValue[0] !== partitionName) {
          setSelectedPartitionValue([partitionName])
        }
        return
      }
    }

    // Otherwise, if nothing is selected, select first and update URL for deep-linking
    if (selectedPartitionValue.length === 0 && items.length > 0) {
      const first = items[0].value
      setSelectedPartitionValue([first])
      if (clusterName) {
        navigate(`/v2/${clusterName}/partitions/${first}`, { replace: true })
      }
    }
  }, [items, partitionName, selectedPartitionValue, clusterName, navigate])

  const selectedPartition = items.find(p => p.value === selectedPartitionValue[0])

  if (!clusterName) {
    return (
      <VStack p={4} align="start">
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>Missing cluster name in route.</Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  return (
    <>
      {/* Overview cards across all partitions */}
      {partitions.length > 0 && (
        <Box px={4} pt={4} pb={2} mb={4}>
          <PartitionOverviewCards partitions={partitions} />
        </Box>
      )}

      <ResizableColumns
        height="calc(100vh - 200px)"
        initialLeftWidth={320}
        minLeftWidth={minLeftWidth}
        maxLeftWidth={maxLeftWidth}
        handleWidth={handleWidth}
        storageKey="partitionsPage.leftWidth"
        left={
          <VStack p={4} gap={4} align="start">
            <Heading size="md">Partitions</Heading>
            {isLoading && (
              <Box display="flex" alignItems="center" gap={2}>
                <Spinner size="sm" />
                <Text>Loading partitions…</Text>
              </Box>
            )}
            {isError && (
              <Alert.Root status="error">
                <Alert.Indicator />
                <Alert.Description>
                  {error instanceof Error ? error.message : 'Failed to load partitions.'}
                </Alert.Description>
              </Alert.Root>
            )}
            {!isLoading && !isError && partitions.length === 0 && (
              <Alert.Root status="info">
                <Alert.Indicator />
                <Alert.Description>No partitions found.</Alert.Description>
              </Alert.Root>
            )}
            {!isLoading && !isError && partitions.length > 0 && (
              <Listbox.Root
                collection={collection}
                value={selectedPartitionValue}
                onValueChange={(details) => {
                  const sel = details.value[0]
                  setSelectedPartitionValue(details.value)
                  // Only navigate if the URL doesn't already match
                  if (sel && clusterName && sel !== partitionName) {
                    navigate(`/v2/${clusterName}/partitions/${sel}`)
                  }
                }}
              >
                <Listbox.Label>Available Partitions</Listbox.Label>
                <Listbox.Input as={Input} placeholder="Type to filter partitions..." onChange={(e) => setFilterValue(e.target.value)} />
                <Listbox.Content>
                  {collection.items.map((partition) => {
                    const gpuUtil = partition.data.total_gpus
                      ? Math.round(((partition.data.gpus_in_use?.length ?? 0) / partition.data.total_gpus) * 100)
                      : 0
                    const statusColor = gpuUtil > 90 ? 'red' : gpuUtil > 50 ? 'yellow' : 'green'

                    return (
                      <Listbox.Item item={partition} key={partition.value}>
                        <VStack align="start" gap={0.5} flex={1}>
                          <HStack justify="space-between" w="100%">
                            <Listbox.ItemText fontWeight="medium">{partition.label}</Listbox.ItemText>
                            <Badge colorPalette={statusColor} size="xs">{gpuUtil}%</Badge>
                          </HStack>
                          <Text fontSize="xs" color="gray.500">
                            {partition.data.nodes?.length ?? 0} nodes • {partition.data.total_cpus ?? 0} CPUs • {partition.data.total_gpus ?? 0} GPUs
                          </Text>
                        </VStack>

                      </Listbox.Item>
                    )
                  })}
                  {collection.items.length === 0 && (
                    <Text color="gray.500" p={2}>No partitions found.</Text>
                  )}
                </Listbox.Content>
              </Listbox.Root>
            )}
          </VStack>
        }
        right={
          <VStack flex={1} p={4} gap={4} align="start" minW="0">
            {selectedPartition ? (
              <>
                <Heading size="md">{selectedPartition.label}</Heading>
                <PartitionSummary partition={selectedPartition.data} />
                <PartitionQueueOverview partition={selectedPartition.data} />
                <PartitionNodes partition={selectedPartition.data} />
                <PartitionGpus partition={selectedPartition.data} />
              </>
            ) : (
              <Text>Select a partition to see details</Text>
            )}
          </VStack>
        }
      />
    </>
  )
}
