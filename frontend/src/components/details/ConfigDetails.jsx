import React from 'react';
import { DataSection, EditYamlButton } from './CommonDetails';

export const ConfigMapDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="flex items-center justify-between mb-2">
            <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Data</h4>
        </div>
        <DataSection data={details.data} />
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

export const SecretDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="flex items-center justify-between mb-2">
            <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Data</h4>
        </div>
        <DataSection data={details.data} isSecret={true} />
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);
