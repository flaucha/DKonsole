import React, { useEffect, useState } from 'react';
import { Activity, HardDrive, Clock } from 'lucide-react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { logger } from '../utils/logger';

const DeploymentMetrics = ({ deployment, namespace }) => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [data, setData] = useState([]);
    const [loading, setLoading] = useState(true);
    const [prometheusEnabled, setPrometheusEnabled] = useState(false);
    const [timeRange, setTimeRange] = useState('1h');

    const timeRanges = [
        { value: '1h', label: '1 Hour' },
        { value: '6h', label: '6 Hours' },
        { value: '12h', label: '12 Hours' },
        { value: '1d', label: '1 Day' },
        { value: '7d', label: '7 Days' },
        { value: '15d', label: '15 Days' },
    ];

    useEffect(() => {
        // Check if Prometheus is enabled
        const checkPrometheus = async () => {
            try {
                const params = new URLSearchParams();
                if (currentCluster) params.append('cluster', currentCluster);

                const response = await authFetch(`/api/prometheus/status?${params.toString()}`);
                const status = await response.json();
                setPrometheusEnabled(status.enabled);
            } catch (error) {
                logger.error('Error checking Prometheus status:', error);
                setPrometheusEnabled(false);
            }
        };

        checkPrometheus();
    }, [currentCluster, authFetch]);

    useEffect(() => {
        if (!prometheusEnabled || !deployment || !namespace) {
            setLoading(false);
            return;
        }

        const fetchMetrics = async () => {
            setLoading(true);
            try {
                const params = new URLSearchParams({
                    deployment: deployment.name,
                    namespace: namespace,
                    range: timeRange,
                });
                if (currentCluster) params.append('cluster', currentCluster);

                const response = await authFetch(`/api/prometheus/metrics?${params.toString()}`);
                const metricsData = await response.json();

                // Transform data for recharts
                const transformedData = [];
                const cpuMap = new Map();
                const memMap = new Map();

                // Build maps by timestamp
                metricsData.cpu?.forEach(point => {
                    cpuMap.set(point.timestamp, point.value);
                });

                metricsData.memory?.forEach(point => {
                    memMap.set(point.timestamp, point.value);
                });

                // Merge data points
                const allTimestamps = new Set([...cpuMap.keys(), ...memMap.keys()]);
                const sortedTimestamps = Array.from(allTimestamps).sort((a, b) => a - b);

                sortedTimestamps.forEach(ts => {
                    const date = new Date(ts);
                    transformedData.push({
                        time: date.toLocaleTimeString('en-US', {
                            hour: '2-digit',
                            minute: '2-digit',
                            ...(timeRange.includes('d') ? { month: 'short', day: 'numeric' } : {})
                        }),
                        cpu: cpuMap.get(ts) || 0,
                        memory: memMap.get(ts) || 0,
                    });
                });

                setData(transformedData);
            } catch (error) {
                logger.error('Error fetching Prometheus metrics:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchMetrics();
    }, [deployment, namespace, timeRange, prometheusEnabled, currentCluster, authFetch]);

    if (!prometheusEnabled) {
        return (
            <div className="p-4 text-gray-500 text-sm italic">
                Prometheus metrics not available. Configure PROMETHEUS_URL to enable historical metrics.
            </div>
        );
    }

    if (loading && data.length === 0) {
        return <div className="text-gray-500 p-4 animate-pulse">Loading metrics...</div>;
    }

    if (data.length === 0) {
        return <div className="text-gray-500 p-4">No metrics data available for this deployment.</div>;
    }

    return (
        <div className="mt-4">
            {/* Time Range Selector */}
            <div className="flex items-center gap-2 mb-4 flex-wrap">
                <Clock size={16} className="text-gray-400" />
                <span className="text-xs text-gray-400">Time Range:</span>
                {timeRanges.map(range => (
                    <button
                        key={range.value}
                        onClick={() => setTimeRange(range.value)}
                        className={`px-3 py-1 text-xs rounded-md transition-colors ${timeRange === range.value
                                ? 'bg-blue-600 text-white'
                                : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700'
                            }`}
                    >
                        {range.label}
                    </button>
                ))}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {/* CPU Chart */}
                <div className="bg-gray-900/50 p-4 rounded-md border border-gray-700">
                    <div className="flex items-center mb-2">
                        <Activity size={16} className="text-blue-400 mr-2" />
                        <h3 className="text-xs font-medium text-gray-300">CPU (millicores)</h3>
                    </div>
                    <div className="h-32 w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={data}>
                                <defs>
                                    <linearGradient id="colorCpu" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#60A5FA" stopOpacity={0.3} />
                                        <stop offset="95%" stopColor="#60A5FA" stopOpacity={0} />
                                    </linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                                <XAxis
                                    dataKey="time"
                                    stroke="#9CA3AF"
                                    fontSize={10}
                                    tick={{ fill: '#9CA3AF' }}
                                    interval="preserveStartEnd"
                                />
                                <YAxis stroke="#9CA3AF" fontSize={10} tick={{ fill: '#9CA3AF' }} />
                                <Tooltip
                                    contentStyle={{ backgroundColor: '#1F2937', borderColor: '#374151', color: '#F3F4F6' }}
                                    itemStyle={{ color: '#60A5FA' }}
                                />
                                <Area
                                    type="monotone"
                                    dataKey="cpu"
                                    stroke="#60A5FA"
                                    fillOpacity={1}
                                    fill="url(#colorCpu)"
                                    isAnimationActive={false}
                                />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                </div>

                {/* Memory Chart */}
                <div className="bg-gray-900/50 p-4 rounded-md border border-gray-700">
                    <div className="flex items-center mb-2">
                        <HardDrive size={16} className="text-purple-400 mr-2" />
                        <h3 className="text-xs font-medium text-gray-300">Memory (MiB)</h3>
                    </div>
                    <div className="h-32 w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={data}>
                                <defs>
                                    <linearGradient id="colorMem" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#A78BFA" stopOpacity={0.3} />
                                        <stop offset="95%" stopColor="#A78BFA" stopOpacity={0} />
                                    </linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                                <XAxis
                                    dataKey="time"
                                    stroke="#9CA3AF"
                                    fontSize={10}
                                    tick={{ fill: '#9CA3AF' }}
                                    interval="preserveStartEnd"
                                />
                                <YAxis stroke="#9CA3AF" fontSize={10} tick={{ fill: '#9CA3AF' }} />
                                <Tooltip
                                    contentStyle={{ backgroundColor: '#1F2937', borderColor: '#374151', color: '#F3F4F6' }}
                                    itemStyle={{ color: '#A78BFA' }}
                                />
                                <Area
                                    type="monotone"
                                    dataKey="memory"
                                    stroke="#A78BFA"
                                    fillOpacity={1}
                                    fill="url(#colorMem)"
                                    isAnimationActive={false}
                                />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default DeploymentMetrics;
