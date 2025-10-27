import { useEffect, useMemo, useState } from 'react'
import { DataList, VStack, Text, Listbox, createListCollection, Accordion, Box, Alert, Spinner, HStack } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'

import { getClusterByClusterNodesByNodenameInfoOptions } from '../../client/@tanstack/react-query.gen'
import type { NodeInfoResponse, GpuCardResponse } from '../../client'

interface Props {
  cluster: string
  nodename: string
}

export const NodeInfoSummary = ({ cluster, nodename }: Props) => {
  const infoQueryOpts = getClusterByClusterNodesByNodenameInfoOptions({
    path: { cluster, nodename },
  })
  const { data: nodeInfoData, isLoading, isError, error } = useQuery({
    ...infoQueryOpts,
    enabled: !!cluster && !!nodename,
  })

  const info: NodeInfoResponse | undefined = useMemo(() => {
    if (!nodeInfoData) return undefined
    const map = nodeInfoData as Record<string, NodeInfoResponse>
    return map[nodename] ?? Object.values(map)[0]
  }, [nodeInfoData, nodename])

  // Partitions selection state and collections must be declared unconditionally (before any returns)
  const [selectedPartition, setSelectedPartition] = useState<string[]>([])
  const partitionItems = useMemo(() => (
    (info?.partitions ?? []).map((p: string) => ({ value: p, label: p }))
  ), [info?.partitions])
  const partitionsCollection = useMemo(
    () => createListCollection<{ value: string; label: string }>({ items: partitionItems }),
    [partitionItems]
  )

  // Initialize or reset selected partition when partitions change
  useEffect(() => {
    if (partitionItems.length > 0) setSelectedPartition([partitionItems[0].value])
    else setSelectedPartition([])
  }, [partitionItems])

  if (isLoading) {
    return (
      <VStack align="start" w="100%" gap={2}>
        <HStack>
          <Spinner size="sm" />
          <Text>Loading node info…</Text>
        </HStack>
      </VStack>
    )
  }

  if (isError) {
    return (
      <VStack align="start" w="100%" gap={2}>
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>
            {error instanceof Error ? error.message : 'Failed to load node info.'}
          </Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  if (!info) {
    return <Text color="gray.600">No additional details available.</Text>
  }

  const memGb = Math.round((info.memory ?? 0) / (1024 * 1024))
  const timeStr = info.time.toLocaleString()

  return (
    <VStack align="start" w="100%" gap={4}>
      <DataList.Root maxW="xl" width="100%" variant="subtle" size="md" orientation="horizontal">
        <DataList.Item>
          <DataList.ItemLabel>Hostname</DataList.ItemLabel>
          <DataList.ItemValue>{info.node}</DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>Cluster</DataList.ItemLabel>
          <DataList.ItemValue>{info.cluster}</DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>OS</DataList.ItemLabel>
          <DataList.ItemValue>{info.os_name} {info.os_release}</DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>Architecture</DataList.ItemLabel>
          <DataList.ItemValue>{info.architecture}</DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>CPU</DataList.ItemLabel>
          <DataList.ItemValue>
            {info.sockets} sockets × {info.cores_per_socket} cores × {info.threads_per_core} threads
          </DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>CPU Model</DataList.ItemLabel>
          <DataList.ItemValue>{info.cpu_model}</DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>Memory</DataList.ItemLabel>
          <DataList.ItemValue>{memGb} GB</DataList.ItemValue>
        </DataList.Item>
        <DataList.Item>
          <DataList.ItemLabel>Last Updated</DataList.ItemLabel>
          <DataList.ItemValue>{timeStr}</DataList.ItemValue>
        </DataList.Item>
      </DataList.Root>

      {Array.isArray(info.partitions) && info.partitions.length > 0 && (
        <VStack align="start" w="100%" gap={1}>
          <Text fontWeight="semibold">Partitions</Text>
          <Listbox.Root
            collection={partitionsCollection}
            value={selectedPartition}
            onValueChange={(details) => setSelectedPartition(details.value)}
            width="100%"
          >
            <Listbox.Content maxH="160px" overflowY="auto">
              {partitionsCollection.items.map((part) => (
                <Listbox.Item item={part} key={part.value}>
                  <Listbox.ItemText>{part.label}</Listbox.ItemText>
                  
                </Listbox.Item>
              ))}
            </Listbox.Content>
          </Listbox.Root>
        </VStack>
      )}

      {Array.isArray(info.cards) && info.cards.length > 0 && (
        <VStack align="start" w="100%" gap={2}>
          <Text fontWeight="semibold">GPUs ({info.cards.length})</Text>
          <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md">
            <Accordion.Root multiple variant="outline" px={4}>
              {info.cards.map((card: GpuCardResponse, idx: number) => (
                <Accordion.Item value={String(card.uuid ?? idx)} key={card.uuid ?? idx}>
                  <Accordion.ItemTrigger px={3} py={2} _hover={{ bg: 'gray.50' }}>
                    <Text fontWeight="semibold">GPU {card.index ?? idx}: {card.manufacturer} {card.model}</Text>
                    <Accordion.ItemIndicator />
                  </Accordion.ItemTrigger>
                  <Accordion.ItemContent>
                    <Accordion.ItemBody>
                      <DataList.Root width="100%" variant="subtle" size="sm">
                        <DataList.Item>
                          <DataList.ItemLabel>Architecture</DataList.ItemLabel>
                          <DataList.ItemValue>{card.architecture}</DataList.ItemValue>
                        </DataList.Item>
                        <DataList.Item>
                          <DataList.ItemLabel>Address</DataList.ItemLabel>
                          <DataList.ItemValue>{card.address}</DataList.ItemValue>
                        </DataList.Item>
                        <DataList.Item>
                          <DataList.ItemLabel>Driver</DataList.ItemLabel>
                          <DataList.ItemValue>{card.driver}</DataList.ItemValue>
                        </DataList.Item>
                        <DataList.Item>
                          <DataList.ItemLabel>Firmware</DataList.ItemLabel>
                          <DataList.ItemValue>{card.firmware}</DataList.ItemValue>
                        </DataList.Item>
                        <DataList.Item>
                          <DataList.ItemLabel>Power Limits</DataList.ItemLabel>
                          <DataList.ItemValue>{card.min_power_limit}–{card.max_power_limit} W</DataList.ItemValue>
                        </DataList.Item>
                        <DataList.Item>
                          <DataList.ItemLabel>Clocks (max)</DataList.ItemLabel>
                          <DataList.ItemValue>CE {card.max_ce_clock} MHz • Mem {card.max_memory_clock} MHz</DataList.ItemValue>
                        </DataList.Item>
                        <DataList.Item>
                          <DataList.ItemLabel>UUID</DataList.ItemLabel>
                          <DataList.ItemValue>{card.uuid}</DataList.ItemValue>
                        </DataList.Item>
                      </DataList.Root>
                    </Accordion.ItemBody>
                  </Accordion.ItemContent>
                </Accordion.Item>
              ))}
            </Accordion.Root>
          </Box>
        </VStack>
      )}
    </VStack>
  )
}
