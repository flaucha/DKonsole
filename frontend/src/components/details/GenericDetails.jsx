import React from 'react';
import { EditYamlButton } from './CommonDetails';

const GenericDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <pre className="text-xs text-gray-400 overflow-auto max-h-40">{JSON.stringify(details, null, 2)}</pre>
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

export default GenericDetails;
