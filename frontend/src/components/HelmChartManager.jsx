import React, { useState } from 'react';
import { Package, RefreshCw, Clock, Tag, MoreVertical, Trash2, CirclePlus, CircleMinus, ArrowUp, X, Info, Download } from 'lucide-react';
import Editor from '@monaco-editor/react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { formatDateTimeShort } from '../utils/dateUtils';
import { getExpandableRowClasses, getExpandableCellClasses, getExpandableRowRowClasses } from '../utils/expandableRow';
import { useHelmReleases } from '../hooks/useHelmReleases';

const HelmChartManager = () => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [expandedReleases, setExpandedReleases] = useState({});
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [upgradeRelease, setUpgradeRelease] = useState(null);
    const [upgradeForm, setUpgradeForm] = useState({
        chart: '',
        version: '',
        repo: '',
        valuesYaml: ''
    });
    const [upgrading, setUpgrading] = useState(false);
    const [installModalOpen, setInstallModalOpen] = useState(false);
    const [installForm, setInstallForm] = useState({
        name: '',
        namespace: '',
        chart: '',
        version: '',
        repo: '',
        valuesYaml: ''
    });
    const [installing, setInstalling] = useState(false);

    const { data: releases = [], isLoading: loading, refetch } = useHelmReleases(authFetch, currentCluster);

    const toggleExpand = (releaseKey) => {
        setExpandedReleases(prev => ({ ...prev, [releaseKey]: !prev[releaseKey] }));
    };

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

    const getAge = (updated) => {
        if (!updated) return 'Unknown';
        try {
            const diff = Date.now() - new Date(updated).getTime();
            const days = Math.floor(diff / (1000 * 60 * 60 * 24));
            if (days > 0) return `${days}d`;
            const hours = Math.floor(diff / (1000 * 60 * 60));
            if (hours > 0) return `${hours}h`;
            const minutes = Math.floor(diff / (1000 * 60));
            return `${minutes}m`;
        } catch {
            return 'Unknown';
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
            alert(`Error uninstalling Helm release: ${err.message}`);
        }
    };

    const handleUpgrade = async () => {
        if (!upgradeRelease) return;

        setUpgrading(true);
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const payload = {
                name: upgradeRelease.name,
                namespace: upgradeRelease.namespace,
                chart: upgradeForm.chart || upgradeRelease.chart,
                version: upgradeForm.version || undefined,
                repo: upgradeForm.repo || undefined,
                valuesYaml: upgradeForm.valuesYaml || undefined
            };

            const res = await authFetch(`/api/helm/releases?${params.toString()}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            });

            if (!res.ok) {
                const errorData = await res.json().catch(() => ({ message: 'Failed to upgrade Helm release' }));
                throw new Error(errorData.message || 'Failed to upgrade Helm release');
            }

            const result = await res.json();
            setUpgradeRelease(null);
            setUpgradeForm({ chart: '', version: '', repo: '', valuesYaml: '' });

            // Show success message
            alert(`Upgrade initiated! Job: ${result.job || 'created'}`);

            // Refresh the list after a delay
            setTimeout(() => {
                refetch();
            }, 2000);
        } catch (err) {
            alert(`Error upgrading Helm release: ${err.message}`);
        } finally {
            setUpgrading(false);
        }
    };

    const handleInstall = async () => {
        if (!installForm.name || !installForm.namespace || !installForm.chart) {
            alert('Please fill in all required fields (Name, Namespace, Chart)');
            return;
        }

        setInstalling(true);
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const payload = {
                name: installForm.name,
                namespace: installForm.namespace,
                chart: installForm.chart,
                version: installForm.version || undefined,
                repo: installForm.repo || undefined,
                valuesYaml: installForm.valuesYaml || undefined
            };

            const res = await authFetch(`/api/helm/releases/install?${params.toString()}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            });

            if (!res.ok) {
                const errorData = await res.json().catch(() => ({ message: 'Failed to install Helm chart' }));
                throw new Error(errorData.message || 'Failed to install Helm chart');
            }

            const result = await res.json();
            setInstallModalOpen(false);
            setInstallForm({ name: '', namespace: '', chart: '', version: '', repo: '', valuesYaml: '' });

            // Show success message
            alert(`Installation initiated! Job: ${result.job || 'created'}`);

            // Refresh the list after a delay
            setTimeout(() => {
                refetch();
            }, 2000);
        } catch (err) {
            alert(`Error installing Helm chart: ${err.message}`);
        } finally {
            setInstalling(false);
        }
    };

    return (
        <div className="p-6">
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center space-x-2">
                    <Package className="text-blue-400" size={20} />
                    <h1 className="text-xl font-semibold text-white">Helm Charts</h1>
                    {loading && <RefreshCw size={16} className="animate-spin text-gray-400" />}
                </div>
                <div className="flex items-center space-x-2">
                    <button
                        onClick={() => setInstallModalOpen(true)}
                        className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-sm transition-colors flex items-center"
                    >
                        <Download size={14} className="mr-2" />
                        Install Chart
                    </button>
                    <button
                        onClick={() => refetch()}
                        className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-white rounded-md border border-gray-700 text-sm transition-colors flex items-center"
                    >
                        <RefreshCw size={14} className="mr-2" />
                        Refresh
                    </button>
                </div>
            </div>

            <div className="bg-gray-800 border border-gray-700 rounded-lg overflow-x-auto">
                <table className="min-w-full border-separate border-spacing-0">
                    <thead>
                        <tr>
                            <th className="w-10 px-2 md:px-4 py-3 bg-gray-900 rounded-tl-lg border-b border-gray-700"></th>
                            <th className="px-3 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider bg-gray-900 border-b border-gray-700">Release</th>
                            <th className="px-2 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider bg-gray-900 border-b border-gray-700">Chart</th>
                            <th className="px-2 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider bg-gray-900 border-b border-gray-700">Version</th>
                            <th className="px-2 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider bg-gray-900 border-b border-gray-700">Namespace</th>
                            <th className="px-2 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider bg-gray-900 border-b border-gray-700">Status</th>
                            <th className="px-2 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider bg-gray-900 border-b border-gray-700">Revision</th>
                            <th className="px-2 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider bg-gray-900 border-b border-gray-700">Updated</th>
                            <th className="w-10 px-2 md:px-4 py-3 bg-gray-900 rounded-tr-lg border-b border-gray-700"></th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-800">
                        {releases.map((release) => {
                            const releaseKey = `${release.namespace}/${release.name}`;
                            return (
                                <React.Fragment key={releaseKey}>
                                    <tr
                                        className={getExpandableRowRowClasses(expandedReleases[releaseKey])}
                                        onClick={() => toggleExpand(releaseKey)}
                                    >
                                        <td className="px-2 md:px-4 py-3 whitespace-nowrap text-gray-400 text-center">
                                            {expandedReleases[releaseKey] ? <CircleMinus size={16} /> : <CirclePlus size={16} />}
                                        </td>
                                        <td className="px-3 md:px-6 py-3 whitespace-nowrap">
                                            <div className="flex items-center">
                                                <div className="flex-shrink-0 h-6 w-6 bg-gray-700 rounded flex items-center justify-center text-gray-400">
                                                    <Package size={14} />
                                                </div>
                                                <div className="ml-4">
                                                    <div className="text-sm font-medium text-white">{release.name}</div>
                                                    {release.description && (
                                                        <div className="text-xs text-gray-400 truncate max-w-xs">{release.description}</div>
                                                    )}
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-2 md:px-6 py-3 whitespace-nowrap">
                                            <div className="text-sm text-gray-300">{release.chart || '-'}</div>
                                            {release.appVersion && (
                                                <div className="text-xs text-gray-500">App: {release.appVersion}</div>
                                            )}
                                        </td>
                                        <td className="px-2 md:px-6 py-3 whitespace-nowrap text-sm text-gray-300">{release.version || '-'}</td>
                                        <td className="px-2 md:px-6 py-3 whitespace-nowrap text-sm text-gray-300">{release.namespace}</td>
                                        <td className="px-2 md:px-6 py-3 whitespace-nowrap">
                                            {getStatusBadge(release.status)}
                                        </td>
                                        <td className="px-2 md:px-6 py-3 whitespace-nowrap text-sm text-gray-300">{release.revision || '-'}</td>
                                        <td className="px-2 md:px-6 py-3 whitespace-nowrap text-sm text-gray-400">
                                            <div className="flex items-center space-x-1">
                                                <Clock size={12} />
                                                <span>{getAge(release.updated)}</span>
                                            </div>
                                        </td>
                                        <td className="px-4 py-3 whitespace-nowrap text-gray-300" onClick={(e) => e.stopPropagation()}>
                                            <div className="relative flex items-center justify-end">
                                                <button
                                                    onClick={() => setMenuOpen(menuOpen === releaseKey ? null : releaseKey)}
                                                    className="p-1 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                                                >
                                                    <MoreVertical size={16} />
                                                </button>
                                                {menuOpen === releaseKey && (
                                                    <div className="absolute right-0 mt-1 w-40 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
                                                        <div className="flex flex-col">
                                                            <button
                                                                onClick={() => {
                                                                    setUpgradeRelease(release);
                                                                    setUpgradeForm({
                                                                        chart: release.chart || '',
                                                                        version: '',
                                                                        repo: '',
                                                                        valuesYaml: ''
                                                                    });
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
                                                        </div>
                                                    </div>
                                                )}
                                            </div>
                                        </td>
                                    </tr>
                                    <tr>
                                        <td colSpan="9" className={getExpandableCellClasses(expandedReleases[releaseKey], 9)}>
                                            <div className={getExpandableRowClasses(expandedReleases[releaseKey], false)}>
                                                {expandedReleases[releaseKey] && (
                                                    <div className="p-4 bg-gray-900/50 rounded-md space-y-6">
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
                                                                <div className="text-sm text-gray-300">{formatDateTimeShort(release.updated)}</div>
                                                                <div className="text-xs text-gray-500 mt-1">Age: {getAge(release.updated)}</div>
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
                                                )}
                                            </div>
                                        </td>
                                    </tr>
                                </React.Fragment>
                            );
                        })}
                    </tbody>
                </table>

                {releases.length === 0 && !loading && (
                    <div className="p-6 text-center text-gray-500">No Helm releases found</div>
                )}
            </div>

            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(null)}
                ></div>
            )}

            {upgradeRelease && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col shadow-xl">
                        {/* Header */}
                        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-700 bg-gray-800">
                            <div className="flex items-center space-x-3">
                                <ArrowUp className="text-blue-400" size={20} />
                                <div>
                                    <h3 className="text-lg font-semibold text-white">
                                        Upgrade Helm Release
                                    </h3>
                                    <p className="text-xs text-gray-400 mt-0.5">
                                        {upgradeRelease.name} <span className="text-gray-500">•</span> {upgradeRelease.namespace}
                                    </p>
                                </div>
                            </div>
                            <button
                                onClick={() => {
                                    setUpgradeRelease(null);
                                    setUpgradeForm({ chart: '', version: '', repo: '', valuesYaml: '' });
                                }}
                                className="p-1.5 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                            >
                                <X size={20} />
                            </button>
                        </div>

                        {/* Current Release Info */}
                        <div className="px-6 py-3 bg-gray-800/50 border-b border-gray-700">
                            <div className="flex items-start space-x-2">
                                <Info size={16} className="text-blue-400 mt-0.5 flex-shrink-0" />
                                <div className="text-xs text-gray-400">
                                    <span className="text-gray-300 font-medium">Current:</span> Chart: <span className="text-white">{upgradeRelease.chart || 'N/A'}</span>
                                    {upgradeRelease.version && (
                                        <> • Version: <span className="text-white">{upgradeRelease.version}</span></>
                                    )}
                                    {upgradeRelease.revision && (
                                        <> • Revision: <span className="text-white">{upgradeRelease.revision}</span></>
                                    )}
                                </div>
                            </div>
                        </div>

                        {/* Content */}
                        <div className="p-6 overflow-y-auto flex-1">
                            <div className="space-y-5">
                                {/* Chart */}
                                <div>
                                    <label className="block text-sm font-medium text-gray-300 mb-2">
                                        Chart Name <span className="text-red-400">*</span>
                                    </label>
                                    <input
                                        type="text"
                                        value={upgradeForm.chart}
                                        onChange={(e) => setUpgradeForm({ ...upgradeForm, chart: e.target.value })}
                                        placeholder="e.g., nginx, vault, prometheus"
                                        className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                    />
                                    {upgradeRelease.chart && !upgradeForm.chart && (
                                        <p className="text-xs text-gray-500 mt-1.5 flex items-center">
                                            <Info size={12} className="mr-1" />
                                            Will use current chart: <span className="text-gray-300 ml-1">{upgradeRelease.chart}</span>
                                        </p>
                                    )}
                                </div>

                                <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                                    {/* Version */}
                                    <div>
                                        <label className="block text-sm font-medium text-gray-300 mb-2">
                                            Chart Version <span className="text-gray-500 text-xs">(optional)</span>
                                        </label>
                                        <input
                                            type="text"
                                            value={upgradeForm.version}
                                            onChange={(e) => setUpgradeForm({ ...upgradeForm, version: e.target.value })}
                                            placeholder="e.g., 1.2.3 (latest if empty)"
                                            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                        />
                                        {upgradeRelease.version && (
                                            <p className="text-xs text-gray-500 mt-1.5">Current: {upgradeRelease.version}</p>
                                        )}
                                    </div>

                                    {/* Repo */}
                                    <div>
                                        <label className="block text-sm font-medium text-gray-300 mb-2">
                                            Repository URL <span className="text-gray-500 text-xs">(optional)</span>
                                        </label>
                                        <input
                                            type="text"
                                            value={upgradeForm.repo}
                                            onChange={(e) => setUpgradeForm({ ...upgradeForm, repo: e.target.value })}
                                            placeholder="e.g., https://charts.helm.sh/stable"
                                            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                        />
                                    </div>
                                </div>

                                {/* Values YAML */}
                                <div>
                                    <label className="block text-sm font-medium text-gray-300 mb-2">
                                        Values Override (YAML) <span className="text-gray-500 text-xs">(optional)</span>
                                    </label>
                                    <div className="border border-gray-700 rounded-md overflow-hidden" style={{ height: '300px' }}>
                                        <Editor
                                            height="100%"
                                            defaultLanguage="yaml"
                                            theme="vs-dark"
                                            value={upgradeForm.valuesYaml}
                                            onChange={(value) => setUpgradeForm({ ...upgradeForm, valuesYaml: value || '' })}
                                            options={{
                                                minimap: { enabled: false },
                                                scrollBeyondLastLine: false,
                                                fontSize: 13,
                                                automaticLayout: true,
                                                lineNumbers: 'on',
                                                wordWrap: 'on',
                                                tabSize: 2,
                                                insertSpaces: true,
                                            }}
                                        />
                                    </div>
                                    <p className="text-xs text-gray-500 mt-1.5">
                                        Override chart values. These will be merged with the default values.
                                    </p>
                                </div>

                                {/* Help Hint */}
                                <div className="mt-4 p-4 bg-blue-900/20 border border-blue-700/50 rounded-md">
                                    <div className="flex items-start space-x-2">
                                        <Info size={16} className="text-blue-400 mt-0.5 flex-shrink-0" />
                                        <div className="text-xs text-blue-300">
                                            <span className="font-medium">💡 Tip:</span> Para obtener los mejores resultados, es recomendable especificar la <span className="font-semibold text-blue-200">URL del repositorio</span> y la <span className="font-semibold text-blue-200">versión del chart</span>. Si no se especifican, el sistema intentará detectarlos automáticamente, pero puede tomar más tiempo o fallar en algunos casos.
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* Footer */}
                        <div className="px-6 py-4 border-t border-gray-700 bg-gray-800 flex justify-between items-center">
                            <div className="text-xs text-gray-500">
                                A Kubernetes Job will be created to execute the upgrade
                            </div>
                            <div className="flex space-x-3">
                                <button
                                    onClick={() => {
                                        setUpgradeRelease(null);
                                        setUpgradeForm({ chart: '', version: '', repo: '', valuesYaml: '' });
                                    }}
                                    className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                                    disabled={upgrading}
                                >
                                    Cancel
                                </button>
                                <button
                                    onClick={handleUpgrade}
                                    disabled={upgrading || !upgradeForm.chart}
                                    className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-md transition-colors flex items-center"
                                >
                                    {upgrading ? (
                                        <>
                                            <RefreshCw size={16} className="mr-2 animate-spin" />
                                            Upgrading...
                                        </>
                                    ) : (
                                        <>
                                            <ArrowUp size={16} className="mr-2" />
                                            Upgrade Release
                                        </>
                                    )}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* Install Chart Modal */}
            {installModalOpen && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col shadow-xl">
                        {/* Header */}
                        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-700 bg-gray-800">
                            <div className="flex items-center space-x-3">
                                <Download className="text-blue-400" size={20} />
                                <div>
                                    <h3 className="text-lg font-semibold text-white">
                                        Install Helm Chart
                                    </h3>
                                </div>
                            </div>
                            <button
                                onClick={() => {
                                    setInstallModalOpen(false);
                                    setInstallForm({ name: '', namespace: '', chart: '', version: '', repo: '', valuesYaml: '' });
                                }}
                                className="p-1.5 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                            >
                                <X size={20} />
                            </button>
                        </div>

                        {/* Content */}
                        <div className="p-6 overflow-y-auto flex-1">
                            <div className="space-y-5">
                                {/* Name and Namespace */}
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                                    <div>
                                        <label className="block text-sm font-medium text-gray-300 mb-2">
                                            Release Name <span className="text-red-400">*</span>
                                        </label>
                                        <input
                                            type="text"
                                            value={installForm.name}
                                            onChange={(e) => setInstallForm({ ...installForm, name: e.target.value })}
                                            placeholder="e.g., my-app, nginx, prometheus"
                                            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-300 mb-2">
                                            Namespace <span className="text-red-400">*</span>
                                        </label>
                                        <input
                                            type="text"
                                            value={installForm.namespace}
                                            onChange={(e) => setInstallForm({ ...installForm, namespace: e.target.value })}
                                            placeholder="e.g., default, production, monitoring"
                                            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                        />
                                    </div>
                                </div>

                                {/* Chart */}
                                <div>
                                    <label className="block text-sm font-medium text-gray-300 mb-2">
                                        Chart Name <span className="text-red-400">*</span>
                                    </label>
                                    <input
                                        type="text"
                                        value={installForm.chart}
                                        onChange={(e) => setInstallForm({ ...installForm, chart: e.target.value })}
                                        placeholder="e.g., nginx, vault, prometheus, kube-prometheus-stack"
                                        className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                    />
                                </div>

                                <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                                    {/* Version */}
                                    <div>
                                        <label className="block text-sm font-medium text-gray-300 mb-2">
                                            Chart Version <span className="text-gray-500 text-xs">(optional)</span>
                                        </label>
                                        <input
                                            type="text"
                                            value={installForm.version}
                                            onChange={(e) => setInstallForm({ ...installForm, version: e.target.value })}
                                            placeholder="e.g., 1.2.3 (latest if empty)"
                                            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                        />
                                    </div>

                                    {/* Repo */}
                                    <div>
                                        <label className="block text-sm font-medium text-gray-300 mb-2">
                                            Repository URL <span className="text-gray-500 text-xs">(optional)</span>
                                        </label>
                                        <input
                                            type="text"
                                            value={installForm.repo}
                                            onChange={(e) => setInstallForm({ ...installForm, repo: e.target.value })}
                                            placeholder="e.g., https://charts.helm.sh/stable"
                                            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                        />
                                    </div>
                                </div>

                                {/* Values YAML */}
                                <div>
                                    <label className="block text-sm font-medium text-gray-300 mb-2">
                                        Values Override (YAML) <span className="text-gray-500 text-xs">(optional)</span>
                                    </label>
                                    <div className="border border-gray-700 rounded-md overflow-hidden" style={{ height: '300px' }}>
                                        <Editor
                                            height="100%"
                                            defaultLanguage="yaml"
                                            theme="vs-dark"
                                            value={installForm.valuesYaml}
                                            onChange={(value) => setInstallForm({ ...installForm, valuesYaml: value || '' })}
                                            options={{
                                                minimap: { enabled: false },
                                                scrollBeyondLastLine: false,
                                                fontSize: 13,
                                                automaticLayout: true,
                                                lineNumbers: 'on',
                                                wordWrap: 'on',
                                                tabSize: 2,
                                                insertSpaces: true,
                                            }}
                                        />
                                    </div>
                                    <p className="text-xs text-gray-500 mt-1.5">
                                        Override chart values. These will be merged with the default values.
                                    </p>
                                </div>

                                {/* Help Hint */}
                                <div className="mt-4 p-4 bg-blue-900/20 border border-blue-700/50 rounded-md">
                                    <div className="flex items-start space-x-2">
                                        <Info size={16} className="text-blue-400 mt-0.5 flex-shrink-0" />
                                        <div className="text-xs text-blue-300">
                                            <span className="font-medium">💡 Tip:</span> Para obtener los mejores resultados, es recomendable especificar la <span className="font-semibold text-blue-200">URL del repositorio</span> y la <span className="font-semibold text-blue-200">versión del chart</span>. Si no se especifican, el sistema intentará detectarlos automáticamente, pero puede tomar más tiempo o fallar en algunos casos.
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* Footer */}
                        <div className="px-6 py-4 border-t border-gray-700 bg-gray-800 flex justify-between items-center">
                            <div className="text-xs text-gray-500">
                                A Kubernetes Job will be created to execute the installation
                            </div>
                            <div className="flex space-x-3">
                                <button
                                    onClick={() => {
                                        setInstallModalOpen(false);
                                        setInstallForm({ name: '', namespace: '', chart: '', version: '', repo: '', valuesYaml: '' });
                                    }}
                                    className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                                    disabled={installing}
                                >
                                    Cancel
                                </button>
                                <button
                                    onClick={handleInstall}
                                    disabled={installing || !installForm.name || !installForm.namespace || !installForm.chart}
                                    className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-md transition-colors flex items-center"
                                >
                                    {installing ? (
                                        <>
                                            <RefreshCw size={16} className="mr-2 animate-spin" />
                                            Installing...
                                        </>
                                    ) : (
                                        <>
                                            <Download size={16} className="mr-2" />
                                            Install Chart
                                        </>
                                    )}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}

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
