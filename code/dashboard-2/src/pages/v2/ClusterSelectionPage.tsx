import { Box, Grid, Heading, Text, VStack } from '@chakra-ui/react'
import { useNavigate } from 'react-router'
import { AVAILABLE_CLUSTERS, getClusterConfig } from '../../config/clusters'
import { useCluster } from '../../hooks/useCluster'
import { getClusterFullName } from '../../config/clusters'
import { loginToCluster } from '../../utils/oidcManager'
import { toaster } from '../../components/ui/toaster'
import { ClusterCard } from '../../components/ClusterCard'

export const ClusterSelectionPage = () => {
  const navigate = useNavigate()
  const { addCluster, removeCluster, selectedClusters } = useCluster()

  const handleClusterSelect = async (clusterId: string) => {
    const clusterConfig = getClusterConfig(clusterId)
    
    if (!clusterConfig) {
      console.error('Cluster config not found:', clusterId)
      return
    }
    
    // Check if cluster requires authentication
    if (clusterConfig.requiresAuth) {
      try {
        // Don't add cluster yet - it will be added after successful authentication
        // Use oidc-client-ts to initiate the authentication flow with PKCE
        const returnPath = `/v2/${getClusterFullName(clusterId)}/overview`
        await loginToCluster(clusterId, returnPath)
      } catch (error) {
        console.error('Failed to initiate login:', error)
        toaster.create({
          title: 'Authentication failed',
          description: error instanceof Error ? error.message : 'Failed to initiate login',
          type: 'error',
          duration: 5000,
        })
      }
    } else {
      // For non-auth clusters, add immediately and navigate
      addCluster(clusterId)
      const clusterFullName = getClusterFullName(clusterId)
      navigate(`/v2/${clusterFullName}/overview`)
    }
  }

  const handleRemoveCluster = async (clusterId: string, event: React.MouseEvent) => {
    event.stopPropagation()
    await removeCluster(clusterId)
  }

  return (
    <Box p={8} maxW="1400px" mx="auto">
      <VStack gap={6} align="start" mb={8}>
        <Heading size="xl">Select a Cluster</Heading>
        <Text fontSize="lg" color="gray.600">
          Choose a cluster to add to your dashboard. You can add multiple clusters and switch between them.
        </Text>
      </VStack>

      <Grid
        templateColumns={{
          base: '1fr',
          md: 'repeat(2, 1fr)',
          lg: 'repeat(3, 1fr)',
        }}
        gap={6}
      >
        {AVAILABLE_CLUSTERS.map((cluster) => (
          <ClusterCard
            key={cluster.id}
            cluster={cluster}
            isSelected={selectedClusters.includes(cluster.id)}
            onSelect={handleClusterSelect}
            onRemove={handleRemoveCluster}
          />
        ))}
      </Grid>

      {selectedClusters.length > 0 && (
        <Box mt={8} p={4} bg="blue.50" borderRadius="md">
          <Text fontSize="sm" color="blue.800">
            <strong>Tip:</strong> You have {selectedClusters.length} cluster(s) added. 
            You can switch between them using the sidebar navigation.
          </Text>
        </Box>
      )}
    </Box>
  )
}
