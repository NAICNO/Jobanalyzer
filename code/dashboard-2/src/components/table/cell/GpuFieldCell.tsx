import { Text } from '@chakra-ui/react'

import { CELL_BACKGROUND_COLORS } from '../../../Constants.ts'
import { CellWrapper } from './CellWrapper.tsx'
import { useColorModeValue } from '../../ui/color-mode.tsx'

interface GpuFieldCellProps {
  value: number;
}

export const GpuFieldCell = ({value}: GpuFieldCellProps) => {

  const isGpuDownOrUndefined = value !== 0 || value === undefined

  const backgroundColor = useColorModeValue(
    isGpuDownOrUndefined ? CELL_BACKGROUND_COLORS.LIGHT.DOWN : CELL_BACKGROUND_COLORS.LIGHT.NA,
    isGpuDownOrUndefined ? CELL_BACKGROUND_COLORS.DARK.DOWN : CELL_BACKGROUND_COLORS.DARK.NA
  )

  return (
    <CellWrapper styleProps={{backgroundColor}}>
      <Text>{value}</Text>
    </CellWrapper>
  )
}
