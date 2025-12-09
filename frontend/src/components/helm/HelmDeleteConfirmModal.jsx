import React from 'react';

const HelmDeleteConfirmModal = ({ release, onCancel, onConfirm }) => {
    if (!release) return null;

    return (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                <h3 className="text-lg font-semibold text-white mb-2">
                    Confirm Uninstall
                </h3>
                <p className="text-sm text-gray-300 mb-4">
                    Are you sure you want to uninstall Helm release <span className="font-bold text-white">"{release.name}"</span> from namespace <span className="font-bold text-white">"{release.namespace}"</span>?
                    <br />
                    <span className="text-sm text-gray-500 mt-2 block">This action cannot be undone. All resources managed by this Helm release will be deleted.</span>
                </p>
                <div className="flex justify-end space-x-3">
                    <button
                        onClick={onCancel}
                        className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={onConfirm}
                        className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors"
                    >
                        Uninstall
                    </button>
                </div>
            </div>
        </div>
    );
};

export default HelmDeleteConfirmModal;
