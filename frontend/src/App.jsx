import React, { useState, useEffect, lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useParams, useLocation } from 'react-router-dom';
import Layout from './components/Layout';
import NamespaceSelector from './components/NamespaceSelector';
import Loading from './components/Loading';
import Settings from './components/Settings';
import About from './components/About';
import { SettingsProvider } from './context/SettingsContext';
import { AuthProvider, useAuth } from './context/AuthContext';
import Login from './components/Login';
import Setup from './components/Setup';
import { canEdit, isAdmin } from './utils/permissions';

// Lazy load large components for code splitting
const WorkloadList = lazy(() => import('./components/WorkloadList'));
const ClusterOverview = lazy(() => import('./components/ClusterOverview'));
const YamlImporter = lazy(() => import('./components/YamlImporter'));
const ApiExplorer = lazy(() => import('./components/ApiExplorer'));
const NamespaceManager = lazy(() => import('./components/NamespaceManager'));
const ResourceQuotaManager = lazy(() => import('./components/ResourceQuotaManager'));
const HelmChartManager = lazy(() => import('./components/HelmChartManager'));

const ProtectedRoute = ({ children }) => {
    const { user, loading, setupRequired } = useAuth();
    if (loading) return <div className="min-h-screen bg-gray-900 flex items-center justify-center text-white">Loading...</div>;
    if (setupRequired) return <Navigate to="/setup" />;
    if (!user) return <Navigate to="/login" />;
    return children;
};

const SetupOrLogin = () => {
    const { setupRequired, loading } = useAuth();
    if (loading) return <div className="min-h-screen bg-gray-900 flex items-center justify-center text-white">Loading...</div>;
    if (setupRequired) return <Setup />;
    return <Login />;
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
    const { user } = useAuth();

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
                            {/* Only show Import YAML if user has edit permission or is admin */}
                            {(isAdmin(user) || canEdit(user, selectedNamespace)) && (
                                <button
                                    onClick={() => setShowImporter(true)}
                                    className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded-md transition-colors border border-blue-500"
                                >
                                    Import YAML
                                </button>
                            )}
                        </div>
                    )
                }
            >
                <Routes>
                    <Route path="/" element={<Navigate to="overview" replace />} />
                    <Route path="overview" element={
                        <Suspense fallback={<Loading message="Loading cluster overview..." />}>
                            <ClusterOverview />
                        </Suspense>
                    } />
                    <Route path="workloads/:kind" element={
                        <Suspense fallback={<Loading message="Loading workloads..." />}>
                            <WorkloadListWrapper namespace={selectedNamespace} />
                        </Suspense>
                    } />
                    <Route path="settings" element={<Settings />} />
                    <Route path="about" element={<About />} />
                    <Route path="api-explorer" element={
                        <Suspense fallback={<Loading message="Loading API explorer..." />}>
                            <ApiExplorer namespace={selectedNamespace} />
                        </Suspense>
                    } />
                    <Route path="namespaces" element={
                        <Suspense fallback={<Loading message="Loading namespace manager..." />}>
                            <NamespaceManager />
                        </Suspense>
                    } />
                    <Route path="resource-quotas" element={
                        <Suspense fallback={<Loading message="Loading resource quotas..." />}>
                            <ResourceQuotaManager namespace={selectedNamespace} />
                        </Suspense>
                    } />
                    <Route path="helm-charts" element={
                        <Suspense fallback={<Loading message="Loading Helm charts..." />}>
                            <HelmChartManager namespace={selectedNamespace} />
                        </Suspense>
                    } />
                    {/* Fallback */}
                    <Route path="*" element={<Navigate to="overview" replace />} />
                </Routes>
                {showImporter && (
                    <Suspense fallback={<Loading message="Loading YAML importer..." />}>
                        <YamlImporter onClose={() => setShowImporter(false)} />
                    </Suspense>
                )}
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
                        <Route path="/setup" element={<Setup />} />
                        <Route path="/login" element={<SetupOrLogin />} />
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
