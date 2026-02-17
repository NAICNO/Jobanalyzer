import { VStack, Heading, Spinner, Alert, Box, HStack, Text, Container } from '@chakra-ui/react'
import { useParams, Link } from 'react-router'
import { useMemo } from 'react'
import { LuArrowLeft } from 'react-icons/lu'

import { useClusterClient } from '../../hooks/useClusterClient'
import { useNodeTopology } from '../../hooks/v2/useNodeQueries'

export const NodeTopologyPage = () => {
  const { clusterName, nodename } = useParams()

  const client = useClusterClient(clusterName)
  const { data, isLoading, isError, error } = useNodeTopology({ cluster: clusterName ?? '', nodename: nodename ?? '', client })

  const svg = useMemo<string | undefined>(() => {
    if (!data) return undefined
    if (typeof data === 'string') return data
    if (typeof data === 'object') {
      const obj = data as Record<string, unknown>
      return (obj.svg as string | undefined)
        || (obj.body as string | undefined)
        || (obj.data as string | undefined)
        || undefined
    }
    return undefined
  }, [data])

  console.log('SVG Data:', svg)

  if (!clusterName || !nodename) {
    return (
      <VStack p={4} align="start">
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>Missing cluster or node name in route.</Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  return (
    <Box w="100%" minH="100vh" bg="gray.50">
      <VStack w="100%" align="start" gap={4} p={6}>
        <Container maxW="8xl">
          <HStack w="100%" justify="space-between">
            <HStack gap={3}>
              <Link to={`/v2/${clusterName}/nodes/${nodename}`}>
                <HStack
                  px={3}
                  py={2}
                  bg="white"
                  borderWidth="1px"
                  borderColor="gray.200"
                  rounded="md"
                  _hover={{ bg: 'gray.100' }}
                  cursor="pointer"
                >
                  <LuArrowLeft />
                  <Text fontSize="sm">Back to Node</Text>
                </HStack>
              </Link>
              <Heading size="lg">{nodename} - Topology</Heading>
            </HStack>
          </HStack>
        </Container>

        <Box
          w="100%"
          bg="white"
          borderWidth="1px"
          borderColor="gray.200"
          rounded="md"
          overflowX="auto"
        >
          {isLoading && (
            <VStack h="400px" justify="center" align="center">
              <Spinner size="lg" />
              <Text>Loading topology…</Text>
            </VStack>
          )}
          {isError && (
            <Box p={6}>
              <Alert.Root status="error">
                <Alert.Indicator />
                <Alert.Description>
                  {error instanceof Error ? error.message : 'Failed to load topology.'}
                </Alert.Description>
              </Alert.Root>
            </Box>
          )}
          {!isLoading && !isError && svg && (
            <Box
              p={6}
              display="inline-block"
              minW="100%"
              dangerouslySetInnerHTML={{ __html: svg }}
            />
          )}
          {!isLoading && !isError && !svg && (
            <VStack h="400px" justify="center" align="center">
              <Text color="gray.600">No topology available.</Text>
            </VStack>
          )}
        </Box>
      </VStack>
    </Box>
  )
}
