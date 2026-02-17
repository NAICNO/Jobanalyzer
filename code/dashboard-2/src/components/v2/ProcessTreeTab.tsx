import { useCallback, useMemo, memo } from 'react'
import {
  Box,
  VStack,
  HStack,
  Text,
  Spinner,
  Alert,
  Button,
} from '@chakra-ui/react'
import { LuExternalLink } from 'react-icons/lu'
import type { Client } from '../../client/client/types.gen'
import { useJobProcessTree } from '../../hooks/v2/useJobQueries'
import { ProcessTreeView } from './ProcessTreeView'

interface Props {
  cluster: string
  jobId: number
  client: Client | null
}

export const ProcessTreeTab = memo(({ cluster, jobId, client }: Props) => {
  const { data: treeData, isLoading, isError, error } = useJobProcessTree({
    cluster,
    jobId,
    client,
  })

  const metadata = useMemo(() => {
    if (!treeData?.nodes) return null
    const firstKey = Object.keys(treeData.nodes)[0]
    return firstKey ? treeData.nodes[firstKey].metadata : null
  }, [treeData])

  const handleOpenFullView = useCallback(() => {
    window.open(`/v2/${cluster}/jobs/${jobId}/process-tree`, '_blank')
  }, [cluster, jobId])

  if (isLoading) {
    return (
      <Box w="100%" h="400px" display="flex" alignItems="center" justifyContent="center">
        <Spinner size="lg" />
      </Box>
    )
  }

  if (isError) {
    return (
      <Alert.Root status="error">
        <Alert.Indicator />
        <Alert.Description>
          Failed to load process tree: {error?.message || 'Unknown error'}
        </Alert.Description>
      </Alert.Root>
    )
  }

  if (!treeData || !Object.keys(treeData.nodes ?? {}).length) {
    return (
      <Alert.Root status="info">
        <Alert.Indicator />
        <Alert.Description>No process tree data available for this job.</Alert.Description>
      </Alert.Root>
    )
  }

  return (
    <VStack w="100%" align="start" gap={3}>
      <HStack w="100%" justify="space-between" align="center">
        {metadata && (
          <HStack gap={3} fontSize="sm">
            <HStack gap={1} px={3} py={1} borderWidth="1px" borderColor="border" rounded="md">
              <Text color="fg.muted">Processes</Text>
              <Text fontWeight="bold">{metadata.total_processes}</Text>
            </HStack>
            <HStack gap={1} px={3} py={1} borderWidth="1px" borderColor="border" rounded="md">
              <Text color="fg.muted">Max Depth</Text>
              <Text fontWeight="bold">{metadata.max_depth}</Text>
            </HStack>
            <HStack gap={1} px={3} py={1} borderWidth="1px" borderColor="border" rounded="md">
              <Text color="fg.muted">Root PID</Text>
              <Text fontWeight="bold">{metadata.root_pid}</Text>
            </HStack>
          </HStack>
        )}
        <Button
          size="xs"
          variant="outline"
          onClick={handleOpenFullView}
        >
          <LuExternalLink />
          Open Full View
        </Button>
      </HStack>

      <Box w="100%" h="600px" borderWidth="1px" borderColor="border" rounded="md">
        <ProcessTreeView cluster={cluster} jobId={jobId} client={client} />
      </Box>
    </VStack>
  )
})

ProcessTreeTab.displayName = 'ProcessTreeTab'
