import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import Login from './Login';
import { BrowserRouter } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';

// Mock dependencies
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => vi.fn(),
        useLocation: () => ({ state: { from: { pathname: '/' } } }),
    };
});

describe('Login Component', () => {
    it('renders login form', () => {
        render(
            <BrowserRouter>
                <AuthContext.Provider value={{ login: vi.fn(), setupStatus: { completed: true } }}>
                    <Login />
                </AuthContext.Provider>
            </BrowserRouter>
        );

        // Check for username and password fields
        expect(screen.getByText(/username/i)).toBeInTheDocument();
        expect(screen.getByText(/password/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
    });

    it('handles input changes', () => {
        render(
            <BrowserRouter>
                <AuthContext.Provider value={{ login: vi.fn(), setupStatus: { completed: true } }}>
                    <Login />
                </AuthContext.Provider>
            </BrowserRouter>
        );

        const usernameInput = screen.getByPlaceholderText(/enter username/i);
        const passwordInput = screen.getByPlaceholderText(/enter password/i);

        fireEvent.change(usernameInput, { target: { value: 'testuser' } });
        fireEvent.change(passwordInput, { target: { value: 'password123' } });

        expect(usernameInput).toHaveValue('testuser');
        expect(passwordInput).toHaveValue('password123');
    });

    it('calls login function on submit', async () => {
        const loginMock = vi.fn().mockResolvedValue(true);
        render(
            <BrowserRouter>
                <AuthContext.Provider value={{ login: loginMock, setupStatus: { completed: true } }}>
                    <Login />
                </AuthContext.Provider>
            </BrowserRouter>
        );

        fireEvent.change(screen.getByPlaceholderText(/enter username/i), { target: { value: 'testuser' } });
        fireEvent.change(screen.getByPlaceholderText(/enter password/i), { target: { value: 'password123' } });
        fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

        // Adjusted expectation to allow extra arguments (like idp='ldap' default)
        expect(loginMock).toHaveBeenCalledWith('testuser', 'password123', expect.anything());
    });
});
