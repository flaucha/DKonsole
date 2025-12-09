import React from 'react';
import { X, Save, FileText, Copy, Loader2 } from 'lucide-react';

const YamlToolbar = ({
    kind,
    name,
    namespace,
    isNew,
    loading,
    saving,
    onSave,
    onCopy,
    onClose,
    copying
}) => {
    return (
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
                {isNew && <span className="text-xs text-blue-400">(New)</span>}
            </div>
            <div className="flex items-center space-x-2">
                <button
                    onClick={onSave}
                    disabled={loading || saving}
                    className="flex items-center px-3 py-1.5 bg-gray-700 hover:bg-gray-600 disabled:opacity-60 disabled:cursor-not-allowed text-white rounded-md text-sm transition-colors"
                >
                    {saving ? <Loader2 size={16} className="animate-spin mr-2" /> : <Save size={16} className="mr-2" />}
                    Apply changes
                </button>
                <button
                    onClick={onCopy}
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
    );
};

export default YamlToolbar;
