import { VStack, Text, Heading, Listbox, Input, useFilter, Spinner, Alert, createListCollection, Box, Tabs } from '@chakra-ui/react'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'

import { getClusterByClusterNodesOptions } from '../../client/@tanstack/react-query.gen'
import { useClusterClient } from '../../hooks/useClusterClient'
import { NodeErrorMessages } from '../../components/v2/NodeErrorMessages'
import { NodeTopology } from '../../components/v2/NodeTopology'
import { NodeInfoSummary } from '../../components/v2/NodeInfoSummary'
import { NodeStates } from '../../components/v2/NodeStates'
import { NodeCpuTimeseries } from '../../components/v2/NodeCpuTimeseries'
import { ResizableColumns } from '../../components/v2/ResizableColumns'
import { NodeOverviewCards } from '../../components/v2/NodeOverviewCards'
import { NodesTable } from '../../components/v2/NodesTable'

export const NodesPage = () => {
  const { clusterName, nodename } = useParams()
  const navigate = useNavigate()
  const minLeftWidth = 220
  const maxLeftWidth = 640
  const handleWidth = 6

  const client = useClusterClient(clusterName)
  if (!client) {
    return <Spinner />
  }

  // Fetch nodes for the cluster from /cluster/:cluster/nodes
  const baseQueryOptions = getClusterByClusterNodesOptions({
    path: { cluster: clusterName ?? '' },
    client,
  })
  const { data, isLoading, isError, error } = useQuery({
    ...baseQueryOptions,
    enabled: !!clusterName,
  })

  const nodes = (data ?? []) as string[]
  type NodeItem = { value: string; label: string }
  const items = useMemo<NodeItem[]>(() => nodes.map(n => ({ value: n, label: n })), [nodes])

  const [selectedNodeValue, setSelectedNodeValue] = useState<string[]>([])
  const { contains } = useFilter({ sensitivity: 'base' })
  const [filterValue, setFilterValue] = useState('')
  const filteredItems = useMemo(() => items.filter((it) => contains(it.label, filterValue)), [items, filterValue, contains])
  const collection = useMemo(() => createListCollection<NodeItem>({ items: filteredItems }), [filteredItems])

  // Sync selection with URL param and initialize when data arrives
  useEffect(() => {
    if (items.length === 0) return

    // If URL contains a nodename, and it exists, sync selection to it
    if (nodename) {
      const exists = items.some(it => it.value === nodename)
      if (exists) {
        if (selectedNodeValue[0] !== nodename) {
          setSelectedNodeValue([nodename])
        }
        return
      }
    }

    // Otherwise, if nothing is selected, select first and update URL for deep-linking
    if (selectedNodeValue.length === 0) {
      const first = items[0].value
      setSelectedNodeValue([first])
      if (clusterName) {
        navigate(`/v2/${clusterName}/nodes/${first}`, { replace: true })
      }
    }
  }, [items, nodename, selectedNodeValue.length, clusterName, navigate])

  const selectedNode = items.find(node => node.value === selectedNodeValue[0])

  if (!clusterName) {
    return (
      <VStack p={4} align="start">
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>Missing cluster name in route.</Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  return (
    <>
      {clusterName ? (
        <VStack align="start" w="100%" p={4} pt={2} mb={4} gap={4}>
          <NodeOverviewCards cluster={clusterName} />
        </VStack>
      ) : null}

      <Tabs.Root defaultValue="nodes" variant="outline" w="100%">
        <Tabs.List px={4}>
          <Tabs.Trigger value="nodes">Node Browser</Tabs.Trigger>
          <Tabs.Trigger value="grid">Nodes Table</Tabs.Trigger>
        </Tabs.List>

        <Tabs.Content value="nodes">
          <ResizableColumns
            height="calc(100vh)"
            initialLeftWidth={320}
            minLeftWidth={minLeftWidth}
            maxLeftWidth={maxLeftWidth}
            handleWidth={handleWidth}
            storageKey="nodesPage.leftWidth"
            left={
              <VStack p={4} pt={0} gap={4} align="start">
                {isLoading && (
                  <Box display="flex" alignItems="center" gap={2}>
                    <Spinner size="sm" />
                    <Text>Loading nodes…</Text>
                  </Box>
                )}
                {isError && (
                  <Alert.Root status="error">
                    <Alert.Indicator />
                    <Alert.Description>
                      {error instanceof Error ? error.message : 'Failed to load nodes.'}
                    </Alert.Description>
                  </Alert.Root>
                )}
                {!isLoading && !isError && (
                  <Listbox.Root
                    collection={collection}
                    value={selectedNodeValue}
                    onValueChange={(details) => {
                      setSelectedNodeValue(details.value)
                      const sel = details.value[0]
                      if (sel && clusterName) {
                        navigate(`/v2/${clusterName}/nodes/${sel}`)
                      }
                    }}
                    width="100%"
                  >
                    <Listbox.Label>Available Nodes</Listbox.Label>
                    <Listbox.Input as={Input} placeholder="Type to filter nodes..." onChange={(e) => setFilterValue(e.target.value)} />
                    <Listbox.Content>
                      {collection.items.map((node) => (
                        <Listbox.Item item={node} key={node.value}>
                          <Listbox.ItemText>{node.label}</Listbox.ItemText>

                        </Listbox.Item>
                      ))}
                      {collection.items.length === 0 && (
                        <Text color="gray.500" p={2}>No nodes found.</Text>
                      )}
                    </Listbox.Content>
                  </Listbox.Root>
                )}
              </VStack>
            }
            right={
              <VStack flex={1} p={4} gap={4} align="start" minW="0">
                {selectedNode ? (
                  <>
                    <Heading size="md">{selectedNode.label} Details</Heading>
                    <NodeInfoSummary cluster={clusterName!} nodename={selectedNode.value} />
                    <NodeErrorMessages cluster={clusterName!} nodename={selectedNode.value} />
                    <NodeStates cluster={clusterName!} nodename={selectedNode.value} />
                    <NodeCpuTimeseries cluster={clusterName!} nodename={selectedNode.value} />
                    <NodeTopology cluster={clusterName!} nodename={selectedNode.value} />
                  </>
                ) : (
                  <Text>Select a node to see details</Text>
                )}
              </VStack>
            }
          />
        </Tabs.Content>

        <Tabs.Content value="grid">
          <VStack p={4} gap={4} align="start" w="100%">
            <Heading size="md">All Nodes</Heading>
            {isLoading && (
              <Box display="flex" alignItems="center" gap={2}>
                <Spinner size="sm" />
                <Text>Loading nodes…</Text>
              </Box>
            )}
            {isError && (
              <Alert.Root status="error">
                <Alert.Indicator />
                <Alert.Description>
                  {error instanceof Error ? error.message : 'Failed to load nodes.'}
                </Alert.Description>
              </Alert.Root>
            )}
            <NodesTable 
              clusterName={clusterName!} 
              onNodeClick={(nodeName) => navigate(`/v2/${clusterName}/nodes/${nodeName}`)}
            />
            <Text fontSize="sm" color="gray.600">
              Click on any row to view detailed information about that node.
            </Text>
          </VStack>
        </Tabs.Content>
      </Tabs.Root>
    </>
  )
}