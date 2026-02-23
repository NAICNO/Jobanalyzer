import { VStack, HStack, Text } from '@chakra-ui/react'
import { useParams } from 'react-router'

import { ClusterOverviewProvider, useClusterOverviewContext } from '../../contexts/ClusterOverviewContext'
import { TimeRangePicker } from '../../components/TimeRangePicker'
import { ClusterStalenessIndicator } from '../../components/v2/ClusterStalenessIndicator'
import { ClusterOverviewCards } from '../../components/v2/ClusterOverviewCards'
import { ClusterHealthStatus } from '../../components/v2/ClusterHealthStatus'
import { ClusterResourceDistribution } from '../../components/v2/ClusterResourceDistribution'
import { ClusterQueueActivity } from '../../components/v2/ClusterQueueActivity'
import { ClusterTimebasedActivity } from '../../components/v2/ClusterTimebasedActivity'
import { ClusterJobAnalytics } from '../../components/v2/ClusterJobAnalytics'
import { ClusterWaitTimeAnalysis } from '../../components/v2/ClusterWaitTimeAnalysis'

const ClusterOverviewContent = () => {
  const { cluster, timeRange, setTimeRange, refetchAll, isFetching, oldestDataUpdatedAt } = useClusterOverviewContext()

  return (
    <VStack w="100%" align="start" gap={6} p={4}>
      <HStack w="100%" justify="space-between" align="center" flexWrap="wrap" gap={3}>
        <Text fontSize="2xl" fontWeight="bold">
          {cluster}
        </Text>
        <HStack gap={3} align="center">
          <ClusterStalenessIndicator
            oldestDataUpdatedAt={oldestDataUpdatedAt}
            isFetching={isFetching}
            onRefresh={refetchAll}
          />
          <TimeRangePicker value={timeRange} onChange={setTimeRange} />
        </HStack>
      </HStack>

      <ClusterOverviewCards />

      <ClusterHealthStatus cluster={cluster} />

      <ClusterResourceDistribution />

      <ClusterQueueActivity />

      <ClusterJobAnalytics cluster={cluster} />

      <ClusterWaitTimeAnalysis />

      <ClusterTimebasedActivity cluster={cluster} />
    </VStack>
  )
}

export const ClusterOverview = () => {
  const { clusterName } = useParams<{ clusterName: string }>()

  if (!clusterName) {
    return <Text>No cluster selected</Text>
  }

  return (
    <ClusterOverviewProvider cluster={clusterName}>
      <ClusterOverviewContent />
    </ClusterOverviewProvider>
  )
}
