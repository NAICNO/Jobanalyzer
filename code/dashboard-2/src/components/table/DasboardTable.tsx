import { Table } from '@chakra-ui/react'
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
    <Table.ScrollArea borderWidth="0.5px" maxHeight="600px" width="100%">
      <Table.Root size="sm" showColumnBorder stickyHeader>
        <TableHeader table={table}/>
        <Table.Body>
          {table.getRowModel().rows.map((row) =>
            <DashboardTableRow row={row} cluster={cluster} key={row.id}/>
          )}
        </Table.Body>
      </Table.Root>
    </Table.ScrollArea>
  )
}
