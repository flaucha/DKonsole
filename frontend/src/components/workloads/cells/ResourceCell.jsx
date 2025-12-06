import React from 'react';

const ResourceCell = ({ value, label }) => (
    <span className="text-sm text-gray-400 text-center" title={label}>
        {value || '-'}
    </span>
);

export default ResourceCell;
