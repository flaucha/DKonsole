import React, { useState } from 'react';
import { Package, Clock, Tag, MoreVertical, Trash2, ArrowUp, ChevronDown } from 'lucide-react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import { formatDateTime } from '../utils/dateUtils';
import { getExpandableRowClasses, getExpandableRowRowClasses } from '../utils/expandableRow';
import { useHelmReleases } from '../hooks/useHelmReleases';
import HelmToolbar from './helm/HelmToolbar';
import HelmUpgradeModal from './helm/HelmUpgradeModal';
import HelmInstallModal from './helm/HelmInstallModal';

const HelmChartManager = ({ namespace }) => {
    const { currentCluster } = useSettings();
    const { authFetch, user } = useAuth();
    const toast = useToast();

    // Check if user is admin (core admin has role='admin', LDAP admin groups have no permissions but are admins)
    const isAdmin = user && user.role === 'admin';
    const [expandedId, setExpandedId] = useState(null);
    const [sortField, setSortField] = useState('name');
    const [sortDirection, setSortDirection] = useState('asc');
    const [filter, setFilter] = useState('');
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);

    // Modal states
    const [upgradeRelease, setUpgradeRelease] = useState(null);
    const [installModalOpen, setInstallModalOpen] = useState(false);
    const [installNamespace, setInstallNamespace] = useState('');

    const { data: releases = [], isLoading: loading, refetch } = useHelmReleases(authFetch, currentCluster);

    const toggleExpand = (releaseKey) => {
        setExpandedId(current => current === releaseKey ? null : releaseKey);
    };

    const handleSort = (field) => {
        setSortField((prevField) => {
            if (prevField === field) {
                setSortDirection((prevDir) => (prevDir === 'asc' ? 'desc' : 'asc'));
                return prevField;
            }
            setSortDirection('asc');
            return field;
        });
    };

    const renderSortIndicator = (field) => {
        if (sortField !== field) return null;
        return sortDirection === 'asc' ? '↑' : '↓';
    };

    // Helper to check if user has edit or admin permission for a namespace
    const hasEditPermission = (ns) => {
        // Admins have full access
        if (isAdmin) return true;
        if (!user || !user.permissions) return false;
        const permission = user.permissions[ns];
        return permission === 'edit' || permission === 'admin';
    };

    // Helper to check if user has view permission for a namespace
    const hasViewPermission = (ns) => {
        // Admins have full access
        if (isAdmin) return true;
        if (!user || !user.permissions) return false;
        const permission = user.permissions[ns];
        return permission === 'view' || permission === 'edit' || permission === 'admin';
    };

    // Filter releases by namespace and search text
    const filteredReleases = releases.filter(release => {
        // If user is admin, show all releases (no namespace filter)
        // If user is not admin, filter by namespace and check permissions
        if (!isAdmin) {
            // If namespace is "all", filter by user's allowed namespaces
            if (namespace === 'all') {
                // Only show releases in namespaces the user has access to
                if (!hasViewPermission(release.namespace)) {
                    return false;
                }
            } else {
                // Show only releases in the selected namespace (if user has access)
                if (release.namespace !== namespace || !hasViewPermission(release.namespace)) {
                    return false;
                }
            }
        }
        // Apply search filter
        if (!filter) return true;
        const searchText = filter.toLowerCase();
        return (
            release.name.toLowerCase().includes(searchText) ||
            release.chart?.toLowerCase().includes(searchText) ||
            release.namespace.toLowerCase().includes(searchText) ||
            release.status?.toLowerCase().includes(searchText)
        );
    });

    const sortedReleases = [...filteredReleases].sort((a, b) => {
        const dir = sortDirection === 'asc' ? 1 : -1;
        const getVal = (item) => {
            switch (sortField) {
                case 'name':
                    return item.name || '';
                case 'chart':
                    return item.chart || '';
                case 'namespace':
                    return item.namespace || '';
                case 'status':
                    return item.status || '';
                case 'version':
                    return item.version || '';
                case 'updated':
                    return new Date(item.updated).getTime() || 0;
                default:
                    return '';
            }
        };
        const va = getVal(a);
        const vb = getVal(b);
        if (typeof va === 'number' && typeof vb === 'number') {
            return (va - vb) * dir;
        }
        return String(va).localeCompare(String(vb)) * dir;
    });

    const getStatusBadge = (status) => {
        const statusLower = (status || 'unknown').toLowerCase();
        if (statusLower === 'deployed') {
            return <span className="px-2 py-1 text-xs rounded-full bg-green-900/50 text-green-300 border border-green-700">Deployed</span>;
        } else if (statusLower === 'failed' || statusLower === 'error') {
            return <span className="px-2 py-1 text-xs rounded-full bg-red-900/50 text-red-300 border border-red-700">Failed</span>;
        } else if (statusLower === 'pending-install' || statusLower === 'pending-upgrade' || statusLower === 'pending-rollback') {
            return <span className="px-2 py-1 text-xs rounded-full bg-yellow-900/50 text-yellow-300 border border-yellow-700">Pending</span>;
        } else {
            return <span className="px-2 py-1 text-xs rounded-full bg-gray-700 text-gray-300 border border-gray-600">{status || 'Unknown'}</span>;
        }
    };

    const handleDelete = async (release) => {
        const params = new URLSearchParams({
            name: release.name,
            namespace: release.namespace
        });
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const res = await authFetch(`/api/helm/releases?${params.toString()}`, {
                method: 'DELETE'
            });

            if (!res.ok) {
                throw new Error('Failed to uninstall Helm release');
            }

            // Refresh the list after deletion
            setTimeout(() => {
                refetch();
            }, 500);
        } catch (err) {
            toast.error(`Error uninstalling Helm release: ${err.message}`);
        }
    };

    const handleSuccess = () => {
        // Refresh the list after a delay
        setTimeout(() => {
            refetch();
        }, 2000);
    };

    if (loading && releases.length === 0) {
        return <div className="text-gray-400 animate-pulse p-6">Loading Helm releases...</div>;
    }

    const showInstallButton = isAdmin || (namespace && namespace !== 'all' && hasEditPermission(namespace));

    return (
        <div className="flex flex-col h-full">
            <HelmToolbar
                filter={filter}
                setFilter={setFilter}
                count={filteredReleases.length}
                onRefresh={() => refetch()}
                onInstallClick={() => {
                    setInstallNamespace(namespace === 'all' ? '' : namespace);
                    setInstallModalOpen(true);
                }}
                showInstallButton={showInstallButton}
            />

            {/* Table Header */}
            <div className="grid grid-cols-12 gap-4 px-6 py-3 border-b border-gray-800 bg-gray-900/50 text-xs font-medium text-gray-500 uppercase tracking-wider">
                <div className="col-span-3 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('name')}>
                    Release {renderSortIndicator('name')}
                </div>
                <div className="col-span-2 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('chart')}>
                    Chart {renderSortIndicator('chart')}
                </div>
                <div className="col-span-1 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('version')}>
                    Version {renderSortIndicator('version')}
                </div>
                <div className="col-span-1 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('namespace')}>
                    Namespace {renderSortIndicator('namespace')}
                </div>
                <div className="col-span-2 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('status')}>
                    Status {renderSortIndicator('status')}
                </div>
                <div className="col-span-2 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('updated')}>
                    Updated {renderSortIndicator('updated')}
                </div>
                <div className="col-span-1"></div>
            </div>

            {/* Table Body */}
            <div className="flex-1 overflow-y-auto">
                {sortedReleases.map((release) => {
                    const releaseKey = `${release.namespace}/${release.name}`;
                    const isExpanded = expandedId === releaseKey;
                    return (
                        <div key={releaseKey} className="border-b border-gray-800 last:border-0">
                            <div
                                onClick={() => toggleExpand(releaseKey)}
                                className={`grid grid-cols-12 gap-4 px-6 py-4 cursor-pointer transition-colors duration-200 items-center ${getExpandableRowRowClasses(isExpanded)}`}
                            >
                                <div className="col-span-3 flex items-center font-medium text-sm text-gray-200">
                                    <ChevronDown
                                        size={16}
                                        className={`mr-2 text-gray-500 transition-transform duration-200 ${isExpanded ? 'transform rotate-180' : ''}`}
                                    />
                                    <Package size={16} className="mr-3 text-gray-500" />
                                    <div className="min-w-0 flex-1">
                                        <span className="truncate block" title={release.name}>{release.name}</span>
                                        {release.description && (
                                            <div className="text-xs text-gray-500 truncate">{release.description}</div>
                                        )}
                                    </div>
                                </div>
                                <div className="col-span-2 text-sm text-gray-300">
                                    {release.chart || '-'}
                                    {release.appVersion && (
                                        <div className="text-xs text-gray-500">App: {release.appVersion}</div>
                                    )}
                                </div>
                                <div className="col-span-1 text-sm text-gray-300">{release.version || '-'}</div>
                                <div className="col-span-1 text-sm text-gray-300">{release.namespace}</div>
                                <div className="col-span-2">
                                    {getStatusBadge(release.status)}
                                </div>
                                <div className="col-span-2 text-sm text-gray-400">
                                    {formatDateTime(release.updated)}
                                </div>
                                <div className="col-span-1 flex justify-end" onClick={(e) => e.stopPropagation()}>
                                    <div className="relative">
                                        <button
                                            onClick={() => setMenuOpen(menuOpen === releaseKey ? null : releaseKey)}
                                            className="p-1 hover:bg-gray-800 rounded text-gray-400 hover:text-white transition-colors"
                                        >
                                            <MoreVertical size={16} />
                                        </button>
                                        {menuOpen === releaseKey && (
                                            <div className="absolute right-0 mt-1 w-40 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
                                                <div className="flex flex-col">
                                                    {hasEditPermission(release.namespace) && (
                                                        <>
                                                            <button
                                                                onClick={() => {
                                                                    setUpgradeRelease(release);
                                                                    setMenuOpen(null);
                                                                }}
                                                                className="w-full text-left px-4 py-2 text-sm text-blue-300 hover:bg-blue-900/40 flex items-center"
                                                            >
                                                                <ArrowUp size={14} className="mr-2" />
                                                                Upgrade
                                                            </button>
                                                            <button
                                                                onClick={() => {
                                                                    setConfirmAction({ release });
                                                                    setMenuOpen(null);
                                                                }}
                                                                className="w-full text-left px-4 py-2 text-sm text-red-300 hover:bg-red-900/40 flex items-center"
                                                            >
                                                                <Trash2 size={14} className="mr-2" />
                                                                Uninstall
                                                            </button>
                                                        </>
                                                    )}
                                                    {!hasEditPermission(release.namespace) && (
                                                        <div className="px-4 py-2 text-xs text-gray-500">
                                                            View only
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>

                            {/* Expanded Details */}
                            <div className={`${getExpandableRowClasses(isExpanded, false)}`}>
                                {isExpanded && (
                                    <div className="px-6 py-4 bg-gray-900/30 border-t border-gray-800">
                                        <div className="bg-gray-900/50 rounded-lg border border-gray-800 overflow-hidden">
                                            <div className="p-4 space-y-6">
                                                {/* Basic Information */}
                                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                    <div>
                                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                                            <Package size={12} className="mr-1" />
                                                            Chart
                                                        </h4>
                                                        <div className="text-sm text-gray-300">{release.chart || '-'}</div>
                                                        {release.appVersion && (
                                                            <div className="text-xs text-gray-500 mt-1">App Version: {release.appVersion}</div>
                                                        )}
                                                    </div>
                                                    <div>
                                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                                            <Tag size={12} className="mr-1" />
                                                            Version
                                                        </h4>
                                                        <div className="text-sm text-gray-300">{release.version || '-'}</div>
                                                    </div>
                                                    <div>
                                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                                            <Clock size={12} className="mr-1" />
                                                            Last Updated
                                                        </h4>
                                                        <div className="text-sm text-gray-300">{formatDateTime(release.updated)}</div>
                                                    </div>
                                                    <div>
                                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                                            <Tag size={12} className="mr-1" />
                                                            Revision
                                                        </h4>
                                                        <div className="text-sm text-gray-300">{release.revision || '-'}</div>
                                                    </div>
                                                </div>

                                                {/* Description */}
                                                {release.description && (
                                                    <div>
                                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Description</h4>
                                                        <div className="text-sm text-gray-300">{release.description}</div>
                                                    </div>
                                                )}

                                                {/* Status */}
                                                <div>
                                                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Status</h4>
                                                    {getStatusBadge(release.status)}
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    );
                })}
                {sortedReleases.length === 0 && (
                    <div className="p-8 text-center text-gray-500">
                        {filter ? 'No releases match your filter.' : 'No Helm releases found.'}
                    </div>
                )}
            </div>

            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(null)}
                ></div>
            )}

            <HelmUpgradeModal
                release={upgradeRelease}
                isOpen={!!upgradeRelease}
                onClose={() => setUpgradeRelease(null)}
                onSuccess={handleSuccess}
                currentCluster={currentCluster}
            />

            <HelmInstallModal
                isOpen={installModalOpen}
                onClose={() => setInstallModalOpen(false)}
                onSuccess={handleSuccess}
                currentCluster={currentCluster}
                prefilledNamespace={installNamespace}
            />

            {confirmAction && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                        <h3 className="text-lg font-semibold text-white mb-2">
                            Confirm Uninstall
                        </h3>
                        <p className="text-sm text-gray-300 mb-4">
                            Are you sure you want to uninstall Helm release <span className="font-bold text-white">"{confirmAction.release.name}"</span> from namespace <span className="font-bold text-white">"{confirmAction.release.namespace}"</span>?
                            <br />
                            <span className="text-sm text-gray-500 mt-2 block">This action cannot be undone. All resources managed by this Helm release will be deleted.</span>
                        </p>
                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setConfirmAction(null)}
                                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={async () => {
                                    await handleDelete(confirmAction.release);
                                    setConfirmAction(null);
                                }}
                                className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors"
                            >
                                Uninstall
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default HelmChartManager;
