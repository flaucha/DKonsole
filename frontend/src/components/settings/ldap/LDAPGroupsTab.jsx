import React, { useState, useEffect } from 'react';
import { Users, Trash2, Plus, Save, AlertCircle, Check } from 'lucide-react';
import { useAuth } from '../../../context/AuthContext';
import { parseErrorResponse, parseError } from '../../../utils/errorParser';

const LDAPGroupsTab = ({ namespaces }) => {
    const { authFetch } = useAuth();
    const [groups, setGroups] = useState({ groups: [] });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    useEffect(() => {
        loadGroups();
    }, []);

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

    return (
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

export default LDAPGroupsTab;
