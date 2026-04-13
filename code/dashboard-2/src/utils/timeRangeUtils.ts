import type { TimeRange } from '../components/TimeRangePicker'

/**
 * Parse duration string (e.g., '15m', '1h', '7d') to milliseconds
 */
const parseDurationToMs = (durationStr: string): number | null => {
  const match = durationStr.match(/^(\d+)([mhdw])$/)
  if (!match) return null
  const num = parseInt(match[1], 10)
  const unit = match[2]
  switch (unit) {
  case 'm': return num * 60 * 1000
  case 'h': return num * 60 * 60 * 1000
  case 'd': return num * 24 * 60 * 60 * 1000
  case 'w': return num * 7 * 24 * 60 * 60 * 1000
  default: return null
  }
}

/**
 * Convert TimeRange to start and end timestamps in seconds since epoch
 * Returns null for both if the range cannot be converted
 */
export const timeRangeToTimestamps = (
  timeRange: TimeRange
): { startTimeInS: number | null; endTimeInS: number | null } => {
  // Absolute range with explicit dates
  if (timeRange.type === 'absolute' && timeRange.start && timeRange.end) {
    return {
      startTimeInS: Math.floor(timeRange.start.getTime() / 1000),
      endTimeInS: Math.floor(timeRange.end.getTime() / 1000),
    }
  }

  // Relative range (e.g., "Last 1 hour")
  if (timeRange.type === 'relative' && timeRange.value) {
    const durationMs = parseDurationToMs(timeRange.value)
    if (!durationMs) {
      return { startTimeInS: null, endTimeInS: null }
    }

    const now = Date.now()
    const endTime = timeRange.endAt === 'now' ? now : now
    const startTime = endTime - durationMs

    return {
      startTimeInS: Math.floor(startTime / 1000),
      endTimeInS: Math.floor(endTime / 1000),
    }
  }

  // Custom or unknown - return null to let API use defaults
  return { startTimeInS: null, endTimeInS: null }
}
