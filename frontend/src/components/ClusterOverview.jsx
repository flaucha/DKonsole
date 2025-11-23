import React, { useState, useEffect } from 'react';
import { Server, Layers, Box, Network, Globe, HardDrive, Activity, Database, Cpu, TrendingUp, Clock, AlertTriangle } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useSettings } from '../context/SettingsContext';

const StatCard = ({ icon: Icon, label, value, color, trend }) => (
    <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 flex items-center shadow-lg">
        <div className={`p-3 rounded-full mr-4 ${color}`}>
            <Icon size={24} className="text-white" />
        </div>
        <div className="flex-1">
            <p className="text-gray-400 text-sm font-medium uppercase tracking-wider">{label}</p>
            <div className="flex items-center justify-between">
                <p className="text-2xl font-bold text-white">{value}</p>
                {trend !== undefined && trend !== null && (
                    <span className={`text-xs ml-2 ${trend > 0 ? 'text-red-400' : trend < 0 ? 'text-green-400' : 'text-gray-400'}`}>
                        {trend > 0 ? '↑' : trend < 0 ? '↓' : '•'} {Math.abs(trend).toFixed(1)}%
                    </span>
                )}
            </div>
        </div>
    </div>
);

const ProgressBar = ({ value, color }) => {
    const colorClasses = {
        blue: 'bg-blue-600',
        purple: 'bg-purple-600',
        green: 'bg-green-600',
    };

    const percentage = Math.min(Math.max(value || 0, 0), 100);

    return (
        <div className="flex items-center gap-2">
            <div className="flex-1 bg-gray-700 rounded-full h-2 overflow-hidden">
                <div
                    className={`h-full ${colorClasses[color]} transition-all duration-300`}
                    style={{ width: `${percentage}%` }}
                />
            </div>
            <span className="text-xs text-gray-300 w-12 text-right">{percentage.toFixed(1)}%</span>
        </div>
    );
};

