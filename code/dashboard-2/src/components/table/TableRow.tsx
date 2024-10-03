import { flexRender, Row } from '@tanstack/react-table'
import { Td, Tr } from '@chakra-ui/react'
import React from 'react'

interface TableRowProps<T> {
  row: Row<T>,
  styles?: React.CSSProperties
}

export const TableRow = ({row, styles}: TableRowProps<any>) => {

  let rowStyles: React.CSSProperties = {
    borderBottom: '1px solid',
    borderColor: 'gray.200',
    ...styles
  }

  return (
    <Tr key={row.id} style={{...rowStyles}}>
      {row.getAllCells().map((cell, cellIndex, cellArray) => {
        const meta: any = cell.column.columnDef.meta
        return (
          <Td
            key={cell.id}
            isNumeric={meta?.isNumeric}
            borderRight={cellIndex !== cellArray.length - 1 ? '1px solid' : 'none'}
            borderColor="gray.200"
            padding={0}
          >
            {flexRender(cell.column.columnDef.cell, cell.getContext())}
          </Td>
        )
      })}
    </Tr>
  )
}
