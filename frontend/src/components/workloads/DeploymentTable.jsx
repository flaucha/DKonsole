import React, { useState } from 'react';
import {
    Box,
    FileText,
    Layers,
    Network,
    HardDrive,
    Tag,
    CirclePlus,
    CircleMinus,
    MoreVertical,
    Plus,
    Minus
} from 'lucide-react';
import YamlEditor from '../YamlEditor';
import { useSettings } from '../../context/SettingsContext';
import { useAuth } from '../../context/AuthContext';
import { formatDateTimeShort } from '../../utils/dateUtils';
import { getExpandableRowClasses, getExpandableCellClasses, getExpandableRowStyles, getExpandableRowRowClasses } from '../../utils/expandableRow';
import { getStatusBadgeClass } from '../../utils/statusBadge';

const DetailRow = ({ label, value, icon: Icon, children }) => (
    <div className="flex items-center justify-between bg-gray-800 px-3 py-2 rounded border border-gray-700 mb-2">
        <div className="flex items-center">
            {Icon && <Icon size={14} className="mr-2 text-gray-500" />}
            <span className="text-xs text-gray-400">{label}</span>
        </div>
        <div className="flex items-center">
            <span className="text-sm font-mono text-white break-all text-right">
                {Array.isArray(value) ? (
                    value.length > 0 ? value.join(', ') : <span className="text-gray-600 italic">None</span>
                ) : (
                    value || <span className="text-gray-600 italic">None</span>
                )}
            </span>
            {children}
        </div>
    </div>
);

const EditYamlButton = ({ onClick }) => (
    <button
        onClick={onClick}
        className="flex items-center px-2.5 py-1 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-md border border-gray-600 text-xs transition-colors"
    >
        <FileText size={12} className="mr-1.5" />
        Edit YAML
    </button>
);

const DeploymentDetails = ({ details, onScale, scaling, res, onEditYAML }) => {
    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <div className="mb-2">
                        <DetailRow label="Replicas" value={`${details.ready} / ${details.replicas}`} icon={Layers}>
                            {onScale && (
                                <div className="flex items-center space-x-1 ml-2">
                                    <button
                                        onClick={() => onScale(-1)}
                                        disabled={scaling}
                                        className="p-1 rounded bg-gray-700 border border-gray-600 text-gray-200 hover:bg-gray-600 disabled:opacity-50"
                                        title="Scale down"
                                    >
                                        <Minus size={12} />
                                    </button>
                                    <button
                                        onClick={() => onScale(1)}
                                        disabled={scaling}
                                        className="p-1 rounded bg-gray-700 border border-gray-600 text-gray-200 hover:bg-gray-600 disabled:opacity-50"
                                        title="Scale up"
                                    >
                                        <Plus size={12} />
                                    </button>
                                </div>
                            )}
                        </DetailRow>
                    </div>
                    <DetailRow label="Images" value={details.images} icon={Box} />
                    <DetailRow label="Ports" value={details.ports?.map(p => p.toString())} icon={Network} />
                </div>
                <div>
                    <DetailRow label="PVCs" value={details.pvcs} icon={HardDrive} />
                    <DetailRow
                        label="Labels"
                        value={details.podLabels ? Object.entries(details.podLabels).map(([k, v]) => `${k}=${v}`) : []}
                        icon={Tag}
                    />
                </div>
            </div>
            <div className="flex justify-end mt-4">
                <EditYamlButton onClick={onEditYAML} />
            </div>
        </div>
    );
};

