import React, { useState, useEffect } from 'react';
import { Settings as SettingsIcon, Users, Key, AlertCircle } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { parseErrorResponse, parseError } from '../../utils/errorParser';
import LDAPConfigTab from './ldap/LDAPConfigTab';
import LDAPGroupsTab from './ldap/LDAPGroupsTab';
import LDAPAdminsTab from './ldap/LDAPAdminsTab';

const LDAPSettings = () => {
    const { authFetch } = useAuth();
    // Config State
    const [config, setConfig] = useState({
        insecureSkipVerify: false,
        caCert: '',
        enabled: false,
        url: '',
        baseDN: '',
        userDN: 'uid',
        groupDN: '',
        userFilter: '',
        requiredGroup: ''
    });
    // Admin Groups State
    const [adminGroups, setAdminGroups] = useState([]);

    // Credentials State
    const [credentials, setCredentials] = useState({
        username: '',
        password: ''
    });
    const [savedCredentials, setSavedCredentials] = useState({
        username: '',
        password: '*****'
    });

    // UI State
    const [loading, setLoading] = useState(false);
    const [testLoading, setTestLoading] = useState(false);
    const [testResult, setTestResult] = useState(null); // null, 'success', 'error'
    const [testMessage, setTestMessage] = useState('');
    const [namespaces, setNamespaces] = useState([]);
    const [ldapActiveTab, setLdapActiveTab] = useState('config');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [isAdmin, setIsAdmin] = useState(false);
    const [checkingAdmin, setCheckingAdmin] = useState(true);

    // Initial Admin Check
    useEffect(() => {
        const checkAdmin = async () => {
            try {
                // Try to access a protected endpoint
                const res = await authFetch('/api/ldap/config');
                if (res.ok || res.status === 404) { // 404 means route exists but config might not, implying access is allowed
                    setIsAdmin(true);
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
        checkAdmin();
    }, [authFetch]);

    useEffect(() => {
        if (isAdmin) {
            loadConfig();
            loadNamespaces();
            loadCredentials();
        }
    }, [authFetch, isAdmin]);

    const loadConfig = async () => {
        try {
            const res = await authFetch('/api/ldap/config');
            if (res.ok) {
                const data = await res.json();
                setConfig(data);
                // Load admin groups from config
                if (data.adminGroups && Array.isArray(data.adminGroups)) {
                    setAdminGroups(data.adminGroups);
                } else {
                    setAdminGroups([]);
                }
            }
        } catch {
            // Ignore errors - LDAP might not be configured
        }
    };

    const loadNamespaces = async () => {
        try {
            const res = await authFetch('/api/namespaces');
            if (res.ok) {
                const data = await res.json();
                setNamespaces(data.map(ns => ns.name || ns));
            }
        } catch {
            // Ignore errors
        }
    };

    const loadCredentials = async () => {
        try {
            const res = await authFetch('/api/ldap/credentials');
            if (res.ok) {
                const data = await res.json();
                if (data.configured && data.username) {
                    setSavedCredentials({
                        username: data.username,
                        password: '*****'
                    });
                    // Pre-fill the form with saved username and masked password
                    setCredentials({
                        username: data.username,
                        password: '*****'
                    });
                }
            }
        } catch {
            // Ignore errors
        }
    };

    const handleSaveConfig = async () => {
        setError('');
        setSuccess('');
        setLoading(true);

        try {
            const res = await authFetch('/api/ldap/config', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ config }),
            });

            if (res.ok) {
                setSuccess('LDAP configuration saved successfully!');
            } else {
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to save LDAP configuration');
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to save LDAP configuration');
        } finally {
            setLoading(false);
        }
    };

    const handleEnabledChange = async (enabled) => {
        const newConfig = { ...config, enabled };
        setConfig(newConfig);

        // Auto-save when enabled checkbox changes
        setError('');
        setSuccess('');
        setLoading(true);

        try {
            const res = await authFetch('/api/ldap/config', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ config: newConfig }),
            });

            if (res.ok) {
                setSuccess('LDAP configuration saved successfully!');
            } else {
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to save LDAP configuration');
                // Revert on error
                setConfig(config);
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to save LDAP configuration');
            // Revert on error
            setConfig(config);
        } finally {
            setLoading(false);
        }
    };

    const handleSaveCredentials = async () => {
        setError('');
        setSuccess('');
        setLoading(true);

        // If password is masked, only send username (password stays the same)
        const credsToSend = {
            username: credentials.username,
            password: credentials.password === '*****' ? '' : credentials.password
        };

        // If username hasn't changed and password is masked, nothing to update
        if (credsToSend.password === '' && credentials.username === savedCredentials.username) {
            setSuccess('No changes to save');
            setLoading(false);
            return;
        }

        try {
            const res = await authFetch('/api/ldap/credentials', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(credsToSend),
            });

            if (res.ok) {
                const responseText = await res.text();
                // Check if it's the "no changes" message
                if (responseText.includes('No changes')) {
                    setSuccess('No changes to save');
                } else {
                    setSuccess('LDAP credentials saved successfully!');
                }
                // Update saved credentials display
                setSavedCredentials({
                    username: credentials.username,
                    password: '*****'
                });
                // Keep username, mask password
                setCredentials({
                    username: credentials.username,
                    password: '*****'
                });
            } else {
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to save LDAP credentials');
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to save LDAP credentials');
        } finally {
            setLoading(false);
        }
    };

    const handleTestConnection = async () => {
        setError('');
        setSuccess('');
        setTestLoading(true);
        setTestResult(null);
        setTestMessage('');

        // Use actual password if not masked, otherwise we can't test
        const testPassword = credentials.password === '*****' ? '' : credentials.password;

        if (!testPassword) {
            setError('Please enter the password to test connection');
            setTestLoading(false);
            return;
        }

        try {
            const res = await authFetch('/api/ldap/test', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    url: config.url,
                    baseDN: config.baseDN,
                    userDN: config.userDN,
                    username: credentials.username,
                    password: testPassword,
                }),
            });

            if (res.ok) {
                setTestResult('success');
                setTestMessage('Connection test successful');
            } else {
                const errorText = await parseErrorResponse(res);
                setTestResult('error');
                setTestMessage(errorText || 'LDAP connection test failed');
            }
        } catch (err) {
            setTestResult('error');
            setTestMessage(parseError(err) || 'LDAP connection test failed');
        } finally {
            setTestLoading(false);
        }
    };

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
