import {
  getUtilizationColor,
  getUtilizationColorPalette,
  calculateUtilization,
  UTILIZATION_COLORS,
  DEFAULT_THRESHOLDS,
} from '../../src/utils/utilizationColors'

describe('getUtilizationColor', () => {
  it('returns HIGH color for >= 80%', () => {
    expect(getUtilizationColor(80)).toBe(UTILIZATION_COLORS.HIGH)
    expect(getUtilizationColor(100)).toBe(UTILIZATION_COLORS.HIGH)
  })

  it('returns MEDIUM color for 50-79%', () => {
    expect(getUtilizationColor(50)).toBe(UTILIZATION_COLORS.MEDIUM)
    expect(getUtilizationColor(79)).toBe(UTILIZATION_COLORS.MEDIUM)
  })

  it('returns LOW color for < 50%', () => {
    expect(getUtilizationColor(0)).toBe(UTILIZATION_COLORS.LOW)
    expect(getUtilizationColor(49)).toBe(UTILIZATION_COLORS.LOW)
  })

  it('accepts custom thresholds', () => {
    const custom = { high: 90, medium: 70 }
    expect(getUtilizationColor(85, custom)).toBe(UTILIZATION_COLORS.MEDIUM)
    expect(getUtilizationColor(95, custom)).toBe(UTILIZATION_COLORS.HIGH)
    expect(getUtilizationColor(50, custom)).toBe(UTILIZATION_COLORS.LOW)
  })
})

describe('getUtilizationColorPalette', () => {
  it('returns red for >= 80%', () => {
    expect(getUtilizationColorPalette(80)).toBe('red')
    expect(getUtilizationColorPalette(100)).toBe('red')
  })

  it('returns yellow for 50-79%', () => {
    expect(getUtilizationColorPalette(50)).toBe('yellow')
    expect(getUtilizationColorPalette(79)).toBe('yellow')
  })

  it('returns green for < 50%', () => {
    expect(getUtilizationColorPalette(0)).toBe('green')
    expect(getUtilizationColorPalette(49)).toBe('green')
  })

  it('accepts custom thresholds', () => {
    const custom = { high: 95, medium: 60 }
    expect(getUtilizationColorPalette(96, custom)).toBe('red')
    expect(getUtilizationColorPalette(70, custom)).toBe('yellow')
    expect(getUtilizationColorPalette(30, custom)).toBe('green')
  })
})

describe('calculateUtilization', () => {
  it('returns 0 when total is 0', () => {
    expect(calculateUtilization(50, 0)).toBe(0)
  })

  it('calculates percentage correctly', () => {
    expect(calculateUtilization(50, 100)).toBe(50)
  })

  it('rounds to nearest integer', () => {
    expect(calculateUtilization(1, 3)).toBe(33)
  })

  it('handles 100% utilization', () => {
    expect(calculateUtilization(200, 200)).toBe(100)
  })
})

describe('DEFAULT_THRESHOLDS', () => {
  it('has expected values', () => {
    expect(DEFAULT_THRESHOLDS.high).toBe(80)
    expect(DEFAULT_THRESHOLDS.medium).toBe(50)
  })
})
