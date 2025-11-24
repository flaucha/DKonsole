import React from 'react';
import { DataSection, EditYamlButton } from './CommonDetails';

export const ConfigMapDetails = ({ details, onEditYAML }) => (
    <div className="p-6">
        <div className="flex items-center justify-between mb-4">
            <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Data</h4>
        </div>
        <DataSection data={details.data} />
        <div className="flex justify-end mt-6">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

export const SecretDetails = ({ details, onEditYAML }) => (
    <div className="p-6">
        <div className="flex items-center justify-between mb-4">
            <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Data</h4>
        </div>
        <DataSection data={details.data} isSecret={true} />
        <div className="flex justify-end mt-6">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);
