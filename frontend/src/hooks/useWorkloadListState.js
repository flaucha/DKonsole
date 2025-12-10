import { useState, useEffect, useRef } from 'react';
import { useLocation } from 'react-router-dom';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { useWorkloads } from './useWorkloads';
import { useWorkloadActions } from './useWorkloadActions';

export const useWorkloadListState = (namespace, kind) => {
    const { currentCluster } = useSettings();
    const { authFetch, user } = useAuth();
    const location = useLocation();

    // UI State
    const [expandedId, setExpandedId] = useState(null);
    const [sortField, setSortField] = useState('name');
    const [sortDirection, setSortDirection] = useState('asc');
    const [filter, setFilter] = useState('');
    const [isSearchFocused, setIsSearchFocused] = useState(false);
    const [menuOpen, setMenuOpen] = useState(null);
    const [draggingColumn, setDraggingColumn] = useState(null);

    // Edit/Modal State
    const [editingResource, setEditingResource] = useState(null);
    const [editingDataResource, setEditingDataResource] = useState(null);

    // Data State
    const [allResources, setAllResources] = useState([]);

    // Fetch Data
    const { data: resourcesData, isLoading: loading, error, refetch } = useWorkloads(authFetch, namespace, kind, currentCluster);

    // Actions Hook
    const actions = useWorkloadActions(authFetch, refetch, currentCluster);

    // Refs for detecting changes
    const prevKindRef = useRef(kind);
    const prevNamespaceRef = useRef(namespace);
    const prevClusterRef = useRef(currentCluster);

    // Reset state on context change
    useEffect(() => {
        const kindChanged = prevKindRef.current !== kind;
        const namespaceChanged = prevNamespaceRef.current !== namespace;
        const clusterChanged = prevClusterRef.current !== currentCluster;

        if (kindChanged || namespaceChanged || clusterChanged) {
            setAllResources([]);
            setExpandedId(null);
            setFilter('');
            actions.setConfirmAction(null);
            actions.setConfirmRollout(null);

            prevKindRef.current = kind;
            prevNamespaceRef.current = namespace;
            prevClusterRef.current = currentCluster;

            if (namespace && kind) {
                refetch();
            }
        }
    }, [namespace, kind, currentCluster, refetch, actions.setConfirmAction, actions.setConfirmRollout]);

    // Sync Data
    useEffect(() => {
        if (resourcesData) {
            let data = [];
            if (Array.isArray(resourcesData)) {
                data = resourcesData;
            } else if (resourcesData.resources && Array.isArray(resourcesData.resources)) {
                data = resourcesData.resources;
            }
            setAllResources(data);

            const searchParams = new URLSearchParams(location.search);
            const search = searchParams.get('search');
            if (search) {
                setFilter(search);
                const found = data.find(r => r.name === search);
                if (found) {
                    setExpandedId(found.uid);
                }
            }
        } else if (!loading) {
            setAllResources([]);
        }
    }, [resourcesData, loading, location.search]);

    const toggleExpand = (uid) => {
        setExpandedId(current => current === uid ? null : uid);
    };

    const handleSort = (field) => {
        setSortField((prevField) => {
            if (prevField === field) {
                setSortDirection((prevDir) => (prevDir === 'asc' ? 'desc' : 'asc'));
                return prevField;
            }
            setSortDirection('asc');
            return field;
        });
    };

    const handleAdd = () => {
        let targetNs = namespace;
        if (namespace === 'all') {
            const firstResource = allResources.find(r => r.namespace);
            targetNs = firstResource?.namespace || 'dkonsole';
        }
        setEditingResource({
            kind: kind,
            namespaced: true,
            isNew: true,
            namespace: targetNs,
            apiVersion: 'v1',
            metadata: {
                namespace: targetNs,
                name: `new-${kind.toLowerCase()}`
            }
        });
    };

    const onDetailsSaved = () => {
        setExpandedId(null);
        setTimeout(() => refetch(), 300);
    };

    return {
        // Data
        resources: allResources,
        loading,
        error,
        refetch,
        user,
        currentCluster,
        authFetch,

        // UI State
        expandedId,
        sortField,
        sortDirection,
        filter,
        isSearchFocused,
        menuOpen,
        draggingColumn,

        // Setters
        setFilter,
        setIsSearchFocused,
        setMenuOpen,
        setDraggingColumn,
        setEditingResource,
        setEditingDataResource,

        // Modal State
        editingResource,
        editingDataResource,

        // Actions
        actions,
        toggleExpand,
        handleSort,
        handleAdd,
        onDetailsSaved
    };
};
