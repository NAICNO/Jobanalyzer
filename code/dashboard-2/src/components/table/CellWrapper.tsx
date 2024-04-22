import { Box, StyleProps } from '@chakra-ui/react'
import { ReactNode } from 'react'

interface CellWrapperProps {
  styleProps?: StyleProps
  children: ReactNode
}

export default function CellWrapper({styleProps, children}: CellWrapperProps) {
  return (
    <Box padding={1} {...styleProps}>
      {children}
    </Box>
  )
}
