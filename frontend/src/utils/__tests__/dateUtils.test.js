import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  formatDateTime,
  formatDateTimeShort,
  formatDate,
  formatRelativeTime,
  formatAge
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

    it('should return "Unknown" for invalid date string', () => {
      expect(formatDateTime('invalid-date')).toBe('Unknown')
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

    it('should return "Unknown" for invalid date string', () => {
      expect(formatDateTimeShort('invalid-date')).toBe('Unknown')
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

    it('should return "Unknown" for invalid date string', () => {
      expect(formatDate('invalid-date')).toBe('Unknown')
    })
  })

  describe('formatRelativeTime', () => {
    it('should return "Just now" for very recent dates', () => {
      const date = new Date('2024-01-15T11:59:00Z')
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

    it('should return "Unknown" for invalid date string', () => {
      expect(formatRelativeTime('invalid-date')).toBe('Unknown')
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

    it('should return "Unknown" for invalid date string', () => {
      expect(formatAge('invalid-date')).toBe('Unknown')
    })
  })
})

