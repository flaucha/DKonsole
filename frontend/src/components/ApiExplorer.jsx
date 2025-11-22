import React, { useEffect, useState, useMemo } from 'react';
import { Search, ListTree, RefreshCw, Globe, MapPin, FileText } from 'lucide-react';
import Editor from '@monaco-editor/react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';

const ApiExplorer = ({ namespace }) => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [apis, setApis] = useState([]);
    const [filter, setFilter] = useState('');
    const [selected, setSelected] = useState(null);
    const [objects, setObjects] = useState([]);
    const [loadingApis, setLoadingApis] = useState(false);
    const [loadingObjects, setLoadingObjects] = useState(false);
    const [scopeFilter, setScopeFilter] = useState('namespaced'); // namespaced | cluster | all
    const [yamlView, setYamlView] = useState(null); // {name, namespace, content}

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
            namespace: selected.namespaced ? namespace : '',
            namespaced: selected.namespaced ? 'true' : 'false',
        });
        if (currentCluster) params.append('cluster', currentCluster);
        setObjects([]);
        setLoadingObjects(true);
        authFetch(`/api/apis/resources?${params.toString()}`)
            .then(res => res.json())
            .then(data => {
                setObjects(data || []);
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
                return matchesText && matchesScope;
            })
            .sort((a, b) => a.kind.localeCompare(b.kind));
    }, [apis, filter, scopeFilter]);

    const handleViewYaml = (obj) => {
        if (!selected) return;
        const params = new URLSearchParams({
            group: selected.group,
            version: selected.version,
            resource: selected.resource,
            name: obj.name,
            namespace: selected.namespaced ? obj.namespace || '' : '',
            namespaced: selected.namespaced ? 'true' : 'false',
        });
        if (currentCluster) params.append('cluster', currentCluster);
        authFetch(`/api/apis/yaml?${params.toString()}`)
            .then((res) => res.text())
            .then((yaml) => setYamlView({ name: obj.name, namespace: obj.namespace, content: yaml }))
            .catch((err) => setYamlView({ name: obj.name, namespace: obj.namespace, content: `# Error: ${err.message}` }));
    };

    return (
        <div className="p-6">
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center space-x-2">
                    <ListTree className="text-blue-400" size={18} />
                    <h1 className="text-xl font-semibold text-white">API Explorer</h1>
                    {loadingApis && <RefreshCw size={16} className="animate-spin text-gray-400" />}
                </div>
                <div className="flex items-center space-x-3">
                    <div className="relative w-64">
                        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
                        <input
                            value={filter}
                            onChange={(e) => setFilter(e.target.value)}
                            className="w-full bg-gray-800 border border-gray-700 rounded-md pl-8 pr-3 py-1.5 text-sm text-gray-200 focus:outline-none focus:border-blue-500"
                            placeholder="Search kind/resource..."
                        />
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
                                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Name</th>
                                        {selected.namespaced && (
                                            <th className="px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Namespace</th>
                                        )}
                                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Status</th>
                                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Created</th>
                                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Actions</th>
                                    </tr>
                                </thead>
                                <tbody className="bg-gray-800 divide-y divide-gray-700">
                                    {objects.map((obj) => (
                                        <tr key={`${obj.namespace}-${obj.name}`}>
                                            <td className="px-4 py-2 text-sm text-gray-200">{obj.name}</td>
                                            {selected.namespaced && (
                                                <td className="px-4 py-2 text-sm text-gray-300">{obj.namespace || '-'}</td>
                                            )}
                                            <td className="px-4 py-2 text-sm text-gray-300">{obj.status || '—'}</td>
                                            <td className="px-4 py-2 text-sm text-gray-400">
                                                {obj.created ? new Date(obj.created).toLocaleDateString() : '—'}
                                            </td>
                                            <td className="px-4 py-2 text-sm text-gray-300">
                                                <button
                                                    onClick={() => handleViewYaml(obj)}
                                                    className="flex items-center px-2.5 py-1 bg-gray-800 hover:bg-gray-700 text-gray-200 rounded-md border border-gray-700 text-xs transition-colors"
                                                >
                                                    <FileText size={12} className="mr-1.5" />
                                                    Ver YAML
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

export default ApiExplorer;
