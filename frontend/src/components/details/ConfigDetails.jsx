import React, { useState } from 'react';
import { DataSection } from './CommonDetails';
import { DataEditor } from './DataEditor';


export const ConfigMapDetails = ({ details, resource, onDataSaved }) => {
    const [editingData, setEditingData] = useState(false);

    return (
        <>
            <div className="p-6">
                <div className="flex items-center justify-between mb-4">
                    <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Data</h4>
                </div>
                <DataSection data={details.data} />

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

export const SecretDetails = ({ details, resource, onDataSaved }) => {
    const [editingData, setEditingData] = useState(false);

    return (
        <>
            <div className="p-6">
                <div className="flex items-center justify-between mb-4">
                    <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Data</h4>
                </div>
                <DataSection data={details.data} isSecret={true} />

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
