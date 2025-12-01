import React, { useState, useMemo } from 'react';
import { Database, RefreshCw, Tag, Clock, MoreVertical, FileText, ChevronDown, Search, X, Plus, GripVertical } from 'lucide-react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import YamlEditor from './YamlEditor';
import { getStatusBadgeClass } from '../utils/statusBadge';
import { formatDateTime, formatDateParts } from '../utils/dateUtils';
import { getExpandableRowClasses, getExpandableRowRowClasses } from '../utils/expandableRow';
import { useNamespaces } from '../hooks/useNamespaces';
import { isAdmin } from '../utils/permissions';
import { useColumnOrder } from '../hooks/useColumnOrder';

const DateStack = ({ value }) => {
    const { date, time } = formatDateParts(value);
    return (
        <div className="flex flex-col items-center leading-tight text-sm text-gray-300">
            <span>{date}</span>
            <span className="text-xs text-gray-500">{time}</span>
        </div>
    );
};

const parseDateValue = (value) => {
    const timestamp = new Date(value).getTime();
    return Number.isFinite(timestamp) ? timestamp : 0;
};

const NamespaceManager = () => {
    const { currentCluster } = useSettings();
    const { authFetch, user } = useAuth();
    const [expandedId, setExpandedId] = useState(null);
    const [sortField, setSortField] = useState('name');
    const [sortDirection, setSortDirection] = useState('asc');
    const [filter, setFilter] = useState('');
    const [isSearchFocused, setIsSearchFocused] = useState(false);
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [editingYaml, setEditingYaml] = useState(null);
    const [draggingColumn, setDraggingColumn] = useState(null);

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

    const dataColumns = useMemo(() => ([
        {
            id: 'name',
            label: 'Name',
            width: 'minmax(220px, 2fr)',
            sortValue: (item) => item.name || '',
            align: 'left',
            renderCell: (item, context = {}) => (
                <div className="flex items-center font-medium text-sm text-gray-200">
                    <ChevronDown
                        size={16}
                        className={`mr-2 text-gray-500 transition-transform duration-200 ${context.isExpanded ? 'transform rotate-180' : ''}`}
                    />
                    <Database size={16} className="mr-3 text-gray-500" />
                    <span className="truncate" title={item.name}>{item.name}</span>
                </div>
            )
        },
        {
            id: 'status',
            label: 'Status',
            width: 'minmax(140px, 1fr)',
            sortValue: (item) => item.status || '',
            align: 'center',
            renderCell: (item) => (
                <span className={`px-2 py-1 rounded-full text-xs font-medium border ${getStatusBadgeClass(item.status || 'Unknown')}`}>
                    {item.status || 'Unknown'}
                </span>
            )
        },
        {
            id: 'labels',
            label: 'Labels',
            width: 'minmax(120px, 0.9fr)',
            sortValue: (item) => item.labels ? Object.keys(item.labels).length : 0,
            align: 'center',
            renderCell: (item) => (
                <span className="text-sm text-gray-400">
                    {item.labels ? Object.keys(item.labels).length : 0}
                </span>
            )
        }
    ]), []);

    const ageColumn = useMemo(() => ({
        id: 'age',
        label: 'Age',
        width: 'minmax(160px, 1fr)',
        sortValue: (item) => parseDateValue(item.created),
        pinned: true,
        align: 'center',
        renderCell: (item) => <DateStack value={item.created} />
    }), []);

    const actionsColumn = {
        id: 'actions',
        label: '',
        width: 'minmax(80px, auto)',
        pinned: true,
        isAction: true,
        renderCell: (item) => (
            <div className="flex justify-end" onClick={(e) => e.stopPropagation()}>
                <div className="relative">
                    <button
                        onClick={() => setMenuOpen(menuOpen === item.name ? null : item.name)}
                        className="p-1 hover:bg-gray-800 rounded text-gray-400 hover:text-white transition-colors"
                    >
                        <MoreVertical size={16} />
                    </button>
                    {menuOpen === item.name && (
                        <div className="absolute right-0 mt-1 w-36 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
                            <div className="flex flex-col">
                                {isAdmin(user) ? (
                                    <>
                                        <button
                                            onClick={() => {
                                                setEditingYaml({ name: item.name, kind: 'Namespace', namespaced: false });
                                                setMenuOpen(null);
                                            }}
                                            className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                        >
                                            Edit YAML
                                        </button>
                                        <button
                                            onClick={() => {
                                                setConfirmAction({ namespace: item.name, force: false });
                                                setMenuOpen(null);
                                            }}
                                            className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                        >
                                            Delete
                                        </button>
                                        <button
                                            onClick={() => {
                                                setConfirmAction({ namespace: item.name, force: true });
                                                setMenuOpen(null);
                                            }}
                                            className="w-full text-left px-4 py-2 text-sm text-red-300 hover:bg-red-900/40"
                                        >
                                            Force Delete
                                        </button>
                                    </>
                                ) : (
                                    <div className="px-4 py-2 text-xs text-gray-500">
                                        View only
                                    </div>
                                )}
                            </div>
                        </div>
                    )}
                </div>
            </div>
        )
    };

    const reorderableColumns = useMemo(
        () => dataColumns.filter((col) => !col.pinned && !col.isAction),
        [dataColumns]
    );

    const { orderedColumns, moveColumn } = useColumnOrder(reorderableColumns, 'namespace-columns', user?.username);

    const sortableColumns = useMemo(
        () => [...dataColumns, ageColumn].filter((col) => typeof col.sortValue === 'function'),
        [dataColumns, ageColumn]
    );

    const sortedNamespaces = useMemo(() => {
        const dir = sortDirection === 'asc' ? 1 : -1;
        const activeColumn = sortableColumns.find((col) => col.id === sortField) || sortableColumns[0];
        if (!activeColumn) return filteredNamespaces;
        return [...filteredNamespaces].sort((a, b) => {
            const va = activeColumn.sortValue(a);
            const vb = activeColumn.sortValue(b);
            if (typeof va === 'number' && typeof vb === 'number') {
                return (va - vb) * dir;
            }
            return String(va).localeCompare(String(vb)) * dir;
        });
    }, [filteredNamespaces, sortDirection, sortField, sortableColumns]);

    const columns = useMemo(
        () => [...orderedColumns, ageColumn, actionsColumn],
        [orderedColumns, ageColumn, actionsColumn]
    );

    const gridTemplateColumns = useMemo(
        () => columns.map((col) => col.width || 'minmax(120px, 1fr)').join(' '),
        [columns]
    );


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
                            className="w-full bg-gray-900 border border-gray-700 text-gray-200 text-sm rounded-md pl-10 pr-10 py-2 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all duration-300"
                        />
                        {filter && (
                            <button
                                onClick={() => setFilter('')}
                                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-200 transition-colors"
                                type="button"
                            >
                                <X size={16} />
                            </button>
                        )}
                    </div>
                    <span className="text-sm text-gray-500">
                        {filteredNamespaces.length} {filteredNamespaces.length === 1 ? 'item' : 'items'}
                    </span>
                </div>
                <div className="flex items-center space-x-2">
                    {isAdmin(user) && (
                        <button
                            onClick={() => {
                                setEditingYaml({
                                    kind: 'Namespace',
                                    namespaced: false,
                                    isNew: true
                                });
                            }}
                            className="flex items-center px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-sm transition-colors"
                            title="Create new namespace"
                        >
                            <Plus size={16} className="mr-1.5" />
                            Add
                        </button>
                    )}
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
            <div
                className="grid gap-4 px-6 py-3 border-b border-gray-800 bg-gray-900/50 text-xs font-medium text-gray-500 uppercase tracking-wider"
                style={{ gridTemplateColumns }}
            >
                {columns.map((column) => {
                    const isSortable = typeof column.sortValue === 'function';
                    const canDrag = !column.pinned && !column.isAction;
                    return (
                        <div
                            key={column.id}
                            className={`flex items-center ${column.align === 'left' ? 'justify-start' : 'justify-center'} ${column.id === 'name' ? 'pl-[0.5cm]' : ''} gap-2`}
                            draggable={canDrag}
                            onDragStart={() => {
                                if (canDrag) setDraggingColumn(column.id);
                            }}
                            onDragOver={(e) => {
                                if (canDrag && draggingColumn) {
                                    e.preventDefault();
                                }
                            }}
                            onDrop={(e) => {
                                if (canDrag && draggingColumn && draggingColumn !== column.id) {
                                    e.preventDefault();
                                    moveColumn(draggingColumn, column.id);
                                    setDraggingColumn(null);
                                }
                            }}
                            onDragEnd={() => setDraggingColumn(null)}
                        >
                            {canDrag && <GripVertical size={12} className="text-gray-600" />}
                            {isSortable ? (
                                <button
                                    type="button"
                                    className="flex items-center gap-1 hover:text-gray-300"
                                    onClick={() => handleSort(column.id)}
                                >
                                    {column.label} {renderSortIndicator(column.id)}
                                </button>
                            ) : (
                                <span>{column.label}</span>
                            )}
                        </div>
                    );
                })}
            </div>

            {/* Table Body */}
            <div className="flex-1 overflow-y-auto">
                {sortedNamespaces.map((ns) => {
                    const isExpanded = expandedId === ns.name;
                    return (
                        <div key={ns.name} className="border-b border-gray-800 last:border-0">
                            <div
                                onClick={() => toggleExpand(ns.name)}
                                className={`grid gap-4 px-6 py-4 cursor-pointer transition-colors duration-200 items-center ${getExpandableRowRowClasses(isExpanded)}`}
                                style={{ gridTemplateColumns }}
                            >
                                {columns.map((column) => (
                                    <div
                                        key={`${ns.name}-${column.id}`}
                                        className={`${column.align === 'left' ? 'justify-start text-left' : 'justify-center text-center'} flex items-center ${column.id === 'name' ? 'pl-[0.5cm]' : ''}`}
                                        onClick={column.isAction ? (e) => e.stopPropagation() : undefined}
                                    >
                                        {column.renderCell(ns, { isExpanded })}
                                    </div>
                                ))}
                            </div>

                            {/* Expanded Details */}
                            <div className={`${getExpandableRowClasses(isExpanded, false)}`}>
                                {isExpanded && (
                                    <div className="px-6 py-4 bg-gray-900/30 border-t border-gray-800">
                                        <div className="bg-gray-900/50 rounded-lg border border-gray-800 overflow-hidden">
                                            <div className="p-4 space-y-6">
                                            {/* Edit YAML Button */}
                                            <div className="flex justify-end mb-2">
                                                <EditYamlButton onClick={() => setEditingYaml({ name: ns.name, kind: 'Namespace', namespaced: false })} />
                                            </div>
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
