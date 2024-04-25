import { Link as ReactRouterLink } from 'react-router-dom'
import { Link as ChakraLink } from '@chakra-ui/react'

import CellWrapper from './CellWrapper.tsx'

interface CellWithLinkProps {
  value: TextWithLink
}

const CellWithLink = ({value}: CellWithLinkProps) => {
  return (
    <CellWrapper styleProps={{paddingLeft: 4}}>
      <ChakraLink as={ReactRouterLink} to={value.link}>
        {value.text}
      </ChakraLink>
    </CellWrapper>
  )
}

export default CellWithLink
