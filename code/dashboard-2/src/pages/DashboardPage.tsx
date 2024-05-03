import { ReactNode, useEffect, useMemo, useRef, useState } from 'react'
import { Navigate, useParams, Link as ReactRouterLink } from 'react-router-dom'
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
import { CLUSTER_INFO, } from '../Constants.ts'
import { isValidateClusterName } from '../util'
import { getDashboardTableColumns } from '../util/TableUtils.ts'
import DashboardTable from '../components/table/DasboardTable.tsx'
import NodeSelectionHelpDrawer from '../components/NodeSelectionHelpDrawer.tsx'
import NodeSelectionInput from '../components/NodeSelectionInput.tsx'

const emptyArray: any[] = []

export default function DashboardPage() {

  const {clusterName} = useParams<string>()

  if (!isValidateClusterName(clusterName)) {
    return (
      <Navigate to="/"/>
    )
  }

  const selectedCluster = CLUSTER_INFO[clusterName!]
  const defaultQuery = selectedCluster.defaultQuery

  const [query, setQuery] = useState<string>(defaultQuery)

  useEffect(() => {
    setQuery(defaultQuery)
  }, [selectedCluster])

  const {data, isFetched} = useFetchDashboard(clusterName!, query)

  const [sorting, setSorting] = useState<SortingState>([])

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

  const {isOpen: isOpenHelpSidebar, onOpen: onOpenHelpSidebar, onClose} = useDisclosure()
  const focusRef = useRef<HTMLInputElement | null>(null)

  return (
    <>
      <Container height="100vh" centerContent>
        <VStack spacing={1}>
          <Heading size="lg" mb={4}>{selectedCluster.name}: Jobanalyzer Dashboard</Heading>
          <Text>Click on hostname for machine details.</Text>
          <Text>
            <ChakraLink as={ReactRouterLink} to="/jobquery" isExternal mr="10px">
              Job query <ExternalLinkIcon mx="2px"/>
            </ChakraLink>
            Aggregates:{' '}
            <ChakraLink as={ReactRouterLink} color="teal.500" href="#">
              nvidia
            </ChakraLink>
          </Text>
          <Text>Recent: 30 mins Longer: 12 hrs{' '} </Text>
          <ViolatorsAndZombiesLinks cluster={selectedCluster}/>
          <NodeSelectionInput
            defaultQuery={defaultQuery}
            onClickSubmit={setQuery}
            onClickHelp={onOpenHelpSidebar}
            focusRef={focusRef}
          />
          <SlideFade in={isFetched}>
            < DashboardTable table={table} cluster={selectedCluster}/>
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

