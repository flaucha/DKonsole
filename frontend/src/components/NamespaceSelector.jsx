import React, { useState, useEffect, useRef } from 'react';
import { ChevronDown, Search, Database } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { logger } from '../utils/logger';

const NamespaceSelector = ({ selected, onSelect }) => {
    const [namespaces, setNamespaces] = useState([]);
    const [isOpen, setIsOpen] = useState(false);
    const [searchTerm, setSearchTerm] = useState('');
    const dropdownRef = useRef(null);
    const searchInputRef = useRef(null);
    const { authFetch } = useAuth();

    useEffect(() => {
        authFetch('/api/namespaces')
            .then(res => {
                if (!res.ok) {
                    throw new Error(`Failed to fetch namespaces: ${res.status} ${res.statusText}`);
                }
                return res.json();
            })
            .then(data => {
                if (Array.isArray(data)) {
                    setNamespaces(data);
                } else {
                    logger.error("Invalid namespaces response:", data);
                    setNamespaces([]);
                }
            })
            .catch(err => {
                logger.error("Failed to fetch namespaces:", err);
                setNamespaces([]);
            });
    }, [authFetch]);

    useEffect(() => {
        const handleClickOutside = (event) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    useEffect(() => {
        if (isOpen && searchInputRef.current) {
            searchInputRef.current.focus();
        }
        if (!isOpen) {
            setSearchTerm('');
        }
    }, [isOpen]);

    const filteredNamespaces = namespaces.filter(ns =>
        ns.name.toLowerCase().includes(searchTerm.toLowerCase())
    );

    return (
        <div className="flex items-center space-x-3" ref={dropdownRef}>
            <span className="text-gray-400 text-sm font-medium hidden md:block">Namespace:</span>
            <div className="relative">
                <button
                    onClick={() => setIsOpen(!isOpen)}
                    className="flex items-center justify-between space-x-2 bg-gray-800 border border-gray-700 text-white px-3 py-1.5 rounded-md hover:bg-gray-700 transition-colors min-w-[200px] text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50"
                >
                    <div className="flex items-center space-x-2 overflow-hidden">
                        <Database size={14} className="text-gray-400 flex-shrink-0" />
                        <span className="truncate">{selected}</span>
                    </div>
                    <ChevronDown size={14} className={`text-gray-400 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`} />
                </button>

                {isOpen && (
                    <div className="absolute top-full right-0 mt-1 w-64 bg-gray-800 border border-gray-700 rounded-md shadow-xl z-50 flex flex-col">
                        <div className="p-2 border-b border-gray-700">
                            <div className="relative">
                                <Search size={14} className="absolute left-2 top-1/2 transform -translate-y-1/2 text-gray-500" />
                                <input
                                    ref={searchInputRef}
                                    type="text"
                                    placeholder="Search namespaces..."
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="w-full bg-gray-900 border border-gray-700 text-gray-300 text-xs rounded pl-8 pr-8 py-1.5 focus:outline-none focus:border-blue-500"
                                />
                                {searchTerm && (
                                    <button
                                        onClick={() => setSearchTerm('')}
                                        className="absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-200 transition-colors"
                                        type="button"
                                    >
                                        <X size={12} />
                                    </button>
                                )}
                            </div>
                        </div>
                        <div className="max-h-60 overflow-y-auto py-1">
                            <button
                                onClick={() => {
                                    onSelect('all');
                                    setIsOpen(false);
                                }}
                                className={`w-full text-left px-4 py-2 text-sm hover:bg-gray-700 transition-colors flex items-center space-x-2 ${selected === 'all' ? 'bg-blue-900/30 text-blue-400 border-l-2 border-blue-500' : 'text-gray-300 border-l-2 border-transparent'}`}
                            >
                                <span>All</span>
                            </button>
                            {filteredNamespaces.length > 0 ? (
                                filteredNamespaces.map(ns => (
                                    <button
                                        key={ns.name}
                                        onClick={() => {
                                            onSelect(ns.name);
                                            setIsOpen(false);
                                        }}
                                        className={`w-full text-left px-4 py-2 text-sm hover:bg-gray-700 transition-colors flex items-center space-x-2 ${selected === ns.name ? 'bg-blue-900/30 text-blue-400 border-l-2 border-blue-500' : 'text-gray-300 border-l-2 border-transparent'}`}
                                    >
                                        <span>{ns.name}</span>
                                        {ns.status === 'Active' && <span className="w-1.5 h-1.5 rounded-full bg-green-500 ml-auto"></span>}
                                    </button>
                                ))
                            ) : (
                                <div className="px-4 py-3 text-xs text-gray-500 text-center italic">
                                    No namespaces found
                                </div>
                            )}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default NamespaceSelector;
