import React, { useEffect, useState, useMemo } from 'react';
import { Search, Package, RefreshCw, Globe, MapPin, FileText, X } from 'lucide-react';
import Editor from '@monaco-editor/react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { getStatusBadgeClass } from '../utils/statusBadge';
import { formatDate } from '../utils/dateUtils';

const CRDExplorer = ({ namespace }) => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [crds, setCrds] = useState([]);
    const [filter, setFilter] = useState('');
    const [selected, setSelected] = useState(null);
    const [resources, setResources] = useState([]);
    const [loadingCRDs, setLoadingCRDs] = useState(false);
    const [loadingResources, setLoadingResources] = useState(false);
    const [scopeFilter, setScopeFilter] = useState('all'); // namespaced | cluster | all
    const [yamlView, setYamlView] = useState(null); // {name, namespace, content}

    const fetchCRDs = () => {
        setLoadingCRDs(true);
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);
        authFetch(`/api/crds?${params.toString()}`)
            .then(res => res.json())
            .then(data => {
                setCrds(data || []);
                setLoadingCRDs(false);
            })
            .catch(() => setLoadingCRDs(false));
    };

    useEffect(() => {
        fetchCRDs();
    }, [currentCluster]);

    useEffect(() => {
        if (!selected) return;
        const params = new URLSearchParams({
            group: selected.group,
            version: selected.version,
            resource: selected.name.split('.')[0], // CRD name format: plural.group
            namespace: selected.scope === 'Namespaced' ? namespace : '',
            namespaced: selected.scope === 'Namespaced' ? 'true' : 'false',
        });
        if (currentCluster) params.append('cluster', currentCluster);
        setResources([]);
        setLoadingResources(true);
        authFetch(`/api/crds/resources?${params.toString()}`)
            .then(res => res.json())
            .then(data => {
                setResources(data || []);
            })
            .catch(() => setResources([]))
            .finally(() => setLoadingResources(false));
    }, [selected, namespace, currentCluster]);

    const filteredCRDs = useMemo(() => {
        const q = filter.toLowerCase();
        return crds
            .filter((crd) => {
                const matchesText = `${crd.kind}/${crd.name}`.toLowerCase().includes(q);
                const matchesScope =
                    scopeFilter === 'all' ? true : scopeFilter === 'namespaced' ? crd.scope === 'Namespaced' : crd.scope === 'Cluster';
                return matchesText && matchesScope;
            })
            .sort((a, b) => a.kind.localeCompare(b.kind));
    }, [crds, filter, scopeFilter]);

    const handleViewYaml = (resource) => {
        if (!selected) return;
        const params = new URLSearchParams({
            group: selected.group,
            version: selected.version,
            resource: selected.name.split('.')[0],
            name: resource.name,
            namespace: selected.scope === 'Namespaced' ? resource.namespace || '' : '',
            namespaced: selected.scope === 'Namespaced' ? 'true' : 'false',
        });
        if (currentCluster) params.append('cluster', currentCluster);
        authFetch(`/api/crds/yaml?${params.toString()}`)
            .then((res) => res.text())
            .then((yaml) => setYamlView({ name: resource.name, namespace: resource.namespace, content: yaml }))
            .catch((err) => setYamlView({ name: resource.name, namespace: resource.namespace, content: `# Error: ${err.message}` }));
    };

    return (
        <div className="p-6">
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center space-x-2">
                    <Package className="text-blue-400" size={18} />
                    <h1 className="text-xl font-semibold text-white">CRD Explorer</h1>
                    {loadingCRDs && <RefreshCw size={16} className="animate-spin text-gray-400" />}
                </div>
                <div className="flex items-center space-x-3">
                    <div className="relative w-64">
                        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
                        <input
                            value={filter}
                            onChange={(e) => setFilter(e.target.value)}
                            className="w-full bg-gray-800 border border-gray-700 rounded-md pl-8 pr-8 py-1.5 text-sm text-gray-200 focus:outline-none focus:border-blue-500"
                            placeholder="Search CRDs..."
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
                    </div>
                    <div className="bg-gray-800 border border-gray-700 rounded-md flex overflow-hidden text-sm">
                        <button
                            onClick={() => setScopeFilter('namespaced')}
                            className={`px-3 py-1.5 flex items-center space-x-1 ${scopeFilter === 'namespaced' ? 'bg-blue-900 text-blue-100' : 'text-gray-300 hover:bg-gray-700'}`}
                        >
                            <MapPin size={14} /> <span>Namespaced</span>
                        </button>
                        <button
                            onClick={() => setScopeFilter('cluster')}
                            className={`px-3 py-1.5 flex items-center space-x-1 ${scopeFilter === 'cluster' ? 'bg-blue-900 text-blue-100' : 'text-gray-300 hover:bg-gray-700'}`}
                        >
                            <Globe size={14} /> <span>Cluster</span>
                        </button>
                        <button
                            onClick={() => setScopeFilter('all')}
                            className={`px-3 py-1.5 flex items-center space-x-1 ${scopeFilter === 'all' ? 'bg-blue-900 text-blue-100' : 'text-gray-300 hover:bg-gray-700'}`}
                        >
                            <span>All</span>
                        </button>
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
                <div className="bg-gray-800 border border-gray-700 rounded-lg overflow-hidden">
                    <div className="px-4 py-3 border-b border-gray-700 text-gray-300 text-sm font-medium">
                        Custom Resource Definitions ({filteredCRDs.length}) {scopeFilter === 'namespaced' ? '(Namespaced)' : scopeFilter === 'cluster' ? '(Cluster-wide)' : ''}
                    </div>
                    <div className="max-h-[70vh] overflow-y-auto">
                        {filteredCRDs.map((crd) => (
                            <button
                                key={`${crd.group}-${crd.version}-${crd.name}`}
                                onClick={() => setSelected(crd)}
                                className={`w-full text-left px-4 py-2 text-sm border-b border-gray-800 hover:bg-gray-700/60 transition-colors ${selected?.name === crd.name && selected?.version === crd.version ? 'bg-gray-700 text-white' : 'text-gray-300'}`}
                            >
                                <div className="flex items-center justify-between">
                                    <span className="font-medium">{crd.kind}</span>
                                    <span className="text-xs text-gray-500">{crd.scope}</span>
                                </div>
                                <div className="text-xs text-gray-500">{crd.group ? `${crd.group}/${crd.version}` : crd.version}</div>
                            </button>
                        ))}
                        {filteredCRDs.length === 0 && (
                            <div className="px-4 py-3 text-xs text-gray-500">No CRDs match this filter.</div>
                        )}
                    </div>
                </div>

                <div className="lg:col-span-2 bg-gray-800 border border-gray-700 rounded-lg overflow-hidden min-h-[70vh]">
                    <div className="px-4 py-3 border-b border-gray-700 flex items-center justify-between">
                        <div className="text-gray-300 text-sm font-medium">
                            {selected ? `${selected.kind} Resources` : 'Select a CRD'}
                        </div>
                        {loadingResources && <RefreshCw size={16} className="animate-spin text-gray-400" />}
                    </div>
                    {selected ? (
                        <div className="overflow-x-auto">
                            <table className="min-w-full">
                                <thead className="bg-gray-900">
                                    <tr>
                                        <th className="px-3 md:px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Name</th>
                                        {selected.scope === 'Namespaced' && (
                                            <th className="px-2 md:px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Namespace</th>
                                        )}
                                        <th className="px-2 md:px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Status</th>
                                        <th className="px-2 md:px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Created</th>
                                        <th className="px-2 md:px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Actions</th>
                                    </tr>
                                </thead>
                                <tbody className="bg-gray-800 divide-y divide-gray-700">
                                    {resources.map((resource) => (
                                        <tr key={`${resource.namespace}-${resource.name}`}>
                                            <td className="px-3 md:px-4 py-2 text-sm text-gray-200">{resource.name}</td>
                                            {selected.scope === 'Namespaced' && (
                                                <td className="px-2 md:px-4 py-2 text-sm text-gray-300">{resource.namespace || '-'}</td>
                                            )}
                                            <td className="px-2 md:px-4 py-2">
                                                {resource.status ? (
                                                    <span className={`px-2 py-1 text-xs rounded-full ${getStatusBadgeClass(resource.status)}`}>
                                                        {resource.status}
                                                    </span>
                                                ) : (
                                                    <span className="text-sm text-gray-300">—</span>
                                                )}
                                            </td>
                                            <td className="px-2 md:px-4 py-2 text-sm text-gray-400">
                                                {resource.created ? formatDate(resource.created) : '—'}
                                            </td>
                                            <td className="px-2 md:px-4 py-2 text-sm text-gray-300">
                                                <button
                                                    onClick={() => handleViewYaml(resource)}
                                                    className="flex items-center px-2.5 py-1 bg-gray-800 hover:bg-gray-700 text-gray-200 rounded-md border border-gray-700 text-xs transition-colors"
                                                >
                                                    <FileText size={12} className="mr-1.5" />
                                                    View YAML
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                    {resources.length === 0 && (
                                        <tr>
                                            <td
                                                colSpan={selected.scope === 'Namespaced' ? 5 : 4}
                                                className="px-4 py-6 text-center text-sm text-gray-500"
                                            >
                                                No resources found in {selected.scope === 'Namespaced' ? `namespace "${namespace}"` : 'cluster'}.
                                            </td>
                                        </tr>
                                    )}
                                </tbody>
                            </table>
                        </div>
                    ) : (
                        <div className="p-6 text-gray-500 text-sm">Choose a Custom Resource Definition to explore its resources.</div>
                    )}
                </div>
            </div>
            {yamlView && (
                <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 w-full max-w-4xl h-[80vh] rounded-lg border border-gray-700 flex flex-col shadow-2xl overflow-hidden">
                        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-700 bg-gray-800">
                            <div className="flex items-center space-x-2">
                                <FileText size={16} className="text-gray-400" />
                                <span className="font-mono text-sm text-gray-200">{yamlView.name}</span>
                                {yamlView.namespace && <span className="text-xs text-gray-500">({yamlView.namespace})</span>}
                            </div>
                            <button
                                onClick={() => setYamlView(null)}
                                className="px-3 py-1 text-sm text-gray-200 bg-gray-800 hover:bg-gray-700 rounded border border-gray-700"
                            >
                                Close
                            </button>
                        </div>
                        <div className="flex-1 relative">
                            <Editor
                                height="100%"
                                defaultLanguage="yaml"
                                theme="vs-dark"
                                value={yamlView.content}
                                options={{
                                    readOnly: true,
                                    minimap: { enabled: false },
                                    scrollBeyondLastLine: false,
                                    fontSize: 12,
                                    automaticLayout: true,
                                }}
                            />
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default CRDExplorer;
