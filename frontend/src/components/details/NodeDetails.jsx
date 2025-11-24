import React from 'react';
import { Server, HardDrive, Box, Activity, Network } from 'lucide-react';
import { DetailRow, EditYamlButton } from './CommonDetails';

const NodeDetails = ({ details, onEditYAML }) => {
    const addresses = details.addresses || [];
    const nodeInfo = details.nodeInfo || {};
    const conditions = details.conditions || {};
    const taints = details.taints || [];

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2 space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">System Info</h4>
                    <DetailRow label="Kernel" value={nodeInfo.kernelVersion} icon={Server} />
                    <DetailRow label="OS Image" value={nodeInfo.osImage} icon={HardDrive} />
                    <DetailRow label="Runtime" value={nodeInfo.containerRuntimeVersion} icon={Box} />
                    <DetailRow label="Kubelet" value={nodeInfo.kubeletVersion} icon={Activity} />
                    <DetailRow label="Arch" value={nodeInfo.architecture} icon={Server} />
                </div>
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Addresses</h4>
                    {addresses.map((addr, i) => (
                        <DetailRow key={i} label={addr.type} value={addr.address} icon={Network} />
                    ))}
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Capacity</h4>
                    <DetailRow label="CPU" value={details.capacity?.cpu} icon={Activity} />
                    <DetailRow label="Memory" value={details.capacity?.memory} icon={HardDrive} />
                    <DetailRow label="Pods" value={details.capacity?.pods} icon={Box} />
                </div>
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Allocatable</h4>
                    <DetailRow label="CPU" value={details.allocatable?.cpu} icon={Activity} />
                    <DetailRow label="Memory" value={details.allocatable?.memory} icon={HardDrive} />
                    <DetailRow label="Pods" value={details.allocatable?.pods} icon={Box} />
                </div>
            </div>

            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Conditions</h4>
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-2">
                    {Object.entries(conditions).map(([type, status]) => (
                        <div key={type} className={`px-3 py-2 rounded border ${status === 'True' && type === 'Ready' ? 'bg-green-900/20 border-green-800 text-green-400' : status === 'True' ? 'bg-red-900/20 border-red-800 text-red-400' : 'bg-gray-800 border-gray-700 text-gray-400'}`}>
                            <div className="text-xs font-medium">{type}</div>
                            <div className="text-xs mt-1">{status}</div>
                        </div>
                    ))}
                </div>
            </div>

            {taints.length > 0 && (
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Taints</h4>
                    <div className="flex flex-wrap gap-2">
                        {taints.map((t, i) => (
                            <span key={i} className="px-2 py-1 bg-yellow-900/20 border border-yellow-800 text-yellow-500 rounded text-xs">
                                {t.key}{t.value ? `=${t.value}` : ''}:{t.effect}
                            </span>
                        ))}
                    </div>
                </div>
            )}
            <div className="flex justify-end mt-4">
                <EditYamlButton onClick={onEditYAML} />
            </div>
        </div>
    );
};

export default NodeDetails;
