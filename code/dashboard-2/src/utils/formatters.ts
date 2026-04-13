import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'
import utc from 'dayjs/plugin/utc'
import { JobState } from '../types/jobStates'

dayjs.extend(duration)
dayjs.extend(utc)

/**
 * Format duration in seconds to human-readable string (e.g., "2d 5h 30m")
 */
export const formatDuration = (seconds: number | undefined): string => {
  if (!seconds) return '-'
  const dur = dayjs.duration(seconds, 'seconds')
  
  const days = Math.floor(dur.asDays())
  const hours = dur.hours()
  const mins = dur.minutes()
  
  const parts: string[] = []
  if (days > 0) parts.push(`${days}d`)
  if (hours > 0) parts.push(`${hours}h`)
  if (mins > 0 || parts.length === 0) parts.push(`${mins}m`)
  
  return parts.join(' ')
}

/**
 * Format memory in MB to human-readable string (e.g., "2.5 GB" or "512 MB")
 */
export const formatMemory = (mb: number | undefined): string => {
  if (!mb) return '-'
  if (mb >= 1024) return `${(mb / 1024).toFixed(1)} GB`
  return `${mb.toFixed(1)} MB`
}

/**
 * Format date to UTC datetime string
 */
export const formatToUtcDateTimeString = (date: Date): string => {
  return dayjs(date).utc().format('YYYY-MM-DDTHH:mm[Z]')
}

/**
 * Format timestamp to human-readable datetime (e.g., "Jan 7, 13:17")
 */
export const formatDateTime = (datetime: number | Date): string => {
  return dayjs(datetime).format('MMM D, HH:mm')
}

/**
 * Format date to locale string (e.g., "1/15/2026, 2:30:00 PM")
 */
export const formatDateTimeToLocaleString = (date?: Date | null): string => {
  if (!date) return 'N/A'
  return new Date(date).toLocaleString()
}

/**
 * Calculate and format duration between two dates
 * Returns a human-readable string like "2d 5h" or "30m 45s"
 */
export const formatDurationBetweenDates = (startTime?: Date | null, endTime?: Date | null): string => {
  if (!startTime || !endTime) return 'N/A'
  const start = new Date(startTime).getTime()
  const end = new Date(endTime).getTime()
  const durationMs = end - start
  if (durationMs <= 0) return 'N/A'

  const seconds = Math.floor(durationMs / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return `${days}d ${hours % 24}h`
  if (hours > 0) return `${hours}h ${minutes % 60}m`
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`
  return `${seconds}s`
}

/**
 * Get the color palette for a job state (for Chakra UI components)
 * @param state - The job state string
 * @returns The Chakra UI color palette name
 */
export const getJobStateColor = (state: string): string => {
  if (state === JobState.RUNNING) return 'green'
  if (state === JobState.COMPLETED) return 'blue'
  if (state === JobState.PENDING) return 'yellow'
  if (state === JobState.SUSPENDED) return 'purple'
  if (
    state === JobState.FAILED ||
    state === JobState.TIMEOUT ||
    state === JobState.NODE_FAIL ||
    state === JobState.BOOT_FAIL ||
    state === JobState.DEADLINE ||
    state === JobState.OUT_OF_MEMORY
  ) return 'red'
  if (state === JobState.CANCELLED || state === JobState.PREEMPTED) return 'orange'
  return 'gray'
}

/**
 * Format CPU hours (e.g., "1,234 CPU-hours")
 */
export const formatCpuHours = (cpuHours: number): string => {
  return `${cpuHours.toLocaleString('en-US', { maximumFractionDigits: 2 })} CPU-hours`
}

/**
 * Format efficiency percentage (e.g., "85.4%")
 */
export const formatEfficiency = (efficiency: number | null): string => {
  if (efficiency === null) return 'N/A'
  return `${efficiency.toFixed(1)}%`
}

/**
 * Format I/O rate in bytes per second (e.g., "125.3 MB/s")
 */
export const formatIORate = (bytesPerSec: number): string => {
  if (bytesPerSec >= 1024 * 1024 * 1024) {
    return `${(bytesPerSec / (1024 * 1024 * 1024)).toFixed(1)} GB/s`
  }
  if (bytesPerSec >= 1024 * 1024) {
    return `${(bytesPerSec / (1024 * 1024)).toFixed(1)} MB/s`
  }
  if (bytesPerSec >= 1024) {
    return `${(bytesPerSec / 1024).toFixed(1)} KB/s`
  }
  return `${bytesPerSec.toFixed(1)} B/s`
}

/**
 * Format I/O operations per second (e.g., "1.2K IOPS" or "450 IOPS")
 */
export const formatIOPS = (ops: number): string => {
  if (ops >= 1_000_000) {
    return `${(ops / 1_000_000).toFixed(1)}M IOPS`
  }
  if (ops >= 1_000) {
    return `${(ops / 1_000).toFixed(1)}K IOPS`
  }
  return `${ops.toFixed(0)} IOPS`
}
