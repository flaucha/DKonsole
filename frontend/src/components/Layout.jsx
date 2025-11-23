import React, { useState, useEffect } from 'react';
import { LayoutDashboard, Box, Settings, Activity, ChevronDown, ChevronRight, Network, HardDrive, Menu, Server, ListTree, Shield, Database, Gauge } from 'lucide-react';
import defaultLogo from '../assets/logo-full.svg';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { LogOut } from 'lucide-react';

const SidebarItem = ({ icon: Icon, label, active, onClick, hasChildren, expanded }) => (
    <div
        onClick={onClick}
        className={`flex items-center justify-between px-4 py-2 cursor-pointer rounded-md transition-colors ${active ? 'bg-gray-800 text-white' : 'text-white hover:bg-gray-800'}`}
    >
        <div className="flex items-center space-x-3">
            <Icon size={20} />
            <span className="font-medium whitespace-nowrap">{label}</span>
        </div>
        {hasChildren && (
            <div className={`transition-transform duration-200 ${expanded ? 'rotate-0' : '-rotate-90'}`}>
                <ChevronDown size={16} />
            </div>
        )}
    </div>
);

const SubItem = ({ label, active, onClick }) => (
    <div
        onClick={onClick}
        className={`pl-12 pr-4 py-1.5 cursor-pointer text-sm transition-colors whitespace-nowrap ${active ? 'text-white font-medium' : 'text-white hover:text-gray-300'}`}
    >
        {label}
    </div>
);

const SubMenu = ({ isOpen, children }) => (
    <div
        className={`overflow-hidden transition-all duration-300 ease-in-out ${isOpen ? 'max-h-[500px] opacity-100' : 'max-h-0 opacity-0'}`}
    >
        <div className="space-y-1 mb-2">
            {children}
        </div>
    </div>
);

