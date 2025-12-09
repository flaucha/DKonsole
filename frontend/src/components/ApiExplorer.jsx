import React, { useEffect, useState, useMemo, useRef } from 'react';
import { Search, ListTree, RefreshCw, Globe, MapPin, FileText, X } from 'lucide-react';
import YamlEditor from './YamlEditor';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { getStatusBadgeClass } from '../utils/statusBadge';
import { formatDate } from '../utils/dateUtils';

const ApiExplorer = ({ namespace }) => {
    const { currentCluster } = useSettings();
    const { authFetch, user } = useAuth();
    const [apis, setApis] = useState([]);
    const [filter, setFilter] = useState('');
    const [selected, setSelected] = useState(null);
    const [objects, setObjects] = useState([]);
    const [loadingApis, setLoadingApis] = useState(false);
    const [loadingObjects, setLoadingObjects] = useState(false);
    const [isAdmin, setIsAdmin] = useState(false);
    const [checkingAdmin, setCheckingAdmin] = useState(true);
    const [scopeFilter, setScopeFilter] = useState('namespaced'); // 'namespaced', 'cluster', or 'all'
    const [yamlResource, setYamlResource] = useState(null); // {name, namespace, kind}
    const [showSuggestions, setShowSuggestions] = useState(false);
    const searchRef = useRef(null);

    // Check if user is admin
    useEffect(() => {
        const checkAdmin = async () => {
            try {
                const res = await authFetch('/api/settings/prometheus/url');
                if (res.ok || res.status === 404) {
                    setIsAdmin(true);
                } else {
                    setIsAdmin(false);
                }
            } catch {
                setIsAdmin(false);
            } finally {
                setCheckingAdmin(false);
            }
        };
        if (user) {
            checkAdmin();
        } else {
            setCheckingAdmin(false);
        }
    }, [authFetch, user]);

    const fetchApis = () => {
        setLoadingApis(true);
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);
        authFetch(`/api/apis?${params.toString()}`)
            .then(res => res.json())
            .then(data => {
                setApis(data || []);
                setLoadingApis(false);
            })
            .catch(() => setLoadingApis(false));
    };

    useEffect(() => {
        fetchApis();
    }, [currentCluster]);

    useEffect(() => {
        if (!selected) return;
        const params = new URLSearchParams({
            group: selected.group,
            version: selected.version,
            resource: selected.resource,
            namespace: namespace || '',
            namespaced: selected.namespaced ? 'true' : 'false',
        });
        if (currentCluster) params.append('cluster', currentCluster);
        setObjects([]);
        setLoadingObjects(true);
        authFetch(`/api/apis/resources?${params.toString()}`)
            .then(res => res.json())
            .then(data => {
                // Backend returns { resources: [...] } structure
                const resources = Array.isArray(data) ? data : (data?.resources || []);
                setObjects(resources);
            })
            .catch(() => setObjects([]))
            .finally(() => setLoadingObjects(false));
    }, [selected, namespace, currentCluster]);

    const filteredApis = useMemo(() => {
        const q = filter.toLowerCase();
        return apis
            .filter((a) => {
                const matchesText = `${a.kind}/${a.resource}`.toLowerCase().includes(q);
                const matchesScope =
                    scopeFilter === 'all' ? true : scopeFilter === 'namespaced' ? a.namespaced : !a.namespaced;
                // Non-admin users can only see namespaced resources
                const matchesPermission = isAdmin || a.namespaced;
                return matchesText && matchesScope && matchesPermission;
            })
            .sort((a, b) => a.kind.localeCompare(b.kind));
    }, [apis, filter, scopeFilter, isAdmin]);

    // Close suggestions when clicking outside
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (searchRef.current && !searchRef.current.contains(event.target)) {
                setShowSuggestions(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleViewYaml = (obj) => {
        if (!selected) return;
        setYamlResource({
            name: obj.name,
            namespace: obj.namespace,
            kind: selected.kind,
            group: selected.group,
            version: selected.version,
            resource: selected.resource,
            namespaced: selected.namespaced
        });
    };

    const filteredSuggestions = useMemo(() => {
        if (!filter) return [];
        const q = filter.toLowerCase();
        return apis.filter(a => `${a.kind}/${a.resource}`.toLowerCase().includes(q)).slice(0, 10);
    }, [apis, filter]);

    return (
        <div className="p-6">
            <div className="flex flex-col md:flex-row md:items-center justify-between mb-4 gap-4">
                <div className="flex items-center space-x-2 min-w-max">
                    <ListTree className="text-blue-400" size={18} />
                    <h1 className="text-xl font-semibold text-white">API Explorer</h1>
                    {loadingApis && <RefreshCw size={16} className="animate-spin text-gray-400" />}
                </div>

                <div className="flex-1 flex items-center gap-4">
                    <div className="relative flex-1 max-w-2xl" ref={searchRef}>
                        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
                        <input
                            value={filter}
                            onChange={(e) => {
                                setFilter(e.target.value);
                                setShowSuggestions(true);
                            }}
                            onFocus={() => setShowSuggestions(true)}
                            className="w-full bg-gray-800 border border-gray-700 rounded-md pl-8 pr-8 py-1.5 text-sm text-gray-200 focus:outline-none focus:border-blue-500 transition-all"
                            placeholder="Search API resources (e.g., pod, deployment, service)..."
                        />
                        {filter && (
                            <button
                                onClick={() => setFilter('')}
                                className="absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-200 transition-colors"
                                type="button"
                            >
                                <X size={14} />
                            </button>
                        )}
                        {showSuggestions && filter && filteredSuggestions.length > 0 && (
                            <div className="absolute top-full left-0 right-0 mt-1 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50 max-h-60 overflow-y-auto">
                                {filteredSuggestions.map((api) => (
                                    <button
                                        key={`${api.group}-${api.version}-${api.resource}`}
                                        onClick={() => {
                                            setSelected(api);
                                            setFilter(`${api.kind}`);
                                            setShowSuggestions(false);
                                        }}
                                        className="w-full text-left px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 flex justify-between items-center"
                                    >
                                        <span>{api.kind}</span>
                                        <span className="text-xs text-gray-500">{api.group ? `${api.group}/${api.version}` : api.version}</span>
                                    </button>
                                ))}
                            </div>
                        )}
                    </div>

                    {!checkingAdmin && isAdmin ? (
                        <div className="bg-gray-800 border border-gray-700 rounded-md flex overflow-hidden text-sm shrink-0">
                            <button
                                onClick={() => setScopeFilter('namespaced')}
                                className={`px-3 py-1.5 flex items-center space-x-1 ${scopeFilter === 'namespaced'
                                        ? 'bg-gray-700 text-white'
                                        : 'text-gray-300 hover:bg-gray-700'
                                    }`}
                            >
                                <MapPin size={14} /> <span className="hidden sm:inline">Namespaced</span>
                            </button>
                            <button
                                onClick={() => setScopeFilter('cluster')}
                                className={`px-3 py-1.5 flex items-center space-x-1 ${scopeFilter === 'cluster'
                                        ? 'bg-gray-700 text-white'
                                        : 'text-gray-300 hover:bg-gray-700'
                                    }`}
                            >
                                <Globe size={14} /> <span className="hidden sm:inline">Cluster</span>
                            </button>
                            <button
                                onClick={() => setScopeFilter('all')}
                                className={`px-3 py-1.5 flex items-center space-x-1 ${scopeFilter === 'all'
                                        ? 'bg-gray-700 text-white'
                                        : 'text-gray-300 hover:bg-gray-700'
                                    }`}
                            >
                                <ListTree size={14} /> <span className="hidden sm:inline">All</span>
                            </button>
                        </div>
                    ) : (
                        <div className="bg-gray-800 border border-gray-700 rounded-md flex overflow-hidden text-sm shrink-0">
                            <div className="px-3 py-1.5 flex items-center space-x-1 bg-gray-700 text-white">
                                <MapPin size={14} /> <span className="hidden sm:inline">Namespaced Only</span>
                            </div>
                        </div>
                    )}
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
                <div className="bg-gray-800 border border-gray-700 rounded-lg overflow-hidden">
                    <div className="px-4 py-3 border-b border-gray-700 text-gray-300 text-sm font-medium">
                        API Resources ({filteredApis.length}) {scopeFilter === 'namespaced' ? '(Namespaced)' : scopeFilter === 'cluster' ? '(Cluster-wide)' : ''}
                    </div>
                    <div className="max-h-[70vh] overflow-y-auto">
                        {filteredApis.map((api) => (
                            <button
                                key={`${api.group}-${api.version}-${api.resource}`}
                                onClick={() => setSelected(api)}
                                className={`w-full text-left px-4 py-2 text-sm border-b border-gray-800 hover:bg-gray-700/60 transition-colors ${selected?.resource === api.resource && selected?.version === api.version ? 'bg-gray-700 text-white' : 'text-gray-300'}`}
                            >
                                <div className="flex items-center justify-between">
                                    <span className="font-medium">{api.kind}</span>
                                    <span className="text-xs text-gray-500">{api.namespaced ? 'Namespaced' : 'Cluster'}</span>
                                </div>
                                <div className="text-xs text-gray-500">{api.group ? `${api.group}/${api.version}` : api.version}</div>
                            </button>
                        ))}
                        {filteredApis.length === 0 && (
                            <div className="px-4 py-3 text-xs text-gray-500">No APIs match this filter.</div>
                        )}
                    </div>
                </div>

                <div className="lg:col-span-2 bg-gray-800 border border-gray-700 rounded-lg overflow-hidden min-h-[70vh]">
                    <div className="px-4 py-3 border-b border-gray-700 flex items-center justify-between">
                        <div className="text-gray-300 text-sm font-medium">
                            {selected ? `${selected.kind} (${selected.resource})` : 'Select an API resource'}
                        </div>
                        {loadingObjects && <RefreshCw size={16} className="animate-spin text-gray-400" />}
                    </div>
                    {selected ? (
                        <div className="overflow-x-auto">
                            <table className="min-w-full">
                                <thead className="bg-gray-900">
                                    <tr>
                                        <th className="px-3 md:px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Name</th>
                                        {selected.namespaced && (
                                            <th className="px-2 md:px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Namespace</th>
                                        )}
                                        <th className="px-2 md:px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Status</th>
                                        <th className="px-2 md:px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Created</th>
                                        <th className="px-2 md:px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Actions</th>
                                    </tr>
                                </thead>
                                <tbody className="bg-gray-800 divide-y divide-gray-700">
                                    {objects.map((obj) => (
                                        <tr key={`${obj.namespace}-${obj.name}`}>
                                            <td className="px-3 md:px-4 py-2 text-sm text-gray-200">{obj.name}</td>
                                            {selected.namespaced && (
                                                <td className="px-2 md:px-4 py-2 text-sm text-gray-300">{obj.namespace || '-'}</td>
                                            )}
                                            <td className="px-2 md:px-4 py-2">
                                                {obj.status ? (
                                                    <span className={`px-2 py-1 text-xs rounded-full ${getStatusBadgeClass(obj.status)}`}>
                                                        {obj.status}
                                                    </span>
                                                ) : (
                                                    <span className="text-sm text-gray-300">—</span>
                                                )}
                                            </td>
                                            <td className="px-2 md:px-4 py-2 text-sm text-gray-400">
                                                {obj.created ? formatDate(obj.created) : '—'}
                                            </td>
                                            <td className="px-2 md:px-4 py-2 text-sm text-gray-300">
                                                <button
                                                    onClick={() => handleViewYaml(obj)}
                                                    className="flex items-center px-2.5 py-1 bg-gray-800 hover:bg-gray-700 text-gray-200 rounded-md border border-gray-700 text-xs transition-colors"
                                                >
                                                    <FileText size={12} className="mr-1.5" />
                                                    Edit YAML
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                    {objects.length === 0 && (
                                        <tr>
                                            <td
                                                colSpan={selected.namespaced ? 4 : 3}
                                                className="px-4 py-6 text-center text-sm text-gray-500"
                                            >
                                                No objects found in {selected.namespaced ? `namespace "${namespace}"` : 'cluster'}.
                                            </td>
                                        </tr>
                                    )}
                                </tbody>
                            </table>
                        </div>
                    ) : (
                        <div className="p-6 text-gray-500 text-sm">Choose an API resource to explore its objects.</div>
                    )}
                </div>
            </div>
            {yamlResource && (
                <YamlEditor
                    resource={yamlResource}
                    onClose={() => setYamlResource(null)}
                    onSaved={() => {
                        setYamlResource(null);
                        // Refresh list if needed, though usually not necessary for edit
                    }}
                />
            )}
        </div>
    );
};

export default ApiExplorer;
