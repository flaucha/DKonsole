import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Lock, User, Key, RefreshCw } from 'lucide-react';
import defaultLogo from '../assets/logo-full.svg';
import { logger } from '../utils/logger';

const Setup = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [jwtSecret, setJwtSecret] = useState('');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [loading, setLoading] = useState(false);
    const [logoSrc, setLogoSrc] = useState(defaultLogo);
    const navigate = useNavigate();

    useEffect(() => {
        // Try to load custom logo from API (no auth required for logo endpoint)
        fetch(`/api/logo?t=${Date.now()}`)
            .then(res => {
                if (res.ok && res.status === 200) {
                    setLogoSrc(`/api/logo?t=${Date.now()}`);
                }
            })
            .catch(() => { });
    }, []);

    const handleLogoError = () => {
        if (logoSrc !== defaultLogo) {
            setLogoSrc(defaultLogo);
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
                        setSuccess(data.message || 'Setup completed successfully! The service has been reloaded and is ready to use.');
                        // Redirect to login after a delay
                        setTimeout(() => {
                            navigate('/login');
                        }, 3000);
                    } else {
                        setError(data.message || data.error || 'Failed to complete setup');
                    }
        } catch (err) {
            logger.error('Setup failed:', err);
            setError('An error occurred while completing setup');
        } finally {
            setLoading(false);
        }
    };

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
