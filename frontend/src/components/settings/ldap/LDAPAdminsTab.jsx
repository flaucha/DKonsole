import React, { useState } from 'react';
import { Key, Trash2, Plus, Save, AlertCircle, Check } from 'lucide-react';
import { useAuth } from '../../../context/AuthContext';
import { parseErrorResponse, parseError } from '../../../utils/errorParser';

const LDAPAdminsTab = ({ config, setConfig, adminGroups, setAdminGroups }) => {
    const { authFetch } = useAuth();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    const handleSaveAdminGroups = async () => {
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
    };

    return (
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
                    onClick={handleSaveAdminGroups}
                    disabled={loading}
                    className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                >
                    <Save size={16} className="mr-2" />
                    {loading ? 'Saving...' : 'Save Admin Groups'}
                </button>

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
        </div>
    );
};

export default LDAPAdminsTab;
