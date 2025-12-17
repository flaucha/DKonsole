import { renderHook, act } from '@testing-library/react';
import { useSetupLogic } from '../useSetupLogic';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useNavigate } from 'react-router-dom';

// Mock dependencies
vi.mock('react-router-dom', () => ({
    useNavigate: vi.fn(),
}));

vi.mock('../utils/logger', () => ({
    logger: {
        error: vi.fn(),
    },
}));

// Mock fetch
global.fetch = vi.fn();

describe('useSetupLogic', () => {
    const mockNavigate = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
        useNavigate.mockReturnValue(mockNavigate);

        // Default fetch mock for logo
        global.fetch.mockImplementation((url) => {
            if (url.includes('/api/logo')) {
                return Promise.resolve({ ok: false });
            }
            if (url === '/api/setup/status') {
                // Default to setup needed, no token update required
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({ setupRequired: true, tokenUpdateRequired: false })
                });
            }
            return Promise.resolve({ ok: false });
        });
    });

    it('should initialize with default states', async () => {
        const { result } = renderHook(() => useSetupLogic());
        expect(result.current.username).toBe('');
        expect(result.current.password).toBe('');
        expect(result.current.tokenOnlyMode).toBe(false);
    });

    it('should set tokenOnlyMode when backend reports tokenUpdateRequired', async () => {
        global.fetch.mockImplementation((url) => {
            if (url === '/api/setup/status') {
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({ setupRequired: false, tokenUpdateRequired: true })
                });
            }
            return Promise.resolve({ ok: false });
        });

        const { result, waitForNextUpdate } = renderHook(() => useSetupLogic());

        // Wait for effect
        await act(async () => {
            await new Promise(resolve => setTimeout(resolve, 0));
        });

        expect(result.current.tokenOnlyMode).toBe(true);
    });

    it('should validate service account token', async () => {
        const { result } = renderHook(() => useSetupLogic());

        await act(async () => {
            result.current.setUsername('admin');
            result.current.setPassword('password123');
            result.current.setConfirmPassword('password123');
            result.current.setServiceAccountToken(''); // Empty
            await result.current.handleSubmit({ preventDefault: vi.fn() });
        });

        expect(result.current.error).toBe('Service Account Token is required');
    });

    it('should call correct endpoint for token update', async () => {
        // Setup in token only mode
        global.fetch.mockImplementation((url) => {
            if (url === '/api/setup/status') {
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({ setupRequired: false, tokenUpdateRequired: true })
                });
            }
            // Setup token call
            if (url === '/api/setup/token') {
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({ message: 'Success' })
                });
            }
            return Promise.resolve({ ok: false });
        });

        const { result } = renderHook(() => useSetupLogic());

        await act(async () => {
            await new Promise(resolve => setTimeout(resolve, 0));
        });

        expect(result.current.tokenOnlyMode).toBe(true);

        // Submit token
        await act(async () => {
            result.current.setServiceAccountToken('new-token');
        });

        await act(async () => {
            await result.current.handleSubmit({ preventDefault: vi.fn() });
        });

        expect(global.fetch).toHaveBeenCalledWith('/api/setup/token', expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({ serviceAccountToken: 'new-token' })
        }));
    });
});
