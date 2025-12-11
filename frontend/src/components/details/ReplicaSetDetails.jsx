import React, { useState } from 'react';
import { Layers, Box, Tag } from 'lucide-react';
import { DetailRow, SmartImage } from './CommonDetails';
import AssociatedPods from './AssociatedPods';

const ReplicaSetDetails = ({ details, res }) => {
    const [activeTab, setActiveTab] = useState('details');

    return (
        <div className="p-6">
            <div className="flex space-x-1 bg-gray-800/50 p-1 rounded-md mb-4 w-fit">
                <button
                    onClick={() => setActiveTab('details')}
                    className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${activeTab === 'details'
                        ? 'bg-gray-700 text-white shadow-sm'
                        : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800'
                        }`}
                >
                    Details
                </button>
                <button
                    onClick={() => setActiveTab('pods')}
                    className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${activeTab === 'pods'
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
                                <DetailRow label="Replicas" value={`${details.ready} / ${details.replicas}`} icon={Layers} />
                            </div>
                            <DetailRow
                                label="Images"
                                value={details.images ? details.images.map((img, i) => <SmartImage key={i} image={img} />) : []}
                                icon={Box}
                            />
                        </div>
                        <div>
                            <DetailRow
                                label="Labels"
                                value={details.labels ? Object.entries(details.labels).map(([k, v]) => `${k}=${v}`) : []}
                                icon={Tag}
                            />
                            <DetailRow
                                label="Pod Selector"
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

export default ReplicaSetDetails;
