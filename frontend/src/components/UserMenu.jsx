import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { User, Settings, LogOut, ChevronDown } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const UserMenu = () => {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const [menuOpen, setMenuOpen] = useState(false);
    const menuRef = useRef(null);

    // Determine user type and format username
    const getUserDisplayName = () => {
        if (!user) return 'Unknown';

        // Use idp field if available, otherwise infer from permissions
        let prefix = 'CORE';
        if (user.idp === 'ldap') {
            prefix = 'LDAP';
        } else if (user.permissions && Object.keys(user.permissions).length > 0) {
            // LDAP user with permissions (view/edit)
            prefix = 'LDAP';
        } else if (user.role === 'admin' && !user.idp) {
            // Core admin (default)
            prefix = 'CORE';
        }

        return `${prefix}\\${user.username}`;
    };

    // Close menu when clicking outside
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (menuRef.current && !menuRef.current.contains(event.target)) {
                setMenuOpen(false);
            }
        };

        if (menuOpen) {
            document.addEventListener('mousedown', handleClickOutside);
        }

        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [menuOpen]);

    const handleSettings = () => {
        setMenuOpen(false);
        navigate('/dashboard/settings');
    };

    const handleLogout = () => {
        setMenuOpen(false);
        logout();
    };

    if (!user) return null;

    return (
        <div className="relative" ref={menuRef}>
            <button
                onClick={() => setMenuOpen(!menuOpen)}
                className="flex items-center space-x-2 px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-gray-300 hover:text-white text-sm rounded-md transition-colors border border-gray-700 hover:border-gray-600"
            >
                <User size={16} />
                <span className="text-xs font-medium">{getUserDisplayName()}</span>
                <ChevronDown size={14} className={`transition-transform duration-200 ${menuOpen ? 'rotate-180' : ''}`} />
            </button>

            {menuOpen && (
                <div className="absolute right-0 mt-2 w-48 bg-gray-800 border border-gray-700 rounded-lg shadow-xl z-50 py-1 animate-in fade-in zoom-in-95 duration-100">
                    {/* Username (disabled) */}
                    <div className="px-4 py-2.5 text-sm text-gray-500 flex items-center cursor-default">
                        <User size={14} className="mr-2" />
                        {getUserDisplayName()}
                    </div>
                    <div className="h-px bg-gray-700 my-1"></div>

                    {/* Settings */}
                    <button
                        onClick={handleSettings}
                        className="w-full text-left px-4 py-2.5 text-sm text-gray-300 hover:bg-gray-700 hover:text-white flex items-center"
                    >
                        <Settings size={14} className="mr-2" /> Settings
                    </button>

                    {/* Logout */}
                    <button
                        onClick={handleLogout}
                        className="w-full text-left px-4 py-2.5 text-sm text-red-400 hover:bg-red-900/20 flex items-center"
                    >
                        <LogOut size={14} className="mr-2" /> Logout
                    </button>
                </div>
            )}
        </div>
    );
};

export default UserMenu;
