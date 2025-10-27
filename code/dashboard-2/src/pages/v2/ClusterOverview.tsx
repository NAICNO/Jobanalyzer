import { VStack, Text } from '@chakra-ui/react'
import { useParams } from 'react-router'

import { ClusterOverviewCards } from '../../components/v2/ClusterOverviewCards'
import { ClusterHealthStatus } from '../../components/v2/ClusterHealthStatus'
import { ClusterResourceDistribution } from '../../components/v2/ClusterResourceDistribution'
import { ClusterQueueActivity } from '../../components/v2/ClusterQueueActivity'
import { ClusterTimebasedActivity } from '../../components/v2/ClusterTimebasedActivity'
import { ClusterJobAnalytics } from '../../components/v2/ClusterJobAnalytics'
import { ClusterWaitTimeAnalysis } from '../../components/v2/ClusterWaitTimeAnalysis'

export const ClusterOverview = () => {
  const { clusterName } = useParams<{ clusterName: string }>()

  if (!clusterName) {
    return <Text>No cluster selected</Text>
  }

  return (
    <VStack w="100%" align="start" gap={6} p={4}>
      <Text fontSize="2xl" fontWeight="bold">
        {clusterName}
      </Text>

      <ClusterOverviewCards cluster={clusterName} />

      <ClusterHealthStatus cluster={clusterName} />

      <ClusterResourceDistribution cluster={clusterName} />

      <ClusterQueueActivity cluster={clusterName} />
{/*
      <ClusterJobAnalytics cluster={clusterName} />

      <ClusterWaitTimeAnalysis cluster={clusterName} /> */}

      <ClusterTimebasedActivity cluster={clusterName} />
    </VStack>
  )
}
