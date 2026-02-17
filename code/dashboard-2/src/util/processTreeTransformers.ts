import type { Edge, Node } from '@xyflow/react'
import { MarkerType, Position } from '@xyflow/react'
import dagre from '@dagrejs/dagre'
import type { ProcessTreeResponse, ProcessData, SampleProcessAccResponse } from '../client/types.gen'

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

const toUnixSec = (time: Date): number => {
  return Math.floor(new Date(time).getTime() / 1000)
}

const findClosestSample = (
  samples: SampleProcessAccResponse[],
  targetTimeSec: number,
): SampleProcessAccResponse | null => {
  if (samples.length === 0) return null
  if (samples.length === 1) return samples[0]

  let closest = samples[0]
  let closestDiff = Math.abs(toUnixSec(samples[0].time) - targetTimeSec)

  for (let i = 1; i < samples.length; i++) {
    const diff = Math.abs(toUnixSec(samples[i].time) - targetTimeSec)
    if (diff < closestDiff) {
      closest = samples[i]
      closestDiff = diff
    } else {
      break
    }
  }

  return closest
}

export const processTreeToFlowAtTime = (
  treeResponse: ProcessTreeResponse,
  targetTimeSec: number,
): { nodes: Node<ProcessNodeData>[]; edges: Edge[] } => {
  const { processes, relations } = treeResponse

  const visiblePids = new Set<string>()
  const nodes: Node<ProcessNodeData>[] = []

  for (const [pid, proc] of Object.entries(processes)) {
    if (proc.data.length === 0) continue

    const firstTime = toUnixSec(proc.data[0].time)
    const lastTime = toUnixSec(proc.data[proc.data.length - 1].time)

    if (targetTimeSec < firstTime || targetTimeSec > lastTime) continue

    visiblePids.add(pid)

    const closestSample = findClosestSample(proc.data, targetTimeSec)

    nodes.push({
      id: pid,
      type: 'processNode',
      position: { x: 0, y: 0 },
      data: {
        label: `${pid}: ${proc.cmd}`,
        pid,
        cmd: proc.cmd,
        user: proc.user,
        cpuUtil: closestSample?.cpu_util ?? null,
        memoryResident: closestSample?.memory_resident ?? null,
        process: proc,
      },
    })
  }

  const edges: Edge[] = relations
    .filter((rel) => visiblePids.has(String(rel.source)) && visiblePids.has(String(rel.target)))
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

export const collectUniqueSampleTimes = (
  treeResponse: ProcessTreeResponse,
): number[] => {
  const timeSet = new Set<number>()
  for (const proc of Object.values(treeResponse.processes)) {
    for (const sample of proc.data) {
      timeSet.add(toUnixSec(sample.time))
    }
  }
  return Array.from(timeSet).sort((a, b) => a - b)
}
