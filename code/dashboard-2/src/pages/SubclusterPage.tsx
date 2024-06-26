import { Heading, HStack, VStack, Text, Box, Link as ChakraLink } from '@chakra-ui/react'
import { Link as ReactRouterLink, Navigate, useParams } from 'react-router-dom'

import { findSubcluster } from '../util'
import { useFetchHostDetails } from '../hooks/useFetchHostDetails.ts'
import MachineDetailsChart from '../components/chart/MachineDetailsChart.tsx'
import { NavigateBackButton } from '../components/NavigateBackButton.tsx'
import { useEffect, useState } from 'react'

export default function SubclusterPage() {

  const {clusterName, subclusterName} = useParams<string>()

  const result = findSubcluster(clusterName, subclusterName)
  if (!result) {
    return (
      <Navigate to="/"/>
    )
  }

  const {cluster, subcluster} = result

  const hostname = `${clusterName}-${subclusterName}`

  const [hostDetails, setHostDetails] = useState<HostDetails>()

  const {
    data
  } = useFetchHostDetails(hostname, 'weekly', true, false, true)

  useEffect(() => {
    if (data) {
      setHostDetails(data)
    }
  }, [data])

  const jobQueryLink = `/jobQuery?cluster=${clusterName}&host=${subcluster.nodes}`


  return (
    <VStack alignItems={'start'}>
      <HStack mb="20px">
        <NavigateBackButton/>
        <Heading ml="20px">{`${cluster.name} (${subcluster.name}) aggregated weekly load`}</Heading>
      </HStack>
      <Box>
        <Text>{hostDetails?.system.description}</Text>
      </Box>

      <ChakraLink as={ReactRouterLink} color="teal.500" to={jobQueryLink}>
        Job query for this subcluster
      </ChakraLink>

      <MachineDetailsChart
        dataItems={hostDetails?.chart?.dataItems || []}
        seriesConfigs={hostDetails?.chart?.seriesConfigs || []}
        containerProps={{
          width: '100%',
          height: 600,
        }}
        isShowDataPoints={false}
      />

      <Text>Data are relative to all system resources (e.g., 100% CPU
        means all cores are completely busy; 100% GPU means all cards are completely busy).
      </Text>

      <Text>Main memory (RAM) can go over 100% due to paging and similar system effects; the
        measurement is the sum of the sizes of the jobs' private memories.
      </Text>
    </VStack>
  )
}
