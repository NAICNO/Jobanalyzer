import { Accordion, VStack, Text, Box, HStack, Tag } from '@chakra-ui/react'
import { Link as RouterLink } from 'react-router'

import type { PartitionResponseOutput } from '../../client'

interface Props {
  partition: PartitionResponseOutput
}

export const PartitionNodes = ({ partition }: Props) => {
  const nodesCount = partition.nodes?.length ?? 0
  const nodesCompactArray = partition.nodes_compact ?? []
  const nodesCompact = Array.isArray(nodesCompactArray) && nodesCompactArray.length > 0
    ? nodesCompactArray.join(', ')
    : ''
  const nodes = partition.nodes ?? []

  return (
    <Accordion.Root variant="outline">
      <Accordion.Item value="nodes">
        <Accordion.ItemTrigger px={3} py={2} _hover={{ bg: 'gray.50' }}>
          <HStack justify="space-between" flex={1}>
            <Text fontWeight="semibold">Nodes</Text>
            <Text fontSize="sm" color="gray.600">{nodesCount} nodes</Text>
          </HStack>
          <Accordion.ItemIndicator />
        </Accordion.ItemTrigger>
        <Accordion.ItemContent>
          <Accordion.ItemBody>
            <VStack align="start" gap={3} w="100%">
              {nodesCompact && (
                <Box>
                  <Text fontSize="sm" fontWeight="medium" mb={1}>Compact Representation</Text>
                  <Tag.Root>
                    <Tag.Label fontFamily="mono" fontSize="sm">{nodesCompact}</Tag.Label>
                  </Tag.Root>
                </Box>
              )}

              {nodes.length > 0 && (
                <Box>
                  <Text fontSize="sm" fontWeight="medium" mb={2}>All Nodes</Text>
                  <HStack gap={2} flexWrap="wrap">
                    {nodes.map((node: string) => (
                      <Tag.Root key={node} size="sm" asChild>
                        <RouterLink to={`/v2/${partition.cluster}/nodes/${node}`}>
                          <Tag.Label>{node}</Tag.Label>
                        </RouterLink>
                      </Tag.Root>
                    ))}
                  </HStack>
                </Box>
              )}

              {nodes.length === 0 && (
                <Text color="gray.500" fontSize="sm">No nodes assigned to this partition</Text>
              )}

              {nodes.length > 0 && (
                <Box mt={2}>
                  <RouterLink to={`/v2/${partition.cluster}/nodes`} style={{ color: 'var(--chakra-colors-blue-500)', textDecoration: 'underline', fontSize: '0.875rem' }}>
                    View all nodes →
                  </RouterLink>
                </Box>
              )}
            </VStack>
          </Accordion.ItemBody>
        </Accordion.ItemContent>
      </Accordion.Item>
    </Accordion.Root>
  )
}
