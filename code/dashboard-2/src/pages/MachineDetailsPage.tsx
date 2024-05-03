import { Heading, HStack, Select, SlideFade, Text, VStack } from '@chakra-ui/react'
import { Navigate, useParams } from 'react-router-dom'
import { isValidateClusterName } from '../util'
import { CLUSTER_INFO, FETCH_FREQUENCIES } from '../Constants.ts'
import { useFetchHostnames } from '../hooks/useFetchHosts.ts'
import { useMemo, useState } from 'react'
import { useFetchHostDetails } from '../hooks/useFetchHostDetails.ts'
import { NavigateBackButton } from '../components/NavigateBackButton.tsx'
import { useFetchViolations } from '../hooks/useFetchViolations.ts'
import {
  getDeadWeightTableColumns,
  getViolatingJobTableColumns,
  getViolatingUserTableColumns
} from '../util/TableUtils.ts'
import { getCoreRowModel, getSortedRowModel, SortingState, useReactTable } from '@tanstack/react-table'
import ViolatingUserTable from '../components/table/ViolatingUserTable.tsx'
import ViolatingJobTable from '../components/table/ViolatingJobTable.tsx'
import DeadWeightTable from '../components/table/DeadWeightTable.tsx'
import { useFetchDeadWeight } from '../hooks/useFetchDeadWeight.ts'

const emptyArray: any[] = []

export default function MachineDetailsPage() {

  const {clusterName, hostname} = useParams<string>()

  if (!isValidateClusterName(clusterName) || !hostname) {
    return (
      <Navigate to="/"/>
    )
  }

  const selectedCluster = CLUSTER_INFO[clusterName!]

  const [selectedFrequency, setSelectedFrequency] = useState(FETCH_FREQUENCIES[0])

  const {data: hostnames} = useFetchHostnames(selectedCluster.cluster)

  const hasDowntime = selectedCluster.hasDowntime

  const isValidHostname = hostnames?.includes(hostname!)

  const {
    data: hostDetails,
    isFetching: isFetchingHostDetails
  } = useFetchHostDetails(hostname!, selectedFrequency.value, isValidHostname)

  const now = new Date()
  const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)

  const filter = {
    afterDate: thirtyDaysAgo,
    hostname: hostname,
  }

  const {
    data: violations,
    isFetching: isFetchingViolations,
    isFetched: isFetchedViolations,
  } = useFetchViolations(clusterName!, filter, isValidHostname)

  const violatingUserTableColumns = useMemo(() => getViolatingUserTableColumns(), [selectedCluster])
  const [violatingUserTableSorting, setViolatingUserTableSorting] = useState<SortingState>([])

  const violatingUserTable = useReactTable({
    columns: violatingUserTableColumns,
    data: violations?.byUser || emptyArray,
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
    data: violations?.byJob || emptyArray,
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
    data: deadweights || emptyArray,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting: sorting,
    }
  })

  return (
    <VStack spacing={4} alignItems="start">
      <HStack mb="20px">
        <NavigateBackButton/>
        <Heading as="h2" size="xl">Machine Details</Heading>
      </HStack>
      <Text>
        Cluster: {hostDetails?.system.hostname}
      </Text>
      <Text>
        Description: {hostDetails?.system.description}
      </Text>

      <Heading as="h3" size="lg">Machine Load</Heading>
      <Select
        value={selectedFrequency.value}
        w="25%"
        size="sm"
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
      <Text>
        Graph will go here
      </Text>
      <Text>Data are relative to all system resources (e.g., 100% CPU
        means all cores are completely busy; 100% GPU means all cards are completely busy).
      </Text>
      <Text>Main memory (RAM) can go over 100% due to paging and similar system effects; the
        measurement is the sum of the sizes of the jobs' private memories.
      </Text>

      <Heading as="h4" size="lg"  mt='20px'>
        Violators last 30 days
      </Heading>

      <Text>
        The following jobs have violated usage policy and are probably not appropriate to run on this cluster. The list
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

      <Heading as="h4" size="lg" mt='20px'>
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
    </VStack>
  )
}
