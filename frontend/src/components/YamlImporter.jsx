import React, { useState } from 'react';
import { Upload, X, Loader2, AlertTriangle } from 'lucide-react';
import Editor from '@monaco-editor/react';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { parseErrorResponse, parseError } from '../utils/errorParser';

const YamlImporter = ({ onClose }) => {
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [content, setContent] = useState('');
    const [importing, setImporting] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    const handleImport = () => {
        setImporting(true);
        setError('');
        setSuccess('');

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
                    throw new Error(text || 'Failed to import');
                }
                const data = await res.json();
                setSuccess(`Imported ${data.count} resource(s): ${data.resources.join(', ')}`);
                setContent('');
            })
            .catch((err) => setError(parseError(err)))
            .finally(() => setImporting(false));
    };

    return (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 w-full max-w-4xl h-[80vh] rounded-lg border border-gray-700 flex flex-col shadow-2xl overflow-hidden">
                <div className="flex items-center justify-between px-4 py-3 border-b border-gray-700 bg-gray-800">
                    <div className="flex items-center space-x-2">
                        <Upload size={18} className="text-gray-400" />
                        <span className="text-sm font-semibold text-white">Import YAML</span>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-1.5 hover:bg-red-900/50 rounded text-gray-400 hover:text-red-400 transition-colors"
                    >
                        <X size={18} />
                    </button>
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

                {success && (
                    <div className="bg-green-900/20 text-green-200 px-4 py-3 flex items-start justify-between border-b border-green-800">
                        <div className="flex items-start space-x-2">
                            <span className="text-sm whitespace-pre-wrap">{success}</span>
                        </div>
                        <button
                            onClick={() => setSuccess('')}
                            className="text-xs px-2 py-1 border border-green-700 rounded hover:bg-green-800/50 transition-colors"
                        >
                            Dismiss
                        </button>
                    </div>
                )}

                <div className="flex-1 flex flex-col relative">
                    <Editor
                        height="100%"
                        defaultLanguage="yaml"
                        theme="vs-dark"
                        value={content}
                        onChange={(value) => setContent(value)}
                        options={{
                            minimap: { enabled: false },
                            scrollBeyondLastLine: false,
                            fontSize: 14,
                            automaticLayout: true,
                        }}
                    />
                    <div className="p-4 border-t border-gray-800 flex justify-end bg-gray-900">
                        <button
                            onClick={handleImport}
                            disabled={importing || !content.trim()}
                            className="flex items-center px-3 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 disabled:cursor-not-allowed text-white rounded-md text-sm transition-colors"
                        >
                            {importing ? <Loader2 size={16} className="animate-spin mr-2" /> : <Upload size={16} className="mr-2" />}
                            Apply YAML
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default YamlImporter;
