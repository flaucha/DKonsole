import React, { useState, useEffect, useRef } from 'react';
import { Server, Network, Activity, Box, HardDrive, Clock } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { DetailRow, EditYamlButton } from './CommonDetails';
import LogViewerInline from './LogViewerInline';
import TerminalViewerInline from './TerminalViewerInline';
import PodMetrics from '../PodMetrics';
import { formatDateTime, formatDateTimeShort } from '../../utils/dateUtils';
import { logger } from '../../utils/logger';

const TabButton = ({ active, label, onClick }) => (
    <button
        onClick={onClick}
        className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${
            active
            ? 'bg-gray-700 text-white shadow-sm'
            : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800'
        }`}
    >
        {label}
    </button>
);

const PodDetails = ({ details, onEditYAML, pod }) => {
    const { authFetch } = useAuth();
    const [activeTab, setActiveTab] = useState('details');
    const [selectedContainer, setSelectedContainer] = useState(null);
    const [events, setEvents] = useState([]);
    const [loadingEvents, setLoadingEvents] = useState(false);
    const containers = details.containers || [];
    const metrics = details.metrics || {};
    const terminalContainerRef = useRef(null);
    const logsContainerRef = useRef(null);

    // Set default container when component mounts or containers change
    useEffect(() => {
        if (containers.length > 0 && !selectedContainer) {
            setSelectedContainer(containers[0]);
        }
    }, [containers, selectedContainer]);

    // Scroll into view when terminal or logs tab is activated
    useEffect(() => {
        if (activeTab === 'terminal' && terminalContainerRef.current) {
            // Use setTimeout to ensure DOM is updated
            setTimeout(() => {
                terminalContainerRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            }, 100);
        } else if (activeTab === 'logs' && logsContainerRef.current) {
            // Use setTimeout to ensure DOM is updated
            setTimeout(() => {
                logsContainerRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            }, 100);
        }
    }, [activeTab]);

    // Fetch events when events tab is activated
    useEffect(() => {
        if (activeTab === 'events' && pod && !loadingEvents && events.length === 0) {
            setLoadingEvents(true);
            authFetch(`/api/pods/events?namespace=${pod.namespace}&pod=${pod.name}`)
                .then(res => res.json())
                .then(data => {
                    setEvents(data || []);
                    setLoadingEvents(false);
                })
                .catch(err => {
                    logger.error('Error fetching events:', err);
                    setEvents([]);
                    setLoadingEvents(false);
                });
        }
    }, [activeTab, pod, authFetch, loadingEvents, events.length]);

    // Determine if we should use full height (for logs/terminal tabs)
    const isFullHeightTab = activeTab === 'logs' || activeTab === 'terminal';
    // Calculate height: viewport height minus header, tabs, and padding (approximately)
    const fullHeight = 'calc(100vh - 300px)';

    return (
        <div className="mt-2 flex flex-col" style={{ height: isFullHeightTab ? fullHeight : 'auto' }}>
            <div className="flex space-x-1 bg-gray-800/50 p-1 rounded-md mb-4 w-fit">
                <TabButton active={activeTab === 'details'} label="Details" onClick={() => setActiveTab('details')} />
                <TabButton active={activeTab === 'logs'} label="Logs" onClick={() => setActiveTab('logs')} />
                <TabButton active={activeTab === 'terminal'} label="Terminal" onClick={() => setActiveTab('terminal')} />
                <TabButton active={activeTab === 'metrics'} label="Metrics" onClick={() => setActiveTab('metrics')} />
                <TabButton active={activeTab === 'events'} label="Events" onClick={() => setActiveTab('events')} />
            </div>

            <div className={`transition-all duration-300 ease-in-out flex-1 flex flex-col ${isFullHeightTab ? 'min-h-0' : ''}`}>
                {activeTab === 'details' ? (
                    <div className="p-4 bg-gray-900/50 rounded-md animate-fadeIn">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-3">
                            <div>
                                <DetailRow label="Node" value={details.node} icon={Server} />
                                <DetailRow label="IP" value={details.ip} icon={Network} />
                            </div>
                            <div>
                                <DetailRow label="Restarts" value={details.restarts} icon={Activity} />
                                <DetailRow label="Containers" value={containers} icon={Box} />
                            </div>
                            <div>
                                {metrics.cpu && <DetailRow label="CPU" value={metrics.cpu} icon={Activity} />}
                                {metrics.memory && <DetailRow label="Memory" value={metrics.memory} icon={HardDrive} />}
                            </div>
                        </div>
                        <div className="flex justify-end space-x-2">
                            <EditYamlButton onClick={onEditYAML} namespace={pod?.namespace} />
                        </div>
                    </div>
                ) : activeTab === 'logs' ? (
                    <div ref={logsContainerRef} className="animate-fadeIn flex-1 flex flex-col min-h-0">
                        {containers.length > 1 && (
                            <div className="mb-3 flex items-center space-x-2 flex-shrink-0">
                                <label className="text-xs text-gray-400">Container:</label>
                                <select
                                    value={selectedContainer || ''}
                                    onChange={(e) => setSelectedContainer(e.target.value)}
                                    className="px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 focus:outline-none focus:border-blue-500"
                                >
                                    {containers.map(c => (
                                        <option key={c} value={c}>{c}</option>
                                    ))}
                                </select>
                            </div>
                        )}
                        {selectedContainer && pod && (
                            <div className="flex-1 min-h-0">
                                <LogViewerInline
                                    namespace={pod.namespace}
                                    pod={pod.name}
                                    container={selectedContainer}
                                />
                            </div>
                        )}
                    </div>
                ) : activeTab === 'terminal' ? (
                    <div ref={terminalContainerRef} className="animate-fadeIn flex-1 flex flex-col min-h-0">
                        {containers.length > 1 && (
                            <div className="mb-3 flex items-center space-x-2 flex-shrink-0">
                                <label className="text-xs text-gray-400">Container:</label>
                                <select
                                    value={selectedContainer || ''}
                                    onChange={(e) => setSelectedContainer(e.target.value)}
                                    className="px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 focus:outline-none focus:border-blue-500"
                                >
                                    {containers.map(c => (
                                        <option key={c} value={c}>{c}</option>
                                    ))}
                                </select>
                            </div>
                        )}
                        {selectedContainer && pod && (
                            <div className="flex-1 min-h-0">
                                <TerminalViewerInline
                                    namespace={pod.namespace}
                                    pod={pod.name}
                                    container={selectedContainer}
                                />
                            </div>
                        )}
                    </div>
                ) : activeTab === 'metrics' ? (
                    <div className="animate-fadeIn">
                        {pod && <PodMetrics pod={{ name: pod.name }} namespace={pod.namespace} />}
                    </div>
                ) : (
                    <div className="animate-fadeIn p-4 bg-gray-900/50 rounded-md">
                        <div className="space-y-4">
                            {/* Events Section */}
                            <div>
                                <h4 className="text-sm font-semibold text-gray-300 mb-3 flex items-center">
                                    <Activity size={16} className="mr-2" />
                                    Pod Events
                                </h4>
                                {loadingEvents ? (
                                    <div className="text-gray-500 italic">Loading events...</div>
                                ) : events.length === 0 ? (
                                    <div className="text-gray-500 italic">No events available</div>
                                ) : (
                                    <div className="space-y-2 max-h-96 overflow-y-auto">
                                        {events.map((event, idx) => (
                                            <div key={idx} className={`p-3 rounded-md border ${event.type === 'Warning' ? 'bg-yellow-900/20 border-yellow-700' : 'bg-blue-900/20 border-blue-700'}`}>
                                                <div className="flex items-start justify-between">
                                                    <div className="flex-1">
                                                        <div className="flex items-center space-x-2 mb-1">
                                                            <span className={`text-xs font-semibold ${event.type === 'Warning' ? 'text-yellow-400' : 'text-blue-400'}`}>
                                                                {event.type}
                                                            </span>
                                                            <span className="text-xs text-gray-400">•</span>
                                                            <span className="text-xs font-medium text-gray-300">{event.reason}</span>
                                                            {event.count > 1 && (
                                                                <>
                                                                    <span className="text-xs text-gray-400">•</span>
                                                                    <span className="text-xs text-gray-500">x{event.count}</span>
                                                                </>
                                                            )}
                                                        </div>
                                                        <p className="text-xs text-gray-300 mb-1">{event.message}</p>
                                                        {event.source && (
                                                            <p className="text-xs text-gray-500">Source: {event.source}</p>
                                                        )}
                                                    </div>
                                                    <div className="text-xs text-gray-500 ml-4 text-right">
                                                        <div>{formatDateTime(event.lastSeen)}</div>
                                                        {event.firstSeen !== event.lastSeen && (
                                                            <div className="text-gray-600 mt-1">
                                                                First seen: {formatDateTimeShort(event.firstSeen)}
                                                            </div>
                                                        )}
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>

                            {/* Container Status Timeline */}
                            <div>
                                <h4 className="text-sm font-semibold text-gray-300 mb-3 flex items-center">
                                    <Clock size={16} className="mr-2" />
                                    Container Status Timeline
                                </h4>
                                {details.containerStatuses && details.containerStatuses.length > 0 ? (
                                    <div className="space-y-3">
                                        {details.containerStatuses.map((containerStatus, idx) => (
                                            <div key={idx} className="p-3 bg-gray-800 rounded-md border border-gray-700">
                                                <div className="flex items-center justify-between mb-2">
                                                    <div className="flex items-center space-x-2">
                                                        <span className="text-sm font-medium text-white">{containerStatus.name}</span>
                                                        <span className={`px-2 py-0.5 text-xs rounded ${containerStatus.ready ? 'bg-green-900/30 text-green-400' : 'bg-red-900/30 text-red-400'}`}>
                                                            {containerStatus.ready ? 'Ready' : 'Not Ready'}
                                                        </span>
                                                        <span className={`px-2 py-0.5 text-xs rounded ${containerStatus.state === 'Running' ? 'bg-blue-900/30 text-blue-400' :
                                                                containerStatus.state === 'Waiting' ? 'bg-yellow-900/30 text-yellow-400' :
                                                                    containerStatus.state === 'Terminated' ? 'bg-gray-700 text-gray-400' :
                                                                        'bg-gray-700 text-gray-400'
                                                            }`}>
                                                            {containerStatus.state || 'Unknown'}
                                                        </span>
                                                    </div>
                                                    <div className="text-xs text-gray-500">
                                                        Restarts: {containerStatus.restartCount || 0}
                                                    </div>
                                                </div>

                                                {containerStatus.state === 'Waiting' && containerStatus.reason && (
                                                    <div className="mt-2 text-xs text-yellow-300">
                                                        <span className="font-medium">Reason:</span> {containerStatus.reason}
                                                        {containerStatus.message && (
                                                            <span className="block mt-1 text-gray-400">{containerStatus.message}</span>
                                                        )}
                                                    </div>
                                                )}

                                                {containerStatus.state === 'Running' && containerStatus.startedAt && (
                                                    <div className="mt-2 text-xs text-gray-400">
                                                        Started: {formatDateTime(containerStatus.startedAt)}
                                                    </div>
                                                )}

                                                {containerStatus.state === 'Terminated' && (
                                                    <div className="mt-2 space-y-1">
                                                        {containerStatus.reason && (
                                                            <div className="text-xs text-red-300">
                                                                <span className="font-medium">Reason:</span> {containerStatus.reason}
                                                            </div>
                                                        )}
                                                        {containerStatus.exitCode !== undefined && (
                                                            <div className="text-xs text-gray-400">
                                                                Exit Code: {containerStatus.exitCode}
                                                            </div>
                                                        )}
                                                        {containerStatus.startedAt && (
                                                            <div className="text-xs text-gray-400">
                                                                Started: {formatDateTime(containerStatus.startedAt)}
                                                            </div>
                                                        )}
                                                        {containerStatus.finishedAt && (
                                                            <div className="text-xs text-gray-400">
                                                                Finished: {formatDateTime(containerStatus.finishedAt)}
                                                            </div>
                                                        )}
                                                    </div>
                                                )}

                                                {containerStatus.image && (
                                                    <div className="mt-2 text-xs text-gray-500">
                                                        Image: {containerStatus.image}
                                                    </div>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="text-gray-500 italic">No container status information available</div>
                                )}
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default PodDetails;
