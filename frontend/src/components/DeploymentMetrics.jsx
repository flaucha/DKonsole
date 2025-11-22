import React, { useState, useEffect, useRef } from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { useAuth } from '../context/AuthContext';
import { Activity, HardDrive } from 'lucide-react';

const parseCpu = (cpuStr) => {
    if (!cpuStr) return 0;
    cpuStr = cpuStr.toString().trim();
    if (cpuStr.endsWith('n')) return parseFloat(cpuStr.replace('n', '')) / 1000000; // nanocores to millicores
    if (cpuStr.endsWith('m')) return parseFloat(cpuStr.replace('m', '')); // millicores
    return parseFloat(cpuStr) * 1000; // cores to millicores
};

const parseMemory = (memStr) => {
    if (!memStr) return 0;
    memStr = memStr.toString().toUpperCase().trim();
    const num = parseFloat(memStr);
    if (isNaN(num)) return 0;
    if (memStr.includes('GI')) return num * 1024; // Gi to Mi
    if (memStr.includes('MI')) return num; // Mi
    if (memStr.includes('KI')) return num / 1024; // Ki to Mi
    return num; // bytes? usually Ki is min, but assume Mi if no unit? No, usually bytes.
    // If no unit, it's bytes.
    // But let's stick to what WorkloadList does, or improve it.
    // WorkloadList logic:
    // if (memStr.includes('GI')) return num * 1024;
    // if (memStr.includes('MI')) return num;
    // if (memStr.includes('KI')) return num / 1024;
    // return num; // This assumes input is Mi if no unit? Or just returns raw?
    // Let's assume output is MiB.
    if (!memStr.match(/[A-Z]/)) return num / 1024 / 1024; // Bytes to MiB
    return num;
};

const DeploymentMetrics = ({ deployment, namespace }) => {
    const { authFetch } = useAuth();
    const [data, setData] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const maxPoints = 30; // 30 points * 3s = 90s history

    useEffect(() => {
        let mounted = true;
        const fetchData = async () => {
            try {
                const params = new URLSearchParams({
                    namespace: namespace,
                    kind: 'Pod'
                });
                const resp = await authFetch(`/api/resources?${params.toString()}`);
                if (!resp.ok) throw new Error('Failed to fetch metrics');
                const pods = await resp.json();

                if (!mounted) return;

                // Filter pods belonging to this deployment
                const matchLabels = deployment.details?.podLabels || {};
                const deploymentPods = pods.filter(pod => {
                    if (!pod.details?.labels) return false;
                    for (const [key, val] of Object.entries(matchLabels)) {
                        if (pod.details.labels[key] !== val) return false;
                    }
                    return true;
                });

                // Aggregate metrics
                let totalCpu = 0;
                let totalMem = 0;

                deploymentPods.forEach(pod => {
                    if (pod.details?.metrics) {
                        totalCpu += parseCpu(pod.details.metrics.cpu);
                        totalMem += parseMemory(pod.details.metrics.memory);
                    }
                });

                const now = new Date();
                const timeStr = now.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' });

                setData(prev => {
                    const newData = [...prev, { time: timeStr, cpu: totalCpu, memory: totalMem }];
                    if (newData.length > maxPoints) return newData.slice(newData.length - maxPoints);
                    return newData;
                });
                setLoading(false);
            } catch (err) {
                console.error("Metrics fetch error:", err);
                if (mounted) setError(err.message);
            }
        };

        fetchData();
        const interval = setInterval(fetchData, 3000); // Poll every 3 seconds

        return () => {
            mounted = false;
            clearInterval(interval);
        };
    }, [deployment, namespace, authFetch]);

    if (error) return <div className="text-red-400 p-4">Error loading metrics: {error}</div>;
    if (loading && data.length === 0) return <div className="text-gray-500 p-4 animate-pulse">Initializing metrics...</div>;

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
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
                            <XAxis dataKey="time" stroke="#9CA3AF" fontSize={10} tick={{ fill: '#9CA3AF' }} />
                            <YAxis stroke="#9CA3AF" fontSize={10} tick={{ fill: '#9CA3AF' }} />
                            <Tooltip
                                contentStyle={{ backgroundColor: '#1F2937', borderColor: '#374151', color: '#F3F4F6' }}
                                itemStyle={{ color: '#60A5FA' }}
                            />
                            <Area type="monotone" dataKey="cpu" stroke="#60A5FA" fillOpacity={1} fill="url(#colorCpu)" isAnimationActive={false} />
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
                            <XAxis dataKey="time" stroke="#9CA3AF" fontSize={10} tick={{ fill: '#9CA3AF' }} />
                            <YAxis stroke="#9CA3AF" fontSize={10} tick={{ fill: '#9CA3AF' }} />
                            <Tooltip
                                contentStyle={{ backgroundColor: '#1F2937', borderColor: '#374151', color: '#F3F4F6' }}
                                itemStyle={{ color: '#A78BFA' }}
                            />
                            <Area type="monotone" dataKey="memory" stroke="#A78BFA" fillOpacity={1} fill="url(#colorMem)" isAnimationActive={false} />
                        </AreaChart>
                    </ResponsiveContainer>
                </div>
            </div>
        </div>
    );
};

export default DeploymentMetrics;
