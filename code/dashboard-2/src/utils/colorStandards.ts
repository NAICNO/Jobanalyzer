/**
 * Standardized color palette for cluster metrics
 * Use these constants and functions to ensure consistent color usage across the dashboard
 */

export const ClusterColors = {
  // State colors
  STATE_HEALTHY: 'green',
  STATE_WARNING: 'yellow',
  STATE_CRITICAL: 'red',
  STATE_INFO: 'blue',
  STATE_NEUTRAL: 'gray',

  // Utilization thresholds
  UTIL_LOW: 'blue',      // <40%
  UTIL_MODERATE: 'green', // 40-70%
  UTIL_HIGH: 'yellow',    // 70-90%
  UTIL_CRITICAL: 'red',   // >90%

  // Job states
  JOB_RUNNING: 'blue',
  JOB_PENDING: 'yellow',
  JOB_COMPLETED: 'green',
  JOB_FAILED: 'red',
  JOB_CANCELLED: 'gray',

  // Node states
  NODE_IDLE: 'green',
  NODE_ALLOCATED: 'blue',
  NODE_MIXED: 'blue',
  NODE_DOWN: 'red',
  NODE_DRAIN: 'red',
  NODE_MAINTENANCE: 'yellow',
} as const

/**
 * Get color for utilization percentage
 * @param percent - Utilization percentage (0-100)
 * @returns Color palette name
 */
export function getUtilizationColor(percent: number): string {
  if (percent > 90) return ClusterColors.UTIL_CRITICAL
  if (percent > 70) return ClusterColors.UTIL_HIGH
  if (percent > 40) return ClusterColors.UTIL_MODERATE
  return ClusterColors.UTIL_LOW
}

/**
 * Get color for success rate percentage
 * @param percent - Success rate percentage (0-100)
 * @returns Color palette name
 */
export function getSuccessRateColor(percent: number): string {
  if (percent >= 80) return ClusterColors.STATE_HEALTHY
  if (percent >= 50) return ClusterColors.STATE_WARNING
  return ClusterColors.STATE_CRITICAL
}

/**
 * Get color for node state
 * @param state - Node state string (e.g., "IDLE", "ALLOCATED", "DOWN")
 * @returns Color palette name
 */
export function getNodeStateColor(state: string): string {
  const s = state.toUpperCase()
  if (s.includes('DOWN') || s.includes('DRAIN') || s.includes('FAIL') || s.includes('INVALID')) {
    return ClusterColors.NODE_DOWN
  }
  if (s.includes('IDLE')) {
    return ClusterColors.NODE_IDLE
  }
  if (s.includes('ALLOC') || s.includes('MIX')) {
    return ClusterColors.NODE_ALLOCATED
  }
  if (s.includes('COMPLETING') || s.includes('PLANNED') || s.includes('RESERVED') || s.includes('MAINTENANCE')) {
    return ClusterColors.NODE_MAINTENANCE
  }
  return ClusterColors.STATE_NEUTRAL
}
