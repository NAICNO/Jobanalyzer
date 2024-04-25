import { Text } from '@chakra-ui/react'
import { CELL_BACKGROUND_COLORS } from '../../../Constants.ts'
import CellWrapper from './CellWrapper.tsx'

interface WorkingFieldCellProps {
  value: number;
}

const WorkingFieldCell = ({value}: WorkingFieldCellProps) => {
  let backgroundColor = CELL_BACKGROUND_COLORS.NA
  if (value >= 75) {
    backgroundColor = CELL_BACKGROUND_COLORS.WORKING_HARD
  } else if (value >= 50) {
    backgroundColor = CELL_BACKGROUND_COLORS.WORKING
  } else if (value >= 25) {
    backgroundColor = CELL_BACKGROUND_COLORS.COASTING
  }
  return (
    <CellWrapper styleProps={{backgroundColor}}>
      <Text>{value}</Text>
    </CellWrapper>
  )
}

export default WorkingFieldCell
