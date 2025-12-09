import React, { useState } from 'react';
import HelmReleaseRow from './HelmReleaseRow';

const HelmReleaseList = ({
    releases,
    sortField,
    sortDirection,
    onSort,
    onUpgrade,
    onDelete,
    hasEditPermission,
    filter
}) => {
    const [expandedId, setExpandedId] = useState(null);

    const renderSortIndicator = (field) => {
        if (sortField !== field) return null;
        return sortDirection === 'asc' ? '↑' : '↓';
    };

    const toggleExpand = (releaseKey) => {
        setExpandedId(current => current === releaseKey ? null : releaseKey);
    };

    return (
        <>
            {/* Table Header */}
            <div className="grid grid-cols-12 gap-4 px-6 py-3 border-b border-gray-800 bg-gray-900/50 text-xs font-medium text-gray-500 uppercase tracking-wider">
                <div className="col-span-3 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => onSort('name')}>
                    Release {renderSortIndicator('name')}
                </div>
                <div className="col-span-2 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => onSort('chart')}>
                    Chart {renderSortIndicator('chart')}
                </div>
                <div className="col-span-1 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => onSort('version')}>
                    Version {renderSortIndicator('version')}
                </div>
                <div className="col-span-1 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => onSort('namespace')}>
                    Namespace {renderSortIndicator('namespace')}
                </div>
                <div className="col-span-2 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => onSort('status')}>
                    Status {renderSortIndicator('status')}
                </div>
                <div className="col-span-2 cursor-pointer hover:text-gray-300 flex items-center" onClick={() => onSort('updated')}>
                    Updated {renderSortIndicator('updated')}
                </div>
                <div className="col-span-1"></div>
            </div>

            {/* Table Body */}
            <div className="flex-1 overflow-y-auto">
                {releases.map((release) => {
                    const releaseKey = `${release.namespace}/${release.name}`;
                    return (
                        <HelmReleaseRow
                            key={releaseKey}
                            release={release}
                            isExpanded={expandedId === releaseKey}
                            onToggle={() => toggleExpand(releaseKey)}
                            onUpgrade={onUpgrade}
                            onDelete={onDelete}
                            hasEditPermission={hasEditPermission(release.namespace)}
                        />
                    );
                })}
                {releases.length === 0 && (
                    <div className="p-8 text-center text-gray-500">
                        {filter ? 'No releases match your filter.' : 'No Helm releases found.'}
                    </div>
                )}
            </div>
        </>
    );
};

export default HelmReleaseList;
