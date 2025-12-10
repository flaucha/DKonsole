import React from 'react';
import ResourceCell from '../../components/workloads/cells/ResourceCell';
import AgeCell from '../../components/workloads/cells/AgeCell';
import { parseDateValue } from '../../utils/workloadUtils';

export const cronJobColumns = [
    {
        id: 'schedule',
        label: 'Schedule',
        width: 'minmax(140px, 1.2fr)',
        sortValue: (item) => item.details?.schedule || '',
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={item.details?.schedule}
                icon={null}
                color="text-yellow-400 font-mono text-xs"
            />
        )
    },
    {
        id: 'suspend',
        label: 'Suspend',
        width: 'minmax(90px, 0.6fr)',
        sortValue: (item) => item.details?.suspend ? 1 : 0,
        align: 'center',
        renderCell: (item) => <span className={item.details?.suspend ? 'text-yellow-500' : 'text-gray-500'}>{item.details?.suspend?.toString()}</span>
    },
    {
        id: 'active',
        label: 'Active',
        width: 'minmax(80px, 0.5fr)',
        sortValue: (item) => (item.details?.active || []).length,
        align: 'center',
        renderCell: (item) => <span>{(item.details?.active || []).length}</span>
    },
    {
        id: 'lastSchedule',
        label: 'Last Run',
        width: 'minmax(120px, 1fr)',
        sortValue: (item) => parseDateValue(item.details?.lastScheduleTime),
        align: 'center',
        renderCell: (item) => <AgeCell created={item.details?.lastScheduleTime} />
    }
];
