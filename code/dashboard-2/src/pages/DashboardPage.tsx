import { Center, Container, Table, Tbody, Td, Th, Thead, Tr, VStack, chakra } from '@chakra-ui/react'
import { useFetchDashboardTable } from '../hooks/useFetchDashboardTable.ts'
import {
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable
} from '@tanstack/react-table'
import { useMemo, useState } from 'react'
import { TriangleDownIcon, TriangleUpIcon } from '@chakra-ui/icons'
import { CLUSTER_INFO, } from '../Constants.ts'
import { Navigate, useParams } from 'react-router-dom'
import { isValidateClusterName } from '../util'
import { getTableColumns } from '../util/TableUtils.ts'

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
      <Center height="100%">
        <VStack spacing={6} width="100%" maxW="md" padding="4">
          <Table size="sm" border="1px solid" borderColor="gray.200">
            <Thead borderBottom="1px solid" borderColor="gray.200">
              {table.getHeaderGroups().map((headerGroup) => (
                <Tr key={headerGroup.id}>
                  {headerGroup.headers.map((header,) => {
                    const meta: any = header.column.columnDef.meta
                    return (
                      <Th
                        key={header.id}
                        borderRight={header.index !== headerGroup.headers.length - 1 ? '1px solid' : 'none'}
                        borderColor="gray.200"
                        onClick={header.column.getToggleSortingHandler()}
                        isNumeric={meta?.isNumeric}
                        colSpan={header.colSpan}
                        style={{textTransform: 'none'}}
                        title={meta?.helpText}
                      >
                        {
                          flexRender(header.column.columnDef.header, header.getContext())
                        }
                        <chakra.span>
                          {
                            header.column.getIsSorted() ? (
                              header.column.getIsSorted() === 'desc' ?
                                <TriangleDownIcon aria-label="sorted descending"/>
                                :
                                <TriangleUpIcon aria-label="sorted ascending"/>

                            ) : null
                          }
                        </chakra.span>
                      </Th>
                    )
                  })}
                </Tr>
              ))}
            </Thead>
            <Tbody>
              {table.getRowModel().rows.map((row) => (
                <Tr key={row.id} borderBottom="1px solid" borderColor="gray.200">
                  {row.getAllCells().map((cell, cellIndex, cellArray) => {
                    const meta: any = cell.column.columnDef.meta
                    return (
                      <Td
                        key={cell.id}
                        isNumeric={meta?.isNumeric}
                        borderRight={cellIndex !== cellArray.length - 1 ? '1px solid' : 'none'}
                        borderColor="gray.200"
                      >
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </Td>
                    )
                  })}
                </Tr>
              ))}
            </Tbody>
          </Table>
        </VStack>
      </Center>
    </Container>
  )

}
