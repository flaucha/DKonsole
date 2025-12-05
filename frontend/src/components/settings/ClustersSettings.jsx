import React from 'react';
import { Server } from 'lucide-react';
import { useSettings } from '../../context/SettingsContext';

const ClustersSettings = () => {
    const { clusters, currentCluster, setCurrentCluster } = useSettings();

    return (
        <div className="w-full space-y-6">
            <div className="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-lg">
                <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                    <Server size={20} className="mr-2 text-blue-400" /> Configured Clusters
                </h2>
                <div className="space-y-3">
                    {clusters.map(cluster => (
                        <div key={cluster} className={`flex items-center justify-between p-4 rounded-lg border transition-all ${cluster === currentCluster ? 'bg-blue-900/20 border-blue-500/50' : 'bg-gray-750 border-gray-700'}`}>
                            <div className="flex items-center">
                                <div className={`w-2 h-2 rounded-full mr-3 ${cluster === currentCluster ? 'bg-green-400 shadow-[0_0_8px_rgba(74,222,128,0.5)]' : 'bg-gray-500'}`}></div>
                                <span className={`font-medium ${cluster === currentCluster ? 'text-white' : 'text-gray-300'}`}>{cluster}</span>
                            </div>
                            {cluster !== currentCluster ? (
                                <button
                                    onClick={() => setCurrentCluster(cluster)}
                                    className="text-xs bg-gray-700 hover:bg-gray-600 text-blue-300 px-3 py-1.5 rounded transition-colors"
                                >
                                    Switch
                                </button>
                            ) : (
                                <span className="text-xs bg-blue-500/20 text-blue-300 px-3 py-1 rounded-full border border-blue-500/30">Active</span>
                            )}
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default ClustersSettings;
