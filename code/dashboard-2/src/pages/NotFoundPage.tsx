import { Box, Button, EmptyState, Text, VStack } from '@chakra-ui/react'
import { useLocation, useNavigate, useRouteError, isRouteErrorResponse } from 'react-router'
import { LuRadar, LuTriangleAlert } from 'react-icons/lu'

export const NotFoundPage = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const error = useRouteError()

  const isError = error != null
  let title = '404 - NOT FOUND'
  let description = `No route matches ${location.pathname}`
  let icon = <LuRadar />

  if (isError) {
    title = 'UNEXPECTED ERROR'
    icon = <LuTriangleAlert />
    if (isRouteErrorResponse(error)) {
      title = `${error.status} - ${error.statusText}`
      description = error.data?.toString() || 'An error occurred while loading this page.'
    } else if (error instanceof Error) {
      description = error.message
    } else {
      description = 'The application encountered an unexpected error.'
    }
  }

  return (
    <Box
      display="flex"
      alignItems="center"
      justifyContent="center"
      minH="70vh"
    >
      <EmptyState.Root size="lg">
        <EmptyState.Content>
          <EmptyState.Indicator>
            {icon}
          </EmptyState.Indicator>
          <VStack textAlign="center" gap="1">
            <EmptyState.Title
              fontFamily="'IBM Plex Mono', monospace"
              fontSize="lg"
              fontWeight="500"
              letterSpacing="0.05em"
            >
              {title}
            </EmptyState.Title>
            <EmptyState.Description>
              <Text as="span" fontFamily="'IBM Plex Mono', monospace" fontSize="sm">
                {description}
              </Text>
            </EmptyState.Description>
          </VStack>
          <Button variant="outline" size="sm" onClick={() => navigate('/select-cluster')}>
            Go to Cluster Selection
          </Button>
        </EmptyState.Content>
      </EmptyState.Root>
    </Box>
  )
}
