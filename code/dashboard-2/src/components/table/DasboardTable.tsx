import { Table, Tbody } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'
import DashboardTableRow from './DashboardTableRow.tsx'
import DashboardTableHeader from './DashboardTableHeader.tsx'

interface DashboardTableProps {
  table: TableType<DashboardTableItem>
  cluster: Cluster
}

const DashboardTable = ({table, cluster}: DashboardTableProps) => {
  return (
    <Table size="sm" border="1px solid" borderColor="gray.200">
      <DashboardTableHeader table={table}/>
      <Tbody>
        {table.getRowModel().rows.map((row) =>
          <DashboardTableRow row={row} cluster={cluster}/>
        )}
      </Tbody>
    </Table>
  )
}

export default DashboardTable
