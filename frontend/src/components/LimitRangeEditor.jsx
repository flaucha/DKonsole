import React, { useEffect, useState } from 'react';
import { X, Save, AlertTriangle, Loader2, Tag } from 'lucide-react';
import Editor from '@monaco-editor/react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';

const LimitRangeEditor = ({ resource, onClose, onSaved }) => {
    const { name, namespace, kind } = resource || {};
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [content, setContent] = useState('');
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState('');

    const buildUrl = () => {
        const params = new URLSearchParams({ kind: 'LimitRange', name });
        if (namespace) params.append('namespace', namespace);
        if (currentCluster) params.append('cluster', currentCluster);
        return `/api/resource/yaml?${params.toString()}`;
    };

    useEffect(() => {
        if (!resource) return;

        // If it's a new resource, provide a template
        if (resource.isNew) {
            const template = `apiVersion: v1
kind: LimitRange
metadata:
  name: example-limitrange
  namespace: ${namespace}
spec:
  limits:
  - max:
      cpu: "2"
      memory: 4Gi
    min:
      cpu: 200m
      memory: 128Mi
    default:
      cpu: 500m
      memory: 512Mi
    defaultRequest:
      cpu: 200m
      memory: 256Mi
    type: Container`;
            setContent(template);
            setLoading(false);
            return;
        }

        setLoading(true);
        setError('');
        authFetch(buildUrl())
            .then(async (res) => {
                if (!res.ok) {
                    const text = await res.text();
                    throw new Error(text || 'Failed to load resource');
                }
                return res.text();
            })
            .then((yaml) => {
                setContent(yaml);
                setLoading(false);
            })
            .catch((err) => {
                setError(err.message);
                setLoading(false);
            });
    }, [resource, currentCluster]);

    const handleSave = () => {
        setSaving(true);
        setError('');

        // For new resources, use import endpoint
        if (resource.isNew) {
            const params = new URLSearchParams();
            if (currentCluster) params.append('cluster', currentCluster);

            authFetch(`/api/resource/import?${params.toString()}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-yaml' },
                body: content,
            })
                .then(async (res) => {
                    if (!res.ok) {
                        const text = await res.text();
                        throw new Error(text || 'Failed to create limit range');
                    }
                    setSaving(false);
                    onSaved?.();
                })
                .catch((err) => {
                    setError(err.message);
                    setSaving(false);
                });
        } else {
            // For existing resources, use update endpoint
            authFetch(buildUrl(), {
                method: 'PUT',
                headers: { 'Content-Type': 'application/x-yaml' },
                body: content,
            })
                .then(async (res) => {
                    if (!res.ok) {
                        const text = await res.text();
                        throw new Error(text || 'Failed to update limit range');
                    }
                    setSaving(false);
                    onSaved?.();
                })
                .catch((err) => {
                    setError(err.message);
                    setSaving(false);
                });
        }
    };

    return (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
            <div className="!bg-gray-900 w-full max-w-4xl h-[85vh] rounded-lg !border-gray-700 border flex flex-col shadow-2xl overflow-hidden">
                {/* Header */}
                <div className="flex items-center justify-between px-4 py-3 !border-gray-700 border-b !bg-gray-800">
                    <div className="flex items-center space-x-2">
                        <Tag size={18} className="text-gray-400" />
                        <span className="font-mono text-sm text-gray-200">LimitRange</span>
                        <span className="text-gray-500">/</span>
                        <span className="font-mono text-sm text-gray-200">{name}</span>
                        {namespace && <span className="text-xs text-gray-500">({namespace})</span>}
                    </div>
                    <div className="flex items-center space-x-2">
                        <button
                            onClick={handleSave}
                            disabled={loading || saving}
                            className="flex items-center px-3 py-1.5 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 disabled:cursor-not-allowed text-white rounded-md text-sm transition-colors"
                        >
                            {saving ? <Loader2 size={16} className="animate-spin mr-2" /> : <Save size={16} className="mr-2" />}
                            Save Changes
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

                <div className="flex-1 flex flex-col relative">
                    {loading ? (
                        <div className="flex-1 flex items-center justify-center text-gray-400">
                            <Loader2 size={20} className="animate-spin mr-2" />
                            Loading YAML...
                        </div>
                    ) : (
                        <Editor
                            height="100%"
                            defaultLanguage="yaml"
                            theme="atom-dark"
                            value={content}
                            onChange={(value) => setContent(value)}
                            options={{
                                minimap: { enabled: false },
                                scrollBeyondLastLine: false,
                                fontSize: 14,
                                automaticLayout: true,
                            }}
                            className="!bg-[#272b34]"
                        />
                    )}
                </div>
            </div>
        </div>
    );
};

export default LimitRangeEditor;
