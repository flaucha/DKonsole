import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import Layout from './Layout';
import { BrowserRouter } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';

vi.mock('./Sidebar', () => ({
    default: () => <div data-testid="sidebar">Sidebar</div>
}));

vi.mock('./Header', () => ({
    default: () => <div data-testid="header">Header</div>
}));

// Mock SettingsContext
vi.mock('../context/SettingsContext', () => ({
    useSettings: () => ({ theme: 'dark', menuAnimation: true, menuAnimationSpeed: 300 })
}));

describe('Layout Component', () => {
    beforeEach(() => {
        global.fetch = vi.fn(() => Promise.resolve({
            ok: true,
            status: 200,
            json: () => Promise.resolve({}),
        }));
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('renders sidebar and header when logged in', async () => {
        const mockAuthFetch = vi.fn().mockResolvedValue({ ok: true, status: 200 });

        render(
            <BrowserRouter>
                <AuthContext.Provider value={{
                    isAuthenticated: true,
                    user: { username: 'testuser', role: 'admin' },
                    logout: vi.fn(),
                    setupStatus: { completed: true },
                    authFetch: mockAuthFetch
                }}>
                    <Layout />
                </AuthContext.Provider>
            </BrowserRouter>
        );

        expect(screen.getByTestId('sidebar')).toBeInTheDocument();
        expect(screen.getByTestId('header')).toBeInTheDocument();

        // Wait for the async checkAdmin and checkPermissions to complete
        // This avoids the "not wrapped in act(...)" warning and potential leaks
        await waitFor(() => {
            expect(mockAuthFetch).toHaveBeenCalled();
        });
    });
});
