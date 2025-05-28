import { Table } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'

import { TableRow } from './TableRow'
import { TableHeader } from './TableHeader'
import { DeadWeightTableItem } from '../../types'

interface DeadWeightTableProps {
  table: TableType<DeadWeightTableItem>
}

export const DeadWeightTable = ({table}: DeadWeightTableProps) => {
  return (
    <Table.Root size="sm">
      <TableHeader table={table}/>
      <Table.Body>
        {table.getRowModel().rows.map((row) =>
          <TableRow row={row} key={row.id}/>
        )}
      </Table.Body>
    </Table.Root>
  )
}
