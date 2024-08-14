import { Table, Tbody } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'
import DashboardTableRow from './DashboardTableRow.tsx'
import TableHeader from './TableHeader.tsx'
import { Cluster } from '../../types/Cluster.ts'

interface DashboardTableProps {
  table: TableType<DashboardTableItem>
  cluster: Cluster
}

const DashboardTable = ({table, cluster}: DashboardTableProps) => {
  return (
    <Table size="sm" border="1px solid" borderColor="gray.200">
      <TableHeader table={table}/>
      <Tbody>
        {table.getRowModel().rows.map((row) =>
          <DashboardTableRow row={row} cluster={cluster} key={row.id}/>
        )}
      </Tbody>
    </Table>
  )
}

export default DashboardTable
