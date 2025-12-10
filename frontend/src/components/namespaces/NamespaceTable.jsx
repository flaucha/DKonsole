import React, { useState } from 'react';
import { Tag, Clock } from 'lucide-react';
import { formatDateTime } from '../../utils/dateUtils';
import { getExpandableRowClasses, getExpandableRowRowClasses } from '../../utils/expandableRow';



const NamespaceTable = ({
    namespaces,
    columns,
    gridTemplateColumns,
    expandedId,
    toggleExpand,
    sortField,
    sortDirection,
    handleSort,
    moveColumn,
    menuOpen,
    setMenuOpen
}) => {
    const [draggingColumn, setDraggingColumn] = useState(null);

    const renderSortIndicator = (field) => {
        if (sortField !== field) return null;
        return sortDirection === 'asc' ? '↑' : '↓';
    };

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
                {namespaces.map((ns) => {
                    // Safety check if ns is null/undefined
                    if (!ns) return null;
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

export default NamespaceTable;
