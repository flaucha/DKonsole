import React, { useEffect, useState } from 'react';
import { Database, RefreshCw, ChevronDown, ChevronRight, Tag, Calendar, Activity, Edit, Plus } from 'lucide-react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import QuotaEditor from './QuotaEditor';
import LimitRangeEditor from './LimitRangeEditor';

const NamespaceManager = () => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [namespaces, setNamespaces] = useState([]);
    const [loading, setLoading] = useState(false);
    const [expandedNs, setExpandedNs] = useState({});
    const [quotas, setQuotas] = useState({});
    const [limitRanges, setLimitRanges] = useState({});
    const [editingQuota, setEditingQuota] = useState(null);
    const [editingLimitRange, setEditingLimitRange] = useState(null);

    const fetchNamespaces = () => {
        setLoading(true);
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        authFetch(`/api/namespaces?${params.toString()}`)
            .then(res => res.json())
            .then(data => {
                setNamespaces(data || []);
                setLoading(false);
            })
            .catch(() => setLoading(false));
    };

    const fetchNamespaceDetails = (nsName) => {
        const params = new URLSearchParams({
            namespace: nsName,
            kind: 'ResourceQuota'
        });
        if (currentCluster) params.append('cluster', currentCluster);

        authFetch(`/api/resources?${params.toString()}`)
            .then(res => res.json())
            .then(data => {
                setQuotas(prev => ({ ...prev, [nsName]: data || [] }));
            })
            .catch(() => setQuotas(prev => ({ ...prev, [nsName]: [] })));

        const limitParams = new URLSearchParams({
            namespace: nsName,
            kind: 'LimitRange'
        });
        if (currentCluster) limitParams.append('cluster', currentCluster);

        authFetch(`/api/resources?${limitParams.toString()}`)
            .then(res => res.json())
            .then(data => {
                setLimitRanges(prev => ({ ...prev, [nsName]: data || [] }));
            })
            .catch(() => setLimitRanges(prev => ({ ...prev, [nsName]: [] })));
    };

    useEffect(() => {
        fetchNamespaces();
    }, [currentCluster]);

    const toggleExpand = (nsName) => {
        const isExpanding = !expandedNs[nsName];
        setExpandedNs(prev => ({ ...prev, [nsName]: isExpanding }));

        if (isExpanding && !quotas[nsName]) {
            fetchNamespaceDetails(nsName);
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

    return (
        <div className="p-6">
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center space-x-2">
                    <Database className="text-blue-400" size={20} />
                    <h1 className="text-2xl font-bold text-white">Namespace Manager</h1>
                    {loading && <RefreshCw size={16} className="animate-spin text-gray-400" />}
                </div>
                <button
                    onClick={fetchNamespaces}
                    className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-white rounded-md border border-gray-700 text-sm transition-colors flex items-center"
                >
                    <RefreshCw size={14} className="mr-2" />
                    Refresh
                </button>
            </div>

            <div className="bg-gray-800 border border-gray-700 rounded-lg overflow-hidden">
                <table className="min-w-full">
                    <thead className="bg-gray-900">
                        <tr>
                            <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider w-8"></th>
                            <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Name</th>
                            <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Status</th>
                            <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Age</th>
                            <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Resource Quotas</th>
                            <th className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Limit Ranges</th>
                        </tr>
                    </thead>
                    <tbody className="bg-gray-800 divide-y divide-gray-700">
                        {namespaces.map((ns) => (
                            <React.Fragment key={ns.name}>
                                <tr className="hover:bg-gray-750 transition-colors">
                                    <td className="px-4 py-3">
                                        <button
                                            onClick={() => toggleExpand(ns.name)}
                                            className="text-gray-400 hover:text-white transition-colors"
                                        >
                                            {expandedNs[ns.name] ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                                        </button>
                                    </td>
                                    <td className="px-4 py-3 text-sm font-medium text-white">{ns.name}</td>
                                    <td className="px-4 py-3">
                                        <span className="px-2 py-1 text-xs rounded-full bg-green-900/30 text-green-400 border border-green-800">
                                            Active
                                        </span>
                                    </td>
                                    <td className="px-4 py-3 text-sm text-gray-400">{getAge(ns.created)}</td>
                                    <td className="px-4 py-3 text-sm text-gray-300">
                                        {quotas[ns.name] ? quotas[ns.name].length : '-'}
                                    </td>
                                    <td className="px-4 py-3 text-sm text-gray-300">
                                        {limitRanges[ns.name] ? limitRanges[ns.name].length : '-'}
                                    </td>
                                </tr>
                                {expandedNs[ns.name] && (
                                    <tr>
                                        <td colSpan="6" className="bg-gray-900/50 px-4 py-4">
                                            <div className="space-y-4">
                                                {/* Resource Quotas Section */}
                                                <div>
                                                    <div className="flex items-center justify-between mb-2">
                                                        <h3 className="text-sm font-semibold text-gray-300 flex items-center">
                                                            <Activity size={14} className="mr-2 text-gray-400" />
                                                            Resource Quotas
                                                        </h3>
                                                        <button
                                                            onClick={() => setEditingQuota({ namespace: ns.name, name: '', kind: 'ResourceQuota', isNew: true })}
                                                            className="flex items-center px-2 py-1 text-xs bg-blue-600 hover:bg-blue-700 text-white rounded border border-blue-500 transition-colors"
                                                        >
                                                            <Plus size={12} className="mr-1" />
                                                            Add Quota
                                                        </button>
                                                    </div>
                                                    {quotas[ns.name]?.length > 0 ? (
                                                        <div className="space-y-2">
                                                            {quotas[ns.name].map((quota) => (
                                                                <div key={quota.name} className="bg-gray-800 border border-gray-700 rounded p-3">
                                                                    <div className="flex items-center justify-between mb-2">
                                                                        <span className="text-sm font-medium text-white">{quota.name}</span>
                                                                        <button
                                                                            onClick={() => setEditingQuota({ ...quota, namespace: ns.name })}
                                                                            className="flex items-center px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-200 rounded border border-gray-600 transition-colors"
                                                                        >
                                                                            <Edit size={12} className="mr-1" />
                                                                            Edit
                                                                        </button>
                                                                    </div>
                                                                    {quota.details?.hard && (
                                                                        <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-xs">
                                                                            {Object.entries(quota.details.hard).map(([key, value]) => (
                                                                                <div key={key} className="text-gray-400">
                                                                                    <span className="font-medium">{key}:</span> {value}
                                                                                    {quota.details.used?.[key] && (
                                                                                        <span className="text-gray-500"> ({quota.details.used[key]} used)</span>
                                                                                    )}
                                                                                </div>
                                                                            ))}
                                                                        </div>
                                                                    )}
                                                                </div>
                                                            ))}
                                                        </div>
                                                    ) : (
                                                        <p className="text-xs text-gray-500 italic">No resource quotas defined</p>
                                                    )}
                                                </div>

                                                {/* Limit Ranges Section */}
                                                <div>
                                                    <div className="flex items-center justify-between mb-2">
                                                        <h3 className="text-sm font-semibold text-gray-300 flex items-center">
                                                            <Tag size={14} className="mr-2 text-gray-400" />
                                                            Limit Ranges
                                                        </h3>
                                                        <button
                                                            onClick={() => setEditingLimitRange({ namespace: ns.name, name: '', kind: 'LimitRange', isNew: true })}
                                                            className="flex items-center px-2 py-1 text-xs bg-blue-600 hover:bg-blue-700 text-white rounded border border-blue-500 transition-colors"
                                                        >
                                                            <Plus size={12} className="mr-1" />
                                                            Add Limit Range
                                                        </button>
                                                    </div>
                                                    {limitRanges[ns.name]?.length > 0 ? (
                                                        <div className="space-y-2">
                                                            {limitRanges[ns.name].map((limitRange) => (
                                                                <div key={limitRange.name} className="bg-gray-800 border border-gray-700 rounded p-3">
                                                                    <div className="flex items-center justify-between mb-2">
                                                                        <span className="text-sm font-medium text-white">{limitRange.name}</span>
                                                                        <button
                                                                            onClick={() => setEditingLimitRange({ ...limitRange, namespace: ns.name })}
                                                                            className="flex items-center px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-200 rounded border border-gray-600 transition-colors"
                                                                        >
                                                                            <Edit size={12} className="mr-1" />
                                                                            Edit
                                                                        </button>
                                                                    </div>
                                                                    {limitRange.details?.limits && (
                                                                        <div className="text-xs text-gray-400">
                                                                            {limitRange.details.limits.length} limit(s) configured
                                                                        </div>
                                                                    )}
                                                                </div>
                                                            ))}
                                                        </div>
                                                    ) : (
                                                        <p className="text-xs text-gray-500 italic">No limit ranges defined</p>
                                                    )}
                                                </div>
                                            </div>
                                        </td>
                                    </tr>
                                )}
                            </React.Fragment>
                        ))}
                    </tbody>
                </table>

                {namespaces.length === 0 && !loading && (
                    <div className="p-6 text-center text-gray-500">No namespaces found</div>
                )}
            </div>

            {editingQuota && (
                <QuotaEditor
                    resource={editingQuota}
                    onClose={() => setEditingQuota(null)}
                    onSaved={() => {
                        setEditingQuota(null);
                        fetchNamespaceDetails(editingQuota.namespace);
                    }}
                />
            )}

            {editingLimitRange && (
                <LimitRangeEditor
                    resource={editingLimitRange}
                    onClose={() => setEditingLimitRange(null)}
                    onSaved={() => {
                        setEditingLimitRange(null);
                        fetchNamespaceDetails(editingLimitRange.namespace);
                    }}
                />
            )}
        </div>
    );
};

export default NamespaceManager;
