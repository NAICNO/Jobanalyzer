/**
 * Mock node data for demo mode.
 *
 * Provides node hardware info, Slurm states, and last-probe timestamps
 * for all 30 nodes in the demo cluster.
 *
 * All date values are ISO 8601 strings — MSW returns raw JSON and the
 * app's transformers handle string-to-Date conversion.
 */

import { CLUSTER_NAME, NODE_NAMES } from './cluster'
import { getGpuCardsForNode } from './gpu-cards'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const KiB = 1
const GiB = 1024 * 1024 * KiB // 1 GiB in KiB
const TiB = 1024 * GiB // 1 TiB in KiB

/** Deterministic pseudo-random seeded by a string, returns 0..1. */
function seededRandom(seed: string): number {
  let hash = 0
  for (let i = 0; i < seed.length; i++) {
    hash = (hash << 5) - hash + seed.charCodeAt(i)
    hash |= 0
  }
  return Math.abs(hash % 1000) / 1000
}

/** Return an ISO timestamp `offsetMs` milliseconds before now. */
function ago(offsetMs: number): string {
  return new Date(Date.now() - offsetMs).toISOString()
}

// ---------------------------------------------------------------------------
// Node hardware profiles
// ---------------------------------------------------------------------------

interface NodeProfile {
  sockets: number
  cores_per_socket: number
  threads_per_core: number
  cpu_model: string
  memory: number // in KiB
  partitions: string[]
}

const COMPUTE_PROFILE: NodeProfile = {
  sockets: 2,
  cores_per_socket: 64,
  threads_per_core: 2,
  cpu_model: 'AMD EPYC 7763 64-Core Processor',
  memory: 512 * GiB,
  partitions: ['cpu'],
}

const GPU_PROFILE: NodeProfile = {
  sockets: 2,
  cores_per_socket: 32,
  threads_per_core: 2,
  cpu_model: 'Intel(R) Xeon(R) Platinum 8358 CPU @ 2.60GHz',
  memory: 1 * TiB,
  partitions: ['gpu'],
}

const HIGHMEM_PROFILE: NodeProfile = {
  sockets: 4,
  cores_per_socket: 28,
  threads_per_core: 2,
  cpu_model: 'Intel(R) Xeon(R) Platinum 8280 CPU @ 2.70GHz',
  memory: 2 * TiB,
  partitions: ['highmem'],
}

function profileFor(node: string): NodeProfile {
  if (node.startsWith('gpu')) return GPU_PROFILE
  if (node.startsWith('hi')) return HIGHMEM_PROFILE
  return COMPUTE_PROFILE
}

// Nodes that also belong to the "interactive" partition
const INTERACTIVE_NODES = new Set(['cn001', 'cn002', 'gpu001', 'hi001'])

// Nodes that are down/drained
const DOWN_NODES = new Set(['cn018', 'cn019'])

// ---------------------------------------------------------------------------
// NodeInfoResponse (keyed by node name)
// ---------------------------------------------------------------------------

function makeNodeInfo(node: string) {
  const profile = profileFor(node)
  const partitions = [...profile.partitions]
  if (INTERACTIVE_NODES.has(node)) {
    partitions.push('interactive')
  }

  const totalCpus =
    profile.sockets * profile.cores_per_socket * profile.threads_per_core

  // Build realistic alloc_tres based on node state
  const isDown = DOWN_NODES.has(node)
  const rng = seededRandom(node)

  const alloc_tres = buildAllocTres(node, totalCpus, profile.memory, rng, isDown)

  return {
    time: new Date().toISOString(),
    cluster: CLUSTER_NAME,
    node,
    os_name: 'Linux',
    os_release: '5.15.0-91-generic',
    architecture: 'x86_64',
    sockets: profile.sockets,
    cores_per_socket: profile.cores_per_socket,
    threads_per_core: profile.threads_per_core,
    cpu_model: profile.cpu_model,
    memory: profile.memory,
    topo_svg: null,
    topo_text: null,
    cards: getGpuCardsForNode(node),
    partitions,
    alloc_tres,
  }
}

