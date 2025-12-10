import React from 'react';

export const configColumns = [
    {
        id: 'data',
        label: 'Data',
        width: 'minmax(90px, 0.7fr)',
        sortValue: (item) => {
            // Support both dataCount (number) and data (object)
            if (typeof item.details?.dataCount === 'number') return item.details.dataCount;
            if (typeof item.details?.data === 'object') return Object.keys(item.details.data || {}).length;
            return item.details?.data || 0;
        },
        align: 'center',
        renderCell: (item) => {
            // Support both dataCount (number) and data (object)
            let dataCount;
            if (typeof item.details?.dataCount === 'number') {
                dataCount = item.details.dataCount;
            } else if (typeof item.details?.data === 'object') {
                dataCount = Object.keys(item.details.data || {}).length;
            } else {
                dataCount = item.details?.data || 0;
            }
            return (
                <span className={dataCount > 0 ? 'text-gray-300' : 'text-gray-500'}>
                    {dataCount} {dataCount === 1 ? 'entry' : 'entries'}
                </span>
            );
        }
    }
];
