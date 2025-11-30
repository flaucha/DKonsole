import React from 'react';
import { Tag, Layers, HardDrive, FileText } from 'lucide-react';
import { DetailRow, EditYamlButton } from './CommonDetails';

const StorageDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="flex justify-end mb-2">
            <EditYamlButton onClick={onEditYAML} />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <DetailRow label="Access Modes" value={details.accessModes} icon={Tag} />
            <DetailRow label="Storage Class" value={details.storageClassName} icon={Layers} />
            {details.requested && (
                <DetailRow label="Requested" value={details.requested} icon={HardDrive} />
            )}
            {details.capacity && (
                <DetailRow label="Capacity" value={details.capacity} icon={HardDrive} />
            )}
            {details.volumeName && <DetailRow label="Volume" value={details.volumeName} icon={HardDrive} />}
            {details.claimRef && (
                <DetailRow
                    label="Claim Ref"
                    value={`${details.claimRef.namespace}/${details.claimRef.name}`}
                    icon={FileText}
                />
            )}
        </div>
    </div>
);

export default StorageDetails;
