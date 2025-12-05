import React, { useState } from 'react';
import { Lock, AlertCircle, Check, Settings as SettingsIcon } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { parseErrorResponse, parseError } from '../../utils/errorParser';

const PasswordSettings = () => {
    const { authFetch } = useAuth();
    const [currentPassword, setCurrentPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [showConfirmDialog, setShowConfirmDialog] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    const validateFields = () => {
        setError('');
        setSuccess('');

        // Validation
        if (!currentPassword || !newPassword || !confirmPassword) {
            setError('All fields are required');
            return false;
        }

        if (newPassword.length < 8) {
            setError('New password must be at least 8 characters long');
            return false;
        }

        if (newPassword !== confirmPassword) {
            setError('New passwords do not match');
            return false;
        }

        return true;
    };

    const handleChangePasswordClick = () => {
        if (validateFields()) {
            setShowConfirmDialog(true);
        }
    };

    const handleConfirmChangePassword = async () => {
        setLoading(true);
        setShowConfirmDialog(false);

        try {
            const res = await authFetch('/api/auth/change-password', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    currentPassword,
                    newPassword,
                }),
            });

            if (res.ok) {
                setSuccess('Password changed successfully! You will be logged out...');
                // Logout and redirect to login
                setTimeout(() => {
                    // Clear auth token
                    document.cookie = 'token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
                    // Reload to trigger login
                    window.location.href = '/login';
                }, 2000);
            } else {
                const errorText = await parseErrorResponse(res);
                setError(errorText || 'Failed to change password');
                setLoading(false);
            }
        } catch (err) {
            setError(parseError(err) || 'Failed to change password');
            setLoading(false);
        }
    };

    return (
        <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
            <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                <SettingsIcon size={20} className="mr-2 text-red-400" /> Change Password
            </h2>
            <div className="space-y-4">
                <div className="space-y-3">
                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">Current Password</label>
                        <div className="relative">
                            <Lock size={18} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500" />
                            <input
                                type="password"
                                value={currentPassword}
                                onChange={(e) => setCurrentPassword(e.target.value)}
                                placeholder="Enter current password"
                                className="w-full pl-10 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">New Password</label>
                        <div className="relative">
                            <Lock size={18} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500" />
                            <input
                                type="password"
                                value={newPassword}
                                onChange={(e) => setNewPassword(e.target.value)}
                                placeholder="Enter new password (min 8 characters)"
                                className="w-full pl-10 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">Confirm New Password</label>
                        <div className="relative">
                            <Lock size={18} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500" />
                            <input
                                type="password"
                                value={confirmPassword}
                                onChange={(e) => setConfirmPassword(e.target.value)}
                                placeholder="Confirm new password"
                                className="w-full pl-10 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                            />
                        </div>
                    </div>
                </div>

                <button
                    onClick={handleChangePasswordClick}
                    disabled={loading}
                    className="w-full flex items-center justify-center px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                >
                    <Lock size={16} className="mr-2" />
                    {loading ? 'Changing Password...' : 'Change Password'}
                </button>

                {error && (
                    <div className="flex items-center text-red-400 text-sm">
                        <AlertCircle size={16} className="mr-2" />
                        {error}
                    </div>
                )}
                {success && (
                    <div className="flex items-center text-green-400 text-sm">
                        <Check size={16} className="mr-2" />
                        {success}
                    </div>
                )}

                {/* Confirmation Dialog */}
                {showConfirmDialog && (
                    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                        <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                            <div className="flex items-center space-x-3 mb-4 text-red-400">
                                <AlertCircle size={24} />
                                <h3 className="text-xl font-bold text-white">Confirm Password Change</h3>
                            </div>
                            <p className="text-gray-300 mb-2 leading-relaxed">
                                Are you sure you want to change your password?
                            </p>
                            <p className="text-sm text-yellow-400 mb-6">
                                ⚠️ Warning: After changing your password, you will be automatically logged out and redirected to the login page.
                            </p>
                            <div className="flex justify-end space-x-3">
                                <button
                                    onClick={() => setShowConfirmDialog(false)}
                                    className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                                >
                                    Cancel
                                </button>
                                <button
                                    onClick={handleConfirmChangePassword}
                                    className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors"
                                >
                                    Change Password
                                </button>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default PasswordSettings;
