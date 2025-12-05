import React, { useState, useEffect } from 'react';
import { Server, Settings as SettingsIcon, Users, Key, Save, Trash2, Plus, Check, AlertCircle, Lock } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { parseErrorResponse, parseError } from '../../utils/errorParser';

const LDAPSettings = () => {
    const { authFetch } = useAuth();
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
    const [credentials, setCredentials] = useState({
        username: '',
        password: ''
    });
    const [savedCredentials, setSavedCredentials] = useState({
        username: '',
        password: '*****'
    });
    const [groups, setGroups] = useState({ groups: [] });
    const [loading, setLoading] = useState(false);
    const [testLoading, setTestLoading] = useState(false);
    const [testResult, setTestResult] = useState(null); // null, 'success', 'error'
    const [testMessage, setTestMessage] = useState('');
    const [namespaces, setNamespaces] = useState([]);
    const [ldapActiveTab, setLdapActiveTab] = useState('config');
    const [adminGroups, setAdminGroups] = useState([]);
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
            loadGroups();
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

    const loadGroups = async () => {
        try {
            const res = await authFetch('/api/ldap/groups');
            if (res.ok) {
                const data = await res.json();
                setGroups(data);
            }
        } catch {
            // Ignore errors
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

    const handleSaveGroups = async () => {
        setError('');
        setSuccess('');
        setLoading(true);

        // Filter out incomplete permissions (empty namespace) before sending
        const groupsToSend = {
            groups: {
                groups: groups.groups.map(group => ({
                    ...group,
                    permissions: group.permissions.filter(perm => perm.namespace !== '')
                }))
            }
        };

        try {
            const res = await authFetch('/api/ldap/groups', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(groupsToSend),
            });

            if (res.ok) {
                setSuccess('LDAP groups configuration saved successfully!');
                // Update local state to remove incomplete permissions
                setGroups(groupsToSend.groups);
            } else {
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to save LDAP groups');
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to save LDAP groups');
        } finally {
            setLoading(false);
        }
    };

    const addGroup = () => {
        setGroups({
            groups: [...groups.groups, { name: '', permissions: [] }]
        });
    };

    const removeGroup = (index) => {
        setGroups({
            groups: groups.groups.filter((_, i) => i !== index)
        });
    };

    const updateGroup = (index, field, value) => {
        const newGroups = [...groups.groups];
        newGroups[index] = { ...newGroups[index], [field]: value };
        setGroups({ groups: newGroups });
    };

    const addPermission = (groupIndex) => {
        const newGroups = [...groups.groups];
        newGroups[groupIndex].permissions = [
            ...newGroups[groupIndex].permissions,
            { namespace: '', permission: 'view' }
        ];
        setGroups({ groups: newGroups });
    };

    const removePermission = (groupIndex, permIndex) => {
        const newGroups = [...groups.groups];
        newGroups[groupIndex].permissions = newGroups[groupIndex].permissions.filter((_, i) => i !== permIndex);
        setGroups({ groups: newGroups });
    };

    const updatePermission = (groupIndex, permIndex, field, value) => {
        const newGroups = [...groups.groups];
        newGroups[groupIndex].permissions[permIndex] = {
            ...newGroups[groupIndex].permissions[permIndex],
            [field]: value
        };
        setGroups({ groups: newGroups });
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
                        <div className="space-y-6">
                            {/* LDAP Server Configuration */}
                            <div>
                                <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                    <Server size={20} className="mr-2 text-blue-400" /> LDAP Server Configuration
                                </h2>
                                <div className="space-y-4">
                                    <div className="flex items-center space-x-3">
                                        <input
                                            type="checkbox"
                                            id="ldap-enabled"
                                            checked={config.enabled}
                                            onChange={(e) => handleEnabledChange(e.target.checked)}
                                            className="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
                                        />
                                        <label htmlFor="ldap-enabled" className="text-gray-300">Enable LDAP Authentication</label>
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">LDAP URL</label>
                                        <input
                                            type="text"
                                            value={config.url}
                                            onChange={(e) => setConfig({ ...config, url: e.target.value })}
                                            placeholder="ldap://ldap.example.com:389"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">Base DN</label>
                                        <input
                                            type="text"
                                            value={config.baseDN}
                                            onChange={(e) => setConfig({ ...config, baseDN: e.target.value })}
                                            placeholder="dc=example,dc=com"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">User DN Attribute</label>
                                        <input
                                            type="text"
                                            value={config.userDN}
                                            onChange={(e) => setConfig({ ...config, userDN: e.target.value })}
                                            placeholder="uid"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">Group DN</label>
                                        <input
                                            type="text"
                                            value={config.groupDN}
                                            onChange={(e) => setConfig({ ...config, groupDN: e.target.value })}
                                            placeholder="ou=groups,dc=example,dc=com"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">Required Group (Optional)</label>
                                        <input
                                            type="text"
                                            value={config.requiredGroup || ''}
                                            onChange={(e) => setConfig({ ...config, requiredGroup: e.target.value })}
                                            placeholder="access_group (leave empty to allow all authenticated users)"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                        <p className="text-xs text-gray-500 mt-1">
                                            Only users in this group will be allowed to login. Leave empty to allow all authenticated LDAP users.
                                        </p>
                                    </div>

                                    {/* TLS Configuration */}
                                    <div className="border-t border-gray-700 pt-4">
                                        <h3 className="text-md font-semibold text-white mb-3 flex items-center">
                                            <Lock size={18} className="mr-2 text-yellow-400" /> TLS Configuration
                                        </h3>

                                        <div className="flex items-center space-x-3 mb-4">
                                            <input
                                                type="checkbox"
                                                id="ldap-insecure-skip-verify"
                                                checked={config.insecureSkipVerify || false}
                                                onChange={(e) => setConfig({ ...config, insecureSkipVerify: e.target.checked })}
                                                className="w-4 h-4 text-yellow-600 bg-gray-700 border-gray-600 rounded focus:ring-yellow-500"
                                            />
                                            <label htmlFor="ldap-insecure-skip-verify" className="text-gray-300">
                                                Skip TLS Certificate Verification
                                            </label>
                                        </div>
                                        {config.insecureSkipVerify && (
                                            <div className="mb-4 p-3 bg-yellow-900/20 border border-yellow-500/50 rounded-lg">
                                                <div className="flex items-start space-x-2">
                                                    <AlertCircle size={16} className="text-yellow-400 mt-0.5 flex-shrink-0" />
                                                    <p className="text-xs text-yellow-400">
                                                        <strong>Warning:</strong> This option disables TLS certificate verification.
                                                        This should only be used in development or with self-signed certificates.
                                                        It makes the connection vulnerable to man-in-the-middle attacks.
                                                    </p>
                                                </div>
                                            </div>
                                        )}

                                        <div>
                                            <label className="block text-sm font-medium text-gray-400 mb-2">CA Certificate (PEM format, optional)</label>
                                            <textarea
                                                value={config.caCert || ''}
                                                onChange={(e) => setConfig({ ...config, caCert: e.target.value })}
                                                placeholder="-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
                                                rows={6}
                                                className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 font-mono text-xs"
                                            />
                                            <p className="text-xs text-gray-500 mt-1">
                                                Paste the CA certificate in PEM format for TLS verification. Leave empty to use system CA certificates.
                                            </p>
                                        </div>
                                    </div>

                                    <button
                                        onClick={handleSaveConfig}
                                        disabled={loading}
                                        className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                                    >
                                        <Save size={16} className="mr-2" />
                                        {loading ? 'Saving...' : 'Save Configuration'}
                                    </button>
                                </div>
                            </div>

                            {/* LDAP Service Account Credentials */}
                            <div className="border-t border-gray-700 pt-6">
                                <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                    <Key size={20} className="mr-2 text-yellow-400" /> LDAP Service Account Credentials
                                </h2>
                                <p className="text-sm text-gray-400 mb-4">
                                    Credentials for the service account used to query LDAP for user groups.
                                </p>
                                <div className="space-y-4">
                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">Username</label>
                                        <input
                                            type="text"
                                            value={credentials.username}
                                            onChange={(e) => setCredentials({ ...credentials, username: e.target.value })}
                                            placeholder="service-account"
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-400 mb-2">Password</label>
                                        <input
                                            type="password"
                                            value={credentials.password}
                                            onChange={(e) => {
                                                setCredentials({ ...credentials, password: e.target.value });
                                            }}
                                            onFocus={() => {
                                                // Clear mask when user focuses on the field
                                                if (credentials.password === '*****') {
                                                    setCredentials({ ...credentials, password: '' });
                                                }
                                            }}
                                            placeholder={savedCredentials.username ? "Enter new password or leave blank to keep current" : "Enter password"}
                                            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                    </div>

                                    <div className="flex items-center space-x-2">
                                        <button
                                            onClick={handleSaveCredentials}
                                            disabled={loading}
                                            className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                                        >
                                            <Save size={16} className="mr-2" />
                                            {loading ? 'Saving...' : 'Save Credentials'}
                                        </button>
                                        <button
                                            onClick={handleTestConnection}
                                            disabled={testLoading || !config.url || !credentials.username || !credentials.password}
                                            className="flex items-center px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                                        >
                                            <Check size={16} className="mr-2" />
                                            {testLoading ? 'Testing...' : 'Test Connection'}
                                        </button>
                                        {testResult && (
                                            <div className={`flex items-center space-x-2 px-3 py-2 rounded-lg ${testResult === 'success'
                                                    ? 'bg-green-900/20 border border-green-500/50 text-green-400'
                                                    : 'bg-red-900/20 border border-red-500/50 text-red-400'
                                                }`}>
                                                {testResult === 'success' ? (
                                                    <>
                                                        <Check size={16} />
                                                        <span className="text-sm">{testMessage}</span>
                                                    </>
                                                ) : (
                                                    <>
                                                        <AlertCircle size={16} />
                                                        <span className="text-sm">{testMessage}</span>
                                                    </>
                                                )}
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}

                    {ldapActiveTab === 'permissions' && (
                        <div>
                            <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                <Users size={20} className="mr-2 text-purple-400" /> LDAP Groups and Permissions
                            </h2>
                            <p className="text-sm text-gray-400 mb-4">
                                Configure which LDAP groups have access to which namespaces and with what permissions.
                            </p>

                            <div className="space-y-4">
                                {groups.groups.map((group, groupIndex) => (
                                    <div key={groupIndex} className="bg-gray-900 p-4 rounded-lg border border-gray-700">
                                        <div className="flex items-center justify-between mb-3">
                                            <input
                                                type="text"
                                                value={group.name}
                                                onChange={(e) => updateGroup(groupIndex, 'name', e.target.value)}
                                                placeholder="Group name (e.g., developers)"
                                                className="flex-1 px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                            />
                                            <button
                                                onClick={() => removeGroup(groupIndex)}
                                                className="ml-2 px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
                                            >
                                                <Trash2 size={16} />
                                            </button>
                                        </div>

                                        <div className="space-y-2">
                                            {group.permissions.map((perm, permIndex) => (
                                                <div key={permIndex} className="flex items-center space-x-2">
                                                    <select
                                                        value={perm.namespace}
                                                        onChange={(e) => updatePermission(groupIndex, permIndex, 'namespace', e.target.value)}
                                                        className="flex-1 px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-blue-500"
                                                    >
                                                        <option value="">Select namespace</option>
                                                        {namespaces.map(ns => (
                                                            <option key={ns} value={ns}>{ns}</option>
                                                        ))}
                                                    </select>
                                                    <select
                                                        value={perm.permission}
                                                        onChange={(e) => updatePermission(groupIndex, permIndex, 'permission', e.target.value)}
                                                        className="px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-blue-500"
                                                    >
                                                        <option value="view">View</option>
                                                        <option value="edit">Edit</option>
                                                    </select>
                                                    <button
                                                        onClick={() => removePermission(groupIndex, permIndex)}
                                                        className="px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
                                                    >
                                                        <Trash2 size={14} />
                                                    </button>
                                                </div>
                                            ))}
                                            <button
                                                onClick={() => addPermission(groupIndex)}
                                                className="flex items-center px-3 py-2 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded-lg transition-colors text-sm"
                                            >
                                                <Plus size={14} className="mr-2" /> Add Permission
                                            </button>
                                        </div>
                                    </div>
                                ))}

                                <button
                                    onClick={addGroup}
                                    className="flex items-center px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded-lg transition-colors"
                                >
                                    <Plus size={16} className="mr-2" /> Add Group
                                </button>

                                <button
                                    onClick={handleSaveGroups}
                                    disabled={loading}
                                    className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                                >
                                    <Save size={16} className="mr-2" />
                                    {loading ? 'Saving...' : 'Save Groups Configuration'}
                                </button>
                            </div>
                        </div>
                    )}

                    {ldapActiveTab === 'admins' && (
                        <div>
                            <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                                <Key size={20} className="mr-2 text-yellow-400" /> LDAP Group Admins
                            </h2>
                            <p className="text-sm text-gray-400 mb-4">
                                Configure which LDAP groups have cluster admin access. Members of these groups will have full access to all namespaces and settings.
                            </p>

                            <div className="space-y-4">
                                {adminGroups.map((group, index) => (
                                    <div key={index} className="flex items-center space-x-2 bg-gray-900 p-4 rounded-lg border border-gray-700">
                                        <input
                                            type="text"
                                            value={group}
                                            onChange={(e) => {
                                                const newGroups = [...adminGroups];
                                                newGroups[index] = e.target.value;
                                                setAdminGroups(newGroups);
                                            }}
                                            placeholder="Group name (e.g., admin)"
                                            className="flex-1 px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                                        />
                                        <button
                                            onClick={() => {
                                                const newGroups = adminGroups.filter((_, i) => i !== index);
                                                setAdminGroups(newGroups);
                                            }}
                                            className="px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
                                        >
                                            <Trash2 size={16} />
                                        </button>
                                    </div>
                                ))}

                                <button
                                    onClick={() => setAdminGroups([...adminGroups, ''])}
                                    className="flex items-center px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors"
                                >
                                    <Plus size={16} className="mr-2" /> Add Admin Group
                                </button>

                                <button
                                    onClick={async () => {
                                        setError('');
                                        setSuccess('');
                                        setLoading(true);

                                        try {
                                            // Update config with admin groups
                                            const updatedConfig = { ...config, adminGroups: adminGroups.filter(g => g.trim() !== '') };
                                            const res = await authFetch('/api/ldap/config', {
                                                method: 'PUT',
                                                headers: { 'Content-Type': 'application/json' },
                                                body: JSON.stringify({ config: updatedConfig }),
                                            });

                                            if (res.ok) {
                                                setSuccess('Admin groups updated successfully!');
                                                setConfig(updatedConfig);
                                            } else {
                                                const errorText = await parseErrorResponse(res);
                                                setError(errorText || 'Failed to update admin groups');
                                            }
                                        } catch (err) {
                                            setError(parseError(err) || 'Failed to update admin groups');
                                        } finally {
                                            setLoading(false);
                                        }
                                    }}
                                    disabled={loading}
                                    className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                                >
                                    <Save size={16} className="mr-2" />
                                    {loading ? 'Saving...' : 'Save Admin Groups'}
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>

            {error && (
                <div className="flex items-center text-red-400 text-sm bg-red-900/20 border border-red-500/50 rounded-lg p-3">
                    <AlertCircle size={16} className="mr-2" />
                    {error}
                </div>
            )}
            {success && (
                <div className="flex items-center text-green-400 text-sm bg-green-900/20 border border-green-500/50 rounded-lg p-3">
                    <Check size={16} className="mr-2" />
                    {success}
                </div>
            )}
        </div>
    );
};

export default LDAPSettings;
