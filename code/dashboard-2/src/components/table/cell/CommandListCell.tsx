import { Text } from '@chakra-ui/react'
import CellWrapper from './CellWrapper.tsx'
import { breakText } from '../../../util'

interface CommandListCellProps {
  value: string;
}

const CommandListCell = ({value}: CommandListCellProps) => {
  const brokenText = breakText(value)

  return (
    <CellWrapper>
      <Text>{brokenText}</Text>
    </CellWrapper>
  )
}

export default CommandListCell
