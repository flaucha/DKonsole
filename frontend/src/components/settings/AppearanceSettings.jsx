import React, { useState } from 'react';
import { Palette, Type, Check, AlertCircle, Menu, Plus, Trash2, Sun, Moon } from 'lucide-react';
import { useSettings } from '../../context/SettingsContext';
import { useAuth } from '../../context/AuthContext';
import { parseErrorResponse, parseError } from '../../utils/errorParser';

const AppearanceSettings = () => {
    const {
        theme, setTheme,
        font, setFont,
        fontSize, setFontSize,
        borderRadius, setBorderRadius,
        menuAnimationSpeed, setMenuAnimationSpeed
    } = useSettings();

    const { authFetch } = useAuth();
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    // Only dark (default) and cream themes - toggle between them
    const isDarkTheme = theme === 'default' || theme !== 'cream';

    const fontGroups = {
        'Sans-Serif': ['Inter', 'Roboto', 'Open Sans', 'Lato', 'Montserrat', 'Source Sans Pro', 'Nunito', 'Quicksand', 'Raleway', 'Ubuntu', 'Oswald'],
        'Serif': ['Merriweather', 'Playfair Display'],
        'Monospace': ['Fira Code', 'Inconsolata'],
    };

    const handleThemeToggle = () => {
        setTheme(isDarkTheme ? 'cream' : 'default');
    };

    return (
        <div className="w-full max-w-4xl mx-auto space-y-6">
            {/* Theme Toggle & UI Preferences */}
            <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                <h2 className="text-lg font-semibold text-white mb-6 flex items-center">
                    <Palette size={20} className="mr-2 text-purple-400" /> Appearance
                </h2>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    {/* Theme Toggle */}
                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">Theme</label>
                        <button
                            onClick={handleThemeToggle}
                            className="w-full flex items-center justify-between p-3 rounded-lg border border-gray-700 bg-gray-900 hover:border-gray-500 transition-all"
                        >
                            <div className="flex items-center gap-3">
                                {isDarkTheme ? (
                                    <Moon size={20} className="text-blue-400" />
                                ) : (
                                    <Sun size={20} className="text-yellow-400" />
                                )}
                                <span className="text-white font-medium">
                                    {isDarkTheme ? 'Dark' : 'Cream'}
                                </span>
                            </div>
                            <div className={`w-12 h-6 rounded-full p-1 transition-colors ${isDarkTheme ? 'bg-gray-700' : 'bg-yellow-500'}`}>
                                <div className={`w-4 h-4 rounded-full bg-white shadow transition-transform ${isDarkTheme ? '' : 'translate-x-6'}`} />
                            </div>
                        </button>
                    </div>

                    {/* Font Size */}
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

                    {/* Border Radius */}
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

            {/* Font Family & Animation Speed */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Font Family Dropdown */}
                <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                    <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                        <Type size={20} className="mr-2 text-blue-400" /> Font Family
                    </h2>
                    <select
                        value={font}
                        onChange={(e) => setFont(e.target.value)}
                        className="w-full p-3 rounded-lg border border-gray-700 bg-gray-900 text-white focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all"
                        style={{ fontFamily: font }}
                    >
                        {Object.entries(fontGroups).map(([group, fonts]) => (
                            <optgroup key={group} label={group}>
                                {fonts.map(f => (
                                    <option key={f} value={f} style={{ fontFamily: f }}>
                                        {f}
                                    </option>
                                ))}
                            </optgroup>
                        ))}
                    </select>
                    <p className="mt-3 text-xs text-gray-500 flex items-center">
                        <AlertCircle size={12} className="mr-1" /> Fonts are loaded dynamically from Google Fonts.
                    </p>
                </div>

                {/* Animation Speed */}
                <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                    <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                        <Menu size={20} className="mr-2 text-blue-400" /> Animation Speed
                    </h2>
                    <p className="text-sm text-gray-400 mb-4">
                        Control the speed of menu slide animations.
                    </p>
                    <div className="flex bg-gray-900 p-1 rounded-lg border border-gray-700">
                        {[
                            { id: 'slow', name: 'Slow', desc: '500ms' },
                            { id: 'medium', name: 'Medium', desc: '300ms' },
                            { id: 'fast', name: 'Fast', desc: '150ms' }
                        ].map(speed => (
                            <button
                                key={speed.id}
                                onClick={() => setMenuAnimationSpeed(speed.id)}
                                className={`flex-1 py-2 text-sm rounded-md transition-all ${menuAnimationSpeed === speed.id
                                    ? 'bg-gray-700 text-white shadow'
                                    : 'text-gray-400 hover:text-gray-200'
                                    }`}
                            >
                                <div className="font-medium">{speed.name}</div>
                                <div className="text-xs opacity-75">{speed.desc}</div>
                            </button>
                        ))}
                    </div>
                </div>
            </div>

            {/* Custom Logo */}
            <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                    <Menu size={20} className="mr-2 text-pink-400" /> Custom Logo
                </h2>
                <div className="space-y-6">
                    <div className="flex-1">
                        <p className="text-sm text-gray-400 mb-4">
                            Upload custom logos (PNG or SVG) to replace the default application logos. You can upload separate logos for dark and light themes.
                        </p>
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm text-gray-300 mb-2">Logo for Dark Themes</label>
                                <div className="flex items-center space-x-3">
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
                                    <button
                                        onClick={() => {
                                            if (!window.confirm('Are you sure you want to reset the dark theme logo to default?')) return;

                                            authFetch('/api/logo?type=normal', {
                                                method: 'DELETE',
                                            })
                                                .then(async (res) => {
                                                    if (res.ok) {
                                                        setSuccess('Logo reset to default successfully! Refreshing...');
                                                        setTimeout(() => window.location.reload(), 1500);
                                                    } else {
                                                        const errorText = await parseErrorResponse(res);
                                                        throw new Error(errorText);
                                                    }
                                                })
                                                .catch((err) => setError(parseError(err)));
                                        }}
                                        className="inline-flex items-center px-4 py-2 bg-red-900/30 hover:bg-red-900/50 text-red-400 hover:text-red-300 text-sm font-medium rounded-lg cursor-pointer transition-colors border border-red-900/50 hover:border-red-800"
                                    >
                                        <Trash2 size={16} className="mr-2" /> Reset Default
                                    </button>
                                </div>
                            </div>
                            <div>
                                <label className="block text-sm text-gray-300 mb-2">Logo for Light Themes</label>
                                <div className="flex items-center space-x-3">
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
                                    <button
                                        onClick={() => {
                                            if (!window.confirm('Are you sure you want to reset the light theme logo to default?')) return;

                                            authFetch('/api/logo?type=light', {
                                                method: 'DELETE',
                                            })
                                                .then(async (res) => {
                                                    if (res.ok) {
                                                        setSuccess('Logo reset to default successfully! Refreshing...');
                                                        setTimeout(() => window.location.reload(), 1500);
                                                    } else {
                                                        const errorText = await parseErrorResponse(res);
                                                        throw new Error(errorText);
                                                    }
                                                })
                                                .catch((err) => setError(parseError(err)));
                                        }}
                                        className="inline-flex items-center px-4 py-2 bg-red-900/30 hover:bg-red-900/50 text-red-400 hover:text-red-300 text-sm font-medium rounded-lg cursor-pointer transition-colors border border-red-900/50 hover:border-red-800"
                                    >
                                        <Trash2 size={16} className="mr-2" /> Reset Default
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {error && (
                <div className="fixed bottom-6 right-6 flex items-center text-red-400 text-sm bg-red-900/90 border border-red-500/50 rounded-lg p-3 shadow-xl z-50">
                    <AlertCircle size={16} className="mr-2" />
                    {error}
                </div>
            )}
            {success && (
                <div className="fixed bottom-6 right-6 flex items-center text-green-400 text-sm bg-green-900/90 border border-green-500/50 rounded-lg p-3 shadow-xl z-50">
                    <Check size={16} className="mr-2" />
                    {success}
                </div>
            )}
        </div>
    );
};

export default AppearanceSettings;
