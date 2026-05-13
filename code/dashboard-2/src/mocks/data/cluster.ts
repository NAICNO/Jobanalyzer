/**
 * Mock cluster configuration for demo mode.
 *
 * Defines the cluster name, node inventory, and partition layout used
 * by all other mock data files.
 */

export const CLUSTER_NAME = 'demo.hpc.example.org'

/** All node hostnames across the cluster. */
export const NODE_NAMES = [
  // 20 compute nodes
  ...Array.from({ length: 20 }, (_, i) => `cn${String(i + 1).padStart(3, '0')}`),
  // 8 GPU nodes
  ...Array.from({ length: 8 }, (_, i) => `gpu${String(i + 1).padStart(3, '0')}`),
  // 2 high-memory nodes
  'hi001',
  'hi002',
]

export const PARTITION_NAMES = ['gpu', 'cpu', 'interactive', 'highmem']

/** Shape matches ClusterResponse (dates as ISO strings for JSON transport). */
export const mockClusterResponse = {
  time: new Date().toISOString(),
  cluster: CLUSTER_NAME,
  slurm: 1,
  partitions: PARTITION_NAMES,
  nodes: NODE_NAMES,
}
