import React from 'react';
import ResourceCell from '../../components/workloads/cells/ResourceCell';
import { SmartImage } from '../../components/details/CommonDetails';
import { parseReadyRatio } from '../../utils/workloadUtils';

export const replicaSetColumns = [
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
            if (!img) return <span className="text-gray-400">-</span>;
            return (
                <div className="flex justify-center text-sm text-gray-400">
                    <SmartImage image={typeof img === 'string' ? img : (img?.image || '')} />
                </div>
            );
        }
    }
];
