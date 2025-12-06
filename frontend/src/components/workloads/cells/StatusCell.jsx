import React from 'react';
import { getStatusBadgeClass } from '../../../utils/statusBadge';

const StatusCell = ({ status }) => (
    <span className={`px-2 py-1 rounded-full text-xs font-medium border ${getStatusBadgeClass(status)}`}>
        {status}
    </span>
);

export default StatusCell;
