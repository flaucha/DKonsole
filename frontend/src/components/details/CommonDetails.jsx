import React, { useState } from 'react';
import { Eye, EyeOff, FileText } from 'lucide-react';

export const DetailRow = ({ label, value, icon: Icon, children }) => (
    <div className="flex items-center justify-between bg-gray-800/50 px-4 py-3 rounded-md border border-gray-700/50 mb-3 hover:bg-gray-800/70 transition-colors">
        <div className="flex items-center min-w-0 flex-1">
            {Icon && <Icon size={16} className="mr-2 text-gray-400 flex-shrink-0" />}
            <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">{label}</span>
        </div>
        <div className="flex items-center ml-4 min-w-0 flex-1 justify-end">
            <span className="text-sm text-gray-200 break-words text-right">
                {Array.isArray(value) ? (
                    value.length > 0 ? (
                        <span className="inline-flex flex-wrap gap-1">
                            {value.map((v, i) => (
                                <span key={i} className="px-2 py-1 bg-gray-700/50 rounded text-xs font-mono">
                                    {v}
                                </span>
                            ))}
                        </span>
                    ) : <span className="text-gray-500 italic">None</span>
                ) : (
                    value || <span className="text-gray-500 italic">None</span>
                )}
            </span>
            {children}
        </div>
    </div>
);

export const DataRow = ({ label, value, isSecret }) => {
    const [revealed, setRevealed] = useState(!isSecret);
    return (
        <div className="bg-gray-800/50 p-3 rounded-md border border-gray-700/50 hover:bg-gray-800/70 transition-colors">
            <div className="flex justify-between items-start mb-2">
                <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">{label}</span>
                {isSecret && (
                    <button
                        onClick={() => setRevealed(!revealed)}
                        className="text-gray-500 hover:text-gray-300 focus:outline-none transition-colors"
                        title={revealed ? 'Hide value' : 'Show value'}
                    >
                        {revealed ? <EyeOff size={14} /> : <Eye size={14} />}
                    </button>
                )}
            </div>
            <div className="text-sm font-mono text-gray-200 break-all whitespace-pre-wrap">
                {revealed ? value : '••••••••'}
            </div>
        </div>
    );
};

export const DataSection = ({ data, isSecret = false }) => {
    if (!data || Object.keys(data).length === 0) return <div className="text-gray-500 italic text-sm py-2">No data.</div>;
    return (
        <div className="space-y-3">
            {Object.entries(data).map(([key, value]) => (
                <DataRow key={key} label={key} value={value} isSecret={isSecret} />
            ))}
        </div>
    );
};

export const EditYamlButton = ({ onClick }) => (
    <button
        onClick={onClick}
        className="flex items-center px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-gray-200 rounded-md border border-gray-600 text-xs transition-colors font-medium"
    >
        <FileText size={14} className="mr-1.5" />
        Edit YAML
    </button>
);
