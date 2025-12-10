import React from 'react';
import { HardDrive } from 'lucide-react';
import ResourceCell from '../../components/workloads/cells/ResourceCell';
import { parseSizeToMi } from '../../utils/workloadUtils';

export const persistentVolumeClaimColumns = [
    {
        id: 'storageClass',
        label: 'Storage Class',
        width: 'minmax(140px, 1.2fr)',
        sortValue: (item) => item.details?.storageClassName || '',
        align: 'center',
        renderCell: (item) => <span>{item.details?.storageClassName || '-'}</span>
    },
    {
        id: 'size',
        label: 'Capacity',
        width: 'minmax(100px, 0.8fr)',
        sortValue: (item) => parseSizeToMi(item.details?.capacity),
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={item.details?.capacity}
                icon={HardDrive}
                color="text-emerald-400"
            />
        )
    },
    {
        id: 'accessModes',
        label: 'Access Modes',
        width: 'minmax(140px, 1.2fr)',
        sortValue: (item) => (item.details?.accessModes || []).join(', '),
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={(item.details?.accessModes || []).join(', ')}
                icon={null}
                color="text-gray-400"
            />
        )
    }
];

export const persistentVolumeColumns = [
    {
        id: 'storageClass',
        label: 'Storage Class',
        width: 'minmax(140px, 1.2fr)',
        sortValue: (item) => item.details?.storageClassName || '',
        align: 'center',
        renderCell: (item) => <span>{item.details?.storageClassName || '-'}</span>
    },
    {
        id: 'capacity',
        label: 'Capacity',
        width: 'minmax(100px, 0.8fr)',
        sortValue: (item) => parseSizeToMi(item.details?.capacity),
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={item.details?.capacity}
                icon={HardDrive}
                color="text-emerald-400"
            />
        )
    },
    {
        id: 'accessModes',
        label: 'Access Modes',
        width: 'minmax(140px, 1.2fr)',
        sortValue: (item) => (item.details?.accessModes || []).join(', '),
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={(item.details?.accessModes || []).join(', ')}
                icon={null}
                color="text-gray-400"
            />
        )
    },
    {
        id: 'reclaim',
        label: 'Reclaim',
        width: 'minmax(110px, 0.8fr)',
        sortValue: (item) => item.details?.reclaimPolicy || '',
        align: 'center',
        renderCell: (item) => <span>{item.details?.reclaimPolicy || '-'}</span>
    }
];

export const storageClassColumns = [
    {
        id: 'provisioner',
        label: 'Provisioner',
        width: 'minmax(200px, 2fr)',
        sortValue: (item) => item.details?.provisioner || '',
        align: 'left',
        renderCell: (item) => (
            <ResourceCell
                value={item.details?.provisioner}
                icon={HardDrive}
                color="text-gray-300"
            />
        )
    },
    {
        id: 'reclaimPolicy',
        label: 'Reclaim Policy',
        width: 'minmax(140px, 1.2fr)',
        sortValue: (item) => item.details?.reclaimPolicy || '',
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={item.details?.reclaimPolicy}
                icon={null}
                color="text-gray-400"
            />
        )
    }
];
