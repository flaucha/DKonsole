import React from 'react';
import { PanelLeftClose, PanelLeftOpen } from 'lucide-react';
import TerminalDock from './TerminalDock';
import UserMenu from './UserMenu';

const Header = ({
    sidebarOpen,
    setSidebarOpen,
    logoSrc,
    handleLogoError,
    checkingAdmin,
    hasPermissions,
    headerContent,
    user
}) => {
    return (
        <header className="bg-gray-900 border-b border-gray-700 flex flex-col shrink-0 z-20 shadow-lg">
            <div className="h-12 flex items-center gap-4 px-4">
                <div className="flex items-center space-x-2 shrink-0">
                    <button
                        onClick={() => setSidebarOpen(!sidebarOpen)}
                        className="p-2 text-gray-400 hover:text-gray-100 hover:bg-gray-800 rounded-md transition-all duration-200 hover:scale-110 border border-transparent hover:border-gray-600"
                        title={sidebarOpen ? "Ocultar menú" : "Mostrar menú"}
                    >
                        {sidebarOpen ? (
                            <PanelLeftClose size={24} className="transition-transform duration-300" />
                        ) : (
                            <PanelLeftOpen size={24} className="transition-transform duration-300" />
                        )}
                    </button>
                    <div className="flex items-center justify-center">
                        <img
                            src={logoSrc}
                            alt="Logo"
                            className="h-9 max-h-9 object-contain"
                            onError={handleLogoError}
                        />
                    </div>
                </div>
                <div className="flex-1 min-w-0">
                    <TerminalDock />
                </div>
                <div className="flex items-center space-x-2 shrink-0">
                    {!checkingAdmin && hasPermissions && headerContent}
                    {!checkingAdmin && user && <UserMenu />}
                </div>
            </div>
        </header>
    );
};

export default Header;
