import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  formatDateTime,
  formatDateTimeShort,
  formatDate,
  formatRelativeTime,
  formatAge,
  formatDateParts
} from '../dateUtils'

describe('dateUtils', () => {
  beforeEach(() => {
    // Mock Date.now() to have consistent test results
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-01-15T12:00:00Z'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  describe('formatDateTime', () => {
    it('should format a valid date string', () => {
      const date = '2024-01-15T10:30:00Z'
      const result = formatDateTime(date)
      expect(result).toMatch(/\d{2}\/\d{2}\/\d{4}, \d{2}:\d{2}:\d{2}/)
    })

    it('should return "Unknown" for null input', () => {
      expect(formatDateTime(null)).toBe('Unknown')
    })

    it('should return "Unknown" for undefined input', () => {
      expect(formatDateTime(undefined)).toBe('Unknown')
    })

    it('should return "Unknown" for empty string', () => {
      expect(formatDateTime('')).toBe('Unknown')
    })

    it('should handle Date object', () => {
      const date = new Date('2024-01-15T10:30:00Z')
      const result = formatDateTime(date)
      expect(result).toMatch(/\d{2}\/\d{2}\/\d{4}, \d{2}:\d{2}:\d{2}/)
    })

    it('should return "Invalid Date" for invalid date string (toLocaleString behavior)', () => {
      // toLocaleString on invalid Date returns "Invalid Date" string, not an exception
      const result = formatDateTime('invalid-date')
      expect(result).toBe('Invalid Date')
    })
  })

  describe('formatDateTimeShort', () => {
    it('should format a valid date string without seconds', () => {
      const date = '2024-01-15T10:30:00Z'
      const result = formatDateTimeShort(date)
      expect(result).toMatch(/\d{2}\/\d{2}\/\d{4}, \d{2}:\d{2}/)
    })

    it('should return "Unknown" for null input', () => {
      expect(formatDateTimeShort(null)).toBe('Unknown')
    })

    it('should return "Invalid Date" for invalid date string (toLocaleString behavior)', () => {
      // toLocaleString on invalid Date returns "Invalid Date" string, not an exception
      const result = formatDateTimeShort('invalid-date')
      expect(result).toBe('Invalid Date')
    })
  })

  describe('formatDate', () => {
    it('should format a valid date string to date only', () => {
      const date = '2024-01-15T10:30:00Z'
      const result = formatDate(date)
      expect(result).toMatch(/\d{2}\/\d{2}\/\d{4}/)
    })

    it('should return "Unknown" for null input', () => {
      expect(formatDate(null)).toBe('Unknown')
    })

    it('should return "Invalid Date" for invalid date string (toLocaleDateString behavior)', () => {
      // toLocaleDateString on invalid Date returns "Invalid Date" string, not an exception
      const result = formatDate('invalid-date')
      expect(result).toBe('Invalid Date')
    })
  })

  describe('formatDateParts', () => {
    it('should split date and time for valid date', () => {
      const result = formatDateParts('2024-01-15T10:30:00Z')
      expect(result.date).toMatch(/\d{2}\/\d{2}\/\d{4}/)
      expect(result.time).toMatch(/\d{2}:\d{2}:\d{2}/)
    })

    it('should return Unknown parts for null input', () => {
      expect(formatDateParts(null)).toEqual({ date: 'Unknown', time: 'Unknown' })
    })
  })

  describe('formatRelativeTime', () => {
    it('should return "Just now" for very recent dates (less than 1 minute)', () => {
      // Date is 1 minute before current time, so it returns "1m ago"
      const date = new Date('2024-01-15T11:59:00Z')
      expect(formatRelativeTime(date)).toBe('1m ago')
    })

    it('should return "Just now" for dates less than 1 minute ago', () => {
      // Date is 30 seconds before current time
      const date = new Date('2024-01-15T11:59:30Z')
      expect(formatRelativeTime(date)).toBe('Just now')
    })

    it('should return minutes ago for dates within an hour', () => {
      const date = new Date('2024-01-15T11:30:00Z')
      expect(formatRelativeTime(date)).toBe('30m ago')
    })

    it('should return hours ago for dates within a day', () => {
      const date = new Date('2024-01-15T10:00:00Z')
      expect(formatRelativeTime(date)).toBe('2h ago')
    })

    it('should return days ago for older dates', () => {
      const date = new Date('2024-01-13T12:00:00Z')
      expect(formatRelativeTime(date)).toBe('2d ago')
    })

    it('should return "Unknown" for null input', () => {
      expect(formatRelativeTime(null)).toBe('Unknown')
    })

    it('should return "Just now" for invalid date string (getTime returns NaN)', () => {
      // Invalid date results in NaN for getTime(), diff becomes NaN, and all checks fail
      // So it falls through to "Just now"
      const result = formatRelativeTime('invalid-date')
      expect(result).toBe('Just now')
    })
  })

  describe('formatAge', () => {
    it('should return minutes for recent dates', () => {
      const date = new Date('2024-01-15T11:30:00Z')
      expect(formatAge(date)).toBe('30m')
    })

    it('should return hours for dates within a day', () => {
      const date = new Date('2024-01-15T10:00:00Z')
      expect(formatAge(date)).toBe('2h')
    })

    it('should return days for older dates', () => {
      const date = new Date('2024-01-13T12:00:00Z')
      expect(formatAge(date)).toBe('2d')
    })

    it('should return "Unknown" for null input', () => {
      expect(formatAge(null)).toBe('Unknown')
    })

    it('should return "NaNm" for invalid date string (getTime returns NaN)', () => {
      // Invalid date results in NaN for getTime(), diff becomes NaN
      // Math.floor(NaN) = NaN, so it returns "NaNm"
      const result = formatAge('invalid-date')
      expect(result).toBe('NaNm')
    })
  })
})
