import { useMemo } from 'react'
import { DataList, VStack, Text, Alert, Spinner, HStack, Tag } from '@chakra-ui/react'

import { useClusterClient } from '../../hooks/useClusterClient'
import { useNodeInfo } from '../../hooks/v2/useNodeQueries'
import type { NodeInfoResponse } from '../../client'

interface Props {
  cluster: string
  nodename: string
}

export const NodeInfoSummary = ({ cluster, nodename }: Props) => {
  const client = useClusterClient(cluster)
  
  const { data: nodeInfoData, isLoading, isError, error } = useNodeInfo({ cluster, nodename, client })

  const info: NodeInfoResponse | undefined = useMemo(() => {
    if (!nodeInfoData) return undefined
    const map = nodeInfoData as Record<string, NodeInfoResponse>
    return map[nodename] ?? Object.values(map)[0]
  }, [nodeInfoData, nodename])

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
  const timeStr = new Date(info.time).toLocaleString()

  return (
    <VStack align="start" w="100%" gap={2}>
      <Text fontWeight="semibold">Node Info</Text>
      <DataList.Root maxW="xl" width="100%" variant="subtle" size="md" orientation="horizontal">
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
        {Array.isArray(info.partitions) && info.partitions.length > 0 && (
          <DataList.Item>
            <DataList.ItemLabel>Partitions</DataList.ItemLabel>
            <DataList.ItemValue>
              <HStack wrap="wrap" gap={1}>
                {info.partitions.map((p: string) => (
                  <Tag.Root key={p} size="lg" variant="surface" colorPalette="blue">
                    <Tag.Label>{p}</Tag.Label>
                  </Tag.Root>
                ))}
              </HStack>
            </DataList.ItemValue>
          </DataList.Item>
        )}
      </DataList.Root>
    </VStack>
  )
}
