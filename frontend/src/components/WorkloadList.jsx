import React, { useState, useEffect, useRef } from 'react';
import {
    Box,
    FileText,
    Key,
    Clock,
    Layers,
    Activity,
    Network,
    Globe,
    ChevronDown,
    Server,
    HardDrive,
    Shield,
    Lock,
    Users,
    Search,
    RefreshCw,
    MoreVertical,
    PlayCircle,
    X
} from 'lucide-react';
import { useLocation, Link } from 'react-router-dom';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { useWorkloads } from '../hooks/useWorkloads';
import { formatDateTime } from '../utils/dateUtils';
import { getExpandableRowClasses, getExpandableRowRowClasses } from '../utils/expandableRow';
import { getStatusBadgeClass } from '../utils/statusBadge';
import { canEdit, isAdmin } from '../utils/permissions';

// Import extracted detail components
import NodeDetails from './details/NodeDetails';
import { ServiceAccountDetails, RoleDetails, BindingDetails } from './details/RbacDetails';
import DeploymentDetails from './details/DeploymentDetails';
import ServiceDetails from './details/ServiceDetails';
import IngressDetails from './details/IngressDetails';
import PodDetails from './details/PodDetails';
import { ConfigMapDetails, SecretDetails } from './details/ConfigDetails';
import NetworkPolicyDetails from './details/NetworkPolicyDetails';
import StorageDetails from './details/StorageDetails';
import StorageClassDetails from './details/StorageClassDetails';
import { JobDetails, CronJobDetails, StatefulSetDetails, DaemonSetDetails, HPADetails } from './details/WorkloadDetails';
import GenericDetails from './details/GenericDetails';
import { EditYamlButton } from './details/CommonDetails';
import YamlEditor from './YamlEditor';

// Map resource kind to an icon component
const getIcon = (kind) => {
    switch (kind) {
        case 'Deployment':
        case 'Pod':
            return Box;
        case 'Node':
            return Server;
        case 'ServiceAccount':
            return Shield;
        case 'Role':
        case 'ClusterRole':
            return Lock;
        case 'RoleBinding':
        case 'ClusterRoleBinding':
            return Users;
        case 'ConfigMap':
            return FileText;
        case 'Secret':
            return Key;
        case 'Job':
        case 'CronJob':
            return Clock;
        case 'StatefulSet':
        case 'DaemonSet':
            return Layers;
        case 'HPA':
            return Activity;
        case 'Service':
            return Network;
        case 'Ingress':
            return Globe;
        case 'NetworkPolicy':
            return Activity;
        case 'PersistentVolumeClaim':
        case 'PersistentVolume':
            return HardDrive;
        case 'StorageClass':
            return HardDrive;
        default:
            return Box;
    }
};

