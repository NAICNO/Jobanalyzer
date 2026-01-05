/**
 * Utilities for calculating utilization colors based on percentage thresholds
 */

export interface UtilizationThresholds {
  high: number   // >= high: red
  medium: number // >= medium: yellow
  // < medium: green
}

/**
 * Default thresholds for resource utilization
 */
export const DEFAULT_THRESHOLDS: UtilizationThresholds = {
  high: 80,
  medium: 50,
} as const

/**
 * Color palette for utilization levels
 */
export const UTILIZATION_COLORS = {
  LOW: '#38a169',    // green (0-50%)
  MEDIUM: '#d69e2e', // yellow (50-80%)
  HIGH: '#e53e3e',   // red (80-100%)
} as const

/**
 * Get color based on utilization percentage
 * @param percentage - Utilization percentage (0-100)
 * @param thresholds - Optional custom thresholds
 * @returns Color hex code
 */
export const getUtilizationColor = (
  percentage: number,
  thresholds: UtilizationThresholds = DEFAULT_THRESHOLDS
): string => {
  if (percentage >= thresholds.high) {
    return UTILIZATION_COLORS.HIGH
  } else if (percentage >= thresholds.medium) {
    return UTILIZATION_COLORS.MEDIUM
  } else {
    return UTILIZATION_COLORS.LOW
  }
}

/**
 * Get Chakra UI color palette name based on utilization percentage
 * Used for components that accept colorPalette prop
 */
export const getUtilizationColorPalette = (
  percentage: number,
  thresholds: UtilizationThresholds = DEFAULT_THRESHOLDS
): 'green' | 'yellow' | 'red' => {
  if (percentage >= thresholds.high) {
    return 'red'
  } else if (percentage >= thresholds.medium) {
    return 'yellow'
  } else {
    return 'green'
  }
}

/**
 * Calculate utilization percentage
 * @param used - Amount used/reserved
 * @param total - Total amount available
 * @returns Percentage (0-100)
 */
export const calculateUtilization = (used: number, total: number): number => {
  if (total === 0) return 0
  return Math.round((used / total) * 100)
}
