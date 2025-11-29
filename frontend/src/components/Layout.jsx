import React, { useState, useEffect } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { LayoutDashboard, Box, Settings, Activity, ChevronDown, ChevronRight, Network, HardDrive, Menu, Server, ListTree, Shield, Database, Gauge, Package, LogOut } from 'lucide-react';
import defaultLogo from '../assets/logo-full.svg';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';

const SidebarItem = ({ icon: Icon, label, to, onClick, hasChildren, expanded }) => {
    if (hasChildren) {
        return (
            <div
                onClick={onClick}
                className={`flex items-center justify-between px-4 py-2 cursor-pointer rounded-md transition-colors text-white hover:bg-gray-800`}
            >
                <div className="flex items-center space-x-3">
                    <Icon size={20} />
                    <span className="font-medium whitespace-nowrap">{label}</span>
                </div>
                <div className={`transition-transform duration-200 ${expanded ? 'rotate-0' : '-rotate-90'}`}>
                    <ChevronDown size={16} />
                </div>
            </div>
        );
    }

    return (
        <NavLink
            to={to}
            className={({ isActive }) =>
                `flex items-center justify-between px-4 py-2 cursor-pointer rounded-md transition-colors ${isActive ? 'bg-gray-800 text-white' : 'text-white hover:bg-gray-800'}`
            }
        >
            <div className="flex items-center space-x-3">
                <Icon size={20} />
                <span className="font-medium whitespace-nowrap">{label}</span>
            </div>
        </NavLink>
    );
};

const SubItem = ({ label, to }) => (
    <NavLink
        to={to}
        className={({ isActive }) =>
            `block pl-12 pr-4 py-1.5 cursor-pointer text-sm transition-colors whitespace-nowrap ${isActive ? 'text-white font-medium' : 'text-white hover:text-gray-300'}`
        }
    >
        {label}
    </NavLink>
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

const Layout = ({ children, headerContent }) => {
    const { currentCluster } = useSettings();
    const { logout, authFetch, user } = useAuth();
    const [isAdmin, setIsAdmin] = useState(false);
    const [checkingAdmin, setCheckingAdmin] = useState(true);
    const [hasPermissions, setHasPermissions] = useState(false);

    useEffect(() => {
        // Check if user is admin (core admin or LDAP admin group member)
        const checkAdmin = async () => {
            try {
                const res = await authFetch('/api/settings/prometheus/url');
                if (res.ok || res.status === 404) {
                    setIsAdmin(true);
                    setHasPermissions(true);
                } else if (res.status === 403) {
                    setIsAdmin(false);
                    // Check if user has LDAP permissions
                    if (user && user.permissions && Object.keys(user.permissions).length > 0) {
                        setHasPermissions(true);
                    } else {
                        setHasPermissions(false);
                    }
                } else {
                    setIsAdmin(false);
                    setHasPermissions(false);
                }
            } catch (err) {
                setIsAdmin(false);
                // Check if user has LDAP permissions
                if (user && user.permissions && Object.keys(user.permissions).length > 0) {
                    setHasPermissions(true);
                } else {
                    setHasPermissions(false);
                }
            } finally {
                setCheckingAdmin(false);
            }
        };
        if (user) {
            checkAdmin();
        } else {
            setCheckingAdmin(false);
            setHasPermissions(false);
        }
    }, [authFetch, user]);
    const [sidebarOpen, setSidebarOpen] = useState(true);
    const [expandedMenus, setExpandedMenus] = useState({
        workloads: true,
        networking: false,
        storage: false,
        accessControl: false,
    });
    const [logoSrc, setLogoSrc] = useState(defaultLogo);
    const location = useLocation();

    useEffect(() => {
        // Add timestamp to prevent browser caching
        authFetch(`/api/logo?t=${Date.now()}`)
            .then(res => {
                if (res.ok && res.status === 200) {
                    // Add timestamp to prevent caching
                    setLogoSrc(`/api/logo?t=${Date.now()}`);
                }
            })
            .catch(() => { });
    }, [authFetch]);

    const handleLogoError = () => {
        // If logo fails to load, fallback to default logo
        if (logoSrc !== defaultLogo) {
            setLogoSrc(defaultLogo);
        }
    };

    const toggleMenu = (menu) => {
        setExpandedMenus(prev => ({ ...prev, [menu]: !prev[menu] }));
    };

    // Helper to map view names to paths
    const getPath = (view) => {
        const kindMap = {
            'Deployments': 'Deployment',
            'Pods': 'Pod',
            'ConfigMaps': 'ConfigMap',
            'Secrets': 'Secret',
            'Jobs': 'Job',
            'CronJobs': 'CronJob',
            'StatefulSets': 'StatefulSet',
            'DaemonSets': 'DaemonSet',
            'HPA': 'HorizontalPodAutoscaler',
            'Services': 'Service',
            'Ingresses': 'Ingress',
            'Network Policies': 'NetworkPolicy',
            'PVCs': 'PersistentVolumeClaim',
            'PVs': 'PersistentVolume',
            'Storage Classes': 'StorageClass',
            'Nodes': 'Node',
            'Service Accounts': 'ServiceAccount',
            'Roles': 'Role',
            'Cluster Roles': 'ClusterRole',
            'Role Bindings': 'RoleBinding',
            'Cluster Role Bindings': 'ClusterRoleBinding'
        };

        if (kindMap[view]) {
            return `/dashboard/workloads/${kindMap[view]}`;
        }

        // Fallback for direct mapping
        return `/dashboard/${view.toLowerCase().replace(' ', '-')}`;
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
                        <img
                            src={logoSrc}
                            alt="Logo"
                            className="h-12 max-h-12 object-contain"
                            onError={handleLogoError}
                        />
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
                            to="/dashboard/overview"
                        />

                        {/* Only show other menu items if user has permissions */}
                        {!checkingAdmin && hasPermissions && (
                            <>
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
                                            to={getPath(item)}
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
                                            to={getPath(item)}
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
                                            to={getPath(item)}
                                        />
                                    ))}
                                </SubMenu>

                                {/* Access Control - Only for admins */}
                                {isAdmin && (
                                    <>
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
                                                    to={getPath(item)}
                                                />
                                            ))}
                                        </SubMenu>
                                    </>
                                )}

                                <SidebarItem
                                    icon={Database}
                                    label="Namespaces"
                                    to="/dashboard/namespaces"
                                />

                                <SidebarItem
                                    icon={ListTree}
                                    label="API Explorer"
                                    to="/dashboard/api-explorer"
                                />

                                <SidebarItem
                                    icon={Package}
                                    label="Helm Charts"
                                    to="/dashboard/helm-charts"
                                />

                                <SidebarItem
                                    icon={Settings}
                                    label="Settings"
                                    to="/dashboard/settings"
                                />
                            </>
                        )}
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
                            <div className="text-xs text-gray-500">User: <span className="text-gray-300 font-medium">{user?.username || 'Unknown'}</span></div>
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
