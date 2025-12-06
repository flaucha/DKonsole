import React, { useState } from 'react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import YamlEditor from './YamlEditor';
import { useNamespaces } from '../hooks/useNamespaces';
import { isAdmin } from '../utils/permissions';
import NamespaceToolbar from './namespaces/NamespaceToolbar';
import NamespaceList from './namespaces/NamespaceList';

const NamespaceManager = () => {
    const { currentCluster } = useSettings();
    const { authFetch, user } = useAuth();
    const toast = useToast();
    const [expandedId, setExpandedId] = useState(null);
    const [sortField, setSortField] = useState('name');
    const [sortDirection, setSortDirection] = useState('asc');
    const [filter, setFilter] = useState('');
    const [isSearchFocused, setIsSearchFocused] = useState(false);
    const [confirmAction, setConfirmAction] = useState(null);
    const [editingYaml, setEditingYaml] = useState(null);

    const { data: namespaces = [], isLoading: loading, refetch } = useNamespaces(authFetch, currentCluster);

    const filteredNamespaces = namespaces.filter(ns => {
        if (!filter) return true;
        return ns.name.toLowerCase().includes(filter.toLowerCase());
    });

    const handleDelete = async (namespace, force = false) => {
        const params = new URLSearchParams({
            kind: 'Namespace',
            name: namespace
        });
        if (force) params.append('force', 'true');
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const res = await authFetch(`/api/resource?${params.toString()}`, {
                method: 'DELETE'
            });

            if (!res.ok) {
                throw new Error('Failed to delete namespace');
            }

            refetch();
        } catch (err) {
            toast.error(`Error deleting namespace: ${err.message}`);
        }
    };

    if (loading && namespaces.length === 0) {
        return <div className="text-gray-400 animate-pulse p-6">Loading namespaces...</div>;
    }

    return (
        <div className="flex flex-col h-full">
            <NamespaceToolbar
                filter={filter}
                setFilter={setFilter}
                isSearchFocused={isSearchFocused}
                setIsSearchFocused={setIsSearchFocused}
                filteredCount={filteredNamespaces.length}
                isAdmin={isAdmin(user)}
                onAdd={() => setEditingYaml({ kind: 'Namespace', namespaced: false, isNew: true })}
                onRefresh={() => refetch()}
            />

            <NamespaceList
                namespaces={filteredNamespaces}
                sortField={sortField}
                setSortField={setSortField}
                sortDirection={sortDirection}
                setSortDirection={setSortDirection}
                expandedId={expandedId}
                setExpandedId={setExpandedId}
                isAdmin={isAdmin(user)}
                onEditYaml={setEditingYaml}
                onDelete={setConfirmAction}
                user={user}
            />

            {editingYaml && (
                <YamlEditor
                    resource={editingYaml}
                    onClose={() => setEditingYaml(null)}
                    onSaved={() => {
                        setEditingYaml(null);
                        refetch();
                    }}
                />
            )}

            {confirmAction && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                        <h3 className="text-lg font-semibold text-white mb-2">
                            Confirm delete
                        </h3>
                        <p className="text-sm text-gray-300 mb-4">
                            {confirmAction.force ? 'Force delete' : 'Delete'} Namespace "{confirmAction.namespace}"?
                        </p>
                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setConfirmAction(null)}
                                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={async () => {
                                    await handleDelete(confirmAction.namespace, confirmAction.force);
                                    setConfirmAction(null);
                                }}
                                className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors"
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

export default NamespaceManager;
