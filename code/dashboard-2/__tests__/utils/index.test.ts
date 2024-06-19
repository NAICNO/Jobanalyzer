import {
  breakText,
  findCluster,
  findSubcluster,
  parseDateString,
  parseRelativeDate,
  reformatHostDescriptions,
  toPercentage,
  validateDateFormat
} from '../../src/util'
import moment from 'moment'
import { describe } from 'vitest'

describe('findCluster', () => {
  it('should return a cluster object for a valid cluster name', () => {
    const clusterName = 'ML nodes'
    const cluster = findCluster('ml')
    expect(cluster).toBeDefined()
    expect(cluster!.name).toBe(clusterName)
  })

  it('should return null for an invalid cluster name', () => {
    const clusterName = 'invalidCluster'
    const cluster = findCluster(clusterName)
    expect(cluster).toBeNull()
  })

  it('should return null for an undefined cluster name', () => {
    const cluster = findCluster(undefined)
    expect(cluster).toBeNull()
  })

  it('should return null for null cluster name', () => {
    const cluster = findCluster(null)
    expect(cluster).toBeNull()
  })

  it('should return null for an empty cluster name', () => {
    const cluster = findCluster('')
    expect(cluster).toBeNull()
  })

  it('should return null for an empty cluster name', () => {
    const cluster = findCluster()
    expect(cluster).toBeNull()
  })

})

describe('findSubcluster', () => {
  it('should return a cluster and subcluster object for a valid cluster and subcluster name', () => {
    const clusterName = 'ML nodes'
    const subclusterName = 'nvidia'
    const result = findSubcluster('ml', 'nvidia')
    const {cluster, subcluster} = result!
    expect(cluster).toBeDefined()
    expect(cluster!.name).toBe(clusterName)
    expect(subcluster).toBeDefined()
    expect(subcluster!.name).toBe(subclusterName)
  })

  it('should return null for an invalid cluster name', () => {
    const clusterName = 'invalidCluster'
    const subclusterName = 'invalidSubcluster'
    const result = findSubcluster(clusterName, subclusterName)
    expect(result).toBeNull()
  })

  it('should return null for an valid cluster name and invalid subcluster name', () => {
    const clusterName = 'ml'
    const subclusterName = 'invalidSubcluster'
    const result = findSubcluster(clusterName, subclusterName)
    expect(result).toBeNull()
  })

  it('should return null for an invalid cluster name and valid subcluster name', () => {
    const clusterName = 'invalidCluster'
    const subclusterName = 'nvidia'
    const result = findSubcluster(clusterName, subclusterName)
    expect(result).toBeNull()
  })

  it('should return null for an undefined cluster name', () => {
    const result = findSubcluster(undefined, undefined)
    expect(result).toBeNull()
  })

  it('should return null for null cluster name', () => {
    const result = findSubcluster(null, null)
    expect(result).toBeNull()
  })

  it('should return null for an empty cluster name', () => {
    const result = findSubcluster('', '')
    expect(result).toBeNull()
  })

  it('should return null for an empty cluster name', () => {
    const result = findSubcluster()
    expect(result).toBeNull()
  })

})

describe('breakText', () => {
  it('should replace spaces with non-breaking spaces', () => {
    const input = 'Hello world'
    const output = breakText(input)
    expect(output).toBe('Hello\xA0world')
  })

  it('should add a space after commas', () => {
    const input = 'one,two,three'
    const output = breakText(input)
    expect(output).toBe('one, two, three')
  })

  it('should handle empty strings correctly', () => {
    const input = ''
    const output = breakText(input)
    expect(output).toBe('')
  })

  it('should handle strings without spaces or commas correctly', () => {
    const input = 'HelloWorld'
    const output = breakText(input)
    expect(output).toBe('HelloWorld')
  })

  it('should handle multiple spaces correctly', () => {
    const input = 'Hello    world'
    const output = breakText(input)
    expect(output).toBe('Hello\xA0\xA0\xA0\xA0world')
  })

  it('should handle multiple commas correctly', () => {
    const input = 'one,two,,three'
    const output = breakText(input)
    expect(output).toBe('one, two, , three')
  })
})

describe('toPercentage', () => {
  it('should convert a whole number to a percentage string with one decimal place', () => {
    const value = 50
    const output = toPercentage(value)
    expect(output).toBe('0.5')
  })

  it('should convert a decimal number to a percentage string with one decimal place', () => {
    const value = 12.34
    const output = toPercentage(value)
    expect(output).toBe('0.1')
  })

  it('should convert a number greater than 100 to a percentage string with one decimal place', () => {
    const value = 123
    const output = toPercentage(value)
    expect(output).toBe('1.2')
  })

  it('should handle zero correctly', () => {
    const value = 0
    const output = toPercentage(value)
    expect(output).toBe('0.0')
  })

  it('should handle negative numbers correctly', () => {
    const value = -50
    const output = toPercentage(value)
    expect(output).toBe('-0.5')
  })

  it('should handle small decimal numbers correctly', () => {
    const value = 0.01
    const output = toPercentage(value)
    expect(output).toBe('0.0')
  })

  it('should handle large numbers correctly', () => {
    const value = 123456789
    const output = toPercentage(value)
    expect(output).toBe('1234567.9')
  })
})

