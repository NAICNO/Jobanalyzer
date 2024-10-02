import { Text } from '@chakra-ui/react'

import { CELL_BACKGROUND_COLORS } from '../../../Constants.ts'
import { CellWrapper } from './CellWrapper.tsx'

interface GpuFieldCellProps {
  value: number;
}

export const GpuFieldCell = ({value}: GpuFieldCellProps) => {
  let backgroundColor = CELL_BACKGROUND_COLORS.NA
  if (value !== 0 || value === undefined) {
    backgroundColor = CELL_BACKGROUND_COLORS.DOWN
  }
  return (
    <CellWrapper styleProps={{backgroundColor}}>
      <Text>{value}</Text>
    </CellWrapper>
  )
}
