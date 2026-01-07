import { Box, Card, Grid, Heading, Text, VStack, HStack, Icon, IconButton } from '@chakra-ui/react'
import { useNavigate } from 'react-router'
import { LuX } from 'react-icons/lu'
import { AVAILABLE_CLUSTERS } from '../../config/clusters'
import { useCluster } from '../../hooks/useCluster'
import { getClusterFullName } from '../../config/clusters'

export const ClusterSelectionPage = () => {
  const navigate = useNavigate()
  const { addCluster, removeCluster, selectedClusters } = useCluster()

  const handleClusterSelect = (clusterId: string) => {
    addCluster(clusterId)
    
    // Navigate to the cluster's overview page
    const clusterFullName = getClusterFullName(clusterId)
    navigate(`/v2/${clusterFullName}/overview`)
  }

  const handleRemoveCluster = (clusterId: string, event: React.MouseEvent) => {
    event.stopPropagation()
    removeCluster(clusterId)
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
        {AVAILABLE_CLUSTERS.map((cluster) => {
          const isSelected = selectedClusters.includes(cluster.id)

          return (
            <Card.Root
              key={cluster.id}
              cursor="pointer"
              transition="all 0.2s"
              borderWidth="2px"
              shadow='md'
              borderColor={isSelected ? 'blue.500' : 'transparent'}
              _hover={{
                borderColor: isSelected ? 'blue.600' : 'gray.200',
                transform: 'translateY(-2px)',
                shadow: 'lg',
              }}
              onClick={() => !isSelected && handleClusterSelect(cluster.id)}
              opacity={isSelected ? 0.6 : 1}
              position="relative"
            >
              <Card.Body>
                {isSelected && (
                  <IconButton
                    position="absolute"
                    top={2}
                    right={2}
                    size="xs"
                    variant="ghost"
                    colorPalette="red"
                    aria-label="Remove cluster"
                    onClick={(e) => handleRemoveCluster(cluster.id, e)}
                  >
                    <LuX />
                  </IconButton>
                )}
                <VStack gap={4} align="start">
                  <HStack gap={3}>
                    <Icon fontSize="3xl" color="blue.500">
                      <cluster.icon />
                    </Icon>
                    <VStack align="start" gap={0}>
                      <Heading size="md">{cluster.name}</Heading>
                      <Text fontSize="xs" color="gray.500">
                        {cluster.id}
                      </Text>
                    </VStack>
                  </HStack>

                  {cluster.description && (
                    <Text fontSize="sm" color="gray.600">
                      {cluster.description}
                    </Text>
                  )}

                  {isSelected && (
                    <Box
                      px={3}
                      py={1}
                      bg="blue.500"
                      color="white"
                      borderRadius="md"
                      fontSize="xs"
                      fontWeight="semibold"
                    >
                      Already Added
                    </Box>
                  )}
                </VStack>
              </Card.Body>
            </Card.Root>
          )
        })}
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
