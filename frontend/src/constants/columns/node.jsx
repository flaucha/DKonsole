import React from 'react';
import { parseCpuToMilli, parseMemoryToMi } from '../../utils/workloadUtils';

export const nodeColumns = [
    {
        id: 'roles',
        label: 'Roles',
        width: 'minmax(120px, 1fr)',
        sortValue: (item) => Object.keys(item.details?.labels || {}).filter(l => l.includes('node-role')).join(', '),
        align: 'center',
        renderCell: (item) => {
            const isMaster = Object.keys(item.details?.labels || {}).some(l => l.includes('control-plane') || l.includes('master'));
            return <span className={`text-xs px-2 py-0.5 rounded-full ${isMaster ? 'bg-purple-900/50 text-purple-300 border border-purple-700' : 'bg-gray-800 text-gray-300'}`}>{isMaster ? 'Control Plane' : 'Worker'}</span>;
        }
    },
    {
        id: 'version',
        label: 'Version',
        width: 'minmax(100px, 0.8fr)',
        sortValue: (item) => item.details?.nodeInfo?.kubeletVersion || '',
        align: 'center',
        renderCell: (item) => <span>{item.details?.nodeInfo?.kubeletVersion}</span>
    },
    {
        id: 'cpu',
        label: 'CPU',
        width: 'minmax(100px, 0.8fr)',
        sortValue: (item) => parseCpuToMilli(item.details?.capacity?.cpu),
        align: 'center',
        renderCell: (item) => <span>{item.details?.capacity?.cpu || '-'}</span>
    },
    {
        id: 'memory',
        label: 'Memory',
        width: 'minmax(100px, 0.8fr)',
        sortValue: (item) => parseMemoryToMi(item.details?.capacity?.memory),
        align: 'center',
        renderCell: (item) => <span>{item.details?.capacity?.memory || '-'}</span>
    }
];
