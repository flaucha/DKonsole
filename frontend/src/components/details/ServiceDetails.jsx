import React, { useState } from 'react';
import { Tag } from 'lucide-react';
import { SmartDNS, EditYamlButton } from './CommonDetails';
import AssociatedPods from './AssociatedPods';

const ServiceDetails = ({ details, onEditYAML, namespace, name }) => {
    const [activeTab, setActiveTab] = useState('details');
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
        <div className="p-6 space-y-6">
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
                    onClick={() => setActiveTab('selector')}
                    className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${
                        activeTab === 'selector'
                        ? 'bg-gray-700 text-white shadow-sm'
                        : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800'
                    }`}
                >
                    Selector
                </button>
            </div>

            {activeTab === 'details' && (
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center space-x-2">
                        <span className={`px-2 py-0.5 rounded text-xs font-medium border ${getTypeColor(type)}`}>
                            {type}
                        </span>
                    </div>
                </div>
            )}

            {activeTab === 'details' && (
                <>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">IP Addresses</h4>
                            <div className="space-y-2">
                                {name && namespace && (
                                    <div className="flex items-center justify-between bg-gray-800/50 px-4 py-3 rounded-md border border-gray-700/50 hover:bg-gray-800/70 transition-colors">
                                        <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">DNS</span>
                                        <div className="flex-1 flex justify-end">
                                            <SmartDNS dns={`${name}.${namespace}.svc.cluster.local`} />
                                        </div>
                                    </div>
                                )}
                                <div className="flex items-center justify-between bg-gray-800/50 px-4 py-3 rounded-md border border-gray-700/50 hover:bg-gray-800/70 transition-colors">
                                    <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">Cluster IP</span>
                                    <span className="text-sm font-mono text-gray-200">{clusterIP}</span>
                                </div>
                                {externalIPs.map((ip, i) => (
                                    <div key={i} className="flex items-center justify-between bg-gray-800/50 px-4 py-3 rounded-md border border-gray-700/50 hover:bg-gray-800/70 transition-colors">
                                        <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">External IP</span>
                                        <span className="text-sm font-mono text-gray-200">{ip}</span>
                                    </div>
                                ))}
                            </div>
                        </div>

                        <div>
                            <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Selector</h4>
                            <div className="flex flex-wrap gap-2">
                                {Object.keys(selector).length > 0 ? (
                                    Object.entries(selector).map(([k, v]) => (
                                        <div key={k} className="flex items-center px-3 py-1.5 bg-gray-800/50 border border-gray-700/50 rounded-md text-xs text-gray-200 hover:bg-gray-800/70 transition-colors">
                                            <Tag size={12} className="mr-1.5 text-gray-400" />
                                            <span className="text-gray-400 mr-1">{k}:</span>
                                            <span className="text-gray-200">{v}</span>
                                        </div>
                                    ))
                                ) : (
                                    <span className="text-sm text-gray-500 italic">No selector</span>
                                )}
                            </div>
                        </div>
                    </div>

                    <div>
                        <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Ports</h4>
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
                                    <div key={i} className="flex items-center justify-between bg-gray-800/50 px-4 py-3 rounded-md border border-gray-700/50 hover:bg-gray-800/70 transition-colors">
                                        <div className="flex items-center">
                                            <div className="w-1.5 h-1.5 rounded-full bg-green-500 mr-2"></div>
                                            <span className="text-sm font-mono text-gray-200">{portNum}</span>
                                            {portNum !== targetPort && (
                                                <>
                                                    <span className="text-gray-500 mx-1">→</span>
                                                    <span className="text-xs text-gray-400">{targetPort}</span>
                                                </>
                                            )}
                                        </div>
                                        <span className="text-[10px] font-bold text-gray-400 bg-gray-900/50 px-1.5 py-0.5 rounded">
                                            {protocol}
                                        </span>
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                    <div className="flex justify-end mt-4">
                        <EditYamlButton onClick={onEditYAML} namespace={namespace} />
                    </div>
                </>
            )}

            {activeTab === 'selector' && (
                <AssociatedPods namespace={namespace} selector={selector} />
            )}
        </div>
    );
};

export default ServiceDetails;
