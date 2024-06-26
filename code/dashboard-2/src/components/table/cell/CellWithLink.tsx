import { Link as ReactRouterLink } from 'react-router-dom'
import { Link as ChakraLink } from '@chakra-ui/react'

import CellWrapper from './CellWrapper.tsx'

interface CellWithLinkProps {
  value: TextWithLink
}

const CellWithLink = ({value}: CellWithLinkProps) => {
  return (
    <CellWrapper styleProps={{paddingLeft: 2}}>
      <ChakraLink as={ReactRouterLink} to={value.link} color={'teal.500'}>
        {value.text}
      </ChakraLink>
    </CellWrapper>
  )
}

export default CellWithLink
