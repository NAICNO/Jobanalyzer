import { Table, Tbody } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'

import { TableRow } from './TableRow'
import { TableHeader } from './TableHeader'
import { DeadWeightTableItem } from '../../types'

interface DeadWeightTableProps {
  table: TableType<DeadWeightTableItem>
}

export const DeadWeightTable = ({table}: DeadWeightTableProps) => {
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
