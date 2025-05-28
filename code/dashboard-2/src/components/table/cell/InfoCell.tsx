import { Icon } from '@chakra-ui/react'
import { MdInfoOutline } from 'react-icons/md'

import { CellWrapper } from './CellWrapper.tsx'
import { Tooltip } from '../../ui/tooltip.tsx'

interface InfoCellProps {
  value: string;
}

export const InfoCell = ({value}: InfoCellProps) => {
  return (
    <CellWrapper>
      <Tooltip
        content={value}
        showArrow
        openDelay={500}
      >
        <Icon>
          <MdInfoOutline />
        </Icon>
      </Tooltip>
    </CellWrapper>
  )
}
