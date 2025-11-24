import React, { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useParams, useLocation } from 'react-router-dom';
import Layout from './components/Layout';
import NamespaceSelector from './components/NamespaceSelector';
import WorkloadList from './components/WorkloadList';
import Settings from './components/Settings';
import ClusterOverview from './components/ClusterOverview';
import { SettingsProvider } from './context/SettingsContext';
import { AuthProvider, useAuth } from './context/AuthContext';
import YamlImporter from './components/YamlImporter';
import ApiExplorer from './components/ApiExplorer';
import NamespaceManager from './components/NamespaceManager';
import ResourceQuotaManager from './components/ResourceQuotaManager';
import HelmChartManager from './components/HelmChartManager';
import Login from './components/Login';

const ProtectedRoute = ({ children }) => {
    const { user, loading } = useAuth();
    if (loading) return <div className="min-h-screen bg-gray-900 flex items-center justify-center text-white">Loading...</div>;
    if (!user) return <Navigate to="/login" />;
    return children;
};

const WorkloadListWrapper = ({ namespace }) => {
    const { kind } = useParams();
    if (!kind) {
        return <div className="text-red-400 p-6">Error: Resource type not specified in URL.</div>;
    }
    return <WorkloadList namespace={namespace} kind={kind} />;
};

const Dashboard = () => {
    // Load saved state from localStorage on mount
    const [selectedNamespace, setSelectedNamespace] = useState(() => {
        const saved = localStorage.getItem('dkonsole_selectedNamespace');
        return saved || 'default';
    });
    const [showImporter, setShowImporter] = useState(false);
    const location = useLocation();

    // Save to localStorage when namespace changes
    useEffect(() => {
        localStorage.setItem('dkonsole_selectedNamespace', selectedNamespace);
    }, [selectedNamespace]);

    const isSettings = location.pathname.includes('/settings');

    return (
        <SettingsProvider>
            <Layout
                headerContent={
                    !isSettings && (
                        <div className="flex items-center space-x-3">
                            <NamespaceSelector
                                selected={selectedNamespace}
                                onSelect={setSelectedNamespace}
                            />
                            <button
                                onClick={() => setShowImporter(true)}
                                className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded-md transition-colors border border-blue-500"
                            >
                                Import YAML
                            </button>
                        </div>
                    )
                }
            >
                <Routes>
                    <Route path="/" element={<Navigate to="overview" replace />} />
                    <Route path="overview" element={<ClusterOverview />} />
                    <Route path="workloads/:kind" element={<WorkloadListWrapper namespace={selectedNamespace} />} />
                    <Route path="settings" element={<Settings />} />
                    <Route path="api-explorer" element={<ApiExplorer namespace={selectedNamespace} />} />
                    <Route path="namespaces" element={<NamespaceManager />} />
                    <Route path="resource-quotas" element={<ResourceQuotaManager namespace={selectedNamespace} />} />
                    <Route path="helm-charts" element={<HelmChartManager />} />
                    {/* Fallback */}
                    <Route path="*" element={<Navigate to="overview" replace />} />
                </Routes>
                {showImporter && <YamlImporter onClose={() => setShowImporter(false)} />}
            </Layout>
        </SettingsProvider>
    );
};

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const queryClient = new QueryClient();

function App() {
    return (
        <QueryClientProvider client={queryClient}>
            <AuthProvider>
                <BrowserRouter>
                    <Routes>
                        <Route path="/login" element={<Login />} />
                        <Route path="/dashboard/*" element={
                            <ProtectedRoute>
                                <Dashboard />
                            </ProtectedRoute>
                        } />
                        <Route path="/" element={<Navigate to="/dashboard/overview" replace />} />
                        <Route path="*" element={<Navigate to="/dashboard/overview" replace />} />
                    </Routes>
                </BrowserRouter>
            </AuthProvider>
        </QueryClientProvider>
    );
}

export default App;

