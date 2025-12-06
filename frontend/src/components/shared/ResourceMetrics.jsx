import React, { useEffect, useState } from 'react';
import { Clock } from 'lucide-react';
import { useSettings } from '../../context/SettingsContext';
import { useAuth } from '../../context/AuthContext';
import { logger } from '../../utils/logger';
import MetricChart from './MetricChart';

const TIME_RANGES = [
    { value: '1h', label: '1 Hour' },
    { value: '6h', label: '6 Hours' },
    { value: '12h', label: '12 Hours' },
    { value: '1d', label: '1 Day' },
    { value: '7d', label: '7 Days' },
    { value: '15d', label: '15 Days' },
];

const ResourceMetrics = ({
    resourceName,
    namespace,
    resourceType, // 'pod' or 'deployment'
    apiEndpoint,
    chartConfigs = [], // Array of { title, dataKey, color, icon, unit }
}) => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [data, setData] = useState([]);
    const [loading, setLoading] = useState(true);
    const [prometheusEnabled, setPrometheusEnabled] = useState(false);
    const [timeRange, setTimeRange] = useState('1h');
    const [autoRefresh, setAutoRefresh] = useState(true);

    useEffect(() => {
        let interval;
        if (autoRefresh && prometheusEnabled) {
            interval = setInterval(() => {
                // Trigger re-fetch by invalidating a counter or just recalling logic?
                // Since fetchMetrics is defined inside another useEffect, we can't call it directly.
                // We can add a 'refreshTrigger' state.
                setRefreshTrigger(prev => prev + 1);
            }, 30000); // 30 seconds refresh
        }
        return () => clearInterval(interval);
    }, [autoRefresh, prometheusEnabled]);

    const [refreshTrigger, setRefreshTrigger] = useState(0);

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
        if (!prometheusEnabled || !resourceName || !namespace) {
            setLoading(false);
            return;
        }

        const fetchMetrics = async () => {
            setLoading(true);
            try {
                const params = new URLSearchParams({
                    [resourceType]: resourceName,
                    namespace: namespace,
                    range: timeRange,
                });
                if (currentCluster) params.append('cluster', currentCluster);

                const response = await authFetch(`${apiEndpoint}?${params.toString()}`);
                if (!response.ok) {
                    const errorText = await response.text();
                    logger.error('Failed to fetch metrics:', errorText);
                    throw new Error(errorText || 'Failed to fetch metrics');
                }
                const metricsData = await response.json();

                // Generic transformation logic
                // Collect all timestamps
                const allTimestamps = new Set();
                const maps = {};

                // Initialize maps for each config
                chartConfigs.forEach(config => {
                    maps[config.dataKey] = new Map();
                    // Identify the source key in response (usually matching dataKey or similar)
                    // We assume the API response key matches the dataKey for simplicity, 
                    // or we check if the API returns separate arrays like 'cpu', 'memory' matching dataKey
                    const sourceKey = config.apiResultKey || config.dataKey;

                    if (metricsData[sourceKey]) {
                        metricsData[sourceKey].forEach(point => {
                            maps[config.dataKey].set(point.timestamp, point.value);
                            allTimestamps.add(point.timestamp);
                        });
                    }
                });

                const sortedTimestamps = Array.from(allTimestamps).sort((a, b) => a - b);
                const transformedData = sortedTimestamps.map(ts => {
                    const date = new Date(ts);
                    const point = {
                        time: date.toLocaleTimeString('en-US', {
                            hour: '2-digit',
                            minute: '2-digit',
                            ...(timeRange.includes('d') ? { month: 'short', day: 'numeric' } : {})
                        }),
                    };
                    chartConfigs.forEach(config => {
                        point[config.dataKey] = maps[config.dataKey]?.get(ts) || 0;
                    });
                    return point;
                });

                setData(transformedData);
            } catch (error) {
                logger.error('Error fetching Prometheus metrics:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchMetrics();
    }, [resourceName, namespace, timeRange, prometheusEnabled, currentCluster, authFetch, resourceType, apiEndpoint, refreshTrigger]);

    if (!prometheusEnabled) {
        if (resourceType === 'deployment') {
            return (
                <div className="p-4 text-gray-500 text-sm italic">
                    Prometheus metrics not available. Configure PROMETHEUS_URL to enable historical metrics.
                </div>
            );
        }
        return null;
    }

    if (loading && data.length === 0) {
        return <div className="text-gray-500 p-4 animate-pulse text-sm">Loading metrics...</div>;
    }

    if (data.length === 0) {
        return <div className="text-gray-500 p-4 text-sm">No metrics data available for this {resourceType}.</div>;
    }

    // Filter charts that have data (optional, but PodMetrics checked hasNetworkData)
    // We can verify if any data point has value > 0 for that key
    const visibleCharts = chartConfigs.filter(config => {
        if (config.alwaysVisible) return true;
        return data.some(d => d[config.dataKey] > 0);
    });

    return (
        <div className={resourceType === 'pod' ? "p-4 bg-gray-900/50 rounded-md" : "mt-4"}>
            {resourceType === 'pod' && (
                <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Historical Metrics</h4>
            )}

            {/* Controls Row */}
            <div className="flex items-center justify-between mb-4 flex-wrap gap-2">
                <div className="flex items-center gap-2">
                    <Clock size={14} className="text-gray-400" />
                    <span className="text-xs text-gray-400">Time Range:</span>
                    <div className="relative">
                        <select
                            value={timeRange}
                            onChange={(e) => setTimeRange(e.target.value)}
                            className="appearance-none bg-gray-800 border border-gray-700 text-gray-300 text-xs rounded-md pl-3 pr-8 py-1 focus:outline-none focus:border-blue-500 cursor-pointer hover:bg-gray-700 transition-colors"
                        >
                            {TIME_RANGES.map(range => (
                                <option key={range.value} value={range.value}>
                                    {range.label}
                                </option>
                            ))}
                        </select>
                        <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-gray-400">
                            <svg className="fill-current h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><path d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z" /></svg>
                        </div>
                    </div>
                </div>

                <div className="flex items-center gap-2">
                    <span className="text-xs text-gray-400">Refresh:</span>
                    <button
                        onClick={() => setAutoRefresh(!autoRefresh)}
                        className={`relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${autoRefresh ? 'bg-blue-600' : 'bg-gray-700'}`}
                        role="switch"
                        aria-checked={autoRefresh}
                    >
                        <span
                            aria-hidden="true"
                            className={`pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${autoRefresh ? 'translate-x-4' : 'translate-x-0'}`}
                        />
                    </button>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {visibleCharts.map(config => (
                    <MetricChart
                        key={config.dataKey}
                        data={data}
                        dataKey={config.dataKey}
                        color={config.color}
                        title={config.title}
                        unit={config.unit}
                        icon={config.icon}
                    />
                ))}
            </div>
        </div>
    );
};

export default ResourceMetrics;
