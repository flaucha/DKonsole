import { describe, it, expect } from 'vitest';

/**
 * Test utility function that mirrors the sorting logic for 'ready' field
 * from WorkloadList component
 */
const getReadyValue = (item, kind) => {
    if (kind !== 'Pod' || !item.details?.ready) return 0;
    const readyStr = item.details.ready.toString();
    const parts = readyStr.split('/');
    if (parts.length === 2) {
        const ready = parseFloat(parts[0]) || 0;
        const total = parseFloat(parts[1]) || 0;
        return total > 0 ? (ready / total) : 0;
    }
    return 0;
};

describe('Ready column sorting logic', () => {
    describe('getReadyValue', () => {
        it('should return 0 for non-Pod resources', () => {
            const item = {
                kind: 'Deployment',
                details: { ready: '2/3' }
            };
            expect(getReadyValue(item, 'Deployment')).toBe(0);
        });

        it('should return 0 when ready is missing', () => {
            const item = {
                kind: 'Pod',
                details: {}
            };
            expect(getReadyValue(item, 'Pod')).toBe(0);
        });

        it('should calculate correct ratio for ready containers', () => {
            const testCases = [
                { ready: '0/3', expected: 0 },
                { ready: '1/3', expected: 1/3 },
                { ready: '2/3', expected: 2/3 },
                { ready: '3/3', expected: 1 },
                { ready: '1/1', expected: 1 },
                { ready: '0/1', expected: 0 },
            ];

            testCases.forEach(({ ready, expected }) => {
                const item = {
                    kind: 'Pod',
                    details: { ready }
                };
                expect(getReadyValue(item, 'Pod')).toBeCloseTo(expected, 5);
            });
        });

        it('should handle zero total containers', () => {
            const item = {
                kind: 'Pod',
                details: { ready: '0/0' }
            };
            expect(getReadyValue(item, 'Pod')).toBe(0);
        });

        it('should handle invalid format gracefully', () => {
            const invalidFormats = [
                'invalid',
                '2',
                '/3',
                '2/',
                '2/3/4',
                '',
            ];

            invalidFormats.forEach(format => {
                const item = {
                    kind: 'Pod',
                    details: { ready: format }
                };
                expect(getReadyValue(item, 'Pod')).toBe(0);
            });
        });

        it('should handle non-numeric values in ready count', () => {
            const item = {
                kind: 'Pod',
                details: { ready: 'abc/3' }
            };
            expect(getReadyValue(item, 'Pod')).toBe(0);
        });

        it('should handle non-numeric values in total count', () => {
            const item = {
                kind: 'Pod',
                details: { ready: '2/abc' }
            };
            expect(getReadyValue(item, 'Pod')).toBe(0);
        });

        it('should sort correctly by ready ratio', () => {
            const pods = [
                { name: 'pod-0', details: { ready: '0/3' } },    // 0.0
                { name: 'pod-1', details: { ready: '3/3' } },    // 1.0
                { name: 'pod-2', details: { ready: '1/3' } },    // 0.333
                { name: 'pod-3', details: { ready: '2/3' } },    // 0.667
            ];

            // Sort by ready ratio ascending
            pods.sort((a, b) => {
                const valA = getReadyValue(a, 'Pod');
                const valB = getReadyValue(b, 'Pod');
                return valA - valB;
            });

            // Verify sorted order
            expect(pods[0].name).toBe('pod-0');
            expect(pods[1].name).toBe('pod-2');
            expect(pods[2].name).toBe('pod-3');
            expect(pods[3].name).toBe('pod-1');
        });

        it('should handle missing details gracefully', () => {
            const item = {
                kind: 'Pod',
                details: null
            };
            expect(getReadyValue(item, 'Pod')).toBe(0);
        });

        it('should handle edge case of all containers ready', () => {
            const item = {
                kind: 'Pod',
                details: { ready: '5/5' }
            };
            expect(getReadyValue(item, 'Pod')).toBe(1);
        });
    });
});
