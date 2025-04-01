import { Icon } from '@chakra-ui/icons'
import { Tooltip } from '@chakra-ui/react'

import { CellWrapper } from './CellWrapper.tsx'
import { MdInfoOutline } from 'react-icons/md'

interface InfoCellProps {
  value: string;
}

export const InfoCell = ({value}: InfoCellProps) => {
  return (
    <CellWrapper>
      <Tooltip
        label={value}
        fontSize="md"
        hasArrow
        openDelay={500}
      >
        <Icon as={MdInfoOutline}/>
      </Tooltip>
    </CellWrapper>
  )
}
