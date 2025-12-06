import React from 'react';
import { Search, X, Plus, RefreshCw } from 'lucide-react';

const NamespaceToolbar = ({
    filter,
    setFilter,
    isSearchFocused,
    setIsSearchFocused,
    filteredCount,
    isAdmin,
    onAdd,
    onRefresh
}) => {
    return (
        <div className="flex items-center justify-between p-4 border-b border-gray-800 bg-gray-900/50">
            <div className="flex items-center space-x-4 flex-1">
                <div className={`relative transition-all duration-300 ${isSearchFocused ? 'w-96' : 'w-64'}`}>
                    <Search className={`absolute left-3 top-1/2 transform -translate-y-1/2 transition-colors duration-300 ${isSearchFocused ? 'text-blue-400' : 'text-gray-500'}`} size={16} />
                    <input
                        type="text"
                        placeholder="Filter namespaces..."
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
                    {filteredCount} {filteredCount === 1 ? 'item' : 'items'}
                </span>
            </div>
            <div className="flex items-center space-x-2">
                {isAdmin && (
                    <button
                        onClick={onAdd}
                        className="flex items-center px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-sm transition-colors"
                        title="Create new namespace"
                    >
                        <Plus size={16} className="mr-1.5" />
                        Add
                    </button>
                )}
                <button
                    onClick={onRefresh}
                    className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-md transition-colors"
                    title="Refresh"
                >
                    <RefreshCw size={16} />
                </button>
            </div>
        </div>
    );
};

export default NamespaceToolbar;