const WorkloadList = ({ namespace, kind }) => {
    const [expandedId, setExpandedId] = useState(null);
    const [sortField, setSortField] = useState('name');
    const [sortDirection, setSortDirection] = useState('asc');
    const { currentCluster } = useSettings();
    const { authFetch, user } = useAuth();
    const location = useLocation();
    const [filter, setFilter] = useState('');
    const [isSearchFocused, setIsSearchFocused] = useState(false);
    const [editingResource, setEditingResource] = useState(null);
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [confirmRollout, setConfirmRollout] = useState(null);
    const [allResources, setAllResources] = useState([]);
    const [scaling, setScaling] = useState(null);
    const [triggering, setTriggering] = useState(null);
    const [createdJob, setCreatedJob] = useState(null);
    const [rollingOut, setRollingOut] = useState(null);

    // Early return if kind is not provided
    if (!kind) {
        return <div className="text-red-400 p-6">Error: Resource type not specified.</div>;
    }

    // Use React Query hook to fetch all resources
    const { data: resourcesData, isLoading: loading, error, refetch } = useWorkloads(authFetch, namespace, kind, currentCluster);

    // Track previous kind to detect changes
    const prevKindRef = useRef(kind);
    const prevNamespaceRef = useRef(namespace);
    const prevClusterRef = useRef(currentCluster);

    // Reset when namespace, kind, or cluster changes - MUST happen before data handling
    useEffect(() => {
        const kindChanged = prevKindRef.current !== kind;
        const namespaceChanged = prevNamespaceRef.current !== namespace;
        const clusterChanged = prevClusterRef.current !== currentCluster;

        if (kindChanged || namespaceChanged || clusterChanged) {
            // Clear state immediately when switching
            setAllResources([]);
            setExpandedId(null);
            setFilter('');

            // Update refs
            prevKindRef.current = kind;
            prevNamespaceRef.current = namespace;
            prevClusterRef.current = currentCluster;

            // Force refetch when kind/namespace/cluster changes to ensure fresh data
            if (namespace && kind) {
                refetch();
            }
        }
    }, [namespace, kind, currentCluster, refetch]);

    // Handle resources data - always expect an array now (no pagination)
    useEffect(() => {
        // Only update if we have data
        if (resourcesData) {
            let data = [];
            if (Array.isArray(resourcesData)) {
                data = resourcesData;
            } else if (resourcesData.resources && Array.isArray(resourcesData.resources)) {
                // Backward compatibility with paginated response
                data = resourcesData.resources;
            }
            setAllResources(data);

            // Check for search param in URL
            const searchParams = new URLSearchParams(location.search);
            const search = searchParams.get('search');
            if (search) {
                setFilter(search);
                // If exact match found, expand it
                const found = data.find(r => r.name === search);
                if (found) {
                    setExpandedId(found.uid);
                }
            }
        } else if (!loading) {
            // Only clear if we're not loading (to avoid flickering)
            setAllResources([]);
        }
    }, [resourcesData, loading, location.search]);

    // Ensure resources is always an array
    const resources = allResources;

    const toggleExpand = (uid) => {
        setExpandedId(current => current === uid ? null : uid);
    };

    const handleSort = (field) => {
        setSortField((prevField) => {
            if (prevField === field) {
                setSortDirection((prevDir) => (prevDir === 'asc' ? 'desc' : 'asc'));
                return prevField;
            }
            setSortDirection('asc');
            return field;
        });
    };

    const filteredResources = resources.filter(r => {
        if (!filter) return true;
        return r.name.toLowerCase().includes(filter.toLowerCase());
    });

    const sortedResources = [...filteredResources].sort((a, b) => {
        const dir = sortDirection === 'asc' ? 1 : -1;
        const getVal = (item) => {
            switch (sortField) {
                case 'name':
                    return item.name || '';
                case 'kind':
                    return item.kind || '';
                case 'status':
                    return item.status || '';
                case 'created':
                    return new Date(item.created).getTime() || 0;
                case 'cpu': {
                    if (kind !== 'Pod' || !item.details?.metrics?.cpu) return 0;
                    const cpuStr = item.details.metrics.cpu.trim();
                    if (cpuStr.endsWith('m')) return parseFloat(cpuStr.replace('m', '')) || 0;
                    const val = parseFloat(cpuStr);
                    return isNaN(val) ? 0 : val * 1000;
                }
                case 'memory': {
                    if (kind !== 'Pod' || !item.details?.metrics?.memory) return 0;
                    const memStr = item.details.metrics.memory.toUpperCase().trim();
                    const num = parseFloat(memStr);
                    if (isNaN(num)) return 0;
                    if (memStr.includes('GI')) return num * 1024;
                    if (memStr.includes('MI')) return num;
                    if (memStr.includes('KI')) return num / 1024;
                    return num;
                }
                case 'size': {
                    if (kind !== 'PersistentVolumeClaim') return 0;
                    const sizeStr = (item.details?.capacity || item.details?.requested || '').toUpperCase().trim();
                    if (!sizeStr || sizeStr === '—') return 0;
                    const num = parseFloat(sizeStr);
                    if (isNaN(num)) return 0;
                    if (sizeStr.includes('TI') || sizeStr.includes('T')) return num * 1024 * 1024;
                    if (sizeStr.includes('GI') || sizeStr.includes('G')) return num * 1024;
                    if (sizeStr.includes('MI') || sizeStr.includes('M')) return num;
                    if (sizeStr.includes('KI') || sizeStr.includes('K')) return num / 1024;
                    return num / (1024 * 1024);
                }
                case 'ready': {
                    if (kind !== 'Pod' || !item.details?.ready) return -1;
                    const readyStr = item.details.ready.toString();
                    const parts = readyStr.split('/');
                    if (parts.length === 2) {
                        const ready = parseFloat(parts[0]) || 0;
                        const total = parseFloat(parts[1]) || 0;
                        return total > 0 ? (ready / total) : -1;
                    }
                    return -1;
                }
                case 'restarts': {
                    if (kind !== 'Pod') return 0;
                    return item.details?.restarts || 0;
                }
                default:
                    return '';
            }
        };
        const va = getVal(a);
        const vb = getVal(b);
        if (typeof va === 'number' && typeof vb === 'number') {
            return (va - vb) * dir;
        }
        return String(va).localeCompare(String(vb)) * dir;
    });

    // No display limit - show all resources
    const limitedResources = sortedResources; // Alias for clarity in code readability

    const renderSortIndicator = (field) => {
        if (sortField !== field) return null;
        return sortDirection === 'asc' ? '↑' : '↓';
    };

    // Helper function to check if a node is a master/control-plane node
    const isMasterNode = (res) => {
        if (res.kind !== 'Node') return false;

        // Check details for labels or taints that indicate control plane
        if (res.details) {
            // Check labels if available
            if (res.details.labels) {
                const labels = res.details.labels;
                if ((labels['node-role.kubernetes.io/control-plane'] &&
                     labels['node-role.kubernetes.io/control-plane'] !== '' &&
                     labels['node-role.kubernetes.io/control-plane'] !== 'false') ||
                    (labels['node-role.kubernetes.io/master'] &&
                     labels['node-role.kubernetes.io/master'] !== '' &&
                     labels['node-role.kubernetes.io/master'] !== 'false')) {
                    return true;
                }
            }

            // Check taints if available
            if (res.details.taints && Array.isArray(res.details.taints)) {
                for (const taint of res.details.taints) {
                    const taintKey = taint.key || (typeof taint === 'string' ? taint : null);
                    if (taintKey && (
                        taintKey === 'node-role.kubernetes.io/control-plane' ||
                        taintKey === 'node-role.kubernetes.io/master')) {
                        return true;
                    }
                }
            }
        }

        return false;
    };

    const handleDelete = async (res, force = false) => {
        const params = new URLSearchParams({ kind: res.kind, name: res.name });
        if (res.namespace) params.append('namespace', res.namespace);
        if (currentCluster) params.append('cluster', currentCluster);
        if (force) params.append('force', 'true');

        try {
            const response = await authFetch(`/api/resource?${params.toString()}`, { method: 'DELETE' });
            if (!response.ok) {
                const errorText = await response.text().catch(() => 'Delete failed');
                throw new Error(errorText || 'Delete failed');
            }
            refetch();
        } catch (err) {
            alert(err.message || 'Failed to delete resource');
        } finally {
            setConfirmAction(null);
        }
    };

    const handleScale = async (res, delta) => {
        if (!res.namespace) return;
        setScaling(res.name);
        const params = new URLSearchParams({
            kind: 'Deployment',
            name: res.name,
            namespace: res.namespace,
            delta: String(delta),
        });
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const resp = await authFetch(`/api/scale?${params.toString()}`, { method: 'POST' });
            if (!resp.ok) {
                const errorText = await resp.text().catch(() => 'Scale failed');
                throw new Error(errorText || 'Scale failed');
            }
            refetch();
        } catch (err) {
            alert(err.message || 'Failed to scale deployment');
        } finally {
            setScaling(null);
        }
    };

    const handleTriggerCronJob = async (res) => {
        if (!res.namespace) return;
        setTriggering(res.name);
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const resp = await authFetch(`/api/cronjobs/trigger?${params.toString()}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    namespace: res.namespace,
                    name: res.name
                })
            });
            if (!resp.ok) {
                const errorText = await resp.text().catch(() => 'Trigger failed');
                throw new Error(errorText || 'Trigger failed');
            }
            const data = await resp.json();
            if (data.jobName) {
                setCreatedJob({
                    name: data.jobName,
                    namespace: res.namespace
                });
            }
        } catch (err) {
            alert(err.message || 'Failed to trigger cronjob');
        } finally {
            setTriggering(null);
        }
    };

    const handleRolloutDeployment = async (res) => {
        if (!res.namespace) return;
        setRollingOut(res.name);
        setConfirmRollout(null);
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const resp = await authFetch(`/api/deployments/rollout?${params.toString()}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    namespace: res.namespace,
                    name: res.name
                })
            });
            if (!resp.ok) {
                const errorText = await resp.text().catch(() => 'Rollout failed');
                throw new Error(errorText || 'Rollout failed');
            }
            // Refresh the list to show updated deployment
            refetch();
        } catch (err) {
            alert(err.message || 'Failed to rollout deployment');
        } finally {
            setRollingOut(null);
        }
    };

    const renderDetails = (res) => {
        const onEditYAML = () => setEditingResource(res);
        const onDataSaved = () => {
            // Close expanded row and refresh data
            setExpandedId(null);
            setTimeout(() => refetch(), 300);
        };
        if (!res.details) return (
            <div className="p-4 text-gray-500 italic">
                No details available.
            </div>
        );
        switch (res.kind) {
            case 'Node': return <NodeDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'ServiceAccount': return <ServiceAccountDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'Role':
            case 'ClusterRole': return <RoleDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'RoleBinding':
            case 'ClusterRoleBinding': return <BindingDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'Deployment': return <DeploymentDetails details={res.details} onScale={(delta) => handleScale(res, delta)} scaling={scaling === res.name} res={res} onEditYAML={onEditYAML} />;
            case 'Service': return <ServiceDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} name={res.name} />;
            case 'Ingress': return <IngressDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} />;
            case 'Pod': return <PodDetails details={res.details} onEditYAML={onEditYAML} pod={res} />;
            case 'ConfigMap': return <ConfigMapDetails details={res.details} onEditYAML={onEditYAML} resource={res} onDataSaved={onDataSaved} />;
            case 'Secret': return <SecretDetails details={res.details} onEditYAML={onEditYAML} resource={res} onDataSaved={onDataSaved} />;
            case 'NetworkPolicy': return <NetworkPolicyDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'PersistentVolumeClaim':
            case 'PersistentVolume': return <StorageDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'StorageClass': return <StorageClassDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'Job': return <JobDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} />;
            case 'CronJob': return <CronJobDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'StatefulSet': return <StatefulSetDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} />;
            case 'DaemonSet': return <DaemonSetDetails details={res.details} onEditYAML={onEditYAML} namespace={res.namespace} />;
            case 'HPA': return <HPADetails details={res.details} onEditYAML={onEditYAML} />;
            default: return <GenericDetails details={res.details} onEditYAML={onEditYAML} />;
        }
    };

    const Icon = getIcon(kind);

    // Show loading state
    if (loading && resources.length === 0) {
        return <div className="text-gray-400 animate-pulse p-6">Loading {kind}s...</div>;
    }

    // Show error state
    if (error) {
        return (
            <div className="flex flex-col h-full">
                <div className="flex items-center justify-between p-4 border-b border-gray-800 bg-gray-900/50">
                    <div className="flex items-center space-x-4 flex-1">
                        <span className="text-sm text-gray-500">0 items</span>
                    </div>
                    <div className="flex items-center space-x-2">
                        <button
                            onClick={() => refetch()}
                            className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-md transition-colors"
                            title="Refresh"
                        >
                            <RefreshCw size={16} />
                        </button>
                    </div>
                </div>
                <div className="flex-1 flex items-center justify-center">
                    <div className="text-red-400 text-center">
                        <p className="text-lg font-semibold mb-2">Error loading {kind}s</p>
                        <p className="text-sm text-gray-500">{error.message}</p>
                    </div>
                </div>
            </div>
        );
    }

    // Show empty state if no resources and no filter (and not loading)
    if (!loading && resources.length === 0 && !filter) {
        return (
            <div className="flex flex-col h-full">
                {/* Toolbar */}
                <div className="flex items-center justify-between p-4 border-b border-gray-800 bg-gray-900/50">
                    <div className="flex items-center space-x-4 flex-1">
                        <div className={`relative transition-all duration-300 ${isSearchFocused ? 'w-96' : 'w-64'}`}>
                            <Search className={`absolute left-3 top-1/2 transform -translate-y-1/2 transition-colors duration-300 ${isSearchFocused ? 'text-blue-400' : 'text-gray-500'}`} size={16} />
                            <input
                                type="text"
                                placeholder={`Filter ${kind}s...`}
                                value={filter}
                                onChange={(e) => setFilter(e.target.value)}
                                onFocus={() => setIsSearchFocused(true)}
                                onBlur={() => setIsSearchFocused(false)}
                                className="w-full bg-gray-900 border border-gray-700 text-gray-200 text-sm rounded-md pl-10 pr-4 py-2 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all duration-300"
                            />
                        </div>
                        <span className="text-sm text-gray-500">
                            0 items
                        </span>
                    </div>
                    <div className="flex items-center space-x-2">
                        <button
                            onClick={() => refetch()}
                            className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-md transition-colors"
                            title="Refresh"
                        >
                            <RefreshCw size={16} />
                        </button>
                    </div>
                </div>
                {/* Empty state message */}
                <div className="flex-1 flex items-center justify-center">
                    <div className="text-gray-500 italic text-center">
                        <Icon size={48} className="mx-auto mb-4 text-gray-600" />
                        <p className="text-lg">No {kind}s found in this namespace.</p>
                        <p className="text-sm mt-2">Try selecting a different namespace or check if resources exist.</p>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full">
            {/* Toolbar */}
            <div className="flex items-center justify-between p-4 border-b border-gray-800 bg-gray-900/50">
                <div className="flex items-center space-x-4 flex-1">
                    <div className={`relative transition-all duration-300 ${isSearchFocused ? 'w-96' : 'w-64'}`}>
                        <Search className={`absolute left-3 top-1/2 transform -translate-y-1/2 transition-colors duration-300 ${isSearchFocused ? 'text-blue-400' : 'text-gray-500'}`} size={16} />
                        <input
                            type="text"
                            placeholder={`Filter ${kind}s...`}
                            value={filter}
                            onChange={(e) => setFilter(e.target.value)}
                            onFocus={() => setIsSearchFocused(true)}
                            onBlur={() => setIsSearchFocused(false)}
                            className="w-full bg-gray-900 border border-gray-700 text-gray-200 text-sm rounded-md pl-10 pr-10 py-2 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all duration-300"
                        />
                        {filter && (
                            <button
                                onClick={() => setFilter('')}
                                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-200 transition-colors"
                                type="button"
                            >
                                <X size={16} />
                            </button>
                        )}
                    </div>
                    <span className="text-sm text-gray-500">
                        {limitedResources.length} {limitedResources.length === 1 ? 'item' : 'items'}
                    </span>
                </div>
                <div className="flex items-center space-x-2">
                    <button
                        onClick={() => refetch()}
                        className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-md transition-colors"
                        title="Refresh"
                    >
                        <RefreshCw size={16} />
                    </button>
                </div>
            </div>

            {/* Table Header */}
            <div className="grid grid-cols-12 gap-4 px-6 py-3 border-b border-gray-800 bg-gray-900/50 text-xs font-medium text-gray-500 uppercase tracking-wider">
                <div className="col-span-4 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('name')}>
                    Name {renderSortIndicator('name')}
                </div>
                <div className="col-span-2 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('status')}>
                    Status {renderSortIndicator('status')}
                </div>
                <div className={`${kind === 'Pod' ? 'col-span-2' : 'col-span-3'} cursor-pointer hover:text-gray-300 flex items-center`} onClick={() => handleSort('created')}>
                    Age {renderSortIndicator('created')}
                </div>
                {kind === 'Pod' && (
                    <>
                        <div className="col-span-1 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('restarts')}>
                            Restarts {renderSortIndicator('restarts')}
                        </div>
                        <div className="col-span-1 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('cpu')}>
                            CPU {renderSortIndicator('cpu')}
                        </div>
                        <div className="col-span-1 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('memory')}>
                            Mem {renderSortIndicator('memory')}
                        </div>
                    </>
                )}
                {kind === 'PersistentVolumeClaim' && (
                    <div className="col-span-1 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => handleSort('size')}>
                        Size {renderSortIndicator('size')}
                    </div>
                )}
                <div className="col-span-1"></div>
            </div>

            {/* Table Body */}
            <div className="flex-1 overflow-y-auto">
                {limitedResources.map((res) => {
                    const isExpanded = expandedId === res.uid;
                    return (
                        <div key={res.uid} className="border-b border-gray-800 last:border-0">
                            <div
                                onClick={() => toggleExpand(res.uid)}
                                className={`grid grid-cols-12 gap-4 px-6 py-4 cursor-pointer transition-colors duration-200 items-center ${getExpandableRowRowClasses(isExpanded)}`}
                            >
                                <div className="col-span-4 flex items-center font-medium text-sm text-gray-200">
                                    <ChevronDown
                                        size={16}
                                        className={`mr-2 text-gray-500 transition-transform duration-200 ${isExpanded ? 'transform rotate-180' : ''}`}
                                    />
                                    <Icon size={16} className="mr-3 text-gray-500" />
                                    <span className="truncate" title={res.name}>{res.name}</span>
                                </div>
                                <div className="col-span-2">
                                    <span className={`px-2 py-1 rounded-full text-xs font-medium border ${getStatusBadgeClass(res.status)}`}>
                                        {res.status}
                                    </span>
                                </div>
                                <div className={`${kind === 'Pod' ? 'col-span-2' : 'col-span-3'} text-sm text-gray-400`}>
                                    {formatDateTime(res.created)}
                                </div>
                                {kind === 'Pod' && (
                                    <>
                                        <div className="col-span-1 text-sm text-gray-400">
                                            {res.details?.restarts || 0}
                                        </div>
                                        <div className="col-span-1 text-sm text-gray-400">
                                            {res.details?.metrics?.cpu || '-'}
                                        </div>
                                        <div className="col-span-1 text-sm text-gray-400">
                                            {res.details?.metrics?.memory || '-'}
                                        </div>
                                    </>
                                )}
                                {kind === 'PersistentVolumeClaim' && (
                                    <div className="col-span-1 text-sm text-gray-400">
                                        {res.details?.capacity || res.details?.requested || '-'}
                                    </div>
                                )}
                                {kind !== 'Pod' && kind !== 'PersistentVolumeClaim' && (
                                    <div className="col-span-1"></div>
                                )}
                                <div className={`${kind === 'Pod' ? 'col-span-1' : kind === 'PersistentVolumeClaim' ? 'col-span-2' : 'col-span-2'} flex justify-end items-center space-x-2 pr-2`} onClick={(e) => e.stopPropagation()}>
                                    {kind === 'CronJob' && (
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleTriggerCronJob(res);
                                            }}
                                            disabled={triggering === res.name}
                                            className="p-1 hover:bg-gray-800 rounded text-gray-400 hover:text-green-400 transition-colors disabled:opacity-50"
                                            title="Trigger manual run"
                                        >
                                            <PlayCircle size={16} />
                                        </button>
                                    )}
                                    {kind === 'Deployment' && (isAdmin(user) || canEdit(user, res.namespace)) && (
                                        <button
                                            onClick={async (e) => {
                                                e.stopPropagation();
                                                // Fetch deployment details with strategy information
                                                try {
                                                    const params = new URLSearchParams();
                                                    if (currentCluster) params.append('cluster', currentCluster);
                                                    const response = await authFetch(
                                                        `/api/namespaces/${res.namespace}/Deployment/${res.name}?${params.toString()}`
                                                    );
                                                    if (response.ok) {
                                                        const deploymentData = await response.json();
                                                        // Extract strategy information from raw deployment
                                                        let strategyInfo = null;
                                                        if (deploymentData.raw && deploymentData.raw.spec) {
                                                            const spec = deploymentData.raw.spec;
                                                            const strategy = spec.strategy || {};
                                                            const strategyType = strategy.type || 'RollingUpdate';

                                                            if (strategyType === 'RollingUpdate') {
                                                                const rollingUpdate = strategy.rollingUpdate || {};
                                                                strategyInfo = {
                                                                    type: 'RollingUpdate',
                                                                    maxSurge: rollingUpdate.maxSurge || '25%',
                                                                    maxUnavailable: rollingUpdate.maxUnavailable || '25%'
                                                                };
                                                            } else {
                                                                strategyInfo = {
                                                                    type: 'Recreate'
                                                                };
                                                            }
                                                        }
                                                        setConfirmRollout({
                                                            ...res,
                                                            details: deploymentData.details || res.details,
                                                            strategy: strategyInfo
                                                        });
                                                    } else {
                                                        setConfirmRollout(res);
                                                    }
                                                } catch (err) {
                                                    setConfirmRollout(res);
                                                }
                                            }}
                                            disabled={rollingOut === res.name}
                                            className="p-1 hover:bg-gray-800 rounded text-gray-400 hover:text-green-400 transition-colors disabled:opacity-50"
                                            title="Rollout deployment"
                                        >
                                            <RefreshCw size={16} className={rollingOut === res.name ? 'animate-spin' : ''} />
                                        </button>
                                    )}
                                    <div className="relative">
                                        <button
                                            onClick={() => setMenuOpen(menuOpen === res.uid ? null : res.uid)}
                                            className="p-1 hover:bg-gray-800 rounded text-gray-400 hover:text-white transition-colors"
                                        >
                                            <MoreVertical size={16} />
                                        </button>
                                        {menuOpen === res.uid && (
                                            <div className="absolute right-0 mt-1 w-36 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
                                                <div className="flex flex-col">
                                                    {/* Special handling for nodes */}
                                                    {res.kind === 'Node' && isMasterNode(res) ? (
                                                        <div className="px-4 py-2 text-xs text-red-400">
                                                            Not Allowed
                                                        </div>
                                                    ) : (isAdmin(user) || canEdit(user, res.namespace)) ? (
                                                        <>
                                                            <button
                                                                onClick={() => {
                                                                    // For nodes, show confirmation dialog
                                                                    if (res.kind === 'Node') {
                                                                        setConfirmAction({ res, force: false });
                                                                    } else {
                                                                        setConfirmAction({ res, force: false });
                                                                    }
                                                                    setMenuOpen(null);
                                                                }}
                                                                className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                            >
                                                                Delete
                                                            </button>
                                                            {res.kind !== 'Node' && (
                                                                <button
                                                                    onClick={() => {
                                                                        setConfirmAction({ res, force: true });
                                                                        setMenuOpen(null);
                                                                    }}
                                                                    className="w-full text-left px-4 py-2 text-sm text-red-300 hover:bg-red-900/40"
                                                                >
                                                                    Force Delete
                                                                </button>
                                                            )}
                                                        </>
                                                    ) : (
                                                        <div className="px-4 py-2 text-xs text-gray-500">
                                                            View only
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>

                            {/* Expanded Details */}
                            <div className={`${getExpandableRowClasses(isExpanded, false)}`}>
                                {isExpanded && (
                                    <div className="px-6 py-4 bg-gray-900/30 border-t border-gray-800">
                                        <div className="bg-gray-900/50 rounded-lg border border-gray-800 overflow-hidden relative">
                                            <div className="absolute top-4 right-4 z-10">
                                                <EditYamlButton onClick={() => setEditingResource(res)} namespace={res.namespace} />
                                            </div>
                                            {renderDetails(res)}
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    );
                })}
                {limitedResources.length === 0 && filter && (
                    <div className="p-8 text-center text-gray-500">
                        No resources match your filter.
                    </div>
                )}
            </div>


            {/* YAML Editor Modal */}
            {editingResource && (
                <YamlEditor
                    resource={editingResource}
                    onClose={() => setEditingResource(null)}
                    onSaved={() => {
                        setEditingResource(null);
                        refetch();
                    }}
                />
            )}

            {/* Menu overlay */}
            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(null)}
                ></div>
            )}

            {/* Delete confirmation modal */}
            {confirmAction && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                        <h3 className="text-lg font-semibold text-white mb-2">
                            Confirm delete
                        </h3>
                        <p className="text-sm text-gray-300 mb-4">
                            {confirmAction.force ? 'Force delete' : 'Delete'} {confirmAction.res.kind} "{confirmAction.res.name}"?
                            {confirmAction.res.kind === 'Node' && (
                                <span className="block mt-2 text-xs text-yellow-400">
                                    ⚠️ Warning: Deleting a node will remove it from the cluster. This action cannot be undone.
                                </span>
                            )}
                            {confirmAction.force && (
                                <span className="block mt-2 text-xs text-red-400">
                                    Warning: Force delete will immediately terminate the resource without graceful shutdown.
                                </span>
                            )}
                        </p>
                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setConfirmAction(null)}
                                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={() => handleDelete(confirmAction.res, confirmAction.force)}
                                className={`px-4 py-2 rounded-md text-white transition-colors ${
                                    confirmAction.force
                                        ? 'bg-red-700 hover:bg-red-800'
                                        : 'bg-orange-600 hover:bg-orange-700'
                                }`}
                            >
                                {confirmAction.force ? 'Force Delete' : 'Delete'}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Rollout confirmation modal */}
            {confirmRollout && (() => {
                const details = confirmRollout.details;
                const strategy = confirmRollout.strategy;
                const replicas = details?.replicas || 0;
                const ready = details?.ready || 0;
                const readyCount = typeof ready === 'string' ? parseInt(ready.split('/')[0]) || 0 : ready;
                const totalReplicas = typeof replicas === 'string' ? parseInt(replicas.split('/')[1]) || parseInt(replicas) || 0 : replicas;

                // Parse maxSurge and maxUnavailable
                const parseStrategyValue = (value, total) => {
                    if (!value) return 0;
                    if (typeof value === 'number') return value;
                    if (typeof value === 'string') {
                        if (value.endsWith('%')) {
                            const percent = parseFloat(value);
                            return Math.ceil((total * percent) / 100);
                        }
                        return parseInt(value) || 0;
                    }
                    return 0;
                };

                const maxSurge = strategy ? parseStrategyValue(strategy.maxSurge, totalReplicas) : Math.ceil(totalReplicas * 0.25);
                const maxUnavailable = strategy ? parseStrategyValue(strategy.maxUnavailable, totalReplicas) : Math.floor(totalReplicas * 0.25);
                const strategyType = strategy?.type || 'RollingUpdate';

                // Determine behavior message based on strategy and replica count
                let behaviorMessage = '';
                let behaviorColor = 'text-yellow-400';
                let strategyDetails = '';

                if (totalReplicas === 0) {
                    behaviorMessage = 'Warning: This deployment has 0 replicas. Rollout will have no effect.';
                    behaviorColor = 'text-gray-400';
                } else if (strategyType === 'Recreate') {
                    behaviorMessage = `This will use the Recreate strategy: all existing pods will be terminated before new pods are created.`;
                    behaviorColor = 'text-orange-400';
                    strategyDetails = `⚠️ Service will be unavailable during the rollout (approximately ${totalReplicas} pod(s) will be restarted sequentially).`;
                } else if (strategyType === 'RollingUpdate') {
                    // Calculate actual pods that can be unavailable
                    const minAvailable = Math.max(0, totalReplicas - maxUnavailable);
                    const maxTotal = totalReplicas + maxSurge;

                    if (totalReplicas === 1) {
                        behaviorMessage = 'This will restart the single pod, causing a brief service interruption.';
                        behaviorColor = 'text-orange-400';
                        strategyDetails = `⚠️ Service will be unavailable during the restart (1 pod will be replaced).`;
                    } else {
                        behaviorMessage = `This will use the RollingUpdate strategy: pods will be updated gradually while maintaining service availability.`;
                        behaviorColor = 'text-green-400';
                        strategyDetails = `✓ Kubernetes will maintain at least ${minAvailable} pod(s) available during the update.\n✓ Up to ${maxSurge} new pod(s) can be created above the desired count.\n✓ Up to ${maxUnavailable} pod(s) can be unavailable during the update.\n✓ Maximum ${maxTotal} pod(s) can exist simultaneously during the rollout.`;
                    }
                } else {
                    behaviorMessage = `This will restart ${totalReplicas} pods using the ${strategyType} strategy.`;
                    behaviorColor = 'text-yellow-400';
                }

                return (
                    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                        <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl max-h-[90vh] overflow-y-auto">
                            <h3 className="text-lg font-semibold text-white mb-2">
                                Confirm rollout
                            </h3>
                            <p className="text-sm text-gray-300 mb-3">
                                Rollout {confirmRollout.kind} "<span className="font-medium text-white">{confirmRollout.name}</span>"?
                            </p>
                            {details && (
                                <div className="mb-3 p-3 bg-gray-800/50 rounded border border-gray-700">
                                    <div className="text-xs text-gray-400 mb-2">Deployment Information:</div>
                                    <div className="text-sm text-gray-300 space-y-1">
                                        <div className="flex justify-between">
                                            <span className="text-gray-400">Total Replicas:</span>
                                            <span className="font-medium text-white">{totalReplicas}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-gray-400">Ready Pods:</span>
                                            <span className={`font-medium ${readyCount === totalReplicas ? 'text-green-400' : 'text-yellow-400'}`}>
                                                {readyCount} / {totalReplicas}
                                            </span>
                                        </div>
                                        {strategy && (
                                            <div className="flex justify-between">
                                                <span className="text-gray-400">Update Strategy:</span>
                                                <span className="font-medium text-blue-400">{strategyType}</span>
                                            </div>
                                        )}
                                        {strategy && strategyType === 'RollingUpdate' && (
                                            <>
                                                <div className="flex justify-between">
                                                    <span className="text-gray-400">Max Surge:</span>
                                                    <span className="font-medium text-white">{strategy.maxSurge || '25%'}</span>
                                                </div>
                                                <div className="flex justify-between">
                                                    <span className="text-gray-400">Max Unavailable:</span>
                                                    <span className="font-medium text-white">{strategy.maxUnavailable || '25%'}</span>
                                                </div>
                                            </>
                                        )}
                                    </div>
                                </div>
                            )}
                            <div className={`text-xs mb-3 p-3 rounded border ${behaviorColor.includes('green') ? 'bg-green-900/20 border-green-700/50' : behaviorColor.includes('orange') ? 'bg-orange-900/20 border-orange-700/50' : 'bg-yellow-900/20 border-yellow-700/50'}`}>
                                <p className={`font-medium mb-1 ${behaviorColor}`}>
                                    {behaviorMessage}
                                </p>
                                {strategyDetails && (
                                    <div className="mt-2 text-xs text-gray-300 whitespace-pre-line">
                                        {strategyDetails}
                                    </div>
                                )}
                            </div>
                            <div className="flex justify-end space-x-3">
                                <button
                                    onClick={() => setConfirmRollout(null)}
                                    className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                                >
                                    Cancel
                                </button>
                                <button
                                    onClick={() => handleRolloutDeployment(confirmRollout)}
                                    disabled={totalReplicas === 0}
                                    className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    Rollout
                                </button>
                            </div>
                        </div>
                    </div>
                );
            })()}

            {/* Job created success modal */}
            {createdJob && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                        <h3 className="text-lg font-semibold text-white mb-2">
                            Job creado
                        </h3>
                        <p className="text-sm text-gray-300 mb-4">
                            Se ha creado el job{' '}
                            <Link
                                to={`/dashboard/workloads/Job?search=${createdJob.name}&namespace=${createdJob.namespace}`}
                                onClick={() => setCreatedJob(null)}
                                className="text-blue-400 hover:text-blue-300 hover:underline font-medium"
                            >
                                {createdJob.name}
                            </Link>
                        </p>
                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setCreatedJob(null)}
                                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                            >
                                Aceptar
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default WorkloadList;
