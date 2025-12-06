import React, { useState } from 'react';
import { Tag, MoreVertical, FileText, Trash2 } from 'lucide-react';

const renderLimitDetails = (resource) => {
    if (!resource.details?.limits || resource.details.limits.length === 0) {
        return <span className="text-gray-500 italic text-xs">No limits configured</span>;
    }

    return (
        <div className="space-y-2">
            {resource.details.limits.map((limit, idx) => (
                <div key={idx} className="bg-gray-800/50 rounded p-2 border border-gray-700/50 text-xs">
                    <div className="font-bold text-gray-400 mb-1 uppercase tracking-wider">{limit.type || 'Container'}</div>
                    <div className="grid grid-cols-2 gap-x-4 gap-y-1">
                        {limit.max && Object.entries(limit.max).map(([k, v]) => (
                            <div key={`max-${k}`} className="text-gray-400">
                                <span className="text-gray-500">Max {k}:</span> <span className="text-gray-200">{v}</span>
                            </div>
                        ))}
                        {limit.min && Object.entries(limit.min).map(([k, v]) => (
                            <div key={`min-${k}`} className="text-gray-400">
                                <span className="text-gray-500">Min {k}:</span> <span className="text-gray-200">{v}</span>
                            </div>
                        ))}
                        {limit.defaultRequest && Object.entries(limit.defaultRequest).map(([k, v]) => (
                            <div key={`dr-${k}`} className="text-gray-400 col-span-2">
                                <span className="text-gray-500">Default Req {k}:</span> <span className="text-gray-200">{v}</span>
                            </div>
                        ))}
                    </div>
                </div>
            ))}
        </div>
    );
};

const LimitRangeList = ({ limitRanges, onEditYaml, onDelete }) => {
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

    if (limitRanges.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center py-20 text-gray-500 bg-gray-900/30 rounded-lg border border-gray-800 border-dashed">
                <Tag size={48} className="mb-4 opacity-20" />
                <p className="text-lg">No limit ranges found</p>
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
                        <th className="px-6 py-3 font-medium">Details</th>
                        <th className="px-6 py-3 font-medium w-20 text-right">Actions</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-gray-800">
                    {limitRanges.map((resource) => (
                        <tr key={`${resource.namespace}-${resource.name}`} className="hover:bg-gray-800/30 transition-colors">
                            <td className="px-6 py-4 align-top">
                                <div className="flex items-center">
                                    <Tag size={16} className="text-green-400 mr-2" />
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
                                {renderLimitDetails(resource)}
                            </td>
                            <td className="px-6 py-4 align-top text-right">
                                <div className="relative inline-block text-left">
                                    <button
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            const uniqueId = `limit-${resource.namespace}-${resource.name}`;
                                            setMenuOpen(menuOpen === uniqueId ? null : uniqueId);
                                        }}
                                        className="p-1.5 hover:bg-gray-700 rounded-lg text-gray-400 hover:text-white transition-colors"
                                    >
                                        <MoreVertical size={16} />
                                    </button>
                                    {menuOpen === `limit-${resource.namespace}-${resource.name}` && (
                                        <div className="absolute right-0 mt-2 w-48 bg-gray-800 border border-gray-700 rounded-lg shadow-xl z-50 py-1 animate-in fade-in zoom-in-95 duration-100">
                                            <button
                                                onClick={() => {
                                                    onEditYaml({ ...resource, kind: 'LimitRange' });
                                                    setMenuOpen(null);
                                                }}
                                                className="w-full text-left px-4 py-2.5 text-sm text-gray-300 hover:bg-gray-700 hover:text-white flex items-center"
                                            >
                                                <FileText size={14} className="mr-2" /> Edit YAML
                                            </button>
                                            <div className="h-px bg-gray-700 my-1"></div>
                                            <button
                                                onClick={() => {
                                                    onDelete({ resource, kind: 'LimitRange', force: false });
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

export default LimitRangeList;
