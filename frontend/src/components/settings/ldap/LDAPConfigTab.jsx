import React from 'react';
import { Server, Lock, AlertCircle, Save, Key, Check } from 'lucide-react';

const LDAPConfigTab = ({
    config,
    setConfig,
    credentials,
    setCredentials,
    savedCredentials,
    loading,
    testLoading,
    testResult,
    testMessage,
    handleSaveConfig,
    handleEnabledChange,
    handleSaveCredentials,
    handleTestConnection
}) => {
    return (
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
    );
};

export default LDAPConfigTab;
