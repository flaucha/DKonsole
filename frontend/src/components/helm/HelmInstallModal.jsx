import React, { useState, useEffect } from 'react';
import { Download, X, Info, RefreshCw } from 'lucide-react';
import Editor from '@monaco-editor/react';
import { defineMonacoTheme } from '../../config/monacoTheme';
import { useAuth } from '../../context/AuthContext';
import { parseErrorResponse } from '../../utils/errorParser';
import { useToast } from '../../context/ToastContext';

const HelmInstallModal = ({ isOpen, onClose, onSuccess, currentCluster, prefilledNamespace }) => {
    const { authFetch } = useAuth();
    const toast = useToast();
    const [form, setForm] = useState({
        name: '',
        namespace: '',
        chart: '',
        version: '',
        repo: '',
        valuesYaml: ''
    });
    const [installing, setInstalling] = useState(false);

    useEffect(() => {
        if (isOpen) {
            setForm(prev => ({ ...prev, namespace: prefilledNamespace || '' }));
        }
    }, [isOpen, prefilledNamespace]);

    const handleEditorWillMount = (monaco) => {
        defineMonacoTheme(monaco);
    };

    const handleInstall = async () => {
        if (!form.name || !form.namespace || !form.chart) {
            toast.warning('Please fill in all required fields (Name, Namespace, Chart)');
            return;
        }

        setInstalling(true);
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const payload = {
                name: form.name,
                namespace: form.namespace,
                chart: form.chart,
                version: form.version || undefined,
                repo: form.repo || undefined,
                valuesYaml: form.valuesYaml || undefined
            };

            const res = await authFetch(`/api/helm/releases/install?${params.toString()}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            });

            if (!res.ok) {
                const errorText = await parseErrorResponse(res);
                throw new Error(errorText || 'Failed to install Helm chart');
            }

            const result = await res.json();

            // Show success message
            toast.success(`Installation initiated! Job: ${result.job || 'created'}`);

            onSuccess();
            onClose();
            // Reset form
            setForm({ name: '', namespace: '', chart: '', version: '', repo: '', valuesYaml: '' });
        } catch (err) {
            toast.error(`Error installing Helm chart: ${err.message}`);
        } finally {
            setInstalling(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 border border-gray-700 rounded-lg w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col shadow-xl">
                {/* Header */}
                <div className="flex items-center justify-between px-6 py-4 border-b border-gray-700 bg-gray-800">
                    <div className="flex items-center space-x-3">
                        <Download className="text-blue-400" size={20} />
                        <div>
                            <h3 className="text-lg font-semibold text-white">
                                Install Helm Chart
                            </h3>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-1.5 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                    >
                        <X size={20} />
                    </button>
                </div>

                {/* Content */}
                <div className="p-6 overflow-y-auto flex-1">
                    <div className="space-y-5">
                        {/* Name and Namespace */}
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                            <div>
                                <label className="block text-sm font-medium text-gray-300 mb-2">
                                    Release Name <span className="text-red-400">*</span>
                                </label>
                                <input
                                    type="text"
                                    value={form.name}
                                    onChange={(e) => setForm({ ...form, name: e.target.value })}
                                    placeholder="e.g., my-app, nginx, prometheus"
                                    className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-300 mb-2">
                                    Namespace <span className="text-red-400">*</span>
                                </label>
                                <input
                                    type="text"
                                    value={form.namespace}
                                    onChange={(e) => setForm({ ...form, namespace: e.target.value })}
                                    placeholder="e.g., default, production, monitoring"
                                    className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                />
                            </div>
                        </div>

                        {/* Chart */}
                        <div>
                            <label className="block text-sm font-medium text-gray-300 mb-2">
                                Chart Name <span className="text-red-400">*</span>
                            </label>
                            <input
                                type="text"
                                value={form.chart}
                                onChange={(e) => setForm({ ...form, chart: e.target.value })}
                                placeholder="e.g., nginx, vault, prometheus, kube-prometheus-stack"
                                className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            />
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
                        A Kubernetes Job will be created to execute the installation
                    </div>
                    <div className="flex space-x-3">
                        <button
                            onClick={onClose}
                            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                            disabled={installing}
                        >
                            Cancel
                        </button>
                        <button
                            onClick={handleInstall}
                            disabled={installing || !form.name || !form.namespace || !form.chart}
                            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-md transition-colors flex items-center"
                        >
                            {installing ? (
                                <>
                                    <RefreshCw size={16} className="mr-2 animate-spin" />
                                    Installing...
                                </>
                            ) : (
                                <>
                                    <Download size={16} className="mr-2" />
                                    Install Chart
                                </>
                            )}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default HelmInstallModal;
