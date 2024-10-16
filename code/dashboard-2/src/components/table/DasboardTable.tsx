import { Table, Tbody } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'

import { DashboardTableRow } from './DashboardTableRow'
import { TableHeader } from './TableHeader.tsx'
import { Cluster, DashboardTableItem } from '../../types/'

interface DashboardTableProps {
  table: TableType<DashboardTableItem>
  cluster: Cluster
}

export const DashboardTable = ({table, cluster}: DashboardTableProps) => {
  return (
    <Table size="sm">
      <TableHeader table={table}/>
      <Tbody>
        {table.getRowModel().rows.map((row) =>
          <DashboardTableRow row={row} cluster={cluster} key={row.id}/>
        )}
      </Tbody>
    </Table>
  )
}
