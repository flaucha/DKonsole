import React from 'react';
import ResourceCell from '../../components/workloads/cells/ResourceCell';

export const networkPolicyColumns = [
    {
        id: 'podSelector',
        label: 'Pod Selector',
        width: 'minmax(160px, 1.5fr)',
        sortValue: (item) => Object.entries(item.details?.podSelector?.matchLabels || {}).map(([k, v]) => `${k}=${v}`).join(', '),
        align: 'left',
        renderCell: (item) => {
            const labels = item.details?.podSelector?.matchLabels || {};
            if (Object.keys(labels).length === 0) return <span className="text-gray-500">All Pods</span>;
            return (
                <div className="flex flex-wrap gap-1">
                    {Object.entries(labels).map(([key, value]) => (
                        <span key={key} className="text-xs bg-gray-800 px-1.5 py-0.5 rounded text-gray-400 border border-gray-700">
                            {key}={value}
                        </span>
                    ))}
                </div>
            );
        }
    },
    {
        id: 'types',
        label: 'Policy Types',
        width: 'minmax(140px, 1.2fr)',
        sortValue: (item) => (item.details?.policyTypes || []).join(', '),
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={(item.details?.policyTypes || []).join(', ')}
                icon={null}
                color="text-blue-300"
            />
        )
    },
];
