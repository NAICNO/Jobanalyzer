import { Accordion, VStack, Text, HStack, Tag } from '@chakra-ui/react'

import type { PartitionResponseOutput } from '../../client'

interface Props {
  partition: PartitionResponseOutput
}

export const PartitionGpus = ({partition}: Props) => {
  const gpusInUse = partition.gpus_in_use ?? []
  const totalGpus = partition.total_gpus ?? 0

  const handleCopyUuid = (uuid: string) => {
    navigator.clipboard.writeText(uuid).catch(() => {
      // Silently ignore clipboard errors
    })
  }

  if (totalGpus === 0) {
    return null // Don't show section if no GPUs
  }

  return (
    <Accordion.Root variant="outline">
      <Accordion.Item value="gpus">
        <Accordion.ItemTrigger px={3} py={2} _hover={{bg: 'gray.50'}}>
          <HStack justify="space-between" flex={1}>
            <Text fontWeight="semibold">GPUs in Use</Text>
            <Text fontSize="sm" color="gray.600">{gpusInUse.length} / {totalGpus} active</Text>
          </HStack>
          <Accordion.ItemIndicator/>
        </Accordion.ItemTrigger>
        <Accordion.ItemContent>
          <Accordion.ItemBody>
            <VStack align="start" gap={3} w="100%">
              {gpusInUse.length > 0 ? (
                <>
                  <Text fontSize="sm" color="gray.600">
                    Click a UUID to copy to clipboard
                  </Text>
                  <HStack gap={2} flexWrap="wrap">
                    {gpusInUse.map((uuid: string) => (
                      <Tag.Root
                        key={uuid}
                        size="sm"
                        cursor="pointer"
                        onClick={() => handleCopyUuid(uuid)}
                        _hover={{bg: 'gray.100'}}
                        title="Click to copy"
                      >
                        <Tag.Label fontFamily="mono" fontSize="xs">{uuid}</Tag.Label>
                      </Tag.Root>
                    ))}
                  </HStack>
                </>
              ) : (
                <Text color="gray.500" fontSize="sm">No GPUs currently in use</Text>
              )}
            </VStack>
          </Accordion.ItemBody>
        </Accordion.ItemContent>
      </Accordion.Item>
    </Accordion.Root>
  )
}
