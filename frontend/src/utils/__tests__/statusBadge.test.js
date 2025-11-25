import { describe, it, expect } from 'vitest'
import { getStatusBadgeClass } from '../statusBadge'

describe('statusBadge', () => {
  describe('getStatusBadgeClass', () => {
    it('should return blue badge for null or undefined status', () => {
      const expected = 'bg-blue-900/50 text-blue-300 border border-blue-700'
      expect(getStatusBadgeClass(null)).toBe(expected)
      expect(getStatusBadgeClass(undefined)).toBe(expected)
    })

    describe('Green badges (Ready/Healthy states)', () => {
      const greenStatuses = [
        'ready', 'Ready', 'READY',
        'running', 'Running', 'RUNNING',
        'active', 'Active', 'ACTIVE',
        'available', 'Available', 'AVAILABLE',
        'bound', 'Bound', 'BOUND',
        'succeeded', 'Succeeded', 'SUCCEEDED',
        'completed', 'Completed', 'COMPLETED',
        'healthy', 'Healthy', 'HEALTHY',
        'ok', 'Ok', 'OK'
      ]

      it.each(greenStatuses)('should return green badge for "%s"', (status) => {
        const result = getStatusBadgeClass(status)
        expect(result).toContain('bg-green-900/50')
        expect(result).toContain('text-green-300')
        expect(result).toContain('border-green-700')
      })
    })

    describe('Red badges (Error/Failed/Critical states)', () => {
      const redStatuses = [
        'error', 'Error', 'ERROR',
        'failed', 'Failed', 'FAILED',
        'crashloopbackoff', 'CrashLoopBackOff',
        'terminated', 'Terminated',
        'notready', 'NotReady',
        'oomkilled', 'OOMKilled',
        'evicted', 'Evicted',
        'unhealthy', 'Unhealthy',
        'deadlineexceeded', 'DeadlineExceeded',
        'outofmemory', 'OutOfMemory',
        'invalid', 'Invalid',
        'rejected', 'Rejected',
        'unknown', 'Unknown'
      ]

      it.each(redStatuses)('should return red badge for "%s"', (status) => {
        const result = getStatusBadgeClass(status)
        expect(result).toContain('bg-red-900/50')
        expect(result).toContain('text-red-300')
        expect(result).toContain('border-red-700')
      })
    })

    describe('Yellow badges (Warning states)', () => {
      const yellowStatuses = [
        'warning', 'Warning', 'WARNING',
        'pending', 'Pending', 'PENDING',
        'imagepullbackoff', 'ImagePullBackOff',
        'errimagepull', 'ErrImagePull',
        'schedulingdisabled', 'SchedulingDisabled',
        'unschedulable', 'Unschedulable',
        'degraded', 'Degraded',
        'partial', 'Partial',
        'terminating', 'Terminating'
      ]

      it.each(yellowStatuses)('should return yellow badge for "%s"', (status) => {
        const result = getStatusBadgeClass(status)
        expect(result).toContain('bg-yellow-900/50')
        expect(result).toContain('text-yellow-300')
        expect(result).toContain('border-yellow-700')
      })
    })

    describe('Blue badges (Informative/Transitional states)', () => {
      const blueStatuses = [
        'containercreating', 'ContainerCreating',
        'init', 'Init', 'INIT',
        'podinitializing', 'PodInitializing',
        'waiting', 'Waiting',
        'creating', 'Creating',
        'updating', 'Updating',
        'deleting', 'Deleting',
        'scaling', 'Scaling',
        'provisioning', 'Provisioning',
        'reconciling', 'Reconciling'
      ]

      it.each(blueStatuses)('should return blue badge for "%s"', (status) => {
        const result = getStatusBadgeClass(status)
        expect(result).toContain('bg-blue-900/50')
        expect(result).toContain('text-blue-300')
        expect(result).toContain('border-blue-700')
      })
    })

    it('should return blue badge for unknown status values', () => {
      const result = getStatusBadgeClass('some-unknown-status')
      expect(result).toContain('bg-blue-900/50')
      expect(result).toContain('text-blue-300')
      expect(result).toContain('border-blue-700')
    })
  })
})





