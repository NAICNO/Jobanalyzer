import {
  formatDuration,
  formatMemory,
  formatToUtcDateTimeString,
  formatDateTime,
  formatDateTimeToLocaleString,
  formatDurationBetweenDates,
  getJobStateColor,
  formatCpuHours,
  formatEfficiency,
  formatIORate,
  formatIOPS,
} from '../../src/utils/formatters'

describe('formatDuration', () => {
  it('returns dash for undefined', () => {
    expect(formatDuration(undefined)).toBe('-')
  })

  it('returns dash for zero', () => {
    expect(formatDuration(0)).toBe('-')
  })

  it('formats minutes only', () => {
    expect(formatDuration(300)).toBe('5m')
  })

  it('formats hours and minutes', () => {
    expect(formatDuration(5400)).toBe('1h 30m')
  })

  it('formats days, hours, and minutes', () => {
    expect(formatDuration(190800)).toBe('2d 5h')
  })

  it('formats exactly one day', () => {
    expect(formatDuration(86400)).toBe('1d')
  })

  it('shows 0m for durations under a minute', () => {
    expect(formatDuration(30)).toBe('0m')
  })
})

describe('formatMemory', () => {
  it('returns dash for undefined', () => {
    expect(formatMemory(undefined)).toBe('-')
  })

  it('returns dash for zero', () => {
    expect(formatMemory(0)).toBe('-')
  })

  it('formats values in MB', () => {
    expect(formatMemory(512)).toBe('512.0 MB')
  })

  it('formats values in GB', () => {
    expect(formatMemory(2560)).toBe('2.5 GB')
  })

  it('formats exactly 1 GB', () => {
    expect(formatMemory(1024)).toBe('1.0 GB')
  })
})

describe('formatToUtcDateTimeString', () => {
  it('formats a date to UTC string', () => {
    const date = new Date('2026-01-15T14:30:00Z')
    expect(formatToUtcDateTimeString(date)).toBe('2026-01-15T14:30Z')
  })
})

describe('formatDateTime', () => {
  it('formats a timestamp to readable string', () => {
    const date = new Date('2026-01-07T13:17:00')
    expect(formatDateTime(date)).toBe('Jan 7, 13:17')
  })
})

describe('formatDateTimeToLocaleString', () => {
  it('returns N/A for null', () => {
    expect(formatDateTimeToLocaleString(null)).toBe('N/A')
  })

  it('returns N/A for undefined', () => {
    expect(formatDateTimeToLocaleString(undefined)).toBe('N/A')
  })

  it('formats a valid date', () => {
    const date = new Date('2026-01-15T14:30:00')
    const result = formatDateTimeToLocaleString(date)
    // Just verify it returns a non-empty string (locale output varies)
    expect(result).not.toBe('N/A')
    expect(result.length).toBeGreaterThan(0)
  })
})

describe('formatDurationBetweenDates', () => {
  it('returns N/A for null start', () => {
    expect(formatDurationBetweenDates(null, new Date())).toBe('N/A')
  })

  it('returns N/A for null end', () => {
    expect(formatDurationBetweenDates(new Date(), null)).toBe('N/A')
  })

  it('returns N/A for negative duration', () => {
    const start = new Date('2026-01-15T15:00:00Z')
    const end = new Date('2026-01-15T14:00:00Z')
    expect(formatDurationBetweenDates(start, end)).toBe('N/A')
  })

  it('formats seconds', () => {
    const start = new Date('2026-01-15T14:00:00Z')
    const end = new Date('2026-01-15T14:00:45Z')
    expect(formatDurationBetweenDates(start, end)).toBe('45s')
  })

  it('formats minutes and seconds', () => {
    const start = new Date('2026-01-15T14:00:00Z')
    const end = new Date('2026-01-15T14:30:45Z')
    expect(formatDurationBetweenDates(start, end)).toBe('30m 45s')
  })

  it('formats hours and minutes', () => {
    const start = new Date('2026-01-15T14:00:00Z')
    const end = new Date('2026-01-15T16:15:00Z')
    expect(formatDurationBetweenDates(start, end)).toBe('2h 15m')
  })

  it('formats days and hours', () => {
    const start = new Date('2026-01-15T14:00:00Z')
    const end = new Date('2026-01-17T19:00:00Z')
    expect(formatDurationBetweenDates(start, end)).toBe('2d 5h')
  })
})

describe('getJobStateColor', () => {
  it('returns green for RUNNING', () => {
    expect(getJobStateColor('RUNNING')).toBe('green')
  })

  it('returns blue for COMPLETED', () => {
    expect(getJobStateColor('COMPLETED')).toBe('blue')
  })

  it('returns yellow for PENDING', () => {
    expect(getJobStateColor('PENDING')).toBe('yellow')
  })

  it('returns purple for SUSPENDED', () => {
    expect(getJobStateColor('SUSPENDED')).toBe('purple')
  })

  it('returns red for failure states', () => {
    expect(getJobStateColor('FAILED')).toBe('red')
    expect(getJobStateColor('TIMEOUT')).toBe('red')
    expect(getJobStateColor('NODE_FAIL')).toBe('red')
    expect(getJobStateColor('BOOT_FAIL')).toBe('red')
    expect(getJobStateColor('DEADLINE')).toBe('red')
    expect(getJobStateColor('OUT_OF_MEMORY')).toBe('red')
  })

  it('returns orange for CANCELLED and PREEMPTED', () => {
    expect(getJobStateColor('CANCELLED')).toBe('orange')
    expect(getJobStateColor('PREEMPTED')).toBe('orange')
  })

  it('returns gray for unknown states', () => {
    expect(getJobStateColor('UNKNOWN_STATE')).toBe('gray')
  })
})

describe('formatCpuHours', () => {
  it('formats integer CPU hours', () => {
    expect(formatCpuHours(1234)).toBe('1,234 CPU-hours')
  })

  it('formats fractional CPU hours', () => {
    expect(formatCpuHours(0.5)).toBe('0.5 CPU-hours')
  })
})

describe('formatEfficiency', () => {
  it('returns N/A for null', () => {
    expect(formatEfficiency(null)).toBe('N/A')
  })

  it('formats efficiency percentage', () => {
    expect(formatEfficiency(85.42)).toBe('85.4%')
  })

  it('formats zero efficiency', () => {
    expect(formatEfficiency(0)).toBe('0.0%')
  })
})

describe('formatIORate', () => {
  it('formats bytes per second', () => {
    expect(formatIORate(500)).toBe('500.0 B/s')
  })

  it('formats kilobytes per second', () => {
    expect(formatIORate(2048)).toBe('2.0 KB/s')
  })

  it('formats megabytes per second', () => {
    expect(formatIORate(5 * 1024 * 1024)).toBe('5.0 MB/s')
  })

  it('formats gigabytes per second', () => {
    expect(formatIORate(2 * 1024 * 1024 * 1024)).toBe('2.0 GB/s')
  })
})

describe('formatIOPS', () => {
  it('formats small IOPS', () => {
    expect(formatIOPS(450)).toBe('450 IOPS')
  })

  it('formats thousands of IOPS', () => {
    expect(formatIOPS(1200)).toBe('1.2K IOPS')
  })

  it('formats millions of IOPS', () => {
    expect(formatIOPS(2500000)).toBe('2.5M IOPS')
  })
})
