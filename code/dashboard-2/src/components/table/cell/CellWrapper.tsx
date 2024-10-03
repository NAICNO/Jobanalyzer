import { ReactNode } from 'react'
import { Box, StyleProps } from '@chakra-ui/react'

interface CellWrapperProps {
  styleProps?: StyleProps
  children: ReactNode
}

export const CellWrapper = ({styleProps, children}: CellWrapperProps) => {
  return (
    <Box padding={1} {...styleProps}>
      {children}
    </Box>
  )
}

