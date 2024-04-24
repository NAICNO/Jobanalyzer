import { Text } from '@chakra-ui/react'
import CellWrapper from './CellWrapper.tsx'

interface GenericCellProps {
  value: number;
}

const GenericCell = ({value}: GenericCellProps) => {
  return (
    <CellWrapper  styleProps={{paddingLeft: 4}}>
      <Text>{value}</Text>
    </CellWrapper>
  )
}

export default GenericCell
