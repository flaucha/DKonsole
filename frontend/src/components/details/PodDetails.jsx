import React, { useState, useEffect, useRef } from 'react';
import { Server, Network, Activity, Box, HardDrive, Clock, Terminal, Pin, PinOff, X, ChevronLeft, ChevronRight } from 'lucide-react';
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

const LIVE_TERMINAL_ID = 'live-terminal';

const createSessionId = () => {
    if (typeof crypto !== 'undefined' && crypto.randomUUID) {
        return crypto.randomUUID();
    }
    return `term-${Date.now()}-${Math.random().toString(16).slice(2)}`;
};

const SessionChip = ({ session, isActive, onSelect, onPinToggle, onClose, pinned, isLive }) => (
    <div
        onClick={onSelect}
        className={`group flex items-center gap-3 px-3 py-2 rounded-lg border cursor-pointer min-w-[200px] transition-all duration-200 ${
            isActive
                ? 'bg-gray-800 border-blue-500 shadow-lg shadow-blue-900/30'
                : 'bg-gray-800/50 border-gray-700 hover:border-gray-500 hover:bg-gray-800/80'
        }`}
    >
        <div className="flex items-center gap-2 truncate">
            <div className="flex flex-col leading-tight truncate">
                <div className="flex items-center gap-2">
                    <Terminal size={14} className="text-gray-300 shrink-0" />
                    <span className="text-xs font-semibold text-gray-100 truncate">{session.podName}</span>
                </div>
                <div className="text-[11px] text-gray-400 truncate">{session.container || '—'}</div>
            </div>
            {isLive && <span className="text-[10px] uppercase tracking-wide text-blue-300 bg-blue-900/30 border border-blue-700 px-1.5 py-0.5 rounded">Live</span>}
            {pinned && <span className="text-[10px] uppercase tracking-wide text-amber-300 bg-amber-900/30 border border-amber-700 px-1.5 py-0.5 rounded">Pinned</span>}
        </div>
        <div className="flex items-center gap-1 ml-auto shrink-0">
            {onPinToggle && (
                <button
                    onClick={(e) => {
                        e.stopPropagation();
                        onPinToggle();
                    }}
                    className={`p-1 rounded-md border text-gray-300 transition-colors ${
                        pinned
                            ? 'bg-amber-900/40 border-amber-700 text-amber-200 hover:bg-amber-900/60'
                            : 'bg-gray-800 border-gray-700 hover:border-gray-500 hover:text-white'
                    }`}
                    title={pinned ? 'Unpin session' : 'Pin session'}
                >
                    {pinned ? <PinOff size={14} /> : <Pin size={14} />}
                </button>
            )}
            {onClose && (
                <button
                    onClick={(e) => {
                        e.stopPropagation();
                        onClose();
                    }}
                    className="p-1 rounded-md border border-gray-700 text-gray-400 hover:text-red-300 hover:border-red-700 hover:bg-red-900/40 transition-colors"
                    title="Close session"
                >
                    <X size={14} />
                </button>
            )}
        </div>
    </div>
);

