import {
  calculateLiveness,
  getLivenessStats,
  LIVENESS_THRESHOLDS,
  LIVENESS_COLORS,
} from '../../src/utils/nodeLiveness'

describe('calculateLiveness', () => {
  it('returns unknown for null timestamp', () => {
    const result = calculateLiveness(null)
    expect(result.icon).toBe('unknown')
    expect(result.color).toBe(LIVENESS_COLORS.UNKNOWN)
    expect(result.ageText).toBe('No data available')
    expect(result.ageMinutes).toBe(Infinity)
  })

  it('returns live for recent timestamp (< 5 min)', () => {
    const twoMinAgo = new Date(Date.now() - 2 * 60 * 1000)
    const result = calculateLiveness(twoMinAgo)
    expect(result.icon).toBe('live')
    expect(result.color).toBe(LIVENESS_COLORS.LIVE)
  })

  it('returns stale for 5-15 min old timestamp', () => {
    const tenMinAgo = new Date(Date.now() - 10 * 60 * 1000)
    const result = calculateLiveness(tenMinAgo)
    expect(result.icon).toBe('stale')
    expect(result.color).toBe(LIVENESS_COLORS.STALE)
  })

  it('returns offline for > 15 min old timestamp', () => {
    const thirtyMinAgo = new Date(Date.now() - 30 * 60 * 1000)
    const result = calculateLiveness(thirtyMinAgo)
    expect(result.icon).toBe('offline')
    expect(result.color).toBe(LIVENESS_COLORS.OFFLINE)
  })

  it('shows "<1 min ago" for very recent timestamps', () => {
    const justNow = new Date(Date.now() - 10 * 1000) // 10 seconds ago
    const result = calculateLiveness(justNow)
    expect(result.ageText).toBe('Last seen <1 min ago')
  })

  it('shows minutes for timestamps < 60 min', () => {
    const tenMinAgo = new Date(Date.now() - 10 * 60 * 1000)
    const result = calculateLiveness(tenMinAgo)
    expect(result.ageText).toBe('Last seen 10 min ago')
  })

  it('shows hours for timestamps >= 60 min', () => {
    const twoHoursAgo = new Date(Date.now() - 120 * 60 * 1000)
    const result = calculateLiveness(twoHoursAgo)
    expect(result.ageText).toBe('Last seen 2 hours ago')
  })

  it('shows singular hour', () => {
    const oneHourAgo = new Date(Date.now() - 60 * 60 * 1000)
    const result = calculateLiveness(oneHourAgo)
    expect(result.ageText).toBe('Last seen 1 hour ago')
  })

  it('boundary: exactly at LIVE threshold', () => {
    const atThreshold = new Date(Date.now() - LIVENESS_THRESHOLDS.LIVE * 60 * 1000)
    const result = calculateLiveness(atThreshold)
    // At exactly 5 min, should be stale (>= 5 is stale)
    expect(result.icon).toBe('stale')
  })
})

describe('getLivenessStats', () => {
  it('returns all zeros for empty map', () => {
    const stats = getLivenessStats({})
    expect(stats.liveNodes).toBe(0)
    expect(stats.staleNodes).toBe(0)
    expect(stats.offlineNodes).toBe(0)
  })

  it('correctly categorizes nodes', () => {
    const now = Date.now()
    const probeMap: Record<string, Date | null> = {
      'node-01': new Date(now - 1 * 60 * 1000),   // live (1 min)
      'node-02': new Date(now - 2 * 60 * 1000),   // live (2 min)
      'node-03': new Date(now - 10 * 60 * 1000),  // stale (10 min)
      'node-04': new Date(now - 30 * 60 * 1000),  // offline (30 min)
      'node-05': null,                              // unknown → offline
    }
    const stats = getLivenessStats(probeMap)
    expect(stats.liveNodes).toBe(2)
    expect(stats.staleNodes).toBe(1)
    expect(stats.offlineNodes).toBe(2) // 1 offline + 1 unknown
  })
})

describe('LIVENESS_THRESHOLDS', () => {
  it('has expected threshold values', () => {
    expect(LIVENESS_THRESHOLDS.LIVE).toBe(5)
    expect(LIVENESS_THRESHOLDS.STALE).toBe(15)
  })
})
