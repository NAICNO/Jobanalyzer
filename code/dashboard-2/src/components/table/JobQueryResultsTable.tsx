import { Table } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'

import { TableRow } from './TableRow'
import { TableHeader } from './TableHeader'
import { JobQueryResultsTableItem } from '../../types'

interface JobQueryResultsTableProps {
  table: TableType<JobQueryResultsTableItem>
}

export const JobQueryResultsTable = ({table}: JobQueryResultsTableProps) => {
  return (
    <Table.ScrollArea borderWidth="1px" maxHeight="800px">
      <Table.Root
        size="sm"
        showColumnBorder
        stickyHeader
        css={{
          tableLayout: 'fixed',
          width: '100%',
          '& td, & th': {
            whiteSpace: 'normal',
            wordBreak: 'break-word',
          },
        }}
      >
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
