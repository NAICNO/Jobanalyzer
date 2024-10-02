import { Table, Tbody } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'

import { TableRow } from './TableRow'
import { TableHeader } from './TableHeader'
import { JobQueryResultsTableItem } from '../../types'

interface JobQueryResultsTableProps {
  table: TableType<JobQueryResultsTableItem>
}

export const JobQueryResultsTable = ({table}: JobQueryResultsTableProps) => {
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
