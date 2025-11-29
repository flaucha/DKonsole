import React, { useState, useEffect } from 'react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import { Server, Palette, Type, Plus, Check, AlertCircle, Trash2, Settings as SettingsIcon, Info, Github, Mail, Coffee, Code, Lock, Save, Users, Key } from 'lucide-react';
import { parseErrorResponse, parseError } from '../utils/errorParser';

const Settings = () => {
    const {
        clusters, currentCluster, setCurrentCluster,
        theme, setTheme,
        font, setFont,
        fontSize, setFontSize,
        borderRadius, setBorderRadius
    } = useSettings();
    const { authFetch, user } = useAuth();
    const [activeTab, setActiveTab] = useState('clusters');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [isAdmin, setIsAdmin] = useState(false);
    const [checkingAdmin, setCheckingAdmin] = useState(true);

    useEffect(() => {
        // Check if user is admin (core admin or LDAP admin group member)
        // We check by trying to access a settings endpoint
        const checkAdmin = async () => {
            try {
                const res = await authFetch('/api/settings/prometheus/url');
                if (res.ok || res.status === 404) {
                    // If we can access it (or it doesn't exist), user is admin
                    setIsAdmin(true);
                } else if (res.status === 403) {
                    // Forbidden - user is not admin
                    setIsAdmin(false);
                } else {
                    // Other error - assume not admin for security
                    setIsAdmin(false);
                }
            } catch (err) {
                // Error accessing - assume not admin for security
                setIsAdmin(false);
            } finally {
                setCheckingAdmin(false);
            }
        };
        if (user) {
            checkAdmin();
        } else {
            setCheckingAdmin(false);
        }
    }, [authFetch, user]);

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
        { id: 'light', name: 'Light Mode', color: '#ffffff' },
        { id: 'cream', name: 'Cream', color: '#f5f5dc' },
    ];

    const fontGroups = {
        'Sans-Serif': ['Inter', 'Roboto', 'Open Sans', 'Lato', 'Montserrat', 'Source Sans Pro', 'Nunito', 'Quicksand', 'Raleway', 'Ubuntu', 'Oswald'],
        'Serif': ['Merriweather', 'Playfair Display'],
        'Monospace': ['Fira Code', 'Inconsolata'],
    };

    return (
        <div className="p-6 max-w-5xl mx-auto">
            <div className="flex justify-between items-center mb-6 h-10">
                <h1 className="text-2xl font-bold text-white">Settings</h1>
                <div className="h-8">
                    {activeTab === 'appearance' ? (
                        <button
                            onClick={handleResetDefaults}
                            className="flex items-center px-3 py-1.5 text-sm bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
                        >
                            <Trash2 size={14} className="mr-2" /> Reset Defaults
                        </button>
                    ) : (
                        <div className="h-8"></div>
                    )}
                </div>
            </div>

            <div className="flex space-x-1 border-b border-gray-700 mb-6 relative">
                <button
                    className={`pb-2 px-4 flex items-center font-medium transition-colors h-10 ${activeTab === 'clusters' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300 border-b-2 border-transparent'}`}
                    onClick={() => setActiveTab('clusters')}
                >
                    <Server size={18} className="mr-2" /> Clusters
                </button>
                <button
                    className={`pb-2 px-4 flex items-center font-medium transition-colors h-10 ${activeTab === 'appearance' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300 border-b-2 border-transparent'}`}
                    onClick={() => setActiveTab('appearance')}
                >
                    <Palette size={18} className="mr-2" /> Appearance
                </button>
                <button
                    className={`pb-2 px-4 flex items-center font-medium transition-colors h-10 ${activeTab === 'general' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300 border-b-2 border-transparent'}`}
                    onClick={() => setActiveTab('general')}
                >
                    <SettingsIcon size={18} className="mr-2" /> General
                </button>
                <button
                    className={`pb-2 px-4 flex items-center font-medium transition-colors h-10 ${activeTab === 'ldap' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300 border-b-2 border-transparent'}`}
                    onClick={() => setActiveTab('ldap')}
                >
                    <Users size={18} className="mr-2" /> LDAP
                </button>
                <button
                    className={`pb-2 px-4 flex items-center font-medium transition-colors h-10 ${activeTab === 'about' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300 border-b-2 border-transparent'}`}
                    onClick={() => setActiveTab('about')}
                >
                    <Info size={18} className="mr-2" /> About
                </button>
            </div>

            {activeTab === 'clusters' && (
                <div className="w-full space-y-6">
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
                <div className="w-full">
                    {checkingAdmin ? (
                        <div className="text-white">Checking permissions...</div>
                    ) : !isAdmin ? (
                        <div className="bg-red-900/20 border border-red-500/50 rounded-lg p-6 text-center">
                            <AlertCircle size={48} className="mx-auto mb-4 text-red-400" />
                            <h2 className="text-xl font-semibold text-white mb-2">Access Denied</h2>
                            <p className="text-gray-400">You need admin privileges to access general settings.</p>
                        </div>
                    ) : (
                        <div className="w-full space-y-6">
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
                </div>
            )}

            {activeTab === 'ldap' && (
                <div className="w-full">
                    {checkingAdmin ? (
                        <div className="text-white">Checking permissions...</div>
                    ) : !isAdmin ? (
                        <div className="bg-red-900/20 border border-red-500/50 rounded-lg p-6 text-center">
                            <AlertCircle size={48} className="mx-auto mb-4 text-red-400" />
                            <h2 className="text-xl font-semibold text-white mb-2">Access Denied</h2>
                            <p className="text-gray-400">You need admin privileges to access LDAP settings.</p>
                        </div>
                    ) : (
                        <LDAPSettings authFetch={authFetch} error={error} setError={setError} success={success} setSuccess={setSuccess} />
                    )}
                </div>
            )}

            {activeTab === 'appearance' && (
                <div className="w-full max-w-5xl mx-auto grid grid-cols-1 lg:grid-cols-3 gap-8">
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
                            <div className="space-y-6">
                                <div className="flex-1">
                                    <p className="text-sm text-gray-400 mb-4">
                                        Upload custom logos (PNG or SVG) to replace the default application logos. You can upload separate logos for dark and light themes.
                                    </p>
                                    <div className="space-y-4">
                                        <div>
                                            <label className="block text-sm text-gray-300 mb-2">Logo for Dark Themes</label>
                                            <label className="inline-flex items-center px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white text-sm font-medium rounded-lg cursor-pointer transition-colors border border-gray-600 hover:border-gray-500 shadow-sm">
                                                <Plus size={16} className="mr-2" /> Upload Dark Theme Logo
                                                <input
                                                    type="file"
                                                    className="hidden"
                                                    accept=".png,.svg"
                                                    onChange={(e) => {
                                                        const file = e.target.files[0];
                                                        if (!file) return;

                                                        const formData = new FormData();
                                                        formData.append('logo', file);
                                                        formData.append('type', 'normal');

                                                        authFetch('/api/logo', {
                                                            method: 'POST',
                                                            body: formData,
                                                        })
                                                            .then(async (res) => {
                                                                if (res.ok) {
                                                                    setSuccess('Logo uploaded successfully! Refreshing...');
                                                                    setTimeout(() => window.location.reload(), 1500);
                                                                } else {
                                                                    const errorText = await parseErrorResponse(res);
                                                                    throw new Error(errorText);
                                                                }
                                                            })
                                                            .catch((err) => setError(parseError(err)));
                                                    }}
                                                />
                                            </label>
                                        </div>
                                        <div>
                                            <label className="block text-sm text-gray-300 mb-2">Logo for Light Themes</label>
                                            <label className="inline-flex items-center px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white text-sm font-medium rounded-lg cursor-pointer transition-colors border border-gray-600 hover:border-gray-500 shadow-sm">
                                                <Plus size={16} className="mr-2" /> Upload Light Theme Logo
                                                <input
                                                    type="file"
                                                    className="hidden"
                                                    accept=".png,.svg"
                                                    onChange={(e) => {
                                                        const file = e.target.files[0];
                                                        if (!file) return;

                                                        const formData = new FormData();
                                                        formData.append('logo', file);
                                                        formData.append('type', 'light');

                                                        authFetch('/api/logo', {
                                                            method: 'POST',
                                                            body: formData,
                                                        })
                                                            .then(async (res) => {
                                                                if (res.ok) {
                                                                    setSuccess('Logo uploaded successfully! Refreshing...');
                                                                    setTimeout(() => window.location.reload(), 1500);
                                                                } else {
                                                                    const errorText = await parseErrorResponse(res);
                                                                    throw new Error(errorText);
                                                                }
                                                            })
                                                            .catch((err) => setError(parseError(err)));
                                                    }}
                                                />
                                            </label>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {activeTab === 'about' && (
                <div className="w-full space-y-6">
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
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to update Prometheus URL');
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to update Prometheus URL');
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
    const [showConfirmDialog, setShowConfirmDialog] = useState(false);

    const validateFields = () => {
        setError('');
        setSuccess('');

        // Validation
        if (!currentPassword || !newPassword || !confirmPassword) {
            setError('All fields are required');
            return false;
        }

        if (newPassword.length < 8) {
            setError('New password must be at least 8 characters long');
            return false;
        }

        if (newPassword !== confirmPassword) {
            setError('New passwords do not match');
            return false;
        }

        return true;
    };

    const handleChangePasswordClick = () => {
        if (validateFields()) {
            setShowConfirmDialog(true);
        }
    };

    const handleConfirmChangePassword = async () => {
        setLoading(true);
        setShowConfirmDialog(false);

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
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to change password');
                setLoading(false);
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to change password');
            setLoading(false);
        }
    };

    return (
        <div className="space-y-4">
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
                onClick={handleChangePasswordClick}
                disabled={loading}
                className="w-full flex items-center justify-center px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
            >
                <Lock size={16} className="mr-2" />
                {loading ? 'Changing Password...' : 'Change Password'}
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

            {/* Confirmation Dialog */}
            {showConfirmDialog && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                        <div className="flex items-center space-x-3 mb-4 text-red-400">
                            <AlertCircle size={24} />
                            <h3 className="text-xl font-bold text-white">Confirm Password Change</h3>
                        </div>
                        <p className="text-gray-300 mb-2 leading-relaxed">
                            Are you sure you want to change your password?
                        </p>
                        <p className="text-sm text-yellow-400 mb-6">
                            ⚠️ Warning: After changing your password, you will be automatically logged out and redirected to the login page.
                        </p>
                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setShowConfirmDialog(false)}
                                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleConfirmChangePassword}
                                className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors"
                            >
                                Change Password
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

// LDAP Settings Component
const LDAPSettings = ({ authFetch, error, setError, success, setSuccess }) => {
    const [config, setConfig] = useState({
        insecureSkipVerify: false,
        caCert: '',
        enabled: false,
        url: '',
        baseDN: '',
        userDN: 'uid',
        groupDN: '',
        userFilter: '',
        requiredGroup: ''
    });
    const [credentials, setCredentials] = useState({
        username: '',
        password: ''
    });
    const [savedCredentials, setSavedCredentials] = useState({
        username: '',
        password: '*****'
    });
    const [groups, setGroups] = useState({ groups: [] });
    const [loading, setLoading] = useState(false);
    const [testLoading, setTestLoading] = useState(false);
    const [testResult, setTestResult] = useState(null); // null, 'success', 'error'
    const [testMessage, setTestMessage] = useState('');
    const [namespaces, setNamespaces] = useState([]);
    const [ldapActiveTab, setLdapActiveTab] = useState('config');
    const [adminGroups, setAdminGroups] = useState([]);

    useEffect(() => {
        loadConfig();
        loadGroups();
        loadNamespaces();
        loadCredentials();
    }, [authFetch]);

    const loadConfig = async () => {
        try {
            const res = await authFetch('/api/ldap/config');
            if (res.ok) {
                const data = await res.json();
                setConfig(data);
                // Load admin groups from config
                if (data.adminGroups && Array.isArray(data.adminGroups)) {
                    setAdminGroups(data.adminGroups);
                } else {
                    setAdminGroups([]);
                }
            }
        } catch (err) {
            // Ignore errors - LDAP might not be configured
        }
    };

    const loadGroups = async () => {
        try {
            const res = await authFetch('/api/ldap/groups');
            if (res.ok) {
                const data = await res.json();
                setGroups(data);
            }
        } catch (err) {
            // Ignore errors
        }
    };

    const loadNamespaces = async () => {
        try {
            const res = await authFetch('/api/namespaces');
            if (res.ok) {
                const data = await res.json();
                setNamespaces(data.map(ns => ns.name || ns));
            }
        } catch (err) {
            // Ignore errors
        }
    };

    const loadCredentials = async () => {
        try {
            const res = await authFetch('/api/ldap/credentials');
            if (res.ok) {
                const data = await res.json();
                if (data.configured && data.username) {
                    setSavedCredentials({
                        username: data.username,
                        password: '*****'
                    });
                    // Pre-fill the form with saved username and masked password
                    setCredentials({
                        username: data.username,
                        password: '*****'
                    });
                }
            }
        } catch (err) {
            // Ignore errors
        }
    };

    const handleSaveConfig = async () => {
        setError('');
        setSuccess('');
        setLoading(true);

        try {
            const res = await authFetch('/api/ldap/config', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ config }),
            });

            if (res.ok) {
                setSuccess('LDAP configuration saved successfully!');
            } else {
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to save LDAP configuration');
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to save LDAP configuration');
        } finally {
            setLoading(false);
        }
    };

    const handleEnabledChange = async (enabled) => {
        const newConfig = { ...config, enabled };
        setConfig(newConfig);

        // Auto-save when enabled checkbox changes
        setError('');
        setSuccess('');
        setLoading(true);

        try {
            const res = await authFetch('/api/ldap/config', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ config: newConfig }),
            });

            if (res.ok) {
                setSuccess('LDAP configuration saved successfully!');
            } else {
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to save LDAP configuration');
                // Revert on error
                setConfig(config);
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to save LDAP configuration');
            // Revert on error
            setConfig(config);
        } finally {
            setLoading(false);
        }
    };

    const handleSaveCredentials = async () => {
        setError('');
        setSuccess('');
        setLoading(true);

        // If password is masked, only send username (password stays the same)
        const credsToSend = {
            username: credentials.username,
            password: credentials.password === '*****' ? '' : credentials.password
        };

        // If username hasn't changed and password is masked, nothing to update
        if (credsToSend.password === '' && credentials.username === savedCredentials.username) {
            setSuccess('No changes to save');
            setLoading(false);
            return;
        }

        try {
            const res = await authFetch('/api/ldap/credentials', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(credsToSend),
            });

            if (res.ok) {
                const responseText = await res.text();
                // Check if it's the "no changes" message
                if (responseText.includes('No changes')) {
                    setSuccess('No changes to save');
                } else {
                    setSuccess('LDAP credentials saved successfully!');
                }
                // Update saved credentials display
                setSavedCredentials({
                    username: credentials.username,
                    password: '*****'
                });
                // Keep username, mask password
                setCredentials({
                    username: credentials.username,
                    password: '*****'
                });
            } else {
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to save LDAP credentials');
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to save LDAP credentials');
        } finally {
            setLoading(false);
        }
    };

    const handleTestConnection = async () => {
        setError('');
        setSuccess('');
        setTestLoading(true);
        setTestResult(null);
        setTestMessage('');

        // Use actual password if not masked, otherwise we can't test
        const testPassword = credentials.password === '*****' ? '' : credentials.password;

        if (!testPassword) {
            setError('Please enter the password to test connection');
            setTestLoading(false);
            return;
        }

        try {
            const res = await authFetch('/api/ldap/test', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    url: config.url,
                    baseDN: config.baseDN,
                    userDN: config.userDN,
                    username: credentials.username,
                    password: testPassword,
                }),
            });

            if (res.ok) {
                setTestResult('success');
                setTestMessage('Connection test successful');
            } else {
                const errorText = await parseErrorResponse(res);
                setTestResult('error');
                setTestMessage(errorText || 'LDAP connection test failed');
            }
        } catch (err) {
            setTestResult('error');
            setTestMessage(parseError(err) || 'LDAP connection test failed');
        } finally {
            setTestLoading(false);
        }
    };

    const handleSaveGroups = async () => {
        setError('');
        setSuccess('');
        setLoading(true);

        // Filter out incomplete permissions (empty namespace) before sending
        const groupsToSend = {
            groups: {
                groups: groups.groups.map(group => ({
                    ...group,
                    permissions: group.permissions.filter(perm => perm.namespace !== '')
                }))
            }
        };

        try {
            const res = await authFetch('/api/ldap/groups', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(groupsToSend),
            });

            if (res.ok) {
                setSuccess('LDAP groups configuration saved successfully!');
                // Update local state to remove incomplete permissions
                setGroups(groupsToSend.groups);
            } else {
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to save LDAP groups');
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to save LDAP groups');
        } finally {
            setLoading(false);
        }
    };

    const addGroup = () => {
        setGroups({
            groups: [...groups.groups, { name: '', permissions: [] }]
        });
    };

    const removeGroup = (index) => {
        setGroups({
            groups: groups.groups.filter((_, i) => i !== index)
        });
    };

    const updateGroup = (index, field, value) => {
        const newGroups = [...groups.groups];
        newGroups[index] = { ...newGroups[index], [field]: value };
        setGroups({ groups: newGroups });
    };

    const addPermission = (groupIndex) => {
        const newGroups = [...groups.groups];
        newGroups[groupIndex].permissions = [
            ...newGroups[groupIndex].permissions,
            { namespace: '', permission: 'view' }
        ];
        setGroups({ groups: newGroups });
    };

    const removePermission = (groupIndex, permIndex) => {
        const newGroups = [...groups.groups];
        newGroups[groupIndex].permissions = newGroups[groupIndex].permissions.filter((_, i) => i !== permIndex);
        setGroups({ groups: newGroups });
    };

    const updatePermission = (groupIndex, permIndex, field, value) => {
        const newGroups = [...groups.groups];
        newGroups[groupIndex].permissions[permIndex] = {
            ...newGroups[groupIndex].permissions[permIndex],
            [field]: value
        };
        setGroups({ groups: newGroups });
    };

    return (
        <div className="space-y-6">
            {/* LDAP Main Container */}
            <div className="bg-gray-800 rounded-lg border border-gray-700 shadow-lg">
                {/* LDAP Tabs */}
                <div className="flex space-x-1 border-b border-gray-700 px-6 pt-4">
                    <button
                        className={`pb-3 px-4 flex items-center font-medium transition-colors ${ldapActiveTab === 'config' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                        onClick={() => setLdapActiveTab('config')}
                    >
                        <SettingsIcon size={18} className="mr-2" /> Config
                    </button>
                    <button
                        className={`pb-3 px-4 flex items-center font-medium transition-colors ${ldapActiveTab === 'permissions' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                        onClick={() => setLdapActiveTab('permissions')}
                    >
                        <Users size={18} className="mr-2" /> Group Permissions
                    </button>
                    <button
                        className={`pb-3 px-4 flex items-center font-medium transition-colors ${ldapActiveTab === 'admins' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                        onClick={() => setLdapActiveTab('admins')}
                    >
                        <Key size={18} className="mr-2" /> Group Admins
                    </button>
                </div>

                {/* Tab Content */}
                <div className="p-6">
                    {ldapActiveTab === 'config' && (
                        <div className="space-y-6">
                            {/* LDAP Server Configuration */}
                            <div>
                                <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                    <Server size={20} className="mr-2 text-blue-400" /> LDAP Server Configuration
                                </h2>
                                <div className="space-y-4">
                                    <div className="flex items-center space-x-3">
                                        <input
                                            type="checkbox"
                                            id="ldap-enabled"
                                            checked={config.enabled}
                                            onChange={(e) => handleEnabledChange(e.target.checked)}
                                            className="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
                                        />
                                        <label htmlFor="ldap-enabled" className="text-gray-300">Enable LDAP Authentication</label>
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">LDAP URL</label>
                                        <input
                                            type="text"
                                            value={config.url}
                                            onChange={(e) => setConfig({ ...config, url: e.target.value })}
                                            placeholder="ldap://ldap.example.com:389"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">Base DN</label>
                                        <input
                                            type="text"
                                            value={config.baseDN}
                                            onChange={(e) => setConfig({ ...config, baseDN: e.target.value })}
                                            placeholder="dc=example,dc=com"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">User DN Attribute</label>
                                        <input
                                            type="text"
                                            value={config.userDN}
                                            onChange={(e) => setConfig({ ...config, userDN: e.target.value })}
                                            placeholder="uid"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">Group DN</label>
                                        <input
                                            type="text"
                                            value={config.groupDN}
                                            onChange={(e) => setConfig({ ...config, groupDN: e.target.value })}
                                            placeholder="ou=groups,dc=example,dc=com"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">Required Group (Optional)</label>
                                        <input
                                            type="text"
                                            value={config.requiredGroup || ''}
                                            onChange={(e) => setConfig({ ...config, requiredGroup: e.target.value })}
                                            placeholder="access_group (leave empty to allow all authenticated users)"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                        <p className="text-xs text-gray-500 mt-1">
                                            Only users in this group will be allowed to login. Leave empty to allow all authenticated LDAP users.
                                        </p>
                                    </div>

                                    {/* TLS Configuration */}
                                    <div className="border-t border-gray-700 pt-4">
                                        <h3 className="text-md font-semibold text-white mb-3 flex items-center">
                                            <Lock size={18} className="mr-2 text-yellow-400" /> TLS Configuration
                                        </h3>

                                        <div className="flex items-center space-x-3 mb-4">
                                            <input
                                                type="checkbox"
                                                id="ldap-insecure-skip-verify"
                                                checked={config.insecureSkipVerify || false}
                                                onChange={(e) => setConfig({ ...config, insecureSkipVerify: e.target.checked })}
                                                className="w-4 h-4 text-yellow-600 bg-gray-700 border-gray-600 rounded focus:ring-yellow-500"
                                            />
                                            <label htmlFor="ldap-insecure-skip-verify" className="text-gray-300">
                                                Skip TLS Certificate Verification
                                            </label>
                                        </div>
                                        {config.insecureSkipVerify && (
                                            <div className="mb-4 p-3 bg-yellow-900/20 border border-yellow-500/50 rounded-lg">
                                                <div className="flex items-start space-x-2">
                                                    <AlertCircle size={16} className="text-yellow-400 mt-0.5 flex-shrink-0" />
                                                    <p className="text-xs text-yellow-400">
                                                        <strong>Warning:</strong> This option disables TLS certificate verification.
                                                        This should only be used in development or with self-signed certificates.
                                                        It makes the connection vulnerable to man-in-the-middle attacks.
                                                    </p>
                                                </div>
                                            </div>
                                        )}

                                        <div>
                                            <label className="block text-sm font-medium text-gray-400 mb-2">CA Certificate (PEM format, optional)</label>
                                            <textarea
                                                value={config.caCert || ''}
                                                onChange={(e) => setConfig({ ...config, caCert: e.target.value })}
                                                placeholder="-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
                                                rows={6}
                                                className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 font-mono text-xs"
                                            />
                                            <p className="text-xs text-gray-500 mt-1">
                                                Paste the CA certificate in PEM format for TLS verification. Leave empty to use system CA certificates.
                                            </p>
                                        </div>
                                    </div>

                                    <button
                                        onClick={handleSaveConfig}
                                        disabled={loading}
                                        className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                                    >
                                        <Save size={16} className="mr-2" />
                                        {loading ? 'Saving...' : 'Save Configuration'}
                                    </button>
                                </div>
                            </div>

                            {/* LDAP Service Account Credentials */}
                            <div className="border-t border-gray-700 pt-6">
                                <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                    <Key size={20} className="mr-2 text-yellow-400" /> LDAP Service Account Credentials
                                </h2>
                                <p className="text-sm text-gray-400 mb-4">
                                    Credentials for the service account used to query LDAP for user groups.
                                </p>
                                <div className="space-y-4">
                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">Username</label>
                                        <input
                                            type="text"
                                            value={credentials.username}
                                            onChange={(e) => setCredentials({ ...credentials, username: e.target.value })}
                                            placeholder="service-account"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">Password</label>
                                        <input
                                            type="password"
                                            value={credentials.password}
                                            onChange={(e) => {
                                                setCredentials({ ...credentials, password: e.target.value });
                                            }}
                                            onFocus={(e) => {
                                                // Clear mask when user focuses on the field
                                                if (credentials.password === '*****') {
                                                    setCredentials({ ...credentials, password: '' });
                                                }
                                            }}
                                            placeholder={savedCredentials.username ? "Enter new password or leave blank to keep current" : "Enter password"}
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div className="flex items-center space-x-2">
                                        <button
                                            onClick={handleSaveCredentials}
                                            disabled={loading}
                                            className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                                        >
                                            <Save size={16} className="mr-2" />
                                            {loading ? 'Saving...' : 'Save Credentials'}
                                        </button>
                                        <button
                                            onClick={handleTestConnection}
                                            disabled={testLoading || !config.url || !credentials.username || !credentials.password}
                                            className="flex items-center px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                                        >
                                            <Check size={16} className="mr-2" />
                                            {testLoading ? 'Testing...' : 'Test Connection'}
                                        </button>
                                        {testResult && (
                                            <div className={`flex items-center space-x-2 px-3 py-2 rounded-lg ${
                                                testResult === 'success'
                                                    ? 'bg-green-900/20 border border-green-500/50 text-green-400'
                                                    : 'bg-red-900/20 border border-red-500/50 text-red-400'
                                            }`}>
                                                {testResult === 'success' ? (
                                                    <>
                                                        <Check size={16} />
                                                        <span className="text-sm">{testMessage}</span>
                                                    </>
                                                ) : (
                                                    <>
                                                        <AlertCircle size={16} />
                                                        <span className="text-sm">{testMessage}</span>
                                                    </>
                                                )}
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}

                    {ldapActiveTab === 'permissions' && (
                        <div>
                            <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                <Users size={20} className="mr-2 text-purple-400" /> LDAP Groups and Permissions
                            </h2>
                            <p className="text-sm text-gray-400 mb-4">
                                Configure which LDAP groups have access to which namespaces and with what permissions.
                            </p>

                            <div className="space-y-4">
                                {groups.groups.map((group, groupIndex) => (
                                    <div key={groupIndex} className="bg-gray-900 p-4 rounded-lg border border-gray-700">
                                        <div className="flex items-center justify-between mb-3">
                                            <input
                                                type="text"
                                                value={group.name}
                                                onChange={(e) => updateGroup(groupIndex, 'name', e.target.value)}
                                                placeholder="Group name (e.g., developers)"
                                                className="flex-1 px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                            />
                                            <button
                                                onClick={() => removeGroup(groupIndex)}
                                                className="ml-2 px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
                                            >
                                                <Trash2 size={16} />
                                            </button>
                                        </div>

                                        <div className="space-y-2">
                                            {group.permissions.map((perm, permIndex) => (
                                                <div key={permIndex} className="flex items-center space-x-2">
                                                    <select
                                                        value={perm.namespace}
                                                        onChange={(e) => updatePermission(groupIndex, permIndex, 'namespace', e.target.value)}
                                                        className="flex-1 px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-blue-500"
                                                    >
                                                        <option value="">Select namespace</option>
                                                        {namespaces.map(ns => (
                                                            <option key={ns} value={ns}>{ns}</option>
                                                        ))}
                                                    </select>
                                                    <select
                                                        value={perm.permission}
                                                        onChange={(e) => updatePermission(groupIndex, permIndex, 'permission', e.target.value)}
                                                        className="px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-blue-500"
                                                    >
                                                        <option value="view">View</option>
                                                        <option value="edit">Edit</option>
                                                    </select>
                                                    <button
                                                        onClick={() => removePermission(groupIndex, permIndex)}
                                                        className="px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
                                                    >
                                                        <Trash2 size={14} />
                                                    </button>
                                                </div>
                                            ))}
                                            <button
                                                onClick={() => addPermission(groupIndex)}
                                                className="flex items-center px-3 py-2 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded-lg transition-colors text-sm"
                                            >
                                                <Plus size={14} className="mr-2" /> Add Permission
                                            </button>
                                        </div>
                                    </div>
                                ))}

                                <button
                                    onClick={addGroup}
                                    className="flex items-center px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded-lg transition-colors"
                                >
                                    <Plus size={16} className="mr-2" /> Add Group
                                </button>

                                <button
                                    onClick={handleSaveGroups}
                                    disabled={loading}
                                    className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                                >
                                    <Save size={16} className="mr-2" />
                                    {loading ? 'Saving...' : 'Save Groups Configuration'}
                                </button>
                            </div>
                        </div>
                    )}

                    {ldapActiveTab === 'admins' && (
                        <div>
                            <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                <Key size={20} className="mr-2 text-yellow-400" /> LDAP Group Admins
                            </h2>
                            <p className="text-sm text-gray-400 mb-4">
                                Configure which LDAP groups have cluster admin access. Members of these groups will have full access to all namespaces and settings.
                            </p>

                            <div className="space-y-4">
                                {adminGroups.map((group, index) => (
                                    <div key={index} className="flex items-center space-x-2 bg-gray-900 p-4 rounded-lg border border-gray-700">
                                        <input
                                            type="text"
                                            value={group}
                                            onChange={(e) => {
                                                const newGroups = [...adminGroups];
                                                newGroups[index] = e.target.value;
                                                setAdminGroups(newGroups);
                                            }}
                                            placeholder="Group name (e.g., admin)"
                                            className="flex-1 px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                        <button
                                            onClick={() => {
                                                const newGroups = adminGroups.filter((_, i) => i !== index);
                                                setAdminGroups(newGroups);
                                            }}
                                            className="px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
                                        >
                                            <Trash2 size={16} />
                                        </button>
                                    </div>
                                ))}

                                <button
                                    onClick={() => setAdminGroups([...adminGroups, ''])}
                                    className="flex items-center px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors"
                                >
                                    <Plus size={16} className="mr-2" /> Add Admin Group
                                </button>

                                <button
                                    onClick={async () => {
                                        setError('');
                                        setSuccess('');
                                        setLoading(true);

                                        try {
                                            // Update config with admin groups
                                            const updatedConfig = { ...config, adminGroups: adminGroups.filter(g => g.trim() !== '') };
                                            const res = await authFetch('/api/ldap/config', {
                                                method: 'PUT',
                                                headers: { 'Content-Type': 'application/json' },
                                                body: JSON.stringify({ config: updatedConfig }),
                                            });

                                            if (res.ok) {
                                                setSuccess('Admin groups updated successfully!');
                                                setConfig(updatedConfig);
                                            } else {
                                                const errorText = await parseErrorResponse(res);
                                                setError(errorText || 'Failed to update admin groups');
                                            }
                                        } catch (err) {
                                            setError(parseError(err) || 'Failed to update admin groups');
                                        } finally {
                                            setLoading(false);
                                        }
                                    }}
                                    disabled={loading}
                                    className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                                >
                                    <Save size={16} className="mr-2" />
                                    {loading ? 'Saving...' : 'Save Admin Groups'}
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>

            {error && (
                <div className="flex items-center text-red-400 text-sm bg-red-900/20 border border-red-500/50 rounded-lg p-3">
                    <AlertCircle size={16} className="mr-2" />
                    {error}
                </div>
            )}
            {success && (
                <div className="flex items-center text-green-400 text-sm bg-green-900/20 border border-green-500/50 rounded-lg p-3">
                    <Check size={16} className="mr-2" />
                    {success}
                </div>
            )}
        </div>
    );
};

export default Settings;
