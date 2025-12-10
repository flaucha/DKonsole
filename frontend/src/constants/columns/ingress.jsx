import React from 'react';
import ResourceCell from '../../components/workloads/cells/ResourceCell';

export const ingressColumns = [
    {
        id: 'hosts',
        label: 'Hosts',
        width: 'minmax(200px, 2fr)',
        sortValue: (item) => (item.details?.rules || []).map(r => r.host).join(', '),
        align: 'left',
        renderCell: (item) => {
            const rules = item.details?.rules || [];
            if (rules.length === 0) return <span className="text-gray-500">-</span>;
            const hosts = rules.map(r => r.host).filter(Boolean);
            if (hosts.length === 0) return <span className="text-gray-500">*</span>;
            return (
                <div className="flex flex-col">
                    {hosts.slice(0, 2).map((host, i) => (
                        <span key={i} className="text-xs text-blue-300 truncate max-w-[180px]" title={host}>{host}</span>
                    ))}
                    {hosts.length > 2 && <span className="text-xs text-gray-500">+{hosts.length - 2} more</span>}
                </div>
            );
        }
    },
    {
        id: 'address',
        label: 'Address',
        width: 'minmax(140px, 1.2fr)',
        sortValue: (item) => (item.details?.loadBalancer || []).map(i => i.ip || i.hostname).join(', '),
        align: 'center',
        renderCell: (item) => {
            const loadBalancer = item.details?.loadBalancer || [];
            if (loadBalancer.length === 0) return <span className="text-gray-500">-</span>;
            return <span className="text-xs">{loadBalancer[0].ip || loadBalancer[0].hostname}</span>;
        }
    }
];
