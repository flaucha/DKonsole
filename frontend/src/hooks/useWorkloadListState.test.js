import { renderHook, act } from '@testing-library/react';
import { vi } from 'vitest';
import { useWorkloadListState } from './useWorkloadListState';

// Mock dependencies
vi.mock('react-router-dom', () => ({
    useLocation: vi.fn(() => ({ search: '' }))
}));

vi.mock('../context/SettingsContext', () => ({
    useSettings: vi.fn(() => ({ currentCluster: 'test-cluster' }))
}));

vi.mock('../context/AuthContext', () => ({
    useAuth: vi.fn(() => ({ authFetch: vi.fn(), user: { username: 'test-user' } }))
}));

vi.mock('./useWorkloads', () => ({
    useWorkloads: vi.fn(() => ({
        data: [],
        isLoading: false,
        error: null,
        refetch: vi.fn()
    }))
}));

vi.mock('./useWorkloadActions', () => ({
    useWorkloadActions: vi.fn(() => ({
        setConfirmAction: vi.fn(),
        setConfirmRollout: vi.fn()
    }))
}));

import { useLocation } from 'react-router-dom';
import { useWorkloads } from './useWorkloads';

describe('useWorkloadListState', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('initializes with default state', () => {
        const { result } = renderHook(() => useWorkloadListState('default', 'Pod'));

        expect(result.current.resources).toEqual([]);
        expect(result.current.sortField).toBe('name');
        expect(result.current.sortDirection).toBe('asc');
        expect(result.current.filter).toBe('');
        expect(result.current.expandedId).toBeNull();
    });

    it('syncs data from useWorkloads', () => {
        const mockData = [{ uid: '1', name: 'pod-1' }];
        useWorkloads.mockReturnValue({
            data: mockData,
            isLoading: false,
            error: null,
            refetch: vi.fn()
        });

        const { result } = renderHook(() => useWorkloadListState('default', 'Pod'));

        expect(result.current.resources).toEqual(mockData);
    });

    it('handles sorting', () => {
        const { result } = renderHook(() => useWorkloadListState('default', 'Pod'));

        // Initial sort is name asc
        act(() => {
            result.current.handleSort('name');
        });
        // Click same field -> toggle desc
        expect(result.current.sortField).toBe('name');
        expect(result.current.sortDirection).toBe('desc');

        // Click different field -> field asc
        act(() => {
            result.current.handleSort('age');
        });
        expect(result.current.sortField).toBe('age');
        expect(result.current.sortDirection).toBe('asc');
    });

    it('toggles expand', () => {
        const { result } = renderHook(() => useWorkloadListState('default', 'Pod'));

        act(() => {
            result.current.toggleExpand('123');
        });
        expect(result.current.expandedId).toBe('123');

        act(() => {
            result.current.toggleExpand('123');
        });
        expect(result.current.expandedId).toBeNull();
    });

    it('handles add new resource', () => {
        const { result } = renderHook(() => useWorkloadListState('default', 'Pod'));

        act(() => {
            result.current.handleAdd();
        });

        expect(result.current.editingResource).toEqual({
            kind: 'Pod',
            namespaced: true,
            isNew: true,
            namespace: 'default',
            apiVersion: 'v1',
            metadata: {
                namespace: 'default',
                name: 'new-pod'
            }
        });
    });

    it('parses search param from URL', () => {
        useLocation.mockReturnValue({ search: '?search=pod-1' });
        const mockData = [{ uid: '1', name: 'pod-1' }];
        useWorkloads.mockReturnValue({
            data: mockData,
            isLoading: false,
            error: null,
            refetch: vi.fn()
        });

        const { result } = renderHook(() => useWorkloadListState('default', 'Pod'));

        expect(result.current.filter).toBe('pod-1');
        expect(result.current.expandedId).toBe('1');
    });
});
