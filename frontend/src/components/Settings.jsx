import React, { useState } from 'react';
import { useSettings } from '../context/SettingsContext';
import { Server, Palette, Settings as SettingsIcon, Users, Info, Trash2 } from 'lucide-react';
import ClustersSettings from './settings/ClustersSettings';
import AppearanceSettings from './settings/AppearanceSettings';
import GeneralSettings from './settings/GeneralSettings';
import LDAPSettings from './settings/LDAPSettings';
import AboutSettings from './settings/AboutSettings';

const Settings = () => {
    const {
        setTheme,
        setFont,
        setFontSize,
        setBorderRadius,
        setMenuAnimationSpeed
    } = useSettings();

    const [activeTab, setActiveTab] = useState('clusters');

    const handleResetDefaults = () => {
        setTheme('default');
        setFont('Inter');
        setFontSize('normal');
        setBorderRadius('md');
        setMenuAnimationSpeed('medium');
    };

    return (
        <div className="flex flex-col h-full max-w-5xl mx-auto w-full">
            <div className="flex-shrink-0 p-6 pb-0 overflow-hidden">
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
            </div>

            <div className="flex-1 overflow-y-auto overflow-x-hidden px-6 pb-6">
                {activeTab === 'clusters' && <ClustersSettings />}
                {activeTab === 'appearance' && <AppearanceSettings />}
                {activeTab === 'general' && <GeneralSettings />}
                {activeTab === 'ldap' && <LDAPSettings />}
                {activeTab === 'about' && <AboutSettings />}
            </div>
        </div>
    );
};

export default Settings;
