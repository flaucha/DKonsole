import React, { useState } from 'react';
import { Edit2 } from 'lucide-react';
import { DataSection } from './CommonDetails';
import { DataEditor } from './DataEditor';
import { useAuth } from '../../context/AuthContext';
import { canEdit, isAdmin } from '../../utils/permissions';

export const ConfigMapDetails = ({ details, onEditYAML, resource, onDataSaved }) => {
    const { user } = useAuth();
    const [editingData, setEditingData] = useState(false);

    return (
        <>
            <div className="p-6">
                <div className="flex items-center justify-between mb-4">
                    <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Data</h4>
                </div>
                <DataSection data={details.data} />
                {(isAdmin(user) || canEdit(user, resource?.namespace)) && (
                    <div className="flex justify-end gap-2 mt-6">
                        <button
                            onClick={() => setEditingData(true)}
                            className="flex items-center px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-gray-200 rounded-md border border-gray-600 text-xs transition-colors font-medium"
                        >
                            <Edit2 size={14} className="mr-1.5" />
                            Edit in place
                        </button>
                    </div>
                )}
            </div>
            {editingData && (
                <DataEditor
                    resource={resource}
                    data={details.data}
                    isSecret={false}
                    onClose={() => setEditingData(false)}
                    onSaved={() => {
                        setEditingData(false);
                        if (onDataSaved) {
                            onDataSaved();
                        }
                    }}
                />
            )}
        </>
    );
};

export const SecretDetails = ({ details, onEditYAML, resource, onDataSaved }) => {
    const { user } = useAuth();
    const [editingData, setEditingData] = useState(false);

    return (
        <>
            <div className="p-6">
                <div className="flex items-center justify-between mb-4">
                    <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Data</h4>
                </div>
                <DataSection data={details.data} isSecret={true} />
                {(isAdmin(user) || canEdit(user, resource?.namespace)) && (
                    <div className="flex justify-end gap-2 mt-6">
                        <button
                            onClick={() => setEditingData(true)}
                            className="flex items-center px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-gray-200 rounded-md border border-gray-600 text-xs transition-colors font-medium"
                        >
                            <Edit2 size={14} className="mr-1.5" />
                            Edit in place
                        </button>
                    </div>
                )}
            </div>
            {editingData && (
                <DataEditor
                    resource={resource}
                    data={details.data}
                    isSecret={true}
                    onClose={() => setEditingData(false)}
                    onSaved={() => {
                        setEditingData(false);
                        if (onDataSaved) {
                            onDataSaved();
                        }
                    }}
                />
            )}
        </>
    );
};
