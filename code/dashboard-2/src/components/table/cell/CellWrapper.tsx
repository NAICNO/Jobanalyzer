import { ReactNode } from 'react'
import { Box, JsxStyleProps } from '@chakra-ui/react'

interface CellWrapperProps {
  styleProps?: JsxStyleProps
  children: ReactNode
}

export const CellWrapper = ({styleProps, children}: CellWrapperProps) => {
  return (
    <Box padding={1} {...styleProps}>
      {children}
    </Box>
  )
}

