import { Table, Tbody } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'
import { ViolatingUserTableItem } from '../../types'

import { TableHeader } from './TableHeader.tsx'
import { TableRow } from './TableRow.tsx'

interface ViolatingUserTableProps {
  table: TableType<ViolatingUserTableItem>
}

export const ViolatingUserTable = ({table}: ViolatingUserTableProps) => {
  return (
    <Table size="sm" border="1px solid" borderColor="gray.200">
      <TableHeader table={table}/>
      <Tbody>
        {table.getRowModel().rows.map((row) =>
          <TableRow row={row} key={row.id}/>
        )}
      </Tbody>
    </Table>
  )
}
