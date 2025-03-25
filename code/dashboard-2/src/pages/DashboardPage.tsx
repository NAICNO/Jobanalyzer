import { ReactNode, useMemo, useRef, useState } from 'react'
import { Navigate, useParams, Link as ReactRouterLink, useSearchParams } from 'react-router'
import {
  Heading,
  VStack,
  Text,
  Link as ChakraLink,
  useDisclosure,
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
import { NodeSelectionHelpDrawer, NodeSelectionInput, PageTitle } from '../components'
import { DashboardTable } from '../components/table'
import { Cluster, Subcluster } from '../types'

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

  const {data} = useFetchDashboard(selectedCluster, query)

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
      to={`/${clusterName}/subcluster/${subcluster.name}`}
    >
      {subcluster.name}
      {' '}
    </ChakraLink>
  ))

  return (
    <>
      <PageTitle title={`${selectedCluster.name} Dashboard`}/>
      <VStack spacing={1}>
        <Heading size={{base: 'md', md: 'lg'}} mb={4}>{selectedCluster.name}: Jobanalyzer Dashboard</Heading>
        <Text>Click on hostname for machine details.</Text>
        <Text>
          <ChakraLink as={ReactRouterLink} to={jobQueryLink} isExternal mr="10px">
            Job query <ExternalLinkIcon mx="2px"/>
          </ChakraLink>
          {
            subclusterLinks.length > 0 &&
            <>
              Aggregates:{' '}
              {subclusterLinks}
            </>
          }
        </Text>
        <ViolatorsAndZombiesLinks cluster={selectedCluster}/>
        <NodeSelectionInput
          defaultQuery={query}
          onClickSubmit={handleSubmitClick}
          onClickHelp={onOpenHelpSidebar}
          focusRef={focusRef}
        />
        <DashboardTable table={table} cluster={selectedCluster}/>
      </VStack>
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

