import { useCallback, useEffect } from 'react'
import { addEdge, Background, Controls, MiniMap, Panel, ReactFlow, useEdgesState, useNodesState, } from '@xyflow/react'
import type { Edge, Node } from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import dagre from '@dagrejs/dagre'
import { MdOutlineAlignHorizontalLeft, MdOutlineAlignVerticalTop } from 'react-icons/md'
import { Box, Heading, SegmentGroup, VStack } from '@chakra-ui/react'

import { useColorMode } from '../components/ui/color-mode.tsx'
import { useFetchJobProcessTree } from '../hooks/useFetchJobProcessTree.ts'
import { useSearchParams } from 'react-router'
import { JobBasicInfoTable } from '../components/table/JobBasicInfoTable.tsx'
import { PageTitle } from '../components'

const dagreGraph = new dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}))

const nodeWidth = 150
const nodeHeight = 150

const getLayoutedElements = (
  nodes: Node[],
  edges: Edge[],
  direction: 'TB' | 'LR' = 'TB'
): { nodes: Node[]; edges: Edge[] } => {
  const isHorizontal = direction === 'LR'
  dagreGraph.setGraph({rankdir: direction})

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, {width: nodeWidth, height: nodeHeight})
  })

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target)
  })

  dagre.layout(dagreGraph)

  const newNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id)
    return {
      ...node,
      targetPosition: isHorizontal ? 'left' : 'top',
      sourcePosition: isHorizontal ? 'right' : 'bottom',
      // We are shifting the dagre node position (anchor=center center) to the top left
      // so it matches the React Flow node anchor point (top left).
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    }
  })

  return {nodes: newNodes, edges}
}

export default function JobProcessTreePage() {

  const [searchParams] = useSearchParams()

  const clusterName = searchParams.get('clusterName')
  const hostname = searchParams.get('hostname')
  const user = searchParams.get('user')
  const jobId = searchParams.get('jobId')

  const {data: jobProcessTree} = useFetchJobProcessTree(clusterName, jobId)

  const {nodes: fetchedNodes, edges: fetchedEdges} = jobProcessTree || {nodes: [], edges: []}

  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])

  useEffect(() => {
    if (fetchedNodes.length > 0 || fetchedEdges.length > 0) {
      const {nodes: layoutedNodes, edges: layoutedEdges} = getLayoutedElements(fetchedNodes, fetchedEdges)
      setNodes(layoutedNodes)
      setEdges(layoutedEdges)
    }
  }, [fetchedNodes, fetchedEdges])

  const onConnect = useCallback(
    (params) => setEdges((eds) => addEdge(params, eds)),
    [setEdges],
  )

  const onLayout = useCallback(
    (direction) => {
      const {nodes: layoutedNodes, edges: layoutedEdges} = getLayoutedElements(
        nodes,
        edges,
        direction,
      )

      setNodes([...layoutedNodes])
      setEdges([...layoutedEdges])
    },
    [nodes, edges],
  )

  const {colorMode} = useColorMode()

  return (
    <>
      <PageTitle title={`Job Tree - ${clusterName} ${jobId}`}/>
      <Box style={{width: '100%', height: '100%', display: 'flex', flexDirection: 'column'}}>
        <Box pt={2} pb={4} maxWidth="400px">
          <VStack gap={2} alignItems="start">
            <Heading as="h3" ml={2} size={{base: 'md', md: 'lg'}}>
              Job Process Tree
            </Heading>
            <JobBasicInfoTable jobId={jobId} user={user} clusterName={clusterName} hostname={hostname}/>
          </VStack>
        </Box>
        <Box flex="1" minHeight={0}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            colorMode={colorMode}
          >
            <Controls/>
            <Panel position="top-right">
              <SegmentGroup.Root defaultValue="TB" onValueChange={e => onLayout(e.value)}>
                <SegmentGroup.Indicator/>
                <SegmentGroup.Items
                  items={[
                    {
                      value: 'TB',
                      label: (
                        <MdOutlineAlignVerticalTop/>
                      ),
                    },
                    {
                      value: 'LR',
                      label: (
                        <MdOutlineAlignHorizontalLeft/>
                      ),
                    },
                  ]}
                />
              </SegmentGroup.Root>
            </Panel>
            <MiniMap/>
            <Background variant="dots" gap={12} size={1}/>
          </ReactFlow>
        </Box>
      </Box>
    </>
  )
}
