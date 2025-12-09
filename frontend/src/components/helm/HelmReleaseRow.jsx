import React, { useState } from 'react';
import { Package, Clock, Tag, MoreVertical, Trash2, ArrowUp, ChevronDown } from 'lucide-react';
import { formatDateTime } from '../../utils/dateUtils';
import { getExpandableRowClasses, getExpandableRowRowClasses } from '../../utils/expandableRow';

const HelmReleaseRow = ({ release, isExpanded, onToggle, onUpgrade, onDelete, hasEditPermission }) => {
    const [menuOpen, setMenuOpen] = useState(false);
    const releaseKey = `${release.namespace}/${release.name}`;

    const getStatusBadge = (status) => {
        const statusLower = (status || 'unknown').toLowerCase();
        if (statusLower === 'deployed') {
            return <span className="px-2 py-1 text-xs rounded-full bg-green-900/50 text-green-300 border border-green-700">Deployed</span>;
        } else if (statusLower === 'failed' || statusLower === 'error') {
            return <span className="px-2 py-1 text-xs rounded-full bg-red-900/50 text-red-300 border border-red-700">Failed</span>;
        } else if (statusLower === 'pending-install' || statusLower === 'pending-upgrade' || statusLower === 'pending-rollback') {
            return <span className="px-2 py-1 text-xs rounded-full bg-yellow-900/50 text-yellow-300 border border-yellow-700">Pending</span>;
        } else {
            return <span className="px-2 py-1 text-xs rounded-full bg-gray-700 text-gray-300 border border-gray-600">{status || 'Unknown'}</span>;
        }
    };

    return (
        <div className="border-b border-gray-800 last:border-0">
            <div
                onClick={onToggle}
                className={`grid grid-cols-12 gap-4 px-6 py-4 cursor-pointer transition-colors duration-200 items-center ${getExpandableRowRowClasses(isExpanded)}`}
            >
                <div className="col-span-3 flex items-center font-medium text-sm text-gray-200">
                    <ChevronDown
                        size={16}
                        className={`mr-2 text-gray-500 transition-transform duration-200 ${isExpanded ? 'transform rotate-180' : ''}`}
                    />
                    <Package size={16} className="mr-3 text-gray-500" />
                    <div className="min-w-0 flex-1">
                        <span className="truncate block" title={release.name}>{release.name}</span>
                        {release.description && (
                            <div className="text-xs text-gray-500 truncate">{release.description}</div>
                        )}
                    </div>
                </div>
                <div className="col-span-2 text-sm text-gray-300">
                    {release.chart || '-'}
                    {release.appVersion && (
                        <div className="text-xs text-gray-500">App: {release.appVersion}</div>
                    )}
                </div>
                <div className="col-span-1 text-sm text-gray-300">{release.version || '-'}</div>
                <div className="col-span-1 text-sm text-gray-300">{release.namespace}</div>
                <div className="col-span-2">
                    {getStatusBadge(release.status)}
                </div>
                <div className="col-span-2 text-sm text-gray-400">
                    {formatDateTime(release.updated)}
                </div>
                <div className="col-span-1 flex justify-end" onClick={(e) => e.stopPropagation()}>
                    <div className="relative">
                        <button
                            onClick={() => setMenuOpen(!menuOpen)}
                            className="p-1 hover:bg-gray-800 rounded text-gray-400 hover:text-white transition-colors"
                        >
                            <MoreVertical size={16} />
                        </button>
                        {menuOpen && (
                            <div className="absolute right-0 mt-1 w-40 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
                                <div className="flex flex-col">
                                    {hasEditPermission ? (
                                        <>
                                            <button
                                                onClick={() => {
                                                    onUpgrade(release);
                                                    setMenuOpen(false);
                                                }}
                                                className="w-full text-left px-4 py-2 text-sm text-blue-300 hover:bg-blue-900/40 flex items-center"
                                            >
                                                <ArrowUp size={14} className="mr-2" />
                                                Upgrade
                                            </button>
                                            <button
                                                onClick={() => {
                                                    onDelete(release);
                                                    setMenuOpen(false);
                                                }}
                                                className="w-full text-left px-4 py-2 text-sm text-red-300 hover:bg-red-900/40 flex items-center"
                                            >
                                                <Trash2 size={14} className="mr-2" />
                                                Uninstall
                                            </button>
                                        </>
                                    ) : (
                                        <div className="px-4 py-2 text-xs text-gray-500">
                                            View only
                                        </div>
                                    )}
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            {/* Expanded Details */}
            <div className={`${getExpandableRowClasses(isExpanded, false)}`}>
                {isExpanded && (
                    <div className="px-6 py-4 bg-gray-900/30 border-t border-gray-800">
                        <div className="bg-gray-900/50 rounded-lg border border-gray-800 overflow-hidden">
                            <div className="p-4 space-y-6">
                                {/* Basic Information */}
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <div>
                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                            <Package size={12} className="mr-1" />
                                            Chart
                                        </h4>
                                        <div className="text-sm text-gray-300">{release.chart || '-'}</div>
                                        {release.appVersion && (
                                            <div className="text-xs text-gray-500 mt-1">App Version: {release.appVersion}</div>
                                        )}
                                    </div>
                                    <div>
                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                            <Tag size={12} className="mr-1" />
                                            Version
                                        </h4>
                                        <div className="text-sm text-gray-300">{release.version || '-'}</div>
                                    </div>
                                    <div>
                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                            <Clock size={12} className="mr-1" />
                                            Last Updated
                                        </h4>
                                        <div className="text-sm text-gray-300">{formatDateTime(release.updated)}</div>
                                    </div>
                                    <div>
                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center">
                                            <Tag size={12} className="mr-1" />
                                            Revision
                                        </h4>
                                        <div className="text-sm text-gray-300">{release.revision || '-'}</div>
                                    </div>
                                </div>

                                {/* Description */}
                                {release.description && (
                                    <div>
                                        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Description</h4>
                                        <div className="text-sm text-gray-300">{release.description}</div>
                                    </div>
                                )}

                                {/* Status */}
                                <div>
                                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Status</h4>
                                    {getStatusBadge(release.status)}
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(false)}
                ></div>
            )}
        </div>
    );
};

export default HelmReleaseRow;
