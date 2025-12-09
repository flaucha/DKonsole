import React from 'react';
import { NavLink } from 'react-router-dom';
import { LayoutDashboard, Box, ChevronDown, Network, HardDrive, Shield, Package, X, Siren } from 'lucide-react';

const SidebarItem = ({ icon: Icon, label, to, onClick, hasChildren, expanded }) => {
    if (hasChildren) {
        return (
            <div
                onClick={onClick}
                className={`flex items-center justify-between px-4 py-2 cursor-pointer rounded-md transition-all duration-200 text-gray-100 hover:bg-gray-800 hover:text-gray-100 border border-transparent hover:border-gray-700 ${expanded ? 'bg-gray-800/50 border-gray-700' : ''}`}
            >
                <div className="flex items-center space-x-3">
                    <Icon size={20} className="text-gray-300 group-hover:text-blue-400" />
                    <span className="font-medium whitespace-nowrap">{label}</span>
                </div>
                <div className={`transition-transform duration-200 text-gray-400 ${expanded ? 'rotate-0 text-blue-400' : '-rotate-90'}`}>
                    <ChevronDown size={16} />
                </div>
            </div>
        );
    }

    return (
        <NavLink
            to={to}
            className={({ isActive }) =>
                `flex items-center justify-between px-4 py-2 cursor-pointer rounded-md transition-all duration-200 border border-transparent ${isActive
                    ? 'bg-gray-800 text-gray-100 border-l-4 border-l-blue-500 shadow-md'
                    : 'text-gray-300 hover:bg-gray-800 hover:text-gray-100 hover:border-gray-700'
                }`
            }
        >
            {({ isActive }) => (
                <div className="flex items-center space-x-3">
                    <Icon size={20} className={isActive ? 'text-blue-400' : 'text-gray-400'} />
                    <span className="font-medium whitespace-nowrap">{label}</span>
                </div>
            )}
        </NavLink>
    );
};

const SubItem = ({ label, to }) => (
    <NavLink
        to={to}
        className={({ isActive }) =>
            `block pl-12 pr-4 py-1.5 cursor-pointer text-xs transition-all duration-200 whitespace-nowrap rounded-md border border-transparent ${isActive
                ? 'text-gray-100 font-semibold bg-gray-800/60 border-l-4 border-l-blue-500 shadow-sm'
                : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800/40 hover:border-gray-700'
            }`
        }
        style={{ fontSize: '0.75rem' }}
    >
        {({ isActive }) => (
            <>
                <span className={`mr-2 ${isActive ? 'text-blue-400' : 'text-gray-500'}`}>•</span>
                {label}
            </>
        )}
    </NavLink>
);

const SubMenu = ({ isOpen, children, animationStyle = 'slide', animationSpeed = 'medium' }) => {
    const getAnimationClasses = () => {
        switch (animationStyle) {
            case 'slide':
                return isOpen
                    ? 'max-h-[500px] opacity-100 translate-y-0'
                    : 'max-h-0 opacity-0 -translate-y-2';
            case 'fade':
                return isOpen
                    ? 'max-h-[500px] opacity-100'
                    : 'max-h-0 opacity-0';
            case 'scale':
                return isOpen
                    ? 'max-h-[500px] opacity-100 scale-100'
                    : 'max-h-0 opacity-0 scale-95';
            case 'rotate':
                return isOpen
                    ? 'max-h-[500px] opacity-100 rotate-0'
                    : 'max-h-0 opacity-0 rotate-[-2deg]';
            default:
                return isOpen
                    ? 'max-h-[500px] opacity-100 translate-y-0'
                    : 'max-h-0 opacity-0 -translate-y-2';
        }
    };

    const getSpeedDuration = () => {
        switch (animationSpeed) {
            case 'slow':
                return 'duration-500';
            case 'fast':
                return 'duration-150';
            case 'medium':
            default:
                return 'duration-300';
        }
    };

    const getTransitionClasses = () => {
        const speedClass = getSpeedDuration();
        switch (animationStyle) {
            case 'slide':
                return `transition-all ${speedClass} ease-out`;
            case 'fade':
                return `transition-all ${speedClass === 'duration-500' ? 'duration-400' : speedClass === 'duration-150' ? 'duration-200' : 'duration-250'} ease-in-out`;
            case 'scale':
                return `transition-all ${speedClass} ease-out transform-gpu`;
            case 'rotate':
                return `transition-all ${speedClass} ease-out transform-gpu origin-top-left`;
            default:
                return `transition-all ${speedClass} ease-out`;
        }
    };

    return (
        <div
            className={`overflow-hidden ${getTransitionClasses()} ${getAnimationClasses()}`}
            style={{
                transformOrigin: animationStyle === 'scale' ? 'top' : animationStyle === 'rotate' ? 'top left' : 'top'
            }}
        >
            <div className="space-y-1 mb-2">
                {children}
            </div>
        </div>
    );
};

