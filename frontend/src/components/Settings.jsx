import React, { useState } from 'react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { Server, Palette, Type, Plus, Check, AlertCircle, Trash2 } from 'lucide-react';

const Settings = () => {
    const { clusters, currentCluster, setCurrentCluster, theme, setTheme, font, setFont, addCluster } = useSettings();
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

    const fonts = [
        'Inter',
        'Roboto',
        'Open Sans',
        'Lato',
        'Montserrat',
        'Poppins',
        'Source Sans Pro',
        'Ubuntu',
        'Merriweather',
        'Playfair Display',
        'Nunito',
        'Raleway',
        'Quicksand',
        'Inconsolata',
        'Fira Code',
        'Oswald',
    ];

    return (
        <div className="p-6 max-w-4xl mx-auto">
            <h1 className="text-2xl font-bold text-white mb-6">Settings</h1>

            <div className="flex space-x-4 border-b border-gray-700 mb-6">
                <button
                    className={`pb-2 px-4 flex items-center ${activeTab === 'clusters' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('clusters')}
                >
                    <Server size={18} className="mr-2" /> Clusters
                </button>
                <button
                    className={`pb-2 px-4 flex items-center ${activeTab === 'appearance' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('appearance')}
                >
                    <Palette size={18} className="mr-2" /> Appearance
                </button>
            </div>

            {activeTab === 'clusters' && (
                <div className="space-y-8">
                    <div className="bg-gray-800 p-6 rounded-lg border border-gray-700">
                        <h2 className="text-lg font-semibold text-white mb-4">Configured Clusters</h2>
                        <div className="space-y-2">
                            {clusters.map(cluster => (
                                <div key={cluster} className="flex items-center justify-between p-3 bg-gray-750 rounded border border-gray-700">
                                    <div className="flex items-center">
                                        <Server size={18} className="text-gray-400 mr-3" />
                                        <span className="text-gray-200 font-medium">{cluster}</span>
                                        {cluster === currentCluster && (
                                            <span className="ml-3 px-2 py-0.5 text-xs bg-blue-900 text-blue-200 rounded-full">Active</span>
                                        )}
                                    </div>
                                    {cluster !== currentCluster && (
                                        <button
                                            onClick={() => setCurrentCluster(cluster)}
                                            className="text-sm text-blue-400 hover:text-blue-300"
                                        >
                                            Switch
                                        </button>
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>

                    <div className="bg-gray-800 p-6 rounded-lg border border-gray-700">
                        <h2 className="text-lg font-semibold text-white mb-4">Add New Cluster</h2>
                        <form onSubmit={handleAddCluster} className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-400 mb-1">Cluster Name</label>
                                <input
                                    type="text"
                                    value={newCluster.name}
                                    onChange={e => setNewCluster({ ...newCluster, name: e.target.value })}
                                    className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:outline-none focus:border-blue-500"
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
                                    className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:outline-none focus:border-blue-500"
                                    placeholder="https://api.k8s.example.com"
                                    required
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-400 mb-1">Bearer Token</label>
                                <textarea
                                    value={newCluster.token}
                                    onChange={e => setNewCluster({ ...newCluster, token: e.target.value })}
                                    className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:outline-none focus:border-blue-500 h-24 font-mono text-xs"
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
                                    className="h-4 w-4 bg-gray-900 border-gray-700 rounded text-blue-500 focus:ring-0"
                                />
                                <label htmlFor="insecure" className="ml-2 text-sm text-gray-400">Skip TLS Verification (Insecure)</label>
                            </div>

                            {error && (
                                <div className="flex items-center text-red-400 text-sm bg-red-900/20 p-3 rounded">
                                    <AlertCircle size={16} className="mr-2" /> {error}
                                </div>
                            )}
                            {success && (
                                <div className="flex items-center text-green-400 text-sm bg-green-900/20 p-3 rounded">
                                    <Check size={16} className="mr-2" /> {success}
                                </div>
                            )}

                            <button
                                type="submit"
                                className="flex items-center justify-center w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded transition-colors"
                            >
                                <Plus size={18} className="mr-2" /> Add Cluster
                            </button>
                        </form>
                    </div>
                </div>
            )}

            {activeTab === 'appearance' && (
                <div className="space-y-8">
                    <div className="bg-gray-800 p-6 rounded-lg border border-gray-700">
                        <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                            <Palette size={20} className="mr-2" /> Color Theme
                        </h2>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {themes.map(t => (
                                <button
                                    key={t.id}
                                    onClick={() => setTheme(t.id)}
                                    className={`flex items-center justify-between p-4 rounded border transition-all ${theme === t.id ? 'border-blue-500 bg-blue-900/20' : 'border-gray-700 bg-gray-750 hover:bg-gray-700'}`}
                                >
                                    <div className="flex items-center">
                                        <div className="w-6 h-6 rounded-full mr-3 border border-gray-600" style={{ backgroundColor: t.color }}></div>
                                        <span className="text-gray-200">{t.name}</span>
                                    </div>
                                    {theme === t.id && <Check size={18} className="text-blue-400" />}
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="bg-gray-800 p-6 rounded-lg border border-gray-700">
                        <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                            <Type size={20} className="mr-2" /> Font Family
                        </h2>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {fonts.map(f => (
                                <button
                                    key={f}
                                    onClick={() => setFont(f)}
                                    className={`flex items-center justify-between p-4 rounded border transition-all ${font === f ? 'border-blue-500 bg-blue-900/20' : 'border-gray-700 bg-gray-750 hover:bg-gray-700'}`}
                                >
                                    <span className="text-gray-200" style={{ fontFamily: f }}>{f}</span>
                                    {font === f && <Check size={18} className="text-blue-400" />}
                                </button>
                            ))}
                        </div>
                        <p className="mt-4 text-sm text-gray-500">
                            Note: Fonts are loaded from Google Fonts. An internet connection is required.
                        </p>
                    </div>

                    <div className="bg-gray-800 p-6 rounded-lg border border-gray-700">
                        <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                            <Server size={20} className="mr-2" /> Custom Logo
                        </h2>
                        <div className="space-y-4">
                            <p className="text-sm text-gray-400">
                                Upload a custom logo (PNG or SVG) to replace the default application logo.
                                The image will be resized to fit the header (max height 48px).
                            </p>
                            <div className="flex items-center space-x-4">
                                <input
                                    type="file"
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
                                                    setSuccess('Logo uploaded successfully! Refresh to see changes.');
                                                    setTimeout(() => window.location.reload(), 1500);
                                                } else {
                                                    throw new Error(await res.text());
                                                }
                                            })
                                            .catch((err) => setError(err.message));
                                    }}
                                    className="block w-full text-sm text-gray-400
                                    file:mr-4 file:py-2 file:px-4
                                    file:rounded-full file:border-0
                                    file:text-sm file:font-semibold
                                    file:bg-blue-600 file:text-white
                                    hover:file:bg-blue-700
                                    cursor-pointer"
                                />
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Settings;
