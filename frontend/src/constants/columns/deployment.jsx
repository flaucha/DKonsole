import React from 'react';
import ResourceCell from '../../components/workloads/cells/ResourceCell';
import { parseReadyRatio } from '../../utils/workloadUtils';

export const deploymentColumns = [
    {
        id: 'ready',
        label: 'Ready',
        width: 'minmax(90px, 0.6fr)',
        sortValue: (item) => parseReadyRatio(`${item.details?.ready}/${item.details?.replicas}`),
        align: 'center',
        renderCell: (item) => (
            <span>
                {item.details?.ready || 0}/{item.details?.replicas || 0}
            </span>
        )
    },
    {
        id: 'upToDate',
        label: 'Up-to-Date',
        width: 'minmax(110px, 0.8fr)',
        sortValue: (item) => item.details?.updated || 0,
        align: 'center',
        renderCell: (item) => <span>{item.details?.updated || 0}</span>
    },
    {
        id: 'available',
        label: 'Available',
        width: 'minmax(100px, 0.7fr)',
        sortValue: (item) => item.details?.available || 0,
        align: 'center',
        renderCell: (item) => <span>{item.details?.available || 0}</span>
    },
    {
        id: 'tag',
        label: 'Tag',
        width: 'minmax(150px, 1.2fr)',
        sortValue: (item) => {
            const img = item.details?.images?.[0];
            return typeof img === 'string' ? img.split(':').pop() : (img?.tag || '');
        },
        align: 'center',
        renderCell: (item) => {
            const img = item.details?.images?.[0];
            const tag = typeof img === 'string' ? img.split(':').pop() : (img?.tag || '-');
            return (
                <ResourceCell
                    value={tag}
                    icon={null}
                    color="text-gray-400"
                />
            );
        }
    }
];
