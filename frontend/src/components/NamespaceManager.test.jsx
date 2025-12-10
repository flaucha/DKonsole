import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import NamespaceManager from './NamespaceManager';

// Mock hooks
vi.mock('../context/AuthContext', () => ({
    useAuth: () => ({ authFetch: vi.fn(), user: { role: 'admin' } })
}));

vi.mock('../context/SettingsContext', () => ({
    useSettings: () => ({ currentCluster: 'default' })
}));

vi.mock('../context/ToastContext', () => ({
    useToast: () => ({ success: vi.fn(), error: vi.fn() })
}));

vi.mock('../hooks/useNamespaces', () => ({
    useNamespaces: () => ({
        data: [
            { metadata: { name: 'default', uid: '1' }, status: 'Active', name: 'default' },
            { metadata: { name: 'kube-system', uid: '2' }, status: 'Active', name: 'kube-system' }
        ],
        isLoading: false,
        error: null,
        refetch: vi.fn()
    })
}));

describe('NamespaceManager Component', () => {
    it('renders namespace list', async () => {
        render(<NamespaceManager />);

        await waitFor(() => {
            expect(screen.getByText('default')).toBeInTheDocument();
            expect(screen.getByText('kube-system')).toBeInTheDocument();
        });
    });
});
