import { Link as ReactRouterLink } from 'react-router-dom'
import { Link as ChakraLink } from '@chakra-ui/react'

import CellWrapper from './CellWrapper.tsx'

interface CellWithLinkProps {
  value: TextWithLink
}

const CellWithLink = ({value}: CellWithLinkProps) => {
  const {text, link, openInNewTab} = value
  return (
    <CellWrapper styleProps={{paddingLeft: 2}}>
      <ChakraLink as={ReactRouterLink} to={link} color={'teal.500'} isExternal={openInNewTab}>
        {text}
      </ChakraLink>
    </CellWrapper>
  )
}

export default CellWithLink
