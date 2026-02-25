import { Card, VStack, HStack, Icon, IconButton, Heading, Text, Box } from '@chakra-ui/react'
import { LuX, LuLock, LuShieldAlert } from 'react-icons/lu'
import type { ClusterConfig } from '../config/clusters'

interface ClusterCardProps {
  cluster: ClusterConfig
  isSelected: boolean
  isAuthExpired?: boolean
  onSelect: (clusterId: string) => void
  onRemove: (clusterId: string, event: React.MouseEvent) => void
}

export const ClusterCard = ({ cluster, isSelected, isAuthExpired, onSelect, onRemove }: ClusterCardProps) => {
  return (
    <Card.Root
      cursor="pointer"
      transition="all 0.2s"
      borderWidth="2px"
      shadow="md"
      borderColor={isSelected ? 'blue.500' : 'transparent'}
      _hover={{
        borderColor: isSelected ? 'blue.600' : 'gray.200',
        transform: 'translateY(-2px)',
        shadow: 'lg',
      }}
      onClick={() => (!isSelected || isAuthExpired) && onSelect(cluster.id)}
      opacity={isSelected && !isAuthExpired ? 0.6 : 1}
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
            onClick={(e) => onRemove(cluster.id, e)}
          >
            <LuX />
          </IconButton>
        )}
        <VStack gap={4} align="start">
          <HStack gap={3} width="full">
            <Icon fontSize="3xl" color="blue.500">
              <cluster.icon />
            </Icon>
            <VStack align="start" gap={0} flex={1}>
              <Heading size="md">{cluster.name}</Heading>
              <HStack gap={1} align="center">
                <Text fontSize="xs" color="gray.500">
                  {cluster.id}
                </Text>
                {cluster.requiresAuth && (
                  <Box title="Requires authentication" display="inline-flex">
                    <Icon fontSize="xs" color="gray.500">
                      <LuLock />
                    </Icon>
                  </Box>
                )}
              </HStack>
            </VStack>
          </HStack>

          {cluster.description && (
            <Text fontSize="sm" color="gray.600">
              {cluster.description}
            </Text>
          )}

          {isSelected && (
            <HStack gap={2}>
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
              {isAuthExpired && (
                <HStack
                  gap={1}
                  px={3}
                  py={1}
                  bg="orange.100"
                  color="orange.800"
                  borderRadius="md"
                  fontSize="xs"
                  fontWeight="semibold"
                >
                  <Icon fontSize="sm">
                    <LuShieldAlert />
                  </Icon>
                  <Text fontSize="xs" fontWeight="semibold">Signed Out</Text>
                </HStack>
              )}
            </HStack>
          )}
        </VStack>
      </Card.Body>
    </Card.Root>
  )
}
