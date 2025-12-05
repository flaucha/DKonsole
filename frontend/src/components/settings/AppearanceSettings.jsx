import React, { useState } from 'react';
import { Palette, Type, Check, AlertCircle, Menu, Plus } from 'lucide-react';
import { useSettings } from '../../context/SettingsContext';
import { useAuth } from '../../context/AuthContext';
import { parseErrorResponse, parseError } from '../../utils/errorParser';

const AppearanceSettings = () => {
    const {
        theme, setTheme,
        font, setFont,
        fontSize, setFontSize,
        borderRadius, setBorderRadius,
        menuAnimation, setMenuAnimation,
        menuAnimationSpeed, setMenuAnimationSpeed
    } = useSettings();

    const { authFetch } = useAuth();
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

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

                {/* Menu Animation Style */}
                <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                    <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                        <Menu size={20} className="mr-2 text-blue-400" /> Menu Animation Style
                    </h2>
                    <p className="text-sm text-gray-400 mb-4">
                        Choose the animation style for sidebar submenus when they open and close.
                    </p>
                    <div className="grid grid-cols-2 gap-3 mb-6">
                        {[
                            { id: 'slide', name: 'Slide', desc: 'Slides down smoothly' },
                            { id: 'fade', name: 'Fade', desc: 'Fades in/out' },
                            { id: 'scale', name: 'Scale', desc: 'Scales from top' },
                            { id: 'rotate', name: 'Rotate', desc: 'Rotates slightly' }
                        ].map(anim => (
                            <button
                                key={anim.id}
                                onClick={() => setMenuAnimation(anim.id)}
                                className={`p-4 rounded-lg border transition-all ${menuAnimation === anim.id
                                        ? 'border-blue-500 bg-blue-900/20 shadow-md'
                                        : 'border-gray-700 bg-gray-750 hover:bg-gray-700 hover:border-gray-600'
                                    }`}
                            >
                                <div className={`font-medium text-left ${menuAnimation === anim.id ? 'text-white' : 'text-gray-300'}`}>
                                    {anim.name}
                                </div>
                                <div className="text-xs text-gray-500 mt-1 text-left">
                                    {anim.desc}
                                </div>
                            </button>
                        ))}
                    </div>

                    {/* Menu Animation Speed */}
                    <div className="border-t border-gray-700 pt-4">
                        <label className="block text-sm font-medium text-gray-400 mb-3">Animation Speed</label>
                        <div className="flex bg-gray-900 p-1 rounded-lg border border-gray-700">
                            {[
                                { id: 'slow', name: 'Lento', desc: '500ms' },
                                { id: 'medium', name: 'Medio', desc: '300ms' },
                                { id: 'fast', name: 'Rápido', desc: '150ms' }
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
