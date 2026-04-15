import { VStack, HStack, Text } from '@chakra-ui/react'
import { useParams } from 'react-router'

import { ClusterOverviewProvider, useClusterOverviewContext } from '../contexts/ClusterOverviewContext'
import { TimeRangePicker } from '../components/TimeRangePicker'
import { ClusterStalenessIndicator } from '../components/ClusterStalenessIndicator'
import { getClusterConfig } from '../config/clusters'
import { ClusterOverviewCards } from '../components/ClusterOverviewCards'
import { ClusterHealthStatus } from '../components/ClusterHealthStatus'
import { ClusterResourceDistribution } from '../components/ClusterResourceDistribution'
import { ClusterQueueActivity } from '../components/ClusterQueueActivity'
import { ClusterTimebasedActivity } from '../components/ClusterTimebasedActivity'
import { ClusterJobAnalytics } from '../components/ClusterJobAnalytics'
import { ClusterWaitTimeAnalysis } from '../components/ClusterWaitTimeAnalysis'
import { ClusterDiskStats } from '../components/ClusterDiskStats'
import { LazySection } from '../components/LazySection'

const ClusterOverviewContent = () => {
  const { cluster, timeRange, setTimeRange, refetchAll, isFetching, oldestDataUpdatedAt } = useClusterOverviewContext()
  const config = getClusterConfig(cluster)

  return (
    <VStack w="100%" align="start" gap={6} p={4}>
      <HStack w="100%" justify="space-between" align="center" flexWrap="wrap" gap={3}>
        <VStack align="start" gap={0}>
          <Text fontSize="2xl" fontWeight="bold">
            {config?.name ?? cluster}
          </Text>
          <Text fontSize="xs" color="gray.500">
            {cluster}
          </Text>
        </VStack>
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

      <LazySection minHeight="250px">
        {(isVisible) => <ClusterHealthStatus cluster={cluster} enabled={isVisible} />}
      </LazySection>

      <ClusterQueueActivity />

      <ClusterWaitTimeAnalysis />

      <LazySection minHeight="500px">
        {(isVisible) => <ClusterJobAnalytics cluster={cluster} enabled={isVisible} />}
      </LazySection>

      <ClusterResourceDistribution />

      <LazySection minHeight="700px">
        {(isVisible) => <ClusterTimebasedActivity cluster={cluster} enabled={isVisible} />}
      </LazySection>

      <LazySection minHeight="800px">
        {(isVisible) => <ClusterDiskStats cluster={cluster} enabled={isVisible} />}
      </LazySection>
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
