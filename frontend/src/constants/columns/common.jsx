import React from 'react';
import { ChevronDown } from 'lucide-react';
import StatusCell from '../../components/workloads/cells/StatusCell';
import AgeCell from '../../components/workloads/cells/AgeCell';
import { parseDateValue } from '../../utils/workloadUtils';

export const getBaseColumns = (kind) => {
    return [
        {
            id: 'name',
            label: 'Name',
            width: 'minmax(220px, 2.2fr)',
            sortValue: (item) => item.name || '',
            align: 'left',
            renderCell: (item, { isExpanded }) => (
                <div className="font-semibold text-gray-200 flex items-center space-x-2 truncate">
                    <ChevronDown
                        size={16}
                        className={`text-gray-500 transition-transform duration-200 flex-shrink-0 ${isExpanded ? 'transform rotate-180' : ''}`}
                    />
                    <span className="truncate" title={item.name}>{item.name}</span>
                </div>
            )
        },
        {
            id: 'status',
            label: kind === 'Secret' || kind === 'Service' ? 'Type' : 'Status',
            width: 'minmax(120px, 0.9fr)',
            sortValue: (item) => item.status || '',
            align: 'center',
            renderCell: (item) => <StatusCell status={item.status} kind={kind} />
        },
        {
            id: 'namespace',
            label: 'Namespace',
            width: 'minmax(140px, 1fr)',
            sortValue: (item) => item.namespace || '',
            align: 'center',
            hiddenByDefault: true,
            renderCell: (item) => (
                <span className="text-gray-300 text-sm" title={item.namespace}>
                    {item.namespace || '-'}
                </span>
            )
        }
    ].filter(col => {
        // Exclude namespace column for cluster-scoped resources
        if (col.id === 'namespace') {
            const clusterScoped = ['Node', 'PersistentVolume', 'StorageClass', 'ClusterRole', 'ClusterRoleBinding', 'Namespace'];
            if (clusterScoped.includes(kind)) return false;
        }
        return true;
    });
};

export const getAgeColumn = () => ({
    id: 'age',
    label: 'Age',
    width: 'minmax(80px, 0.6fr)',
    sortValue: (item) => parseDateValue(item.created),
    align: 'center',
    renderCell: (item) => <AgeCell created={item.created} />
});
