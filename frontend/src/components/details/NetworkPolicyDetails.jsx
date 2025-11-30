import React from 'react';
import { Tag } from 'lucide-react';
import { DetailRow } from './CommonDetails';

const NetworkPolicyDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <DetailRow label="Policy Types" value={details.policyTypes} icon={Tag} />
        <DetailRow
            label="Pod Selector"
            value={details.podSelector ? Object.entries(details.podSelector).map(([k, v]) => `${k}=${v}`) : []}
            icon={Tag}
        />
    </div>
);

export default NetworkPolicyDetails;
