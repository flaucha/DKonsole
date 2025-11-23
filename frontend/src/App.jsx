import React, { useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
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
import Login from './components/Login';

const ProtectedRoute = ({ children }) => {
    const { user, loading } = useAuth();
    if (loading) return <div className="min-h-screen bg-gray-900 flex items-center justify-center text-white">Loading...</div>;
    if (!user) return <Navigate to="/login" />;
    return children;
};

const Dashboard = () => {
    const [selectedNamespace, setSelectedNamespace] = useState('default');
    const [currentView, setCurrentView] = useState('Overview');
    const [showImporter, setShowImporter] = useState(false);

    // Map view names to API kinds
    const getKind = (view) => {
        const map = {
            'Deployments': 'Deployment',
            'Pods': 'Pod',
            'ConfigMaps': 'ConfigMap',
            'Secrets': 'Secret',
            'Jobs': 'Job',
            'CronJobs': 'CronJob',
            'StatefulSets': 'StatefulSet',
            'DaemonSets': 'DaemonSet',
            'HPA': 'HorizontalPodAutoscaler',
            'Services': 'Service',
            'Ingresses': 'Ingress',
            'Network Policies': 'NetworkPolicy',
            'PVCs': 'PersistentVolumeClaim',
            'PVs': 'PersistentVolume',
            'Storage Classes': 'StorageClass',
            'Nodes': 'Node',
            'Service Accounts': 'ServiceAccount',
            'Roles': 'Role',
            'Cluster Roles': 'ClusterRole',
            'Role Bindings': 'RoleBinding',
            'Cluster Role Bindings': 'ClusterRoleBinding'
        };
        return map[view] || '';
    };

    return (
        <SettingsProvider>
            <Layout
                currentView={currentView}
                onViewChange={setCurrentView}
                headerContent={
                    currentView !== 'Settings' && (
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
                {currentView === 'Settings' ? (
                    <Settings />
                ) : currentView === 'API Explorer' ? (
                    <div className="p-0">
                        <ApiExplorer namespace={selectedNamespace} />
                    </div>
                ) : currentView === 'Namespaces' ? (
                    <div className="p-0">
                        <NamespaceManager />
                    </div>
                ) : currentView === 'Resource Quotas' ? (
                    <div className="p-0">
                        <ResourceQuotaManager namespace={selectedNamespace} />
                    </div>
                ) : (
                    <div className="p-6">
                        {currentView === 'Overview' ? (
                            <ClusterOverview />
                        ) : (
                            <WorkloadList
                                namespace={selectedNamespace}
                                kind={getKind(currentView)}
                            />
                        )}
                    </div>
                )}
                {showImporter && <YamlImporter onClose={() => setShowImporter(false)} />}
            </Layout>
        </SettingsProvider>
    );
};

function App() {
    return (
        <AuthProvider>
            <BrowserRouter>
                <Routes>
                    <Route path="/login" element={<Login />} />
                    <Route path="/*" element={
                        <ProtectedRoute>
                            <Dashboard />
                        </ProtectedRoute>
                    } />
                </Routes>
            </BrowserRouter>
        </AuthProvider>
    );
}

export default App;
