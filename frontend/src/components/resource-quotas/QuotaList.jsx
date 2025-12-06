import React, { useState } from 'react';
import { Activity, MoreVertical, FileText, Trash2 } from 'lucide-react';
import { calculatePercentage } from '../../utils/resourceParser';

const ProgressBar = ({ used, hard, label }) => {
    const percentage = calculatePercentage(used, hard);
    let colorClass = 'bg-blue-500';
    if (percentage > 90) colorClass = 'bg-red-500';
    else if (percentage > 75) colorClass = 'bg-yellow-500';

    return (
        <div className="mb-2 last:mb-0">
            <div className="flex justify-between text-xs mb-0.5">
                <span className="text-gray-400 font-medium">{label}</span>
                <span className="text-gray-300">
                    {used} / {hard} <span className="ml-1 text-gray-500">({percentage}%)</span>
                </span>
            </div>
            <div className="w-full bg-gray-700 rounded-full h-1.5 overflow-hidden">
                <div
                    className={`h-1.5 rounded-full transition-all duration-500 ${colorClass}`}
                    style={{ width: `${percentage}%` }}
                ></div>
            </div>
        </div>
    );
};

const renderQuotaUsage = (resource) => {
    if (!resource.details?.hard || Object.keys(resource.details.hard).length === 0) {
        return <span className="text-gray-500 italic text-xs">No limits configured</span>;
    }

    const hard = resource.details.hard;
    const used = resource.details.used || {};

    // Prioritize CPU and Memory
    const priorityKeys = ['requests.cpu', 'limits.cpu', 'requests.memory', 'limits.memory'];
    const otherKeys = Object.keys(hard).filter(k => !priorityKeys.includes(k));

    return (
        <div className="space-y-2 max-w-md">
            {priorityKeys.map(key => {
                if (hard[key]) {
                    return (
                        <ProgressBar
                            key={key}
                            label={key}
                            used={used[key] || '0'}
                            hard={hard[key]}
                        />
                    );
                }
                return null;
            })}
            {otherKeys.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                    {otherKeys.map(key => (
                        <span key={key} className="bg-gray-800 text-gray-400 px-1.5 py-0.5 rounded text-[10px] border border-gray-700">
                            {key}: {used[key] || '0'}/{hard[key]}
                        </span>
                    ))}
                </div>
            )}
        </div>
    );
};

const QuotaList = ({ quotas, onEditYaml, onDelete }) => {
    const [menuOpen, setMenuOpen] = useState(null);

    const getAge = (created) => {
        if (!created) return 'Unknown';
        const diff = Date.now() - new Date(created).getTime();
        const days = Math.floor(diff / (1000 * 60 * 60 * 24));
        if (days > 0) return `${days}d`;
        const hours = Math.floor(diff / (1000 * 60 * 60));
        if (hours > 0) return `${hours}h`;
        const minutes = Math.floor(diff / (1000 * 60));
        return `${minutes}m`;
    };

    if (quotas.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center py-20 text-gray-500 bg-gray-900/30 rounded-lg border border-gray-800 border-dashed">
                <Activity size={48} className="mb-4 opacity-20" />
                <p className="text-lg">No resource quotas found</p>
                <p className="text-sm opacity-60">Create one to get started</p>
            </div>
        );
    }

    return (
        <div className="bg-gray-900 border border-gray-800 rounded-lg overflow-hidden">
            <table className="w-full text-left border-collapse">
                <thead>
                    <tr className="bg-gray-800/50 text-gray-400 text-xs uppercase tracking-wider border-b border-gray-700">
                        <th className="px-6 py-3 font-medium">Name</th>
                        <th className="px-6 py-3 font-medium">Namespace</th>
                        <th className="px-6 py-3 font-medium w-24 text-center">Age</th>
                        <th className="px-6 py-3 font-medium">Usage / Limits</th>
                        <th className="px-6 py-3 font-medium w-20 text-right">Actions</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-gray-800">
                    {quotas.map((resource) => (
                        <tr key={`${resource.namespace}-${resource.name}`} className="hover:bg-gray-800/30 transition-colors">
                            <td className="px-6 py-4 align-top">
                                <div className="flex items-center">
                                    <Activity size={16} className="text-blue-400 mr-2" />
                                    <span className="font-medium text-gray-200">{resource.name}</span>
                                </div>
                            </td>
                            <td className="px-6 py-4 align-top text-sm text-gray-300">
                                {resource.namespace}
                            </td>
                            <td className="px-6 py-4 align-top text-sm text-gray-400 text-center whitespace-nowrap">
                                {getAge(resource.created)}
                            </td>
                            <td className="px-6 py-4 align-top">
                                {renderQuotaUsage(resource)}
                            </td>
                            <td className="px-6 py-4 align-top text-right">
                                <div className="relative inline-block text-left">
                                    <button
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            const uniqueId = `quota-${resource.namespace}-${resource.name}`;
                                            setMenuOpen(menuOpen === uniqueId ? null : uniqueId);
                                        }}
                                        className="p-1.5 hover:bg-gray-700 rounded-lg text-gray-400 hover:text-white transition-colors"
                                    >
                                        <MoreVertical size={16} />
                                    </button>
                                    {menuOpen === `quota-${resource.namespace}-${resource.name}` && (
                                        <div className="absolute right-0 mt-2 w-48 bg-gray-800 border border-gray-700 rounded-lg shadow-xl z-50 py-1 animate-in fade-in zoom-in-95 duration-100">
                                            <button
                                                onClick={() => {
                                                    onEditYaml({ ...resource, kind: 'ResourceQuota' });
                                                    setMenuOpen(null);
                                                }}
                                                className="w-full text-left px-4 py-2.5 text-sm text-gray-300 hover:bg-gray-700 hover:text-white flex items-center"
                                            >
                                                <FileText size={14} className="mr-2" /> Edit YAML
                                            </button>
                                            <div className="h-px bg-gray-700 my-1"></div>
                                            <button
                                                onClick={() => {
                                                    onDelete({ resource, kind: 'ResourceQuota', force: false });
                                                    setMenuOpen(null);
                                                }}
                                                className="w-full text-left px-4 py-2.5 text-sm text-red-400 hover:bg-red-900/20 flex items-center"
                                            >
                                                <Trash2 size={14} className="mr-2" /> Delete
                                            </button>
                                        </div>
                                    )}
                                </div>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
            {menuOpen && (
                <div className="fixed inset-0 z-40" onClick={() => setMenuOpen(null)}></div>
            )}
        </div>
    );
};

export default QuotaList;
