import { useCallback, useEffect, useMemo, useState, memo } from 'react'
import {
  Background,
  BackgroundVariant,
  Controls,
  Handle,
  MiniMap,
  Panel,
  Position,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useReactFlow,
} from '@xyflow/react'
import type { Node, NodeProps } from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import {
  Box,
  VStack,
  HStack,
  Text,
  Spinner,
  Alert,
  Badge,
  SegmentGroup,
  SimpleGrid,
  Stat,
} from '@chakra-ui/react'
import { MdOutlineAlignHorizontalLeft, MdOutlineAlignVerticalTop } from 'react-icons/md'
import type { Client } from '../../client/client/types.gen'
import { useColorMode } from '../ui/color-mode'
import { useJobProcessTree } from '../../hooks/useJobProcessTree'
import {
  processTreeToFlow,
  getLayoutedElements,
  type ProcessNodeData,
} from '../../util/processTreeTransformers'

interface Props {
  cluster: string
  jobId: number
  client: Client | null
}

const formatKiB = (kib: number): string => {
  if (kib >= 1048576) return `${(kib / 1048576).toFixed(1)} GiB`
  if (kib >= 1024) return `${(kib / 1024).toFixed(1)} MiB`
  return `${kib.toFixed(0)} KiB`
}

const ProcessNode = memo(({ data }: NodeProps<Node<ProcessNodeData>>) => {
  return (
    <>
      <Handle type="target" position={Position.Top} />
      <VStack gap={0} align="start" p={2} minW="180px">
        <Text fontSize="xs" fontWeight="bold" truncate maxW="170px">
          {data.cmd}
        </Text>
        <Text fontSize="xs" color="fg.muted">
          PID: {data.pid}
        </Text>
        <HStack gap={2} mt={1}>
          {data.cpuUtil !== null && (
            <Badge size="xs" colorPalette="blue" variant="subtle">
              CPU {data.cpuUtil.toFixed(1)}%
            </Badge>
          )}
          {data.memoryResident !== null && (
            <Badge size="xs" colorPalette="green" variant="subtle">
              {formatKiB(data.memoryResident)}
            </Badge>
          )}
        </HStack>
      </VStack>
      <Handle type="source" position={Position.Bottom} />
    </>
  )
})

ProcessNode.displayName = 'ProcessNode'

const nodeTypes = { processNode: ProcessNode }

