import React, { useEffect, useState } from 'react';
import { Database, RefreshCw, CirclePlus, CircleMinus, Tag, Calendar, Clock, MoreVertical, FileText } from 'lucide-react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import YamlEditor from './YamlEditor';
import { getStatusBadgeClass } from '../utils/statusBadge';

const NamespaceManager = () => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [namespaces, setNamespaces] = useState([]);
    const [loading, setLoading] = useState(false);
    const [expandedNs, setExpandedNs] = useState({});
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [editingYaml, setEditingYaml] = useState(null);

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

    useEffect(() => {
        fetchNamespaces();
    }, [currentCluster]);

    const toggleExpand = (nsName) => {
        setExpandedNs(prev => ({ ...prev, [nsName]: !prev[nsName] }));
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

    const formatDate = (dateString) => {
        if (!dateString) return 'Unknown';
        return new Date(dateString).toLocaleString();
    };

    const handleDelete = async (namespace, force = false) => {
        const params = new URLSearchParams({
            kind: 'Namespace',
            name: namespace
        });
        if (force) params.append('force', 'true');
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const res = await authFetch(`/api/resource?${params.toString()}`, {
                method: 'DELETE'
            });

            if (!res.ok) {
                throw new Error('Failed to delete namespace');
            }

            fetchNamespaces();
        } catch (err) {
            alert(`Error deleting namespace: ${err.message}`);
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
                <table className="min-w-full border-separate border-spacing-0">
                    <thead>
                        <tr>
                            <th className="w-10 px-4 py-3 bg-gray-900 rounded-tl-lg border-b border-gray-700"></th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider bg-gray-900 border-b border-gray-700">Name</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider bg-gray-900 border-b border-gray-700">Status</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider bg-gray-900 border-b border-gray-700">Labels</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider bg-gray-900 border-b border-gray-700">Age</th>
                            <th className="w-10 px-4 py-3 bg-gray-900 rounded-tr-lg border-b border-gray-700"></th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-800">
                        {namespaces.map((ns) => (
                            <React.Fragment key={ns.name}>
                                <tr
                                    className={`group hover:bg-gray-800/50 transition-colors cursor-pointer ${expandedNs[ns.name] ? 'bg-gray-800/30' : ''}`}
                                    onClick={() => toggleExpand(ns.name)}
                                >
                                    <td className="px-4 py-3 whitespace-nowrap text-gray-400 text-center">
                                        {expandedNs[ns.name] ? <CircleMinus size={16} /> : <CirclePlus size={16} />}
                                    </td>
                                    <td className="px-6 py-3">
                                        <div className="flex items-center min-w-0">
                                            <div className="flex-shrink-0 h-6 w-6 bg-gray-700 rounded flex items-center justify-center text-gray-400">
                                                <Database size={14} />
                                            </div>
                                            <div className="ml-4 min-w-0 flex-1">
                                                <div className="text-sm font-medium text-white truncate">{ns.name}</div>
                                            </div>
                                        </div>
                                    </td>
                                    <td className="px-6 py-3 whitespace-nowrap">
                                        <span className={`px-2 py-1 text-xs rounded-full ${getStatusBadgeClass(ns.status)}`}>
                                            {ns.status || 'Unknown'}
                                        </span>
                                    </td>
                                    <td className="px-6 py-3 whitespace-nowrap text-sm text-gray-400">
                                        {ns.labels ? Object.keys(ns.labels).length : 0}
                                    </td>
                                    <td className="px-6 py-3 whitespace-nowrap text-sm text-gray-400">{getAge(ns.created)}</td>
                                    <td className="px-4 py-3 whitespace-nowrap text-gray-300" onClick={(e) => e.stopPropagation()}>
                                        <div className="relative flex items-center justify-end">
                                            <button
                                                onClick={() => setMenuOpen(menuOpen === ns.name ? null : ns.name)}
                                                className="p-1 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                                            >
                                                <MoreVertical size={16} />
                                            </button>
                                            {menuOpen === ns.name && (
                                                <div className="absolute right-0 mt-1 w-36 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
                                                    <div className="flex flex-col">
                                                        <button
                                                            onClick={() => {
                                                                setEditingYaml({ name: ns.name, kind: 'Namespace', namespaced: false });
                                                                setMenuOpen(null);
                                                            }}
                                                            className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                        >
                                                            Edit YAML
                                                        </button>
                                                        <button
                                                            onClick={() => {
                                                                setConfirmAction({ namespace: ns.name, force: false });
                                                                setMenuOpen(null);
                                                            }}
                                                            className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                        >
                                                            Delete
                                                        </button>
                                                        <button
                                                            onClick={() => {
                                                                setConfirmAction({ namespace: ns.name, force: true });
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
                                    <td colSpan="6" className={`px-6 pt-0 bg-gray-800 border-0 ${expandedNs[ns.name] ? 'border-b border-gray-700' : ''}`}>
                                        <div
                                            className={`transition-all duration-300 ease-in-out ${expandedNs[ns.name] ? 'opacity-100 pb-4' : 'max-h-0 opacity-0 overflow-hidden'}`}
                                        >
                                            {expandedNs[ns.name] && (
                                                <div className="p-4 bg-gray-900/50 rounded-md space-y-6">
                                                {/* Basic Information */}
                                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                    <div>
                                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                                            <Clock size={12} className="mr-1" />
                                                            Creation Time
                                                        </h4>
                                                        <div className="text-sm text-gray-300">{formatDate(ns.created)}</div>
                                                    </div>
                                                    <div>
                                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                                            <Tag size={12} className="mr-1" />
                                                            Age
                                                        </h4>
                                                        <div className="text-sm text-gray-300">{getAge(ns.created)}</div>
                                                    </div>
                                                </div>

                                                {/* Labels Section */}
                                                <div>
                                                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                                        <Tag size={12} className="mr-1" />
                                                        Labels
                                                    </h4>
                                                    <div className="flex flex-wrap gap-2">
                                                        {ns.labels && Object.keys(ns.labels).length > 0 ? (
                                                            Object.entries(ns.labels).map(([k, v]) => (
                                                                <span key={k} className="px-2 py-1 bg-gray-800 border border-gray-700 rounded text-xs text-gray-300">
                                                                    {k}={v}
                                                                </span>
                                                            ))
                                                        ) : (
                                                            <span className="text-sm text-gray-500 italic">No labels</span>
                                                        )}
                                                    </div>
                                                </div>

                                                {/* Annotations Section */}
                                                {ns.annotations && Object.keys(ns.annotations).length > 0 && (
                                                    <div>
                                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Annotations</h4>
                                                        <div className="space-y-1">
                                                            {Object.entries(ns.annotations).map(([k, v]) => (
                                                                <div key={k} className="bg-gray-800 border border-gray-700 rounded p-2 text-xs">
                                                                    <span className="font-medium text-gray-400">{k}:</span>
                                                                    <span className="ml-2 text-gray-300 break-words">{v}</span>
                                                                </div>
                                                            ))}
                                                        </div>
                                                    </div>
                                                )}

                                                {/* Actions */}
                                                <div className="flex justify-end mt-4">
                                                    <EditYamlButton onClick={() => setEditingYaml({ name: ns.name, kind: 'Namespace', namespaced: false })} />
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

                {namespaces.length === 0 && !loading && (
                    <div className="p-6 text-center text-gray-500">No namespaces found</div>
                )}
            </div>

            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(null)}
                ></div>
            )}

            {editingYaml && (
                <YamlEditor
                    resource={editingYaml}
                    onClose={() => setEditingYaml(null)}
                    onSaved={() => {
                        setEditingYaml(null);
                        fetchNamespaces();
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
                            {confirmAction.force ? 'Force delete' : 'Delete'} Namespace "{confirmAction.namespace}"?
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
                                    await handleDelete(confirmAction.namespace, confirmAction.force);
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

export default NamespaceManager;
