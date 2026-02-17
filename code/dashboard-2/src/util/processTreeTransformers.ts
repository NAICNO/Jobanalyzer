import type { Edge, Node } from '@xyflow/react'
import { MarkerType, Position } from '@xyflow/react'
import dagre from '@dagrejs/dagre'
import type { ProcessTreeResponse, ProcessData } from '../client/types.gen'

export interface ProcessNodeData extends Record<string, unknown> {
  label: string
  pid: string
  cmd: string
  user: string
  cpuUtil: number | null
  memoryResident: number | null
  process: ProcessData
}

const NODE_WIDTH = 200
const NODE_HEIGHT = 80

export const processTreeToFlow = (
  treeResponse: ProcessTreeResponse,
): { nodes: Node<ProcessNodeData>[]; edges: Edge[] } => {
  const { processes, relations } = treeResponse

  const nodes: Node<ProcessNodeData>[] = Object.entries(processes).map(([pid, proc]) => {
    const latestSample = proc.data.length > 0 ? proc.data[proc.data.length - 1] : null

    return {
      id: pid,
      type: 'processNode',
      position: { x: 0, y: 0 },
      data: {
        label: `${pid}: ${proc.cmd}`,
        pid,
        cmd: proc.cmd,
        user: proc.user,
        cpuUtil: latestSample?.cpu_util ?? null,
        memoryResident: latestSample?.memory_resident ?? null,
        process: proc,
      },
    }
  })

  const processIds = new Set(Object.keys(processes))

  const edges: Edge[] = relations
    .filter((rel) => processIds.has(String(rel.source)) && processIds.has(String(rel.target)))
    .map((rel) => ({
      id: rel.relation_id,
      source: String(rel.source),
      target: String(rel.target),
      type: 'default',
      markerEnd: {
        type: MarkerType.ArrowClosed,
        width: 15,
        height: 15,
      },
      style: {
        strokeWidth: 1.5,
      },
    }))

  return { nodes, edges }
}

export const getLayoutedElements = (
  nodes: Node[],
  edges: Edge[],
  direction: 'TB' | 'LR' = 'TB',
): { nodes: Node[]; edges: Edge[] } => {
  const graph = new dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}))
  const isHorizontal = direction === 'LR'
  graph.setGraph({ rankdir: direction })

  nodes.forEach((node) => {
    graph.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT })
  })

  edges.forEach((edge) => {
    graph.setEdge(edge.source, edge.target)
  })

  dagre.layout(graph)

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = graph.node(node.id)
    return {
      ...node,
      targetPosition: isHorizontal ? Position.Left : Position.Top,
      sourcePosition: isHorizontal ? Position.Right : Position.Bottom,
      position: {
        x: nodeWithPosition.x - NODE_WIDTH / 2,
        y: nodeWithPosition.y - NODE_HEIGHT / 2,
      },
    }
  })

  return { nodes: layoutedNodes, edges }
}
