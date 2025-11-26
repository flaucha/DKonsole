import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { Box } from 'lucide-react';

const AssociatedPods = ({ namespace, selector }) => {
    const { authFetch } = useAuth();
    const [pods, setPods] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    useEffect(() => {
        if (!selector || Object.keys(selector).length === 0) return;

        const fetchPods = async () => {
            setLoading(true);
            try {
                const labelSelector = Object.entries(selector)
                    .map(([k, v]) => `${k}=${v}`)
                    .join(',');

                const res = await authFetch(`/api/resources?kind=Pod&namespace=${namespace}&labelSelector=${encodeURIComponent(labelSelector)}`);
                if (res.ok) {
                    const data = await res.json();
                    setPods(data || []);
                } else {
                    setError("Failed to fetch pods");
                }
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchPods();
    }, [namespace, selector, authFetch]);

    if (!selector || Object.keys(selector).length === 0) {
         return <div className="text-gray-500 italic text-sm p-4">No selector defined.</div>;
    }

    if (loading) return <div className="text-gray-400 text-sm p-4">Loading pods...</div>;
    if (error) return <div className="text-red-400 text-sm p-4">Error: {error}</div>;
    if (pods.length === 0) return <div className="text-gray-500 italic text-sm p-4">No associated pods found.</div>;

    return (
        <div className="overflow-x-auto">
            <table className="min-w-full text-left text-sm text-gray-400">
                <thead className="bg-gray-800/50 text-xs uppercase font-medium text-gray-400">
                    <tr>
                        <th className="px-4 py-2 rounded-tl-md">Name</th>
                        <th className="px-4 py-2">Status</th>
                        <th className="px-4 py-2">Ready</th>
                        <th className="px-4 py-2">Restarts</th>
                        <th className="px-4 py-2 rounded-tr-md">Created</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-gray-800/50">
                    {pods.map((pod) => (
                        <tr key={pod.uid} className="hover:bg-gray-800/30 transition-colors">
                            <td className="px-4 py-2 font-medium text-blue-400">
                                <Link to={`/dashboard/workloads/Pod?search=${pod.name}`} className="hover:underline flex items-center">
                                     <Box size={14} className="mr-2" />
                                     {pod.name}
                                </Link>
                            </td>
                            <td className="px-4 py-2">
                                <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                                    pod.status === 'Running' || pod.status === 'Completed' ? 'bg-green-900/30 text-green-400 border border-green-700/50' :
                                    pod.status === 'Pending' ? 'bg-yellow-900/30 text-yellow-400 border border-yellow-700/50' :
                                    'bg-red-900/30 text-red-400 border border-red-700/50'
                                }`}>
                                    {pod.status}
                                </span>
                            </td>
                            <td className="px-4 py-2 text-gray-300">{pod.details?.ready}</td>
                            <td className="px-4 py-2 text-gray-300">{pod.details?.restarts}</td>
                            <td className="px-4 py-2 text-gray-500">{new Date(pod.created).toLocaleString()}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
};

export default AssociatedPods;
