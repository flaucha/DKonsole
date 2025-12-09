import React, { useState, useEffect } from 'react';
import logoFullDark from '../assets/logo-full-dark.png';
import logoFullLight from '../assets/logo-full-light.png';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import Header from './Header';
import Sidebar from './Sidebar';

const Layout = ({ children, headerContent }) => {
    const { theme, menuAnimation, menuAnimationSpeed } = useSettings();
    const { authFetch, user } = useAuth();
    const [isAdmin, setIsAdmin] = useState(false);
    const [checkingAdmin, setCheckingAdmin] = useState(true);
    const [hasPermissions, setHasPermissions] = useState(false);

    useEffect(() => {
        // Check if user is admin (core admin or LDAP admin group member)
        const checkAdmin = async () => {
            try {
                const res = await authFetch('/api/settings/prometheus/url');
                if (res.ok || res.status === 404) {
                    setIsAdmin(true);
                    setHasPermissions(true);
                } else if (res.status === 403) {
                    setIsAdmin(false);
                } else {
                    setIsAdmin(false);
                }
            } catch {
                setIsAdmin(false);
            } finally {
                setCheckingAdmin(false);
            }
        };

        // Check if user has permissions (admin or LDAP permissions)
        const checkPermissions = () => {
            if (!user) {
                setHasPermissions(false);
                return;
            }
            // Admin always has permissions
            if (user.role === 'admin') {
                setHasPermissions(true);
                return;
            }
            // Check if user has LDAP permissions
            if (user.permissions && Object.keys(user.permissions).length > 0) {
                setHasPermissions(true);
            } else {
                setHasPermissions(false);
            }
        };

        if (user) {
            checkAdmin();
            checkPermissions();
        } else {
            setCheckingAdmin(false);
            setHasPermissions(false);
        }
    }, [authFetch, user]);
    const [sidebarOpen, setSidebarOpen] = useState(true);
    const [expandedMenus, setExpandedMenus] = useState({
        workloads: false,
        networking: false,
        storage: false,
        accessControl: false,
        adminArea: false,
    });
    // Get current theme and determine default logo immediately
    const currentTheme = theme || localStorage.getItem('theme') || 'default';
    const isLightTheme = currentTheme === 'light' || currentTheme === 'cream';
    const defaultLogoSrc = isLightTheme ? logoFullLight : logoFullDark;

    const [logoSrc, setLogoSrc] = useState(defaultLogoSrc); // Show default logo immediately

    useEffect(() => {
        // Use fetch (not authFetch) since /api/logo GET is public
        // Check for custom logo in background, but show default immediately
        const logoType = isLightTheme ? 'light' : 'normal';
        const timestamp = Date.now();

        fetch(`/api/logo?type=${logoType}&t=${timestamp}`)
            .then(res => {
                if (res.ok && res.status === 200) {
                    // Custom logo exists, use it
                    setLogoSrc(`/api/logo?type=${logoType}&t=${timestamp}`);
                }
                // If not found (404), keep default logo (already set)
            })
            .catch(() => {
                // On error, keep default logo (already set)
            });
    }, [theme, isLightTheme]);

    const handleLogoError = () => {
        // If logo fails to load, fallback to theme-appropriate default logo
        const currentTheme = theme || localStorage.getItem('theme') || 'default';
        const isLightTheme = currentTheme === 'light' || currentTheme === 'cream';
        const fallbackLogo = isLightTheme ? logoFullLight : logoFullDark;

        if (logoSrc !== fallbackLogo) {
            setLogoSrc(fallbackLogo);
        }
    };

    const toggleMenu = (menu) => {
        setExpandedMenus(prev => {
            // If clicking the same menu, toggle it
            if (prev[menu]) {
                return { ...prev, [menu]: false };
            }
            // Otherwise, close all and open only the clicked one
            return {
                workloads: false,
                networking: false,
                storage: false,
                accessControl: false,
                adminArea: false,
                [menu]: true
            };
        });
    };

    return (
        <div className="flex flex-col h-screen bg-gray-900">
            <Header
                sidebarOpen={sidebarOpen}
                setSidebarOpen={setSidebarOpen}
                logoSrc={logoSrc}
                handleLogoError={handleLogoError}
                checkingAdmin={checkingAdmin}
                hasPermissions={hasPermissions}
                headerContent={headerContent}
                user={user}
            />

            <div className="flex flex-1 overflow-hidden relative">
                <Sidebar
                    sidebarOpen={sidebarOpen}
                    setSidebarOpen={setSidebarOpen}
                    expandedMenus={expandedMenus}
                    toggleMenu={toggleMenu}
                    isAdmin={isAdmin}
                    hasPermissions={hasPermissions}
                    menuAnimation={menuAnimation}
                    menuAnimationSpeed={menuAnimationSpeed}
                    checkingAdmin={checkingAdmin}
                    user={user}
                />

                {/* Main Content */}
                <div className="flex-1 overflow-auto bg-gray-900 relative">
                    {children}
                </div>
            </div>
        </div>
    );
};

export default Layout;
