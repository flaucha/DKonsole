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

import { getExpandableRowClasses, getExpandableRowRowClasses } from '../utils/expandableRow';

// Detail components
import NodeDetails from './details/NodeDetails';
import { ServiceAccountDetails, RoleDetails, BindingDetails } from './details/RbacDetails';
import DeploymentDetails from './details/DeploymentDetails';
import ServiceDetails from './details/ServiceDetails';
import IngressDetails from './details/IngressDetails';
import PodDetails from './details/PodDetails';
import { ConfigMapDetails, SecretDetails } from './details/ConfigDetails';
import NetworkPolicyDetails from './details/NetworkPolicyDetails';
import StorageDetails from './details/StorageDetails';
import StorageClassDetails from './details/StorageClassDetails';
import { JobDetails, CronJobDetails, StatefulSetDetails, DaemonSetDetails, HPADetails } from './details/WorkloadDetails';
import GenericDetails from './details/GenericDetails';
import { EditYamlButton } from './details/CommonDetails';
import YamlEditor from './YamlEditor';

// Cells and Modals
import ActionsCell from './workloads/cells/ActionsCell';
import DeleteConfirmationModal from './workloads/modals/DeleteConfirmationModal';
import RolloutConfirmationModal from './workloads/modals/RolloutConfirmationModal';
import JobCreatedModal from './workloads/modals/JobCreatedModal';
import WorkloadToolbar from './workloads/WorkloadToolbar';

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

    const renderDetails = (res) => {
        const onEditYAML = () => setEditingResource(res);
        const onDataSaved = () => {
            setExpandedId(null);
            setTimeout(() => refetch(), 300);
        };
        if (!res.details) return (
            <div className="p-4 text-gray-500 italic">
                No details available.
            </div>
        );
        switch (res.kind) {
            case 'Node': return <NodeDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'ServiceAccount': return <ServiceAccountDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'Role':
            case 'ClusterRole': return <RoleDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'RoleBinding':
            case 'ClusterRoleBinding': return <BindingDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'Deployment': return <DeploymentDetails details={res.details} onScale={(delta) => handleScale(res, delta)} scaling={scaling === res.name} res={res} onEditYAML={onEditYAML} />;
            case 'Service': return <ServiceDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} name={res.name} />;
            case 'Ingress': return <IngressDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} />;
            case 'Pod': return <PodDetails details={res.details} onEditYAML={onEditYAML} pod={res} />;
            case 'ConfigMap': return <ConfigMapDetails details={res.details} onEditYAML={onEditYAML} resource={res} onDataSaved={onDataSaved} />;
            case 'Secret': return <SecretDetails details={res.details} onEditYAML={onEditYAML} resource={res} onDataSaved={onDataSaved} />;
            case 'NetworkPolicy': return <NetworkPolicyDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'PersistentVolumeClaim':
            case 'PersistentVolume': return <StorageDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} />;
            case 'StorageClass': return <StorageClassDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'Job': return <JobDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} />;
            case 'CronJob': return <CronJobDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'StatefulSet': return <StatefulSetDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} />;
            case 'DaemonSet': return <DaemonSetDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} />;
            case 'HPA': return <HPADetails details={res.details} onEditYAML={onEditYAML} />;
            default: return <GenericDetails details={res.details} onEditYAML={onEditYAML} />;
        }
    };

    const Icon = getIcon(kind);

    // Get columns from hook
    const { dataColumns, ageColumn } = useWorkloadColumns(kind);

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
                handleDelete={handleDelete}
                setConfirmRollout={setConfirmRollout}
            />
        )
    }), [kind, user, currentCluster, authFetch, handleTriggerCronJob, triggering, rollingOut, handleDelete, setConfirmRollout]);

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

    const renderSortIndicator = (id) => {
        if (sortField !== id) return null;
        return sortDirection === 'asc' ? '↑' : '↓';
    };

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
                    ageColumn={ageColumn}
                    hidden={hidden}
                    toggleVisibility={toggleVisibility}
                    resetOrder={resetOrder}
                />
                {/* Empty state message */}
                <div className="flex-1 flex items-center justify-center">
                    <div className="text-gray-500 italic text-center">
                        <Icon size={48} className="mx-auto mb-4 text-gray-600" />
                        <p className="text-lg">No {kind}s found in this namespace.</p>
                        <p className="text-sm mt-2">Try selecting a different namespace or check if resources exist.</p>
                    </div>
                </div>
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
                ageColumn={ageColumn}
                hidden={hidden}
                toggleVisibility={toggleVisibility}
                resetOrder={resetOrder}
            />

            {/* Table Header */}
            <div
                className="grid gap-4 px-6 py-3 border-b border-gray-800 bg-gray-900/50 text-xs font-medium text-gray-500 uppercase tracking-wider"
                style={{ gridTemplateColumns }}
            >
                {columns.map((column) => {
                    const isSortable = typeof column.sortValue === 'function';
                    const canDrag = !column.pinned && !column.isAction;
                    const headerLabel = column.label || '';
                    const dataTestKey = column.label ? `${column.label.replace(/\s+/g, '').toLowerCase()}-header` : undefined;
                    const alignmentClass = column.align === 'left' ? 'justify-start text-left' : 'justify-center text-center';
                    return (
                        <div
                            key={column.id}
                            className={`flex items-center ${alignmentClass} ${column.id === 'name' ? 'pl-[0.5cm]' : ''} gap-2 ${isSortable ? 'cursor-pointer' : ''}`}
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
                                    className={`flex items-center gap-1 hover:text-gray-300 uppercase ${alignmentClass}`}
                                    onClick={() => handleSort(column.id)}
                                    data-testid={dataTestKey}
                                >
                                    {headerLabel} {renderSortIndicator(column.id)}
                                </button>
                            ) : (
                                <span data-testid={dataTestKey} className={`${alignmentClass} uppercase`}>{headerLabel}</span>
                            )}
                        </div>
                    );
                })}
            </div>

            {/* Table Body */}
            <div className="flex-1 overflow-y-auto">
                {limitedResources.map((res) => {
                    const isExpanded = expandedId === res.uid;
                    return (
                        <div key={res.uid} className="border-b border-gray-800 last:border-0">
                            <div
                                onClick={() => toggleExpand(res.uid)}
                                className={`grid gap-4 px-6 py-1.5 cursor-pointer transition-colors duration-200 items-center ${getExpandableRowRowClasses(isExpanded)}`}
                                style={{ gridTemplateColumns }}
                            >
                                {columns.map((column) => (
                                    <div
                                        key={`${res.uid}-${column.id}`}
                                        className={`${column.align === 'left' ? 'justify-start text-left' : 'justify-center text-center'} flex items-center ${column.id === 'name' ? 'pl-[0.5cm]' : ''}`}
                                        onClick={column.isAction ? (e) => e.stopPropagation() : undefined}
                                    >
                                        {column.renderCell(res, { isExpanded })}
                                    </div>
                                ))}
                            </div>

                            {/* Expanded Details */}
                            <div className={`${getExpandableRowClasses(isExpanded, false)}`}>
                                {isExpanded && (
                                    <div className="px-6 py-4 bg-gray-900/30 border-t border-gray-800">
                                        <div className="bg-gray-900/50 rounded-lg border border-gray-800 overflow-hidden relative">
                                            {/* Show Edit YAML button only for resources that don't have it in their detail component */}
                                            {res.kind !== 'Deployment' && res.kind !== 'Pod' && res.kind !== 'ClusterRole' && res.kind !== 'ClusterRoleBinding' &&
                                                res.kind !== 'Role' && res.kind !== 'RoleBinding' &&
                                                res.kind !== 'CronJob' && res.kind !== 'StatefulSet' && res.kind !== 'DaemonSet' && res.kind !== 'HPA' &&
                                                res.kind !== 'Job' && res.kind !== 'PersistentVolumeClaim' && res.kind !== 'PersistentVolume' && res.kind !== 'StorageClass' &&
                                                res.kind !== 'ConfigMap' && res.kind !== 'Secret' && res.kind !== 'NetworkPolicy' && res.kind !== 'Service' && res.kind !== 'Ingress' && (
                                                    <div className="absolute top-4 right-4 z-10">
                                                        <EditYamlButton onClick={() => setEditingResource(res)} namespace={res.namespace} />
                                                    </div>
                                                )}
                                            {renderDetails(res)}
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    );
                })}
                {limitedResources.length === 0 && filter && (
                    <div className="p-8 text-center text-gray-500">
                        No resources match your filter.
                    </div>
                )}
            </div>


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
        </div>
    );
};

export default WorkloadList;
