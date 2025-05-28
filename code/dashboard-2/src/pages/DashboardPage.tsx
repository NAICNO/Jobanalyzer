import {
  ReactNode,
  useMemo,
  useRef,
  useState
} from 'react'
import {
  Link as ReactRouterLink,
  Navigate,
  useParams,
  useSearchParams
} from 'react-router'
import {
  HStack,
  Heading,
  Link as ChakraLink,
  Text,
  VStack,
  useDisclosure,
} from '@chakra-ui/react'
import {
  SortingState,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table'

import { useFetchDashboard } from '../hooks/useFetchDashboard.ts'
import { EMPTY_ARRAY, } from '../Constants.ts'
import { findCluster } from '../util'
import { getDashboardTableColumns } from '../util/TableUtils.ts'
import { NodeSelectionHelpDrawer, NodeSelectionInput, PageTitle } from '../components'
import { DashboardTable } from '../components/table'
import { Cluster, Subcluster } from '../types'
import { LuExternalLink } from 'react-icons/lu'

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

  const {open: isOpenHelpSidebar, onOpen: onOpenHelpSidebar, onClose} = useDisclosure()
  const focusRef = useRef<HTMLInputElement | null>(null)

  const jobQueryLink = `/jobquery?cluster=${clusterName}`

  const handleSubmitClick = (query: string) => {
    setSearchParams({query})
  }

  const subclusterLinks = selectedCluster.subclusters.map((subcluster: Subcluster) => (
    <ChakraLink key={subcluster.name} asChild variant="underline" colorPalette="blue">
      <ReactRouterLink to={`/${clusterName}/subcluster/${subcluster.name}`}>
        {subcluster.name}
        {' '}
      </ReactRouterLink>
    </ChakraLink>
  ))


  const SubClusterLinks = () => {
    return (
      <HStack display="inline-flex" gap={1}>
        {selectedCluster.subclusters.map((subcluster: Subcluster) => (
          <ChakraLink key={subcluster.name} asChild variant="underline" colorPalette="blue">
            <ReactRouterLink to={`/${clusterName}/subcluster/${subcluster.name}`}>
              {subcluster.name}
            </ReactRouterLink>
          </ChakraLink>
        ))}
      </HStack>
    )
  }

  return (
    <>
      <PageTitle title={`${selectedCluster.name} Dashboard`}/>
      <VStack gap={1}>
        <Heading size={{base: 'lg', md: 'xl'}} mb={1}>{selectedCluster.name}: Jobanalyzer Dashboard</Heading>
        <Text>Click on hostname for machine details.</Text>
        <Text>
          <ChakraLink asChild mr="10px" target="_blank" variant="underline" colorPalette="blue">
            <ReactRouterLink to={jobQueryLink}>
              Job query <LuExternalLink/>
            </ReactRouterLink>
          </ChakraLink>
          {
            subclusterLinks.length > 0 &&
            <>
              Aggregates:{' '}
              <SubClusterLinks/>
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
      <ChakraLink asChild variant="underline" colorPalette="blue">
        <ReactRouterLink to={to}>
          {children}
        </ReactRouterLink>
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

