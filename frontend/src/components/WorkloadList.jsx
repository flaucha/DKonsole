import React, { useState, useEffect } from 'react';
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
    Tag,
    Eye,
    EyeOff,
    Terminal,
    CirclePlus,
    CircleMinus,
    MoreVertical,
    Check,
    Pause,
    AlertTriangle,
    X,
    Plus,
    Minus,
    Shield, // Added
    Lock,   // Added
    Users,  // Added
    Play,   // Added
    Database, // Added
    Search, // Added
    RefreshCw // Added
} from 'lucide-react';
import LogViewer from './LogViewer';
import TerminalViewer from './TerminalViewer';
import YamlEditor from './YamlEditor';
import DeploymentMetrics from './DeploymentMetrics';
import PodMetrics from './PodMetrics';

import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';

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

const DetailRow = ({ label, value, icon: Icon, children }) => (
    <div className="flex items-center justify-between bg-gray-800 px-3 py-2 rounded border border-gray-700 mb-2">
        <div className="flex items-center">
            {Icon && <Icon size={14} className="mr-2 text-gray-500" />}
            <span className="text-xs text-gray-400">{label}</span>
        </div>
        <div className="flex items-center">
            <span className="text-sm font-mono text-white break-all text-right">
                {Array.isArray(value) ? (
                    value.length > 0 ? value.join(', ') : <span className="text-gray-600 italic">None</span>
                ) : (
                    value || <span className="text-gray-600 italic">None</span>
                )}
            </span>
            {children}
        </div>
    </div>
);

const DataSection = ({ data, isSecret = false }) => {
    if (!data || Object.keys(data).length === 0) return <div className="text-gray-500 italic text-sm">No data.</div>;
    return (
        <div className="mt-2 space-y-2">
            {Object.entries(data).map(([key, value]) => (
                <DataRow key={key} label={key} value={value} isSecret={isSecret} />
            ))}
        </div>
    );
};

const DataRow = ({ label, value, isSecret }) => {
    const [revealed, setRevealed] = useState(!isSecret);
    return (
        <div className="bg-gray-800 p-2 rounded border border-gray-700">
            <div className="flex justify-between items-start">
                <span className="text-xs font-medium text-gray-400 mb-1 block">{label}</span>
                {isSecret && (
                    <button
                        onClick={() => setRevealed(!revealed)}
                        className="text-gray-500 hover:text-gray-300 focus:outline-none"
                        title={revealed ? 'Hide value' : 'Show value'}
                    >
                        {revealed ? <EyeOff size={14} /> : <Eye size={14} />}
                    </button>
                )}
            </div>
            <div className="text-sm font-mono text-gray-300 break-all whitespace-pre-wrap">
                {revealed ? value : '••••••••'}
            </div>
        </div>
    );
};

// Detail components for each resource kind
const NodeDetails = ({ details, onEditYAML }) => {
    const addresses = details.addresses || [];
    const nodeInfo = details.nodeInfo || {};
    const conditions = details.conditions || {};
    const taints = details.taints || [];

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2 space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">System Info</h4>
                    <DetailRow label="Kernel" value={nodeInfo.kernelVersion} icon={Server} />
                    <DetailRow label="OS Image" value={nodeInfo.osImage} icon={HardDrive} />
                    <DetailRow label="Runtime" value={nodeInfo.containerRuntimeVersion} icon={Box} />
                    <DetailRow label="Kubelet" value={nodeInfo.kubeletVersion} icon={Activity} />
                    <DetailRow label="Arch" value={nodeInfo.architecture} icon={Server} />
                </div>
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Addresses</h4>
                    {addresses.map((addr, i) => (
                        <DetailRow key={i} label={addr.type} value={addr.address} icon={Network} />
                    ))}
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Capacity</h4>
                    <DetailRow label="CPU" value={details.capacity?.cpu} icon={Activity} />
                    <DetailRow label="Memory" value={details.capacity?.memory} icon={HardDrive} />
                    <DetailRow label="Pods" value={details.capacity?.pods} icon={Box} />
                </div>
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Allocatable</h4>
                    <DetailRow label="CPU" value={details.allocatable?.cpu} icon={Activity} />
                    <DetailRow label="Memory" value={details.allocatable?.memory} icon={HardDrive} />
                    <DetailRow label="Pods" value={details.allocatable?.pods} icon={Box} />
                </div>
            </div>

            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Conditions</h4>
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-2">
                    {Object.entries(conditions).map(([type, status]) => (
                        <div key={type} className={`px-3 py-2 rounded border ${status === 'True' && type === 'Ready' ? 'bg-green-900/20 border-green-800 text-green-400' : status === 'True' ? 'bg-red-900/20 border-red-800 text-red-400' : 'bg-gray-800 border-gray-700 text-gray-400'}`}>
                            <div className="text-xs font-medium">{type}</div>
                            <div className="text-xs mt-1">{status}</div>
                        </div>
                    ))}
                </div>
            </div>

            {taints.length > 0 && (
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Taints</h4>
                    <div className="flex flex-wrap gap-2">
                        {taints.map((t, i) => (
                            <span key={i} className="px-2 py-1 bg-yellow-900/20 border border-yellow-800 text-yellow-500 rounded text-xs">
                                {t.key}{t.value ? `=${t.value}` : ''}:{t.effect}
                            </span>
                        ))}
                    </div>
                </div>
            )}
            <div className="flex justify-end mt-4">
                <EditYamlButton onClick={onEditYAML} />
            </div>
        </div>
    );
};

