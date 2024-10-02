import React from 'react'
import { Row } from '@tanstack/react-table'

import { CELL_BACKGROUND_COLORS } from '../../Constants.ts'
import { TableRow } from './TableRow'
import { Cluster, DashboardTableItem } from '../../types'

interface DashboardTableRowProps {
  row: Row<DashboardTableItem>
  cluster: Cluster
}

export const DashboardTableRow = ({row, cluster}: DashboardTableRowProps) => {

  let rowStyles: React.CSSProperties = {
    borderBottom: '1px solid',
    borderColor: 'gray.200',
  }

  if (cluster.uptime) {
    const allCells = row.getAllCells()

    const cpuStatusCell = allCells.find(cell => cell.column.id === 'cpu_status')

    if (cpuStatusCell?.getValue() != 0) {
      rowStyles = {
        ...rowStyles,
        backgroundColor: CELL_BACKGROUND_COLORS.DOWN,
      }
    }
  }

  return (
    <TableRow row={row} styles={rowStyles}/>
  )
}

