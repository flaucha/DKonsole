import React from 'react';
import { getExpandableRowClasses, getExpandableRowRowClasses } from '../../utils/expandableRow';

// Detail components
import NodeDetails from '../details/NodeDetails';
import { ServiceAccountDetails, RoleDetails, BindingDetails } from '../details/RbacDetails';
import DeploymentDetails from '../details/DeploymentDetails';
import ReplicaSetDetails from '../details/ReplicaSetDetails';
import ServiceDetails from '../details/ServiceDetails';
import IngressDetails from '../details/IngressDetails';
import PodDetails from '../details/PodDetails';
import { ConfigMapDetails, SecretDetails } from '../details/ConfigDetails';
import NetworkPolicyDetails from '../details/NetworkPolicyDetails';
import StorageDetails from '../details/StorageDetails';
import StorageClassDetails from '../details/StorageClassDetails';
import { JobDetails, CronJobDetails, StatefulSetDetails, DaemonSetDetails, HPADetails } from '../details/WorkloadDetails';
import GenericDetails from '../details/GenericDetails';


const WorkloadTable = ({
    resources,
    columns,
    gridTemplateColumns,
    expandedId,
    toggleExpand,
    sortField,
    sortDirection,
    handleSort,
    draggingColumn,
    setDraggingColumn,
    moveColumn,
    setEditingResource,
    handleScale,
    scaling,
    onDetailsSaved,
    filter
}) => {

    const renderSortIndicator = (id) => {
        if (sortField !== id) return null;
        return sortDirection === 'asc' ? '↑' : '↓';
    };

    const renderDetails = (res) => {
        const onEditYAML = () => setEditingResource(res);

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
            case 'ReplicaSet': return <ReplicaSetDetails details={res.details} res={res} />;
            case 'Service': return <ServiceDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} name={res.name} />;
            case 'Ingress': return <IngressDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} />;
            case 'Pod': return <PodDetails details={res.details} onEditYAML={onEditYAML} pod={res} />;
            case 'ConfigMap': return <ConfigMapDetails details={res.details} onEditYAML={onEditYAML} resource={res} onDataSaved={onDetailsSaved} />;
            case 'Secret': return <SecretDetails details={res.details} onEditYAML={onEditYAML} resource={res} onDataSaved={onDetailsSaved} />;
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

    return (
        <>
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
                {resources.map((res) => {
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
                                            {/* Edit YAML buttons removed as per UI refinement request (redundant with Kebab menu) */}
                                            {renderDetails(res)}
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    );
                })}
                {resources.length === 0 && filter && (
                    <div className="p-8 text-center text-gray-500">
                        No resources match your filter.
                    </div>
                )}
            </div>
        </>
    );
};

export default WorkloadTable;