const PodDetails = ({ details, onEditYAML, pod }) => {
    const { authFetch } = useAuth();
    const [activeTab, setActiveTab] = useState('details');
    const [selectedContainer, setSelectedContainer] = useState(null);
    const [events, setEvents] = useState([]);
    const [loadingEvents, setLoadingEvents] = useState(false);
    const [pinnedTerminals, setPinnedTerminals] = useState([]);
    const [activeTerminalId, setActiveTerminalId] = useState(LIVE_TERMINAL_ID);
    const containers = details.containers || [];
    const metrics = details.metrics || {};
    const terminalContainerRef = useRef(null);
    const logsContainerRef = useRef(null);
    const pinnedScrollRef = useRef(null);
    const [canScrollLeft, setCanScrollLeft] = useState(false);
    const [canScrollRight, setCanScrollRight] = useState(false);

    // Set default container when component mounts or containers change
    useEffect(() => {
        if (containers.length > 0 && (!selectedContainer || !containers.includes(selectedContainer))) {
            setSelectedContainer(containers[0]);
            setActiveTerminalId(LIVE_TERMINAL_ID);
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

    const liveSession = pod && selectedContainer ? {
        id: LIVE_TERMINAL_ID,
        namespace: pod.namespace,
        podName: pod.name,
        container: selectedContainer
    } : null;

    const handlePinCurrentTerminal = () => {
        if (!liveSession) return;
        setPinnedTerminals(prev => {
            const existing = prev.find(session =>
                session.namespace === liveSession.namespace &&
                session.podName === liveSession.podName &&
                session.container === liveSession.container
            );
            if (existing) {
                setActiveTerminalId(existing.id);
                return prev;
            }
            const newSession = { ...liveSession, id: createSessionId() };
            setActiveTerminalId(newSession.id);
            return [...prev, newSession];
        });
    };

    const handleClosePinnedTerminal = (id) => {
        setPinnedTerminals(prev => {
            const remaining = prev.filter(session => session.id !== id);
            if (activeTerminalId === id) {
                const fallback = remaining[remaining.length - 1];
                setActiveTerminalId(fallback ? fallback.id : LIVE_TERMINAL_ID);
            }
            return remaining;
        });
    };

    const handleSelectSession = (id) => {
        setActiveTerminalId(id);
    };

    const handleUnpinTerminal = (session) => {
        handleClosePinnedTerminal(session.id);
    };

    const updateCarouselControls = () => {
        const el = pinnedScrollRef.current;
        if (!el) return;
        const maxScrollLeft = el.scrollWidth - el.clientWidth;
        setCanScrollLeft(el.scrollLeft > 4);
        setCanScrollRight(el.scrollLeft < maxScrollLeft - 4);
    };

    const scrollPinnedSessions = (direction) => {
        const el = pinnedScrollRef.current;
        if (!el) return;
        const delta = Math.max(el.clientWidth * 0.6, 240);
        el.scrollBy({ left: direction === 'left' ? -delta : delta, behavior: 'smooth' });
    };

    // Reset pinned sessions when switching pods
    useEffect(() => {
        setPinnedTerminals([]);
        setActiveTerminalId(LIVE_TERMINAL_ID);
    }, [pod?.name, pod?.namespace, pod?.uid]);

    // Ensure active session exists
    useEffect(() => {
        if (activeTerminalId !== LIVE_TERMINAL_ID && !pinnedTerminals.some(session => session.id === activeTerminalId)) {
            setActiveTerminalId(LIVE_TERMINAL_ID);
        }
    }, [pinnedTerminals, activeTerminalId]);

    // Update carousel controls on changes
    useEffect(() => {
        updateCarouselControls();
        const el = pinnedScrollRef.current;
        const handleResize = () => updateCarouselControls();

        if (el) {
            el.addEventListener('scroll', updateCarouselControls);
        }
        window.addEventListener('resize', handleResize);

        return () => {
            if (el) {
                el.removeEventListener('scroll', updateCarouselControls);
            }
            window.removeEventListener('resize', handleResize);
        };
    }, [pinnedTerminals.length, selectedContainer]);

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
                        <div className="flex justify-end mt-4">
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
                    <div ref={terminalContainerRef} className="animate-fadeIn flex-1 flex flex-col min-h-0 space-y-3">
                        {containers.length > 1 && (
                            <div className="mb-1 flex items-center space-x-2 flex-shrink-0">
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

                        <div className="bg-gray-900/60 border border-gray-800 rounded-lg p-3 shadow-inner flex flex-col gap-3">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2 text-xs uppercase tracking-wide text-gray-400">
                                    <Terminal size={14} className="text-gray-300" />
                                    <span>Terminal sessions</span>
                                </div>
                                <div className="flex items-center gap-2">
                                    <button
                                        onClick={() => handleSelectSession(LIVE_TERMINAL_ID)}
                                        className={`px-3 py-1.5 rounded-md border text-xs transition-colors ${activeTerminalId === LIVE_TERMINAL_ID
                                            ? 'border-blue-500 text-blue-200 bg-blue-900/30'
                                            : 'border-gray-700 text-gray-300 bg-gray-800 hover:border-gray-500 hover:text-white'
                                        }`}
                                    >
                                        Current view
                                    </button>
                                    <button
                                        onClick={handlePinCurrentTerminal}
                                        disabled={!liveSession}
                                        className="px-3 py-1.5 rounded-md border text-xs transition-colors bg-blue-600 text-white border-blue-500 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                                    >
                                        Pin current
                                    </button>
                                </div>
                            </div>

                            <div className="relative">
                                {canScrollLeft && (
                                    <button
                                        onClick={() => scrollPinnedSessions('left')}
                                        className="absolute left-0 top-1/2 -translate-y-1/2 z-10 p-1 rounded-full bg-gray-800 border border-gray-700 text-gray-300 hover:text-white hover:border-gray-500 shadow"
                                        title="Scroll left"
                                    >
                                        <ChevronLeft size={16} />
                                    </button>
                                )}
                                <div
                                    ref={pinnedScrollRef}
                                    className="flex gap-2 overflow-x-auto pb-1 pr-8"
                                >
                                    {liveSession ? (
                                        <SessionChip
                                            session={liveSession}
                                            isActive={activeTerminalId === LIVE_TERMINAL_ID}
                                            onSelect={() => handleSelectSession(LIVE_TERMINAL_ID)}
                                            onPinToggle={handlePinCurrentTerminal}
                                            pinned={false}
                                            isLive
                                        />
                                    ) : (
                                        <div className="text-xs text-gray-500 px-3 py-2">
                                            Select a container to start a terminal session.
                                        </div>
                                    )}
                                    {pinnedTerminals.map(session => (
                                        <SessionChip
                                            key={session.id}
                                            session={session}
                                            pinned
                                            isActive={activeTerminalId === session.id}
                                            onSelect={() => handleSelectSession(session.id)}
                                            onPinToggle={() => handleUnpinTerminal(session)}
                                            onClose={() => handleClosePinnedTerminal(session.id)}
                                        />
                                    ))}
                                </div>
                                {canScrollRight && (
                                    <button
                                        onClick={() => scrollPinnedSessions('right')}
                                        className="absolute right-0 top-1/2 -translate-y-1/2 z-10 p-1 rounded-full bg-gray-800 border border-gray-700 text-gray-300 hover:text-white hover:border-gray-500 shadow"
                                        title="Scroll right"
                                    >
                                        <ChevronRight size={16} />
                                    </button>
                                )}
                            </div>
                        </div>

                        <div className="flex-1 min-h-0 relative">
                            {liveSession && (
                                <TerminalViewerInline
                                    key={`${liveSession.namespace}-${liveSession.podName}-${liveSession.container}`}
                                    namespace={liveSession.namespace}
                                    pod={liveSession.podName}
                                    container={liveSession.container}
                                    isActive={activeTerminalId === LIVE_TERMINAL_ID}
                                    onPinToggle={handlePinCurrentTerminal}
                                    sessionLabel={`${liveSession.podName} / ${liveSession.container}`}
                                />
                            )}
                            {pinnedTerminals.map(session => (
                                <TerminalViewerInline
                                    key={session.id}
                                    namespace={session.namespace}
                                    pod={session.podName}
                                    container={session.container}
                                    isActive={activeTerminalId === session.id}
                                    onPinToggle={() => handleUnpinTerminal(session)}
                                    onClose={() => handleClosePinnedTerminal(session.id)}
                                    pinned
                                    sessionLabel={`${session.podName} / ${session.container}`}
                                />
                            ))}
                            {!liveSession && (
                                <div className="h-full border border-dashed border-gray-700 rounded-lg flex items-center justify-center text-gray-500 text-sm">
                                    Select a container to open a terminal session.
                                </div>
                            )}
                        </div>
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
