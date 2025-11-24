import React, { useState } from 'react';
import { Database, RefreshCw, Tag, Clock, MoreVertical, FileText, ChevronDown, Search } from 'lucide-react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import YamlEditor from './YamlEditor';
import { getStatusBadgeClass } from '../utils/statusBadge';
import { formatDateTime } from '../utils/dateUtils';
import { getExpandableRowClasses, getExpandableCellClasses, getExpandableRowRowClasses } from '../utils/expandableRow';
import { useNamespaces } from '../hooks/useNamespaces';

const NamespaceManager = () => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [expandedId, setExpandedId] = useState(null);
    const [sortField, setSortField] = useState('name');
    const [sortDirection, setSortDirection] = useState('asc');
    const [filter, setFilter] = useState('');
    const [isSearchFocused, setIsSearchFocused] = useState(false);
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [editingYaml, setEditingYaml] = useState(null);

    const { data: namespaces = [], isLoading: loading, refetch } = useNamespaces(authFetch, currentCluster);

    const toggleExpand = (nsName) => {
        setExpandedId(current => current === nsName ? null : nsName);
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

    const filteredNamespaces = namespaces.filter(ns => {
        if (!filter) return true;
        return ns.name.toLowerCase().includes(filter.toLowerCase());
    });

    const sortedNamespaces = [...filteredNamespaces].sort((a, b) => {
        const dir = sortDirection === 'asc' ? 1 : -1;
        const getVal = (item) => {
            switch (sortField) {
                case 'name':
                    return item.name || '';
                case 'status':
                    return item.status || '';
                case 'created':
                    return new Date(item.created).getTime() || 0;
                case 'labels':
                    return item.labels ? Object.keys(item.labels).length : 0;
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

            refetch();
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

    if (loading && namespaces.length === 0) {
        return <div className="text-gray-400 animate-pulse p-6">Loading namespaces...</div>;
    }

    return (
        <div className="flex flex-col h-full">
            {/* Toolbar */}
            <div className="flex items-center justify-between p-4 border-b border-gray-800 bg-gray-900/50">
                <div className="flex items-center space-x-4 flex-1">
                    <div className={`relative transition-all duration-300 ${isSearchFocused ? 'w-96' : 'w-64'}`}>
                        <Search className={`absolute left-3 top-1/2 transform -translate-y-1/2 transition-colors duration-300 ${isSearchFocused ? 'text-blue-400' : 'text-gray-500'}`} size={16} />
                        <input
                            type="text"
                            placeholder="Filter namespaces..."
                            value={filter}
                            onChange={(e) => setFilter(e.target.value)}
                            onFocus={() => setIsSearchFocused(true)}
                            onBlur={() => setIsSearchFocused(false)}
                            className="w-full bg-gray-900 border border-gray-700 text-gray-200 text-sm rounded-md pl-10 pr-4 py-2 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all duration-300"
                        />
                    </div>
                    <span className="text-sm text-gray-500">
                        {filteredNamespaces.length} {filteredNamespaces.length === 1 ? 'item' : 'items'}
                    </span>
                </div>
                <div className="flex items-center space-x-2">
                    <button
                        onClick={() => refetch()}
                        className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-md transition-colors"
                        title="Refresh"
                    >
                        <RefreshCw size={16} />
                    </button>
                </div>
            </div>

            {/* Table Header */}
            <div className="grid grid-cols-12 gap-4 px-6 py-3 border-b border-gray-800 bg-gray-900/50 text-xs font-medium text-gray-500 uppercase tracking-wider">
                <div className="col-span-4 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('name')}>
                    Name {renderSortIndicator('name')}
                </div>
                <div className="col-span-2 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('status')}>
                    Status {renderSortIndicator('status')}
                </div>
                <div className="col-span-2 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('labels')}>
                    Labels {renderSortIndicator('labels')}
                </div>
                <div className="col-span-3 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('created')}>
                    Age {renderSortIndicator('created')}
                </div>
                <div className="col-span-1"></div>
            </div>

            {/* Table Body */}
            <div className="flex-1 overflow-y-auto">
                {sortedNamespaces.map((ns) => {
                    const isExpanded = expandedId === ns.name;
                    return (
                        <div key={ns.name} className="border-b border-gray-800 last:border-0">
                            <div
                                onClick={() => toggleExpand(ns.name)}
                                className={`grid grid-cols-12 gap-4 px-6 py-4 cursor-pointer transition-colors duration-200 items-center ${getExpandableRowRowClasses(isExpanded)}`}
                            >
                                <div className="col-span-4 flex items-center font-medium text-sm text-gray-200">
                                    <ChevronDown
                                        size={16}
                                        className={`mr-2 text-gray-500 transition-transform duration-200 ${isExpanded ? 'transform rotate-180' : ''}`}
                                    />
                                    <Database size={16} className="mr-3 text-gray-500" />
                                    <span className="truncate" title={ns.name}>{ns.name}</span>
                                </div>
                                <div className="col-span-2">
                                    <span className={`px-2 py-1 rounded-full text-xs font-medium border ${getStatusBadgeClass(ns.status || 'Unknown')}`}>
                                        {ns.status || 'Unknown'}
                                    </span>
                                </div>
                                <div className="col-span-2 text-sm text-gray-400">
                                    {ns.labels ? Object.keys(ns.labels).length : 0}
                                </div>
                                <div className="col-span-3 text-sm text-gray-400">
                                    {formatDateTime(ns.created)}
                                </div>
                                <div className="col-span-1 flex justify-end" onClick={(e) => e.stopPropagation()}>
                                    <div className="relative">
                                        <button
                                            onClick={() => setMenuOpen(menuOpen === ns.name ? null : ns.name)}
                                            className="p-1 hover:bg-gray-800 rounded text-gray-400 hover:text-white transition-colors"
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
                                                        <Clock size={12} className="mr-1" />
                                                        Creation Time
                                                    </h4>
                                                    <div className="text-sm text-gray-300">{formatDateTime(ns.created)}</div>
                                                </div>
                                                <div>
                                                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                                        <Tag size={12} className="mr-1" />
                                                        Labels Count
                                                    </h4>
                                                    <div className="text-sm text-gray-300">{ns.labels ? Object.keys(ns.labels).length : 0}</div>
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
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    );
                })}
                {sortedNamespaces.length === 0 && (
                    <div className="p-8 text-center text-gray-500">
                        {filter ? 'No namespaces match your filter.' : 'No namespaces found.'}
                    </div>
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
                        refetch();
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
