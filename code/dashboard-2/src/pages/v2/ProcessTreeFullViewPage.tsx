import { useMemo } from 'react'
import { Box, HStack, Text, Badge, Spinner, Alert } from '@chakra-ui/react'
import { useParams } from 'react-router'
import { useClusterClient } from '../../hooks/useClusterClient'
import { ProcessTreeView } from '../../components/v2/ProcessTreeView'

export const ProcessTreeFullViewPage = () => {
  const { clusterName, jobId } = useParams<{ clusterName: string; jobId: string }>()
  const client = useClusterClient(clusterName)

  const jobIdNum = useMemo(() => {
    const parsed = parseInt(jobId ?? '', 10)
    return isNaN(parsed) ? undefined : parsed
  }, [jobId])

  if (!client) {
    return (
      <Box w="100vw" h="100vh" display="flex" alignItems="center" justifyContent="center">
        <Spinner size="lg" />
      </Box>
    )
  }

  if (!clusterName || !jobIdNum) {
    return (
      <Box w="100vw" h="100vh" p={4}>
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>Invalid job ID or cluster name</Alert.Description>
        </Alert.Root>
      </Box>
    )
  }

  return (
    <Box w="100vw" h="100vh" overflow="hidden">
      <HStack px={4} py={2} borderBottomWidth="1px" borderColor="border" justify="space-between">
        <HStack gap={2}>
          <Text fontSize="sm" fontWeight="bold">Process Tree</Text>
          <Badge size="sm" variant="subtle">Job {jobIdNum}</Badge>
          <Text fontSize="xs" color="fg.muted">{clusterName}</Text>
        </HStack>
      </HStack>
      <Box w="100%" h="calc(100vh - 41px)">
        <ProcessTreeView cluster={clusterName} jobId={jobIdNum} client={client} showTimeline />
      </Box>
    </Box>
  )
}
