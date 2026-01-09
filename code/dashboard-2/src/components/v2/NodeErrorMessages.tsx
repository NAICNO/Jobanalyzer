import { useMemo } from 'react'
import { VStack, Text, HStack, Spinner, Alert } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'

import { getClusterByClusterNodesByNodenameErrorMessagesOptions } from '../../client/@tanstack/react-query.gen'
import type { ErrorMessageResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'

type Props = {
  cluster: string
  nodename: string
}

export const NodeErrorMessages = ({ cluster, nodename }: Props) => {
  const client = useClusterClient(cluster)
  
  if (!client) {
    return <Spinner />
  }
  
  const queryOpts = getClusterByClusterNodesByNodenameErrorMessagesOptions({
    path: { cluster, nodename },
    client,
  })

  const { data, isLoading, isError, error } = useQuery({
    ...queryOpts,
    enabled: !!cluster && !!nodename,
  })

  const messages: ErrorMessageResponse[] = useMemo(() => {
    if (!data) return []
    const map = data as Record<string, ErrorMessageResponse>
    return Object.values(map)
  }, [data])

  return (
    <VStack align="start" w="100%" mt={4} gap={2}>
      <Text fontWeight="semibold">Error messages</Text>
      {isLoading && (
        <HStack>
          <Spinner size="sm" />
          <Text>Loading error messages…</Text>
        </HStack>
      )}
      {isError && (
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>
            {error instanceof Error ? error.message : 'Failed to load error messages.'}
          </Alert.Description>
        </Alert.Root>
      )}
      {!isLoading && !isError && messages.length === 0 && (
        <Text color="gray.600">No recent error messages.</Text>
      )}
      {!isLoading && !isError && messages.length > 0 && (
        <VStack align="start" w="100%" gap={2}>
          {messages.map((msg, i) => (
            <Alert.Root key={`${msg.node}-${i}`} status="warning" variant="subtle">
              <Alert.Indicator />
              <VStack align="start" p={1} gap={0}>
                <Text fontWeight="semibold">{msg.node}</Text>
                <Text fontSize="sm" color="gray.700">{msg.details}</Text>
                <Text fontSize="xs" color="gray.500">{new Date(msg.time).toLocaleString()}</Text>
              </VStack>
            </Alert.Root>
          ))}
        </VStack>
      )}
    </VStack>
  )
}
