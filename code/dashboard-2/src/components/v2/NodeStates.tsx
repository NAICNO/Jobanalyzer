import { useMemo, useState } from 'react'
import { VStack, Text, HStack, Spinner, Alert, Box, Accordion, Tag } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'

import { getClusterByClusterNodesByNodenameStatesOptions } from '../../client/@tanstack/react-query.gen'
import type { NodeStateResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'

interface Props {
  cluster: string
  nodename: string
  initialCollapsed?: boolean
}

export const NodeStates = ({ cluster, nodename, initialCollapsed = false }: Props) => {
  const client = useClusterClient(cluster)
  const [value, setValue] = useState<string[]>(initialCollapsed ? [] : ['states'])
  const expanded = value.includes('states')

  if (!client) {
    return null
  }

  const queryOpts = getClusterByClusterNodesByNodenameStatesOptions({
    path: { cluster, nodename },
    client,
  })

  const { data, isLoading, isError, error } = useQuery({
    ...queryOpts,
    enabled: !!cluster && !!nodename && expanded,
  })

  const items = useMemo<NodeStateResponse[]>(() => (Array.isArray(data) ? (data as NodeStateResponse[]) : []), [data])

  return (
    <VStack align="start" w="100%" mt={2} gap={2}>
      <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md">
        <Accordion.Root value={value} onValueChange={(e) => setValue(e.value)} collapsible variant="plain">
          <Accordion.Item value="states">
            <Accordion.ItemTrigger px={3} py={2} _hover={{ bg: 'gray.50' }}>
              <HStack flex="1" justify="space-between" align="center">
                <Text fontWeight="semibold">States</Text>
              </HStack>
              <Accordion.ItemIndicator />
            </Accordion.ItemTrigger>
            <Accordion.ItemContent>
              <Accordion.ItemBody>
                {isLoading && (
                  <HStack>
                    <Spinner size="sm" />
                    <Text>Loading states…</Text>
                  </HStack>
                )}
                {isError && (
                  <Alert.Root status="error">
                    <Alert.Indicator />
                    <Alert.Description>
                      {error instanceof Error ? error.message : 'Failed to load states.'}
                    </Alert.Description>
                  </Alert.Root>
                )}
                {!isLoading && !isError && items.length > 0 && (
                  <VStack align="start" gap={2} paddingX={4}>
                    {items.map((it, idx) => {
                      const when = it.time.toLocaleString()
                      return (
                        <HStack key={idx} justify="space-between" w="100%" align="center">
                          <HStack wrap="wrap" gap={2}>
                            {it.states.map((s, i) => (
                              <Tag.Root key={i} colorPalette="blue" variant="surface" size="sm">
                                <Tag.Label>{s}</Tag.Label>
                              </Tag.Root>
                            ))}
                          </HStack>
                          <Text color="gray.600" fontSize="sm">{when}</Text>
                        </HStack>
                      )
                    })}
                  </VStack>
                )}
                {!isLoading && !isError && items.length === 0 && (
                  <Text color="gray.600">No states available.</Text>
                )}
              </Accordion.ItemBody>
            </Accordion.ItemContent>
          </Accordion.Item>
        </Accordion.Root>
      </Box>
    </VStack>
  )
}
