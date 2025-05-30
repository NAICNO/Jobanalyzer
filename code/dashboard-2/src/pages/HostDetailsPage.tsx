import { useState } from 'react'
import {
  Box,
  Button,
  Checkbox,
  Group,
  HStack,
  Heading,
  Portal,
  Select,
  Separator,
  Spacer,
  Text,
  VStack,
  createListCollection,
} from '@chakra-ui/react'
import { Link as ReactRouterLink, Navigate, useParams } from 'react-router'

import { FETCH_FREQUENCIES } from '../Constants.ts'
import { useFetchHostnames } from '../hooks/useFetchHosts.ts'
import { useFetchHostDetails } from '../hooks/useFetchHostDetails.ts'
import { findCluster } from '../util'
import { NavigateBackButton, PageTitle } from '../components'
import { MachineDetailsChart } from '../components/chart/MachineDetailsChart'
import { IoSearchOutline } from 'react-icons/io5'
import { GiShamblingZombie } from 'react-icons/gi'
import { PiGavel } from 'react-icons/pi'


export default function HostDetailsPage() {

  const {clusterName, hostname} = useParams<string>()

  const selectedCluster = findCluster(clusterName)
  if (!selectedCluster || !hostname) {
    return (
      <Navigate to="/"/>
    )
  }

  const [selectedFrequency, setSelectedFrequency] = useState(FETCH_FREQUENCIES[0])
  const [isShowData, setIsShowData] = useState<boolean>(true)
  const [isShowDowntime, setIsShowDowntime] = useState<boolean>(false)
  const [isShowDataPoints, setIsShowDataPoints] = useState<boolean>(false)

  const {data: hostnames} = useFetchHostnames(selectedCluster)

  const hasDowntime = selectedCluster.hasDowntime

  const isValidHostname = hostnames?.includes(hostname!)

  const {
    data: hostDetails
  } = useFetchHostDetails(selectedCluster.canonical, hostname!, selectedFrequency.value, isShowData, isShowDowntime, isValidHostname)

  const jobQueryLink = `/jobquery?cluster=${clusterName}&host=${hostname}`
  const violatorsLink = `/${clusterName}/${hostname}/violators`
  const deadWeightLink = `/${clusterName}/${hostname}/deadweight`

  const collection = createListCollection({
    items: FETCH_FREQUENCIES,
  })

  return (
    <>
      <PageTitle title={`${hostDetails?.system.hostname} Details`}/>
      <VStack gap={4} alignItems="start">
        <HStack mb={1}>
          <NavigateBackButton/>
          <Heading as="h3" ml={2} size={{base: 'lg', md: 'xl'}}>
            Machine Details: {hostDetails?.system.hostname}
          </Heading>
        </HStack>
        <Text>
          Description :{'\t'}{hostDetails?.system.description}
        </Text>

        <Group gap="2">
          <Button asChild variant="surface" size="xs" colorPalette="blue">
            <ReactRouterLink to={jobQueryLink}>
              <IoSearchOutline/> Job Query
            </ReactRouterLink>
          </Button>
          {selectedCluster.violators &&
            <Button asChild variant="surface" size="xs" colorPalette="blue">
              <ReactRouterLink to={violatorsLink}>
                <PiGavel/> View Violators
              </ReactRouterLink>
            </Button>
          }
          {selectedCluster.deadweight &&
            <Button asChild variant="surface" size="xs" colorPalette="blue">
              <ReactRouterLink to={deadWeightLink}>
                <GiShamblingZombie/> View Zombies
              </ReactRouterLink>
            </Button>
          }
        </Group>

        <Separator size="lg"/>

        <Heading as="h4" size="lg" my="0px">Machine Load</Heading>

        <HStack w="100%">
          <Box w="50%">
            <Select.Root
              collection={collection}
              value={[selectedFrequency.value]}
              size="sm"
              maxW="50%"
              onValueChange={(event) => {
                const value = event.value
                setSelectedFrequency(FETCH_FREQUENCIES.find((frequency) => frequency.value === value[0])!)
              }}>
              <Select.HiddenSelect/>
              <Select.Control>
                <Select.Trigger>
                  <Select.ValueText/>
                </Select.Trigger>
                <Select.IndicatorGroup>
                  <Select.Indicator/>
                </Select.IndicatorGroup>
              </Select.Control>
              <Portal>
                <Select.Positioner>
                  <Select.Content>
                    {
                      collection.items.map((frequency) => (
                        <Select.Item key={frequency.value} item={frequency}>
                          <Select.ItemText>
                            {frequency.text}
                          </Select.ItemText>
                        </Select.Item>
                      ))
                    }
                  </Select.Content>
                </Select.Positioner>
              </Portal>
            </Select.Root>
          </Box>
          <Spacer/>
          <Checkbox.Root
            colorPalette={'blue'}
            variant={'subtle'}
            checked={isShowData}
            onCheckedChange={(event) => setIsShowData(!!event.checked)}>
            <Checkbox.HiddenInput/>
            <Checkbox.Control/>
            <Checkbox.Label>
              Show data
            </Checkbox.Label>
          </Checkbox.Root>
          <Checkbox.Root
            colorPalette={'blue'}
            variant={'subtle'}
            checked={isShowDataPoints}
            onCheckedChange={(event) => setIsShowDataPoints(!!event.checked)}>
            <Checkbox.HiddenInput/>
            <Checkbox.Control/>
            <Checkbox.Label>
              Show data points
            </Checkbox.Label>
          </Checkbox.Root>
          {hasDowntime &&
            <Checkbox.Root
              colorPalette={'blue'}
              variant={'subtle'}
              checked={isShowDowntime}
              onCheckedChange={(event) => setIsShowDowntime(!!event.checked)}>
              <Checkbox.HiddenInput/>
              <Checkbox.Control/>
              <Checkbox.Label>
                Show downtime
              </Checkbox.Label>
            </Checkbox.Root>
          }
        </HStack>
        <MachineDetailsChart
          dataItems={hostDetails?.chart?.dataItems || []}
          seriesConfigs={hostDetails?.chart?.seriesConfigs || []}
          containerProps={{
            width: '100%',
            height: 600,
          }}
          yAxisDomain={([dataMin, dataMax]) => {
            const min = dataMin
            const max = Math.floor((dataMax + 10) / 10) * 10
            return [min, max]
          }}
          isShowDataPoints={isShowDataPoints}
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
