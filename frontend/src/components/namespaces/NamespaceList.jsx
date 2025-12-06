import React, { useMemo, useState } from 'react';
import { Database, MoreVertical, FileText, ChevronDown, GripVertical, Clock, Tag } from 'lucide-react';
import { getStatusBadgeClass } from '../../utils/statusBadge';
import { formatDateTime, formatDateParts } from '../../utils/dateUtils';
import { getExpandableRowClasses, getExpandableRowRowClasses } from '../../utils/expandableRow';
import { useColumnOrder } from '../../hooks/useColumnOrder';

const DateStack = ({ value }) => {
    const { date, time } = formatDateParts(value);
    return (
        <div className="flex flex-col items-center leading-tight text-sm text-gray-300">
            <span>{date}</span>
            <span className="text-xs text-gray-500">{time}</span>
        </div>
    );
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

const parseDateValue = (value) => {
    const timestamp = new Date(value).getTime();
    return Number.isFinite(timestamp) ? timestamp : 0;
};

const NamespaceList = ({
    namespaces,
    sortField,
    setSortField,
    sortDirection,
    setSortDirection,
    expandedId,
    setExpandedId,
    isAdmin,
    onEditYaml,
    onDelete,
    user
}) => {
    const [draggingColumn, setDraggingColumn] = useState(null);
    const [menuOpen, setMenuOpen] = useState(null);

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
                                {isAdmin ? (
                                    <>
                                        <button
                                            onClick={() => {
                                                onEditYaml({ name: item.name, kind: 'Namespace', namespaced: false });
                                                setMenuOpen(null);
                                            }}
                                            className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                        >
                                            Edit YAML
                                        </button>
                                        <button
                                            onClick={() => {
                                                onDelete({ namespace: item.name, force: false });
                                                setMenuOpen(null);
                                            }}
                                            className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                        >
                                            Delete
                                        </button>
                                        <button
                                            onClick={() => {
                                                onDelete({ namespace: item.name, force: true });
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
        if (!activeColumn) return namespaces;
        return [...namespaces].sort((a, b) => {
            const va = activeColumn.sortValue(a);
            const vb = activeColumn.sortValue(b);
            if (typeof va === 'number' && typeof vb === 'number') {
                return (va - vb) * dir;
            }
            return String(va).localeCompare(String(vb)) * dir;
        });
    }, [namespaces, sortDirection, sortField, sortableColumns]);

    const columns = useMemo(
        () => [...orderedColumns, ageColumn, actionsColumn],
        [orderedColumns, ageColumn, actionsColumn]
    );

    const gridTemplateColumns = useMemo(
        () => columns.map((col) => col.width || 'minmax(120px, 1fr)').join(' '),
        [columns]
    );

    return (
        <div className="flex-1 flex flex-col overflow-hidden">
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
                            className={`flex items-center ${column.align === 'left' ? 'justify-start' : 'justify-center'} gap-2`}
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
                            {isSortable ? (
                                <button
                                    type="button"
                                    className="flex items-center gap-1 hover:text-gray-300 uppercase"
                                    onClick={() => handleSort(column.id)}
                                >
                                    {column.label} {renderSortIndicator(column.id)}
                                </button>
                            ) : (
                                <span className="uppercase">{column.label}</span>
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
                                className={`grid gap-4 px-6 py-1.5 cursor-pointer transition-colors duration-200 items-center ${getExpandableRowRowClasses(isExpanded)}`}
                                style={{ gridTemplateColumns }}
                            >
                                {columns.map((column) => (
                                    <div
                                        key={`${ns.name}-${column.id}`}
                                        className={`${column.align === 'left' ? 'justify-start text-left' : 'justify-center text-center'} flex items-center`}
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
                                                    <EditYamlButton onClick={() => onEditYaml({ name: ns.name, kind: 'Namespace', namespaced: false })} />
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
                {namespaces.length === 0 && (
                    <div className="p-8 text-center text-gray-500">
                        {/* Message passed from parent usually, but good enough here */}
                        No namespaces found.
                    </div>
                )}
            </div>

            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(null)}
                ></div>
            )}
        </div>
    );
};

export default NamespaceList;
