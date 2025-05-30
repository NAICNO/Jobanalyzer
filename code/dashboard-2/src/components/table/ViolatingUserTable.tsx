import { Table } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'
import { ViolatingUserTableItem } from '../../types'

import { TableHeader } from './TableHeader.tsx'
import { TableRow } from './TableRow.tsx'

interface ViolatingUserTableProps {
  table: TableType<ViolatingUserTableItem>
}

export const ViolatingUserTable = ({table}: ViolatingUserTableProps) => {
  return (
    <Table.ScrollArea borderWidth="1px"  maxHeight="500px">
      <Table.Root size="sm" showColumnBorder stickyHeader>
        <TableHeader table={table}/>
        <Table.Body>
          {table.getRowModel().rows.map((row) =>
            <TableRow row={row} key={row.id}/>
          )}
        </Table.Body>
      </Table.Root>
    </Table.ScrollArea>
  )
}