const ClusterOverview = () => {
    const [stats, setStats] = useState(null);
    const [loading, setLoading] = useState(true);
    const [prometheusEnabled, setPrometheusEnabled] = useState(false);
    const [nodeMetrics, setNodeMetrics] = useState([]);
    const [clusterStats, setClusterStats] = useState(null);
    const [timeRange, setTimeRange] = useState('1h');
    const { authFetch } = useAuth();
    const { currentCluster } = useSettings();

    const timeRanges = [
        { value: '1h', label: '1 Hour' },
        { value: '6h', label: '6 Hours' },
        { value: '12h', label: '12 Hours' },
        { value: '1d', label: '1 Day' },
    ];

    // Check Prometheus status
    useEffect(() => {
        const checkPrometheus = async () => {
            try {
                const params = new URLSearchParams();
                if (currentCluster) params.append('cluster', currentCluster);

                const response = await authFetch(`/api/prometheus/status?${params.toString()}`);
                const status = await response.json();
                setPrometheusEnabled(status.enabled);
            } catch (error) {
                console.error('Error checking Prometheus status:', error);
                setPrometheusEnabled(false);
            }
        };

        checkPrometheus();
    }, [currentCluster, authFetch]);

    // Fetch basic cluster stats
    useEffect(() => {
        authFetch('/api/overview')
            .then(res => res.json())
            .then(data => {
                setStats(data);
                setLoading(false);
            })
            .catch(err => {
                console.error('Failed to fetch cluster stats:', err);
                setLoading(false);
            });
    }, [authFetch]);

    // Fetch Prometheus metrics if enabled
    useEffect(() => {
        if (!prometheusEnabled) return;

        const fetchClusterMetrics = async () => {
            try {
                const params = new URLSearchParams({ range: timeRange });
                if (currentCluster) params.append('cluster', currentCluster);

                const response = await authFetch(`/api/prometheus/cluster-overview?${params.toString()}`);
                const data = await response.json();

                setNodeMetrics(data.nodeMetrics || []);
                setClusterStats(data.clusterStats || null);
            } catch (error) {
                console.error('Error fetching cluster metrics:', error);
            }
        };

        fetchClusterMetrics();
    }, [prometheusEnabled, timeRange, currentCluster, authFetch]);

    if (loading) {
        return <div className="text-gray-400 animate-pulse p-6">Loading cluster overview...</div>;
    }

    if (!stats) {
        return <div className="text-red-400 p-6">Failed to load cluster statistics.</div>;
    }

    return (
        <div className="p-6 space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                        <TrendingUp size={28} className="text-blue-400" />
                        Cluster Overview
                    </h2>
                    <p className="text-sm text-gray-400 mt-1">
                        {prometheusEnabled ? 'Real-time cluster metrics and node statistics' : 'High-level cluster resource summary'}
                    </p>
                </div>

                {/* Time Range Selector - Only show if Prometheus is enabled */}
                {prometheusEnabled && (
                    <div className="flex items-center gap-2">
                        <Clock size={14} className="text-gray-400" />
                        <span className="text-xs text-gray-400">Time Range:</span>
                        {timeRanges.map(range => (
                            <button
                                key={range.value}
                                onClick={() => setTimeRange(range.value)}
                                className={`px-3 py-1.5 text-xs rounded-md transition-all duration-200 ${timeRange === range.value
                                    ? 'bg-blue-600 text-white shadow-md'
                                    : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700'
                                    }`}
                            >
                                {range.label}
                            </button>
                        ))}
                    </div>
                )}
            </div>

            {/* Prometheus Metrics Stats - Only if enabled */}
            {prometheusEnabled && clusterStats && (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                    <StatCard
                        icon={Server}
                        label="Total Nodes"
                        value={clusterStats.totalNodes}
                        color="bg-blue-600"
                    />
                    <StatCard
                        icon={Cpu}
                        label="Avg CPU Usage"
                        value={`${clusterStats.avgCpuUsage?.toFixed(1)}%`}
                        color="bg-purple-600"
                        trend={clusterStats.cpuTrend}
                    />
                    <StatCard
                        icon={HardDrive}
                        label="Avg Memory Usage"
                        value={`${clusterStats.avgMemoryUsage?.toFixed(1)}%`}
                        color="bg-green-600"
                        trend={clusterStats.memoryTrend}
                    />
                    <StatCard
                        icon={Network}
                        label="Network Traffic"
                        value={`${clusterStats.networkTraffic?.toFixed(2)} MB/s`}
                        color="bg-yellow-600"
                    />
                </div>
            )}

            {/* Basic Resource Stats */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard
                    icon={Server}
                    label="Nodes"
                    value={stats.nodes}
                    color="bg-blue-600"
                />
                <StatCard
                    icon={Layers}
                    label="Namespaces"
                    value={stats.namespaces}
                    color="bg-purple-600"
                />
                <StatCard
                    icon={Box}
                    label="Pods"
                    value={stats.pods}
                    color="bg-green-600"
                />
                <StatCard
                    icon={Database}
                    label="Deployments"
                    value={stats.deployments}
                    color="bg-indigo-600"
                />
                <StatCard
                    icon={Network}
                    label="Services"
                    value={stats.services}
                    color="bg-yellow-600"
                />
                <StatCard
                    icon={Globe}
                    label="Ingresses"
                    value={stats.ingresses}
                    color="bg-pink-600"
                />
                <StatCard
                    icon={HardDrive}
                    label="PVCs"
                    value={stats.pvcs}
                    color="bg-orange-600"
                />
                <StatCard
                    icon={HardDrive}
                    label="PVs"
                    value={stats.pvs}
                    color="bg-red-600"
                />
            </div>

            {/* Node Metrics Table - Only if Prometheus is enabled */}
            {prometheusEnabled && nodeMetrics.length > 0 && (
                <div className="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
                    <div className="p-4 border-b border-gray-700">
                        <h3 className="text-lg font-semibold text-gray-200 flex items-center gap-2">
                            <Server size={20} className="text-blue-400" />
                            Node Metrics
                        </h3>
                    </div>
                    <div className="overflow-x-auto">
                        <table className="w-full">
                            <thead className="bg-gray-750">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Node</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">CPU Usage</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Memory Usage</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Disk Usage</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Network RX</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Network TX</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Status</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-700">
                                {nodeMetrics.map((node, idx) => (
                                    <tr key={idx} className="hover:bg-gray-750 transition-colors">
                                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-200">{node.name}</td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm">
                                            <ProgressBar value={node.cpuUsage} color="purple" />
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm">
                                            <ProgressBar value={node.memoryUsage} color="green" />
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm">
                                            <ProgressBar value={node.diskUsage} color="blue" />
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{node.networkRx?.toFixed(2)} KB/s</td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{node.networkTx?.toFixed(2)} KB/s</td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className={`px-2 py-1 text-xs rounded-full ${node.status === 'Ready'
                                                ? 'bg-green-900/50 text-green-300 border border-green-700'
                                                : 'bg-red-900/50 text-red-300 border border-red-700'
                                                }`}>
                                                {node.status}
                                            </span>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            )}

            {/* Cluster Health */}
            <div className="bg-gray-800 p-6 rounded-lg border border-gray-700">
                <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                    <Activity size={20} className="text-green-400" />
                    Cluster Health
                </h3>
                <div className="flex items-center space-x-2">
                    <div className="h-3 w-3 rounded-full bg-green-500"></div>
                    <span className="text-gray-300">All systems operational</span>
                </div>
                <p className="text-sm text-gray-500 mt-2">
                    {prometheusEnabled
                        ? 'Real-time metrics are being collected from Prometheus. Navigate to specific sections for detailed views.'
                        : 'This dashboard provides a high-level summary of your cluster\'s resources. Configure Prometheus for detailed metrics.'}
                </p>
            </div>
        </div>
    );
};

export default ClusterOverview;
