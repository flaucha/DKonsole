import React from 'react';
import { Cpu, Database } from 'lucide-react';
import { parseReadyRatio, parseCpuToMilli, parseMemoryToMi } from '../../utils/workloadUtils';
import ResourceCell from '../../components/workloads/cells/ResourceCell';

export const podColumns = [
    {
        id: 'ready',
        label: 'Ready',
        width: 'minmax(90px, 0.8fr)',
        sortValue: (item) => parseReadyRatio(item.details?.ready),
        align: 'center',
        renderCell: (item) => <span>{item.details?.ready || '-'}</span>
    },
    {
        id: 'restarts',
        label: 'Restarts',
        width: 'minmax(90px, 0.6fr)',
        sortValue: (item) => item.details?.restarts || 0,
        align: 'center',
        renderCell: (item) => <span>{item.details?.restarts || 0}</span>
    },
    {
        id: 'cpu',
        label: 'CPU',
        width: 'minmax(110px, 0.9fr)',
        sortValue: (item) => parseCpuToMilli(item.details?.metrics?.cpu),
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={item.details?.metrics?.cpu}
                icon={Cpu}
                color="text-amber-400"
            />
        )
    },
    {
        id: 'memory',
        label: 'Memory',
        width: 'minmax(110px, 0.9fr)',
        sortValue: (item) => parseMemoryToMi(item.details?.metrics?.memory),
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={item.details?.metrics?.memory}
                icon={Database}
                color="text-blue-400"
            />
        )
    }
];
