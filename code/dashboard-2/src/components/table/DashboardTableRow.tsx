import { flexRender, Row } from '@tanstack/react-table'
import { Td, Tr } from '@chakra-ui/react'
import { CELL_BACKGROUND_COLORS } from '../../Constants.ts'
import React from 'react'

interface DashboardTableRowProps {
  row: Row<DashboardTableItem>
  cluster: Cluster
}

const DashboardTableRow = ({row, cluster}: DashboardTableRowProps) => {

  let rowStyles : React.CSSProperties = {
    borderBottom: '1px solid',
    borderColor: 'gray.200',
  }

  if(cluster.uptime){
    const allCells = row.getAllCells()

    const cpuStatusCell = allCells.find(cell => cell.column.id === 'cpu_status')

    if(cpuStatusCell?.getValue() !=0) {
      rowStyles = {
        ...rowStyles,
        backgroundColor: CELL_BACKGROUND_COLORS.DOWN,
      }
    }
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

export default DashboardTableRow
