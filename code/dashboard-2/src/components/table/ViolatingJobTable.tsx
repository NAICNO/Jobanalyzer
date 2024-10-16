import { Table, Tbody } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'

import { ViolatingJobTableItem } from '../../types'
import { TableHeader } from './TableHeader.tsx'
import { TableRow } from './TableRow.tsx'

interface ViolatingJobTableProps {
  table: TableType<ViolatingJobTableItem>
}

export const ViolatingJobTable = ({table}: ViolatingJobTableProps) => {
  return (
    <Table size="sm">
      <TableHeader table={table}/>
      <Tbody>
        {table.getRowModel().rows.map((row) =>
          <TableRow row={row} key={row.id}/>
        )}
      </Tbody>
    </Table>
  )
}
