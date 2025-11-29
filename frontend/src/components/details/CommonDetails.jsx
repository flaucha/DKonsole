import React, { useState } from 'react';
import { Eye, EyeOff, FileText } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { canEdit } from '../../utils/permissions';

export const SmartImage = ({ image }) => {
    const [copied, setCopied] = useState(false);

    const handleClick = (e) => {
        e.stopPropagation();
        // Extract SHA if present, otherwise copy full image string
        const shaMatch = image.match(/@sha256:([a-f0-9]+)/);
        const textToCopy = shaMatch ? image : image;
        // Requirement says "click on shortened hash copies it".
        // Actually "click on shortened hash that it copies AND notifies".
        // If I copy the FULL image string it's safer for reuse.

        navigator.clipboard.writeText(textToCopy);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    if (!image) return null;

    const parts = image.split('@sha256:');
    const isSha = parts.length > 1;
    const displayText = isSha ? (
        <span>
            {parts[0]}@sha256:<span className="font-bold text-blue-300 hover:underline">{parts[1].substring(0, 8)}...</span>
        </span>
    ) : image;

    return (
        <div className="relative inline-block group">
            <span
                className={`cursor-pointer hover:text-blue-400 transition-colors ${isSha ? 'font-mono' : ''}`}
                onClick={handleClick}
                title={image} // Show full image on hover
            >
                {displayText}
            </span>
            {copied && (
                <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-green-600 text-white text-xs rounded shadow-lg animate-fade-out pointer-events-none whitespace-nowrap z-50 border border-green-500">
                    Version copied
                </div>
            )}
        </div>
    );
};

export const SmartDNS = ({ dns }) => {
    const [copied, setCopied] = useState(false);

    const handleClick = (e) => {
        e.stopPropagation();
        navigator.clipboard.writeText(dns);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    if (!dns) return null;

    return (
        <div className="relative inline-block group">
            <span
                className="cursor-pointer hover:text-blue-400 transition-colors font-mono"
                onClick={handleClick}
                title={dns} // Show full DNS on hover
            >
                {dns}
            </span>
            {copied && (
                <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-green-600 text-white text-xs rounded shadow-lg animate-fade-out pointer-events-none whitespace-nowrap z-50 border border-green-500">
                    DNS copied
                </div>
            )}
        </div>
    );
};

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

export const EditYamlButton = ({ onClick, namespace }) => {
    const { user } = useAuth();
    // Only show if user has edit permission or is admin
    if (!canEdit(user, namespace)) {
        return null;
    }
    return (
        <button
            onClick={onClick}
            className="flex items-center px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-gray-200 rounded-md border border-gray-600 text-xs transition-colors font-medium"
        >
            <FileText size={14} className="mr-1.5" />
            Edit YAML
        </button>
    );
};
