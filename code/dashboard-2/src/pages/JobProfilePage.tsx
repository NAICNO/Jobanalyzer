import { useState } from 'react'
import {
  Heading,
  HStack,
  SegmentGroup,
  SimpleGrid,
  Spacer,
  Table,
  Text,
  VStack,
} from '@chakra-ui/react'
import { useSearchParams } from 'react-router'
import { IoGridOutline, IoListOutline } from 'react-icons/io5'

import { PageTitle } from '../components'
import { useFetchJobProfile } from '../hooks/useFetchJobProfile.ts'
import { JobProfileChart } from '../components/chart/JobProfileChart.tsx'
import { useColorModeValue } from '../components/ui/color-mode.tsx'

export default function JobProfilePage() {

  const [searchParams] = useSearchParams()

  const clusterName = searchParams.get('clusterName')
  const hostname = searchParams.get('hostname')
  const user = searchParams.get('user')
  const jobId = searchParams.get('jobId')
  const from = searchParams.get('from') || ''
  const to = searchParams.get('to') || ''

  const {data} = useFetchJobProfile(clusterName, hostname, jobId, from, to)

  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')

  const chartSize = viewMode === 'grid' ? {width: '100%', height: 400} : {width: '100%', height: 800}

  const numberOfColumns = viewMode === 'grid' ? 2 : 1

  // Get theme-aware border color
  const borderColor = useColorModeValue('gray.200', 'gray.600')

  return (
    <>
      <PageTitle title={'Job Profile Details'}/>
      <VStack gap={2} alignItems="start">
        <Heading as="h3" ml={2} size={{base: 'md', md: 'lg'}}>
          Job Profile Details
        </Heading>
        <Table.ScrollArea borderWidth="1px">
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
        <HStack alignItems="start" justifyContent={'space-between'} width={'100%'}>
          <Text fontSize="sm" color="gray.500" mt={4} mb={4}>
            Tip: Drag the brush below to zoom in and explore specific time ranges. Timestamps are in UTC.
          </Text>
          <Spacer/>
          <HStack hideBelow="md">
            <Text fontSize="md" mt={4} mb={4}>
              View as:
            </Text>
            <SegmentGroup.Root defaultValue="grid" size="md" value={viewMode} onValueChange={e => setViewMode(e.value)}>
              <SegmentGroup.Indicator/>
              <SegmentGroup.Items
                items={[
                  {
                    value: 'grid',
                    label: (
                      <HStack>
                        <IoGridOutline/>
                        Grid
                      </HStack>
                    ),
                  },
                  {
                    value: 'list',
                    label: (
                      <HStack>
                        <IoListOutline/>
                        List
                      </HStack>
                    ),
                  },
                ]}
              />
            </SegmentGroup.Root>
          </HStack>
        </HStack>
        <SimpleGrid columns={{base: 1, md: numberOfColumns}} rowGap={4} columnGap={4} width="100%" pt={4}>
          {
            data?.map((item, index) => {
                return (
                  <JobProfileChart
                    key={index}
                    profileInfo={item.profileInfo}
                    dataItems={item.dataItems}
                    seriesConfigs={item.seriesConfigs}
                    chartSize={chartSize}
                    maxTooltipItems={viewMode === 'grid' ? 5 : undefined}
                    maxLegendItems={viewMode === 'grid' ? 5 : undefined}
                    syncId={'1'}
                    wrapperStyles={{
                      borderWidth: '1px',
                      borderColor: borderColor,
                      borderRadius: '12px',
                      padding: '24px',
                    }}
                  />
                )
              }
            )
          }
        </SimpleGrid>
      </VStack>
    </>
  )
}
