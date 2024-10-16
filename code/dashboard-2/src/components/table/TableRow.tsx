import { flexRender, Row } from '@tanstack/react-table'
import { Td, Tr } from '@chakra-ui/react'
import React from 'react'

interface TableRowProps<T> {
  row: Row<T>,
  styles?: React.CSSProperties
}

export const TableRow = ({row, styles}: TableRowProps<any>) => {

  return (
    <Tr key={row.id} style={styles}>
      {row.getAllCells().map((cell, cellIndex, cellArray) => {
        const meta: any = cell.column.columnDef.meta
        return (
          <Td
            key={cell.id}
            isNumeric={meta?.isNumeric}
            padding={0}
          >
            {flexRender(cell.column.columnDef.cell, cell.getContext())}
          </Td>
        )
      })}
    </Tr>
  )
}
