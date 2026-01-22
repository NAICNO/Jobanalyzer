import { useEffect, useState, useRef } from 'react'
import { useNavigate } from 'react-router'
import { Box, Heading, Text, Spinner, VStack } from '@chakra-ui/react'
import { handleAuthCallback } from '../../utils/oidcManager'
import { setClusterAuth } from '../../utils/secureStorage'
import { toaster } from '../../components/ui/toaster'
import { useAuth } from '../../hooks/useAuth'
import { useCluster } from '../../hooks/useCluster'

export const CallbackPage = () => {
  const navigate = useNavigate()
  const { refreshAuthState } = useAuth()
  const { addCluster } = useCluster()
  const [status, setStatus] = useState<'processing' | 'success' | 'error'>('processing')
  const [errorMessage, setErrorMessage] = useState<string>('')
  const hasProcessed = useRef(false) // Prevent double processing in React StrictMode

  useEffect(() => {
    const processCallback = async () => {
      // Prevent double processing (React StrictMode issue)
      if (hasProcessed.current) {
        return
      }
      hasProcessed.current = true

      try {
        setStatus('processing')

        // Handle the OIDC callback
        const result = await handleAuthCallback()

        // Store tokens in secureStorage
        if (result.user) {
          await setClusterAuth(result.clusterId, {
            accessToken: result.user.access_token,
            refreshToken: result.user.refresh_token,
            expiresAt: result.user.expires_at,
            userId: result.user.profile.sub,
          })

          // Refresh auth state in context
          await refreshAuthState(result.clusterId)

          // Add cluster to selected clusters after successful authentication
          addCluster(result.clusterId)

          setStatus('success')

          // Show success toast
          toaster.create({
            title: 'Authentication successful',
            description: 'You have been successfully authenticated',
            type: 'success',
            duration: 3000,
          })

          // Redirect to the stored path or cluster overview immediately
          navigate(result.returnPath, { replace: true })
        } else {
          throw new Error('No user data received from authentication')
        }
      } catch (error) {
        console.error('Callback processing failed:', error)
        setStatus('error')
        
        const message = error instanceof Error ? error.message : 'Unknown error occurred'
        setErrorMessage(message)

        toaster.create({
          title: 'Authentication failed',
          description: message,
          type: 'error',
          duration: 5000,
        })

        // Redirect to cluster selection after error
        setTimeout(() => {
          navigate('/v2/select-cluster', { replace: true })
        }, 3000)
      }
    }

    processCallback()
  }, [navigate, refreshAuthState])

  return (
    <Box
      display="flex"
      alignItems="center"
      justifyContent="center"
      minH="100vh"
      bg="gray.50"
    >
      <VStack gap={6} p={8} bg="white" borderRadius="lg" shadow="md" maxW="md">
        {status === 'processing' && (
          <>
            <Spinner size="xl" color="blue.500" />
            <Heading size="lg">Processing Authentication</Heading>
            <Text color="gray.600" textAlign="center">
              Please wait while we complete your authentication...
            </Text>
          </>
        )}

        {status === 'success' && (
          <>
            <Box fontSize="4xl" color="green.500">
              ✓
            </Box>
            <Heading size="lg" color="green.600">
              Authentication Successful
            </Heading>
            <Text color="gray.600" textAlign="center">
              Redirecting you to your cluster...
            </Text>
          </>
        )}

        {status === 'error' && (
          <>
            <Box fontSize="4xl" color="red.500">
              ✗
            </Box>
            <Heading size="lg" color="red.600">
              Authentication Failed
            </Heading>
            <Text color="gray.600" textAlign="center">
              {errorMessage || 'An error occurred during authentication'}
            </Text>
            <Text fontSize="sm" color="gray.500" textAlign="center">
              Redirecting to cluster selection...
            </Text>
          </>
        )}
      </VStack>
    </Box>
  )
}
