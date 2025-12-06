import React, { useState, useEffect } from 'react';
import { ArrowUp, X, Info, RefreshCw } from 'lucide-react';
import Editor from '@monaco-editor/react';
import { defineMonacoTheme } from '../../config/monacoTheme';
import { useAuth } from '../../context/AuthContext';
import { parseErrorResponse } from '../../utils/errorParser';
import { useToast } from '../../context/ToastContext';

const HelmUpgradeModal = ({ release, isOpen, onClose, onSuccess, currentCluster }) => {
    const { authFetch } = useAuth();
    const toast = useToast();
    const [form, setForm] = useState({
        chart: '',
        version: '',
        repo: '',
        valuesYaml: ''
    });
    const [upgrading, setUpgrading] = useState(false);

    useEffect(() => {
        if (release) {
            setForm({
                chart: release.chart || '',
                version: '',
                repo: '',
                valuesYaml: ''
            });
        }
    }, [release]);

    const handleEditorWillMount = (monaco) => {
        defineMonacoTheme(monaco);
    };

    const handleUpgrade = async () => {
        if (!release) return;

        setUpgrading(true);
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const payload = {
                name: release.name,
                namespace: release.namespace,
                chart: form.chart || release.chart,
                version: form.version || undefined,
                repo: form.repo || undefined,
                valuesYaml: form.valuesYaml || undefined
            };

            const res = await authFetch(`/api/helm/releases?${params.toString()}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            });

            if (!res.ok) {
                const errorText = await parseErrorResponse(res);
                throw new Error(errorText || 'Failed to upgrade Helm release');
            }

            const result = await res.json();

            // Show success message
            toast.success(`Upgrade initiated! Job: ${result.job || 'created'}`);

            onSuccess();
            onClose();
        } catch (err) {
            toast.error(`Error upgrading Helm release: ${err.message}`);
        } finally {
            setUpgrading(false);
        }
    };

    if (!isOpen || !release) return null;

    return (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 border border-gray-700 rounded-lg w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col shadow-xl">
                {/* Header */}
                <div className="flex items-center justify-between px-6 py-4 border-b border-gray-700 bg-gray-800">
                    <div className="flex items-center space-x-3">
                        <ArrowUp className="text-blue-400" size={20} />
                        <div>
                            <h3 className="text-lg font-semibold text-white">
                                Upgrade Helm Release
                            </h3>
                            <p className="text-xs text-gray-400 mt-0.5">
                                {release.name} <span className="text-gray-500">•</span> {release.namespace}
                            </p>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-1.5 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                    >
                        <X size={20} />
                    </button>
                </div>

                {/* Current Release Info */}
                <div className="px-6 py-3 bg-gray-800/50 border-b border-gray-700">
                    <div className="flex items-start space-x-2">
                        <Info size={16} className="text-blue-400 mt-0.5 flex-shrink-0" />
                        <div className="text-xs text-gray-400">
                            <span className="text-gray-300 font-medium">Current:</span> Chart: <span className="text-white">{release.chart || 'N/A'}</span>
                            {release.version && (
                                <> • Version: <span className="text-white">{release.version}</span></>
                            )}
                            {release.revision && (
                                <> • Revision: <span className="text-white">{release.revision}</span></>
                            )}
                        </div>
                    </div>
                </div>

                {/* Content */}
                <div className="p-6 overflow-y-auto flex-1">
                    <div className="space-y-5">
                        {/* Chart */}
                        <div>
                            <label className="block text-sm font-medium text-gray-300 mb-2">
                                Chart Name <span className="text-red-400">*</span>
                            </label>
                            <input
                                type="text"
                                value={form.chart}
                                onChange={(e) => setForm({ ...form, chart: e.target.value })}
                                placeholder="e.g., nginx, vault, prometheus"
                                className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            />
                            {release.chart && !form.chart && (
                                <p className="text-xs text-gray-500 mt-1.5 flex items-center">
                                    <Info size={12} className="mr-1" />
                                    Will use current chart: <span className="text-gray-300 ml-1">{release.chart}</span>
                                </p>
                            )}
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                            {/* Version */}
                            <div>
                                <label className="block text-sm font-medium text-gray-300 mb-2">
                                    Chart Version <span className="text-gray-500 text-xs">(optional)</span>
                                </label>
                                <input
                                    type="text"
                                    value={form.version}
                                    onChange={(e) => setForm({ ...form, version: e.target.value })}
                                    placeholder="e.g., 1.2.3 (latest if empty)"
                                    className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                />
                                {release.version && (
                                    <p className="text-xs text-gray-500 mt-1.5">Current: {release.version}</p>
                                )}
                            </div>

                            {/* Repo */}
                            <div>
                                <label className="block text-sm font-medium text-gray-300 mb-2">
                                    Repository URL <span className="text-gray-500 text-xs">(optional)</span>
                                </label>
                                <input
                                    type="text"
                                    value={form.repo}
                                    onChange={(e) => setForm({ ...form, repo: e.target.value })}
                                    placeholder="e.g., https://charts.helm.sh/stable"
                                    className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                />
                            </div>
                        </div>

                        {/* Values YAML */}
                        <div>
                            <label className="block text-sm font-medium text-gray-300 mb-2">
                                Values Override (YAML) <span className="text-gray-500 text-xs">(optional)</span>
                            </label>
                            <div className="border border-gray-700 rounded-md overflow-hidden monaco-editor-container" style={{ height: '300px' }}>
                                <Editor
                                    height="100%"
                                    defaultLanguage="yaml"
                                    theme="dkonsole-dark"
                                    beforeMount={handleEditorWillMount}
                                    value={form.valuesYaml}
                                    onChange={(value) => setForm({ ...form, valuesYaml: value || '' })}
                                    options={{
                                        minimap: { enabled: false },
                                        scrollBeyondLastLine: false,
                                        fontSize: 13,
                                        automaticLayout: true,
                                        lineNumbers: 'on',
                                        wordWrap: 'on',
                                        tabSize: 2,
                                        insertSpaces: true,
                                    }}
                                />
                            </div>
                            <p className="text-xs text-gray-500 mt-1.5">
                                Override chart values. These will be merged with the default values.
                            </p>
                        </div>

                        {/* Help Hint */}
                        <div className="mt-4 p-4 bg-blue-900/20 border border-blue-700/50 rounded-md">
                            <div className="flex items-start space-x-2">
                                <Info size={16} className="text-blue-400 mt-0.5 flex-shrink-0" />
                                <div className="text-xs text-blue-300">
                                    <span className="font-medium">💡 Tip:</span> For best results, specify the <span className="font-semibold text-blue-200">Repository URL</span> and <span className="font-semibold text-blue-200">Chart Version</span>. If unspecified, the system will try to auto-detect them, which might take longer.
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Footer */}
                <div className="px-6 py-4 border-t border-gray-700 bg-gray-800 flex justify-between items-center">
                    <div className="text-xs text-gray-500">
                        A Kubernetes Job will be created to execute the upgrade
                    </div>
                    <div className="flex space-x-3">
                        <button
                            onClick={onClose}
                            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                            disabled={upgrading}
                        >
                            Cancel
                        </button>
                        <button
                            onClick={handleUpgrade}
                            disabled={upgrading || !form.chart}
                            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-md transition-colors flex items-center"
                        >
                            {upgrading ? (
                                <>
                                    <RefreshCw size={16} className="mr-2 animate-spin" />
                                    Upgrading...
                                </>
                            ) : (
                                <>
                                    <ArrowUp size={16} className="mr-2" />
                                    Upgrade Release
                                </>
                            )}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default HelmUpgradeModal;
