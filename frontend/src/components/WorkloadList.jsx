import React, { useMemo } from 'react';
import { useWorkloadListState } from '../hooks/useWorkloadListState';
import useWorkloadColumns from '../hooks/useWorkloadColumns';
import { useColumnOrder } from '../hooks/useColumnOrder';
import { getIcon } from '../utils/workloadUtils';
import { RefreshCw } from 'lucide-react';

// Components
import YamlEditor from './YamlEditor';
import ActionsCell from './workloads/cells/ActionsCell';
import WorkloadToolbar from './workloads/WorkloadToolbar';
import WorkloadTable from './workloads/WorkloadTable';
import WorkloadModals from './WorkloadModals';

const WorkloadList = ({ namespace, kind }) => {
    // Early return if kind is not provided
    if (!kind) {
        return <div className="text-red-400 p-6">Error: Resource type not specified.</div>;
    }

    const {
        // Data
        resources,
        loading,
        error,
        refetch,
        user,
        currentCluster,
        authFetch,

        // UI State
        expandedId,
        sortField,
        sortDirection,
        filter,
        isSearchFocused,
        menuOpen,
        draggingColumn,

        // Setters
        setFilter,
        setIsSearchFocused,
        setMenuOpen,
        setDraggingColumn,
        setEditingResource,
        setEditingDataResource,

        // Modal State
        editingResource,
        editingDataResource,

        // Actions
        actions,
        toggleExpand,
        handleSort,
        handleAdd,
        onDetailsSaved
    } = useWorkloadListState(namespace, kind);

    const {
        setConfirmAction,
        setConfirmRollout,
        triggering,
        rollingOut,
        handleTriggerCronJob,
        scaling,
        handleDelete,
        handleScale,
        confirmAction,
        confirmRollout,
        handleRolloutDeployment,
        createdJob,
        setCreatedJob
    } = actions;

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

            <WorkloadModals
                editingResource={editingResource}
                setEditingResource={setEditingResource}
                editingDataResource={editingDataResource}
                setEditingDataResource={setEditingDataResource}
                confirmAction={confirmAction}
                setConfirmAction={setConfirmAction}
                confirmRollout={confirmRollout}
                setConfirmRollout={setConfirmRollout}
                createdJob={createdJob}
                setCreatedJob={setCreatedJob}
                handleDelete={handleDelete}
                handleRolloutDeployment={handleRolloutDeployment}
                refetch={refetch}
                menuOpen={menuOpen}
                setMenuOpen={setMenuOpen}
            />
        </div>
    );
};

export default WorkloadList;
