import React from 'react';
import { useMemo } from 'react';
import {
    Cpu,
    Database,
    HardDrive,
    Network,
    ChevronDown
} from 'lucide-react';

import {
    parseDateValue,
    parseReadyRatio,
    parseCpuToMilli,
    parseMemoryToMi,
    parseSizeToMi
} from '../utils/workloadUtils';

import StatusCell from '../components/workloads/cells/StatusCell';
import AgeCell from '../components/workloads/cells/AgeCell';
import ResourceCell from '../components/workloads/cells/ResourceCell';

const useWorkloadColumns = (kind) => {
    // Memoize the dataColumns to prevent unnecessary re-renders
    const dataColumns = useMemo(() => {
        // Common columns for all workloads
        const baseColumns = [
            {
                id: 'name',
                label: 'Name',
                width: 'minmax(220px, 2.2fr)',
                sortValue: (item) => item.name || '',
                align: 'left',
                renderCell: (item, { isExpanded }) => (
                    <div className="font-semibold text-gray-200 flex items-center space-x-2 truncate">
                        <ChevronDown
                            size={16}
                            className={`text-gray-500 transition-transform duration-200 flex-shrink-0 ${isExpanded ? 'transform rotate-180' : ''}`}
                        />
                        <span className="truncate" title={item.name}>{item.name}</span>
                    </div>
                )
            },
            {
                id: 'status',
                label: 'Status',
                width: 'minmax(120px, 0.9fr)',
                sortValue: (item) => item.status || '',
                align: 'center',
                renderCell: (item) => <StatusCell status={item.status} kind={kind} />
            },
            {
                id: 'namespace',
                label: 'Namespace',
                width: 'minmax(140px, 1fr)',
                sortValue: (item) => item.namespace || '',
                align: 'center',
                hiddenByDefault: true,
                renderCell: (item) => (
                    <span className="text-gray-300 text-sm" title={item.namespace}>
                        {item.namespace || '-'}
                    </span>
                )
            }
        ].filter(col => {
            // Exclude namespace column for cluster-scoped resources
            if (col.id === 'namespace') {
                const clusterScoped = ['Node', 'PersistentVolume', 'StorageClass', 'ClusterRole', 'ClusterRoleBinding', 'Namespace'];
                if (clusterScoped.includes(kind)) return false;
            }
            return true;
        });

        // Kind-specific columns
        let specificColumns = [];

        switch (kind) {
            case 'Pod':
                specificColumns = [
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
                break;
            case 'Deployment':
                specificColumns = [
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
                        id: 'upToDate',
                        label: 'Up-to-Date',
                        width: 'minmax(110px, 0.8fr)',
                        sortValue: (item) => item.details?.updated || 0,
                        align: 'center',
                        renderCell: (item) => <span>{item.details?.updated || 0}</span>
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
                            const tag = typeof img === 'string' ? img.split(':').pop() : (img?.tag || '-');
                            return (
                                <ResourceCell
                                    value={tag}
                                    icon={null}
                                    color="text-gray-400"
                                />
                            );
                        }
                    }
                ];
                break;
            case 'Service':
                specificColumns = [
                    {
                        id: 'type',
                        label: 'Type',
                        width: 'minmax(120px, 1fr)',
                        sortValue: (item) => item.details?.type || '',
                        align: 'center',
                        renderCell: (item) => <span>{item.details?.type || '-'}</span>
                    },
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
                        sortValue: (item) => (item.details?.ports || []).map(p => `${p.port}/${p.protocol}`).join(', '),
                        align: 'left',
                        renderCell: (item) => {
                            const ports = item.details?.ports || [];
                            if (ports.length === 0) return <span className="text-gray-500">-</span>;
                            return (
                                <div className="flex flex-wrap gap-1">
                                    {ports.slice(0, 2).map((p, i) => (
                                        <span key={i} className="text-xs bg-gray-800 px-1.5 py-0.5 rounded text-gray-300 border border-gray-700">
                                            {p.port}/{p.protocol}
                                        </span>
                                    ))}
                                    {ports.length > 2 && <span className="text-xs text-gray-500">+{ports.length - 2}</span>}
                                </div>
                            );
                        }
                    }
                ];
                break;
            case 'Ingress':
                specificColumns = [
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
                        sortValue: (item) => (item.details?.loadBalancer?.ingress || []).map(i => i.ip || i.hostname).join(', '),
                        align: 'center',
                        renderCell: (item) => {
                            const ingress = item.details?.loadBalancer?.ingress || [];
                            if (ingress.length === 0) return <span className="text-gray-500">-</span>;
                            return <span className="text-xs">{ingress[0].ip || ingress[0].hostname}</span>;
                        }
                    }
                ];
                break;
            case 'Node':
                specificColumns = [
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
                break;
            case 'Job':
                specificColumns = [
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
                break;
            case 'CronJob':
                specificColumns = [
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
                break;
            case 'PersistentVolumeClaim':
                specificColumns = [
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
                        label: 'Size',
                        width: 'minmax(100px, 0.8fr)',
                        sortValue: (item) => parseSizeToMi(item.details?.resources?.requests?.storage || item.details?.capacity?.storage),
                        align: 'center',
                        renderCell: (item) => (
                            <ResourceCell
                                value={item.details?.resources?.requests?.storage || item.details?.capacity?.storage}
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
                break;
            case 'PersistentVolume':
                specificColumns = [
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
                        sortValue: (item) => parseSizeToMi(item.details?.capacity?.storage),
                        align: 'center',
                        renderCell: (item) => (
                            <ResourceCell
                                value={item.details?.capacity?.storage}
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
                        sortValue: (item) => item.details?.persistentVolumeReclaimPolicy || '',
                        align: 'center',
                        renderCell: (item) => <span>{item.details?.persistentVolumeReclaimPolicy}</span>
                    }
                ];
                break;
            case 'StorageClass':
                specificColumns = [
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
                break;
            case 'NetworkPolicy':
                specificColumns = [
                    {
                        id: 'podSelector',
                        label: 'Pod Selector',
                        width: 'minmax(160px, 1.5fr)',
                        sortValue: (item) => Object.entries(item.details?.podSelector?.matchLabels || {}).map(([k, v]) => `${k}=${v}`).join(', '),
                        align: 'left',
                        renderCell: (item) => {
                            const labels = item.details?.podSelector?.matchLabels || {};
                            if (Object.keys(labels).length === 0) return <span className="text-gray-500">All Pods</span>;
                            return (
                                <div className="flex flex-wrap gap-1">
                                    {Object.entries(labels).map(([key, value]) => (
                                        <span key={key} className="text-xs bg-gray-800 px-1.5 py-0.5 rounded text-gray-400 border border-gray-700">
                                            {key}={value}
                                        </span>
                                    ))}
                                </div>
                            );
                        }
                    },
                    {
                        id: 'types',
                        label: 'Policy Types',
                        width: 'minmax(140px, 1.2fr)',
                        sortValue: (item) => (item.details?.policyTypes || []).join(', '),
                        align: 'center',
                        renderCell: (item) => (
                            <ResourceCell
                                value={(item.details?.policyTypes || []).join(', ')}
                                icon={null}
                                color="text-blue-300"
                            />
                        )
                    },
                ];
                break;

            // Generic fallback
            default:
                break;
        }

        const ageColumn = {
            id: 'age',
            label: 'Age',
            width: 'minmax(80px, 0.6fr)',
            sortValue: (item) => parseDateValue(item.created),
            align: 'center',
            renderCell: (item) => <AgeCell created={item.created} />
        };

        return [...baseColumns, ...specificColumns, ageColumn];
    }, [kind]);

    return { dataColumns };
};

export default useWorkloadColumns;
