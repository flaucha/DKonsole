import React, { useEffect, useState } from 'react';
import { X, Save, AlertTriangle, Loader2, FileText, Copy } from 'lucide-react';
import Editor from '@monaco-editor/react';
import { defineMonacoTheme } from '../config/monacoTheme';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { parseErrorResponse, parseError } from '../utils/errorParser';

const YamlEditor = ({ resource, onClose, onSaved }) => {
    const { name, namespace, kind, group, version, resource: resourceName, namespaced } = resource || {};
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [content, setContent] = useState('');
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState('');
    const [copying, setCopying] = useState(false);

    const handleEditorWillMount = (monaco) => {
        defineMonacoTheme(monaco);
    };

    const handleEditorDidMount = (editor, monaco) => {
        // Ensure font metrics are measured after the font is available to avoid cursor drift
        monaco.editor.remeasureFonts();
        editor.updateOptions({
            fontFamily: '"Fira Code", "JetBrains Mono", "Cascadia Code", Menlo, Consolas, "Courier New", monospace',
            fontLigatures: true,
            lineHeight: 22,
            letterSpacing: 0,
        });
    };

    const buildUrl = () => {
        const params = new URLSearchParams({ kind, name });
        if (namespace) params.append('namespace', namespace);
        if (group) params.append('group', group);
        if (version) params.append('version', version);
        if (resourceName) params.append('resource', resourceName);
        if (namespaced !== undefined) params.append('namespaced', namespaced.toString());
        if (currentCluster) params.append('cluster', currentCluster);
        return `/api/resource/yaml?${params.toString()}`;
    };

    useEffect(() => {
        if (!resource) return;

        // If it's a new resource, provide a template
        if (resource.isNew) {
            let template = '';
            if (kind === 'Namespace') {
                template = `apiVersion: v1
kind: Namespace
metadata:
  name: example-namespace
  labels:
    app: example
`;
            } else {
                // Generic template for other resources
                template = `apiVersion: v1
kind: ${kind}
metadata:
  name: example-${kind.toLowerCase()}
${namespace ? `  namespace: ${namespace}` : ''}
`;
            }
            setContent(template);
            setLoading(false);
            return;
        }

        setLoading(true);
        setError('');
        authFetch(buildUrl())
            .then(async (res) => {
                if (!res.ok) {
                    const text = await parseErrorResponse(res);
                    throw new Error(text || 'Failed to load resource');
                }
                return res.text();
            })
            .then((yaml) => {
                setContent(yaml);
                setLoading(false);
            })
            .catch((err) => {
                setError(parseError(err));
                setLoading(false);
            });
    }, [resource, currentCluster]);

    const handleSave = () => {
        setSaving(true);
        setError('');

        // Use import endpoint which uses Server-Side Apply (equivalent to kubectl apply -f)
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        authFetch(`/api/resource/import?${params.toString()}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/x-yaml' },
            body: content,
        })
            .then(async (res) => {
                if (!res.ok) {
                    const text = await parseErrorResponse(res);
                    throw new Error(text || 'Failed to apply resource');
                }
                setSaving(false);
                onSaved?.();
            })
            .catch((err) => {
                setError(parseError(err));
                setSaving(false);
            });
    };

    const handleCopy = async () => {
        try {
            setCopying(true);
            await navigator.clipboard.writeText(content);
            setTimeout(() => setCopying(false), 800);
        } catch {
            setCopying(false);
        }
    };

    return (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 w-full max-w-5xl h-[85vh] rounded-lg border border-gray-700 flex flex-col shadow-2xl overflow-hidden">
                {/* Header */}
                <div className="flex items-center justify-between px-4 py-3 border-b border-gray-700 bg-gray-800">
                    <div className="flex items-center space-x-2">
                        <FileText size={18} className="text-gray-400" />
                        <span className="font-mono text-sm text-gray-200">{kind}</span>
                        {name && (
                            <>
                                <span className="text-gray-500">/</span>
                                <span className="font-mono text-sm text-gray-200">{name}</span>
                            </>
                        )}
                        {namespace && <span className="text-xs text-gray-500">({namespace})</span>}
                        {resource?.isNew && <span className="text-xs text-blue-400">(New)</span>}
                    </div>
                    <div className="flex items-center space-x-2">
                        <button
                            onClick={handleSave}
                            disabled={loading || saving}
                            className="flex items-center px-3 py-1.5 bg-gray-700 hover:bg-gray-600 disabled:opacity-60 disabled:cursor-not-allowed text-white rounded-md text-sm transition-colors"
                        >
                            {saving ? <Loader2 size={16} className="animate-spin mr-2" /> : <Save size={16} className="mr-2" />}
                            Apply changes
                        </button>
                        <button
                            onClick={handleCopy}
                            disabled={loading}
                            className="flex items-center px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-white rounded-md text-sm transition-colors border border-gray-700 disabled:opacity-60 disabled:cursor-not-allowed"
                            title="Copy YAML to clipboard"
                        >
                            <Copy size={16} className="mr-2" />
                            {copying ? 'Copied' : 'Copy'}
                        </button>
                        <button
                            onClick={onClose}
                            className="p-1.5 hover:bg-red-900/50 rounded text-gray-400 hover:text-red-400 transition-colors"
                        >
                            <X size={18} />
                        </button>
                    </div>
                </div>

                {error && (
                    <div className="bg-red-900/30 text-red-200 px-4 py-3 flex items-start justify-between border-b border-red-800">
                        <div className="flex items-start space-x-2">
                            <AlertTriangle size={18} className="mt-0.5" />
                            <span className="text-sm whitespace-pre-wrap">{error}</span>
                        </div>
                        <button
                            onClick={() => setError('')}
                            className="text-xs px-2 py-1 border border-red-700 rounded hover:bg-red-800/50 transition-colors"
                        >
                            Dismiss
                        </button>
                    </div>
                )}

                <div className="flex-1 flex flex-col relative monaco-editor-container" style={{ backgroundColor: '#1f2937' }}>
                    {loading ? (
                        <div className="flex-1 flex items-center justify-center text-gray-400">
                            <Loader2 size={20} className="animate-spin mr-2" />
                            Loading YAML...
                        </div>
                    ) : (
                        <Editor
                            height="100%"
                            defaultLanguage="yaml"
                            theme="dkonsole-dark"
                            beforeMount={handleEditorWillMount}
                            onMount={handleEditorDidMount}
                            value={content}
                            onChange={(value) => setContent(value)}
                            options={{
                                minimap: { enabled: false },
                                scrollBeyondLastLine: false,
                                fontSize: 14,
                                automaticLayout: true,
                            }}
                        />
                    )}
                </div>
            </div>
        </div>
    );
};

export default YamlEditor;
