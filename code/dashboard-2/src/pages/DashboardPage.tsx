import { ReactNode, useMemo, useState } from 'react'
import { Navigate, useParams, Link as ReactRouterLink } from 'react-router-dom'
import {
  Container,
  Heading,
  VStack,
  Text,
  Link as ChakraLink
} from '@chakra-ui/react'
import { ExternalLinkIcon } from '@chakra-ui/icons'
import {
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'

import { useFetchDashboard } from '../hooks/useFetchDashboard.ts'
import { CLUSTER_INFO, } from '../Constants.ts'
import { isValidateClusterName } from '../util'
import { getDashboardTableColumns } from '../util/TableUtils.ts'
import DashboardTable from '../components/table/DasboardTable.tsx'

const emptyArray: any[] = []

export default function DashboardPage() {

  const {clusterName} = useParams<string>()

  if (!isValidateClusterName(clusterName)) {
    return (
      <Navigate to="/"/>
    )
  }

  const {data} = useFetchDashboard(clusterName!)

  const [sorting, setSorting] = useState<SortingState>([])

  const selectedCluster = CLUSTER_INFO[clusterName!]
  const tableColumns = useMemo(() => getDashboardTableColumns(selectedCluster), [clusterName])

  const table = useReactTable({
    columns: tableColumns,
    data: data || emptyArray,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting,
    }
  })

  return (
    <Container maxW="xl" height="100vh" centerContent>
      <VStack spacing={2} padding="4">
        <Heading size="lg" mb={4}>{selectedCluster.name}: Jobanalyzer Dashboard</Heading>
        <Text>Click on hostname for machine details.</Text>
        <ChakraLink as={ReactRouterLink} to="/jobquery" isExternal>
          Job query <ExternalLinkIcon mx="2px"/>
        </ChakraLink>
        <Text>
          Aggregates:{' '}
          <ChakraLink as={ReactRouterLink} color="teal.500" href="#">
            nvidia
          </ChakraLink>
        </Text>
        <Text>Recent: 30 mins Longer: 12 hrs{' '} </Text>
        <ViolatorsAndZombiesLinks cluster={selectedCluster}/>
        <DashboardTable table={table} cluster={selectedCluster}/>
      </VStack>
    </Container>
  )
}

const ViolatorsAndZombiesLinks = ({cluster}: { cluster: Cluster }) => {
  const ConditionalLink = ({
    isActive,
    to,
    children,
  }: {
    isActive: boolean,
    to: string,
    children: ReactNode,
  }) => {
    if (!isActive) {
      return null
    }

    return (
      <ChakraLink
        as={ReactRouterLink}
        color="teal.500"
        to={to}
      >
        {children}
      </ChakraLink>
    )
  }

  const {violators, deadweight} = cluster

  return (
    <Text>
      <ConditionalLink isActive={violators} to="violators">
        Violators
      </ConditionalLink>
      {violators && deadweight && ' and '}
      <ConditionalLink isActive={deadweight} to="deadweight">
        Zombies
      </ConditionalLink>
    </Text>
  )
}

