import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Lock, User, Key, RefreshCw, Loader2 } from 'lucide-react';
import defaultLogo from '../assets/logo-full.svg';
import logoLight from '../assets/logo-light.svg';
import { logger } from '../utils/logger';

const Setup = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [jwtSecret, setJwtSecret] = useState('');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [loading, setLoading] = useState(false);
    const [reloading, setReloading] = useState(false);
    const [setupCompleted, setSetupCompleted] = useState(false);
    const [checkingStatus, setCheckingStatus] = useState(true);
    const [logoSrc, setLogoSrc] = useState(defaultLogo);
    const navigate = useNavigate();

    useEffect(() => {
        // Check if setup is already completed
        const checkSetupStatus = async () => {
            try {
                const res = await fetch('/api/setup/status');
                if (res.ok) {
                    const data = await res.json();
                    if (!data.setupRequired) {
                        // Setup is already completed
                        setSetupCompleted(true);
                    }
                }
            } catch (err) {
                logger.error('Failed to check setup status:', err);
            } finally {
                setCheckingStatus(false);
            }
        };

        checkSetupStatus();

        // Get current theme from localStorage
        const currentTheme = localStorage.getItem('theme') || 'default';
        const isLightTheme = currentTheme === 'light' || currentTheme === 'cream';
        // Use /logo-light.svg for light themes (served from static directory)
        const defaultLogoSrc = isLightTheme ? '/logo-light.svg' : defaultLogo;
        setLogoSrc(defaultLogoSrc);

        // Try to load custom logo from API (no auth required for logo endpoint)
        // Use logo-light for light themes, logo normal for dark themes
        // Add timestamp to prevent browser caching (use same timestamp for consistency)
        const logoType = isLightTheme ? 'light' : 'normal';
        const timestamp = Date.now();
        fetch(`/api/logo?type=${logoType}&t=${timestamp}`)
            .then(res => {
                if (res.ok && res.status === 200) {
                    setLogoSrc(`/api/logo?type=${logoType}&t=${timestamp}`);
                } else {
                    // No custom logo, use theme-appropriate default
                    setLogoSrc(defaultLogoSrc);
                }
            })
            .catch(() => {
                // On error, use theme-appropriate default
                setLogoSrc(defaultLogoSrc);
            });
    }, []);

    const handleLogoError = () => {
        // If logo fails to load, fallback to theme-appropriate default logo
        const currentTheme = localStorage.getItem('theme') || 'default';
        const isLightTheme = currentTheme === 'light' || currentTheme === 'cream';
        const fallbackLogo = isLightTheme ? '/logo-light.svg' : defaultLogo;

        if (logoSrc !== fallbackLogo) {
            setLogoSrc(fallbackLogo);
        }
    };

    const generateJWTSecret = () => {
        // Generate a secure random JWT secret (44 characters base64)
        const array = new Uint8Array(32);
        crypto.getRandomValues(array);
        const base64 = btoa(String.fromCharCode(...array));
        setJwtSecret(base64);
    };

    const validateForm = () => {
        if (!username.trim()) {
            setError('Username is required');
            return false;
        }

        if (!password) {
            setError('Password is required');
            return false;
        }

        if (password.length < 8) {
            setError('Password must be at least 8 characters long');
            return false;
        }

        if (password !== confirmPassword) {
            setError('Passwords do not match');
            return false;
        }

        if (jwtSecret && jwtSecret.length < 32) {
            setError('JWT secret must be at least 32 characters long');
            return false;
        }

        return true;
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        setSuccess('');

        if (!validateForm()) {
            return;
        }

        setLoading(true);

        try {
            const response = await fetch('/api/setup/complete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    username,
                    password,
                    jwtSecret: jwtSecret || undefined, // Send undefined if empty to trigger auto-generation
                }),
            });

            const data = await response.json();

            if (response.ok) {
                // Hide form and show reloading screen
                setReloading(true);
                setLoading(false);

                // Wait 5 seconds then navigate to login
                setTimeout(() => {
                    navigate('/login');
                }, 5000);
            } else {
                setError(data.message || data.error || 'Failed to complete setup');
                setLoading(false);
            }
        } catch (err) {
            logger.error('Setup failed:', err);
            setError('An error occurred while completing setup');
            setLoading(false);
        }
    };

    // Show loading while checking setup status
    if (checkingStatus) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-900">
                <div className="flex flex-col items-center space-y-4">
                    <Loader2 className="w-12 h-12 animate-spin text-blue-500" />
                    <p className="text-gray-400 text-sm">Checking setup status...</p>
                </div>
            </div>
        );
    }

    // Show "Setup completed" message if setup is already done
    if (setupCompleted) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-900">
                <div className="bg-gray-800 p-8 rounded-lg shadow-lg w-full max-w-md border border-gray-700 text-center">
                    <div className="text-center mb-8 flex justify-center items-center min-h-[80px]">
                        {logoSrc && (
                            <img
                                src={logoSrc}
                                alt="DKonsole Logo"
                                className="h-20 w-auto max-w-full object-contain"
                                onError={handleLogoError}
                                style={{ display: 'block' }}
                            />
                        )}
                    </div>
                    <div className="flex flex-col items-center space-y-4">
                        <div className="w-16 h-16 bg-green-900/20 rounded-full flex items-center justify-center">
                            <svg className="w-8 h-8 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                            </svg>
                        </div>
                        <h2 className="text-2xl font-bold text-white">Setup Completed</h2>
                        <p className="text-gray-400 text-sm">
                            The initial setup has already been completed. You can now log in to access DKonsole.
                        </p>
                        <button
                            onClick={() => navigate('/login')}
                            className="mt-4 w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
                        >
                            Go to Login
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    // Show reloading screen if setup was just completed
    if (reloading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-900">
                <div className="flex flex-col items-center space-y-4">
                    <Loader2 className="w-12 h-12 animate-spin text-blue-500" />
                    <p className="text-gray-300 text-lg">Setup completed successfully!</p>
                    <p className="text-gray-400 text-sm">Redirecting to login...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-900">
            <div className="bg-gray-800 p-8 rounded-lg shadow-lg w-full max-w-md border border-gray-700">
                <div className="text-center mb-8 flex justify-center items-center min-h-[80px]">
                    {logoSrc && (
                        <img
                            src={logoSrc}
                            alt="DKonsole Logo"
                            className="h-20 w-auto max-w-full object-contain"
                            onError={handleLogoError}
                            style={{ display: 'block' }}
                        />
                    )}
                </div>

                <h2 className="text-2xl font-bold text-white mb-2 text-center">Initial Setup</h2>
                <p className="text-sm text-gray-400 mb-6 text-center">
                    Configure your DKonsole administrator account
                </p>

                {error && (
                    <div className="bg-red-900/20 border border-red-900 text-red-400 px-4 py-3 rounded mb-6 text-sm">
                        {error}
                    </div>
                )}

                {success && (
                    <div className="bg-green-900/20 border border-green-900 text-green-400 px-4 py-3 rounded mb-6 text-sm">
                        {success}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-6">
                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">Username</label>
                        <div className="relative">
                            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                <User size={18} className="text-gray-500" />
                            </div>
                            <input
                                type="text"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                className="block w-full pl-10 pr-3 py-2 border border-gray-700 rounded-md leading-5 bg-gray-900 text-gray-300 placeholder-gray-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 sm:text-sm"
                                placeholder="Enter admin username"
                                required
                                disabled={loading}
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">Password</label>
                        <div className="relative">
                            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                <Lock size={18} className="text-gray-500" />
                            </div>
                            <input
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="block w-full pl-10 pr-3 py-2 border border-gray-700 rounded-md leading-5 bg-gray-900 text-gray-300 placeholder-gray-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 sm:text-sm"
                                placeholder="Enter password (min. 8 characters)"
                                required
                                disabled={loading}
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">Confirm Password</label>
                        <div className="relative">
                            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                <Lock size={18} className="text-gray-500" />
                            </div>
                            <input
                                type="password"
                                value={confirmPassword}
                                onChange={(e) => setConfirmPassword(e.target.value)}
                                className="block w-full pl-10 pr-3 py-2 border border-gray-700 rounded-md leading-5 bg-gray-900 text-gray-300 placeholder-gray-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 sm:text-sm"
                                placeholder="Confirm password"
                                required
                                disabled={loading}
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">
                            JWT Secret (optional)
                            <span className="text-xs text-gray-500 ml-2">Leave empty to auto-generate</span>
                        </label>
                        <div className="relative">
                            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                <Key size={18} className="text-gray-500" />
                            </div>
                            <input
                                type="text"
                                value={jwtSecret}
                                onChange={(e) => setJwtSecret(e.target.value)}
                                className="block w-full pl-10 pr-12 py-2 border border-gray-700 rounded-md leading-5 bg-gray-900 text-gray-300 placeholder-gray-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 sm:text-sm"
                                placeholder="Auto-generated if empty"
                                disabled={loading}
                            />
                            <button
                                type="button"
                                onClick={generateJWTSecret}
                                className="absolute inset-y-0 right-0 pr-3 flex items-center text-blue-500 hover:text-blue-400 transition-colors"
                                title="Generate random JWT secret"
                                disabled={loading}
                            >
                                <RefreshCw size={18} />
                            </button>
                        </div>
                        <p className="mt-1 text-xs text-gray-500">
                            Must be at least 32 characters long if provided manually
                        </p>
                    </div>

                    <button
                        type="submit"
                        disabled={loading}
                        className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {loading ? 'Completing Setup...' : 'Complete Setup'}
                    </button>
                </form>
            </div>
        </div>
    );
};

export default Setup;
