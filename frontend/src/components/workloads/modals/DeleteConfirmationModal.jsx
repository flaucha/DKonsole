import React from 'react';

const DeleteConfirmationModal = ({ confirmAction, setConfirmAction, handleDelete }) => {
    if (!confirmAction) return null;

    return (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                <h3 className="text-lg font-semibold text-white mb-2">
                    Confirm delete
                </h3>
                <p className="text-sm text-gray-300 mb-4">
                    {confirmAction.force ? 'Force delete' : 'Delete'} {confirmAction.res.kind} "{confirmAction.res.name}"?
                    {confirmAction.res.kind === 'Node' && (
                        <span className="block mt-2 text-xs text-yellow-400">
                            ⚠️ Warning: Deleting a node will remove it from the cluster. This action cannot be undone.
                        </span>
                    )}
                    {confirmAction.force && (
                        <span className="block mt-2 text-xs text-red-400">
                            Warning: Force delete will immediately terminate the resource without graceful shutdown.
                        </span>
                    )}
                </p>
                <div className="flex justify-end space-x-3">
                    <button
                        onClick={() => setConfirmAction(null)}
                        className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={() => handleDelete(confirmAction.res, confirmAction.force)}
                        className={`px-4 py-2 rounded-md text-white transition-colors ${confirmAction.force
                            ? 'bg-red-700 hover:bg-red-800'
                            : 'bg-orange-600 hover:bg-orange-700'
                            }`}
                    >
                        {confirmAction.force ? 'Force Delete' : 'Delete'}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default DeleteConfirmationModal;
