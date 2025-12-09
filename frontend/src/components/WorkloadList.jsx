import React, { useState, useEffect, useRef, useMemo } from 'react';
import { useLocation } from 'react-router-dom';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { useWorkloads } from '../hooks/useWorkloads';
import { useColumnOrder } from '../hooks/useColumnOrder';
import { useWorkloadActions } from '../hooks/useWorkloadActions';
import useWorkloadColumns from '../hooks/useWorkloadColumns';
import { getIcon } from '../utils/workloadUtils';
import { RefreshCw } from 'lucide-react';

import YamlEditor from './YamlEditor';
import { DataEditor } from './details/DataEditor';

// Cells and Modals
import ActionsCell from './workloads/cells/ActionsCell';
import DeleteConfirmationModal from './workloads/modals/DeleteConfirmationModal';
import RolloutConfirmationModal from './workloads/modals/RolloutConfirmationModal';
import JobCreatedModal from './workloads/modals/JobCreatedModal';
import WorkloadToolbar from './workloads/WorkloadToolbar';
import WorkloadTable from './workloads/WorkloadTable';

const WorkloadList = ({ namespace, kind }) => {
    const [expandedId, setExpandedId] = useState(null);
    const [sortField, setSortField] = useState('name');
    const [sortDirection, setSortDirection] = useState('asc');
    const { currentCluster } = useSettings();
    const { authFetch, user } = useAuth();
    const location = useLocation();
    const [filter, setFilter] = useState('');
    const [isSearchFocused, setIsSearchFocused] = useState(false);
    const [editingResource, setEditingResource] = useState(null);
    const [editingDataResource, setEditingDataResource] = useState(null);
    const [menuOpen, setMenuOpen] = useState(null);
    const [draggingColumn, setDraggingColumn] = useState(null);

    // Early return if kind is not provided
    if (!kind) {
        return <div className="text-red-400 p-6">Error: Resource type not specified.</div>;
    }

    // Use React Query hook to fetch all resources
    const { data: resourcesData, isLoading: loading, error, refetch } = useWorkloads(authFetch, namespace, kind, currentCluster);

    const {
        confirmAction,
        setConfirmAction,
        confirmRollout,
        setConfirmRollout,
        scaling,
        triggering,
        createdJob,
        setCreatedJob,
        rollingOut,
        handleDelete,
        handleScale,
        handleTriggerCronJob,
        handleRolloutDeployment
    } = useWorkloadActions(authFetch, refetch, currentCluster);

    // Track previous kind to detect changes
    const prevKindRef = useRef(kind);
    const prevNamespaceRef = useRef(namespace);
    const prevClusterRef = useRef(currentCluster);

    // Reset when namespace, kind, or cluster changes
    const [allResources, setAllResources] = useState([]);

    useEffect(() => {
        const kindChanged = prevKindRef.current !== kind;
        const namespaceChanged = prevNamespaceRef.current !== namespace;
        const clusterChanged = prevClusterRef.current !== currentCluster;

        if (kindChanged || namespaceChanged || clusterChanged) {
            setAllResources([]);
            setExpandedId(null);
            setFilter('');
            setConfirmAction(null);
            setConfirmRollout(null);

            prevKindRef.current = kind;
            prevNamespaceRef.current = namespace;
            prevClusterRef.current = currentCluster;

            if (namespace && kind) {
                refetch();
            }
        }
    }, [namespace, kind, currentCluster, refetch, setConfirmAction, setConfirmRollout]);

    // Update resources state
    useEffect(() => {
        if (resourcesData) {
            let data = [];
            if (Array.isArray(resourcesData)) {
                data = resourcesData;
            } else if (resourcesData.resources && Array.isArray(resourcesData.resources)) {
                data = resourcesData.resources;
            }
            setAllResources(data);

            const searchParams = new URLSearchParams(location.search);
            const search = searchParams.get('search');
            if (search) {
                setFilter(search);
                const found = data.find(r => r.name === search);
                if (found) {
                    setExpandedId(found.uid);
                }
            }
        } else if (!loading) {
            setAllResources([]);
        }
    }, [resourcesData, loading, location.search]);

    const resources = allResources;

    const toggleExpand = (uid) => {
        setExpandedId(current => current === uid ? null : uid);
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

    const filteredResources = resources.filter(r => {
        if (!filter) return true;
        return r.name.toLowerCase().includes(filter.toLowerCase());
    });

    const Icon = getIcon(kind);

    // Get columns from hook
    const { dataColumns } = useWorkloadColumns(kind);

    // Actions column definition
    const actionsColumn = useMemo(() => ({
        id: 'actions',
        label: '',
        width: '60px',
        pinned: true,
        isAction: true,
        align: 'center',
        renderCell: (item) => (
            <ActionsCell
                res={item}
                kind={kind}
                user={user}
                currentCluster={currentCluster}
                authFetch={authFetch}
                handleTriggerCronJob={handleTriggerCronJob}
                triggering={triggering}
                rollingOut={rollingOut}
                setConfirmAction={setConfirmAction}
                setConfirmRollout={setConfirmRollout}
                onEditYaml={(item) => setEditingResource(item)}
                onEditInPlace={(item) => setEditingDataResource(item)}
            />
        )
    }), [kind, user, currentCluster, authFetch, handleTriggerCronJob, triggering, rollingOut, setConfirmAction, setConfirmRollout]);

    const reorderableColumns = useMemo(
        () => dataColumns.filter((col) => !col.pinned && !col.isAction),
        [dataColumns]
    );

    const {
        orderedColumns: orderedDataColumns,
        visibleColumns: visibleDataColumns,
        moveColumn,
        hidden,
        toggleVisibility,
        resetOrder
    } = useColumnOrder(
        reorderableColumns,
        `dkonsole-columns-${kind}`,
        user?.username
    );

    const sortableColumns = useMemo(
        () => dataColumns.filter((col) => typeof col.sortValue === 'function'),
        [dataColumns]
    );

    useEffect(() => {
        if (sortableColumns.length === 0) return;
        const validIds = new Set(sortableColumns.map((col) => col.id));
        if (!validIds.has(sortField)) {
            setSortField('name');
            setSortDirection('asc');
        }
    }, [sortableColumns, sortField]);

    const sortedResources = useMemo(() => {
        const dir = sortDirection === 'asc' ? 1 : -1;
        const activeColumn = sortableColumns.find((col) => col.id === sortField) || sortableColumns[0];
        if (!activeColumn) return filteredResources;
        return [...filteredResources].sort((a, b) => {
            const va = activeColumn.sortValue(a);
            const vb = activeColumn.sortValue(b);
            if (typeof va === 'number' && typeof vb === 'number') {
                return (va - vb) * dir;
            }
            return String(va).localeCompare(String(vb)) * dir;
        });
    }, [filteredResources, sortDirection, sortField, sortableColumns]);

    const limitedResources = sortedResources;

    const columns = useMemo(
        () => {
            const visibleCols = visibleDataColumns;
            return [...visibleCols, actionsColumn];
        },
        [visibleDataColumns, actionsColumn]
    );

    const gridTemplateColumns = useMemo(
        () => columns.map((col) => col.width || 'minmax(120px, 1fr)').join(' '),
        [columns]
    );

    // Show loading state
    if (loading && resources.length === 0) {
        return <div className="text-gray-400 animate-pulse p-6">Loading {kind}s...</div>;
    }

    // Show error state
    if (error) {
        return (
            <div className="flex flex-col h-full">
                <div className="flex items-center justify-between p-4 border-b border-gray-800 bg-gray-900/50">
                    <div className="flex items-center space-x-4 flex-1">
                        <span className="text-sm text-gray-500">0 items</span>
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
                <div className="flex-1 flex items-center justify-center">
                    <div className="text-red-400 text-center">
                        <p className="text-lg font-semibold mb-2">Error loading {kind}s</p>
                        <p className="text-sm text-gray-500">{error.message}</p>
                    </div>
                </div>
            </div>
        );
    }

    const handleAdd = () => {
        // Determine the best namespace to use for new resources
        let targetNs = namespace;
        if (namespace === 'all') {
            // When viewing all namespaces, use the first available from resources or fallback to dkonsole
            const firstResource = allResources.find(r => r.namespace);
            targetNs = firstResource?.namespace || 'dkonsole';
        }
        setEditingResource({
            kind: kind,
            namespaced: true,
            isNew: true,
            namespace: targetNs,
            apiVersion: 'v1', // This should technically vary by kind, YamlEditor might handle defaults
            metadata: {
                namespace: targetNs,
                name: `new-${kind.toLowerCase()}`
            }
        });
    };

    // Show empty state if no resources and no filter (and not loading)
    if (!loading && resources.length === 0 && !filter) {
        return (
            <div className="flex flex-col h-full">
                <WorkloadToolbar
                    kind={kind}
                    filter={filter}
                    setFilter={setFilter}
                    isSearchFocused={isSearchFocused}
                    setIsSearchFocused={setIsSearchFocused}
                    refetch={refetch}
                    resourcesCount={0}
                    menuOpen={menuOpen}
                    setMenuOpen={setMenuOpen}
                    orderedDataColumns={orderedDataColumns}

                    hidden={hidden}
                    toggleVisibility={toggleVisibility}
                    resetOrder={resetOrder}
                    onAdd={handleAdd}
                />
                {/* Empty state message */}
                <div className="flex-1 flex items-center justify-center">
                    <div className="text-gray-500 italic text-center">
                        <Icon size={48} className="mx-auto mb-4 text-gray-600" />
                        <p className="text-lg">No {kind}s found in this namespace.</p>
                        <p className="text-sm mt-2">Try selecting a different namespace or check if resources exist.</p>
                    </div>
                </div>

                {/* YAML Editor Modal - also needed in empty state */}
                {editingResource && (
                    <YamlEditor
                        resource={editingResource}
                        onClose={() => setEditingResource(null)}
                        onSaved={() => {
                            setEditingResource(null);
                            refetch();
                        }}
                    />
                )}
            </div>
        );
    }

    const onDetailsSaved = () => {
        setExpandedId(null);
        setTimeout(() => refetch(), 300);
    };

    return (
        <div className="flex flex-col h-full">
            <WorkloadToolbar
                kind={kind}
                filter={filter}
                setFilter={setFilter}
                isSearchFocused={isSearchFocused}
                setIsSearchFocused={setIsSearchFocused}
                refetch={refetch}
                resourcesCount={limitedResources.length}
                menuOpen={menuOpen}
                setMenuOpen={setMenuOpen}
                orderedDataColumns={orderedDataColumns}

                hidden={hidden}
                toggleVisibility={toggleVisibility}
                resetOrder={resetOrder}
                onAdd={handleAdd}
            />

            <WorkloadTable
                kind={kind}
                resources={limitedResources}
                columns={columns}
                gridTemplateColumns={gridTemplateColumns}
                expandedId={expandedId}
                toggleExpand={toggleExpand}
                sortField={sortField}
                sortDirection={sortDirection}
                handleSort={handleSort}
                draggingColumn={draggingColumn}
                setDraggingColumn={setDraggingColumn}
                moveColumn={moveColumn}
                setEditingResource={setEditingResource}
                handleScale={handleScale}
                scaling={scaling}
                onDetailsSaved={onDetailsSaved}
                filter={filter}
            />

            {/* YAML Editor Modal */}
            {editingResource && (
                <YamlEditor
                    resource={editingResource}
                    onClose={() => setEditingResource(null)}
                    onSaved={() => {
                        setEditingResource(null);
                        refetch();
                    }}
                />
            )}

            {/* Menu overlay */}
            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(null)}
                ></div>
            )}

            {/* Delete confirmation modal */}
            <DeleteConfirmationModal
                confirmAction={confirmAction}
                setConfirmAction={setConfirmAction}
                handleDelete={handleDelete}
            />

            {/* Rollout confirmation modal */}
            <RolloutConfirmationModal
                confirmRollout={confirmRollout}
                setConfirmRollout={setConfirmRollout}
                handleRolloutDeployment={handleRolloutDeployment}
            />

            {/* Job created success modal */}
            <JobCreatedModal
                createdJob={createdJob}
                setCreatedJob={setCreatedJob}
            />

            {/* Data Editor Modal - Edit In Place */}
            {editingDataResource && (
                <DataEditor
                    resource={editingDataResource}
                    data={editingDataResource.details?.data || {}}
                    isSecret={editingDataResource.kind === 'Secret'}
                    onClose={() => setEditingDataResource(null)}
                    onSaved={() => {
                        setEditingDataResource(null);
                        refetch();
                    }}
                />
            )}
        </div>
    );
};

export default WorkloadList;
