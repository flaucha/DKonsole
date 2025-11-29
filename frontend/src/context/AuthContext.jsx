import React, { createContext, useContext, useState, useEffect } from 'react';
import { logger } from '../utils/logger';

const AuthContext = createContext();

export const useAuth = () => useContext(AuthContext);

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const [setupRequired, setSetupRequired] = useState(false);

    useEffect(() => {
        checkSetupStatus();
    }, []);

    const checkSetupStatus = async () => {
        try {
            const res = await fetch('/api/setup/status');
            if (res.ok) {
                const data = await res.json();
                setSetupRequired(data.setupRequired || false);

                // Only check session if setup is not required
                if (!data.setupRequired) {
                    await checkSession();
                } else {
                    setLoading(false);
                }
            } else {
                // If endpoint fails, assume setup is not required and check session
                await checkSession();
            }
        } catch (error) {
            logger.error('Setup status check failed:', error);
            // On error, assume setup is not required and check session
            await checkSession();
        }
    };

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

    const login = async (username, password, idp = '') => {
        try {
            const res = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password, idp }),
            });

            if (res.ok) {
                const data = await res.json();
                // Token is handled by HttpOnly cookie
                // After successful login, fetch full user data including permissions
                try {
                    const meRes = await fetch('/api/me');
                    if (meRes.ok) {
                        const meData = await meRes.json();
                        setUser(meData);
                    } else {
                        // Fallback to basic user data if /api/me fails
                        setUser({ username, role: data.role });
                    }
                } catch (meError) {
                    logger.error('Failed to fetch user data after login:', meError);
                    // Fallback to basic user data
                    setUser({ username, role: data.role });
                }
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
        <AuthContext.Provider value={{ user, login, logout, loading, authFetch, setupRequired, checkSetupStatus }}>
            {children}
        </AuthContext.Provider>
    );
};
