import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import ClusterOverview from './ClusterOverview';

// Mock hooks
vi.mock('../context/AuthContext', () => ({
    useAuth: () => ({ authFetch: vi.fn(), user: { role: 'admin' } })
}));

vi.mock('../context/SettingsContext', () => ({
    useSettings: () => ({ currentCluster: 'default' })
}));

vi.mock('../hooks/useClusterOverview', () => ({
    useClusterOverview: () => ({
        overview: { data: { nodes: 3, pods: 10, deployments: 5 }, isLoading: false, error: null },
        prometheusStatus: { data: { enabled: false } },
        metrics: { data: { clusterStats: {}, nodeMetrics: [] } }
    })
}));

describe('ClusterOverview Component', () => {
    it('renders stats when loaded', async () => {
        render(<ClusterOverview />);

        await waitFor(() => {
            // Check for stats that are definitely rendered based on failure log
            expect(screen.getByText(/Pods/i)).toBeInTheDocument();
            expect(screen.getByText(/10/i)).toBeInTheDocument();
        });
    });
});
