import { ReactNode, useMemo, useRef, useState } from 'react'
import { Navigate, useParams, Link as ReactRouterLink, useSearchParams } from 'react-router-dom'
import {
  Container,
  Heading,
  VStack,
  Text,
  Link as ChakraLink,
  useDisclosure,
  SlideFade,
} from '@chakra-ui/react'
import { ExternalLinkIcon } from '@chakra-ui/icons'
import {
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'

import { useFetchDashboard } from '../hooks/useFetchDashboard.ts'
import { EMPTY_ARRAY, } from '../Constants.ts'
import { findCluster } from '../util'
import { getDashboardTableColumns } from '../util/TableUtils.ts'
import DashboardTable from '../components/table/DasboardTable.tsx'
import NodeSelectionHelpDrawer from '../components/NodeSelectionHelpDrawer.tsx'
import NodeSelectionInput from '../components/NodeSelectionInput.tsx'
import PageTitle from '../components/PageTitle.tsx'
import { Cluster } from '../types/Cluster.ts'

export default function DashboardPage() {

  const {clusterName} = useParams<string>()

  const selectedCluster = findCluster(clusterName!)

  if (!selectedCluster) {
    return (
      <Navigate to="/"/>
    )
  }

  const [searchParams, setSearchParams] = useSearchParams()
  const defaultQuery = selectedCluster.defaultQuery

  const query = searchParams.get('query') || defaultQuery

  const {data, isFetched} = useFetchDashboard(selectedCluster, query)

  const [sorting, setSorting] = useState<SortingState>([])

  const tableColumns = useMemo(() => getDashboardTableColumns(selectedCluster), [clusterName])

  const table = useReactTable({
    columns: tableColumns,
    data: data || EMPTY_ARRAY,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting,
    }
  })

  const {isOpen: isOpenHelpSidebar, onOpen: onOpenHelpSidebar, onClose} = useDisclosure()
  const focusRef = useRef<HTMLInputElement | null>(null)

  const jobQueryLink = `/jobquery?cluster=${clusterName}`

  const handleSubmitClick = (query: string) => {
    setSearchParams({query})
  }

  const subclusterLinks = selectedCluster.subclusters.map((subcluster: Subcluster) => (
    <ChakraLink
      key={subcluster.name}
      as={ReactRouterLink}
      color="teal.500"
      to={`/${clusterName}/subcluster/${subcluster.name}`}
    >
      {subcluster.name}
      {' '}
    </ChakraLink>
  ))

  return (
    <>
      <PageTitle title={`${selectedCluster.name} Dashboard`}/>
      <Container centerContent>
        <VStack spacing={1}>
          <Heading size="lg" mb={4}>{selectedCluster.name}: Jobanalyzer Dashboard</Heading>
          <Text>Click on hostname for machine details.</Text>
          <Text>
            <ChakraLink as={ReactRouterLink} to={jobQueryLink} isExternal mr="10px" color="teal.500">
              Job query <ExternalLinkIcon mx="2px"/>
            </ChakraLink>
            Aggregates:{' '}
            {subclusterLinks}
          </Text>
          <Text>Recent: 30 mins Longer: 12 hrs{' '} </Text>
          <ViolatorsAndZombiesLinks cluster={selectedCluster}/>
          <NodeSelectionInput
            defaultQuery={query}
            onClickSubmit={handleSubmitClick}
            onClickHelp={onOpenHelpSidebar}
            focusRef={focusRef}
          />
          <SlideFade in={isFetched}>
            <DashboardTable table={table} cluster={selectedCluster}/>
          </SlideFade>
        </VStack>
      </Container>
      <NodeSelectionHelpDrawer isOpen={isOpenHelpSidebar} onClose={onClose} finalFocusRef={focusRef}/>
    </>
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

  const {violators, deadweight, cluster: clusterName} = cluster

  const violatorsLink = `/${clusterName}/violators`
  const deadweightLink = `/${clusterName}/deadweight`

  return (
    <Text>
      <ConditionalLink isActive={violators} to={violatorsLink}>
        Violators
      </ConditionalLink>
      {violators && deadweight && ' and '}
      <ConditionalLink isActive={deadweight} to={deadweightLink}>
        Zombies
      </ConditionalLink>
    </Text>
  )
}

