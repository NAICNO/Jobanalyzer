import { useEffect, useState } from 'react'
import { Heading, HStack, VStack, Text, Link as ChakraLink } from '@chakra-ui/react'
import { Link as ReactRouterLink, Navigate, useParams } from 'react-router'

import { findSubcluster } from '../util'
import { useFetchHostDetails } from '../hooks/useFetchHostDetails.ts'
import { MachineDetailsChart } from '../components/chart/MachineDetailsChart.tsx'
import { NavigateBackButton, PageTitle } from '../components'
import { HostDetails } from '../types'

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
  } = useFetchHostDetails(cluster.canonical, hostname, 'weekly', true, false, true)

  useEffect(() => {
    if (data) {
      setHostDetails(data)
    }
  }, [data])

  const jobQueryLink = `/jobQuery?cluster=${clusterName}&host=${subcluster.nodes}`


  return (
    <>
      <PageTitle title={`${cluster.name} (${subcluster.name}) aggregated weekly load`}/>
      <VStack alignItems={'start'}>
        <HStack mb="20px" alignItems={'start'}>
          <NavigateBackButton/>
          <VStack alignItems={'start'} ml={'20px'} spacing={4}>
            <Heading size={{base: 'md', md: 'lg'}}>{`${cluster.name} (${subcluster.name}) aggregated weekly load`}</Heading>
            <Text whiteSpace={'pre-line'}>{hostDetails?.system.description}</Text>
            <ChakraLink as={ReactRouterLink} to={jobQueryLink}>
              Job query for this subcluster
            </ChakraLink>

          </VStack>
        </HStack>

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
          measurement is the sum of the sizes of the jobs&apos; private memories.
        </Text>
      </VStack>
    </>
  )
}
