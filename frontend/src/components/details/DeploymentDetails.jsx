import React, { useState } from 'react';
import { Layers, Box, Network, HardDrive, Tag, Minus, Plus } from 'lucide-react';
import { DetailRow, SmartImage } from './CommonDetails';
import AssociatedPods from './AssociatedPods';
import { useAuth } from '../../context/AuthContext';
import { canEdit, isAdmin } from '../../utils/permissions';

const DeploymentDetails = ({ details, onScale, scaling, res, onEditYAML }) => {
    const { user } = useAuth();
    const [activeTab, setActiveTab] = useState('details');

    return (
        <div className="p-6">
            <div className="flex space-x-1 bg-gray-800/50 p-1 rounded-md mb-4 w-fit">
                <button
                    onClick={() => setActiveTab('details')}
                    className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${
                        activeTab === 'details'
                        ? 'bg-gray-700 text-white shadow-sm'
                        : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800'
                    }`}
                >
                    Details
                </button>
                <button
                    onClick={() => setActiveTab('pods')}
                    className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${
                        activeTab === 'pods'
                        ? 'bg-gray-700 text-white shadow-sm'
                        : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800'
                    }`}
                >
                    Pod List
                </button>
            </div>

            {activeTab === 'details' && (
                <>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <div className="mb-2">
                                <DetailRow label="Replicas" value={`${details.ready} / ${details.replicas}`} icon={Layers}>
                                    {onScale && (isAdmin(user) || canEdit(user, res?.namespace)) && (
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
                            <DetailRow
                                label="Images"
                                value={details.images ? details.images.map((img, i) => <SmartImage key={i} image={img} />) : []}
                                icon={Box}
                            />
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
                </>
            )}

            {activeTab === 'pods' && (
                <AssociatedPods namespace={res.namespace} selector={details.podLabels} />
            )}
        </div>
    );
};

export default DeploymentDetails;
