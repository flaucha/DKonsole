import React from 'react';
import { Network } from 'lucide-react';
import ResourceCell from '../../components/workloads/cells/ResourceCell';

export const serviceColumns = [
    {
        id: 'clusterIP',
        label: 'Cluster IP',
        width: 'minmax(140px, 1.2fr)',
        sortValue: (item) => item.details?.clusterIP || '',
        align: 'center',
        renderCell: (item) => (
            <ResourceCell
                value={item.details?.clusterIP}
                icon={Network}
                color="text-indigo-400"
            />
        )
    },
    {
        id: 'ports',
        label: 'Ports',
        width: 'minmax(180px, 1.5fr)',
        sortValue: (item) => (item.details?.ports || []).join(', '),
        align: 'left',
        renderCell: (item) => {
            const ports = item.details?.ports || [];
            if (ports.length === 0) return <span className="text-gray-500">-</span>;
            return (
                <div className="flex flex-wrap gap-1">
                    {ports.slice(0, 2).map((port, i) => {
                        // Parse format like "80:30080/TCP" or "80/TCP"
                        const parts = String(port || '').split('/');
                        const protocol = parts[1] || 'TCP';
                        const portDisplay = parts[0] || port;
                        return (
                            <span key={i} className="text-xs bg-gray-800 px-1.5 py-0.5 rounded text-gray-300 border border-gray-700">
                                {portDisplay}/{protocol}
                            </span>
                        );
                    })}
                    {ports.length > 2 && <span className="text-xs text-gray-500">+{ports.length - 2}</span>}
                </div>
            );
        }
    }
];
