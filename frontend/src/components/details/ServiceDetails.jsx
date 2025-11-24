import React from 'react';
import { Tag } from 'lucide-react';
import { EditYamlButton } from './CommonDetails';

const ServiceDetails = ({ details, onEditYAML }) => {
    const type = details.type || 'ClusterIP';
    const clusterIP = details.clusterIP;
    const externalIPs = details.externalIPs || [];
    const ports = details.ports || [];
    const selector = details.selector || {};

    const getTypeColor = (t) => {
        switch (t) {
            case 'LoadBalancer': return 'text-blue-400 border-blue-400/30 bg-blue-400/10';
            case 'NodePort': return 'text-purple-400 border-purple-400/30 bg-purple-400/10';
            default: return 'text-gray-400 border-gray-600 bg-gray-800';
        }
    };

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2 space-y-4">
            <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                    <span className={`px-2 py-0.5 rounded text-xs font-medium border ${getTypeColor(type)}`}>
                        {type}
                    </span>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">IP Addresses</h4>
                    <div className="space-y-2">
                        <div className="flex items-center justify-between bg-gray-800 px-3 py-2 rounded border border-gray-700">
                            <span className="text-xs text-gray-400">Cluster IP</span>
                            <span className="text-sm font-mono text-white">{clusterIP}</span>
                        </div>
                        {externalIPs.map((ip, i) => (
                            <div key={i} className="flex items-center justify-between bg-gray-800 px-3 py-2 rounded border border-gray-700">
                                <span className="text-xs text-gray-400">External IP</span>
                                <span className="text-sm font-mono text-white">{ip}</span>
                            </div>
                        ))}
                    </div>
                </div>

                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Selector</h4>
                    <div className="flex flex-wrap gap-2">
                        {Object.keys(selector).length > 0 ? (
                            Object.entries(selector).map(([k, v]) => (
                                <div key={k} className="flex items-center px-2 py-1 bg-gray-800 border border-gray-700 rounded text-xs text-gray-300">
                                    <Tag size={12} className="mr-1.5 text-gray-500" />
                                    <span className="text-gray-400 mr-1">{k}:</span>
                                    <span className="text-white">{v}</span>
                                </div>
                            ))
                        ) : (
                            <span className="text-sm text-gray-500 italic">No selector</span>
                        )}
                    </div>
                </div>
            </div>

            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Ports</h4>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
                    {ports.map((port, i) => {
                        if (typeof port !== 'string') return null;
                        // Parse "80:30080/TCP" or similar format
                        const parts = port.split('/');
                        const protocol = parts[1] || 'TCP';
                        const portMap = parts[0].split(':');
                        const portNum = portMap[0];
                        const targetPort = portMap[1] || portNum;

                        return (
                            <div key={i} className="flex items-center justify-between bg-gray-800 px-3 py-2 rounded border border-gray-700">
                                <div className="flex items-center">
                                    <div className="w-1.5 h-1.5 rounded-full bg-green-500 mr-2"></div>
                                    <span className="text-sm font-mono text-white">{portNum}</span>
                                    {portNum !== targetPort && (
                                        <>
                                            <span className="text-gray-500 mx-1">→</span>
                                            <span className="text-xs text-gray-400">{targetPort}</span>
                                        </>
                                    )}
                                </div>
                                <span className="text-[10px] font-bold text-gray-500 bg-gray-900 px-1.5 py-0.5 rounded">
                                    {protocol}
                                </span>
                            </div>
                        );
                    })}
                </div>
            </div>
            <div className="flex justify-end mt-4">
                <EditYamlButton onClick={onEditYAML} />
            </div>
        </div>
    );
};

export default ServiceDetails;
