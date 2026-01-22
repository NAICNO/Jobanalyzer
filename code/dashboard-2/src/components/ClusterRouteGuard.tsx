import { useEffect, useState } from 'react'
import { useNavigate, useLocation, Outlet, useParams } from 'react-router'
import { Box, Spinner, VStack, Text } from '@chakra-ui/react'
import { useCluster } from '../hooks/useCluster'
import { useAuth } from '../hooks/useAuth'
import { getClusterConfig } from '../config/clusters'
import { toaster } from './ui/toaster'

/**
 * Route guard that:
 * 1. Redirects to cluster selection page if user has no clusters selected
 * 2. Checks authentication for clusters that require it
 * 3. Redirects to login if authentication is required but not present
 */
export const ClusterRouteGuard = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { clusterName } = useParams<{ clusterName: string }>()
  const { hasSelectedClusters } = useCluster()
  const { isAuthenticated, login } = useAuth()
  const [isCheckingAuth, setIsCheckingAuth] = useState(true)

  useEffect(() => {
    const checkAccess = async () => {
      // Don't redirect if we're already on the selection page
      if (location.pathname === '/v2/select-cluster') {
        setIsCheckingAuth(false)
        return
      }

      // Redirect to selection page if no clusters are selected
      if (!hasSelectedClusters) {
        navigate('/v2/select-cluster', { replace: true })
        setIsCheckingAuth(false)
        return
      }

      // Check authentication if cluster requires it
      if (clusterName) {
        const clusterConfig = getClusterConfig(clusterName)
        
        if (clusterConfig?.requiresAuth) {
          const authenticated = isAuthenticated(clusterName)
          
          if (!authenticated) {
            // Show toast notification
            toaster.create({
              title: 'Authentication Required',
              description: `You need to authenticate to access ${clusterConfig.name}`,
              type: 'info',
              duration: 4000,
            })

            try {
              // Store the current path to return after auth
              await login(clusterName, location.pathname)
              setIsCheckingAuth(false) // Add this
              return
            } catch (error) {
              console.error('Failed to initiate login:', error)
              setIsCheckingAuth(false) // Add this
              navigate('/v2/select-cluster', { replace: true })
              return
            }
          }
        }
      }

      setIsCheckingAuth(false)
    }

    checkAccess()
  }, [hasSelectedClusters, clusterName, isAuthenticated, login, navigate, location.pathname])

  // Show loading state while checking authentication
  if (isCheckingAuth) {
    return (
      <Box
        display="flex"
        alignItems="center"
        justifyContent="center"
        minH="60vh"
      >
        <VStack gap={4}>
          <Spinner size="xl" color="blue.500" />
          <Text color="gray.600">Checking access...</Text>
        </VStack>
      </Box>
    )
  }

  // If no clusters selected and not on selection page, don't render children
  if (!hasSelectedClusters && location.pathname !== '/v2/select-cluster') {
    return null
  }

  return <Outlet />
}
