import React, { useEffect, useState, useRef } from 'react';
import { Activity, RefreshCw, Tag, Plus, MoreVertical, FileText, Trash2, AlertCircle, Globe, MapPin } from 'lucide-react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import QuotaEditor from './QuotaEditor';
import LimitRangeEditor from './LimitRangeEditor';
import YamlEditor from './YamlEditor';
import { calculatePercentage } from '../utils/resourceParser';
import { useResourceQuotas } from '../hooks/useResourceQuotas';
import { useNamespaces } from '../hooks/useNamespaces';
import { DEFAULT_NAMESPACE } from '../config/constants';

const ResourceQuotaManager = ({ namespace }) => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const toast = useToast();
    const location = useLocation();
    const navigate = useNavigate();

    const [activeTab, setActiveTab] = useState('quotas'); // 'quotas' or 'limits'
    const [editingQuota, setEditingQuota] = useState(null);
    const [editingLimitRange, setEditingLimitRange] = useState(null);
    const [editingYaml, setEditingYaml] = useState(null);
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [namespaceFilter, setNamespaceFilter] = useState('all'); // 'all' or specific namespace
    const [createMenuOpen, setCreateMenuOpen] = useState(false);
    const createMenuRef = useRef(null);

    // Fetch namespaces for the dropdown/create menu
    const { data: namespaces = [] } = useNamespaces(authFetch, currentCluster);

    // Fetch quotas and limit ranges
    const { quotas: quotasQuery, limitRanges: limitRangesQuery } = useResourceQuotas(authFetch, namespaceFilter, currentCluster);

    const quotas = quotasQuery.data || [];
    const limitRanges = limitRangesQuery.data || [];
    const loading = quotasQuery.isLoading || limitRangesQuery.isLoading;

    // Initialize namespaceFilter from prop
    useEffect(() => {
        if (namespace) {
            setNamespaceFilter(namespace);
        }
    }, [namespace]);

    // Sync tab with URL
    useEffect(() => {
        const params = new URLSearchParams(location.search);
        const tab = params.get('tab');
        if (tab === 'quotas' || tab === 'limits') {
            setActiveTab(tab);
        }
    }, [location.search]);

    const handleTabChange = (tab) => {
        setActiveTab(tab);
        const params = new URLSearchParams(location.search);
        params.set('tab', tab);
        navigate(`${location.pathname}?${params.toString()}`, { replace: true });
    };

    // Close create menu when clicking outside
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (createMenuRef.current && !createMenuRef.current.contains(event.target)) {
                setCreateMenuOpen(false);
            }
        };
        if (createMenuOpen) {
            document.addEventListener('mousedown', handleClickOutside);
        }
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [createMenuOpen]);


    const handleDelete = async (resource, kind, force = false) => {
        const params = new URLSearchParams({
            kind: kind,
            name: resource.name
        });
        if (resource.namespace) params.append('namespace', resource.namespace);
        if (force) params.append('force', 'true');
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const res = await authFetch(`/api/resource?${params.toString()}`, {
                method: 'DELETE'
            });

            if (!res.ok) {
                const errorText = await res.text();
                throw new Error(errorText || 'Failed to delete resource');
            }

            // Close the confirmation modal first
            setConfirmAction(null);

            // Refresh the appropriate list
            if (kind === 'ResourceQuota') {
                quotasQuery.refetch();
            } else {
                limitRangesQuery.refetch();
            }
        } catch (err) {
            toast.error(`Error deleting ${kind}: ${err.message}`);
            throw err; // Re-throw to allow caller to handle
        }
    };

    const getAge = (created) => {
        if (!created) return 'Unknown';
        const diff = Date.now() - new Date(created).getTime();
        const days = Math.floor(diff / (1000 * 60 * 60 * 24));
        if (days > 0) return `${days}d`;
        const hours = Math.floor(diff / (1000 * 60 * 60));
        if (hours > 0) return `${hours}h`;
        const minutes = Math.floor(diff / (1000 * 60));
        return `${minutes}m`;
    };

    const ProgressBar = ({ used, hard, label }) => {
        const percentage = calculatePercentage(used, hard);
        let colorClass = 'bg-blue-500';
        if (percentage > 90) colorClass = 'bg-red-500';
        else if (percentage > 75) colorClass = 'bg-yellow-500';

        return (
            <div className="mb-2 last:mb-0">
                <div className="flex justify-between text-xs mb-0.5">
                    <span className="text-gray-400 font-medium">{label}</span>
                    <span className="text-gray-300">
                        {used} / {hard} <span className="ml-1 text-gray-500">({percentage}%)</span>
                    </span>
                </div>
                <div className="w-full bg-gray-700 rounded-full h-1.5 overflow-hidden">
                    <div
                        className={`h-1.5 rounded-full transition-all duration-500 ${colorClass}`}
                        style={{ width: `${percentage}%` }}
                    ></div>
                </div>
            </div>
        );
    };

    const renderQuotaUsage = (resource) => {
        if (!resource.details?.hard || Object.keys(resource.details.hard).length === 0) {
            return <span className="text-gray-500 italic text-xs">No limits configured</span>;
        }

        const hard = resource.details.hard;
        const used = resource.details.used || {};

        // Prioritize CPU and Memory
        const priorityKeys = ['requests.cpu', 'limits.cpu', 'requests.memory', 'limits.memory'];
        const otherKeys = Object.keys(hard).filter(k => !priorityKeys.includes(k));

        return (
            <div className="space-y-2 max-w-md">
                {priorityKeys.map(key => {
                    if (hard[key]) {
                        return (
                            <ProgressBar
                                key={key}
                                label={key}
                                used={used[key] || '0'}
                                hard={hard[key]}
                            />
                        );
                    }
                    return null;
                })}
                {otherKeys.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-2">
                        {otherKeys.map(key => (
                            <span key={key} className="bg-gray-800 text-gray-400 px-1.5 py-0.5 rounded text-[10px] border border-gray-700">
                                {key}: {used[key] || '0'}/{hard[key]}
                            </span>
                        ))}
                    </div>
                )}
            </div>
        );
    };

    const renderLimitDetails = (resource) => {
        if (!resource.details?.limits || resource.details.limits.length === 0) {
            return <span className="text-gray-500 italic text-xs">No limits configured</span>;
        }

        return (
            <div className="space-y-2">
                {resource.details.limits.map((limit, idx) => (
                    <div key={idx} className="bg-gray-800/50 rounded p-2 border border-gray-700/50 text-xs">
                        <div className="font-bold text-gray-400 mb-1 uppercase tracking-wider">{limit.type || 'Container'}</div>
                        <div className="grid grid-cols-2 gap-x-4 gap-y-1">
                            {limit.max && Object.entries(limit.max).map(([k, v]) => (
                                <div key={`max-${k}`} className="text-gray-400">
                                    <span className="text-gray-500">Max {k}:</span> <span className="text-gray-200">{v}</span>
                                </div>
                            ))}
                            {limit.min && Object.entries(limit.min).map(([k, v]) => (
                                <div key={`min-${k}`} className="text-gray-400">
                                    <span className="text-gray-500">Min {k}:</span> <span className="text-gray-200">{v}</span>
                                </div>
                            ))}
                            {limit.defaultRequest && Object.entries(limit.defaultRequest).map(([k, v]) => (
                                <div key={`dr-${k}`} className="text-gray-400 col-span-2">
                                    <span className="text-gray-500">Default Req {k}:</span> <span className="text-gray-200">{v}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                ))}
            </div>
        );
    };

    const renderTable = (data, type) => {
        const isQuota = type === 'quota';

        if (data.length === 0) {
            return (
                <div className="flex flex-col items-center justify-center py-20 text-gray-500 bg-gray-900/30 rounded-lg border border-gray-800 border-dashed">
                    {isQuota ? <Activity size={48} className="mb-4 opacity-20" /> : <Tag size={48} className="mb-4 opacity-20" />}
                    <p className="text-lg">No {isQuota ? 'resource quotas' : 'limit ranges'} found</p>
                    <p className="text-sm opacity-60">Create one to get started</p>
                </div>
            );
        }

        return (
            <div className="bg-gray-900 border border-gray-800 rounded-lg overflow-hidden">
                <table className="w-full text-left border-collapse">
                    <thead>
                        <tr className="bg-gray-800/50 text-gray-400 text-xs uppercase tracking-wider border-b border-gray-700">
                            <th className="px-6 py-3 font-medium">Name</th>
                            <th className="px-6 py-3 font-medium">Namespace</th>
                            <th className="px-6 py-3 font-medium w-24 text-center">Age</th>
                            <th className="px-6 py-3 font-medium">{isQuota ? 'Usage / Limits' : 'Details'}</th>
                            <th className="px-6 py-3 font-medium w-20 text-right">Actions</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-800">
                        {data.map((resource) => (
                            <tr key={`${resource.namespace}-${resource.name}`} className="hover:bg-gray-800/30 transition-colors">
                                <td className="px-6 py-4 align-top">
                                    <div className="flex items-center">
                                        {isQuota ? <Activity size={16} className="text-blue-400 mr-2" /> : <Tag size={16} className="text-green-400 mr-2" />}
                                        <span className="font-medium text-gray-200">{resource.name}</span>
                                    </div>
                                </td>
                                <td className="px-6 py-4 align-top text-sm text-gray-300">
                                    {resource.namespace}
                                </td>
                                <td className="px-6 py-4 align-top text-sm text-gray-400 text-center whitespace-nowrap">
                                    {getAge(resource.created)}
                                </td>
                                <td className="px-6 py-4 align-top">
                                    {isQuota ? renderQuotaUsage(resource) : renderLimitDetails(resource)}
                                </td>
                                <td className="px-6 py-4 align-top text-right">
                                    <div className="relative inline-block text-left">
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                const uniqueId = `${type}-${resource.namespace}-${resource.name}`;
                                                setMenuOpen(menuOpen === uniqueId ? null : uniqueId);
                                            }}
                                            className="p-1.5 hover:bg-gray-700 rounded-lg text-gray-400 hover:text-white transition-colors"
                                        >
                                            <MoreVertical size={16} />
                                        </button>
                                        {menuOpen === `${type}-${resource.namespace}-${resource.name}` && (
                                            <div className="absolute right-0 mt-2 w-48 bg-gray-800 border border-gray-700 rounded-lg shadow-xl z-50 py-1 animate-in fade-in zoom-in-95 duration-100">
                                                <button
                                                    onClick={() => {
                                                        setEditingYaml({ ...resource, kind: isQuota ? 'ResourceQuota' : 'LimitRange' });
                                                        setMenuOpen(null);
                                                    }}
                                                    className="w-full text-left px-4 py-2.5 text-sm text-gray-300 hover:bg-gray-700 hover:text-white flex items-center"
                                                >
                                                    <FileText size={14} className="mr-2" /> Edit YAML
                                                </button>
                                                <div className="h-px bg-gray-700 my-1"></div>
                                                <button
                                                    onClick={() => {
                                                        setConfirmAction({ resource, kind: isQuota ? 'ResourceQuota' : 'LimitRange', force: false });
                                                        setMenuOpen(null);
                                                    }}
                                                    className="w-full text-left px-4 py-2.5 text-sm text-red-400 hover:bg-red-900/20 flex items-center"
                                                >
                                                    <Trash2 size={14} className="mr-2" /> Delete
                                                </button>
                                            </div>
                                        )}
                                    </div>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        );
    };

    return (
        <div className="p-6">
            {/* Header Section */}
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center space-x-3">
                    {activeTab === 'quotas' ? <Activity className="text-blue-400" size={24} /> : <Tag className="text-green-400" size={24} />}
                    <h1 className="text-2xl font-semibold text-white">
                        {activeTab === 'quotas' ? 'Resource Quotas' : 'Limit Ranges'}
                    </h1>
                    {loading && <RefreshCw size={18} className="animate-spin text-gray-400 ml-2" />}
                </div>
                <div className="flex items-center space-x-3">
                    <div className="bg-gray-800 border border-gray-700 rounded-md flex overflow-hidden text-sm shrink-0">
                        <button
                            onClick={() => setNamespaceFilter('all')}
                            className={`px-3 py-1.5 flex items-center space-x-1 ${namespaceFilter === 'all' ? 'bg-gray-700 text-white' : 'text-gray-300 hover:bg-gray-700'}`}
                            title="Show all namespaces"
                        >
                            <Globe size={14} /> <span className="hidden sm:inline">All</span>
                        </button>
                        <button
                            onClick={() => {
                                // Toggle between 'all' and the namespace from header
                                if (namespaceFilter === 'all' && namespace && namespace !== 'all') {
                                    setNamespaceFilter(namespace);
                                } else {
                                    setNamespaceFilter('all');
                                }
                            }}
                            className={`px-3 py-1.5 flex items-center space-x-1 ${namespaceFilter !== 'all' ? 'bg-gray-700 text-white' : 'text-gray-300 hover:bg-gray-700'}`}
                            title={namespaceFilter !== 'all' ? `Showing namespace: ${namespaceFilter}` : namespace && namespace !== 'all' ? `Click to show namespace: ${namespace}` : 'No namespace selected'}
                        >
                            <MapPin size={14} /> <span className="hidden sm:inline max-w-[120px] truncate">{namespaceFilter !== 'all' ? namespaceFilter : (namespace && namespace !== 'all' ? namespace : 'Namespace')}</span>
                        </button>
                    </div>
                    <button
                        onClick={() => {
                            quotasQuery.refetch();
                            limitRangesQuery.refetch();
                        }}
                        className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-white rounded-md border border-gray-700 text-sm transition-colors flex items-center"
                        title="Refresh"
                    >
                        <RefreshCw size={16} className={loading ? "animate-spin mr-2" : "mr-2"} />
                        Refresh
                    </button>

                    {/* Direct Add Button based on active tab */}
                    <button
                        onClick={() => {
                            const selectedNs = namespaceFilter !== 'all' ? namespaceFilter : (namespace && namespace !== 'all' ? namespace : (namespaces[0]?.name || DEFAULT_NAMESPACE));
                            if (activeTab === 'quotas') {
                                setEditingQuota({ namespace: selectedNs, name: '', kind: 'ResourceQuota', isNew: true });
                            } else {
                                setEditingLimitRange({ namespace: selectedNs, name: '', kind: 'LimitRange', isNew: true });
                            }
                        }}
                        className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md border border-blue-500 text-sm transition-colors flex items-center shadow-lg shadow-blue-900/20"
                    >
                        <Plus size={16} className="mr-2" />
                        Add {activeTab === 'quotas' ? 'Quota' : 'Limit Range'}...
                    </button>
                </div>
            </div>

            {/* Tabs */}
            <div className="flex space-x-1 bg-gray-800/50 p-1 rounded-lg w-fit mb-6 border border-gray-700/50">
                <button
                    onClick={() => handleTabChange('quotas')}
                    className={`px-6 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 flex items-center ${activeTab === 'quotas'
                        ? 'bg-gray-700 text-white shadow-sm'
                        : 'text-gray-400 hover:text-white hover:bg-gray-700/50'
                        }`}
                >
                    <Activity size={16} className="mr-2" />
                    Resource Quotas
                    <span className="ml-2 bg-black/20 px-2 py-0.5 rounded-full text-xs">{quotas.length}</span>
                </button>
                <button
                    onClick={() => handleTabChange('limits')}
                    className={`px-6 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 flex items-center ${activeTab === 'limits'
                        ? 'bg-gray-700 text-white shadow-sm'
                        : 'text-gray-400 hover:text-white hover:bg-gray-700/50'
                        }`}
                >
                    <Tag size={16} className="mr-2" />
                    Limit Ranges
                    <span className="ml-2 bg-black/20 px-2 py-0.5 rounded-full text-xs">{limitRanges.length}</span>
                </button>
            </div>

            {/* Content Table */}
            <div className="animate-in fade-in duration-300">
                {activeTab === 'quotas' ? renderTable(quotas, 'quota') : renderTable(limitRanges, 'limit')}
            </div>

            {/* Modals */}
            {menuOpen && (
                <div className="fixed inset-0 z-40" onClick={() => setMenuOpen(null)}></div>
            )}

            {editingQuota && (
                <QuotaEditor
                    resource={editingQuota}
                    onClose={() => setEditingQuota(null)}
                    onSaved={() => {
                        setEditingQuota(null);
                        quotasQuery.refetch();
                    }}
                />
            )}

            {editingLimitRange && (
                <LimitRangeEditor
                    resource={editingLimitRange}
                    onClose={() => setEditingLimitRange(null)}
                    onSaved={() => {
                        setEditingLimitRange(null);
                        limitRangesQuery.refetch();
                    }}
                />
            )}

            {editingYaml && (
                <YamlEditor
                    resource={editingYaml}
                    onClose={() => setEditingYaml(null)}
                    onSaved={() => {
                        setEditingYaml(null);
                        if (activeTab === 'quotas') {
                            quotasQuery.refetch();
                        } else {
                            limitRangesQuery.refetch();
                        }
                    }}
                />
            )}

            {confirmAction && (
                <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4 animate-in fade-in duration-200">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                        <div className="flex items-center space-x-3 mb-4 text-red-400">
                            <AlertCircle size={24} />
                            <h3 className="text-xl font-bold text-white">Confirm Deletion</h3>
                        </div>
                        <p className="text-gray-300 mb-6 leading-relaxed">
                            Are you sure you want to {confirmAction.force ? 'force ' : ''}delete the {confirmAction.kind === 'ResourceQuota' ? 'quota' : 'limit range'} <span className="font-bold text-white">"{confirmAction.resource.name}"</span>?
                            <br />
                            <span className="text-sm text-gray-500 mt-2 block">This action cannot be undone.</span>
                        </p>
                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setConfirmAction(null)}
                                className="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-lg transition-colors font-medium"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={async () => {
                                    await handleDelete(confirmAction.resource, confirmAction.kind, confirmAction.force);
                                    // handleDelete already closes the modal and refreshes
                                }}
                                className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors font-medium shadow-lg shadow-red-900/30"
                            >
                                {confirmAction.force ? 'Force Delete' : 'Delete'}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ResourceQuotaManager;
