import { useState } from 'react'
import {
  Box, Button, ButtonGroup,
  Checkbox,
  Divider,
  Heading,
  HStack,
  Select,
  Spacer,
  Text,
  VStack
} from '@chakra-ui/react'
import { Link as ReactRouterLink, Navigate, useParams } from 'react-router-dom'

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

  return (
    <>
      <PageTitle title={`${hostDetails?.system.hostname} Details`}/>
      <VStack spacing={4} alignItems="start">
        <HStack mb="15px">
          <NavigateBackButton/>
          <Heading as="h3" ml={2} size={{base: 'md', md: 'lg'}}>
            Machine Details: {hostDetails?.system.hostname}
          </Heading>
        </HStack>
        <Text> Description :{'\t'}{hostDetails?.system.description}
        </Text>

        <ButtonGroup size={'sm'} spacing="2" variant={'outline'}>
          <Button as={ReactRouterLink} to={jobQueryLink} leftIcon={<IoSearchOutline/>}>
            Job Query
          </Button>
          {selectedCluster.violators &&
            <Button as={ReactRouterLink} to={violatorsLink} leftIcon={<PiGavel/>}>
              View Violators
            </Button>
          }
          {selectedCluster.deadweight &&
            <Button as={ReactRouterLink} to={deadWeightLink} leftIcon={<GiShamblingZombie/>}>
              View Zombies
            </Button>
          }
        </ButtonGroup>
        <Divider/>

        <Heading as="h4" size="lg" my="0px">Machine Load</Heading>

        <HStack w="100%">
          <Box w="50%">
            <Select
              value={selectedFrequency.value}
              size="sm"
              maxW="50%"
              onChange={(event) => {
                const value = event.target.value
                setSelectedFrequency(FETCH_FREQUENCIES.find((frequency) => frequency.value === value)!)
              }}>
              {
                FETCH_FREQUENCIES?.map((frequency) => (
                  <option key={frequency.value} value={frequency.value}>
                    {frequency.text}
                  </option>
                ))
              }
            </Select>
          </Box>
          <Spacer/>
          <Checkbox
            isChecked={isShowData}
            onChange={(event) => setIsShowData(event.target.checked)}>
            Show data
          </Checkbox>
          <Checkbox
            isChecked={isShowDataPoints}
            onChange={(event) => setIsShowDataPoints(event.target.checked)}>
            Show data points
          </Checkbox>
          {hasDowntime &&
            <Checkbox
              isChecked={isShowDowntime}
              onChange={(event) => setIsShowDowntime(event.target.checked)}>
              Show downtime
            </Checkbox>
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
