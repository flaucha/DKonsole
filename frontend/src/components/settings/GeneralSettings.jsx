import React, { useState, useEffect } from 'react';
import { AlertCircle, Server, Settings as SettingsIcon } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { isCoreAdmin } from '../../utils/permissions';
import PrometheusSettings from './PrometheusSettings';
import PasswordSettings from './PasswordSettings';

const GeneralSettings = () => {
    const { authFetch, user } = useAuth();
    const [isAdmin, setIsAdmin] = useState(false);
    const [checkingAdmin, setCheckingAdmin] = useState(true);

    useEffect(() => {
        // Check if user is admin (core admin or LDAP admin group member)
        // We check by trying to access a settings endpoint
        const checkAdmin = async () => {
            try {
                const res = await authFetch('/api/settings/prometheus/url');
                if (res.ok || res.status === 404) {
                    // If we can access it (or it doesn't exist), user is admin
                    setIsAdmin(true);
                } else if (res.status === 403) {
                    // Forbidden - user is not admin
                    setIsAdmin(false);
                } else {
                    // Other error - assume not admin for security
                    setIsAdmin(false);
                }
            } catch {
                // Error accessing - assume not admin for security
                setIsAdmin(false);
            } finally {
                setCheckingAdmin(false);
            }
        };
        if (user) {
            checkAdmin();
        } else {
            setCheckingAdmin(false);
        }
    }, [authFetch, user]);

    return (
        <div className="w-full">
            {checkingAdmin ? (
                <div className="text-white">Checking permissions...</div>
            ) : !isAdmin ? (
                <div className="bg-red-900/20 border border-red-500/50 rounded-lg p-6 text-center">
                    <AlertCircle size={48} className="mx-auto mb-4 text-red-400" />
                    <h2 className="text-xl font-semibold text-white mb-2">Access Denied</h2>
                    <p className="text-gray-400">You need admin privileges to access general settings.</p>
                </div>
            ) : (
                <div className="w-full space-y-6">
                    {/* Prometheus URL Settings */}
                    <PrometheusSettings />

                    {/* Password Change Settings - Only for CORE admin, not LDAP admin or regular users */}
                    {user && isCoreAdmin(user) && (
                        <PasswordSettings />
                    )}
                </div>
            )}
        </div>
    );
};

export default GeneralSettings;
