import React, { useState, useEffect } from 'react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import { Server, Palette, Type, Plus, Check, AlertCircle, Trash2, Settings as SettingsIcon, Info, Github, Mail, Coffee, Code, Lock, Save } from 'lucide-react';

const Settings = () => {
    const {
        clusters, currentCluster, setCurrentCluster,
        theme, setTheme,
        font, setFont,
        fontSize, setFontSize,
        borderRadius, setBorderRadius
    } = useSettings();
    const { authFetch } = useAuth();
    const [activeTab, setActiveTab] = useState('clusters');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    const handleResetDefaults = () => {
        setTheme('default');
        setFont('Inter');
        setFontSize('normal');
        setBorderRadius('md');
    };

    const themes = [
        { id: 'default', name: 'Default (Dark)', color: '#1f2937' },
        { id: 'ocean', name: 'Ocean Blue', color: '#0f172a' },
        { id: 'forest', name: 'Forest Green', color: '#064e3b' },
        { id: 'sunset', name: 'Sunset Orange', color: '#7c2d12' },
        { id: 'midnight', name: 'Midnight Purple', color: '#150524' },
        { id: 'nebula', name: 'Nebula Pink', color: '#0a0a1e' },
        { id: 'dracula', name: 'Dracula', color: '#282a36' },
        { id: 'cyberpunk', name: 'Cyberpunk', color: '#000000' },
    ];

    const fontGroups = {
        'Sans-Serif': ['Inter', 'Roboto', 'Open Sans', 'Lato', 'Montserrat', 'Source Sans Pro', 'Nunito', 'Quicksand', 'Raleway', 'Ubuntu', 'Oswald'],
        'Serif': ['Merriweather', 'Playfair Display'],
        'Monospace': ['Fira Code', 'Inconsolata'],
    };

    return (
        <div className="p-6 max-w-5xl mx-auto">
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-2xl font-bold text-white">Settings</h1>
                {activeTab === 'appearance' && (
                    <button
                        onClick={handleResetDefaults}
                        className="flex items-center px-3 py-1.5 text-sm bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
                    >
                        <Trash2 size={14} className="mr-2" /> Reset Defaults
                    </button>
                )}
            </div>

            <div className="flex space-x-1 border-b border-gray-700 mb-6">
                <button
                    className={`pb-2 px-4 flex items-center font-medium transition-colors ${activeTab === 'clusters' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('clusters')}
                >
                    <Server size={18} className="mr-2" /> Clusters
                </button>
                <button
                    className={`pb-2 px-4 flex items-center font-medium transition-colors ${activeTab === 'appearance' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('appearance')}
                >
                    <Palette size={18} className="mr-2" /> Appearance
                </button>
                <button
                    className={`pb-2 px-4 flex items-center font-medium transition-colors ${activeTab === 'general' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('general')}
                >
                    <SettingsIcon size={18} className="mr-2" /> General
                </button>
                <button
                    className={`pb-2 px-4 flex items-center font-medium transition-colors ${activeTab === 'about' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('about')}
                >
                    <Info size={18} className="mr-2" /> About
                </button>
            </div>

            {activeTab === 'clusters' && (
                <div className="space-y-6">
                    <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                        <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                            <Server size={20} className="mr-2 text-blue-400" /> Configured Clusters
                        </h2>
                        <div className="space-y-3">
                            {clusters.map(cluster => (
                                <div key={cluster} className={`flex items-center justify-between p-4 rounded-lg border transition-all ${cluster === currentCluster ? 'bg-blue-900/20 border-blue-500/50' : 'bg-gray-750 border-gray-700'}`}>
                                    <div className="flex items-center">
                                        <div className={`w-2 h-2 rounded-full mr-3 ${cluster === currentCluster ? 'bg-green-400 shadow-[0_0_8px_rgba(74,222,128,0.5)]' : 'bg-gray-500'}`}></div>
                                        <span className={`font-medium ${cluster === currentCluster ? 'text-white' : 'text-gray-300'}`}>{cluster}</span>
                                    </div>
                                    {cluster !== currentCluster ? (
                                        <button
                                            onClick={() => setCurrentCluster(cluster)}
                                            className="text-xs bg-gray-700 hover:bg-gray-600 text-blue-300 px-3 py-1.5 rounded transition-colors"
                                        >
                                            Switch
                                        </button>
                                    ) : (
                                        <span className="text-xs bg-blue-500/20 text-blue-300 px-3 py-1 rounded-full border border-blue-500/30">Active</span>
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            )}

            {activeTab === 'general' && (
                <div className="space-y-6">
                    {/* Prometheus URL Settings */}
                    <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                        <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                            <Server size={20} className="mr-2 text-blue-400" /> Prometheus Configuration
                        </h2>
                        <PrometheusURLSettings authFetch={authFetch} error={error} setError={setError} success={success} setSuccess={setSuccess} />
                    </div>

                    {/* Password Change Settings */}
                    <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                        <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                            <SettingsIcon size={20} className="mr-2 text-red-400" /> Change Password
                        </h2>
                        <PasswordChangeSettings authFetch={authFetch} error={error} setError={setError} success={success} setSuccess={setSuccess} />
                    </div>
                </div>
            )}

            {activeTab === 'appearance' && (
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                    {/* Theme Column */}
                    <div className="lg:col-span-1 space-y-6">
                        <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg h-full">
                            <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                <Palette size={20} className="mr-2 text-purple-400" /> Color Theme
                            </h2>
                            <div className="space-y-2">
                                {themes.map(t => (
                                    <button
                                        key={t.id}
                                        onClick={() => setTheme(t.id)}
                                        className={`w-full flex items-center justify-between p-3 rounded-lg border transition-all ${theme === t.id ? 'border-blue-500 bg-blue-900/20 shadow-md' : 'border-gray-700 bg-gray-750 hover:bg-gray-700 hover:border-gray-600'}`}
                                    >
                                        <div className="flex items-center">
                                            <div className="w-8 h-8 rounded-full mr-3 border-2 border-gray-600 shadow-sm" style={{ backgroundColor: t.color }}></div>
                                            <span className={`font-medium ${theme === t.id ? 'text-white' : 'text-gray-300'}`}>{t.name}</span>
                                        </div>
                                        {theme === t.id && <Check size={18} className="text-blue-400" />}
                                    </button>
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* Font & UI Column */}
                    <div className="lg:col-span-2 space-y-6">
                        {/* UI Preferences */}
                        <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                            <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                <Type size={20} className="mr-2 text-yellow-400" /> Interface Preferences
                            </h2>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div>
                                    <label className="block text-sm font-medium text-gray-400 mb-2">Font Size</label>
                                    <div className="flex bg-gray-900 p-1 rounded-lg border border-gray-700">
                                        {['small', 'normal', 'large', 'xl'].map((size) => (
                                            <button
                                                key={size}
                                                onClick={() => setFontSize(size)}
                                                className={`flex-1 py-1.5 text-sm rounded-md transition-all ${fontSize === size ? 'bg-gray-700 text-white shadow' : 'text-gray-400 hover:text-gray-200'}`}
                                            >
                                                {size.charAt(0).toUpperCase() + size.slice(1)}
                                            </button>
                                        ))}
                                    </div>
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-400 mb-2">Border Radius</label>
                                    <div className="flex bg-gray-900 p-1 rounded-lg border border-gray-700">
                                        {['none', 'sm', 'md', 'lg'].map((radius) => (
                                            <button
                                                key={radius}
                                                onClick={() => setBorderRadius(radius)}
                                                className={`flex-1 py-1.5 text-sm rounded-md transition-all ${borderRadius === radius ? 'bg-gray-700 text-white shadow' : 'text-gray-400 hover:text-gray-200'}`}
                                            >
                                                {radius === 'none' ? 'Square' : radius.toUpperCase()}
                                            </button>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* Font Family */}
                        <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                            <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                <Type size={20} className="mr-2 text-blue-400" /> Font Family
                            </h2>
                            <div className="space-y-6">
                                {Object.entries(fontGroups).map(([group, groupFonts]) => (
                                    <div key={group}>
                                        <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 ml-1">{group}</h3>
                                        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                                            {groupFonts.map(f => (
                                                <button
                                                    key={f}
                                                    onClick={() => setFont(f)}
                                                    className={`flex items-center justify-center p-3 rounded-lg border transition-all text-center ${font === f ? 'border-blue-500 bg-blue-900/20 text-white shadow-md' : 'border-gray-700 bg-gray-750 text-gray-300 hover:bg-gray-700 hover:border-gray-600'}`}
                                                >
                                                    <span style={{ fontFamily: f }}>{f}</span>
                                                </button>
                                            ))}
                                        </div>
                                    </div>
                                ))}
                            </div>
                            <p className="mt-4 text-xs text-gray-500 flex items-center">
                                <AlertCircle size={12} className="mr-1" /> Fonts are loaded dynamically from Google Fonts.
                            </p>
                        </div>

                        {/* Custom Logo */}
                        <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                            <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                <Server size={20} className="mr-2 text-pink-400" /> Custom Logo
                            </h2>
                            <div className="flex items-start space-x-4">
                                <div className="flex-1">
                                    <p className="text-sm text-gray-400 mb-4">
                                        Upload a custom logo (PNG or SVG) to replace the default application logo.
                                    </p>
                                    <label className="inline-flex items-center px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white text-sm font-medium rounded-lg cursor-pointer transition-colors border border-gray-600 hover:border-gray-500 shadow-sm">
                                        <Plus size={16} className="mr-2" /> Upload New Logo
                                        <input
                                            type="file"
                                            className="hidden"
                                            accept=".png,.svg"
                                            onChange={(e) => {
                                                const file = e.target.files[0];
                                                if (!file) return;

                                                const formData = new FormData();
                                                formData.append('logo', file);

                                                authFetch('/api/logo', {
                                                    method: 'POST',
                                                    body: formData,
                                                })
                                                    .then(async (res) => {
                                                        if (res.ok) {
                                                            setSuccess('Logo uploaded successfully! Refreshing...');
                                                            setTimeout(() => window.location.reload(), 1500);
                                                        } else {
                                                            throw new Error(await res.text());
                                                        }
                                                    })
                                                    .catch((err) => setError(err.message));
                                            }}
                                        />
                                    </label>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {activeTab === 'about' && (
                <div className="space-y-6">
                    <div className="bg-gray-800 p-8 rounded-lg border border-gray-700 shadow-lg">
                        <div className="flex items-center mb-6">
                            <Code size={32} className="mr-3 text-blue-400" />
                            <h2 className="text-3xl font-bold text-white">DKonsole</h2>
                        </div>

                        <p className="text-gray-300 mb-6 text-lg">
                            A modern, lightweight Kubernetes dashboard built entirely with <strong>Artificial Intelligence</strong>.
                        </p>

                        <div className="space-y-4 mb-8">
                            <div className="flex items-center text-gray-300">
                                <span className="font-semibold mr-2">Version:</span>
                                <span className="px-2 py-1 bg-blue-900/30 text-blue-300 rounded text-sm font-mono">{import.meta.env.VITE_APP_VERSION || '1.1.6'}</span>
                            </div>

                            <div className="flex items-center space-x-4">
                                <a
                                    href="https://github.com/flaucha/DKonsole"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="flex items-center px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-200 rounded-md transition-colors"
                                >
                                    <Github size={18} className="mr-2" />
                                    GitHub
                                </a>

                                <a
                                    href="mailto:flaucha@gmail.com"
                                    className="flex items-center px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-200 rounded-md transition-colors"
                                >
                                    <Mail size={18} className="mr-2" />
                                    Email
                                </a>

                                <a
                                    href="https://buymeacoffee.com/flaucha"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="flex items-center px-4 py-2 bg-yellow-600 hover:bg-yellow-700 text-white rounded-md transition-colors"
                                >
                                    <Coffee size={18} className="mr-2" />
                                    Buy me a coffee
                                </a>
                            </div>
                        </div>

                        <div className="border-t border-gray-700 pt-6">
                            <h3 className="text-lg font-semibold text-white mb-4">Features</h3>
                            <ul className="space-y-2 text-gray-300">
                                <li className="flex items-start">
                                    <span className="text-blue-400 mr-2">•</span>
                                    <span>Resource Management: View and manage Deployments, Pods, Services, ConfigMaps, Secrets, and more</span>
                                </li>
                                <li className="flex items-start">
                                    <span className="text-blue-400 mr-2">•</span>
                                    <span>Prometheus Integration: Historical metrics for Pods with customizable time ranges</span>
                                </li>
                                <li className="flex items-start">
                                    <span className="text-blue-400 mr-2">•</span>
                                    <span>Live Logs: Stream logs from containers in real-time</span>
                                </li>
                                <li className="flex items-start">
                                    <span className="text-blue-400 mr-2">•</span>
                                    <span>Terminal Access: Execute commands directly in pod containers</span>
                                </li>
                                <li className="flex items-start">
                                    <span className="text-blue-400 mr-2">•</span>
                                    <span>YAML Editor: Edit resources with a built-in YAML editor</span>
                                </li>
                                <li className="flex items-start">
                                    <span className="text-blue-400 mr-2">•</span>
                                    <span>Secure Authentication: Argon2 password hashing and JWT-based sessions</span>
                                </li>
                                <li className="flex items-start">
                                    <span className="text-blue-400 mr-2">•</span>
                                    <span>Multi-Cluster Support: Manage multiple Kubernetes clusters from a single interface</span>
                                </li>
                            </ul>
                        </div>

                        <div className="border-t border-gray-700 pt-6 mt-6">
                            <p className="text-sm text-gray-500">
                                DKonsole is licensed under the MIT License.
                            </p>
                        </div>
                    </div>
                </div>
            )}

        </div>
    );
};

// Prometheus URL Settings Component
const PrometheusURLSettings = ({ authFetch, error, setError, success, setSuccess }) => {
    const [prometheusURL, setPrometheusURL] = useState('');
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        // Load current Prometheus URL
        authFetch('/api/settings/prometheus/url')
            .then(async (res) => {
                if (res.ok) {
                    const data = await res.json();
                    setPrometheusURL(data.url || '');
                }
            })
            .catch(() => {});
    }, [authFetch]);

    const handleSave = async () => {
        setError('');
        setSuccess('');
        setLoading(true);

        try {
            const res = await authFetch('/api/settings/prometheus/url', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ url: prometheusURL }),
            });

            if (res.ok) {
                setSuccess('Prometheus URL updated successfully! Reloading app...');
                setTimeout(() => window.location.reload(), 1500);
            } else {
                const errorText = await res.text();
                setError(errorText || 'Failed to update Prometheus URL');
            }
        } catch (err) {
            setError(err.message || 'Failed to update Prometheus URL');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="space-y-4">
            <p className="text-sm text-gray-400 mb-4">
                Configure the Prometheus endpoint URL for historical metrics. Leave empty to disable.
            </p>
            <div className="flex items-center space-x-2">
                <input
                    type="text"
                    value={prometheusURL}
                    onChange={(e) => setPrometheusURL(e.target.value)}
                    placeholder="http://prometheus-server.monitoring.svc.cluster.local:9090"
                    className="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                />
                <button
                    onClick={handleSave}
                    disabled={loading}
                    className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                >
                    <Save size={16} className="mr-2" />
                    {loading ? 'Saving...' : 'Save'}
                </button>
            </div>
            {error && (
                <div className="flex items-center text-red-400 text-sm">
                    <AlertCircle size={16} className="mr-2" />
                    {error}
                </div>
            )}
            {success && (
                <div className="flex items-center text-green-400 text-sm">
                    <Check size={16} className="mr-2" />
                    {success}
                </div>
            )}
        </div>
    );
};

// Password Change Settings Component
const PasswordChangeSettings = ({ authFetch, error, setError, success, setSuccess }) => {
    const navigate = useNavigate();
    const [currentPassword, setCurrentPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [showWarning, setShowWarning] = useState(false);

    const handleChangePassword = async () => {
        setError('');
        setSuccess('');

        // Validation
        if (!currentPassword || !newPassword || !confirmPassword) {
            setError('All fields are required');
            return;
        }

        if (newPassword.length < 8) {
            setError('New password must be at least 8 characters long');
            return;
        }

        if (newPassword !== confirmPassword) {
            setError('New passwords do not match');
            return;
        }

        if (!showWarning) {
            setShowWarning(true);
            return;
        }

        setLoading(true);

        try {
            const res = await authFetch('/api/auth/change-password', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    currentPassword,
                    newPassword,
                }),
            });

            if (res.ok) {
                setSuccess('Password changed successfully! You will be logged out...');
                // Logout and redirect to login
                setTimeout(() => {
                    // Clear auth token
                    document.cookie = 'token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
                    // Reload to trigger login
                    window.location.href = '/login';
                }, 2000);
            } else {
                const errorText = await res.text();
                setError(errorText || 'Failed to change password');
                setLoading(false);
            }
        } catch (err) {
            setError(err.message || 'Failed to change password');
            setLoading(false);
        }
    };

    return (
        <div className="space-y-4">
            {showWarning && (
                <div className="bg-yellow-900/20 border border-yellow-600/50 rounded-lg p-4 mb-4">
                    <div className="flex items-start">
                        <AlertCircle size={20} className="text-yellow-400 mr-3 mt-0.5" />
                        <div>
                            <h4 className="text-yellow-400 font-semibold mb-1">Warning: You will be logged out</h4>
                            <p className="text-yellow-300/80 text-sm">
                                After changing your password, you will be automatically logged out and redirected to the login page.
                            </p>
                        </div>
                    </div>
                </div>
            )}

            <div className="space-y-3">
                <div>
                    <label className="block text-sm font-medium text-gray-400 mb-2">Current Password</label>
                    <div className="relative">
                        <Lock size={18} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500" />
                        <input
                            type="password"
                            value={currentPassword}
                            onChange={(e) => setCurrentPassword(e.target.value)}
                            placeholder="Enter current password"
                            className="w-full pl-10 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                        />
                    </div>
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-400 mb-2">New Password</label>
                    <div className="relative">
                        <Lock size={18} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500" />
                        <input
                            type="password"
                            value={newPassword}
                            onChange={(e) => setNewPassword(e.target.value)}
                            placeholder="Enter new password (min 8 characters)"
                            className="w-full pl-10 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                        />
                    </div>
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-400 mb-2">Confirm New Password</label>
                    <div className="relative">
                        <Lock size={18} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500" />
                        <input
                            type="password"
                            value={confirmPassword}
                            onChange={(e) => setConfirmPassword(e.target.value)}
                            placeholder="Confirm new password"
                            className="w-full pl-10 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                        />
                    </div>
                </div>
            </div>

            <button
                onClick={handleChangePassword}
                disabled={loading}
                className="w-full flex items-center justify-center px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
            >
                <Lock size={16} className="mr-2" />
                {loading ? 'Changing Password...' : showWarning ? 'Confirm Password Change' : 'Change Password'}
            </button>

            {error && (
                <div className="flex items-center text-red-400 text-sm">
                    <AlertCircle size={16} className="mr-2" />
                    {error}
                </div>
            )}
            {success && (
                <div className="flex items-center text-green-400 text-sm">
                    <Check size={16} className="mr-2" />
                    {success}
                </div>
            )}
        </div>
    );
};

export default Settings;
