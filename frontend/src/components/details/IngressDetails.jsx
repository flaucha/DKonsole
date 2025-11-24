import React from 'react';
import { Globe } from 'lucide-react';
import { EditYamlButton } from './CommonDetails';

const IngressDetails = ({ details, onEditYAML }) => {
    const rules = details.rules || [];
    const tls = details.tls || [];
    const loadBalancer = details.loadBalancer || [];

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2 space-y-6">
            {/* LoadBalancer Status */}
            {loadBalancer.length > 0 && (
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Address</h4>
                    <div className="flex flex-wrap gap-2">
                        {loadBalancer.map((lb, i) => (
                            <div key={i} className="flex items-center px-2 py-1 bg-gray-800 border border-gray-700 rounded text-sm text-white">
                                <Globe size={14} className="mr-2 text-blue-400" />
                                {lb.ip || lb.hostname}
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* Rules Section */}
            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Rules</h4>
                {rules.length > 0 ? (
                    <div className="space-y-3">
                        {rules.map((rule, i) => (
                            <div key={i} className="bg-gray-800 rounded border border-gray-700 overflow-hidden">
                                <div className="px-3 py-2 bg-gray-800/50 border-b border-gray-700 flex items-center">
                                    <span className="text-xs text-gray-500 uppercase mr-2">Host:</span>
                                    <span className="text-sm font-medium text-white">{rule.host || '*'}</span>
                                </div>
                                <div className="p-2 space-y-1">
                                    {rule.paths && rule.paths.map((path, j) => (
                                        <div key={j} className="flex items-center text-xs text-gray-300 pl-2">
                                            <div className="w-1.5 h-1.5 rounded-full bg-gray-600 mr-2"></div>
                                            {path}
                                        </div>
                                    ))}
                                </div>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className="text-sm text-gray-500 italic">No rules defined</div>
                )}
            </div>

            {/* TLS Section */}
            {tls.length > 0 && (
                <div>
                    <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">TLS Configuration</h4>
                    <div className="space-y-2">
                        {tls.map((t, i) => (
                            <div key={i} className="flex flex-col bg-gray-800 px-3 py-2 rounded border border-gray-700">
                                <div className="flex items-center mb-1">
                                    <span className="text-xs text-gray-500 uppercase mr-2">Secret:</span>
                                    <span className="text-sm font-medium text-white">{t.secretName}</span>
                                </div>
                                {t.hosts && t.hosts.length > 0 && (
                                    <div className="ml-6 text-xs text-gray-400">
                                        <span className="text-gray-500 mr-1">Hosts:</span>
                                        {t.hosts.join(', ')}
                                    </div>
                                )}
                            </div>
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

export default IngressDetails;
