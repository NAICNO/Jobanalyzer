import { Table } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'

import { ViolatingJobTableItem } from '../../types'
import { TableHeader } from './TableHeader.tsx'
import { TableRow } from './TableRow.tsx'

interface ViolatingJobTableProps {
  table: TableType<ViolatingJobTableItem>
}

export const ViolatingJobTable = ({table}: ViolatingJobTableProps) => {
  return (
    <Table.ScrollArea borderWidth="1px" maxHeight="600px">
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