const Layout = ({ children, currentView, onViewChange, headerContent }) => {
    const { currentCluster } = useSettings();
    const { logout, authFetch } = useAuth();
    const [sidebarOpen, setSidebarOpen] = useState(true);
    const [expandedMenus, setExpandedMenus] = useState({
        workloads: true,
        networking: false,
        storage: false,
        accessControl: false,
    });
    const [logoSrc, setLogoSrc] = useState(defaultLogo);

    useEffect(() => {
        authFetch('/api/logo')
            .then(res => {
                // Only set custom logo if response is OK (200) and has content
                if (res.ok && res.status === 200) {
                    setLogoSrc('/api/logo');
                }
                // If 404 or other error, keep default logo (no action needed)
            })
            .catch(() => {
                // Silently handle errors (network, etc.) - keep default logo
            });
    }, []);

    const toggleMenu = (menu) => {
        setExpandedMenus(prev => ({ ...prev, [menu]: !prev[menu] }));
    };

    return (
        <div className="flex flex-col h-screen bg-gray-900">
            {/* Header */}
            <header className="h-16 bg-black border-b border-gray-800 flex items-center justify-between px-4 shrink-0 z-20">
                <div className="flex items-center space-x-2">
                    <button
                        onClick={() => setSidebarOpen(!sidebarOpen)}
                        className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-md transition-colors"
                    >
                        <Menu size={24} />
                    </button>
                    <div className="flex items-center justify-center">
                        <img src={logoSrc} alt="Logo" className="h-12 max-h-12 object-contain" />
                    </div>
                </div>
                <div className="flex items-center">
                    {headerContent}
                </div>
            </header>

            <div className="flex flex-1 overflow-hidden">
                {/* Sidebar */}
                <div
                    className={`bg-black border-r border-gray-800 flex flex-col transition-all duration-300 ease-in-out ${sidebarOpen ? 'w-64 translate-x-0' : 'w-0 -translate-x-full opacity-0 overflow-hidden border-none'}`}
                >
                    <nav className="flex-1 px-2 space-y-1 mt-4 overflow-y-auto">
                        <SidebarItem
                            icon={LayoutDashboard}
                            label="Overview"
                            active={currentView === 'Overview'}
                            onClick={() => onViewChange('Overview')}
                        />

                        {/* Workloads */}
                        <SidebarItem
                            icon={Box}
                            label="Workloads"
                            hasChildren
                            expanded={expandedMenus.workloads}
                            onClick={() => toggleMenu('workloads')}
                        />
                        <SubMenu isOpen={expandedMenus.workloads}>
                            {['Deployments', 'Pods', 'ConfigMaps', 'Secrets', 'Jobs', 'CronJobs', 'StatefulSets', 'DaemonSets', 'HPA'].map(item => (
                                <SubItem
                                    key={item}
                                    label={item}
                                    active={currentView === item}
                                    onClick={() => onViewChange(item)}
                                />
                            ))}
                        </SubMenu>

                        {/* Networking */}
                        <SidebarItem
                            icon={Network}
                            label="Networking"
                            hasChildren
                            expanded={expandedMenus.networking}
                            onClick={() => toggleMenu('networking')}
                        />
                        <SubMenu isOpen={expandedMenus.networking}>
                            {['Services', 'Ingresses', 'Network Policies'].map(item => (
                                <SubItem
                                    key={item}
                                    label={item}
                                    active={currentView === item}
                                    onClick={() => onViewChange(item)}
                                />
                            ))}
                        </SubMenu>

                        {/* Storage */}
                        <SidebarItem
                            icon={HardDrive}
                            label="Storage"
                            hasChildren
                            expanded={expandedMenus.storage}
                            onClick={() => toggleMenu('storage')}
                        />
                        <SubMenu isOpen={expandedMenus.storage}>
                            {['PVCs', 'PVs', 'Storage Classes'].map(item => (
                                <SubItem
                                    key={item}
                                    label={item}
                                    active={currentView === item}
                                    onClick={() => onViewChange(item)}
                                />
                            ))}
                        </SubMenu>

                        {/* Access Control */}
                        <SidebarItem
                            icon={Shield}
                            label="Access Control"
                            hasChildren
                            expanded={expandedMenus.accessControl}
                            onClick={() => toggleMenu('accessControl')}
                        />
                        <SubMenu isOpen={expandedMenus.accessControl}>
                            {['Service Accounts', 'Roles', 'Role Bindings', 'Cluster Roles', 'Cluster Role Bindings'].map(item => (
                                <SubItem
                                    key={item}
                                    label={item}
                                    active={currentView === item}
                                    onClick={() => onViewChange(item)}
                                />
                            ))}
                        </SubMenu>

                        <SidebarItem
                            icon={Server}
                            label="Nodes"
                            active={currentView === 'Nodes'}
                            onClick={() => onViewChange('Nodes')}
                        />

                        <SidebarItem
                            icon={Database}
                            label="Namespaces"
                            active={currentView === 'Namespaces'}
                            onClick={() => onViewChange('Namespaces')}
                        />

                        <SidebarItem
                            icon={Gauge}
                            label="Resource Quotas"
                            active={currentView === 'Resource Quotas'}
                            onClick={() => onViewChange('Resource Quotas')}
                        />

                        <SidebarItem
                            icon={ListTree}
                            label="API Explorer"
                            active={currentView === 'API Explorer'}
                            onClick={() => onViewChange('API Explorer')}
                        />

                        <SidebarItem
                            icon={Settings}
                            label="Settings"
                            active={currentView === 'Settings'}
                            onClick={() => onViewChange('Settings')}
                        />
                    </nav>

                    <div className="mt-auto border-t border-gray-800">
                        <button
                            onClick={logout}
                            className="w-full flex items-center px-4 py-2 text-gray-400 hover:text-white hover:bg-gray-800 transition-colors"
                        >
                            <LogOut size={20} className="mr-3" />
                            <span className="font-medium">Logout</span>
                        </button>
                        <div className="px-4 py-2 border-t border-gray-800 whitespace-nowrap overflow-hidden">
                            <div className="text-xs text-gray-500">Cluster: <span className="text-gray-300 font-medium">{currentCluster}</span></div>
                        </div>
                    </div>
                </div>

                {/* Main Content */}
                <div className="flex-1 overflow-auto bg-gray-900 relative">
                    {children}
                </div>
            </div>
        </div>
    );
};

export default Layout;
