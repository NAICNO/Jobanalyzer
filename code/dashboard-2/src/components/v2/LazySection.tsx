import type { ReactNode } from 'react'
import { Box, Skeleton, VStack } from '@chakra-ui/react'
import { useIntersectionObserver } from '../../hooks/useIntersectionObserver'

interface LazySectionProps {
  children: (isVisible: boolean) => ReactNode
  minHeight?: string | number
}

export const LazySection = ({ children, minHeight = '200px' }: LazySectionProps) => {
  const { ref, hasBeenVisible } = useIntersectionObserver()

  return (
    <Box ref={ref} w="100%" minH={hasBeenVisible ? undefined : minHeight}>
      {hasBeenVisible ? (
        children(true)
      ) : (
        <VStack w="100%" gap={3}>
          <Skeleton height="20px" width="200px" />
          <Skeleton height={minHeight} width="100%" />
        </VStack>
      )}
    </Box>
  )
}
