import React, { createContext, useContext, useState, useEffect } from 'react';
import { logger } from '../utils/logger';

const AuthContext = createContext();

export const useAuth = () => useContext(AuthContext);

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        checkSession();
    }, []);

    const checkSession = async () => {
        try {
            const res = await fetch('/api/me');
            if (res.ok) {
                const data = await res.json();
                setUser(data);
            }
        } catch (error) {
            logger.error('Session check failed:', error);
        } finally {
            setLoading(false);
        }
    };

    const login = async (username, password) => {
        try {
            const res = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password }),
            });

            if (res.ok) {
                const data = await res.json();
                // Token is handled by HttpOnly cookie
                setUser({ username, role: data.role });
                return true;
            } else {
                throw new Error('Invalid credentials');
            }
        } catch (error) {
            logger.error('Login failed:', error);
            throw error;
        }
    };

    const logout = async () => {
        try {
            await fetch('/api/logout', { method: 'POST' });
        } catch (error) {
            logger.error('Logout failed:', error);
        }
        setUser(null);
    };

    const authFetch = async (url, options = {}) => {
        // No need to manually attach token, cookies handle it
        const res = await fetch(url, options);
        if (res.status === 401) {
            setUser(null);
            throw new Error('Session expired');
        }
        return res;
    };

    return (
        <AuthContext.Provider value={{ user, login, logout, loading, authFetch }}>
            {children}
        </AuthContext.Provider>
    );
};
