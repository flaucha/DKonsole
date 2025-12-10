import React from 'react';
import { NavLink } from 'react-router-dom';

const SubItem = ({ label, to }) => (
    <NavLink
        to={to}
        className={({ isActive }) =>
            `block pl-12 pr-4 py-1.5 cursor-pointer text-xs transition-all duration-200 whitespace-nowrap rounded-md border border-transparent ${isActive
                ? 'text-gray-100 font-semibold bg-gray-800/60 border-l-4 border-l-blue-500 shadow-sm'
                : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800/40 hover:border-gray-700'
            }`
        }
        style={{ fontSize: '0.75rem' }}
    >
        {({ isActive }) => (
            <>
                <span className={`mr-2 ${isActive ? 'text-blue-400' : 'text-gray-500'}`}>•</span>
                {label}
            </>
        )}
    </NavLink>
);

export default SubItem;
