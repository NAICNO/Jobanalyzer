import { useState } from 'react'
import {
  Box,
  Button,
  Heading,
  HStack,
  SegmentGroup,
  SimpleGrid,
  Spacer,
  Tag,
  Text,
  VStack,
} from '@chakra-ui/react'
import { useSearchParams } from 'react-router'
import { IoGridOutline, IoListOutline } from 'react-icons/io5'

import { PageTitle } from '../components'
import { useFetchJobProfile } from '../hooks/useFetchJobProfile.ts'
import { JobProfileChart } from '../components/chart/JobProfileChart.tsx'
import { useColorModeValue } from '../components/ui/color-mode.tsx'
import { JobBasicInfoTable } from '../components/table/JobBasicInfoTable.tsx'

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

  const query = new URLSearchParams(
    Object.entries(
      {
        clusterName: clusterName,
        hostname: hostname,
        user: user,
        jobId: jobId,
      }
    ).reduce<Record<string, string>>((acc, [key, val]) => {
      acc[key] = String(val)
      return acc
    }, {})
  ).toString()

  const treeUrl = `/jobs/tree?${query}`

  return (
    <>
      <PageTitle title={`Job Profile - ${clusterName} ${jobId}`}/>
      <VStack gap={2} alignItems="start">
        <Heading as="h3" ml={2} size={{base: 'md', md: 'lg'}}>
          Job Profile Details
        </Heading>
        <HStack alignItems="start">
          <JobBasicInfoTable jobId={jobId} user={user} clusterName={clusterName} hostname={hostname}/>
          <Box position="relative" display="inline-block" pl={4}>
            <Button asChild colorPalette={'blue'}>
              <a href={treeUrl} target="_blank" rel="noopener noreferrer">View Process Tree</a>
            </Button>
            <Tag.Root
              position="absolute"
              top="-1"
              right="-1"
              fontSize="xs"
              color="white"
              colorPalette={'orange'}
              variant="solid"
              px={2}
              py={0.5}
              borderRadius="md"
              transform="translate(25%, -50%)"
            >
              <Tag.Label>
                Experimental
              </Tag.Label>
            </Tag.Root>
          </Box>
        </HStack>

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
