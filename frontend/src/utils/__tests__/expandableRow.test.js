import { describe, it, expect } from 'vitest'
import {
  getExpandableRowClasses,
  getExpandableCellClasses,
  getExpandableRowStyles,
  getExpandableRowRowClasses
} from '../expandableRow'

describe('expandableRow', () => {
  describe('getExpandableRowClasses', () => {
    it('should return expanded classes when isExpanded is true', () => {
      const result = getExpandableRowClasses(true)
      expect(result).toContain('transition-all')
      expect(result).toContain('duration-300')
      expect(result).toContain('ease-in-out')
      expect(result).toContain('pl-12')
      expect(result).toContain('opacity-100')
      expect(result).toContain('max-h-[80vh]')
      expect(result).toContain('overflow-y-auto')
    })

    it('should return collapsed classes when isExpanded is false', () => {
      const result = getExpandableRowClasses(false)
      expect(result).toContain('transition-all')
      expect(result).toContain('duration-300')
      expect(result).toContain('ease-in-out')
      expect(result).toContain('pl-12')
      expect(result).toContain('max-h-0')
      expect(result).toContain('opacity-0')
      expect(result).toContain('overflow-hidden')
    })

    it('should not include padding when hasPadding is false', () => {
      const result = getExpandableRowClasses(true, false)
      expect(result).not.toContain('pl-12')
    })

    it('should include padding by default', () => {
      const result = getExpandableRowClasses(true)
      expect(result).toContain('pl-12')
    })
  })

  describe('getExpandableCellClasses', () => {
    it('should return base classes with border when expanded', () => {
      const result = getExpandableCellClasses(true, 5)
      expect(result).toContain('px-6')
      expect(result).toContain('pt-0')
      expect(result).toContain('bg-gray-800')
      expect(result).toContain('border-0')
      expect(result).toContain('border-b')
      expect(result).toContain('border-gray-700')
    })

    it('should return base classes without border when collapsed', () => {
      const result = getExpandableCellClasses(false, 5)
      expect(result).toContain('px-6')
      expect(result).toContain('pt-0')
      expect(result).toContain('bg-gray-800')
      expect(result).toContain('border-0')
      expect(result).not.toContain('border-b')
      expect(result).not.toContain('border-gray-700')
    })
  })

  describe('getExpandableRowStyles', () => {
    it('should return empty object when not expanded', () => {
      const result = getExpandableRowStyles(false)
      expect(result).toEqual({})
    })

    it('should return custom max height when provided', () => {
      const result = getExpandableRowStyles(true, null, '500px')
      expect(result).toEqual({ maxHeight: '500px' })
    })

    it('should return Pod-specific max height for Pod kind', () => {
      const result = getExpandableRowStyles(true, 'Pod')
      expect(result).toEqual({ maxHeight: 'calc(100vh - 250px)' })
    })

    it('should return empty object for non-Pod kind when expanded', () => {
      const result = getExpandableRowStyles(true, 'Deployment')
      expect(result).toEqual({})
    })

    it('should prioritize custom max height over Pod kind', () => {
      const result = getExpandableRowStyles(true, 'Pod', '600px')
      expect(result).toEqual({ maxHeight: '600px' })
    })
  })

  describe('getExpandableRowRowClasses', () => {
    it('should return expanded classes when isExpanded is true', () => {
      const result = getExpandableRowRowClasses(true)
      expect(result).toContain('group')
      expect(result).toContain('hover:bg-gray-800/50')
      expect(result).toContain('transition-colors')
      expect(result).toContain('cursor-pointer')
      expect(result).toContain('bg-gray-800/30')
    })

    it('should return collapsed classes when isExpanded is false', () => {
      const result = getExpandableRowRowClasses(false)
      expect(result).toContain('group')
      expect(result).toContain('hover:bg-gray-800/50')
      expect(result).toContain('transition-colors')
      expect(result).toContain('cursor-pointer')
      expect(result).not.toContain('bg-gray-800/30')
    })
  })
})





