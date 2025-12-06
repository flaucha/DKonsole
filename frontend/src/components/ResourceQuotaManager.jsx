import React, { useEffect, useState } from 'react';
import { AlertCircle } from 'lucide-react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import QuotaEditor from './QuotaEditor';
import LimitRangeEditor from './LimitRangeEditor';
import YamlEditor from './YamlEditor';
import { useResourceQuotas } from '../hooks/useResourceQuotas';
import { useNamespaces } from '../hooks/useNamespaces';
import QuotaToolbar from './resource-quotas/QuotaToolbar';
import QuotaList from './resource-quotas/QuotaList';
import LimitRangeList from './resource-quotas/LimitRangeList';

const ResourceQuotaManager = ({ namespace }) => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const toast = useToast();
    const location = useLocation();
    const navigate = useNavigate();

    const [activeTab, setActiveTab] = useState('quotas');
    const [editingQuota, setEditingQuota] = useState(null);
    const [editingLimitRange, setEditingLimitRange] = useState(null);
    const [editingYaml, setEditingYaml] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [namespaceFilter, setNamespaceFilter] = useState('all');

    const { data: namespaces = [] } = useNamespaces(authFetch, currentCluster);
    const { quotas: quotasQuery, limitRanges: limitRangesQuery } = useResourceQuotas(authFetch, namespaceFilter, currentCluster);

    const quotas = quotasQuery.data || [];
    const limitRanges = limitRangesQuery.data || [];
    const loading = quotasQuery.isLoading || limitRangesQuery.isLoading;

    useEffect(() => {
        if (namespace) {
            setNamespaceFilter(namespace);
        }
    }, [namespace]);

    useEffect(() => {
        const params = new URLSearchParams(location.search);
        const tab = params.get('tab');
        if (tab === 'quotas' || tab === 'limits') {
            setActiveTab(tab);
        }
    }, [location.search]);

    const handleTabChange = (tab) => {
        setActiveTab(tab);
        const params = new URLSearchParams(location.search);
        params.set('tab', tab);
        navigate(`${location.pathname}?${params.toString()}`, { replace: true });
    };

    const handleDelete = async (resource, kind, force = false) => {
        const params = new URLSearchParams({
            kind: kind,
            name: resource.name
        });
        if (resource.namespace) params.append('namespace', resource.namespace);
        if (force) params.append('force', 'true');
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const res = await authFetch(`/api/resource?${params.toString()}`, {
                method: 'DELETE'
            });

            if (!res.ok) {
                const errorText = await res.text();
                throw new Error(errorText || 'Failed to delete resource');
            }

            setConfirmAction(null);

            if (kind === 'ResourceQuota') {
                quotasQuery.refetch();
            } else {
                limitRangesQuery.refetch();
            }
        } catch (err) {
            toast.error(`Error deleting ${kind}: ${err.message}`);
            throw err;
        }
    };

    return (
        <div className="p-6">
            <QuotaToolbar
                activeTab={activeTab}
                setActiveTab={handleTabChange}
                namespaceFilter={namespaceFilter}
                setNamespaceFilter={setNamespaceFilter}
                namespaceFromUrl={namespace}
                loading={loading}
                quotasCount={quotas.length}
                limitRangesCount={limitRanges.length}
                onRefresh={() => {
                    quotasQuery.refetch();
                    limitRangesQuery.refetch();
                }}
                onAddQuota={(ns) => setEditingQuota({ namespace: ns, name: '', kind: 'ResourceQuota', isNew: true })}
                onAddLimitRange={(ns) => setEditingLimitRange({ namespace: ns, name: '', kind: 'LimitRange', isNew: true })}
                namespaces={namespaces}
            />

            <div className="animate-in fade-in duration-300">
                {activeTab === 'quotas' ? (
                    <QuotaList
                        quotas={quotas}
                        onEditYaml={setEditingYaml}
                        onDelete={setConfirmAction}
                    />
                ) : (
                    <LimitRangeList
                        limitRanges={limitRanges}
                        onEditYaml={setEditingYaml}
                        onDelete={setConfirmAction}
                    />
                )}
            </div>

            {editingQuota && (
                <QuotaEditor
                    resource={editingQuota}
                    onClose={() => setEditingQuota(null)}
                    onSaved={() => {
                        setEditingQuota(null);
                        quotasQuery.refetch();
                    }}
                />
            )}

            {editingLimitRange && (
                <LimitRangeEditor
                    resource={editingLimitRange}
                    onClose={() => setEditingLimitRange(null)}
                    onSaved={() => {
                        setEditingLimitRange(null);
                        limitRangesQuery.refetch();
                    }}
                />
            )}

            {editingYaml && (
                <YamlEditor
                    resource={editingYaml}
                    onClose={() => setEditingYaml(null)}
                    onSaved={() => {
                        setEditingYaml(null);
                        if (activeTab === 'quotas') {
                            quotasQuery.refetch();
                        } else {
                            limitRangesQuery.refetch();
                        }
                    }}
                />
            )}

            {confirmAction && (
                <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4 animate-in fade-in duration-200">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                        <div className="flex items-center space-x-3 mb-4 text-red-400">
                            <AlertCircle size={24} />
                            <h3 className="text-xl font-bold text-white">Confirm Deletion</h3>
                        </div>
                        <p className="text-gray-300 mb-6 leading-relaxed">
                            Are you sure you want to {confirmAction.force ? 'force ' : ''}delete the {confirmAction.kind === 'ResourceQuota' ? 'quota' : 'limit range'} <span className="font-bold text-white">"{confirmAction.resource.name}"</span>?
                            <br />
                            <span className="text-sm text-gray-500 mt-2 block">This action cannot be undone.</span>
                        </p>
                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setConfirmAction(null)}
                                className="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-lg transition-colors font-medium"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={async () => {
                                    await handleDelete(confirmAction.resource, confirmAction.kind, confirmAction.force);
                                }}
                                className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors font-medium shadow-lg shadow-red-900/30"
                            >
                                {confirmAction.force ? 'Force Delete' : 'Delete'}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ResourceQuotaManager;
