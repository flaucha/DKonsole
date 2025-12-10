import React from 'react';
import { Tag } from 'lucide-react';
import { DetailRow } from './CommonDetails';

const NetworkPolicyDetails = ({ details }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <DetailRow label="Policy Types" value={details.policyTypes} icon={Tag} />
        <DetailRow
            label="Pod Selector"
            value={details.podSelector?.matchLabels ? Object.entries(details.podSelector.matchLabels).map(([k, v]) => `${k}=${v}`) : (Object.keys(details.podSelector?.matchLabels || {}).length === 0 ? ['All Pods'] : [])}
            icon={Tag}
        />
    </div>
);

export default NetworkPolicyDetails;