function buildAllocTres(
  node: string,
  totalCpus: number,
  totalMemory: number,
  rng: number,
  isDown: boolean,
) {
  if (isDown) {
    return { cpu: 0, memory: 0 }
  }

  if (node.startsWith('gpu')) {
    // GPU nodes are heavily utilised
    const gpuAlloc = rng > 0.3 ? 4 : 3
    const cpuFraction = 0.6 + rng * 0.35 // 60-95 %
    const memFraction = 0.5 + rng * 0.4 // 50-90 %
    return {
      cpu: Math.round(totalCpus * cpuFraction),
      memory: Math.round(totalMemory * memFraction),
      gpu: gpuAlloc,
      node: 1,
      billing: Math.round(totalCpus * cpuFraction),
    }
  }

  if (node.startsWith('hi')) {
    // High-memory nodes: moderate CPU, high memory
    const cpuFraction = 0.3 + rng * 0.4
    const memFraction = 0.6 + rng * 0.3
    return {
      cpu: Math.round(totalCpus * cpuFraction),
      memory: Math.round(totalMemory * memFraction),
      node: 1,
      billing: Math.round(totalCpus * cpuFraction),
    }
  }

  // Compute nodes: varied utilisation
  const cpuFraction = rng * 0.95
  const memFraction = rng * 0.85
  return {
    cpu: Math.round(totalCpus * cpuFraction),
    memory: Math.round(totalMemory * memFraction),
    node: cpuFraction > 0.1 ? 1 : 0,
    billing: Math.round(totalCpus * cpuFraction),
  }
}

// ---------------------------------------------------------------------------
// NodeStateResponse (array of { time, cluster, node, states })
// ---------------------------------------------------------------------------

function statesForNode(node: string): string[] {
  if (DOWN_NODES.has(node)) return ['down', 'drain']

  if (node.startsWith('gpu')) {
    // GPU nodes mostly allocated
    const rng = seededRandom(node + '-state')
    if (rng < 0.15) return ['mixed']
    return ['alloc']
  }

  if (node.startsWith('hi')) {
    const rng = seededRandom(node + '-state')
    return rng < 0.5 ? ['alloc'] : ['mixed']
  }

  // Compute nodes: mix of states
  const rng = seededRandom(node + '-state')
  if (rng < 0.3) return ['idle']
  if (rng < 0.7) return ['alloc']
  return ['mixed']
}

function makeNodeState(node: string) {
  return {
    time: new Date().toISOString(),
    cluster: CLUSTER_NAME,
    node,
    states: statesForNode(node),
  }
}

// ---------------------------------------------------------------------------
// Last-probe timestamps ({ [nodeName]: isoString | null })
// ---------------------------------------------------------------------------

function lastProbeFor(node: string): string | null {
  if (DOWN_NODES.has(node)) return null

  // Recent probe within last 5 minutes, varied per node
  const offsetMs = Math.round(seededRandom(node + '-probe') * 5 * 60 * 1000)
  return ago(offsetMs)
}

// ---------------------------------------------------------------------------
// Exported data & accessors
// ---------------------------------------------------------------------------

/** All node info keyed by hostname. Shape: Record<string, NodeInfoResponse>. */
export const mockNodesInfo: Record<string, ReturnType<typeof makeNodeInfo>> = {}

/** All node states as an array. Shape: NodeStateResponse[]. */
export const mockNodeStates: ReturnType<typeof makeNodeState>[] = []

/** Last probe timestamps keyed by hostname. Shape: Record<string, string | null>. */
export const mockLastProbeTimestamps: Record<string, string | null> = {}

for (const node of NODE_NAMES) {
  mockNodesInfo[node] = makeNodeInfo(node)
  mockNodeStates.push(makeNodeState(node))
  mockLastProbeTimestamps[node] = lastProbeFor(node)
}

/** Look up a single node's info. Returns undefined for unknown nodes. */
export function getNodeInfo(
  nodeName: string,
): ReturnType<typeof makeNodeInfo> | undefined {
  return mockNodesInfo[nodeName]
}

/** Look up a single node's state entry. Returns undefined for unknown nodes. */
export function getNodeStates(
  nodeName: string,
): ReturnType<typeof makeNodeState> | undefined {
  return mockNodeStates.find((s) => s.node === nodeName)
}
