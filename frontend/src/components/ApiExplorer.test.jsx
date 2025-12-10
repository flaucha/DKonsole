import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import ApiExplorer from './ApiExplorer';
import { k8sApi } from '../api/k8sApi';

// Mock dependnecies
vi.mock('../api/k8sApi', () => ({
    k8sApi: {
        listAPIResources: vi.fn(),
        listAPIResourceObjects: vi.fn()
    }
}));

vi.mock('../context/AuthContext', () => ({
    useAuth: vi.fn(() => ({ authFetch: vi.fn().mockResolvedValue({ ok: true, json: async () => [] }), user: { role: 'admin' } }))
}));

vi.mock('../context/SettingsContext', () => ({
    useSettings: () => ({ currentCluster: 'default' })
}));

describe('ApiExplorer Component', () => {
    it('renders API explorer', () => {
        vi.mocked(k8sApi.listAPIResources).mockResolvedValue({
            groups: []
        });
        render(<ApiExplorer />);
        expect(screen.getByText(/API Explorer/i)).toBeInTheDocument();
    });

    // Note: Since ApiExplorer uses authFetch directly in useEffect instead of k8sApi (based on code reading in Step 374),
    // we mocked authFetch. The k8sApi mock might not be used if component uses authFetch directly.
    // Reading Step 374: `fetchApis` uses `authFetch('/api/apis?...')`.
    // So mocking k8sApi is irrelevant if it doesn't use it?
    // Step 346 ApiExplorer.test.jsx mocked k8sApi assuming it used it.
    // The component uses authFetch. So we mock authFetch response.

    it('loads and displays API groups', async () => {
        // Mock authFetch to return API list
        const authFetchMock = vi.fn().mockImplementation((url) => {
            if (url.includes('/api/apis')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => ([
                        { kind: 'Deployment', resource: 'deployments', group: 'apps', version: 'v1', namespaced: true }
                    ])
                });
            }
            return Promise.resolve({ ok: true, json: async () => ({}) });
        });

        // Re-mock useAuth with specific fetch
        vi.mocked(await import('../context/AuthContext')).useAuth.mockReturnValue({
            authFetch: authFetchMock,
            user: { role: 'admin' }
        });

        render(<ApiExplorer />);

        await waitFor(() => {
            expect(screen.getByText(/Deployment/i)).toBeInTheDocument();
        });
    });
});
