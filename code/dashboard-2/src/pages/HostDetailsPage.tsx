import { useMemo, useState } from 'react'
import {
  Box,
  Checkbox,
  Divider,
  Heading,
  HStack,
  Select,
  SlideFade,
  Spacer,
  Text,
  VStack
} from '@chakra-ui/react'
import { Navigate, useParams } from 'react-router-dom'
import { getCoreRowModel, getSortedRowModel, SortingState, useReactTable } from '@tanstack/react-table'

import { EMPTY_ARRAY, FETCH_FREQUENCIES } from '../Constants.ts'
import { useFetchHostnames } from '../hooks/useFetchHosts.ts'
import { useFetchHostDetails } from '../hooks/useFetchHostDetails.ts'
import { NavigateBackButton } from '../components/NavigateBackButton.tsx'
import { useFetchViolations } from '../hooks/useFetchViolations.ts'
import { findCluster } from '../util'
import {
  getDeadWeightTableColumns,
  getViolatingJobTableColumns,
  getViolatingUserTableColumns
} from '../util/TableUtils.ts'
import ViolatingUserTable from '../components/table/ViolatingUserTable.tsx'
import ViolatingJobTable from '../components/table/ViolatingJobTable.tsx'
import DeadWeightTable from '../components/table/DeadWeightTable.tsx'
import { useFetchDeadWeight } from '../hooks/useFetchDeadWeight.ts'
import MachineDetailsChart from '../components/chart/MachineDetailsChart.tsx'

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

  const {data: hostnames} = useFetchHostnames(selectedCluster.cluster)

  const hasDowntime = selectedCluster.hasDowntime

  const isValidHostname = hostnames?.includes(hostname!)

  const {
    data: hostDetails
  } = useFetchHostDetails(hostname!, selectedFrequency.value, isShowData, isShowDowntime, isValidHostname)

  const now = new Date()
  const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)

  const filter = {
    afterDate: thirtyDaysAgo,
    hostname: hostname,
  }

  const {
    data: violations,
    isFetched: isFetchedViolations,
  } = useFetchViolations(clusterName!, filter, isValidHostname)

  const violatingUserTableColumns = useMemo(() => getViolatingUserTableColumns(), [selectedCluster])
  const [violatingUserTableSorting, setViolatingUserTableSorting] = useState<SortingState>([])

  const violatingUserTable = useReactTable({
    columns: violatingUserTableColumns,
    data: violations?.byUser || EMPTY_ARRAY,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setViolatingUserTableSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting: violatingUserTableSorting,
    }
  })

  const violatingJobTableColumns = useMemo(() => getViolatingJobTableColumns(), [selectedCluster])
  const [violatingJobTableSorting, setViolatingJobTableSorting] = useState<SortingState>([])

  const violatingJobTable = useReactTable({
    columns: violatingJobTableColumns,
    data: violations?.byJob || EMPTY_ARRAY,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setViolatingJobTableSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting: violatingJobTableSorting,
    }
  })

  const {data: deadweights, isFetched: isFetchedDeadweights} = useFetchDeadWeight(clusterName!, filter, isValidHostname)

  const deadWeightJobTableColumns = useMemo(() => getDeadWeightTableColumns(), [clusterName])
  const [sorting, setSorting] = useState<SortingState>([])

  const deadWeightTable = useReactTable({
    columns: deadWeightJobTableColumns,
    data: deadweights || EMPTY_ARRAY,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting: sorting,
    }
  })

  return (
    <VStack spacing={4} alignItems="start">
      <HStack mb="15px">
        <NavigateBackButton/>
        <Heading as="h3" size="lg">
          Machine Details: {hostDetails?.system.hostname}
        </Heading>
      </HStack>
      <Text> Description :{'\t'}{hostDetails?.system.description}
      </Text>

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
        measurement is the sum of the sizes of the jobs' private memories.
      </Text>

      {selectedCluster.violators &&
        <>
          <Heading as="h4" size="lg" mt="20px">
            Violators last 30 days
          </Heading>

          <Text>
            The following jobs have violated usage policy and are probably not appropriate to run on this cluster. The
            list
            is recomputed at noon and midnight and goes back four weeks.
          </Text>

          <Heading as="h4" size="md">By user</Heading>

          <SlideFade in={isFetchedViolations}>
            <ViolatingUserTable table={violatingUserTable}/>
          </SlideFade>
          <Heading as="h4" size="md" mt="20px">
            By job and time
          </Heading>
          <SlideFade in={isFetchedViolations}>
            <ViolatingJobTable table={violatingJobTable}/>
          </SlideFade>
        </>
      }

      {selectedCluster.deadweight &&
        <>
          <Heading as="h4" size="lg" mt="20px">
            Deadweight processes last 30 days
          </Heading>
          <Text>
            The following processes and jobs are zombies or defuncts or
            otherwise dead and may be bogging down the system. The list is
            recomputed at noon and midnight and goes back four weeks.
          </Text>

          <SlideFade in={isFetchedDeadweights}>
            <DeadWeightTable table={deadWeightTable}/>
          </SlideFade>
        </>
      }
    </VStack>
  )
}
