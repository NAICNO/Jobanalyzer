import { Text } from '@chakra-ui/react'
import CellWrapper from './CellWrapper.tsx'
import { HOSTNAMES_ALIAS } from '../../../Constants.ts'

interface HostNameFieldCellProps {
  value: string;
}

const HostNameFieldCell = ({value}: HostNameFieldCellProps) => {
  const text = HOSTNAMES_ALIAS[value] || value
  return (
    <CellWrapper styleProps={{paddingLeft: 4}}>
      <Text>{text}</Text>
    </CellWrapper>
  )
}

export default HostNameFieldCell
