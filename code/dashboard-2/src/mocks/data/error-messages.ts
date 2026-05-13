/**
 * Mock error message data for demo mode.
 *
 * ErrorMessageResponse has `time` as a plain string (not a Date object),
 * which aligns with the generated type definition.
 */

import type { ErrorMessageResponse } from '../../client/types.gen'
import { CLUSTER_NAME } from './cluster'

// ---------------------------------------------------------------------------
// Time helpers
// ---------------------------------------------------------------------------

const NOW = new Date()

function hoursAgo(h: number): string {
  return new Date(NOW.getTime() - h * 3600_000).toISOString()
}

function minutesAgo(m: number): string {
  return new Date(NOW.getTime() - m * 60_000).toISOString()
}

// ---------------------------------------------------------------------------
// Error messages
// ---------------------------------------------------------------------------

export const mockErrorMessages: ErrorMessageResponse[] = [
  {
    cluster: CLUSTER_NAME,
    node: 'gpu005',
    details:
      'GPU 2 (UUID: GPU-gpu005-02-0000-0000-000000000000) temperature exceeded 90\u00B0C, throttling applied',
    time: minutesAgo(30),
  },
  {
    cluster: CLUSTER_NAME,
    node: 'cn018',
    details: 'Node unreachable - network timeout after 30s',
    time: hoursAgo(2),
  },
  {
    cluster: CLUSTER_NAME,
    node: 'cn019',
    details:
      'DRAM ECC error detected on DIMM_A2, node set to drain state',
    time: hoursAgo(5),
  },
  {
    cluster: CLUSTER_NAME,
    node: 'cn012',
    details:
      'InfiniBand link flapping on mlx5_0 port 1 - 3 transitions in last 60s',
    time: hoursAgo(8),
  },
]

// ---------------------------------------------------------------------------
// Exports
// ---------------------------------------------------------------------------

/**
 * Filter error messages for a specific node.
 * Returns all errors if nodeName is undefined or empty.
 */
export function mockNodeErrorMessages(
  nodeName?: string,
): ErrorMessageResponse[] {
  if (!nodeName) return mockErrorMessages
  return mockErrorMessages.filter((e) => e.node === nodeName)
}
