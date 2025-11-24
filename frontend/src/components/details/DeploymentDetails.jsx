import React from 'react';
import { Layers, Box, Network, HardDrive, Tag, Minus, Plus } from 'lucide-react';
import { DetailRow, EditYamlButton } from './CommonDetails';

const DeploymentDetails = ({ details, onScale, scaling, res, onEditYAML }) => {
    return (
        <div className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <div className="mb-2">
                        <DetailRow label="Replicas" value={`${details.ready} / ${details.replicas}`} icon={Layers}>
                            {onScale && (
                                <div className="flex items-center space-x-1 ml-2">
                                    <button
                                        onClick={() => onScale(-1)}
                                        disabled={scaling}
                                        className="p-1 rounded bg-gray-700 border border-gray-600 text-gray-200 hover:bg-gray-600 disabled:opacity-50"
                                        title="Scale down"
                                    >
                                        <Minus size={12} />
                                    </button>
                                    <button
                                        onClick={() => onScale(1)}
                                        disabled={scaling}
                                        className="p-1 rounded bg-gray-700 border border-gray-600 text-gray-200 hover:bg-gray-600 disabled:opacity-50"
                                        title="Scale up"
                                    >
                                        <Plus size={12} />
                                    </button>
                                </div>
                            )}
                        </DetailRow>
                    </div>
                    <DetailRow label="Images" value={details.images} icon={Box} />
                    <DetailRow label="Ports" value={details.ports?.map(p => p.toString())} icon={Network} />
                </div>
                <div>
                    <DetailRow label="PVCs" value={details.pvcs} icon={HardDrive} />
                    <DetailRow
                        label="Labels"
                        value={details.podLabels ? Object.entries(details.podLabels).map(([k, v]) => `${k}=${v}`) : []}
                        icon={Tag}
                    />
                </div>
            </div>
            <div className="flex justify-end mt-4">
                <EditYamlButton onClick={onEditYAML} />
            </div>
        </div>
    );
};

export default DeploymentDetails;
