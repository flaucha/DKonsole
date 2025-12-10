import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import { Lock, User, Shield, Users } from 'lucide-react';
import logoFullDark from '../assets/logo-full-dark.png';
import logoFullLight from '../assets/logo-full-light.png';

const Login = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [ldapEnabled, setLdapEnabled] = useState(false);
    const [activeTab, setActiveTab] = useState('ldap'); // 'core' or 'ldap' - default to LDAP
    const { login } = useAuth();
    const navigate = useNavigate();

    // Get current theme and determine default logo immediately
    const currentTheme = localStorage.getItem('theme') || 'default';
    const isCreamTheme = currentTheme === 'cream';
    // Only support Cream or Dark (default). Light theme is removed.
    const defaultLogoSrc = isCreamTheme ? logoFullLight : logoFullDark;
    const [logoSrc, setLogoSrc] = useState(defaultLogoSrc); // Show default logo immediately

    useEffect(() => {
        // Try to load custom logo from API (no auth required for logo endpoint)
        // Check in background, but show default immediately
        // Use 'light' logo type for cream theme as they share similar brightness
        const logoType = isCreamTheme ? 'light' : 'normal';
        const timestamp = Date.now();
        fetch(`/api/logo?type=${logoType}&t=${timestamp}`)
            .then(res => {
                if (res.ok && res.status === 200) {
                    // Custom logo exists, use it
                    setLogoSrc(`/api/logo?type=${logoType}&t=${timestamp}`);
                }
                // If not found (404), keep default logo (already set)
            })
            .catch(() => {
                // On error, keep default logo (already set)
            });

        // Check if LDAP is enabled
        fetch('/api/ldap/status')
            .then(res => {
                if (res.ok) {
                    return res.json();
                }
            })
            .then(data => {
                if (data && data.enabled) {
                    setLdapEnabled(true);
                    setActiveTab('ldap'); // Default to LDAP tab when enabled
                } else {
                    setActiveTab('core'); // Default to core if LDAP is not enabled
                }
            })
            .catch(() => { });
    }, []);

    const handleLogoError = () => {
        // If logo fails to load, fallback to theme-appropriate default logo
        const currentTheme = localStorage.getItem('theme') || 'default';
        const isCreamTheme = currentTheme === 'cream';
        const fallbackLogo = isCreamTheme ? logoFullLight : logoFullDark;

        if (logoSrc !== fallbackLogo) {
            setLogoSrc(fallbackLogo);
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        try {
            // Determine IDP based on active tab
            const idp = activeTab === 'ldap' ? 'ldap' : 'core';
            await login(username, password, idp);
            navigate('/');
        } catch {
            setError('Invalid username or password');
        }
    };

    // Theme classes - now relying on CSS variables via data-theme
    // bg-gray-900 adapts: Dark (Dark), Light (White), Cream (Cream)
    // text-white adapts: Dark (White), Light (Dark), Cream (Brown)

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-900 text-white" data-theme={currentTheme}>
            <div className="bg-gray-800 border-gray-700 p-8 rounded-lg shadow-lg w-full max-w-md border">
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

                {error && (
                    <div className="bg-red-900/20 border border-red-900 text-red-400 px-4 py-3 rounded mb-6 text-sm">
                        {error}
                    </div>
                )}

                {/* Tabs for IDP selection when LDAP is enabled */}
                {ldapEnabled && (
                    <div className="flex space-x-1 border-b border-gray-700 mb-6">
                        <button
                            type="button"
                            className={`flex-1 pb-2 px-4 flex items-center justify-center font-medium transition-colors ${activeTab === 'ldap' ? 'border-b-2 border-blue-500 text-blue-500' : 'text-gray-400 hover:text-gray-300'}`}
                            onClick={() => setActiveTab('ldap')}
                        >
                            <Users size={18} className="mr-2" /> LDAP
                        </button>
                        <button
                            type="button"
                            className={`flex-1 pb-2 px-4 flex items-center justify-center font-medium transition-colors ${activeTab === 'core' ? 'border-b-2 border-blue-500 text-blue-500' : 'text-gray-400 hover:text-gray-300'}`}
                            onClick={() => setActiveTab('core')}
                        >
                            <Shield size={18} className="mr-2" /> CORE
                        </button>
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
                                className="block w-full pl-10 pr-3 py-2 border rounded-md leading-5 placeholder-gray-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 sm:text-sm bg-gray-900 text-gray-300 border-gray-700"
                                placeholder="Enter username"
                                required
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
                                className="block w-full pl-10 pr-3 py-2 border rounded-md leading-5 placeholder-gray-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 sm:text-sm bg-gray-900 text-gray-300 border-gray-700"
                                placeholder="Enter password"
                                required
                            />
                        </div>
                    </div>

                    <button
                        type="submit"
                        className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
                    >
                        Sign In
                    </button>
                </form>
            </div>
        </div>
    );
};

export default Login;
