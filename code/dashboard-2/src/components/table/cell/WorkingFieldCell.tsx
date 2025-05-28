import { Text } from '@chakra-ui/react'

import { CELL_BACKGROUND_COLORS } from '../../../Constants.ts'
import { CellWrapper } from './CellWrapper.tsx'
import { useColorModeValue } from '../../ui/color-mode.tsx'

interface WorkingFieldCellProps {
  value: number;
}

export const WorkingFieldCell = ({value}: WorkingFieldCellProps) => {
  const backgroundColor = useColorModeValue(
    value >= 75
      ? CELL_BACKGROUND_COLORS.LIGHT.WORKING_HARD
      : value >= 50
        ? CELL_BACKGROUND_COLORS.LIGHT.WORKING
        : value >= 25
          ? CELL_BACKGROUND_COLORS.LIGHT.COASTING
          : CELL_BACKGROUND_COLORS.LIGHT.NA,
    value >= 75
      ? CELL_BACKGROUND_COLORS.DARK.WORKING_HARD
      : value >= 50
        ? CELL_BACKGROUND_COLORS.DARK.WORKING
        : value >= 25
          ? CELL_BACKGROUND_COLORS.DARK.COASTING
          : CELL_BACKGROUND_COLORS.DARK.NA
  )
  return (
    <CellWrapper styleProps={{backgroundColor}}>
      <Text>{value}</Text>
    </CellWrapper>
  )
}
