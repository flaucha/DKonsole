import React, { useEffect, useState } from 'react';
import { Activity, RefreshCw, Tag, Plus, CircleMinus, CirclePlus, MoreVertical, FileText } from 'lucide-react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import QuotaEditor from './QuotaEditor';
import LimitRangeEditor from './LimitRangeEditor';
import YamlEditor from './YamlEditor';
import { getStatusBadgeClass } from '../utils/statusBadge';

const ResourceQuotaManager = () => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [quotas, setQuotas] = useState([]);
    const [limitRanges, setLimitRanges] = useState([]);
    const [loading, setLoading] = useState(false);
    const [expandedId, setExpandedId] = useState(null);
    const [activeTab, setActiveTab] = useState('quotas'); // 'quotas' or 'limits'
    const [editingQuota, setEditingQuota] = useState(null);
    const [editingLimitRange, setEditingLimitRange] = useState(null);
    const [editingYaml, setEditingYaml] = useState(null);
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [namespaces, setNamespaces] = useState([]);

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
        const params = new URLSearchParams({ kind: 'ResourceQuota' });
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
        const params = new URLSearchParams({ kind: 'LimitRange' });
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

    useEffect(() => {
        fetchAll();
    }, [currentCluster, activeTab]);

    const toggleExpand = (id) => {
        setExpandedId(expandedId === id ? null : id);
    };

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
                throw new Error('Failed to delete resource');
            }

            if (kind === 'ResourceQuota') {
                fetchQuotas();
            } else {
                fetchLimitRanges();
            }
        } catch (err) {
            alert(`Error deleting ${kind}: ${err.message}`);
        }
    };

    const EditYamlButton = ({ onClick }) => (
        <button
            onClick={onClick}
            className="flex items-center px-2.5 py-1 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-md border border-gray-600 text-xs transition-colors"
        >
            <FileText size={12} className="mr-1.5" />
            Edit YAML
        </button>
    );

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

    const sortedQuotas = [...quotas].sort((a, b) => {
        if (a.namespace !== b.namespace) {
            return (a.namespace || '').localeCompare(b.namespace || '');
        }
        return (a.name || '').localeCompare(b.name || '');
    });

    const sortedLimitRanges = [...limitRanges].sort((a, b) => {
        if (a.namespace !== b.namespace) {
            return (a.namespace || '').localeCompare(b.namespace || '');
        }
        return (a.name || '').localeCompare(b.name || '');
    });

    return (
        <div className="p-6">
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center space-x-2">
                    <Activity className="text-blue-400" size={20} />
                    <h1 className="text-2xl font-bold text-white">Resource Quotas & Limits</h1>
                    {loading && <RefreshCw size={16} className="animate-spin text-gray-400" />}
                </div>
                <button
                    onClick={fetchAll}
                    className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-white rounded-md border border-gray-700 text-sm transition-colors flex items-center"
                >
                    <RefreshCw size={14} className="mr-2" />
                    Refresh
                </button>
            </div>

            {/* Tabs */}
            <div className="flex space-x-2 mb-4 border-b border-gray-700">
                <button
                    onClick={() => {
                        setActiveTab('quotas');
                        setExpandedId(null);
                    }}
                    className={`px-4 py-2 text-sm font-medium transition-colors ${
                        activeTab === 'quotas'
                            ? 'text-blue-400 border-b-2 border-blue-400'
                            : 'text-gray-400 hover:text-gray-300'
                    }`}
                >
                    Resource Quotas ({quotas.length})
                </button>
                <button
                    onClick={() => {
                        setActiveTab('limits');
                        setExpandedId(null);
                    }}
                    className={`px-4 py-2 text-sm font-medium transition-colors ${
                        activeTab === 'limits'
                            ? 'text-blue-400 border-b-2 border-blue-400'
                            : 'text-gray-400 hover:text-gray-300'
                    }`}
                >
                    Limit Ranges ({limitRanges.length})
                </button>
            </div>

            {/* Resource Quotas Tab */}
            {activeTab === 'quotas' && (
                <div className="bg-gray-800 border border-gray-700 rounded-lg overflow-hidden">
                    <div className="flex items-center justify-between px-4 py-3 bg-gray-900 border-b border-gray-700">
                        <h2 className="text-sm font-semibold text-gray-300">Resource Quotas</h2>
                        <select
                            onChange={(e) => {
                                if (e.target.value) {
                                    setEditingQuota({ namespace: e.target.value, name: '', kind: 'ResourceQuota', isNew: true });
                                }
                            }}
                            className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md border border-blue-500 text-sm transition-colors"
                            defaultValue=""
                        >
                            <option value="">Create New Quota...</option>
                            {namespaces.map(ns => (
                                <option key={ns.name} value={ns.name}>{ns.name}</option>
                            ))}
                        </select>
                    </div>
                    <table className="min-w-full">
                        <thead className="bg-gray-900">
                            <tr>
                                <th className="w-10 px-4 py-3 bg-gray-900 rounded-tl-lg border-b border-gray-700"></th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Name</th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Namespace</th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Status</th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Age</th>
                                <th className="w-10 px-4 py-3 bg-gray-900 rounded-tr-lg border-b border-gray-700"></th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-800">
                            {sortedQuotas.map((quota) => (
                                <React.Fragment key={`${quota.namespace}-${quota.name}`}>
                                    <tr
                                        onClick={() => toggleExpand(`quota-${quota.namespace}-${quota.name}`)}
                                        className={`group hover:bg-gray-800/50 transition-colors cursor-pointer ${expandedId === `quota-${quota.namespace}-${quota.name}` ? 'bg-gray-800/30' : ''}`}
                                    >
                                        <td className="px-4 py-3 whitespace-nowrap text-gray-400 text-center">
                                            {expandedId === `quota-${quota.namespace}-${quota.name}` ? <CircleMinus size={16} /> : <CirclePlus size={16} />}
                                        </td>
                                        <td className="px-6 py-3 whitespace-nowrap">
                                            <div className="flex items-center">
                                                <div className="flex-shrink-0 h-6 w-6 bg-gray-700 rounded flex items-center justify-center text-gray-400">
                                                    <Activity size={14} />
                                                </div>
                                                <div className="ml-4">
                                                    <div className="text-sm font-medium text-white">{quota.name}</div>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-3 whitespace-nowrap text-sm text-gray-300">{quota.namespace || '—'}</td>
                                        <td className="px-6 py-3 whitespace-nowrap">
                                            <span className={`px-2 py-1 text-xs rounded-full ${getStatusBadgeClass(quota.status)}`}>
                                                {quota.status || 'Active'}
                                            </span>
                                        </td>
                                        <td className="px-6 py-3 whitespace-nowrap text-sm text-gray-400">{getAge(quota.created)}</td>
                                        <td className="px-4 py-3 whitespace-nowrap text-gray-300" onClick={(e) => e.stopPropagation()}>
                                            <div className="relative flex items-center justify-end">
                                                <button
                                                    onClick={() => setMenuOpen(menuOpen === `quota-${quota.namespace}-${quota.name}` ? null : `quota-${quota.namespace}-${quota.name}`)}
                                                    className="p-1 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                                                >
                                                    <MoreVertical size={16} />
                                                </button>
                                                {menuOpen === `quota-${quota.namespace}-${quota.name}` && (
                                                    <div className="absolute right-0 mt-1 w-36 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
                                                        <div className="flex flex-col">
                                                            <button
                                                                onClick={() => {
                                                                    setEditingQuota({ ...quota, namespace: quota.namespace });
                                                                    setMenuOpen(null);
                                                                }}
                                                                className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                            >
                                                                Edit
                                                            </button>
                                                            <button
                                                                onClick={() => {
                                                                    setEditingYaml({ ...quota, namespace: quota.namespace, kind: 'ResourceQuota' });
                                                                    setMenuOpen(null);
                                                                }}
                                                                className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                            >
                                                                Edit YAML
                                                            </button>
                                                            <button
                                                                onClick={() => {
                                                                    setConfirmAction({ resource: quota, kind: 'ResourceQuota', force: false });
                                                                    setMenuOpen(null);
                                                                }}
                                                                className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                            >
                                                                Delete
                                                            </button>
                                                            <button
                                                                onClick={() => {
                                                                    setConfirmAction({ resource: quota, kind: 'ResourceQuota', force: true });
                                                                    setMenuOpen(null);
                                                                }}
                                                                className="w-full text-left px-4 py-2 text-sm text-red-300 hover:bg-red-900/40"
                                                            >
                                                                Force Delete
                                                            </button>
                                                        </div>
                                                    </div>
                                                )}
                                            </div>
                                        </td>
                                    </tr>
                                    <tr>
                                        <td colSpan={6} className={`px-6 pt-0 bg-gray-800 border-0 ${expandedId === `quota-${quota.namespace}-${quota.name}` ? 'border-b border-gray-700' : ''}`}>
                                            <div
                                                className={`pl-12 transition-all duration-300 ease-in-out ${expandedId === `quota-${quota.namespace}-${quota.name}` ? 'opacity-100 pb-4' : 'max-h-0 opacity-0 overflow-hidden'}`}
                                            >
                                                {expandedId === `quota-${quota.namespace}-${quota.name}` && (
                                                    <div className="p-4 bg-gray-900/50 rounded-md space-y-4">
                                                        {quota.details?.hard && (
                                                            <div>
                                                                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Hard Limits</h4>
                                                                <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                                                                    {Object.entries(quota.details.hard).map(([key, value]) => (
                                                                        <div key={key} className="bg-gray-800 border border-gray-700 rounded p-2 text-xs">
                                                                            <div className="font-medium text-gray-400 mb-1">{key}</div>
                                                                            <div className="text-gray-300">{value}</div>
                                                                            {quota.details.used?.[key] && (
                                                                                <div className="text-gray-500 mt-1">
                                                                                    Used: {quota.details.used[key]}
                                                                                </div>
                                                                            )}
                                                                        </div>
                                                                    ))}
                                                                </div>
                                                            </div>
                                                        )}
                                                        {(!quota.details?.hard || Object.keys(quota.details.hard).length === 0) && (
                                                            <p className="text-sm text-gray-500 italic">No limits configured</p>
                                                        )}
                                                        <div className="flex justify-end mt-4">
                                                            <EditYamlButton onClick={() => setEditingYaml({ ...quota, namespace: quota.namespace, kind: 'ResourceQuota' })} />
                                                        </div>
                                                    </div>
                                                )}
                                            </div>
                                        </td>
                                    </tr>
                                </React.Fragment>
                            ))}
                        </tbody>
                    </table>
                    {sortedQuotas.length === 0 && !loading && (
                        <div className="p-6 text-center text-gray-500">No resource quotas found</div>
                    )}
                </div>
            )}

            {/* Limit Ranges Tab */}
            {activeTab === 'limits' && (
                <div className="bg-gray-800 border border-gray-700 rounded-lg overflow-hidden">
                    <div className="flex items-center justify-between px-4 py-3 bg-gray-900 border-b border-gray-700">
                        <h2 className="text-sm font-semibold text-gray-300">Limit Ranges</h2>
                        <select
                            onChange={(e) => {
                                if (e.target.value) {
                                    setEditingLimitRange({ namespace: e.target.value, name: '', kind: 'LimitRange', isNew: true });
                                }
                            }}
                            className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md border border-blue-500 text-sm transition-colors"
                            defaultValue=""
                        >
                            <option value="">Create New Limit Range...</option>
                            {namespaces.map(ns => (
                                <option key={ns.name} value={ns.name}>{ns.name}</option>
                            ))}
                        </select>
                    </div>
                    <table className="min-w-full">
                        <thead className="bg-gray-900">
                            <tr>
                                <th className="w-10 px-4 py-3 bg-gray-900 rounded-tl-lg border-b border-gray-700"></th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Name</th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Namespace</th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Status</th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Age</th>
                                <th className="w-10 px-4 py-3 bg-gray-900 rounded-tr-lg border-b border-gray-700"></th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-800">
                            {sortedLimitRanges.map((limitRange) => (
                                <React.Fragment key={`${limitRange.namespace}-${limitRange.name}`}>
                                    <tr
                                        onClick={() => toggleExpand(`limit-${limitRange.namespace}-${limitRange.name}`)}
                                        className={`group hover:bg-gray-800/50 transition-colors cursor-pointer ${expandedId === `limit-${limitRange.namespace}-${limitRange.name}` ? 'bg-gray-800/30' : ''}`}
                                    >
                                        <td className="px-4 py-3 whitespace-nowrap text-gray-400 text-center">
                                            {expandedId === `limit-${limitRange.namespace}-${limitRange.name}` ? <CircleMinus size={16} /> : <CirclePlus size={16} />}
                                        </td>
                                        <td className="px-6 py-3 whitespace-nowrap">
                                            <div className="flex items-center">
                                                <div className="flex-shrink-0 h-6 w-6 bg-gray-700 rounded flex items-center justify-center text-gray-400">
                                                    <Tag size={14} />
                                                </div>
                                                <div className="ml-4">
                                                    <div className="text-sm font-medium text-white">{limitRange.name}</div>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-3 whitespace-nowrap text-sm text-gray-300">{limitRange.namespace || '—'}</td>
                                        <td className="px-6 py-3 whitespace-nowrap">
                                            <span className={`px-2 py-1 text-xs rounded-full ${getStatusBadgeClass(limitRange.status)}`}>
                                                {limitRange.status || 'Active'}
                                            </span>
                                        </td>
                                        <td className="px-6 py-3 whitespace-nowrap text-sm text-gray-400">{getAge(limitRange.created)}</td>
                                        <td className="px-4 py-3 whitespace-nowrap text-gray-300" onClick={(e) => e.stopPropagation()}>
                                            <div className="relative flex items-center justify-end">
                                                <button
                                                    onClick={() => setMenuOpen(menuOpen === `limit-${limitRange.namespace}-${limitRange.name}` ? null : `limit-${limitRange.namespace}-${limitRange.name}`)}
                                                    className="p-1 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                                                >
                                                    <MoreVertical size={16} />
                                                </button>
                                                {menuOpen === `limit-${limitRange.namespace}-${limitRange.name}` && (
                                                    <div className="absolute right-0 mt-1 w-36 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
                                                        <div className="flex flex-col">
                                                            <button
                                                                onClick={() => {
                                                                    setEditingLimitRange({ ...limitRange, namespace: limitRange.namespace });
                                                                    setMenuOpen(null);
                                                                }}
                                                                className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                            >
                                                                Edit
                                                            </button>
                                                            <button
                                                                onClick={() => {
                                                                    setEditingYaml({ ...limitRange, namespace: limitRange.namespace, kind: 'LimitRange' });
                                                                    setMenuOpen(null);
                                                                }}
                                                                className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                            >
                                                                Edit YAML
                                                            </button>
                                                            <button
                                                                onClick={() => {
                                                                    setConfirmAction({ resource: limitRange, kind: 'LimitRange', force: false });
                                                                    setMenuOpen(null);
                                                                }}
                                                                className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                            >
                                                                Delete
                                                            </button>
                                                            <button
                                                                onClick={() => {
                                                                    setConfirmAction({ resource: limitRange, kind: 'LimitRange', force: true });
                                                                    setMenuOpen(null);
                                                                }}
                                                                className="w-full text-left px-4 py-2 text-sm text-red-300 hover:bg-red-900/40"
                                                            >
                                                                Force Delete
                                                            </button>
                                                        </div>
                                                    </div>
                                                )}
                                            </div>
                                        </td>
                                    </tr>
                                    <tr>
                                        <td colSpan={6} className={`px-6 pt-0 bg-gray-800 border-0 ${expandedId === `limit-${limitRange.namespace}-${limitRange.name}` ? 'border-b border-gray-700' : ''}`}>
                                            <div
                                                className={`pl-12 transition-all duration-300 ease-in-out ${expandedId === `limit-${limitRange.namespace}-${limitRange.name}` ? 'opacity-100 pb-4' : 'max-h-0 opacity-0 overflow-hidden'}`}
                                            >
                                                {expandedId === `limit-${limitRange.namespace}-${limitRange.name}` && (
                                                    <div className="p-4 bg-gray-900/50 rounded-md space-y-4">
                                                        {limitRange.details?.limits && limitRange.details.limits.length > 0 ? (
                                                            <div>
                                                                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Limits</h4>
                                                                <div className="space-y-3">
                                                                    {limitRange.details.limits.map((limit, idx) => (
                                                                        <div key={idx} className="bg-gray-800 border border-gray-700 rounded p-3">
                                                                            <div className="text-xs font-medium text-gray-400 mb-2">Type: {limit.type || 'Container'}</div>
                                                                            <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-xs">
                                                                                {limit.max && Object.entries(limit.max).map(([key, value]) => (
                                                                                    <div key={`max-${key}`} className="text-gray-400">
                                                                                        <span className="font-medium">Max {key}:</span> {value}
                                                                                    </div>
                                                                                ))}
                                                                                {limit.min && Object.entries(limit.min).map(([key, value]) => (
                                                                                    <div key={`min-${key}`} className="text-gray-400">
                                                                                        <span className="font-medium">Min {key}:</span> {value}
                                                                                    </div>
                                                                                ))}
                                                                                {limit.default && Object.entries(limit.default).map(([key, value]) => (
                                                                                    <div key={`default-${key}`} className="text-gray-400">
                                                                                        <span className="font-medium">Default {key}:</span> {value}
                                                                                    </div>
                                                                                ))}
                                                                                {limit.defaultRequest && Object.entries(limit.defaultRequest).map(([key, value]) => (
                                                                                    <div key={`defaultRequest-${key}`} className="text-gray-400">
                                                                                        <span className="font-medium">Default Request {key}:</span> {value}
                                                                                    </div>
                                                                                ))}
                                                                            </div>
                                                                        </div>
                                                                    ))}
                                                                </div>
                                                            </div>
                                                        ) : (
                                                            <p className="text-sm text-gray-500 italic">No limits configured</p>
                                                        )}
                                                        <div className="flex justify-end mt-4">
                                                            <EditYamlButton onClick={() => setEditingYaml({ ...limitRange, namespace: limitRange.namespace, kind: 'LimitRange' })} />
                                                        </div>
                                                    </div>
                                                )}
                                            </div>
                                        </td>
                                    </tr>
                                </React.Fragment>
                            ))}
                        </tbody>
                    </table>
                    {sortedLimitRanges.length === 0 && !loading && (
                        <div className="p-6 text-center text-gray-500">No limit ranges found</div>
                    )}
                </div>
            )}

            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(null)}
                ></div>
            )}

            {editingQuota && (
                <QuotaEditor
                    resource={editingQuota}
                    onClose={() => setEditingQuota(null)}
                    onSaved={() => {
                        setEditingQuota(null);
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
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                        <h3 className="text-lg font-semibold text-white mb-2">
                            Confirm delete
                        </h3>
                        <p className="text-sm text-gray-300 mb-4">
                            {confirmAction.force ? 'Force delete' : 'Delete'} {confirmAction.kind} "{confirmAction.resource.name}"?
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
                                    await handleDelete(confirmAction.resource, confirmAction.kind, confirmAction.force);
                                    setConfirmAction(null);
                                }}
                                className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors"
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

