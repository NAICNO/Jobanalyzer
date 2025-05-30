import React from 'react'
import { Row } from '@tanstack/react-table'

import { CELL_BACKGROUND_COLORS } from '../../Constants.ts'
import { TableRow } from './TableRow'
import { Cluster, DashboardTableItem } from '../../types'
import { useColorModeValue } from '../ui/color-mode.tsx'

interface DashboardTableRowProps {
  row: Row<DashboardTableItem>
  cluster: Cluster
}

export const DashboardTableRow = ({row, cluster}: DashboardTableRowProps) => {

  let rowStyles: React.CSSProperties = {}

  const downBackgroundColor = useColorModeValue(CELL_BACKGROUND_COLORS.LIGHT.DOWN, CELL_BACKGROUND_COLORS.DARK.DOWN)

  if (cluster.uptime) {
    const allCells = row.getAllCells()

    const cpuStatusCell = allCells.find(cell => cell.column.id === 'cpu_status')

    if (cpuStatusCell?.getValue() != 0) {
      rowStyles = {
        backgroundColor: downBackgroundColor,
      }
    }
  }

  return (
    <TableRow row={row} styles={rowStyles}/>
  )
}

