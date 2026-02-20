import { VStack, Heading, Alert } from '@chakra-ui/react'
import { useNavigate, useParams } from 'react-router'

import { NodeOverviewCards } from '../../components/v2/NodeOverviewCards'
import { NodesTable } from '../../components/v2/NodesTable'
import { NodeDetailDrawer } from '../../components/v2/NodeDetailDrawer'

export const NodesPage = () => {
  const { clusterName, nodename } = useParams()
  const navigate = useNavigate()

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
      <VStack align="start" w="100%" p={4} pt={2} mb={4} gap={4}>
        <NodeOverviewCards cluster={clusterName} />
      </VStack>
      <VStack px={4} gap={2} align="start" w="100%">
        <Heading size="md">All Nodes</Heading>
        <NodesTable
          clusterName={clusterName}
          onNodeClick={(nodeName) => navigate(`/v2/${clusterName}/nodes/${nodeName}`)}
        />
      </VStack>
      <NodeDetailDrawer
        cluster={clusterName}
        nodename={nodename}
        onClose={() => navigate(`/v2/${clusterName}/nodes`)}
      />
    </>
  )
}
