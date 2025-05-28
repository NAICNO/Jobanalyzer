import { flexRender, Row } from '@tanstack/react-table'
import { Table } from '@chakra-ui/react'
import React from 'react'

interface TableRowProps<T> {
  row: Row<T>,
  styles?: React.CSSProperties
}

export const TableRow = ({row, styles}: TableRowProps<any>) => {

  return (
    <Table.Row key={row.id} style={styles}>
      {row.getAllCells().map((cell, cellIndex, cellArray) => {
        const meta: any = cell.column.columnDef.meta
        return (
          <Table.Cell
            key={cell.id}
            padding={0}
          >
            {flexRender(cell.column.columnDef.cell, cell.getContext())}
          </Table.Cell>
        )
      })}
    </Table.Row>
  )
}
