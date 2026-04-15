import {
  ClusterColors,
  getUtilizationColor,
  getSuccessRateColor,
  getNodeStateColor,
} from '../../src/utils/colorStandards'

describe('getUtilizationColor', () => {
  it('returns critical (red) for > 90%', () => {
    expect(getUtilizationColor(95)).toBe(ClusterColors.UTIL_CRITICAL)
  })

  it('returns high (yellow) for 71-90%', () => {
    expect(getUtilizationColor(80)).toBe(ClusterColors.UTIL_HIGH)
    expect(getUtilizationColor(71)).toBe(ClusterColors.UTIL_HIGH)
  })

  it('returns moderate (green) for 41-70%', () => {
    expect(getUtilizationColor(50)).toBe(ClusterColors.UTIL_MODERATE)
    expect(getUtilizationColor(41)).toBe(ClusterColors.UTIL_MODERATE)
  })

  it('returns low (blue) for <= 40%', () => {
    expect(getUtilizationColor(40)).toBe(ClusterColors.UTIL_LOW)
    expect(getUtilizationColor(10)).toBe(ClusterColors.UTIL_LOW)
  })

  it('handles boundary at exactly 90', () => {
    expect(getUtilizationColor(90)).toBe(ClusterColors.UTIL_HIGH)
  })

  it('handles zero', () => {
    expect(getUtilizationColor(0)).toBe(ClusterColors.UTIL_LOW)
  })
})

describe('getSuccessRateColor', () => {
  it('returns healthy (green) for >= 80%', () => {
    expect(getSuccessRateColor(80)).toBe(ClusterColors.STATE_HEALTHY)
    expect(getSuccessRateColor(100)).toBe(ClusterColors.STATE_HEALTHY)
  })

  it('returns warning (yellow) for 50-79%', () => {
    expect(getSuccessRateColor(50)).toBe(ClusterColors.STATE_WARNING)
    expect(getSuccessRateColor(79)).toBe(ClusterColors.STATE_WARNING)
  })

  it('returns critical (red) for < 50%', () => {
    expect(getSuccessRateColor(49)).toBe(ClusterColors.STATE_CRITICAL)
    expect(getSuccessRateColor(0)).toBe(ClusterColors.STATE_CRITICAL)
  })
})

describe('getNodeStateColor', () => {
  it('returns red for DOWN states', () => {
    expect(getNodeStateColor('DOWN')).toBe(ClusterColors.NODE_DOWN)
    expect(getNodeStateColor('DOWN*')).toBe(ClusterColors.NODE_DOWN)
  })

  it('returns red for DRAIN states', () => {
    expect(getNodeStateColor('DRAIN')).toBe(ClusterColors.NODE_DOWN)
    expect(getNodeStateColor('DRAINED')).toBe(ClusterColors.NODE_DOWN)
  })

  it('returns red for FAIL states', () => {
    expect(getNodeStateColor('FAIL')).toBe(ClusterColors.NODE_DOWN)
  })

  it('returns green for IDLE states', () => {
    expect(getNodeStateColor('IDLE')).toBe(ClusterColors.NODE_IDLE)
  })

  it('returns blue for ALLOCATED states', () => {
    expect(getNodeStateColor('ALLOCATED')).toBe(ClusterColors.NODE_ALLOCATED)
  })

  it('returns blue for MIXED states', () => {
    expect(getNodeStateColor('MIXED')).toBe(ClusterColors.NODE_ALLOCATED)
  })

  it('returns yellow for MAINTENANCE and related states', () => {
    expect(getNodeStateColor('MAINTENANCE')).toBe(ClusterColors.NODE_MAINTENANCE)
    expect(getNodeStateColor('COMPLETING')).toBe(ClusterColors.NODE_MAINTENANCE)
    expect(getNodeStateColor('PLANNED')).toBe(ClusterColors.NODE_MAINTENANCE)
    expect(getNodeStateColor('RESERVED')).toBe(ClusterColors.NODE_MAINTENANCE)
  })

  it('returns gray for unknown states', () => {
    expect(getNodeStateColor('SOME_UNKNOWN')).toBe(ClusterColors.STATE_NEUTRAL)
  })

  it('is case-insensitive', () => {
    expect(getNodeStateColor('idle')).toBe(ClusterColors.NODE_IDLE)
    expect(getNodeStateColor('down')).toBe(ClusterColors.NODE_DOWN)
  })
})
