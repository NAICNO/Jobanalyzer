import { useMemo, useState } from 'react'
import { VStack, Text, HStack, Spinner, Alert, Box, Accordion, IconButton } from '@chakra-ui/react'
import { LuExternalLink } from 'react-icons/lu'
import DOMPurify from 'dompurify'

import { useClusterClient } from '../../hooks/useClusterClient'
import { useNodeTopology } from '../../hooks/v2/useNodeQueries'

type Props = {
  cluster: string
  nodename: string
  initialCollapsed?: boolean
}

export const NodeTopology = ({ cluster, nodename, initialCollapsed = true }: Props) => {
  const client = useClusterClient(cluster)
  
  // Use Accordion in controlled mode so we can gate the fetch by expansion state
  const [value, setValue] = useState<string[]>(initialCollapsed ? [] : ['topology'])
  const expanded = value.includes('topology')
  
  const { data, isLoading, isError, error } = useNodeTopology({ cluster, nodename, client, enabled: expanded })

  const svg = useMemo<string | undefined>(() => {
    if (!data) return undefined
    let raw: string | undefined
    if (typeof data === 'string') {
      raw = data
    } else if (typeof data === 'object') {
      const obj = data as Record<string, unknown>
      raw = (obj.svg as string | undefined)
        || (obj.body as string | undefined)
        || (obj.data as string | undefined)
        || undefined
    }
    return raw ? DOMPurify.sanitize(raw, { USE_PROFILES: { svg: true } }) : undefined
  }, [data])

  return (
    <VStack align="start" w="100%" mt={4} gap={2}>
      <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md">
        <Accordion.Root
          value={value}
          onValueChange={(e) => setValue(e.value)}
          collapsible
          variant="plain"
        >
          <Accordion.Item value="topology">
            <Accordion.ItemTrigger px={3} py={2} _hover={{ bg: 'gray.50' }}>
              <HStack flex="1" justify="space-between" align="center">
                <Text fontWeight="semibold">Topology</Text>
                <IconButton
                  aria-label="Open topology in new tab"
                  size="sm"
                  variant="ghost"
                  onClick={(e) => {
                    e.stopPropagation()
                    const url = `/v2/${cluster}/nodes/${nodename}/topology`
                    window.open(url, '_blank')
                  }}
                >
                  <LuExternalLink />
                </IconButton>
              </HStack>
              <Accordion.ItemIndicator />
            </Accordion.ItemTrigger>
            <Accordion.ItemContent>
              <Accordion.ItemBody>
                {isLoading && (
                  <HStack>
                    <Spinner size="sm" />
                    <Text>Loading topology…</Text>
                  </HStack>
                )}
                {isError && (
                  <Alert.Root status="error">
                    <Alert.Indicator />
                    <Alert.Description>
                      {error instanceof Error ? error.message : 'Failed to load topology.'}
                    </Alert.Description>
                  </Alert.Root>
                )}
                {!isLoading && !isError && svg && (
                  <Box p={2} w="100%" overflow="auto"
                    css={{
                      '& svg': { overflow: 'visible' },
                      '& svg text': { fontSize: '10px' },
                    }}
                    dangerouslySetInnerHTML={{ __html: svg }}
                  />
                )}
                {!isLoading && !isError && !svg && (
                  <Text color="gray.600">No topology available.</Text>
                )}
              </Accordion.ItemBody>
            </Accordion.ItemContent>
          </Accordion.Item>
        </Accordion.Root>
      </Box>
    </VStack>
  )
}
