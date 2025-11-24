import React, { useState } from 'react';
import { Eye, EyeOff, FileText } from 'lucide-react';

export const DetailRow = ({ label, value, icon: Icon, children }) => (
    <div className="flex items-center justify-between bg-gray-800 px-3 py-2 rounded border border-gray-700 mb-2">
        <div className="flex items-center">
            {Icon && <Icon size={14} className="mr-2 text-gray-500" />}
            <span className="text-xs text-gray-400">{label}</span>
        </div>
        <div className="flex items-center">
            <span className="text-sm font-mono text-white break-all text-right">
                {Array.isArray(value) ? (
                    value.length > 0 ? value.join(', ') : <span className="text-gray-600 italic">None</span>
                ) : (
                    value || <span className="text-gray-600 italic">None</span>
                )}
            </span>
            {children}
        </div>
    </div>
);

export const DataRow = ({ label, value, isSecret }) => {
    const [revealed, setRevealed] = useState(!isSecret);
    return (
        <div className="bg-gray-800 p-2 rounded border border-gray-700">
            <div className="flex justify-between items-start">
                <span className="text-xs font-medium text-gray-400 mb-1 block">{label}</span>
                {isSecret && (
                    <button
                        onClick={() => setRevealed(!revealed)}
                        className="text-gray-500 hover:text-gray-300 focus:outline-none"
                        title={revealed ? 'Hide value' : 'Show value'}
                    >
                        {revealed ? <EyeOff size={14} /> : <Eye size={14} />}
                    </button>
                )}
            </div>
            <div className="text-sm font-mono text-gray-300 break-all whitespace-pre-wrap">
                {revealed ? value : '••••••••'}
            </div>
        </div>
    );
};

export const DataSection = ({ data, isSecret = false }) => {
    if (!data || Object.keys(data).length === 0) return <div className="text-gray-500 italic text-sm">No data.</div>;
    return (
        <div className="mt-2 space-y-2">
            {Object.entries(data).map(([key, value]) => (
                <DataRow key={key} label={key} value={value} isSecret={isSecret} />
            ))}
        </div>
    );
};

export const EditYamlButton = ({ onClick }) => (
    <button
        onClick={onClick}
        className="flex items-center px-2.5 py-1 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-md border border-gray-600 text-xs transition-colors"
    >
        <FileText size={12} className="mr-1.5" />
        Edit YAML
    </button>
);
