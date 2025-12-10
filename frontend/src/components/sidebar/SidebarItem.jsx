import React from 'react';
import { NavLink } from 'react-router-dom';
import { ChevronDown } from 'lucide-react';

const SidebarItem = ({ icon: Icon, label, to, onClick, hasChildren, expanded }) => {
    if (hasChildren) {
        return (
            <div
                onClick={onClick}
                className={`flex items-center justify-between px-4 py-2 cursor-pointer rounded-md transition-all duration-200 text-gray-100 hover:bg-gray-800 hover:text-gray-100 border border-transparent hover:border-gray-700 ${expanded ? 'bg-gray-800/50 border-gray-700' : ''}`}
            >
                <div className="flex items-center space-x-3">
                    <Icon size={20} className="text-gray-300 group-hover:text-blue-400" />
                    <span className="font-medium whitespace-nowrap">{label}</span>
                </div>
                <div className={`transition-transform duration-200 text-gray-400 ${expanded ? 'rotate-0 text-blue-400' : '-rotate-90'}`}>
                    <ChevronDown size={16} />
                </div>
            </div>
        );
    }

    return (
        <NavLink
            to={to}
            className={({ isActive }) =>
                `flex items-center justify-between px-4 py-2 cursor-pointer rounded-md transition-all duration-200 border border-transparent ${isActive
                    ? 'bg-gray-800 text-gray-100 border-l-4 border-l-blue-500 shadow-md'
                    : 'text-gray-300 hover:bg-gray-800 hover:text-gray-100 hover:border-gray-700'
                }`
            }
        >
            {({ isActive }) => (
                <div className="flex items-center space-x-3">
                    <Icon size={20} className={isActive ? 'text-blue-400' : 'text-gray-400'} />
                    <span className="font-medium whitespace-nowrap">{label}</span>
                </div>
            )}
        </NavLink>
    );
};

export default SidebarItem;
