import { Table, Tbody } from '@chakra-ui/react'
import { Table as TableType } from '@tanstack/react-table'
import TableRow from './TableRow.tsx'
import TableHeader from './TableHeader.tsx'

interface ViolatingJobTableProps {
  table: TableType<ViolatingJobTableItem>
}

const ViolatingJobTable = ({table}: ViolatingJobTableProps) => {
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

export default ViolatingJobTable
