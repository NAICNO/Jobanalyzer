import { Edge, MarkerType, Node } from '@xyflow/react'
import { ProcessEntry } from '../types/ProcessEntry.ts'

export const generateProcessTree = (csv: string) => {
  const processes = parseCSV(csv)
  const pidToLabel = buildPidToLabel(processes)
  const uniqueEdges = generateEdges(processes, pidToLabel)

  // Generate React Flow nodes and edges first
  const nodes: Node[] = Array.from(pidToLabel.values()).map(label => ({
    id: label,
    position: {x: 0, y: 0},
    data: {label},
  }))
  const formattedEdges = formatEdges(uniqueEdges)

  return {nodes, edges: formattedEdges}
}

// Helper to parse CSV and deduplicate lines
export const parseCSV = (csv: string): ProcessEntry[] => {
  const lines = csv.trim().split('\n')
  const seen = new Set<string>()
  const processes: ProcessEntry[] = []

  for (const line of lines) {
    if (seen.has(line)) continue
    seen.add(line)
    const [pidStr, ppidStr, command] = line.split(',')
    processes.push({
      pid: parseInt(pidStr, 10),
      ppid: parseInt(ppidStr, 10),
      command: command.trim(),
    })
  }
  return processes
}

// Helper to build pid -> label map, including unknown parents
export const buildPidToLabel = (processes: ProcessEntry[]): Map<number, string> => {
  const pidToLabel = new Map<number, string>()
  processes.forEach(proc => {
    pidToLabel.set(proc.pid, `${proc.pid}:${proc.command}`)
  })
  // Add parent pids that are not in the process list
  processes.forEach(proc => {
    if (!pidToLabel.has(proc.ppid)) {
      pidToLabel.set(proc.ppid, `${proc.ppid}:unknown`)
    }
  })
  return pidToLabel
}

// Helper to generate unique edges as [sourceLabel, targetLabel] tuples
export const generateEdges = (processes: ProcessEntry[], pidToLabel: Map<number, string>): [string, string][] => {
  const edges: [string, string][] = []
  processes.forEach(proc => {
    const childLabel = pidToLabel.get(proc.pid)!
    const parentLabel = pidToLabel.get(proc.ppid)!
    edges.push([parentLabel, childLabel])
  })
  // Deduplicate edges by serializing to string and back
  return Array.from(new Set(edges.map(edge => edge.join('|')))).map(e => e.split('|') as [string, string])
}

// Helper to format edges ([source, target]) to Edge[]
export const formatEdges = (edges: [string, string][]): Edge[] => {
  return edges.map(([source, target]) => ({
    id: `${source}-${target}`,
    source,
    target,
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
}