const DeploymentTable = ({ namespace, resources, loading, onReload }) => {
    const [expandedId, setExpandedId] = useState(null);
    const [editingResource, setEditingResource] = useState(null);
    const [sortField, setSortField] = useState('name');
    const [sortDirection, setSortDirection] = useState('asc');
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [scaling, setScaling] = useState(null);
    const [filter, setFilter] = useState('');
    const [isSearchFocused, setIsSearchFocused] = useState(false);

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

    const filteredResources = resources.filter(r => {
        if (!filter) return true;
        return r.name.toLowerCase().includes(filter.toLowerCase());
    });

    const sortedResources = [...filteredResources].sort((a, b) => {
        const dir = sortDirection === 'asc' ? 1 : -1;
        const getVal = (item) => {
            switch (sortField) {
                case 'name':
                    return item.name || '';
                case 'status':
                    return item.status || '';
                case 'created':
                    return new Date(item.created).getTime() || 0;
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

    const renderSortIndicator = (field) => {
        if (sortField !== field) return null;
        return sortDirection === 'asc' ? '↑' : '↓';
    };

    const toggleExpand = (uid) => {
        setExpandedId(current => current === uid ? null : uid);
    };

    const renderDetails = (res) => {
        const onEditYAML = () => setEditingResource(res);
        if (!res.details) return (
            <div className="p-4 text-gray-500 italic">
                No details available.
                <div className="flex justify-end mt-4">
                    <EditYamlButton onClick={onEditYAML} />
                </div>
            </div>
        );
        return (
            <DeploymentDetails
                details={res.details}
                onScale={(delta) => handleScale(res, delta)}
                scaling={scaling === res.name}
                res={res}
                onEditYAML={onEditYAML}
            />
        );
    };

    const handleScale = (res, delta) => {
        if (!res.namespace) return;
        setScaling(res.name);
        const params = new URLSearchParams({
            kind: 'Deployment',
            name: res.name,
            namespace: res.namespace,
            delta: String(delta),
        });
        if (currentCluster) params.append('cluster', currentCluster);

        authFetch(`/api/scale?${params.toString()}`, { method: 'POST' })
            .then(async (resp) => {
                if (!resp.ok) throw new Error('Scale failed');
                onReload();
            })
            .catch((err) => alert(err.message))
            .finally(() => setScaling(null));
    };

    const triggerDelete = (res, force = false) => {
        const params = new URLSearchParams({ kind: res.kind, name: res.name });
        if (res.namespace) params.append('namespace', res.namespace);
        if (currentCluster) params.append('cluster', currentCluster);
        if (force) params.append('force', 'true');

        authFetch(`/api/resource?${params.toString()}`, { method: 'DELETE' })
            .then(async (resp) => {
                if (!resp.ok) {
                    const text = await resp.text();
                    throw new Error(text || 'Delete failed');
                }
                onReload();
            })
            .catch((err) => {
                alert(err.message || 'Failed to delete resource');
            })
            .finally(() => setConfirmAction(null));
    };

    return (
        <>
            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(null)}
                ></div>
            )}
            <div className="bg-gray-800 rounded-lg border border-gray-700 overflow-x-auto">
                <table className="min-w-full border-separate border-spacing-0">
                    <thead>
                        <tr>
                            <th className="w-8 px-2 md:px-4 py-3 bg-gray-900 rounded-tl-lg border-b border-gray-700"></th>
                            <th
                                scope="col"
                                className="px-3 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('name')}
                            >
                                Name <span className="inline-block text-[10px]">{renderSortIndicator('name')}</span>
                            </th>
                            <th
                                scope="col"
                                className="px-2 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('status')}
                            >
                                Status <span className="inline-block text-[10px]">{renderSortIndicator('status')}</span>
                            </th>
                            <th
                                scope="col"
                                className="px-2 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('created')}
                            >
                                Created <span className="inline-block text-[10px]">{renderSortIndicator('created')}</span>
                            </th>
                            <th className="w-10 px-4 py-3 bg-gray-900 rounded-tr-lg border-b border-gray-700"></th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-800">
                        {sortedResources.map((res) => (
                            <React.Fragment key={res.uid}>
                                <tr
                                    onClick={() => toggleExpand(res.uid)}
                                    className={getExpandableRowRowClasses(expandedId === res.uid)}
                                >
                                    <td className="px-2 md:px-4 py-3 whitespace-nowrap text-gray-400 text-center">
                                        {expandedId === res.uid ? <CircleMinus size={16} /> : <CirclePlus size={16} />}
                                    </td>
                                    <td className="px-3 md:px-6 py-3 whitespace-nowrap">
                                        <div className="flex items-center">
                                            <div className="flex-shrink-0 h-6 w-6 bg-gray-700 rounded flex items-center justify-center text-gray-400">
                                                <Box size={14} />
                                            </div>
                                            <div className="ml-4">
                                                <div className="text-sm font-medium text-white">{res.name}</div>
                                            </div>
                                        </div>
                                    </td>
                                    <td className="px-2 md:px-6 py-3 whitespace-nowrap">
                                        <span className={`px-2 py-1 text-xs rounded-full ${getStatusBadgeClass(res.status)}`}>
                                            {res.status}
                                        </span>
                                    </td>
                                    <td className="px-2 md:px-6 py-3 whitespace-nowrap text-sm text-gray-400">
                                        {formatDateTimeShort(res.created)}
                                    </td>
                                    <td className="px-4 py-3 whitespace-nowrap text-gray-300" onClick={(e) => e.stopPropagation()}>
                                        <div className="relative flex items-center justify-end space-x-1">
                                            <button
                                                onClick={() => setMenuOpen(menuOpen === res.name ? null : res.name)}
                                                className="p-1 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                                            >
                                                <MoreVertical size={16} />
                                            </button>
                                            {menuOpen === res.name && (
                                                <div className="absolute right-0 mt-1 w-36 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
                                                    <div className="flex flex-col">
                                                        <button
                                                            onClick={() => {
                                                                setConfirmAction({ res, force: false });
                                                                setMenuOpen(null);
                                                            }}
                                                            className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                        >
                                                            Delete
                                                        </button>
                                                        <button
                                                            onClick={() => {
                                                                setConfirmAction({ res, force: true });
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
                                    <td colSpan={5} className={getExpandableCellClasses(expandedId === res.uid, 5)}>
                                        <div
                                            className={getExpandableRowClasses(expandedId === res.uid, true)}
                                            style={getExpandableRowStyles(expandedId === res.uid, res.kind)}
                                        >
                                            {expandedId === res.uid && renderDetails(res)}
                                        </div>
                                    </td>
                                </tr>
                            </React.Fragment>
                        ))}
                    </tbody>
                </table>
            </div>
            {editingResource && (
                <YamlEditor
                    resource={editingResource}
                    onClose={() => setEditingResource(null)}
                    onSaved={() => {
                        setEditingResource(null);
                        onReload();
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
                            {confirmAction.force ? 'Force delete' : 'Delete'} {confirmAction.res.kind} "{confirmAction.res.name}"?
                        </p>
                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setConfirmAction(null)}
                                className="px-4 py-2 bg-gray-800 text-gray-200 rounded-md hover:bg-gray-700 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={() => {
                                    triggerDelete(confirmAction.res, confirmAction.force);
                                }}
                                className={`px-4 py-2 rounded-md text-white transition-colors ${confirmAction.force ? 'bg-red-700 hover:bg-red-800' : 'bg-orange-600 hover:bg-orange-700'}`}
                            >
                                {confirmAction.force ? 'Force delete' : 'Delete'}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </>
    );
};

export default DeploymentTable;