const Sidebar = ({
    sidebarOpen,
    setSidebarOpen,
    expandedMenus,
    toggleMenu,
    isAdmin,
    hasPermissions,
    menuAnimation,
    menuAnimationSpeed,
    checkingAdmin,
    user
}) => {
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
        <div
            className={`bg-gray-900 border-r border-gray-700 flex flex-col transition-all duration-500 ease-[cubic-bezier(0.4,0,0.2,1)] shadow-xl ${sidebarOpen
                ? 'min-w-[200px] w-auto translate-x-0 opacity-100'
                : 'w-0 -translate-x-full opacity-0 pointer-events-none'
                }`}
            style={{
                transform: sidebarOpen ? 'translateX(0)' : 'translateX(-100%)',
                transition: 'all 0.5s cubic-bezier(0.4, 0, 0.2, 1)',
            }}
        >
            <nav className={`flex-1 px-2 pt-4 space-y-1 overflow-y-auto transition-opacity duration-300 ${sidebarOpen ? 'opacity-100' : 'opacity-0'}`}>
                {/* Overview con botón de cerrar */}
                <div className="flex items-center justify-between px-4 py-2 mb-1">
                    <NavLink
                        to="/dashboard/overview"
                        className={({ isActive }) =>
                            `flex items-center space-x-3 flex-1 px-4 py-2 cursor-pointer rounded-md transition-all duration-200 border border-transparent ${isActive
                                ? 'bg-gray-800 text-gray-100 border-l-4 border-l-blue-500 shadow-md'
                                : 'text-gray-300 hover:bg-gray-800 hover:text-gray-100 hover:border-gray-700'
                            }`
                        }
                    >
                        {({ isActive }) => (
                            <>
                                <LayoutDashboard size={20} className={isActive ? 'text-blue-400' : 'text-gray-400'} />
                                <span className="font-medium whitespace-nowrap">Overview</span>
                            </>
                        )}
                    </NavLink>
                    {sidebarOpen && (
                        <button
                            onClick={() => setSidebarOpen(false)}
                            className="p-1.5 text-gray-400 hover:text-gray-100 hover:bg-gray-800 rounded-md transition-all duration-200 hover:scale-110 hover:rotate-90 ml-2 border border-transparent hover:border-gray-600"
                            title="Ocultar menú"
                        >
                            <X size={16} />
                        </button>
                    )}
                </div>

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
                        <SubMenu isOpen={expandedMenus.workloads} animationStyle={menuAnimation} animationSpeed={menuAnimationSpeed}>
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
                        <SubMenu isOpen={expandedMenus.networking} animationStyle={menuAnimation} animationSpeed={menuAnimationSpeed}>
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
                        <SubMenu isOpen={expandedMenus.storage} animationStyle={menuAnimation} animationSpeed={menuAnimationSpeed}>
                            {['PVCs', ...(isAdmin ? ['PVs', 'Storage Classes'] : [])].map(item => (
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
                                <SubMenu isOpen={expandedMenus.accessControl} animationStyle={menuAnimation} animationSpeed={menuAnimationSpeed}>
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

                        {/* Admin Area - Only for admins */}
                        {isAdmin && (
                            <>
                                <SidebarItem
                                    icon={Siren}
                                    label="Admin Area"
                                    hasChildren
                                    expanded={expandedMenus.adminArea}
                                    onClick={() => toggleMenu('adminArea')}
                                />
                                <SubMenu isOpen={expandedMenus.adminArea} animationStyle={menuAnimation} animationSpeed={menuAnimationSpeed}>
                                    <SubItem
                                        label="Nodes"
                                        to="/dashboard/workloads/Node"
                                    />
                                    <SubItem
                                        label="Namespaces"
                                        to="/dashboard/namespaces"
                                    />
                                    {/* Consolidated Quotas and Limits menu item */}
                                    <SubItem
                                        label="Quotas / Limits"
                                        to="/dashboard/resource-quotas"
                                    />
                                    <SubItem
                                        label="API Explorer"
                                        to="/dashboard/api-explorer"
                                    />
                                    <SubItem
                                        label="Helm Charts"
                                        to="/dashboard/helm-charts"
                                    />
                                    <SubItem
                                        label="Settings"
                                        to="/dashboard/settings"
                                    />
                                </SubMenu>
                            </>
                        )}

                        {/* Helm Charts - Only for users with namespace permissions (non-admin) */}
                        {!isAdmin && hasPermissions && (
                            <SidebarItem
                                icon={Package}
                                label="Helm Charts"
                                to="/dashboard/helm-charts"
                            />
                        )}
                    </>
                )}
            </nav>

            <div className="mt-auto border-t border-gray-700 bg-gray-800/30">
                <div className="px-4 py-2 whitespace-nowrap overflow-hidden">
                    <div className="text-xs text-gray-500">User: <span className="text-gray-300 font-medium">{user?.username || 'Unknown'}</span></div>
                </div>
            </div>
        </div>
    );
};

export default Sidebar;
