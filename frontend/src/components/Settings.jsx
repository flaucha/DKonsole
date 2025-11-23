import React, { useState } from 'react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { Server, Palette, Type, Plus, Check, AlertCircle, Trash2 } from 'lucide-react';

const Settings = () => {
    const {
        clusters, currentCluster, setCurrentCluster,
        theme, setTheme,
        font, setFont,
        fontSize, setFontSize,
        borderRadius, setBorderRadius,
        addCluster
    } = useSettings();
    const { authFetch } = useAuth();
    const [activeTab, setActiveTab] = useState('clusters');
    const [newCluster, setNewCluster] = useState({ name: '', host: '', token: '', insecure: false });
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    const handleAddCluster = async (e) => {
        e.preventDefault();
        setError('');
        setSuccess('');
        try {
            await addCluster(newCluster);
            setSuccess('Cluster added successfully!');
            setNewCluster({ name: '', host: '', token: '', insecure: false });
        } catch (err) {
            setError(err.message);
        }
    };

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
            </div>

            {activeTab === 'clusters' && (
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
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

                    <div className="space-y-6">
                        <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                            <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                <Plus size={20} className="mr-2 text-green-400" /> Add New Cluster
                            </h2>
                            <form onSubmit={handleAddCluster} className="space-y-4">
                                <div>
                                    <label className="block text-sm font-medium text-gray-400 mb-1">Cluster Name</label>
                                    <input
                                        type="text"
                                        value={newCluster.name}
                                        onChange={e => setNewCluster({ ...newCluster, name: e.target.value })}
                                        className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:outline-none focus:border-blue-500 transition-colors"
                                        placeholder="e.g., production-us-east"
                                        required
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-400 mb-1">API Server URL</label>
                                    <input
                                        type="text"
                                        value={newCluster.host}
                                        onChange={e => setNewCluster({ ...newCluster, host: e.target.value })}
                                        className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:outline-none focus:border-blue-500 transition-colors"
                                        placeholder="https://api.k8s.example.com"
                                        required
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-400 mb-1">Bearer Token</label>
                                    <textarea
                                        value={newCluster.token}
                                        onChange={e => setNewCluster({ ...newCluster, token: e.target.value })}
                                        className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:outline-none focus:border-blue-500 h-24 font-mono text-xs transition-colors"
                                        placeholder="ey..."
                                        required
                                    />
                                </div>
                                <div className="flex items-center">
                                    <input
                                        type="checkbox"
                                        id="insecure"
                                        checked={newCluster.insecure}
                                        onChange={e => setNewCluster({ ...newCluster, insecure: e.target.checked })}
                                        className="h-4 w-4 bg-gray-900 border-gray-700 rounded text-blue-500 focus:ring-0 cursor-pointer"
                                    />
                                    <label htmlFor="insecure" className="ml-2 text-sm text-gray-400 cursor-pointer select-none">Skip TLS Verification (Insecure)</label>
                                </div>

                                {error && (
                                    <div className="flex items-center text-red-400 text-sm bg-red-900/20 p-3 rounded border border-red-900/50">
                                        <AlertCircle size={16} className="mr-2 flex-shrink-0" /> {error}
                                    </div>
                                )}
                                {success && (
                                    <div className="flex items-center text-green-400 text-sm bg-green-900/20 p-3 rounded border border-green-900/50">
                                        <Check size={16} className="mr-2 flex-shrink-0" /> {success}
                                    </div>
                                )}

                                <button
                                    type="submit"
                                    className="flex items-center justify-center w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded transition-all hover:shadow-lg hover:shadow-blue-900/20"
                                >
                                    <Plus size={18} className="mr-2" /> Add Cluster
                                </button>
                            </form>
                        </div>
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
        </div>
    );
};

export default Settings;
