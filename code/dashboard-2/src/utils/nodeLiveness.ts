/**
 * Utilities for calculating and managing node liveness status
 */

export interface LivenessStatus {
  color: string
  icon: 'live' | 'stale' | 'offline' | 'unknown'
  ageText: string
  ageMinutes: number
}

export interface LivenessStats {
  liveNodes: number
  staleNodes: number
  offlineNodes: number
  totalReporting: number
  livePercentage: number
}

/**
 * Thresholds for liveness classification (in minutes)
 */
export const LIVENESS_THRESHOLDS = {
  LIVE: 5,      // < 5 minutes: Live
  STALE: 15,    // 5-15 minutes: Stale
  // > 15 minutes: Offline
} as const

/**
 * Colors for liveness states
 */
export const LIVENESS_COLORS = {
  LIVE: '#38a169',    // green
  STALE: '#d69e2e',   // yellow
  OFFLINE: '#e53e3e', // red
  UNKNOWN: '#a0aec0', // gray
} as const

/**
 * Calculate liveness status for a node based on its last probe timestamp
 */
export const calculateLiveness = (timestamp: Date | null): LivenessStatus => {
  if (!timestamp) {
    return {
      color: LIVENESS_COLORS.UNKNOWN,
      icon: 'unknown',
      ageText: 'No data available',
      ageMinutes: Infinity,
    }
  }

  const now = new Date()
  const nodeTime = new Date(timestamp)
  const ageMs = now.getTime() - nodeTime.getTime()
  const ageMinutes = ageMs / 1000 / 60

  // Format age text
  let ageText: string
  if (ageMinutes < 1) {
    ageText = 'Last seen <1 min ago'
  } else if (ageMinutes < 60) {
    ageText = `Last seen ${Math.round(ageMinutes)} min ago`
  } else {
    const ageHours = Math.round(ageMinutes / 60)
    ageText = `Last seen ${ageHours} hour${ageHours > 1 ? 's' : ''} ago`
  }

  // Determine status based on thresholds
  if (ageMinutes < LIVENESS_THRESHOLDS.LIVE) {
    return {
      color: LIVENESS_COLORS.LIVE,
      icon: 'live',
      ageText,
      ageMinutes,
    }
  } else if (ageMinutes < LIVENESS_THRESHOLDS.STALE) {
    return {
      color: LIVENESS_COLORS.STALE,
      icon: 'stale',
      ageText,
      ageMinutes,
    }
  } else {
    return {
      color: LIVENESS_COLORS.OFFLINE,
      icon: 'offline',
      ageText,
      ageMinutes,
    }
  }
}

/**
 * Calculate liveness statistics from a map of node timestamps
 */
export const getLivenessStats = (
  lastProbeMap: Record<string, Date | null>
): LivenessStats => {
  let liveNodes = 0
  let staleNodes = 0
  let offlineNodes = 0

  Object.entries(lastProbeMap).forEach(([_, timestamp]) => {
    const status = calculateLiveness(timestamp)
    
    switch (status.icon) {
    case 'live':
      liveNodes++
      break
    case 'stale':
      staleNodes++
      break
    case 'offline':
    case 'unknown':
      offlineNodes++
      break
    }
  })


  return {
    liveNodes,
    staleNodes,
    offlineNodes,
  }
}
