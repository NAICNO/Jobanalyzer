import { timeRangeToTimestamps } from '../../src/utils/timeRangeUtils'
import type { TimeRange } from '../../src/components/TimeRangePicker'

describe('timeRangeToTimestamps', () => {
  describe('absolute ranges', () => {
    it('converts absolute range with start and end', () => {
      const start = new Date('2026-01-15T14:00:00Z')
      const end = new Date('2026-01-15T15:00:00Z')
      const range: TimeRange = {
        type: 'absolute',
        label: 'Custom',
        value: '',
        start,
        end,
      }
      const result = timeRangeToTimestamps(range)
      expect(result.startTimeInS).toBe(Math.floor(start.getTime() / 1000))
      expect(result.endTimeInS).toBe(Math.floor(end.getTime() / 1000))
    })

    it('returns null for absolute range without start', () => {
      const range: TimeRange = {
        type: 'absolute',
        label: 'Custom',
        value: '',
        end: new Date(),
      }
      const result = timeRangeToTimestamps(range)
      expect(result.startTimeInS).toBeNull()
      expect(result.endTimeInS).toBeNull()
    })
  })

  describe('relative ranges', () => {
    it('converts minutes', () => {
      const range: TimeRange = {
        type: 'relative',
        label: 'Last 15 minutes',
        value: '15m',
        endAt: 'now',
      }
      const before = Math.floor(Date.now() / 1000)
      const result = timeRangeToTimestamps(range)
      const after = Math.floor(Date.now() / 1000)

      expect(result.endTimeInS).toBeGreaterThanOrEqual(before)
      expect(result.endTimeInS).toBeLessThanOrEqual(after)
      expect(result.startTimeInS).toBe(result.endTimeInS! - 15 * 60)
    })

    it('converts hours', () => {
      const range: TimeRange = {
        type: 'relative',
        label: 'Last 1 hour',
        value: '1h',
        endAt: 'now',
      }
      const result = timeRangeToTimestamps(range)
      expect(result.endTimeInS! - result.startTimeInS!).toBe(3600)
    })

    it('converts days', () => {
      const range: TimeRange = {
        type: 'relative',
        label: 'Last 7 days',
        value: '7d',
        endAt: 'now',
      }
      const result = timeRangeToTimestamps(range)
      expect(result.endTimeInS! - result.startTimeInS!).toBe(7 * 24 * 3600)
    })

    it('converts weeks', () => {
      const range: TimeRange = {
        type: 'relative',
        label: 'Last 2 weeks',
        value: '2w',
        endAt: 'now',
      }
      const result = timeRangeToTimestamps(range)
      expect(result.endTimeInS! - result.startTimeInS!).toBe(14 * 24 * 3600)
    })

    it('returns null for invalid duration format', () => {
      const range: TimeRange = {
        type: 'relative',
        label: 'Invalid',
        value: 'abc',
        endAt: 'now',
      }
      const result = timeRangeToTimestamps(range)
      expect(result.startTimeInS).toBeNull()
      expect(result.endTimeInS).toBeNull()
    })

    it('returns null for relative range without value', () => {
      const range: TimeRange = {
        type: 'relative',
        label: 'Empty',
        value: '',
      }
      const result = timeRangeToTimestamps(range)
      expect(result.startTimeInS).toBeNull()
      expect(result.endTimeInS).toBeNull()
    })
  })

  describe('custom/unknown ranges', () => {
    it('returns null for custom type', () => {
      const range: TimeRange = {
        type: 'custom',
        label: 'Custom',
        value: 'some-value',
      }
      const result = timeRangeToTimestamps(range)
      expect(result.startTimeInS).toBeNull()
      expect(result.endTimeInS).toBeNull()
    })
  })
})