const ServiceAccountDetails = ({ details, onEditYAML }) => {
    const secrets = details.secrets || [];
    const imagePullSecrets = details.imagePullSecrets || [];

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2 space-y-4">
            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Secrets</h4>
                {secrets.length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                        {secrets.map((s, i) => (
                            <span key={i} className="px-2 py-1 bg-gray-800 border border-gray-700 rounded text-xs text-gray-300 flex items-center">
                                <Key size={10} className="mr-1" /> {s.name}
                            </span>
                        ))}
                    </div>
                ) : (
                    <div className="text-sm text-gray-500 italic">No secrets</div>
                )}
            </div>
            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Image Pull Secrets</h4>
                {imagePullSecrets.length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                        {imagePullSecrets.map((s, i) => (
                            <span key={i} className="px-2 py-1 bg-gray-800 border border-gray-700 rounded text-xs text-gray-300 flex items-center">
                                <Database size={10} className="mr-1" /> {s.name}
                            </span>
                        ))}
                    </div>
                ) : (
                    <div className="text-sm text-gray-500 italic">No image pull secrets</div>
                )}
            </div>
            <div className="flex justify-end mt-4">
                <EditYamlButton onClick={onEditYAML} />
            </div>
        </div>
    );
};

const RoleDetails = ({ details, onEditYAML }) => {
    const rules = details.rules || [];

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2">
            <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Policy Rules</h4>
            <div className="overflow-x-auto">
                <table className="min-w-full text-xs text-left">
                    <thead>
                        <tr className="border-b border-gray-700">
                            <th className="py-2 px-2 font-medium text-gray-400">Resources</th>
                            <th className="py-2 px-2 font-medium text-gray-400">Verbs</th>
                            <th className="py-2 px-2 font-medium text-gray-400">API Groups</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-800">
                        {rules.map((rule, i) => (
                            <tr key={i}>
                                <td className="py-2 px-2 text-gray-300">
                                    {(rule.resources || []).join(', ') || '*'}
                                </td>
                                <td className="py-2 px-2 text-gray-300">
                                    {(rule.verbs || []).join(', ') || '*'}
                                </td>
                                <td className="py-2 px-2 text-gray-300">
                                    {(rule.apiGroups || []).join(', ') || '""'}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
            <div className="flex justify-end mt-4">
                <EditYamlButton onClick={onEditYAML} />
            </div>
        </div>
    );
};

const BindingDetails = ({ details, onEditYAML }) => {
    const subjects = details.subjects || [];
    const roleRef = details.roleRef || {};

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2 space-y-4">
            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Role Reference</h4>
                <div className="flex items-center space-x-2 text-sm text-gray-300">
                    <Lock size={14} className="text-gray-500" />
                    <span className="font-medium">{roleRef.kind}:</span>
                    <span>{roleRef.name}</span>
                </div>
            </div>
            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Subjects</h4>
                <div className="overflow-x-auto">
                    <table className="min-w-full text-xs text-left">
                        <thead>
                            <tr className="border-b border-gray-700">
                                <th className="py-2 px-2 font-medium text-gray-400">Kind</th>
                                <th className="py-2 px-2 font-medium text-gray-400">Name</th>
                                <th className="py-2 px-2 font-medium text-gray-400">Namespace</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-800">
                            {subjects.map((sub, i) => (
                                <tr key={i}>
                                    <td className="py-2 px-2 text-gray-300">{sub.kind}</td>
                                    <td className="py-2 px-2 text-gray-300">{sub.name}</td>
                                    <td className="py-2 px-2 text-gray-300">{sub.namespace || '—'}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
            <div className="flex justify-end mt-4">
                <EditYamlButton onClick={onEditYAML} />
            </div>
        </div>
    );
};

const DeploymentDetails = ({ details, onScale, scaling, res, onEditYAML }) => {
    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <div className="mb-2">
                        <DetailRow label="Replicas" value={`${details.ready} / ${details.replicas}`} icon={Layers}>
                            {onScale && (
                                <div className="flex items-center space-x-1 ml-2">
                                    <button
                                        onClick={() => onScale(-1)}
                                        disabled={scaling}
                                        className="p-1 rounded bg-gray-700 border border-gray-600 text-gray-200 hover:bg-gray-600 disabled:opacity-50"
                                        title="Scale down"
                                    >
                                        <Minus size={12} />
                                    </button>
                                    <button
                                        onClick={() => onScale(1)}
                                        disabled={scaling}
                                        className="p-1 rounded bg-gray-700 border border-gray-600 text-gray-200 hover:bg-gray-600 disabled:opacity-50"
                                        title="Scale up"
                                    >
                                        <Plus size={12} />
                                    </button>
                                </div>
                            )}
                        </DetailRow>
                    </div>
                    <DetailRow label="Images" value={details.images} icon={Box} />
                    <DetailRow label="Ports" value={details.ports?.map(p => p.toString())} icon={Network} />
                </div>
                <div>
                    <DetailRow label="PVCs" value={details.pvcs} icon={HardDrive} />
                    <DetailRow
                        label="Labels"
                        value={details.podLabels ? Object.entries(details.podLabels).map(([k, v]) => `${k}=${v}`) : []}
                        icon={Tag}
                    />
                </div>
            </div>
            <div className="flex justify-end mt-4">
                <EditYamlButton onClick={onEditYAML} />
            </div>
        </div>
    );
};

const ServiceDetails = ({ details, onEditYAML }) => {
    const type = details.type || 'ClusterIP';
    const clusterIP = details.clusterIP;
    const externalIPs = details.externalIPs || [];
    const ports = details.ports || [];
    const selector = details.selector || {};

    const getTypeColor = (t) => {
        switch (t) {
            case 'LoadBalancer': return 'text-blue-400 border-blue-400/30 bg-blue-400/10';
            case 'NodePort': return 'text-purple-400 border-purple-400/30 bg-purple-400/10';
            default: return 'text-gray-400 border-gray-600 bg-gray-800';
        }
    };

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2 space-y-4">
            <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                    <span className={`px-2 py-0.5 rounded text-xs font-medium border ${getTypeColor(type)}`}>
                        {type}
                    </span>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">IP Addresses</h4>
                    <div className="space-y-2">
                        <div className="flex items-center justify-between bg-gray-800 px-3 py-2 rounded border border-gray-700">
                            <span className="text-xs text-gray-400">Cluster IP</span>
                            <span className="text-sm font-mono text-white">{clusterIP}</span>
                        </div>
                        {externalIPs.map((ip, i) => (
                            <div key={i} className="flex items-center justify-between bg-gray-800 px-3 py-2 rounded border border-gray-700">
                                <span className="text-xs text-gray-400">External IP</span>
                                <span className="text-sm font-mono text-white">{ip}</span>
                            </div>
                        ))}
                    </div>
                </div>

                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Selector</h4>
                    <div className="flex flex-wrap gap-2">
                        {Object.keys(selector).length > 0 ? (
                            Object.entries(selector).map(([k, v]) => (
                                <div key={k} className="flex items-center px-2 py-1 bg-gray-800 border border-gray-700 rounded text-xs text-gray-300">
                                    <Tag size={12} className="mr-1.5 text-gray-500" />
                                    <span className="text-gray-400 mr-1">{k}:</span>
                                    <span className="text-white">{v}</span>
                                </div>
                            ))
                        ) : (
                            <span className="text-sm text-gray-500 italic">No selector</span>
                        )}
                    </div>
                </div>
            </div>

            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Ports</h4>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
                    {ports.map((port, i) => {
                        if (typeof port !== 'string') return null;
                        // Parse "80:30080/TCP" or similar format
                        const parts = port.split('/');
                        const protocol = parts[1] || 'TCP';
                        const portMap = parts[0].split(':');
                        const portNum = portMap[0];
                        const targetPort = portMap[1] || portNum;

                        return (
                            <div key={i} className="flex items-center justify-between bg-gray-800 px-3 py-2 rounded border border-gray-700">
                                <div className="flex items-center">
                                    <div className="w-1.5 h-1.5 rounded-full bg-green-500 mr-2"></div>
                                    <span className="text-sm font-mono text-white">{portNum}</span>
                                    {portNum !== targetPort && (
                                        <>
                                            <span className="text-gray-500 mx-1">→</span>
                                            <span className="text-xs text-gray-400">{targetPort}</span>
                                        </>
                                    )}
                                </div>
                                <span className="text-[10px] font-bold text-gray-500 bg-gray-900 px-1.5 py-0.5 rounded">
                                    {protocol}
                                </span>
                            </div>
                        );
                    })}
                </div>
            </div>
            <div className="flex justify-end mt-4">
                <EditYamlButton onClick={onEditYAML} />
            </div>
        </div>
    );
};

const IngressDetails = ({ details, onEditYAML }) => {
    const rules = details.rules || [];
    const tls = details.tls || [];
    const annotations = details.annotations || {};
    const loadBalancer = details.loadBalancer || [];

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2 space-y-6">
            {/* LoadBalancer Status */}
            {loadBalancer.length > 0 && (
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Address</h4>
                    <div className="flex flex-wrap gap-2">
                        {loadBalancer.map((lb, i) => (
                            <div key={i} className="flex items-center px-2 py-1 bg-gray-800 border border-gray-700 rounded text-sm text-white">
                                <Globe size={14} className="mr-2 text-blue-400" />
                                {lb.ip || lb.hostname}
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* Rules Section */}
            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Rules</h4>
                {rules.length > 0 ? (
                    <div className="space-y-3">
                        {rules.map((rule, i) => (
                            <div key={i} className="bg-gray-800 rounded border border-gray-700 overflow-hidden">
                                <div className="px-3 py-2 bg-gray-800/50 border-b border-gray-700 flex items-center">
                                    <span className="text-xs text-gray-500 uppercase mr-2">Host:</span>
                                    <span className="text-sm font-medium text-white">{rule.host || '*'}</span>
                                </div>
                                <div className="p-2 space-y-1">
                                    {rule.paths && rule.paths.map((path, j) => (
                                        <div key={j} className="flex items-center text-xs text-gray-300 pl-2">
                                            <div className="w-1.5 h-1.5 rounded-full bg-gray-600 mr-2"></div>
                                            {path}
                                        </div>
                                    ))}
                                </div>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className="text-sm text-gray-500 italic">No rules defined</div>
                )}
            </div>

            {/* TLS Section */}
            {tls.length > 0 && (
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">TLS Configuration</h4>
                    <div className="space-y-2">
                        {tls.map((t, i) => (
                            <div key={i} className="flex flex-col bg-gray-800 px-3 py-2 rounded border border-gray-700">
                                <div className="flex items-center mb-1">
                                    <Lock size={14} className="mr-2 text-green-400" />
                                    <span className="text-xs text-gray-500 uppercase mr-2">Secret:</span>
                                    <span className="text-sm font-medium text-white">{t.secretName}</span>
                                </div>
                                {t.hosts && t.hosts.length > 0 && (
                                    <div className="ml-6 text-xs text-gray-400">
                                        <span className="text-gray-500 mr-1">Hosts:</span>
                                        {t.hosts.join(', ')}
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* Annotations Section */}
            {Object.keys(annotations).length > 0 && (
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Annotations</h4>
                    <div className="bg-gray-800 rounded border border-gray-700 p-2">
                        <div className="grid grid-cols-1 gap-1">
                            {Object.entries(annotations)
                                .filter(([k]) => k !== 'kubectl.kubernetes.io/last-applied-configuration')
                                .map(([k, v]) => {
                                    let displayValue = v;
                                    let isJson = false;
                                    try {
                                        if (typeof v === 'string' && (v.startsWith('{') || v.startsWith('['))) {
                                            const parsed = JSON.parse(v);
                                            displayValue = JSON.stringify(parsed, null, 2);
                                            isJson = true;
                                        }
                                    } catch (e) { }

                                    return (
                                        <div key={k} className="text-xs break-all flex flex-col mb-2 border-b border-gray-800 pb-2 last:border-0">
                                            <span className="text-gray-500 font-medium mb-1">{k}:</span>
                                            {isJson ? (
                                                <pre className="text-gray-300 bg-gray-900 p-2 rounded overflow-x-auto font-mono">
                                                    {displayValue}
                                                </pre>
                                            ) : (
                                                <span className="text-gray-300">{v}</span>
                                            )}
                                        </div>
                                    );
                                })}
                        </div>
                    </div>
                </div>
            )}
            <div className="flex justify-end mt-4">
                <EditYamlButton onClick={onEditYAML} />
            </div>
        </div>
    );
};

const PodDetails = ({ details, onStreamLogs, onOpenTerminal, onEditYAML, pod }) => {
    const [showMenu, setShowMenu] = useState(false);
    const [activeTab, setActiveTab] = useState('details');
    const containers = details.containers || [];
    const metrics = details.metrics || {};

    return (
        <div className="mt-2">
            <div className="flex space-x-4 border-b border-gray-700 mb-4">
                <button
                    className={`pb-2 text-sm font-medium transition-colors ${activeTab === 'details' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('details')}
                >
                    Details
                </button>
                <button
                    className={`pb-2 text-sm font-medium transition-colors ${activeTab === 'metrics' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('metrics')}
                >
                    Metrics
                </button>
            </div>

            {activeTab === 'details' ? (
                <div className="p-4 bg-gray-900/50 rounded-md">
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-3">
                        <div>
                            <DetailRow label="Node" value={details.node} icon={Server} />
                            <DetailRow label="IP" value={details.ip} icon={Network} />
                        </div>
                        <div>
                            <DetailRow label="Restarts" value={details.restarts} icon={Activity} />
                            <DetailRow label="Containers" value={containers} icon={Box} />
                        </div>
                        <div>
                            {metrics.cpu && <DetailRow label="CPU" value={metrics.cpu} icon={Activity} />}
                            {metrics.memory && <DetailRow label="Memory" value={metrics.memory} icon={HardDrive} />}
                        </div>
                    </div>
                    <div className="flex justify-end space-x-2">
                        {/* Stream Logs */}
                        {containers.length > 1 ? (
                            <div className="relative">
                                <button
                                    onClick={() => setShowMenu(!showMenu)}
                                    className="flex items-center px-2.5 py-1 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-md border border-gray-600 text-xs transition-colors"
                                >
                                    <Terminal size={12} className="mr-1.5" />
                                    Stream Logs
                                    <ChevronDown size={12} className="ml-1.5" />
                                </button>
                                {showMenu && (
                                    <div className="absolute right-0 bottom-full mb-1 w-48 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-10">
                                        {containers.map(c => (
                                            <button
                                                key={c}
                                                onClick={() => {
                                                    onStreamLogs(c);
                                                    setShowMenu(false);
                                                }}
                                                className="w-full text-left px-3 py-2 text-xs text-gray-300 hover:bg-gray-700 first:rounded-t-md last:rounded-b-md"
                                            >
                                                {c}
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </div>
                        ) : (
                            <button
                                onClick={() => onStreamLogs(containers[0])}
                                className="flex items-center px-2.5 py-1 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-md border border-gray-600 text-xs transition-colors"
                            >
                                <Terminal size={12} className="mr-1.5" />
                                Stream Logs
                            </button>
                        )}
                        {/* Open Terminal */}
                        <button
                            onClick={() => onOpenTerminal(containers[0])}
                            className="flex items-center px-2.5 py-1 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-md border border-gray-600 text-xs transition-colors"
                        >
                            <Terminal size={12} className="mr-1.5" />
                            Open Terminal
                        </button>
                        <EditYamlButton onClick={onEditYAML} />
                    </div>
                </div>
            ) : (
                pod && <PodMetrics pod={{ name: pod.name }} namespace={pod.namespace} />
            )}
        </div>
    );
};

const ConfigMapDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="flex items-center justify-between mb-2">
            <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Data</h4>
        </div>
        <DataSection data={details.data} />
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

const SecretDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="flex items-center justify-between mb-2">
            <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Data</h4>
        </div>
        <DataSection data={details.data} isSecret={true} />
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

const NetworkPolicyDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <DetailRow label="Policy Types" value={details.policyTypes} icon={Tag} />
        <DetailRow
            label="Pod Selector"
            value={details.podSelector ? Object.entries(details.podSelector).map(([k, v]) => `${k}=${v}`) : []}
            icon={Tag}
        />
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

const StorageDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <DetailRow label="Access Modes" value={details.accessModes} icon={Tag} />
        <DetailRow label="Capacity" value={details.capacity} icon={HardDrive} />
        <DetailRow label="Storage Class" value={details.storageClassName} icon={Layers} />
        {details.volumeName && <DetailRow label="Volume" value={details.volumeName} icon={HardDrive} />}
        {details.claimRef && (
            <DetailRow
                label="Claim Ref"
                value={`${details.claimRef.namespace}/${details.claimRef.name}`}
                icon={FileText}
            />
        )}
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

const GenericDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <pre className="text-xs text-gray-400 overflow-auto max-h-40">{JSON.stringify(details, null, 2)}</pre>
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

const EditYamlButton = ({ onClick }) => (
    <button
        onClick={onClick}
        className="flex items-center px-2.5 py-1 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-md border border-gray-600 text-xs transition-colors"
    >
        <FileText size={12} className="mr-1.5" />
        Edit YAML
    </button>
);

const WorkloadList = ({ namespace, kind }) => {
    const [resources, setResources] = useState([]);
    const [loading, setLoading] = useState(false);
    const [expandedId, setExpandedId] = useState(null);
    const [loggingPod, setLoggingPod] = useState(null);
    const [terminalPod, setTerminalPod] = useState(null);
    const [editingResource, setEditingResource] = useState(null);
    const [reloadKey, setReloadKey] = useState(0);
    const [sortField, setSortField] = useState('name');
    const [sortDirection, setSortDirection] = useState('asc');
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [scaling, setScaling] = useState(null);

    const [filter, setFilter] = useState('');
    const [isSearchFocused, setIsSearchFocused] = useState(false);

    // Reset state when view context changes
    useEffect(() => {
        setResources([]);
        setExpandedId(null);
        setLoading(true);
        setFilter(''); // Reset filter on view change
    }, [namespace, kind, currentCluster]);

    useEffect(() => {
        if (!kind) return;
        let isMounted = true;
        const requestKind = kind;
        // Only set loading if resources are empty (initial load or view change)
        if (resources.length === 0) setLoading(true);

        const params = new URLSearchParams({ namespace, kind });
        if (currentCluster) params.append('cluster', currentCluster);
        authFetch(`/api/resources?${params.toString()}`)
            .then(res => res.json())
            .then(data => {
                if (isMounted && kind === requestKind) {
                    setResources(data || []);
                    setLoading(false);
                }
            })
            .catch(err => {
                if (isMounted && kind === requestKind) {
                    console.error('Failed to fetch resources:', err);
                    setLoading(false);
                }
            });
        return () => {
            isMounted = false;
        };
    }, [namespace, kind, reloadKey, currentCluster]);

    useEffect(() => {
        if (!kind) return;
        const params = new URLSearchParams({ namespace, kind });
        if (currentCluster) params.append('cluster', currentCluster);
        const es = new EventSource(`/api/resources/watch?${params.toString()}`);
        es.onmessage = () => setReloadKey((v) => v + 1);
        es.onerror = () => {
            es.close();
        };
        return () => es.close();
    }, [namespace, kind, currentCluster]);

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

    const renderSortIndicator = (field) => {
        if (sortField !== field) return null;
        return sortDirection === 'asc' ? '↑' : '↓';
    };



    const renderDetails = (res) => {
        const onEditYAML = () => setEditingResource(res);
        if (!res.details) return (
            <div className="p-4 text-gray-500 italic">
                No details available.
                <div className="flex justify-end mt-4">
                    <EditYamlButton onClick={onEditYAML} />
                </div>
            </div>
        );
        switch (res.kind) {
            case 'Node':
                return <NodeDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'ServiceAccount':
                return <ServiceAccountDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'Role':
            case 'ClusterRole':
                return <RoleDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'RoleBinding':
            case 'ClusterRoleBinding':
                return <BindingDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'Deployment':
                return (
                    <DeploymentDetails
                        details={res.details}
                        onScale={(delta) => handleScale(res, delta)}
                        scaling={scaling === res.name}
                        res={res}
                        onEditYAML={onEditYAML}
                    />
                );
            case 'Service':
                return <ServiceDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'Ingress':
                return <IngressDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'Pod':
                return (
                    <PodDetails
                        details={res.details}
                        onStreamLogs={(container) => setLoggingPod({ ...res, container })}
                        onOpenTerminal={(container) => setTerminalPod({ ...res, container })}
                        onEditYAML={onEditYAML}
                        pod={res}
                    />
                );
            case 'ConfigMap':
                return <ConfigMapDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'Secret':
                return <SecretDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'NetworkPolicy':
                return <NetworkPolicyDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'PersistentVolumeClaim':
            case 'PersistentVolume':
                return <StorageDetails details={res.details} onEditYAML={onEditYAML} />;
            case 'StorageClass':
                return (
                    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <DetailRow label="Provisioner" value={res.details?.provisioner} icon={HardDrive} />
                            <DetailRow label="Reclaim Policy" value={res.details?.reclaimPolicy} icon={Activity} />
                            <DetailRow label="Binding Mode" value={res.details?.volumeBindingMode} icon={Network} />
                            <DetailRow label="Volume Expansion" value={String(res.details?.allowVolumeExpansion)} icon={Layers} />
                            <DetailRow
                                label="Parameters"
                                value={res.details?.parameters ? Object.entries(res.details.parameters).map(([k, v]) => `${k}=${v}`) : []}
                                icon={Tag}
                            />
                            <DetailRow label="Mount Options" value={res.details?.mountOptions} icon={HardDrive} />
                        </div>
                        <div className="flex justify-end mt-4">
                            <EditYamlButton onClick={onEditYAML} />
                        </div>
                    </div>
                );
            case 'Job':
                return (
                    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <DetailRow label="Active" value={res.details?.active} icon={Activity} />
                            <DetailRow label="Succeeded" value={res.details?.succeeded} icon={Check} />
                            <DetailRow label="Failed" value={res.details?.failed} icon={X} />
                            <DetailRow label="Parallelism" value={res.details?.parallelism} icon={Layers} />
                            <DetailRow label="Completions" value={res.details?.completions} icon={Layers} />
                            <DetailRow label="Backoff Limit" value={res.details?.backoffLimit} icon={AlertTriangle} />
                        </div>
                        <div className="flex justify-end mt-4">
                            <EditYamlButton onClick={onEditYAML} />
                        </div>
                    </div>
                );
            case 'CronJob':
                return (
                    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <DetailRow label="Schedule" value={res.details?.schedule} icon={Clock} />
                            <DetailRow label="Suspend" value={String(res.details?.suspend)} icon={Pause} />
                            <DetailRow label="Concurrency" value={res.details?.concurrency} icon={Layers} />
                            <DetailRow label="Start Deadline" value={res.details?.startingDeadline} icon={Clock} />
                            <DetailRow label="Last Schedule" value={res.details?.lastSchedule} icon={Clock} />
                        </div>
                        <div className="flex justify-end mt-4">
                            <EditYamlButton onClick={onEditYAML} />
                        </div>
                    </div>
                );
            case 'StatefulSet':
                return (
                    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <DetailRow label="Replicas" value={`${res.details?.ready}/${res.details?.replicas}`} icon={Layers} />
                            <DetailRow label="Current" value={res.details?.current} icon={Activity} />
                            <DetailRow label="Updated" value={res.details?.update} icon={Activity} />
                            <DetailRow label="Service" value={res.details?.serviceName} icon={Network} />
                            <DetailRow label="Pod Mgmt" value={res.details?.podManagement} icon={Box} />
                            <DetailRow label="Update Strategy" value={res.details?.updateStrategy?.type} icon={Layers} />
                        </div>
                        <div className="flex justify-end mt-4">
                            <EditYamlButton onClick={onEditYAML} />
                        </div>
                    </div>
                );
            case 'DaemonSet':
                return (
                    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <DetailRow label="Desired" value={res.details?.desired} icon={Activity} />
                            <DetailRow label="Current" value={res.details?.current} icon={Activity} />
                            <DetailRow label="Ready" value={res.details?.ready} icon={Activity} />
                            <DetailRow label="Available" value={res.details?.available} icon={Check} />
                            <DetailRow label="Updated" value={res.details?.updated} icon={Layers} />
                            <DetailRow label="Misscheduled" value={res.details?.misscheduled} icon={AlertTriangle} />
                        </div>
                        <div className="flex justify-end mt-4">
                            <EditYamlButton onClick={onEditYAML} />
                        </div>
                    </div>
                );
            case 'HPA':
                return (
                    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <DetailRow label="Min Replicas" value={res.details?.minReplicas} icon={Layers} />
                            <DetailRow label="Max Replicas" value={res.details?.maxReplicas} icon={Layers} />
                            <DetailRow label="Current" value={res.details?.current} icon={Activity} />
                            <DetailRow label="Desired" value={res.details?.desired} icon={Activity} />
                            <DetailRow label="Metrics" value={res.details?.metrics ? res.details.metrics.map((m) => m.type).join(', ') : ''} icon={Activity} />
                            <DetailRow label="Last Scale" value={res.details?.lastScaleTime} icon={Clock} />
                        </div>
                        <div className="flex justify-end mt-4">
                            <EditYamlButton onClick={onEditYAML} />
                        </div>
                    </div>
                );
            default:
                return <GenericDetails details={res.details} onEditYAML={onEditYAML} />;
        }
    };

    if (loading) {
        return <div className="text-gray-400 animate-pulse p-6">Loading {kind}s...</div>;
    }

    if (resources.length === 0 && !filter) {
        return <div className="text-gray-500 italic p-6">No {kind}s found in this namespace.</div>;
    }

    const Icon = getIcon(kind);

    const triggerDelete = (res, force = false) => {
        const params = new URLSearchParams({ kind: res.kind, name: res.name });
        if (res.namespace) params.append('namespace', res.namespace);
        if (currentCluster) params.append('cluster', currentCluster);
        if (force) params.append('force', 'true');

        authFetch(`/api/resource?${params.toString()}`, { method: 'DELETE' })
            .then(async (resp) => {
                if (!resp.ok) {
                    const text = await resp.text();
                    throw new Error(text || 'Delete failed');
                }
                setReloadKey((v) => v + 1);
            })
            .catch((err) => {
                alert(err.message || 'Failed to delete resource');
            })
            .finally(() => setConfirmAction(null));
    };

    const handleScale = (res, delta) => {
        if (!res.namespace) return;
        setScaling(res.name);
        const params = new URLSearchParams({
            kind: 'Deployment',
            name: res.name,
            namespace: res.namespace,
            delta: String(delta),
        });
        if (currentCluster) params.append('cluster', currentCluster);


        authFetch(`/api/scale?${params.toString()}`, { method: 'POST' })
            .then(async (resp) => {
                if (!resp.ok) throw new Error('Scale failed');
                setReloadKey((v) => v + 1);
            })
            .catch((err) => alert(err.message))
            .finally(() => setScaling(null));
    };

    const triggerCronJob = (res) => {
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        authFetch(`/api/cronjobs/trigger?${params.toString()}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ namespace: res.namespace, name: res.name })
        })
            .then(async (resp) => {
                if (!resp.ok) throw new Error('Failed to trigger CronJob');
                await resp.json();
                setReloadKey(v => v + 1);
            })
            .catch(err => alert(err.message))
            .finally(() => setConfirmAction(null));
    };

    return (
        <>
            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(null)}
                ></div>
            )}
            <div className="flex flex-col md:flex-row md:items-center justify-between mb-6 gap-4">
                <div className="flex items-center space-x-3">
                    <div className={`p-2 rounded-lg bg-gray-800 ${loading ? 'animate-pulse' : ''}`}>
                        <Icon size={24} className="text-blue-400" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-white flex items-center gap-3">
                            {kind}
                            <span className="px-2.5 py-0.5 rounded-full bg-gray-800 text-gray-400 text-sm font-medium border border-gray-700">
                                {filteredResources.length}
                            </span>
                        </h1>
                        {!['Node', 'ClusterRole', 'ClusterRoleBinding', 'PersistentVolume', 'StorageClass'].includes(kind) && (
                            <p className="text-gray-400 text-sm mt-0.5">
                                {namespace === 'all' ? 'All Namespaces' : namespace}
                            </p>
                        )}
                    </div>
                </div>

                <div className="flex items-center gap-3">
                    {(kind === 'ClusterRole' || kind === 'ClusterRoleBinding') && (
                        <div className={`relative transition-all duration-200 ${isSearchFocused ? 'w-64' : 'w-48'}`}>
                            <Search size={16} className={`absolute left-3 top-1/2 -translate-y-1/2 transition-colors ${isSearchFocused ? 'text-blue-400' : 'text-gray-500'}`} />
                            <input
                                type="text"
                                value={filter}
                                onChange={(e) => setFilter(e.target.value)}
                                onFocus={() => setIsSearchFocused(true)}
                                onBlur={() => setIsSearchFocused(false)}
                                placeholder={`Search ${kind}s...`}
                                className={`w-full bg-gray-800 border rounded-md pl-9 pr-4 py-1.5 text-sm text-gray-200 focus:outline-none transition-all ${isSearchFocused ? 'border-blue-500 ring-1 ring-blue-500/20' : 'border-gray-700 hover:border-gray-600'}`}
                            />
                            {filter && (
                                <button
                                    onClick={() => setFilter('')}
                                    className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
                                >
                                    <X size={14} />
                                </button>
                            )}
                        </div>
                    )}
                    <button
                        onClick={() => setReloadKey(v => v + 1)}
                        disabled={loading}
                        className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg transition-colors disabled:opacity-50"
                        title="Refresh"
                    >
                        <RefreshCw size={20} className={loading ? 'animate-spin' : ''} />
                    </button>
                </div>
            </div>
            <div className="bg-gray-800 rounded-lg border border-gray-700">
                <table className="min-w-full border-separate border-spacing-0">
                    <thead>
                        <tr>
                            <th className="w-8 px-4 py-3 bg-gray-900 rounded-tl-lg border-b border-gray-700"></th>
                            <th
                                scope="col"
                                className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('name')}
                            >
                                Name <span className="inline-block text-[10px]">{renderSortIndicator('name')}</span>
                            </th>
                            <th
                                scope="col"
                                className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('kind')}
                            >
                                Kind <span className="inline-block text-[10px]">{renderSortIndicator('kind')}</span>
                            </th>
                            <th
                                scope="col"
                                className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('status')}
                            >
                                Status <span className="inline-block text-[10px]">{renderSortIndicator('status')}</span>
                            </th>
                            <th
                                scope="col"
                                className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('created')}
                            >
                                Created <span className="inline-block text-[10px]">{renderSortIndicator('created')}</span>
                            </th>
                            {kind === 'Pod' && (
                                <>
                                    <th
                                        scope="col"
                                        className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                        onClick={() => handleSort('cpu')}
                                    >
                                        CPU <span className="inline-block text-[10px]">{renderSortIndicator('cpu')}</span>
                                    </th>
                                    <th
                                        scope="col"
                                        className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                        onClick={() => handleSort('memory')}
                                    >
                                        Memory <span className="inline-block text-[10px]">{renderSortIndicator('memory')}</span>
                                    </th>
                                </>
                            )}
                            <th className="w-10 px-4 py-3 bg-gray-900 rounded-tr-lg border-b border-gray-700"></th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-800">
                        {filteredResources.map((res) => (
                            <React.Fragment key={res.uid}>
                                <tr
                                    onClick={() => toggleExpand(res.uid)}
                                    className={`group hover:bg-gray-800/50 transition-colors cursor-pointer ${expandedId === res.uid ? 'bg-gray-800/30' : ''}`}
                                >
                                    <td className="px-4 py-3 whitespace-nowrap text-gray-400 text-center">
                                        {expandedId === res.uid ? <CircleMinus size={16} /> : <CirclePlus size={16} />}
                                    </td>
                                    <td className="px-6 py-3 whitespace-nowrap">
                                        <div className="flex items-center">
                                            <div className="flex-shrink-0 h-6 w-6 bg-gray-700 rounded flex items-center justify-center text-gray-400">
                                                <Icon size={14} />
                                            </div>
                                            <div className="ml-4">
                                                <div className="text-sm font-medium text-white">{res.name}</div>
                                            </div>
                                        </div>
                                    </td>
                                    <td className="px-6 py-3 whitespace-nowrap">
                                        <div className="text-sm text-gray-300">{res.kind}</div>
                                    </td>
                                    <td className="px-6 py-3 whitespace-nowrap">
                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-700 text-gray-200">
                                            {res.status}
                                        </span>
                                    </td>
                                    <td className="px-6 py-3 whitespace-nowrap text-sm text-gray-400">
                                        {new Date(res.created).toLocaleDateString()}
                                    </td>
                                    {kind === 'Pod' && (
                                        <>
                                            <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-300">
                                                {res.details?.metrics?.cpu || '—'}
                                            </td>
                                            <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-300">
                                                {res.details?.metrics?.memory || '—'}
                                            </td>
                                        </>
                                    )}
                                    <td className="px-4 py-3 whitespace-nowrap text-gray-300" onClick={(e) => e.stopPropagation()}>
                                        <div className="relative flex items-center justify-end space-x-1">
                                            {res.kind === 'CronJob' && (
                                                <button
                                                    onClick={() => setConfirmAction({ res, action: 'trigger' })}
                                                    className="p-1 hover:bg-gray-700 rounded text-green-400 hover:text-green-300 transition-colors mr-1"
                                                    title="Run Now"
                                                >
                                                    <Play size={16} />
                                                </button>
                                            )}
                                            <button
                                                onClick={() => setMenuOpen(menuOpen === res.name ? null : res.name)}
                                                className="p-1 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                                            >
                                                <MoreVertical size={16} />
                                            </button>
                                            {menuOpen === res.name && (
                                                <div className="absolute right-0 mt-1 w-36 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
                                                    <div className="flex flex-col">

                                                        <button
                                                            onClick={() => {
                                                                setConfirmAction({ res, force: false });
                                                                setMenuOpen(null);
                                                            }}
                                                            className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                        >
                                                            Delete
                                                        </button>
                                                        <button
                                                            onClick={() => {
                                                                setConfirmAction({ res, force: true });
                                                                setMenuOpen(null);
                                                            }}
                                                            className="w-full text-left px-4 py-2 text-sm text-red-300 hover:bg-red-900/40"
                                                        >
                                                            Force Delete
                                                        </button>
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td colSpan={kind === 'Pod' ? 8 : 6} className={`px-6 pt-0 bg-gray-800 border-0 ${expandedId === res.uid ? 'border-b border-gray-700' : ''}`}>
                                        <div
                                            className={`pl-12 overflow-y-auto transition-all duration-300 ease-in-out ${expandedId === res.uid ? 'max-h-[500px] opacity-100 pb-4' : 'max-h-0 opacity-0'}`}
                                        >
                                            {expandedId === res.uid && renderDetails(res)}
                                        </div>
                                    </td>
                                </tr>
                            </React.Fragment>
                        ))}
                    </tbody>
                </table>
            </div>
            {loggingPod && (
                <LogViewer
                    namespace={loggingPod.namespace}
                    pod={loggingPod.name}
                    container={loggingPod.container}
                    onClose={() => setLoggingPod(null)}
                />
            )}
            {terminalPod && (
                <TerminalViewer
                    namespace={terminalPod.namespace}
                    pod={terminalPod.name}
                    container={terminalPod.container}
                    onClose={() => setTerminalPod(null)}
                />
            )}
            {editingResource && (
                <YamlEditor
                    resource={editingResource}
                    onClose={() => setEditingResource(null)}
                    onSaved={() => {
                        setEditingResource(null);
                        setReloadKey((v) => v + 1);
                    }}
                />
            )}
            {confirmAction && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                        <h3 className="text-lg font-semibold text-white mb-2">
                            {confirmAction.action === 'trigger' ? 'Confirm Run' : 'Confirm delete'}
                        </h3>
                        <p className="text-sm text-gray-300 mb-4">
                            {confirmAction.action === 'trigger' ? (
                                <>Run CronJob "{confirmAction.res.name}" now?</>
                            ) : (
                                <>{confirmAction.force ? 'Force delete' : 'Delete'} {confirmAction.res.kind} "{confirmAction.res.name}"?</>
                            )}
                        </p>
                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setConfirmAction(null)}
                                className="px-4 py-2 bg-gray-800 text-gray-200 rounded-md hover:bg-gray-700 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={() => {
                                    if (confirmAction.action === 'trigger') {
                                        triggerCronJob(confirmAction.res);
                                    } else {
                                        triggerDelete(confirmAction.res, confirmAction.force);
                                    }
                                }}
                                className={`px-4 py-2 rounded-md text-white transition-colors ${confirmAction.action === 'trigger'
                                    ? 'bg-green-600 hover:bg-green-700'
                                    : confirmAction.force ? 'bg-red-700 hover:bg-red-800' : 'bg-orange-600 hover:bg-orange-700'
                                    }`}
                            >
                                {confirmAction.action === 'trigger' ? 'Run Now' : (confirmAction.force ? 'Force delete' : 'Delete')}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </>
    );
};

export default WorkloadList;
