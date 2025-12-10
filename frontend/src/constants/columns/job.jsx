import React from 'react';
import ResourceCell from '../../components/workloads/cells/ResourceCell';

export const jobColumns = [
    {
        id: 'completions',
        label: 'Completions',
        width: 'minmax(100px, 0.8fr)',
        sortValue: (item) => `${item.details?.succeeded || 0}/${item.details?.completions || 1}`,
        align: 'center',
        renderCell: (item) => <span>{item.details?.succeeded || 0}/{item.details?.completions || 1}</span>
    },
    {
        id: 'duration',
        label: 'Duration',
        width: 'minmax(100px, 0.8fr)',
        sortValue: (item) => {
            if (!item.details?.startTime) return 0;
            const start = new Date(item.details.startTime).getTime();
            const end = item.details.completionTime ? new Date(item.details.completionTime).getTime() : Date.now();
            return end - start;
        },
        align: 'center',
        renderCell: (item) => {
            if (!item.details?.startTime) return '-';
            const start = new Date(item.details.startTime);
            const end = item.details.completionTime ? new Date(item.details.completionTime) : new Date();
            const diff = Math.max(0, Math.floor((end - start) / 1000));

            if (diff < 60) return `${diff}s`;
            if (diff < 3600) return `${Math.floor(diff / 60)}m`;
            return `${Math.floor(diff / 3600)}h`;
        }
    },
    {
        id: 'succeeded',
        label: 'Succeeded',
        width: 'minmax(90px, 0.6fr)',
        sortValue: (item) => item.details?.succeeded || 0,
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={item.details?.succeeded || 0}
                icon={null}
                color="text-green-400"
            />
        )
    },
    {
        id: 'failed',
        label: 'Failed',
        width: 'minmax(90px, 0.6fr)',
        sortValue: (item) => item.details?.failed || 0,
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={item.details?.failed || 0}
                icon={null}
                color={item.details?.failed > 0 ? "text-red-400 font-bold" : "text-gray-500"}
            />
        )
    }
];
