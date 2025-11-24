import { describe, it, expect } from 'vitest'
import {
  parseCpu,
  parseMemory,
  calculatePercentage,
  formatResourceValue
} from '../resourceParser'

describe('resourceParser', () => {
  describe('parseCpu', () => {
    it('should parse CPU value in millicores (m)', () => {
      expect(parseCpu('500m')).toBe(500)
      expect(parseCpu('1000m')).toBe(1000)
      expect(parseCpu('250m')).toBe(250)
    })

    it('should parse CPU value in cores (converts to millicores)', () => {
      expect(parseCpu('1')).toBe(1000)
      expect(parseCpu('2.5')).toBe(2500)
      expect(parseCpu('0.5')).toBe(500)
    })

    it('should return 0 for null or undefined', () => {
      expect(parseCpu(null)).toBe(0)
      expect(parseCpu(undefined)).toBe(0)
      expect(parseCpu('')).toBe(0)
    })
  })

  describe('parseMemory', () => {
    it('should parse memory in Ki (kibibytes) to bytes', () => {
      // parseMemory converts to bytes: 1024Ki = 1024 * 1024 = 1048576 bytes
      expect(parseMemory('1024Ki')).toBe(1024 * 1024)
      expect(parseMemory('512Ki')).toBe(512 * 1024)
    })

    it('should parse memory in Mi (mebibytes)', () => {
      expect(parseMemory('512Mi')).toBe(512 * 1024 * 1024)
      expect(parseMemory('1Mi')).toBe(1024 * 1024)
    })

    it('should parse memory in Gi (gibibytes)', () => {
      expect(parseMemory('2Gi')).toBe(2 * 1024 * 1024 * 1024)
      expect(parseMemory('0.5Gi')).toBe(0.5 * 1024 * 1024 * 1024)
    })

    it('should parse memory in Ti (tebibytes)', () => {
      expect(parseMemory('1Ti')).toBe(1024 * 1024 * 1024 * 1024)
    })

    it('should parse memory as bytes if no unit', () => {
      expect(parseMemory('1024')).toBe(1024)
      expect(parseMemory('2048')).toBe(2048)
    })

    it('should return 0 for null or undefined', () => {
      expect(parseMemory(null)).toBe(0)
      expect(parseMemory(undefined)).toBe(0)
      expect(parseMemory('')).toBe(0)
    })
  })

  describe('calculatePercentage', () => {
    it('should calculate percentage for CPU values', () => {
      expect(calculatePercentage('500m', '1000m')).toBe(50)
      expect(calculatePercentage('250m', '1000m')).toBe(25)
      expect(calculatePercentage('1000m', '1000m')).toBe(100)
    })

    it('should calculate percentage for memory values', () => {
      expect(calculatePercentage('512Mi', '1024Mi')).toBe(50)
      expect(calculatePercentage('256Mi', '1024Mi')).toBe(25)
      expect(calculatePercentage('1024Mi', '1024Mi')).toBe(100)
    })

    it('should return 0 if used or hard is missing', () => {
      expect(calculatePercentage(null, '1000m')).toBe(0)
      expect(calculatePercentage('500m', null)).toBe(0)
      expect(calculatePercentage(null, null)).toBe(0)
    })

    it('should return 0 if hard value is 0', () => {
      expect(calculatePercentage('500m', '0')).toBe(0)
    })

    it('should cap percentage at 100', () => {
      expect(calculatePercentage('2000m', '1000m')).toBe(100)
      expect(calculatePercentage('2048Mi', '1024Mi')).toBe(100)
    })
  })

  describe('formatResourceValue', () => {
    it('should return the value as-is', () => {
      expect(formatResourceValue('500m')).toBe('500m')
      expect(formatResourceValue('512Mi')).toBe('512Mi')
      expect(formatResourceValue('100')).toBe('100')
    })

    it('should return "-" for null or undefined', () => {
      expect(formatResourceValue(null)).toBe('-')
      expect(formatResourceValue(undefined)).toBe('-')
      expect(formatResourceValue('')).toBe('-')
    })
  })
})

