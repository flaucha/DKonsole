import React from 'react';

const GenericDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <pre className="text-xs text-gray-400 overflow-auto max-h-40">{JSON.stringify(details, null, 2)}</pre>
    </div>
);

export default GenericDetails;
