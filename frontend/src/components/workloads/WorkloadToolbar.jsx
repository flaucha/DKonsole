import React from 'react';
import { Search, X, RefreshCw, Columns } from 'lucide-react';

const WorkloadToolbar = ({
    kind,
    filter,
    setFilter,
    isSearchFocused,
    setIsSearchFocused,
    refetch,
    resourcesCount,
    menuOpen,
    setMenuOpen,
    orderedDataColumns,
    ageColumn,
    hidden,
    toggleVisibility,
    resetOrder
}) => {
    return (
        <div className="flex items-center justify-between p-4 border-b border-gray-800 bg-gray-900/50">
            <div className="flex items-center space-x-4 flex-1">
                <div className={`relative transition-all duration-300 ${isSearchFocused ? 'w-96' : 'w-64'}`}>
                    <Search className={`absolute left-3 top-1/2 transform -translate-y-1/2 transition-colors duration-300 ${isSearchFocused ? 'text-blue-400' : 'text-gray-500'}`} size={16} />
                    <input
                        type="text"
                        placeholder={`Filter ${kind}s...`}
                        value={filter}
                        onChange={(e) => setFilter(e.target.value)}
                        onFocus={() => setIsSearchFocused(true)}
                        onBlur={() => setIsSearchFocused(false)}
                        className="w-full bg-gray-900 border border-gray-700 text-gray-200 text-sm rounded-md pl-10 pr-10 py-2 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all duration-300"
                    />
                    {filter && (
                        <button
                            onClick={() => setFilter('')}
                            className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-200 transition-colors"
                            type="button"
                        >
                            <X size={16} />
                        </button>
                    )}
                </div>
                <span className="text-sm text-gray-500">
                    {resourcesCount} {resourcesCount === 1 ? 'item' : 'items'}
                </span>
            </div>
            <div className="flex items-center space-x-2">
                <div className="relative">
                    <button
                        onClick={() => setMenuOpen(menuOpen === 'columns' ? null : 'columns')}
                        className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-md transition-colors"
                        title="Manage Columns"
                    >
                        <Columns size={16} />
                    </button>
                    {menuOpen === 'columns' && (
                        <div className="absolute right-0 mt-2 w-56 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50 p-2">
                            <div className="text-xs font-semibold text-gray-500 mb-2 px-2 uppercase tracking-wider">
                                Visible Columns
                            </div>
                            <div className="space-y-1 max-h-64 overflow-y-auto">
                                {orderedDataColumns.concat([ageColumn]).map((col) => (
                                    <label key={col.id} className="flex items-center px-2 py-1.5 hover:bg-gray-700 rounded cursor-pointer">
                                        <input
                                            type="checkbox"
                                            checked={!hidden.includes(col.id)}
                                            onChange={() => toggleVisibility(col.id)}
                                            className="form-checkbox h-4 w-4 text-blue-500 rounded border-gray-600 bg-gray-700 focus:ring-blue-500 focus:ring-offset-gray-800"
                                        />
                                        <span className="ml-2 text-sm text-gray-200">{col.label}</span>
                                    </label>
                                ))}
                            </div>
                            <div className="border-t border-gray-700 mt-2 pt-2 px-2">
                                <button
                                    onClick={resetOrder}
                                    className="text-xs text-blue-400 hover:text-blue-300 w-full text-left"
                                >
                                    Reset to Default
                                </button>
                            </div>
                        </div>
                    )}
                </div>
                <button
                    onClick={() => refetch()}
                    className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-md transition-colors"
                    title="Refresh"
                >
                    <RefreshCw size={16} />
                </button>
            </div>
        </div>
    );
};

export default WorkloadToolbar;
