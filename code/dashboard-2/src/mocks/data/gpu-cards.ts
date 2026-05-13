/**
 * Mock GPU card data for demo mode.
 *
 * Each GPU node (gpu001-gpu008) is equipped with 4x NVIDIA A100-SXM4-80GB
 * accelerators. UUIDs are deterministic so mock data is reproducible.
 */

import { CLUSTER_NAME } from './cluster'

const GPU_NODE_COUNT = 8
const GPUS_PER_NODE = 4

/** Build a deterministic UUID for a GPU card. */
function gpuUuid(node: string, index: number): string {
  const padIdx = String(index).padStart(2, '0')
  return `GPU-${node}-${padIdx}-0000-0000-000000000000`
}

/** Create a single GpuCardResponse-shaped object (dates as ISO strings). */
function makeGpuCard(node: string, index: number) {
  return {
    time: new Date().toISOString(),
    uuid: gpuUuid(node, index),
    manufacturer: 'NVIDIA',
    model: 'NVIDIA A100-SXM4-80GB',
    architecture: 'Ampere',
    memory: 80 * 1024 * 1024, // 80 GB in KiB
    cluster: CLUSTER_NAME,
    node,
    index,
    address: `00000000:${(0x41 + index).toString(16).toUpperCase()}:00.0`,
    driver: '535.129.03',
    firmware: '96.00.5E.00.01',
    max_power_limit: 400,
    min_power_limit: 100,
    max_ce_clock: 1410,
    max_memory_clock: 1593,
    last_active: new Date().toISOString(),
  }
}

/**
 * Return the GPU cards installed in `node`.
 * Non-GPU nodes return an empty array.
 */
export function getGpuCardsForNode(node: string): ReturnType<typeof makeGpuCard>[] {
  if (!node.startsWith('gpu')) return []
  return Array.from({ length: GPUS_PER_NODE }, (_, i) => makeGpuCard(node, i))
}

/**
 * Pre-computed lookup: node name -> list of GPU UUIDs.
 * Useful when generating sample-level data that references cards by UUID.
 */
export const ALL_GPU_UUIDS: Record<string, string[]> = {}

for (let i = 1; i <= GPU_NODE_COUNT; i++) {
  const node = `gpu${String(i).padStart(3, '0')}`
  ALL_GPU_UUIDS[node] = Array.from({ length: GPUS_PER_NODE }, (_, j) =>
    gpuUuid(node, j),
  )
}
