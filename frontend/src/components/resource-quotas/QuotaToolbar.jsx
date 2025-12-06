import React, { useRef, useEffect, useState } from 'react';
import { Activity, Tag, RefreshCw, Globe, MapPin, Plus } from 'lucide-react';
import { DEFAULT_NAMESPACE } from '../../config/constants';

const QuotaToolbar = ({
    activeTab,
    setActiveTab,
    namespaceFilter,
    setNamespaceFilter,
    namespaceFromUrl,
    loading,
    quotasCount,
    limitRangesCount,
    onRefresh,
    onAddQuota,
    onAddLimitRange,
    namespaces = []
}) => {
    const [createMenuOpen, setCreateMenuOpen] = useState(false);
    const createMenuRef = useRef(null);

    // Close create menu when clicking outside
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (createMenuRef.current && !createMenuRef.current.contains(event.target)) {
                setCreateMenuOpen(false);
            }
        };
        if (createMenuOpen) {
            document.addEventListener('mousedown', handleClickOutside);
        }
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [createMenuOpen]);

    const handleAddClick = () => {
        const selectedNs = namespaceFilter !== 'all' ? namespaceFilter : (namespaceFromUrl && namespaceFromUrl !== 'all' ? namespaceFromUrl : (namespaces[0]?.name || DEFAULT_NAMESPACE));
        if (activeTab === 'quotas') {
            onAddQuota(selectedNs);
        } else {
            onAddLimitRange(selectedNs);
        }
    };

    return (
        <>
            {/* Header Section */}
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center space-x-3">
                    {activeTab === 'quotas' ? <Activity className="text-blue-400" size={24} /> : <Tag className="text-green-400" size={24} />}
                    <h1 className="text-2xl font-semibold text-white">
                        {activeTab === 'quotas' ? 'Resource Quotas' : 'Limit Ranges'}
                    </h1>
                    {loading && <RefreshCw size={18} className="animate-spin text-gray-400 ml-2" />}
                </div>
                <div className="flex items-center space-x-3">
                    <div className="bg-gray-800 border border-gray-700 rounded-md flex overflow-hidden text-sm shrink-0">
                        <button
                            onClick={() => setNamespaceFilter('all')}
                            className={`px-3 py-1.5 flex items-center space-x-1 ${namespaceFilter === 'all' ? 'bg-gray-700 text-white' : 'text-gray-300 hover:bg-gray-700'}`}
                            title="Show all namespaces"
                        >
                            <Globe size={14} /> <span className="hidden sm:inline">All</span>
                        </button>
                        <button
                            onClick={() => {
                                // Toggle between 'all' and the namespace from header/url
                                if (namespaceFilter === 'all' && namespaceFromUrl && namespaceFromUrl !== 'all') {
                                    setNamespaceFilter(namespaceFromUrl);
                                } else {
                                    setNamespaceFilter('all');
                                }
                            }}
                            className={`px-3 py-1.5 flex items-center space-x-1 ${namespaceFilter !== 'all' ? 'bg-gray-700 text-white' : 'text-gray-300 hover:bg-gray-700'}`}
                            title={namespaceFilter !== 'all' ? `Showing namespace: ${namespaceFilter}` : namespaceFromUrl && namespaceFromUrl !== 'all' ? `Click to show namespace: ${namespaceFromUrl}` : 'No namespace selected'}
                        >
                            <MapPin size={14} /> <span className="hidden sm:inline max-w-[120px] truncate">{namespaceFilter !== 'all' ? namespaceFilter : (namespaceFromUrl && namespaceFromUrl !== 'all' ? namespaceFromUrl : 'Namespace')}</span>
                        </button>
                    </div>
                    <button
                        onClick={onRefresh}
                        className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-white rounded-md border border-gray-700 text-sm transition-colors flex items-center"
                        title="Refresh"
                    >
                        <RefreshCw size={16} className={loading ? "animate-spin mr-2" : "mr-2"} />
                        Refresh
                    </button>

                    {/* Direct Add Button based on active tab */}
                    <button
                        onClick={handleAddClick}
                        className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md border border-blue-500 text-sm transition-colors flex items-center shadow-lg shadow-blue-900/20"
                    >
                        <Plus size={16} className="mr-2" />
                        Add {activeTab === 'quotas' ? 'Quota' : 'Limit Range'}...
                    </button>
                </div>
            </div>

            {/* Tabs */}
            <div className="flex space-x-1 bg-gray-800/50 p-1 rounded-lg w-fit mb-6 border border-gray-700/50">
                <button
                    onClick={() => setActiveTab('quotas')}
                    className={`px-6 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 flex items-center ${activeTab === 'quotas'
                        ? 'bg-gray-700 text-white shadow-sm'
                        : 'text-gray-400 hover:text-white hover:bg-gray-700/50'
                        }`}
                >
                    <Activity size={16} className="mr-2" />
                    Resource Quotas
                    <span className="ml-2 bg-black/20 px-2 py-0.5 rounded-full text-xs">{quotasCount}</span>
                </button>
                <button
                    onClick={() => setActiveTab('limits')}
                    className={`px-6 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 flex items-center ${activeTab === 'limits'
                        ? 'bg-gray-700 text-white shadow-sm'
                        : 'text-gray-400 hover:text-white hover:bg-gray-700/50'
                        }`}
                >
                    <Tag size={16} className="mr-2" />
                    Limit Ranges
                    <span className="ml-2 bg-black/20 px-2 py-0.5 rounded-full text-xs">{limitRangesCount}</span>
                </button>
            </div>
        </>
    );
};

export default QuotaToolbar;
