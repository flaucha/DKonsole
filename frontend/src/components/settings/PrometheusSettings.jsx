import React, { useState, useEffect } from 'react';
import { Server, Save, AlertCircle, Check } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { parseErrorResponse, parseError } from '../../utils/errorParser';

const PrometheusSettings = () => {
    const { authFetch } = useAuth();
    const [prometheusURL, setPrometheusURL] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    useEffect(() => {
        // Load current Prometheus URL
        authFetch('/api/settings/prometheus/url')
            .then(async (res) => {
                if (res.ok) {
                    const data = await res.json();
                    setPrometheusURL(data.url || '');
                }
            })
            .catch(() => { });
    }, [authFetch]);

    const handleSave = async () => {
        setError('');
        setSuccess('');
        setLoading(true);

        try {
            const res = await authFetch('/api/settings/prometheus/url', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ url: prometheusURL }),
            });

            if (res.ok) {
                setSuccess('Prometheus URL updated successfully! Reloading app...');
                setTimeout(() => window.location.reload(), 1500);
            } else {
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to update Prometheus URL');
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to update Prometheus URL');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
            <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                <Server size={20} className="mr-2 text-blue-400" /> Prometheus Configuration
            </h2>
            <div className="space-y-4">
                <p className="text-sm text-gray-400 mb-4">
                    Configure the Prometheus endpoint URL for historical metrics. Leave empty to disable.
                </p>
                <div className="flex items-center space-x-2">
                    <input
                        type="text"
                        value={prometheusURL}
                        onChange={(e) => setPrometheusURL(e.target.value)}
                        placeholder="http://prometheus-server.monitoring.svc.cluster.local:9090"
                        className="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                    />
                    <button
                        onClick={handleSave}
                        disabled={loading}
                        className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                    >
                        <Save size={16} className="mr-2" />
                        {loading ? 'Saving...' : 'Save'}
                    </button>
                </div>
                {error && (
                    <div className="flex items-center text-red-400 text-sm">
                        <AlertCircle size={16} className="mr-2" />
                        {error}
                    </div>
                )}
                {success && (
                    <div className="flex items-center text-green-400 text-sm">
                        <Check size={16} className="mr-2" />
                        {success}
                    </div>
                )}
            </div>
        </div>
    );
};

export default PrometheusSettings;
