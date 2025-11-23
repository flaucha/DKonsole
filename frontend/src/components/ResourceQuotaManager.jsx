import React, { useEffect, useState, useRef } from 'react';
import { Activity, RefreshCw, Tag, Plus, MoreVertical, FileText, Trash2, AlertCircle, Box, Cpu, HardDrive, Globe, MapPin } from 'lucide-react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import QuotaEditor from './QuotaEditor';
import LimitRangeEditor from './LimitRangeEditor';
import YamlEditor from './YamlEditor';
import { getStatusBadgeClass } from '../utils/statusBadge';
import { calculatePercentage } from '../utils/resourceParser';

const ResourceQuotaManager = ({ namespace }) => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [quotas, setQuotas] = useState([]);
    const [limitRanges, setLimitRanges] = useState([]);
    const [loading, setLoading] = useState(false);
    const [activeTab, setActiveTab] = useState('quotas'); // 'quotas' or 'limits'
    const [editingQuota, setEditingQuota] = useState(null);
    const [editingLimitRange, setEditingLimitRange] = useState(null);
    const [editingYaml, setEditingYaml] = useState(null);
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [namespaces, setNamespaces] = useState([]);
    const [namespaceFilter, setNamespaceFilter] = useState('all'); // 'all' or specific namespace
    const [createMenuOpen, setCreateMenuOpen] = useState(false);
    const createMenuRef = useRef(null);

    const fetchNamespaces = () => {
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        authFetch(`/api/namespaces?${params.toString()}`)
            .then(res => res.json())
            .then(data => {
                setNamespaces(data || []);
            })
            .catch(() => setNamespaces([]));
    };

    const fetchQuotas = () => {
        setLoading(true);
        const params = new URLSearchParams({ kind: 'ResourceQuota', namespace: namespaceFilter });
        if (currentCluster) params.append('cluster', currentCluster);

        authFetch(`/api/resources?${params.toString()}`)
            .then(res => res.json())
            .then(data => {
                setQuotas(data || []);
                setLoading(false);
            })
            .catch(() => {
                setQuotas([]);
                setLoading(false);
            });
    };

    const fetchLimitRanges = () => {
        setLoading(true);
        const params = new URLSearchParams({ kind: 'LimitRange', namespace: namespaceFilter });
        if (currentCluster) params.append('cluster', currentCluster);

        authFetch(`/api/resources?${params.toString()}`)
            .then(res => res.json())
            .then(data => {
                setLimitRanges(data || []);
                setLoading(false);
            })
            .catch(() => {
                setLimitRanges([]);
                setLoading(false);
            });
    };

    const fetchAll = () => {
        fetchNamespaces();
        if (activeTab === 'quotas') {
            fetchQuotas();
        } else {
            fetchLimitRanges();
        }
    };

    // Initialize namespaceFilter from prop
    useEffect(() => {
        if (namespace) {
            setNamespaceFilter(namespace);
        }
    }, [namespace]);

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

    useEffect(() => {
        fetchAll();
    }, [currentCluster, activeTab, namespaceFilter]);


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

            // Refresh the appropriate list immediately after successful deletion
            // Use setTimeout to ensure the modal is closed before refreshing
            setTimeout(() => {
                if (kind === 'ResourceQuota') {
                    fetchQuotas();
                } else {
                    fetchLimitRanges();
                }
            }, 100);
        } catch (err) {
            alert(`Error deleting ${kind}: ${err.message}`);
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

        return (
            <div className="mb-3">
                <div className="flex justify-between text-xs mb-1">
                    <span className="text-gray-400 font-medium">{label}</span>
                    <span className="text-gray-300">
                        {used} / {hard} <span className="ml-1 text-gray-500">({percentage}%)</span>
                    </span>
                </div>
                <div className="w-full bg-gray-700 rounded-full h-2 overflow-hidden">
                    <div
                        className="h-2 rounded-full transition-all duration-500 bg-gray-500"
                        style={{ width: `${percentage}%` }}
                    ></div>
                </div>
            </div>
        );
    };

    const ResourceCard = ({ resource, type }) => {
        const isQuota = type === 'quota';
        const uniqueId = `${type}-${resource.namespace}-${resource.name}`;

        return (
            <div className="bg-gray-800 border border-gray-700 rounded-lg p-5 hover:border-gray-600 transition-colors group relative">
                <div className="flex justify-between items-start mb-4">
                    <div className="flex items-center space-x-3">
                        <div className="p-2 rounded-lg bg-gray-700/50 text-gray-400">
                            {isQuota ? <Activity size={20} /> : <Tag size={20} />}
                        </div>
                        <div>
                            <h3 className="text-lg font-bold text-white leading-tight">{resource.name}</h3>
                            <div className="flex items-center text-xs text-gray-400 mt-1">
                                <Box size={12} className="mr-1" />
                                {resource.namespace}
                                <span className="mx-2">•</span>
                                <span>{getAge(resource.created)}</span>
                            </div>
                        </div>
                    </div>
                    <div className="relative">
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                setMenuOpen(menuOpen === uniqueId ? null : uniqueId);
                            }}
                            className="p-1.5 hover:bg-gray-700 rounded-lg text-gray-400 hover:text-white transition-colors"
                        >
                            <MoreVertical size={18} />
                        </button>

                        {menuOpen === uniqueId && (
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
                </div>

                <div className="space-y-4">
                    {isQuota ? (
                        <>
                            {resource.details?.hard && Object.keys(resource.details.hard).length > 0 ? (
                                <div className="space-y-2">
                                    {/* Prioritize CPU and Memory for visual bars */}
                                    {['requests.cpu', 'limits.cpu', 'requests.memory', 'limits.memory'].map(key => {
                                        if (resource.details.hard[key]) {
                                            return (
                                                <ProgressBar
                                                    key={key}
                                                    label={key}
                                                    used={resource.details.used?.[key] || '0'}
                                                    hard={resource.details.hard[key]}
                                                />
                                            );
                                        }
                                        return null;
                                    })}

                                    {/* Show other quotas as simple tags */}
                                    <div className="flex flex-wrap gap-2 mt-3">
                                        {Object.entries(resource.details.hard).filter(([key]) =>
                                            !['requests.cpu', 'limits.cpu', 'requests.memory', 'limits.memory'].includes(key)
                                        ).map(([key, value]) => (
                                            <div key={key} className="bg-gray-700/50 rounded px-2 py-1 text-xs text-gray-300 border border-gray-700">
                                                <span className="text-gray-500 mr-1">{key}:</span>
                                                {resource.details.used?.[key] || '0'} / {value}
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            ) : (
                                <div className="flex flex-col items-center justify-center py-6 text-gray-500 bg-gray-900/30 rounded-lg border border-gray-700/50 border-dashed">
                                    <AlertCircle size={24} className="mb-2 opacity-50" />
                                    <span className="text-sm">No limits configured</span>
                                </div>
                            )}
                        </>
                    ) : (
                        <div className="space-y-3">
                            {resource.details?.limits && resource.details.limits.length > 0 ? (
                                resource.details.limits.map((limit, idx) => (
                                    <div key={idx} className="bg-gray-900/50 rounded-lg p-3 border border-gray-700/50">
                                        <div className="text-xs font-bold text-gray-400 mb-2 uppercase tracking-wider flex items-center">
                                            {limit.type || 'Container'}
                                        </div>
                                        <div className="grid grid-cols-2 gap-2 text-xs">
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
                                ))
                            ) : (
                                <div className="flex flex-col items-center justify-center py-6 text-gray-500 bg-gray-900/30 rounded-lg border border-gray-700/50 border-dashed">
                                    <AlertCircle size={24} className="mb-2 opacity-50" />
                                    <span className="text-sm">No limits configured</span>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>
        );
    };

    return (
        <div className="p-6">
            {/* Header Section */}
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center space-x-2">
                    <Activity className="text-blue-400" size={18} />
                    <h1 className="text-xl font-semibold text-white">Resource Quotas</h1>
                    {loading && <RefreshCw size={16} className="animate-spin text-gray-400" />}
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
                        onClick={fetchAll}
                        className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-white rounded-md border border-gray-700 text-sm transition-colors flex items-center"
                        title="Refresh"
                    >
                        <RefreshCw size={16} className={loading ? "animate-spin mr-2" : "mr-2"} />
                        Refresh
                    </button>
                    <div className="relative" ref={createMenuRef}>
                        <button 
                            onClick={() => setCreateMenuOpen(!createMenuOpen)}
                            className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-white rounded-md border border-gray-700 text-sm transition-colors flex items-center"
                        >
                            <Plus size={16} className="mr-2" />
                            Create New
                        </button>
                        {createMenuOpen && (
                            <div className="absolute right-0 mt-2 w-48 bg-gray-800 border border-gray-700 rounded-lg shadow-xl z-50">
                                <div className="py-1">
                                    <button
                                        onClick={() => {
                                            const selectedNs = namespaceFilter !== 'all' ? namespaceFilter : (namespace && namespace !== 'all' ? namespace : (namespaces[0]?.name || 'default'));
                                            setEditingQuota({ namespace: selectedNs, name: '', kind: 'ResourceQuota', isNew: true });
                                            setActiveTab('quotas'); // Switch to quotas tab when creating a quota
                                            setCreateMenuOpen(false);
                                        }}
                                        className="block w-full text-left px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 hover:text-white"
                                    >
                                        Resource Quota
                                    </button>
                                    <button
                                        onClick={() => {
                                            const selectedNs = namespaceFilter !== 'all' ? namespaceFilter : (namespace && namespace !== 'all' ? namespace : (namespaces[0]?.name || 'default'));
                                            setEditingLimitRange({ namespace: selectedNs, name: '', kind: 'LimitRange', isNew: true });
                                            setActiveTab('limits'); // Switch to limits tab when creating a limit range
                                            setCreateMenuOpen(false);
                                        }}
                                        className="block w-full text-left px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 hover:text-white"
                                    >
                                        Limit Range
                                    </button>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            {/* Tabs */}
            <div className="flex space-x-1 bg-gray-800/50 p-1 rounded-lg w-fit mb-6 border border-gray-700/50">
                <button
                    onClick={() => setActiveTab('quotas')}
                    className={`px-6 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 flex items-center ${activeTab === 'quotas'
                        ? 'bg-gray-700 text-white'
                        : 'text-gray-400 hover:text-white hover:bg-gray-700/50'
                        }`}
                >
                    <Activity size={16} className="mr-2" />
                    Resource Quotas
                    <span className="ml-2 bg-black/20 px-2 py-0.5 rounded-full text-xs">{quotas.length}</span>
                </button>
                <button
                    onClick={() => setActiveTab('limits')}
                    className={`px-6 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 flex items-center ${activeTab === 'limits'
                        ? 'bg-gray-700 text-white'
                        : 'text-gray-400 hover:text-white hover:bg-gray-700/50'
                        }`}
                >
                    <Tag size={16} className="mr-2" />
                    Limit Ranges
                    <span className="ml-2 bg-black/20 px-2 py-0.5 rounded-full text-xs">{limitRanges.length}</span>
                </button>
            </div>

            {/* Content Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {activeTab === 'quotas' ? (
                    quotas.length > 0 ? (
                        quotas.map(quota => (
                            <ResourceCard key={`${quota.namespace}-${quota.name}`} resource={quota} type="quota" />
                        ))
                    ) : (
                        <div className="col-span-full flex flex-col items-center justify-center py-20 text-gray-500">
                            <Activity size={48} className="mb-4 opacity-20" />
                            <p className="text-lg">No resource quotas found</p>
                            <p className="text-sm opacity-60">Create one to get started</p>
                        </div>
                    )
                ) : (
                    limitRanges.length > 0 ? (
                        limitRanges.map(range => (
                            <ResourceCard key={`${range.namespace}-${range.name}`} resource={range} type="limit" />
                        ))
                    ) : (
                        <div className="col-span-full flex flex-col items-center justify-center py-20 text-gray-500">
                            <Tag size={48} className="mb-4 opacity-20" />
                            <p className="text-lg">No limit ranges found</p>
                            <p className="text-sm opacity-60">Create one to get started</p>
                        </div>
                    )
                )}
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
                        setActiveTab('quotas'); // Ensure we're on the quotas tab
                        fetchQuotas();
                    }}
                />
            )}

            {editingLimitRange && (
                <LimitRangeEditor
                    resource={editingLimitRange}
                    onClose={() => setEditingLimitRange(null)}
                    onSaved={() => {
                        setEditingLimitRange(null);
                        setActiveTab('limits'); // Ensure we're on the limits tab
                        fetchLimitRanges();
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
                            fetchQuotas();
                        } else {
                            fetchLimitRanges();
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
