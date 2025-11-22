import React, { useState, useEffect } from 'react';
import { Server, Layers, Box, Network, Globe, HardDrive, Activity, Database } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const StatCard = ({ icon: Icon, label, value, color }) => (
    <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 flex items-center shadow-lg">
        <div className={`p-3 rounded-full mr-4 ${color}`}>
            <Icon size={24} className="text-white" />
        </div>
        <div>
            <p className="text-gray-400 text-sm font-medium uppercase tracking-wider">{label}</p>
            <p className="text-2xl font-bold text-white">{value}</p>
        </div>
    </div>
);

const ClusterOverview = () => {
    const [stats, setStats] = useState(null);
    const [loading, setLoading] = useState(true);
    const { authFetch } = useAuth();

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
    }, []);

    if (loading) {
        return <div className="text-gray-400 animate-pulse p-6">Loading cluster overview...</div>;
    }

    if (!stats) {
        return <div className="text-red-400 p-6">Failed to load cluster statistics.</div>;
    }

    return (
        <div className="p-6">
            <h2 className="text-xl font-bold text-white mb-6 flex items-center">
                <Activity className="mr-2 text-blue-500" /> Cluster Overview
            </h2>

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

            <div className="mt-8 bg-gray-800 p-6 rounded-lg border border-gray-700">
                <h3 className="text-lg font-semibold text-white mb-4">Cluster Health</h3>
                <div className="flex items-center space-x-2">
                    <div className="h-3 w-3 rounded-full bg-green-500"></div>
                    <span className="text-gray-300">All systems operational</span>
                </div>
                <p className="text-sm text-gray-500 mt-2">
                    This dashboard provides a high-level summary of your cluster's resources.
                    Navigate to specific sections in the sidebar for detailed views.
                </p>
            </div>
        </div>
    );
};

export default ClusterOverview;