describe('parseDateString', () => {
  it('should parse a date string with the default format', () => {
    const dateString = '2024-06-07 01:05'
    const expectedDate = new Date(Date.UTC(2024, 5, 7, 1, 5)) // Month is 0-indexed
    const output = parseDateString(dateString)
    expect(output.toISOString()).toBe(expectedDate.toISOString())
  })

  it('should parse a date string with a custom format', () => {
    const dateString = '07/06/2024 01:05'
    const format = 'DD/MM/YYYY HH:mm'
    const expectedDate = new Date(Date.UTC(2024, 5, 7, 1, 5)) // Month is 0-indexed
    const output = parseDateString(dateString, format)
    expect(output.toISOString()).toBe(expectedDate.toISOString())
  })

  it('should handle invalid date strings', () => {
    const dateString = 'invalid date'
    const output = parseDateString(dateString)
    expect(output.toString()).toBe('Invalid Date')
  })

  it('should handle different time zones correctly', () => {
    const dateString = '2024-06-07 01:05+02:00'
    const expectedDate = new Date(Date.UTC(2024, 5, 6, 23, 5)) // Convert to UTC
    const output = parseDateString(dateString, 'YYYY-MM-DD HH:mmZ')
    expect(output.toISOString()).toBe(expectedDate.toISOString())
  })

  it('should parse a date string without a time part', () => {
    const dateString = '2024-06-07'
    const expectedDate = new Date(Date.UTC(2024, 5, 7)) // Month is 0-indexed
    const output = parseDateString(dateString, 'YYYY-MM-DD')
    expect(output.toISOString()).toBe(expectedDate.toISOString())
  })

  it('should parse a date string with a time part only', () => {
    const dateString = '01:05'
    const format = 'HH:mm'
    const baseDate = moment.utc().format('YYYY-MM-DD')
    const expectedDate = moment.utc(`${baseDate} 01:05`, 'YYYY-MM-DD HH:mm').toDate()
    const output = parseDateString(dateString, format)
    expect(output.toISOString()).toBe(expectedDate.toISOString())
  })

  it('should handle leap year dates correctly', () => {
    const dateString = '2024-02-29 12:00'
    const expectedDate = new Date(Date.UTC(2024, 1, 29, 12, 0)) // Month is 0-indexed
    const output = parseDateString(dateString)
    expect(output.toISOString()).toBe(expectedDate.toISOString())
  })
})

describe('validateDateFormat', () => {
  it('should return true for valid absolute date format', () => {
    expect(validateDateFormat('2024-12-23')).toBe(true)
    expect(validateDateFormat('2000-04-02')).toBe(true)
  })

  it('should return true for valid relative date format', () => {
    expect(validateDateFormat('1w')).toBe(true)
    expect(validateDateFormat('5d')).toBe(true)
  })

  it('should return false for invalid date format', () => {
    expect(validateDateFormat('2024-13-15')).toBe(false)
    expect(validateDateFormat('2015-12-36')).toBe(false)
    expect(validateDateFormat('2023-02-29')).toBe(false) // 2023 is not a leap year
    expect(validateDateFormat('10x')).toBe(false)
  })

  it('should return true for empty value', () => {
    expect(validateDateFormat(undefined)).toBe(true)
    expect(validateDateFormat('')).toBe(true)
  })
})

describe('parseRelativeDate', () => {
  it('should parse relative date to absolute date', () => {
    const now = moment()
    expect(parseRelativeDate('1w').isSame(now.clone().add(1, 'weeks'), 'day')).toBe(true)
    expect(parseRelativeDate('5d').isSame(now.clone().add(5, 'days'), 'day')).toBe(true)
  })

  it('should parse absolute date correctly', () => {
    expect(parseRelativeDate('2024-12-23').isSame(moment('2024-12-23', 'YYYY-MM-DD', true))).toBe(true)
    expect(parseRelativeDate('2000-04-02').isSame(moment('2000-04-02', 'YYYY-MM-DD', true))).toBe(true)
  })

  it('should return an invalid moment object for invalid date format', () => {
    expect(parseRelativeDate('2024-13-15').isValid()).toBe(false)
    expect(parseRelativeDate('2015-12-36').isValid()).toBe(false)
    expect(parseRelativeDate('10x').isValid()).toBe(false)
  })
})

describe('reformatHostDescriptions', () => {
  test('should handle a single description', () => {
    const description = 'apple'
    const result = reformatHostDescriptions(description)
    expect(result).toBe('1x apple')
  })

  test('should handle multiple descriptions with different counts', () => {
    const description = 'apple|||banana|||apple|||apple|||banana'
    const result = reformatHostDescriptions(description)
    expect(result).toBe('3x apple\n2x banana')
  })

  test('should handle descriptions with the same counts sorted alphabetically', () => {
    const description = 'banana|||apple'
    const result = reformatHostDescriptions(description)
    expect(result).toBe('1x apple\n1x banana')
  })

  test('should handle an empty string', () => {
    const description = ''
    const result = reformatHostDescriptions(description)
    expect(result).toBe('')
  })

  test('should handle descriptions with special characters', () => {
    const description = 'apple|||b@n@n@|||apple|||b@n@n@'
    const result = reformatHostDescriptions(description)
    expect(result).toBe('2x apple\n2x b@n@n@')
  })

  test('should handle descriptions with spaces', () => {
    const description = 'apple|||banana split|||apple|||banana split|||apple'
    const result = reformatHostDescriptions(description)
    expect(result).toBe('3x apple\n2x banana split')
  })

  test('should handle descriptions with varying counts and order', () => {
    const description = 'cat|||dog|||cat|||bird|||dog|||dog|||cat|||bird|||cat'
    const result = reformatHostDescriptions(description)
    expect(result).toBe('4x cat\n3x dog\n2x bird')
  })
})

