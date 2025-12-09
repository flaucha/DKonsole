import React, { useState } from 'react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import { useHelmReleases } from '../hooks/useHelmReleases';
import HelmToolbar from './helm/HelmToolbar';
import HelmUpgradeModal from './helm/HelmUpgradeModal';
import HelmInstallModal from './helm/HelmInstallModal';
import HelmReleaseList from './helm/HelmReleaseList';
import HelmDeleteConfirmModal from './helm/HelmDeleteConfirmModal';

const HelmChartManager = ({ namespace }) => {
    const { currentCluster } = useSettings();
    const { authFetch, user } = useAuth();
    const toast = useToast();

    // Check if user is admin (core admin has role='admin', LDAP admin groups have no permissions but are admins)
    const isAdmin = user && user.role === 'admin';
    const [sortField, setSortField] = useState('name');
    const [sortDirection, setSortDirection] = useState('asc');
    const [filter, setFilter] = useState('');
    const [confirmAction, setConfirmAction] = useState(null);

    // Modal states
    const [upgradeRelease, setUpgradeRelease] = useState(null);
    const [installModalOpen, setInstallModalOpen] = useState(false);
    const [installNamespace, setInstallNamespace] = useState('');

    const { data: releases = [], isLoading: loading, refetch } = useHelmReleases(authFetch, currentCluster);

    // Helper to check if user has edit or admin permission for a namespace
    const hasEditPermission = (ns) => {
        if (isAdmin) return true;
        if (!user || !user.permissions) return false;
        const permission = user.permissions[ns];
        return permission === 'edit' || permission === 'admin';
    };

    // Helper to check if user has view permission for a namespace
    const hasViewPermission = (ns) => {
        if (isAdmin) return true;
        if (!user || !user.permissions) return false;
        const permission = user.permissions[ns];
        return permission === 'view' || permission === 'edit' || permission === 'admin';
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

    // Filter releases by namespace and search text
    const filteredReleases = releases.filter(release => {
        if (!isAdmin) {
            if (namespace === 'all') {
                if (!hasViewPermission(release.namespace)) return false;
            } else {
                if (release.namespace !== namespace || !hasViewPermission(release.namespace)) {
                    return false;
                }
            }
        }
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
                case 'name': return item.name || '';
                case 'chart': return item.chart || '';
                case 'namespace': return item.namespace || '';
                case 'status': return item.status || '';
                case 'version': return item.version || '';
                case 'updated': return new Date(item.updated).getTime() || 0;
                default: return '';
            }
        };
        const va = getVal(a);
        const vb = getVal(b);
        if (typeof va === 'number' && typeof vb === 'number') {
            return (va - vb) * dir;
        }
        return String(va).localeCompare(String(vb)) * dir;
    });

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

            <HelmReleaseList
                releases={sortedReleases}
                sortField={sortField}
                sortDirection={sortDirection}
                onSort={handleSort}
                onUpgrade={setUpgradeRelease}
                onDelete={(release) => setConfirmAction({ release })}
                hasEditPermission={hasEditPermission}
                filter={filter}
            />

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

            <HelmDeleteConfirmModal
                release={confirmAction?.release}
                onCancel={() => setConfirmAction(null)}
                onConfirm={async () => {
                    await handleDelete(confirmAction.release);
                    setConfirmAction(null);
                }}
            />
        </div>
    );
};

export default HelmChartManager;
