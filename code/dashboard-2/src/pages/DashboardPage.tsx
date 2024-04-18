import { Container, Heading, VStack } from '@chakra-ui/react'
import { useFetchDashboardTable } from '../hooks/useFetchDashboardTable.ts'
import {
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable
} from '@tanstack/react-table'
import { useMemo, useState } from 'react'
import { CLUSTER_INFO, } from '../Constants.ts'
import { Navigate, useParams } from 'react-router-dom'
import { isValidateClusterName } from '../util'
import { getTableColumns } from '../util/TableUtils.ts'
import DashboardTable from '../components/table/DasboardTable.tsx'

const emptyArray: any[] = []

export default function DashboardPage() {

  const {clusterName} = useParams<string>()

  if (!isValidateClusterName(clusterName)) {
    return (
      <Navigate to="/"/>
    )
  }

  const {data} = useFetchDashboardTable(clusterName!)

  const [sorting, setSorting] = useState<SortingState>([])

  const selectedCluster = CLUSTER_INFO[clusterName!]
  const tableColumns = useMemo(() => getTableColumns(selectedCluster), [clusterName])

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
      <VStack spacing={6} width="100%" padding="4">
        <Heading size="lg" mb={4}>{selectedCluster.name}: Jobanalyzer Dashboard</Heading>
        <DashboardTable table={table} cluster={selectedCluster}/>
      </VStack>
    </Container>
  )
}
