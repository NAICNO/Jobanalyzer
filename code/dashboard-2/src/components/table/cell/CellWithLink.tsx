import { Link as ReactRouterLink } from 'react-router'
import { Link as ChakraLink } from '@chakra-ui/react'

import { CellWrapper } from './CellWrapper.tsx'
import { TextWithLink } from '../../../types'

interface CellWithLinkProps {
  value: TextWithLink
}

export const CellWithLink = ({value}: CellWithLinkProps) => {
  const {text, link, openInNewTab} = value
  return (
    <CellWrapper styleProps={{paddingLeft: 2}}>
      <ChakraLink as={ReactRouterLink} to={link} isExternal={openInNewTab}>
        {text}
      </ChakraLink>
    </CellWrapper>
  )
}
