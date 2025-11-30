import React from 'react';
import { HardDrive, Activity, Network, Layers, Tag } from 'lucide-react';
import { DetailRow, EditYamlButton } from './CommonDetails';

const StorageClassDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="flex justify-end mb-2">
            <EditYamlButton onClick={onEditYAML} />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <DetailRow label="Provisioner" value={details.provisioner} icon={HardDrive} />
            <DetailRow label="Reclaim Policy" value={details.reclaimPolicy} icon={Activity} />
            <DetailRow label="Binding Mode" value={details.volumeBindingMode} icon={Network} />
            <DetailRow label="Volume Expansion" value={String(details.allowVolumeExpansion)} icon={Layers} />
            <DetailRow
                label="Parameters"
                value={details.parameters ? Object.entries(details.parameters).map(([k, v]) => `${k}=${v}`) : []}
                icon={Tag}
            />
            <DetailRow label="Mount Options" value={details.mountOptions} icon={HardDrive} />
        </div>
    </div>
);

export default StorageClassDetails;
