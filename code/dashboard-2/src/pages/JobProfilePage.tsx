import {
  Heading,
  Table,
  TableContainer,
  Tbody,
  Td, Text,
  Th,
  Thead,
  Tr,
  VStack
} from '@chakra-ui/react'
import { useSearchParams } from 'react-router'

import { PageTitle } from '../components'
import { useFetchJobProfile } from '../hooks/useFetchJobProfile.ts'
import { JobProfileChart } from '../components/chart/JobProfileChart.tsx'

export default function JobProfilePage() {

  const [searchParams] = useSearchParams()

  const clusterName = searchParams.get('clusterName')
  const hostname = searchParams.get('hostname')
  const user = searchParams.get('user')
  const jobId = searchParams.get('jobId')
  const from = searchParams.get('from') || ''
  const to = searchParams.get('to') || ''

  const {data} = useFetchJobProfile(clusterName, hostname, jobId, from, to)

  return (
    <>
      <PageTitle title={'Job Profile Details'}/>
      <VStack spacing={8} alignItems="start">
        <Heading as="h3" ml={2} size={{base: 'md', md: 'lg'}}>
          Job Profile Details
        </Heading>
        <TableContainer>
          <Table size={'sm'}>
            <Thead>
              <Tr>
                <Th>Job #</Th>
                <Th>User</Th>
                <Th>Cluster</Th>
                <Th>Node</Th>
              </Tr>
            </Thead>
            <Tbody>
              <Tr>
                <Td>{jobId}</Td>
                <Td>{user}</Td>
                <Td>{clusterName}</Td>
                <Td>{hostname}</Td>
              </Tr>
            </Tbody>
          </Table>
        </TableContainer>
        <Text fontSize="sm" color="gray.500">
          Tip: Drag the brush below to zoom in and explore specific time ranges. Timestamps are in UTC.
        </Text>
        <VStack w={'100%'} spacing={8}>
          {
            data?.map((item, index) => {
                return (
                  <JobProfileChart
                    key={index}
                    profileInfo={item.profileInfo}
                    dataItems={item.dataItems}
                    seriesConfigs={item.seriesConfigs}
                    containerProps={{width: '100%', height: 600}}
                    syncId={'1'}
                  />
                )
              }
            )
          }
        </VStack>
      </VStack>
    </>
  )
}
