import {
  Heading,
  Table,
  Text,
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
      <VStack gap={2} alignItems="start">
        <Heading as="h3" ml={2} size={{base: 'md', md: 'lg'}}>
          Job Profile Details
        </Heading>
        <Table.ScrollArea  borderWidth="1px" >
          <Table.Root size={'sm'} showColumnBorder variant="outline">
            <Table.Header>
              <Table.Row>
                <Table.ColumnHeader>Job #</Table.ColumnHeader>
                <Table.ColumnHeader>User</Table.ColumnHeader>
                <Table.ColumnHeader>Cluster</Table.ColumnHeader>
                <Table.ColumnHeader>Node</Table.ColumnHeader>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              <Table.Row>
                <Table.Cell>{jobId}</Table.Cell>
                <Table.Cell>{user}</Table.Cell>
                <Table.Cell>{clusterName}</Table.Cell>
                <Table.Cell>{hostname}</Table.Cell>
              </Table.Row>
            </Table.Body>
          </Table.Root>
        </Table.ScrollArea>
        <Text fontSize="sm" color="gray.500" mt={4} mb={4}>
          Tip: Drag the brush below to zoom in and explore specific time ranges. Timestamps are in UTC.
        </Text>
        <VStack w={'100%'} gap={8}>
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
