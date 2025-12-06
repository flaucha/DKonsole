import React, { useState, useEffect } from 'react';
import Editor from '@monaco-editor/react';
import { X, Save, FileJson, AlertCircle } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useSettings } from '../../context/SettingsContext';

const ResourceYamlEditor = ({
    isOpen,
    onClose,
    onSave,
    resourceName,
    namespace,
    resourceType, // Display name e.g. "Limit Range"
    apiEndpoint,  // Base endpoint e.g. "/api/k8s/limitranges"
    templateYaml
}) => {
    const { authFetch } = useAuth();
    const { currentCluster } = useSettings();
    const [yaml, setYaml] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        if (isOpen) {
            if (resourceName) {
                fetchResourceYaml();
            } else {
                // New resource
                setYaml(templateYaml.trim());
                setError(null);
            }
        } else {
            setYaml('');
            setError(null);
        }
    }, [isOpen, resourceName, namespace, currentCluster]);

    const fetchResourceYaml = async () => {
        setLoading(true);
        setError(null);
        try {
            const params = new URLSearchParams({
                namespace: namespace,
                name: resourceName
            });
            if (currentCluster) params.append('cluster', currentCluster);

            const response = await authFetch(`${apiEndpoint}/yaml?${params.toString()}`);
            if (!response.ok) {
                throw new Error('Failed to fetch YAML');
            }
            const data = await response.text();
            setYaml(data);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        setSaving(true);
        setError(null);
        try {
            const params = new URLSearchParams({
                namespace: namespace
            });
            if (resourceName) {
                params.append('name', resourceName);
            }
            if (currentCluster) params.append('cluster', currentCluster);

            const method = resourceName ? 'PUT' : 'POST';
            const url = resourceName
                ? `${apiEndpoint}?${params.toString()}` // Update
                : `${apiEndpoint}?${params.toString()}`; // Create (name usually in body or generated, but endpoint might need adjustment depending on backend implementation. Assuming standard POST to collection)

            // Note: backend implementation for Create usually takes namespace in URL or body. 
            // Existing implementation check: 
            // LimitRangeEditor used: PUT for update, POST for create.

            const response = await authFetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/yaml',
                },
                body: yaml,
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || `Failed to ${resourceName ? 'update' : 'create'} ${resourceType}`);
            }

            onSave();
            onClose();
        } catch (err) {
            setError(err.message);
        } finally {
            setSaving(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
            <div className="bg-gray-800 rounded-lg shadow-xl w-full max-w-4xl h-[80vh] flex flex-col border border-gray-700">
                {/* Header */}
                <div className="flex items-center justify-between p-4 border-b border-gray-700 bg-gray-900/50">
                    <div className="flex items-center space-x-3">
                        <div className="p-2 bg-blue-500/10 rounded-lg">
                            <FileJson className="w-5 h-5 text-blue-400" />
                        </div>
                        <div>
                            <h2 className="text-lg font-semibold text-white">
                                {resourceName ? `Edit ${resourceType}` : `Create ${resourceType}`}
                            </h2>
                            <p className="text-xs text-gray-400">
                                {namespace ? `Namespace: ${namespace}` : 'Cluster Scope'}
                            </p>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-2 hover:bg-gray-700 rounded-lg transition-colors text-gray-400 hover:text-white"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Editor Content */}
                <div className="flex-1 overflow-hidden relative">
                    {loading ? (
                        <div className="absolute inset-0 flex items-center justify-center bg-gray-900/50">
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
                        </div>
                    ) : (
                        <Editor
                            height="100%"
                            defaultLanguage="yaml"
                            value={yaml}
                            onChange={(value) => setYaml(value)}
                            theme="dkonsole-dark"
                            options={{
                                minimap: { enabled: false },
                                fontSize: 13,
                                lineNumbers: 'on',
                                scrollBeyondLastLine: false,
                                automaticLayout: true,
                                renderWhitespace: 'selection',
                                tabSize: 2,
                                scrollbar: {
                                    vertical: 'visible',
                                    horizontal: 'visible'
                                }
                            }}
                        />
                    )}
                </div>

                {/* Error Banner */}
                {error && (
                    <div className="p-3 bg-red-500/10 border-t border-red-500/20 flex items-center gap-2">
                        <AlertCircle className="w-4 h-4 text-red-400 flex-shrink-0" />
                        <span className="text-sm text-red-200 truncate">{error}</span>
                    </div>
                )}

                {/* Footer */}
                <div className="p-4 border-t border-gray-700 bg-gray-900/50 flex justify-end gap-3">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-sm font-medium text-gray-300 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        disabled={saving || loading}
                        className="flex items-center gap-2 px-4 py-2 text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {saving ? (
                            <>
                                <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                Saving...
                            </>
                        ) : (
                            <>
                                <Save className="w-4 h-4" />
                                Save Changes
                            </>
                        )}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default ResourceYamlEditor;
