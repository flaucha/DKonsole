import React, { useState } from 'react';
import { Settings as SettingsIcon, Users, Key, AlertCircle } from 'lucide-react';
import { useLDAPSettings } from '../../hooks/useLDAPSettings';
import LDAPConfigTab from './ldap/LDAPConfigTab';
import LDAPGroupsTab from './ldap/LDAPGroupsTab';
import LDAPAdminsTab from './ldap/LDAPAdminsTab';

const LDAPSettings = () => {
    const {
        // State
        config,
        setConfig,
        adminGroups,
        setAdminGroups,
        credentials,
        setCredentials,
        savedCredentials,
        loading,
        testLoading,
        testResult,
        testMessage,
        namespaces,
        error,
        success,
        isAdmin,
        checkingAdmin,

        // Handlers
        handleSaveConfig,
        handleEnabledChange,
        handleSaveCredentials,
        handleTestConnection
    } = useLDAPSettings();

    const [ldapActiveTab, setLdapActiveTab] = useState('config');

    if (checkingAdmin) {
        return <div className="text-white">Checking permissions...</div>;
    }

    if (!isAdmin) {
        return (
            <div className="bg-red-900/20 border border-red-500/50 rounded-lg p-6 text-center w-full">
                <AlertCircle size={48} className="mx-auto mb-4 text-red-400" />
                <h2 className="text-xl font-semibold text-white mb-2">Access Denied</h2>
                <p className="text-gray-400">You need admin privileges to access LDAP settings.</p>
            </div>
        );
    }

    return (
        <div className="space-y-6 w-full">
            {/* LDAP Main Container */}
            <div className="bg-gray-800 rounded-lg border border-gray-700 shadow-lg w-full">
                {/* LDAP Tabs */}
                <div className="flex space-x-1 border-b border-gray-700 px-6 pt-4">
                    <button
                        className={`pb-3 px-4 flex items-center font-medium transition-colors ${ldapActiveTab === 'config' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                        onClick={() => setLdapActiveTab('config')}
                    >
                        <SettingsIcon size={18} className="mr-2" /> Config
                    </button>
                    <button
                        className={`pb-3 px-4 flex items-center font-medium transition-colors ${ldapActiveTab === 'permissions' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                        onClick={() => setLdapActiveTab('permissions')}
                    >
                        <Users size={18} className="mr-2" /> Group Permissions
                    </button>
                    <button
                        className={`pb-3 px-4 flex items-center font-medium transition-colors ${ldapActiveTab === 'admins' ? 'border-b-2 border-blue-500 text-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                        onClick={() => setLdapActiveTab('admins')}
                    >
                        <Key size={18} className="mr-2" /> Group Admins
                    </button>
                </div>

                {/* Tab Content */}
                <div className="p-6">
                    {ldapActiveTab === 'config' && (
                        <LDAPConfigTab
                            config={config}
                            setConfig={setConfig}
                            credentials={credentials}
                            setCredentials={setCredentials}
                            savedCredentials={savedCredentials}
                            loading={loading}
                            testLoading={testLoading}
                            testResult={testResult}
                            testMessage={testMessage}
                            handleSaveConfig={handleSaveConfig}
                            handleEnabledChange={handleEnabledChange}
                            handleSaveCredentials={handleSaveCredentials}
                            handleTestConnection={handleTestConnection}
                        />
                    )}

                    {ldapActiveTab === 'permissions' && (
                        <LDAPGroupsTab namespaces={namespaces} />
                    )}

                    {ldapActiveTab === 'admins' && (
                        <LDAPAdminsTab
                            config={config}
                            setConfig={setConfig}
                            adminGroups={adminGroups}
                            setAdminGroups={setAdminGroups}
                        />
                    )}
                </div>
            </div>

            {(error || success) && (
                <div className={`flex items-center text-sm rounded-lg p-3 ${error ? 'bg-red-900/20 border border-red-500/50 text-red-400' : 'bg-green-900/20 border border-green-500/50 text-green-400'}`}>
                    <AlertCircle size={16} className="mr-2" />
                    {error || success}
                </div>
            )}
        </div>
    );
};

export default LDAPSettings;