const ProcessTreeFlow = memo(({ cluster, jobId, client }: Props) => {
  const [direction, setDirection] = useState<'TB' | 'LR'>('TB')
  const { colorMode } = useColorMode()
  const { fitView } = useReactFlow()

  const { data: treeData, isLoading, isError, error } = useJobProcessTree({
    cluster,
    jobId,
    client,
  })

  const nodeNames = useMemo(() => {
    if (!treeData?.nodes) return []
    return Object.keys(treeData.nodes)
  }, [treeData])

  const [selectedNode, setSelectedNode] = useState<string | null>(null)
  const activeNode = selectedNode ?? nodeNames[0] ?? null

  const { flowNodes, flowEdges, metadata } = useMemo(() => {
    if (!treeData?.nodes || !activeNode || !treeData.nodes[activeNode]) {
      return { flowNodes: [], flowEdges: [], metadata: null }
    }
    const nodeData = treeData.nodes[activeNode]
    const { nodes, edges } = processTreeToFlow(nodeData)
    const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(nodes, edges, direction)
    return { flowNodes: layoutedNodes, flowEdges: layoutedEdges, metadata: nodeData.metadata }
  }, [treeData, activeNode, direction])

  const [nodes, setNodes, onNodesChange] = useNodesState(flowNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(flowEdges)

  // Sync when flow data changes and fit view
  useEffect(() => {
    setNodes(flowNodes)
    setEdges(flowEdges)
    // Allow React Flow to render the new nodes before fitting
    requestAnimationFrame(() => {
      fitView({ padding: 0.2 })
    })
  }, [flowNodes, flowEdges, setNodes, setEdges, fitView])

  const onLayout = useCallback(
    (dir: string) => {
      const d = dir as 'TB' | 'LR'
      setDirection(d)
    },
    [],
  )

  if (isLoading) {
    return (
      <Box w="100%" h="400px" display="flex" alignItems="center" justifyContent="center">
        <Spinner size="lg" />
      </Box>
    )
  }

  if (isError) {
    return (
      <Alert.Root status="error">
        <Alert.Indicator />
        <Alert.Description>
          Failed to load process tree: {error?.message || 'Unknown error'}
        </Alert.Description>
      </Alert.Root>
    )
  }

  if (!treeData || nodeNames.length === 0) {
    return (
      <Alert.Root status="info">
        <Alert.Indicator />
        <Alert.Description>No process tree data available for this job.</Alert.Description>
      </Alert.Root>
    )
  }

  return (
    <VStack w="100%" align="start" gap={4}>
      {/* Metadata summary */}
      {metadata && (
        <SimpleGrid columns={{ base: 2, md: 4 }} gap={3} w="100%">
          <Box borderWidth="1px" borderColor="border" rounded="md" p={3}>
            <Stat.Root size="sm">
              <Stat.Label fontSize="sm">Total Processes</Stat.Label>
              <Stat.ValueText>{metadata.total_processes}</Stat.ValueText>
            </Stat.Root>
          </Box>
          <Box borderWidth="1px" borderColor="border" rounded="md" p={3}>
            <Stat.Root size="sm">
              <Stat.Label fontSize="sm">Max Depth</Stat.Label>
              <Stat.ValueText>{metadata.max_depth}</Stat.ValueText>
            </Stat.Root>
          </Box>
          <Box borderWidth="1px" borderColor="border" rounded="md" p={3}>
            <Stat.Root size="sm">
              <Stat.Label fontSize="sm">Root PID</Stat.Label>
              <Stat.ValueText>{metadata.root_pid}</Stat.ValueText>
            </Stat.Root>
          </Box>
          {nodeNames.length > 1 && (
            <Box borderWidth="1px" borderColor="border" rounded="md" p={3}>
              <Stat.Root size="sm">
                <Stat.Label fontSize="sm">Node</Stat.Label>
                <Stat.ValueText fontSize="sm">
                  <SegmentGroup.Root
                    size="sm"
                    value={activeNode ?? ''}
                    onValueChange={(e) => setSelectedNode(e.value)}
                  >
                    <SegmentGroup.Indicator />
                    <SegmentGroup.Items
                      items={nodeNames.map((name) => ({ value: name, label: name }))}
                    />
                  </SegmentGroup.Root>
                </Stat.ValueText>
              </Stat.Root>
            </Box>
          )}
        </SimpleGrid>
      )}

      {/* React Flow tree */}
      <Box w="100%" h="600px" borderWidth="1px" borderColor="border" rounded="md">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          nodeTypes={nodeTypes}
          colorMode={colorMode}
          fitView
          nodesConnectable={false}
          nodesDraggable={true}
          elementsSelectable={true}
        >
          <Controls />
          <Panel position="top-right">
            <SegmentGroup.Root defaultValue="TB" onValueChange={(e) => e.value && onLayout(e.value)}>
              <SegmentGroup.Indicator />
              <SegmentGroup.Items
                items={[
                  {
                    value: 'TB',
                    label: <MdOutlineAlignVerticalTop />,
                  },
                  {
                    value: 'LR',
                    label: <MdOutlineAlignHorizontalLeft />,
                  },
                ]}
              />
            </SegmentGroup.Root>
          </Panel>
          <MiniMap />
          <Background variant={BackgroundVariant.Dots} gap={12} size={1} />
        </ReactFlow>
      </Box>
    </VStack>
  )
})

ProcessTreeFlow.displayName = 'ProcessTreeFlow'

export const ProcessTreeTab = memo(({ cluster, jobId, client }: Props) => {
  return (
    <ReactFlowProvider>
      <ProcessTreeFlow cluster={cluster} jobId={jobId} client={client} />
    </ReactFlowProvider>
  )
})

ProcessTreeTab.displayName = 'ProcessTreeTab'
