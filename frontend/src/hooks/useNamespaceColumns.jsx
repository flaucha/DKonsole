import React, { useMemo } from 'react';
import { Database, MoreVertical, FileText, ChevronDown } from 'lucide-react';
import { getStatusBadgeClass } from '../utils/statusBadge';
import { formatDateParts } from '../utils/dateUtils';

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

export const useNamespaceColumns = ({
    isAdmin,
    onEditYaml,
    onDelete,
    menuOpen,
    setMenuOpen
}) => {
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

    const actionsColumn = useMemo(() => ({
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
    }), [menuOpen, isAdmin, onEditYaml, onDelete, setMenuOpen]);

    return { dataColumns, ageColumn, actionsColumn };
};
